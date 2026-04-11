package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/ismael/football-analytics/internal/domain"
)

func TestMigrateAppliesPendingMigrations(t *testing.T) {
	tx := &fakeTx{
		rowValues: map[string][]any{
			"SELECT COUNT(1) FROM schema_migrations": {0},
		},
	}
	store := newStoreFromBeginTxer(&fakeDB{tx: tx})

	if err := store.Migrate(context.Background()); err != nil {
		t.Fatalf("Migrate() error = %v", err)
	}

	if !tx.committed {
		t.Fatal("Migrate() did not commit transaction")
	}
	if len(tx.execs) < 3 {
		t.Fatalf("expected bootstrap + migration execs, got %d", len(tx.execs))
	}
	if !strings.Contains(tx.execs[1].query, "CREATE TABLE IF NOT EXISTS schema_migrations") &&
		!strings.Contains(tx.execs[1].query, "PRAGMA foreign_keys = ON") {
		t.Fatalf("unexpected migration exec query %q", tx.execs[1].query)
	}
}

func TestCreateWorldAssignsLastInsertID(t *testing.T) {
	store := newStoreFromBeginTxer(&fakeDB{
		execResults: map[string]fakeResult{
			"INSERT INTO worlds": {lastInsertID: 42},
		},
	})

	world, err := store.CreateWorld(context.Background(), domain.World{Name: "Test", Seed: 7})
	if err != nil {
		t.Fatalf("CreateWorld() error = %v", err)
	}
	if world.ID != 42 {
		t.Fatalf("CreateWorld() id = %d, want 42", world.ID)
	}
}

func TestListPlayersByClubDecodesPositions(t *testing.T) {
	clubID := int64(99)
	birthDate := time.Date(2000, 2, 3, 0, 0, 0, 0, time.UTC)
	store := newStoreFromBeginTxer(&fakeDB{
		queryRows: map[string]*fakeRows{
			"FROM players": {
				values: [][]any{
					{int64(1), int64(10), sql.NullInt64{Int64: clubID, Valid: true}, "ext", "Ada", "Lovelace", "Ada", sql.NullTime{Time: birthDate, Valid: true}, "England", "CM", "AM,RW", 81, 88},
				},
			},
		},
	})

	players, err := store.ListPlayersByClub(context.Background(), 10, 99)
	if err != nil {
		t.Fatalf("ListPlayersByClub() error = %v", err)
	}
	if len(players) != 1 {
		t.Fatalf("ListPlayersByClub() len = %d, want 1", len(players))
	}
	if got, want := players[0].PrimaryPosition, domain.PositionCM; got != want {
		t.Fatalf("PrimaryPosition = %q, want %q", got, want)
	}
	if got, want := players[0].SecondaryPositions, []domain.Position{domain.PositionAM, domain.PositionRW}; !reflect.DeepEqual(got, want) {
		t.Fatalf("SecondaryPositions = %#v, want %#v", got, want)
	}
}

func TestEncodePositionsRejectsInvalidPosition(t *testing.T) {
	_, err := encodePositions([]domain.Position{"??"})
	if err == nil {
		t.Fatal("encodePositions() error = nil, want error")
	}
}

type fakeDB struct {
	tx          *fakeTx
	execResults map[string]fakeResult
	queryRows   map[string]*fakeRows
	rowValues   map[string][]any
}

func (f *fakeDB) BeginTx(context.Context, *sql.TxOptions) (transaction, error) {
	if f.tx == nil {
		return nil, errors.New("no transaction configured")
	}
	return f.tx, nil
}

func (f *fakeDB) ExecContext(_ context.Context, query string, _ ...any) (result, error) {
	for fragment, res := range f.execResults {
		if strings.Contains(query, fragment) {
			return res, nil
		}
	}
	return fakeResult{lastInsertID: 1}, nil
}

func (f *fakeDB) QueryContext(_ context.Context, query string, _ ...any) (rows, error) {
	for fragment, rows := range f.queryRows {
		if strings.Contains(query, fragment) {
			rows.index = 0
			return rows, nil
		}
	}
	return &fakeRows{}, nil
}

func (f *fakeDB) QueryRowContext(_ context.Context, query string, _ ...any) row {
	for fragment, values := range f.rowValues {
		if strings.Contains(query, fragment) {
			return &fakeRow{values: values}
		}
	}
	return &fakeRow{err: sql.ErrNoRows}
}

type fakeTx struct {
	execs     []fakeExec
	rowValues map[string][]any
	committed bool
}

func (f *fakeTx) ExecContext(_ context.Context, query string, _ ...any) (result, error) {
	f.execs = append(f.execs, fakeExec{query: query})
	return fakeResult{lastInsertID: int64(len(f.execs))}, nil
}

func (f *fakeTx) QueryContext(_ context.Context, _ string, _ ...any) (rows, error) {
	return &fakeRows{}, nil
}

func (f *fakeTx) QueryRowContext(_ context.Context, query string, _ ...any) row {
	for fragment, values := range f.rowValues {
		if strings.Contains(query, fragment) {
			return &fakeRow{values: values}
		}
	}
	return &fakeRow{err: sql.ErrNoRows}
}

func (f *fakeTx) Commit() error {
	f.committed = true
	return nil
}

func (f *fakeTx) Rollback() error {
	return nil
}

type fakeExec struct {
	query string
}

type fakeResult struct {
	lastInsertID int64
}

func (f fakeResult) LastInsertId() (int64, error) {
	return f.lastInsertID, nil
}

type fakeRows struct {
	values [][]any
	index  int
}

func (f *fakeRows) Close() error { return nil }
func (f *fakeRows) Err() error   { return nil }
func (f *fakeRows) Next() bool {
	if f.index >= len(f.values) {
		return false
	}
	f.index++
	return true
}

func (f *fakeRows) Scan(dest ...any) error {
	if f.index == 0 || f.index > len(f.values) {
		return errors.New("scan called out of sequence")
	}
	return assign(dest, f.values[f.index-1])
}

type fakeRow struct {
	values []any
	err    error
}

func (f *fakeRow) Scan(dest ...any) error {
	if f.err != nil {
		return f.err
	}
	return assign(dest, f.values)
}

func assign(dest []any, values []any) error {
	if len(dest) != len(values) {
		return errors.New("destination/value length mismatch")
	}
	for i := range dest {
		switch d := dest[i].(type) {
		case *int:
			*d = values[i].(int)
		case *int64:
			*d = values[i].(int64)
		case *string:
			*d = values[i].(string)
		case *sql.NullInt64:
			*d = values[i].(sql.NullInt64)
		case *sql.NullTime:
			*d = values[i].(sql.NullTime)
		case *time.Time:
			*d = values[i].(time.Time)
		default:
			return errors.New("unsupported scan destination")
		}
	}
	return nil
}
