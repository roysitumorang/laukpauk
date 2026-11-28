package database

import (
	"context"
	"os"
	"testing"
	"time"
	"uuid"
)

// TestMariaDB runs against a real MariaDB when TEST_DATABASE_URL is set (as in CI).
func TestMariaDB(t *testing.T) {
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("TEST_DATABASE_URL is not set")
	}
	ctx := context.Background()
	db, err := Open(ctx, dsn, 1) // one connection: the temporary table lives on it
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	defer db.Close()
	if _, err = db.Exec(ctx, `CREATE TEMPORARY TABLE adapter_test (
		id UUID NOT NULL PRIMARY KEY DEFAULT (UUID_v7())
		, "code" VARCHAR(255) NOT NULL
		, ref_id UUID
		, slots JSON NOT NULL
		, at DATETIME(6) NOT NULL
		, done_at DATETIME(6)
		, flag BOOLEAN NOT NULL
		, CONSTRAINT adapter_test_code_key UNIQUE ("code")
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
	// $n placeholders out of order and repeated, a UUID, a JSON slice, a time and || concatenation
	if err = db.QueryRow(ctx, `INSERT INTO adapter_test ("code", ref_id, slots, at, flag)
		VALUES ($2 || '-' || $2, $1, $3, $4, $5)
		RETURNING id, "code", ref_id, slots, at, done_at, flag`,
		ref, "A", []int64{5, 10}, at, true,
	).Scan(&id, &code, &gotRef, &slots, &gotAt, &doneAt, &flag); err != nil {
		t.Fatalf("insert: %v", err)
	}
	if id.Compare(uuid.Nil()) == 0 || code != "A-A" || gotRef == nil || gotRef.Compare(ref) != 0 ||
		len(slots) != 2 || slots[1] != 10 || !gotAt.Equal(at) || gotAt.Location() != time.Local || doneAt != nil || !flag {
		t.Fatalf("got id=%v code=%q ref=%v slots=%v at=%v done=%v flag=%v", id, code, gotRef, slots, gotAt, doneAt, flag)
	}

	// UPDATE ... RETURNING (MariaDB 13)
	done := at.Add(time.Hour)
	if err = db.QueryRow(ctx, `UPDATE adapter_test SET done_at = $1, ref_id = $2 WHERE id = $3 RETURNING done_at, ref_id`,
		done, nil, id).Scan(&doneAt, &gotRef); err != nil {
		t.Fatalf("update returning: %v", err)
	}
	if doneAt == nil || !doneAt.Equal(done) || gotRef != nil {
		t.Fatalf("after update done=%v ref=%v", doneAt, gotRef)
	}

	// ON DUPLICATE KEY UPDATE ... RETURNING gives the updated row
	var n int64
	if err = db.QueryRow(ctx, `INSERT INTO adapter_test ("code", slots, at, flag) VALUES ($1, '[]', $2, FALSE)
		ON DUPLICATE KEY UPDATE slots = '[1,2,3]'
		RETURNING JSON_LENGTH(slots)`, "A-A", at).Scan(&n); err != nil {
		t.Fatalf("upsert returning: %v", err)
	}
	if n != 3 {
		t.Fatalf("upsert returned %d slots; want the updated row's 3", n)
	}

	// duplicate entries name their key
	_, err = db.Exec(ctx, `UPDATE adapter_test SET "code" = 'A-A'`)
	if err == nil {
		if _, err = db.Exec(ctx, `INSERT INTO adapter_test ("code", slots, at, flag) VALUES ('A-A', '[]', $1, FALSE)`, at); err == nil {
			t.Fatal("want a duplicate entry error")
		}
	}
	if !IsUniqueViolation(err, "adapter_test_code_key") {
		t.Fatalf("IsUniqueViolation(%v) = false", err)
	}

	// transactions
	tx, err := db.BeginTx(ctx, TxOptions{IsoLevel: Serializable})
	if err != nil {
		t.Fatalf("begin: %v", err)
	}
	if _, err = tx.Exec(ctx, `DELETE FROM adapter_test WHERE id = $1`, id); err != nil {
		t.Fatalf("delete: %v", err)
	}
	if err = tx.Rollback(ctx); err != nil {
		t.Fatalf("rollback: %v", err)
	}
	if err = tx.Commit(ctx); err != ErrTxClosed {
		t.Fatalf("commit after rollback: err = %v; want ErrTxClosed", err)
	}
	if err = db.QueryRow(ctx, `SELECT id FROM adapter_test WHERE id = $1`, id).Scan(&id); err != nil {
		t.Fatalf("row gone after rollback: %v", err)
	}
	if err = db.QueryRow(ctx, `SELECT id FROM adapter_test WHERE "code" = $1`, "none").Scan(&id); !IsNoRows(err) {
		t.Fatalf("no rows: err = %v", err)
	}
}
