package repositories

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/roysitumorang/laukpauk/models"
)

// TestMariaDBSQLRoundTrip checks UUIDs, JSON columns and times through database/sql, MariaDB's RETURNING and the
// duplicate-key mapping on one connection (the temporary table lives on it).
func TestMariaDBSQLRoundTrip(t *testing.T) {
	db := testPool(t)
	ctx := context.Background()
	conn, err := db.Conn(ctx)
	if err != nil {
		t.Fatalf("conn: %v", err)
	}
	defer func() { _ = conn.Close() }()
	if _, err = conn.ExecContext(ctx, "CREATE TEMPORARY TABLE sql_test (id UUID NOT NULL PRIMARY KEY DEFAULT (UUID_v7()), `code` VARCHAR(255) NOT NULL, ref_id UUID, slots JSON NOT NULL, at DATETIME(6) NOT NULL, done_at DATETIME(6), flag BOOLEAN NOT NULL, CONSTRAINT sql_test_code_key UNIQUE (`code`))"); err != nil {
		t.Fatalf("create: %v", err)
	}
	ref := newTestUUID(t)
	at := time.Date(2026, 11, 28, 10, 47, 1, 123456000, time.Local)
	var id uuid.UUID
	var gotRef *uuid.UUID
	var slots models.Int64s
	var gotAt time.Time
	var doneAt *time.Time
	var flag bool
	var code string
	// a UUID, a JSON slice, a time and CONCAT
	if err = conn.QueryRowContext(ctx, "INSERT INTO sql_test (`code`, ref_id, slots, at, flag) VALUES (CONCAT(?, '-', ?), ?, ?, ?, ?) RETURNING id, `code`, ref_id, slots, at, done_at, flag",
		"A", "A", ref, models.Int64s{5, 10}, at, true,
	).Scan(&id, &code, &gotRef, &slots, &gotAt, &doneAt, &flag); err != nil {
		t.Fatalf("insert: %v", err)
	}
	if id == uuid.Nil || code != "A-A" || gotRef == nil || *gotRef != ref ||
		len(slots) != 2 || slots[1] != 10 || !gotAt.Equal(at) || doneAt != nil || !flag {
		t.Fatalf("got id=%v code=%q ref=%v slots=%v at=%v done=%v flag=%v", id, code, gotRef, slots, gotAt, doneAt, flag)
	}

	// UPDATE ... RETURNING (MariaDB 13)
	done := at.Add(time.Hour)
	if err = conn.QueryRowContext(ctx, `UPDATE sql_test SET done_at = ?, ref_id = ? WHERE id = ? RETURNING done_at, ref_id`,
		done, nil, id).Scan(&doneAt, &gotRef); err != nil {
		t.Fatalf("update returning: %v", err)
	}
	if doneAt == nil || !doneAt.Equal(done) || gotRef != nil {
		t.Fatalf("after update done=%v ref=%v", doneAt, gotRef)
	}

	// ON DUPLICATE KEY UPDATE ... RETURNING gives the updated row
	var n int64
	if err = conn.QueryRowContext(ctx, "INSERT INTO sql_test (`code`, slots, at, flag) VALUES (?, '[]', ?, FALSE) ON DUPLICATE KEY UPDATE slots = '[1,2,3]' RETURNING JSON_LENGTH(slots)", "A-A", at).Scan(&n); err != nil {
		t.Fatalf("upsert returning: %v", err)
	}
	if n != 3 {
		t.Fatalf("upsert returned %d slots; want the updated row's 3", n)
	}

	// duplicate entries name their key
	_, err = conn.ExecContext(ctx, "UPDATE sql_test SET `code` = 'A-A'")
	if err == nil {
		if _, err = conn.ExecContext(ctx, "INSERT INTO sql_test (`code`, slots, at, flag) VALUES ('A-A', '[]', ?, FALSE)", at); err == nil {
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
	if _, err = tx.ExecContext(ctx, `DELETE FROM sql_test WHERE id = ?`, id); err != nil {
		t.Fatalf("delete: %v", err)
	}
	if err = tx.Rollback(); err != nil {
		t.Fatalf("rollback: %v", err)
	}
	if err = tx.Commit(); !errors.Is(err, sql.ErrTxDone) {
		t.Fatalf("commit after rollback: err = %v; want sql.ErrTxDone", err)
	}
	if err = conn.QueryRowContext(ctx, `SELECT id FROM sql_test WHERE id = ?`, id).Scan(&id); err != nil {
		t.Fatalf("row gone after rollback: %v", err)
	}
	if err = conn.QueryRowContext(ctx, "SELECT id FROM sql_test WHERE `code` = ?", "none").Scan(&id); !errors.Is(err, sql.ErrNoRows) {
		t.Fatalf("no rows: err = %v", err)
	}
}
