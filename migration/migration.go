package migration

import (
	"context"
	"errors"
	"fmt"
	"os"
	"sort"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/roysitumorang/laukpauk/helper"
	"go.uber.org/zap"
)

type (
	Migration struct {
		tx pgx.Tx
	}
)

var (
	Migrations = map[int64]func(ctx context.Context, tx pgx.Tx) error{}
)

func NewMigration(tx pgx.Tx) *Migration {
	return &Migration{
		tx: tx,
	}
}

func (m *Migration) Migrate(ctx context.Context) {
	ctxt := "Migration-Migrate"
	var exists bool
	if err := m.tx.QueryRow(
		ctx,
		`SELECT EXISTS(
			SELECT 1
			FROM information_schema.tables
			WHERE table_name = 'migrations'
		)`,
	).Scan(&exists); err != nil {
		_ = m.tx.Rollback(ctx)
		helper.Capture(ctx, zap.FatalLevel, err, ctxt, "ErrScan")
		return
	}
	if !exists {
		if _, err := m.tx.Exec(
			ctx,
			`CREATE TABLE migrations (
				"version" bigint NOT NULL PRIMARY KEY
			)`,
		); err != nil {
			_ = m.tx.Rollback(ctx)
			helper.Capture(ctx, zap.FatalLevel, err, ctxt, "ErrExec")
			return
		}
	}
	rows, err := m.tx.Query(ctx, `SELECT "version" FROM "migrations" ORDER BY "version"`)
	if errors.Is(err, pgx.ErrNoRows) {
		err = nil
	}
	if err != nil {
		_ = m.tx.Rollback(ctx)
		helper.Capture(ctx, zap.FatalLevel, err, ctxt, "ErrQuery")
		return
	}
	defer rows.Close()
	mapVersions := map[int64]int{}
	for rows.Next() {
		var version int64
		if err = rows.Scan(&version); err != nil {
			_ = m.tx.Rollback(ctx)
			helper.Capture(ctx, zap.FatalLevel, err, ctxt, "ErrScan")
			return
		}
		mapVersions[version] = 1
	}
	sortedVersions := make([]int64, len(Migrations))
	var i int
	for version := range Migrations {
		sortedVersions[i] = version
		i++
	}
	if len(sortedVersions) > 0 {
		sort.Slice(
			sortedVersions,
			func(i, j int) bool {
				return sortedVersions[i] < sortedVersions[j]
			},
		)
	}
	for _, version := range sortedVersions {
		if _, ok := mapVersions[version]; ok {
			continue
		}
		function, ok := Migrations[version]
		if !ok {
			_ = m.tx.Rollback(ctx)
			err = fmt.Errorf("migration function for version %d not found", version)
			helper.Capture(ctx, zap.FatalLevel, err, ctxt, "ErrOK")
			return
		}
		if err = function(ctx, m.tx); err != nil {
			_ = m.tx.Rollback(ctx)
			helper.Capture(ctx, zap.FatalLevel, err, ctxt, "ErrFunction")
			return
		}
		if _, err = m.tx.Exec(ctx, `INSERT INTO "migrations" ("version") VALUES ($1)`, version); err != nil {
			_ = m.tx.Rollback(ctx)
			helper.Capture(ctx, zap.FatalLevel, err, ctxt, "ErrExec")
			return
		}
	}
	_ = m.tx.Commit(ctx)
}

func (m *Migration) CreateMigrationFile(_ context.Context) error {
	now := time.Now().UTC().UnixNano()
	filepath := fmt.Sprintf("./migration/%d.go", now)
	content := fmt.Sprintf(
		`package migration

import (
	"context"

	"github.com/jackc/pgx/v5"
)

func init() {
	Migrations[%d] = func(ctx context.Context, tx pgx.Tx) (err error) {
		return
	}
}`,
		now,
	)
	return os.WriteFile(
		filepath,
		helper.String2ByteSlice(content),
		0600,
	)
}
