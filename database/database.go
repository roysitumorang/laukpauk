// Package database is a small pgx-style layer over database/sql for MariaDB: the pool, transactions and rows keep
// the Exec / Query / QueryRow(ctx, sql, args...) shape the repositories were written for, with ? placeholders, and
// UUIDs, JSON values and times are converted on the way in and out.
package database

import (
	"context"
	"database/sql"
	"errors"
)

var (
	// ErrNoRows is returned by Row.Scan when the query selected no rows.
	ErrNoRows = sql.ErrNoRows
	// ErrTxClosed is returned by Commit and Rollback on a transaction that is already committed or rolled back.
	ErrTxClosed = sql.ErrTxDone
)

// IsoLevel is a transaction isolation level.
type IsoLevel = sql.IsolationLevel

const (
	ReadCommitted  IsoLevel = sql.LevelReadCommitted
	RepeatableRead IsoLevel = sql.LevelRepeatableRead
	Serializable   IsoLevel = sql.LevelSerializable
)

// TxOptions are the options of Pool.BeginTx.
type TxOptions struct {
	IsoLevel IsoLevel
	ReadOnly bool
}

// Result is what Exec returns.
type Result interface {
	RowsAffected() int64
}

// Row is the result of QueryRow.
type Row interface {
	Scan(dest ...any) error
}

// Rows is the result of Query.
type Rows interface {
	Next() bool
	Scan(dest ...any) error
	Close()
	Err() error
}

// Querier runs statements; the pool and transactions both implement it.
type Querier interface {
	Exec(ctx context.Context, query string, args ...any) (Result, error)
	Query(ctx context.Context, query string, args ...any) (Rows, error)
	QueryRow(ctx context.Context, query string, args ...any) Row
}

// Tx is a database transaction.
type Tx interface {
	Querier
	Commit(ctx context.Context) error
	Rollback(ctx context.Context) error
}

// Pool is a pool of connections.
type Pool struct {
	db *sql.DB
}

// NewPool wraps an open *sql.DB.
func NewPool(db *sql.DB) *Pool {
	return &Pool{db: db}
}

// DB is the underlying *sql.DB.
func (p *Pool) DB() *sql.DB {
	return p.db
}

func (p *Pool) Ping(ctx context.Context) error {
	return p.db.PingContext(ctx)
}

func (p *Pool) Close() {
	_ = p.db.Close()
}

func (p *Pool) Exec(ctx context.Context, query string, args ...any) (Result, error) {
	return exec(ctx, p.db, query, args)
}

func (p *Pool) Query(ctx context.Context, query string, args ...any) (Rows, error) {
	return queryRows(ctx, p.db, query, args)
}

func (p *Pool) QueryRow(ctx context.Context, query string, args ...any) Row {
	return queryRow(ctx, p.db, query, args)
}

// Begin starts a transaction with the server's default isolation level.
func (p *Pool) Begin(ctx context.Context) (Tx, error) {
	return p.BeginTx(ctx, TxOptions{})
}

// BeginTx starts a transaction.
func (p *Pool) BeginTx(ctx context.Context, opts TxOptions) (Tx, error) {
	tx, err := p.db.BeginTx(ctx, &sql.TxOptions{Isolation: opts.IsoLevel, ReadOnly: opts.ReadOnly})
	if err != nil {
		return nil, err
	}
	return &transaction{tx: tx}, nil
}

type transaction struct {
	tx *sql.Tx
}

func (t *transaction) Exec(ctx context.Context, query string, args ...any) (Result, error) {
	return exec(ctx, t.tx, query, args)
}

func (t *transaction) Query(ctx context.Context, query string, args ...any) (Rows, error) {
	return queryRows(ctx, t.tx, query, args)
}

func (t *transaction) QueryRow(ctx context.Context, query string, args ...any) Row {
	return queryRow(ctx, t.tx, query, args)
}

func (t *transaction) Commit(context.Context) error {
	return t.tx.Commit()
}

func (t *transaction) Rollback(context.Context) error {
	return t.tx.Rollback()
}

// runner is what *sql.DB and *sql.Tx have in common.
type runner interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
}

type result struct {
	rowsAffected int64
}

func (r result) RowsAffected() int64 {
	return r.rowsAffected
}

func exec(ctx context.Context, r runner, query string, args []any) (Result, error) {
	args, err := prepare(query, args)
	if err != nil {
		return nil, err
	}
	res, err := r.ExecContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return nil, err
	}
	return result{rowsAffected: n}, nil
}

func queryRows(ctx context.Context, r runner, query string, args []any) (Rows, error) {
	args, err := prepare(query, args)
	if err != nil {
		return nil, err
	}
	rows, err := r.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	return &resultRows{rows: rows}, nil
}

func queryRow(ctx context.Context, r runner, query string, args []any) Row {
	rows, err := queryRows(ctx, r, query, args)
	return &resultRow{rows: rows, err: err}
}

type resultRows struct {
	rows *sql.Rows
}

func (r *resultRows) Next() bool {
	return r.rows.Next()
}

func (r *resultRows) Scan(dest ...any) error {
	return scan(r.rows, dest)
}

func (r *resultRows) Close() {
	_ = r.rows.Close()
}

func (r *resultRows) Err() error {
	return r.rows.Err()
}

// resultRow scans the first row and closes the rows, like *sql.Row, so the connection is free again right after.
type resultRow struct {
	rows Rows
	err  error
}

func (r *resultRow) Scan(dest ...any) error {
	if r.err != nil {
		return r.err
	}
	defer r.rows.Close()
	if !r.rows.Next() {
		if err := r.rows.Err(); err != nil {
			return err
		}
		return ErrNoRows
	}
	if err := r.rows.Scan(dest...); err != nil {
		return err
	}
	return r.rows.Err()
}

// scan scans one row, wrapping the destinations database/sql can't fill by itself.
func scan(rows *sql.Rows, dest []any) error {
	wrapped := make([]any, len(dest))
	for i, d := range dest {
		wrapped[i] = wrapDest(d)
	}
	return rows.Scan(wrapped...)
}

// IsNoRows reports whether err means the query selected no rows.
func IsNoRows(err error) bool {
	return errors.Is(err, ErrNoRows)
}
