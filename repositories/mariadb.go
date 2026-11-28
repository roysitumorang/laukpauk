package repositories

import (
	"errors"
	"strings"

	"github.com/go-sql-driver/mysql"
)

// MariaDB's error number for a duplicate entry on a unique key; its message ends with "for key '<name>'".
const (
	errDuplicateEntry   = 1062
	duplicateKeyMessage = " for key '"
)

// uniqueViolationKey returns the name of the unique key a duplicate-entry error is about; ok is false for other
// errors.
func uniqueViolationKey(err error) (key string, ok bool) {
	var e *mysql.MySQLError
	if !errors.As(err, &e) || e.Number != errDuplicateEntry {
		return "", false
	}
	i := strings.LastIndex(e.Message, duplicateKeyMessage)
	if i < 0 {
		return "", true
	}
	key = strings.TrimSuffix(e.Message[i+len(duplicateKeyMessage):], "'")
	// MariaDB may prefix the key with the table name ("table.key")
	if j := strings.LastIndexByte(key, '.'); j >= 0 {
		key = key[j+1:]
	}
	return key, true
}

// isUniqueViolation reports whether err is a duplicate entry on the unique key named key.
func isUniqueViolation(err error, key string) bool {
	name, ok := uniqueViolationKey(err)
	return ok && name == key
}
