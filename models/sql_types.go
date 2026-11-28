package models

import (
	"database/sql/driver"
	"fmt"

	"github.com/goccy/go-json"
)

// Int64s is a list of numbers kept in a JSON column.
type Int64s []int64

// Value implements driver.Valuer: the list as JSON, [] when nil.
func (q Int64s) Value() (driver.Value, error) {
	if q == nil {
		return "[]", nil
	}
	b, err := json.Marshal([]int64(q))
	return string(b), err
}

// Scan implements sql.Scanner from a JSON column; NULL gives nil.
func (q *Int64s) Scan(src any) error {
	return scanJSON(src, (*[]int64)(q))
}

// StringMap is a string map kept in a JSON column.
type StringMap map[string]string

// Value implements driver.Valuer: the map as JSON, {} when nil.
func (q StringMap) Value() (driver.Value, error) {
	if q == nil {
		return "{}", nil
	}
	b, err := json.Marshal(map[string]string(q))
	return string(b), err
}

// Scan implements sql.Scanner from a JSON column; NULL gives nil.
func (q *StringMap) Scan(src any) error {
	return scanJSON(src, (*map[string]string)(q))
}

func scanJSON(src, dst any) error {
	switch s := src.(type) {
	case nil:
		return json.Unmarshal([]byte("null"), dst)
	case []byte:
		return json.Unmarshal(s, dst)
	case string:
		return json.Unmarshal([]byte(s), dst)
	}
	return fmt.Errorf("models: cannot scan %T from a JSON column", src)
}
