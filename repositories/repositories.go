package repositories

import (
	"context"
	"log/slog"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/roysitumorang/laukpauk/helper"
)

var (
	dbWrite *pgxpool.Pool
)

func SetDbWrite(db *pgxpool.Pool) {
	dbWrite = db
}

func BeginTx(ctx context.Context) (pgx.Tx, error) {
	ctxt := "Repositories-BeginTx"
	tx, err := dbWrite.Begin(ctx)
	if err != nil {
		helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrBegin")
	}
	return tx, err
}
