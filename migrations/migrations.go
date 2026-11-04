package migrations

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"maps"
	"os"
	"slices"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/roysitumorang/laukpauk/helper"
)

type Migration struct {
	dbWrite *pgxpool.Pool
}

var Migrations = map[int64]func(ctx context.Context, tx pgx.Tx) error{}

func NewMigration(
	dbWrite *pgxpool.Pool,
) *Migration {
	return &Migration{
		dbWrite: dbWrite,
	}
}

func (m *Migration) Migrate(ctx context.Context) error {
	ctxt := "Migration-Migrate"
	tx, err := m.dbWrite.BeginTx(ctx, pgx.TxOptions{
		IsoLevel: pgx.Serializable,
	})
	if err != nil {
		helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrBeginTx")
		return err
	}
	defer func(ctx context.Context) {
		if err = tx.Rollback(ctx); errors.Is(err, pgx.ErrTxClosed) {
			err = nil
		}
		if err != nil {
			helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrRollback")
		}
	}(ctx)
	if _, err = tx.Exec(
		ctx,
		"CREATE EXTENSION IF NOT EXISTS btree_gist;",
	); err != nil {
		helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
		return err
	}
	if _, err = tx.Exec(
		ctx,
		`CREATE COLLATION IF NOT EXISTS natural_sort (
			PROVIDER = 'icu'
			, LOCALE = 'en-US-u-kn-true'
		);`,
	); err != nil {
		helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
		return err
	}
	if _, err = tx.Exec(
		ctx,
		`CREATE TABLE IF NOT EXISTS migrations (
			"version" bigint NOT NULL PRIMARY KEY
		)`,
	); err != nil {
		helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
		return err
	}
	rows, err := tx.Query(ctx, `SELECT "version" FROM "migrations" ORDER BY "version"`)
	if errors.Is(err, pgx.ErrNoRows) {
		err = nil
	}
	if err != nil {
		helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrQuery")
		return err
	}
	defer rows.Close()
	mapVersions := map[int64]struct{}{}
	for rows.Next() {
		var version int64
		if err := rows.Scan(&version); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrScan")
			return err
		}
		mapVersions[version] = struct{}{}
	}
	sortedVersions := slices.Collect(maps.Keys(Migrations))
	slices.Sort(sortedVersions)
	for _, version := range sortedVersions {
		if _, ok := mapVersions[version]; ok {
			continue
		}
		function, ok := Migrations[version]
		if !ok {
			return fmt.Errorf("migration function for version %d not found", version)
		}
		if err = function(ctx, tx); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrFunction")
			return err
		}
		if _, err = tx.Exec(ctx, `INSERT INTO "migrations" ("version") VALUES ($1)`, version); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
			return err
		}
	}
	if err = tx.Commit(ctx); err != nil {
		helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrCommit")
		return err
	}
	clear(Migrations)
	return nil
}

func (m *Migration) CreateMigrationFile() error {
	now := time.Now().UTC().UnixNano()
	filepath := helper.Sprintf("./migrations/%d.go", now)
	content := helper.Sprintf(
		`package migrations

import (
	"context"

	"github.com/jackc/pgx/v5"
)

func init() {
	Migrations[%d] = func(ctx context.Context, tx pgx.Tx) (err error) {
		ctxt := "Migration-%d"
		return
	}
}`,
		now,
		now,
	)
	return os.WriteFile(
		filepath,
		helper.String2ByteSlice(content),
		0600,
	)
}
