package migrations

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"maps"
	"os"
	"slices"
	"time"

	"github.com/roysitumorang/laukpauk/helper"
)

type Migrator struct {
	dbWrite *sql.DB
}

type Migration struct {
	Version int64
	Up      func(ctx context.Context, tx *sql.Tx) error
	Down    func(ctx context.Context, tx *sql.Tx) error
}

var Migrations = map[int64]Migration{}

func NewMigrator(
	dbWrite *sql.DB,
) *Migrator {
	return &Migrator{
		dbWrite: dbWrite,
	}
}

// Migrate runs the migrations that haven't run yet, in version order. MariaDB commits each CREATE / ALTER / DROP on
// its own, so a migration that fails halfway is not rolled back by the transaction: its Down undoes what it can.
func (m *Migrator) Migrate(ctx context.Context) error {
	ctxt := "Migrator-Migrate"
	tx, err := m.dbWrite.BeginTx(ctx, &sql.TxOptions{
		Isolation: sql.LevelSerializable,
	})
	if err != nil {
		helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrBeginTx")
		return err
	}
	defer func(ctx context.Context) {
		if err = tx.Rollback(); errors.Is(err, sql.ErrTxDone) {
			err = nil
		}
		if err != nil {
			helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrRollback")
		}
	}(ctx)
	if _, err = tx.ExecContext(
		ctx,
		`CREATE TABLE IF NOT EXISTS migrations (
			"version" bigint NOT NULL PRIMARY KEY
		) ENGINE=InnoDB`,
	); err != nil {
		helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
		return err
	}
	rows, err := tx.QueryContext(ctx, `SELECT "version" FROM "migrations" ORDER BY "version"`)
	if errors.Is(err, sql.ErrNoRows) {
		err = nil
	}
	if err != nil {
		helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrQuery")
		return err
	}
	defer func() { _ = rows.Close() }()
	mapVersions := map[int64]struct{}{}
	for rows.Next() {
		var version int64
		if err = rows.Scan(helper.SQLScan(&version)...); err != nil {
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
		migration, ok := Migrations[version]
		if !ok {
			return fmt.Errorf("migration version %d not found", version)
		}
		if err = migration.Up(ctx, tx); err != nil {
			if errRollback := migration.Down(ctx, tx); errRollback != nil {
				helper.Capture(ctx, slog.LevelError, errRollback, ctxt, "ErrUp")
			}
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrUp")
			return err
		}
		if _, err = tx.ExecContext(ctx, `INSERT INTO "migrations" ("version") VALUES (?)`, helper.SQLArgs(version)...); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
			return err
		}
	}
	if err = tx.Commit(); err != nil {
		helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrCommit")
		return err
	}
	clear(Migrations)
	return nil
}

func (m *Migrator) CreateMigrationFile() error {
	now := time.Now().UTC().UnixNano()
	filepath := helper.Sprintf("./migrations/%d.go", now)
	content := helper.Sprintf(
		`package migrations

import (
	"context"

)

func init() {
	Migrations[%d] = Migration{
		Version: %d,
		Up: func(ctx context.Context, tx *sql.Tx) (err error) {
			ctxt := "Migration-Up"
			return
		},
		Down: func(ctx context.Context, tx *sql.Tx) (err error) {
			ctxt := "Migration-Down"
			return
		},
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

// seedTime parses a seed timestamp such as "2024-12-11 07:41:55.000 +0700": MariaDB doesn't take time zone offsets in
// DATETIME text, so seeds pass time.Time values.
func seedTime(value string) time.Time {
	t, err := time.Parse("2006-01-02 15:04:05.000 -0700", value)
	if err != nil {
		panic(err)
	}
	return t
}
