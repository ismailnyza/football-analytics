package sqlite

import (
	"context"
	"database/sql"
)

type result interface {
	LastInsertId() (int64, error)
}

type row interface {
	Scan(dest ...any) error
}

type rows interface {
	Close() error
	Err() error
	Next() bool
	Scan(dest ...any) error
}

type queryer interface {
	ExecContext(ctx context.Context, query string, args ...any) (result, error)
	QueryContext(ctx context.Context, query string, args ...any) (rows, error)
	QueryRowContext(ctx context.Context, query string, args ...any) row
}

type transaction interface {
	queryer
	Commit() error
	Rollback() error
}

type beginTxer interface {
	queryer
	BeginTx(ctx context.Context, opts *sql.TxOptions) (transaction, error)
}

type sqlDB struct {
	*sql.DB
}

func (db sqlDB) ExecContext(ctx context.Context, query string, args ...any) (result, error) {
	return db.DB.ExecContext(ctx, query, args...)
}

func (db sqlDB) QueryContext(ctx context.Context, query string, args ...any) (rows, error) {
	return db.DB.QueryContext(ctx, query, args...)
}

func (db sqlDB) QueryRowContext(ctx context.Context, query string, args ...any) row {
	return db.DB.QueryRowContext(ctx, query, args...)
}

func (db sqlDB) BeginTx(ctx context.Context, opts *sql.TxOptions) (transaction, error) {
	tx, err := db.DB.BeginTx(ctx, opts)
	if err != nil {
		return nil, err
	}
	return sqlTx{Tx: tx}, nil
}

type sqlTx struct {
	*sql.Tx
}

func (tx sqlTx) ExecContext(ctx context.Context, query string, args ...any) (result, error) {
	return tx.Tx.ExecContext(ctx, query, args...)
}

func (tx sqlTx) QueryContext(ctx context.Context, query string, args ...any) (rows, error) {
	return tx.Tx.QueryContext(ctx, query, args...)
}

func (tx sqlTx) QueryRowContext(ctx context.Context, query string, args ...any) row {
	return tx.Tx.QueryRowContext(ctx, query, args...)
}
