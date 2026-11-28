package repositories

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"
	"uuid"

	"github.com/roysitumorang/laukpauk/helper"
)

// TestMariaDBSQLRoundTrip checks the value conversions of helper.SQLArgs / helper.SQLScan, MariaDB's RETURNING and
// the duplicate-key mapping on one connection (the temporary table lives on it).
func TestMariaDBSQLRoundTrip(t *testing.T) {
	db := testPool(t)
	ctx := context.Background()
	conn, err := db.Conn(ctx)
	if err != nil {
		t.Fatalf("conn: %v", err)
	}
	defer func() { _ = conn.Close() }()
	if _, err = conn.ExecContext(ctx, `CREATE TEMPORARY TABLE sql_test (
		id UUID NOT NULL PRIMARY KEY DEFAULT (UUID_v7())
		, "code" VARCHAR(255) NOT NULL
		, ref_id UUID
		, slots JSON NOT NULL
		, at DATETIME(6) NOT NULL
		, done_at DATETIME(6)
		, flag BOOLEAN NOT NULL
		, CONSTRAINT sql_test_code_key UNIQUE ("code")
	)`); err != nil {
		t.Fatalf("create: %v", err)
	}
	ref := uuid.NewV7()
	at := time.Date(2026, 11, 28, 10, 47, 1, 123456000, time.Local)
	var id uuid.UUID
	var gotRef *uuid.UUID
	var slots []int64
	var gotAt time.Time
	var doneAt *time.Time
	var flag bool
	var code string
	// a UUID, a JSON slice, a time and || concatenation
	if err = conn.QueryRowContext(ctx, `INSERT INTO sql_test ("code", ref_id, slots, at, flag)
		VALUES (? || '-' || ?, ?, ?, ?, ?)
		RETURNING id, "code", ref_id, slots, at, done_at, flag`,
		helper.SQLArgs("A", "A", ref, []int64{5, 10}, at, true)...,
	).Scan(helper.SQLScan(&id, &code, &gotRef, &slots, &gotAt, &doneAt, &flag)...); err != nil {
		t.Fatalf("insert: %v", err)
	}
	if id.Compare(uuid.Nil()) == 0 || code != "A-A" || gotRef == nil || gotRef.Compare(ref) != 0 ||
		len(slots) != 2 || slots[1] != 10 || !gotAt.Equal(at) || gotAt.Location() != time.Local || doneAt != nil || !flag {
		t.Fatalf("got id=%v code=%q ref=%v slots=%v at=%v done=%v flag=%v", id, code, gotRef, slots, gotAt, doneAt, flag)
	}

	// UPDATE ... RETURNING (MariaDB 13)
	done := at.Add(time.Hour)
	if err = conn.QueryRowContext(ctx, `UPDATE sql_test SET done_at = ?, ref_id = ? WHERE id = ? RETURNING done_at, ref_id`,
		helper.SQLArgs(done, nil, id)...).Scan(helper.SQLScan(&doneAt, &gotRef)...); err != nil {
		t.Fatalf("update returning: %v", err)
	}
	if doneAt == nil || !doneAt.Equal(done) || gotRef != nil {
		t.Fatalf("after update done=%v ref=%v", doneAt, gotRef)
	}

	// ON DUPLICATE KEY UPDATE ... RETURNING gives the updated row
	var n int64
	if err = conn.QueryRowContext(ctx, `INSERT INTO sql_test ("code", slots, at, flag) VALUES (?, '[]', ?, FALSE)
		ON DUPLICATE KEY UPDATE slots = '[1,2,3]'
		RETURNING JSON_LENGTH(slots)`, "A-A", at).Scan(helper.SQLScan(&n)...); err != nil {
		t.Fatalf("upsert returning: %v", err)
	}
	if n != 3 {
		t.Fatalf("upsert returned %d slots; want the updated row's 3", n)
	}

	// duplicate entries name their key
	_, err = conn.ExecContext(ctx, `UPDATE sql_test SET "code" = 'A-A'`)
	if err == nil {
		if _, err = conn.ExecContext(ctx, `INSERT INTO sql_test ("code", slots, at, flag) VALUES ('A-A', '[]', ?, FALSE)`, at); err == nil {
			t.Fatal("want a duplicate entry error")
		}
	}
	if !isUniqueViolation(err, "sql_test_code_key") {
		t.Fatalf("isUniqueViolation(%v) = false", err)
	}

	// transactions
	tx, err := conn.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		t.Fatalf("begin: %v", err)
	}
	if _, err = tx.ExecContext(ctx, `DELETE FROM sql_test WHERE id = ?`, helper.SQLArgs(id)...); err != nil {
		t.Fatalf("delete: %v", err)
	}
	if err = tx.Rollback(); err != nil {
		t.Fatalf("rollback: %v", err)
	}
	if err = tx.Commit(); !errors.Is(err, sql.ErrTxDone) {
		t.Fatalf("commit after rollback: err = %v; want sql.ErrTxDone", err)
	}
	if err = conn.QueryRowContext(ctx, `SELECT id FROM sql_test WHERE id = ?`, helper.SQLArgs(id)...).Scan(helper.SQLScan(&id)...); err != nil {
		t.Fatalf("row gone after rollback: %v", err)
	}
	if err = conn.QueryRowContext(ctx, `SELECT id FROM sql_test WHERE "code" = ?`, "none").Scan(helper.SQLScan(&id)...); !errors.Is(err, sql.ErrNoRows) {
		t.Fatalf("no rows: err = %v", err)
	}
}
