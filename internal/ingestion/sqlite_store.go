package ingestion

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"time"

	storagesqlite "github.com/ismael/football-analytics/internal/storage/sqlite"
	_ "modernc.org/sqlite"
)

// SQLiteStagingStore persists staged ingest records in the app SQLite database.
type SQLiteStagingStore struct {
	db *sql.DB
}

func NewSQLiteStagingStore(stateDir string) (*SQLiteStagingStore, func() error, error) {
	if err := os.MkdirAll(stateDir, 0o755); err != nil {
		return nil, nil, fmt.Errorf("create state dir: %w", err)
	}
	db, err := sql.Open("sqlite", storagesqlite.DefaultPath(stateDir))
	if err != nil {
		return nil, nil, fmt.Errorf("open sqlite staging store: %w", err)
	}
	if err := storagesqlite.NewStore(db).Migrate(context.Background()); err != nil {
		_ = db.Close()
		return nil, nil, fmt.Errorf("migrate sqlite staging store: %w", err)
	}
	return &SQLiteStagingStore{db: db}, db.Close, nil
}

func (s *SQLiteStagingStore) StageRecord(ctx context.Context, record SourceRecord) error {
	if record.SourceName == "" || record.ExternalID == "" {
		return fmt.Errorf("staging: source name and external ID are required")
	}
	sourceID, err := s.ensureSource(ctx, record.SourceName)
	if err != nil {
		return err
	}
	fetchedAt := record.FetchedAt
	if fetchedAt.IsZero() {
		fetchedAt = time.Now().UTC()
	}
	_, err = s.db.ExecContext(
		ctx,
		`INSERT INTO raw_payloads(source_id, external_id, fetched_at, payload_json, checksum, processed_at)
		VALUES (?, ?, ?, ?, ?, NULL)`,
		sourceID,
		record.ExternalID,
		fetchedAt.Format(time.RFC3339Nano),
		record.RawJSON,
		record.EntityType,
	)
	if err != nil {
		return fmt.Errorf("insert raw payload: %w", err)
	}
	return nil
}

func (s *SQLiteStagingStore) ListStaged(ctx context.Context, source, entityType string) ([]SourceRecord, error) {
	query := `SELECT rs.code, COALESCE(rp.checksum, ''), rp.external_id, rp.payload_json, rp.fetched_at
		FROM raw_payloads rp
		INNER JOIN raw_sources rs ON rs.id = rp.source_id
		WHERE rp.processed_at IS NULL`
	args := make([]any, 0, 2)
	if source != "" {
		query += ` AND rs.code = ?`
		args = append(args, source)
	}
	if entityType != "" {
		query += ` AND COALESCE(rp.checksum, '') = ?`
		args = append(args, entityType)
	}
	query += ` ORDER BY rp.id`

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("query staged payloads: %w", err)
	}
	defer rows.Close()

	var out []SourceRecord
	for rows.Next() {
		var record SourceRecord
		var fetchedAt string
		if err := rows.Scan(&record.SourceName, &record.EntityType, &record.ExternalID, &record.RawJSON, &fetchedAt); err != nil {
			return nil, fmt.Errorf("scan staged payload: %w", err)
		}
		if fetchedAt != "" {
			if parsed, err := time.Parse(time.RFC3339Nano, fetchedAt); err == nil {
				record.FetchedAt = parsed
			}
		}
		out = append(out, record)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate staged payloads: %w", err)
	}
	return out, nil
}

func (s *SQLiteStagingStore) MarkProcessed(ctx context.Context, source, externalID string) error {
	_, err := s.db.ExecContext(
		ctx,
		`UPDATE raw_payloads
		SET processed_at = ?
		WHERE external_id = ? AND source_id = (SELECT id FROM raw_sources WHERE code = ? LIMIT 1) AND processed_at IS NULL`,
		time.Now().UTC().Format(time.RFC3339Nano),
		externalID,
		source,
	)
	if err != nil {
		return fmt.Errorf("mark processed: %w", err)
	}
	return nil
}

func (s *SQLiteStagingStore) ensureSource(ctx context.Context, sourceName string) (int64, error) {
	_, err := s.db.ExecContext(
		ctx,
		`INSERT INTO raw_sources(code, description) VALUES (?, ?) ON CONFLICT(code) DO NOTHING`,
		sourceName,
		sourceName+" staging source",
	)
	if err != nil {
		return 0, fmt.Errorf("ensure source: %w", err)
	}
	var sourceID int64
	if err := s.db.QueryRowContext(ctx, `SELECT id FROM raw_sources WHERE code = ?`, sourceName).Scan(&sourceID); err != nil {
		return 0, fmt.Errorf("lookup source id: %w", err)
	}
	return sourceID, nil
}

func resolveSQLiteStateStore(stateDir string) (*SQLiteStagingStore, func() error, error) {
	store, closeFn, err := NewSQLiteStagingStore(stateDir)
	if err != nil {
		return nil, nil, err
	}
	return store, closeFn, nil
}
