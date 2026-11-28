package database

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"time"

	"github.com/go-sql-driver/mysql"
)

// sqlMode keeps the PostgreSQL habits the queries rely on: "double quoted" identifiers and || for concatenation, and
// makes invalid or truncated values errors instead of warnings.
const sqlMode = "'STRICT_ALL_TABLES,ANSI_QUOTES,PIPES_AS_CONCAT,NO_ZERO_IN_DATE,NO_ZERO_DATE,ERROR_FOR_DIVISION_BY_ZERO,NO_ENGINE_SUBSTITUTION'"

// Config turns a go-sql-driver/mysql DSN (user:password@tcp(host:3306)/dbname?...) into the driver config the
// backend needs: times are parsed and kept in UTC on the server, with the SQL mode above.
func Config(dsn string) (*mysql.Config, error) {
	cfg, err := mysql.ParseDSN(dsn)
	if err != nil {
		return nil, err
	}
	cfg.ParseTime = true
	cfg.Loc = time.UTC
	if cfg.Params == nil {
		cfg.Params = map[string]string{}
	}
	cfg.Params["time_zone"] = "'+00:00'"
	cfg.Params["sql_mode"] = sqlMode
	return cfg, nil
}

// Open connects to MariaDB with at most maxConns open connections and checks the connection.
func Open(ctx context.Context, dsn string, maxConns int) (*Pool, error) {
	cfg, err := Config(dsn)
	if err != nil {
		return nil, err
	}
	connector, err := mysql.NewConnector(cfg)
	if err != nil {
		return nil, err
	}
	db := sql.OpenDB(connector)
	db.SetMaxOpenConns(maxConns)
	db.SetMaxIdleConns(maxConns)
	db.SetConnMaxIdleTime(5 * time.Minute)
	pool := NewPool(db)
	if err = pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, err
	}
	return pool, nil
}

// MariaDB error numbers.
const (
	errDuplicateEntry   = 1062
	errLockDeadlock     = 1213
	errLockWaitTimeout  = 1205
	errNoReferencedRow  = 1452
	errRowIsReferenced  = 1451
	duplicateKeyMessage = " for key '"
)

// duplicateKey returns the name of the unique key a duplicate-entry error is about.
func duplicateKey(err error) (string, bool) {
	var e *mysql.MySQLError
	if !errors.As(err, &e) || e.Number != errDuplicateEntry {
		return "", false
	}
	i := strings.LastIndex(e.Message, duplicateKeyMessage)
	if i < 0 {
		return "", true
	}
	key := strings.TrimSuffix(e.Message[i+len(duplicateKeyMessage):], "'")
	// MariaDB may prefix the key with the table name ("table.key")
	if j := strings.LastIndexByte(key, '.'); j >= 0 {
		key = key[j+1:]
	}
	return key, true
}

// UniqueViolationKey returns the name of the unique key a duplicate-entry error is about; ok is false for other
// errors.
func UniqueViolationKey(err error) (key string, ok bool) {
	return duplicateKey(err)
}

// IsUniqueViolation reports whether err is a duplicate entry on the unique key named key (any unique key when key is
// empty).
func IsUniqueViolation(err error, key string) bool {
	name, ok := duplicateKey(err)
	return ok && (key == "" || name == key)
}

// IsSerializationFailure reports whether the transaction failed on a deadlock or lock wait timeout and can be retried.
func IsSerializationFailure(err error) bool {
	var e *mysql.MySQLError
	return errors.As(err, &e) && (e.Number == errLockDeadlock || e.Number == errLockWaitTimeout)
}

// IsForeignKeyViolation reports whether err is about a missing or still referenced row.
func IsForeignKeyViolation(err error) bool {
	var e *mysql.MySQLError
	return errors.As(err, &e) && (e.Number == errNoReferencedRow || e.Number == errRowIsReferenced)
}
