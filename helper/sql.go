package helper

import (
	"database/sql"
	"database/sql/driver"
	"encoding"
	"encoding/json"
	"fmt"
	"reflect"
	"time"
)

// SQLArgs converts query arguments to values database/sql and the MariaDB driver take: a uuid.UUID (which has no
// driver.Valuer) becomes its text form, a type with only a text form (MarshalText) becomes that text, and a slice,
// map or struct becomes JSON. Everything else (numbers, strings, booleans, times, Valuers, pointers to them) is left
// to database/sql.
func SQLArgs(args ...any) []any {
	out := make([]any, len(args))
	for i, a := range args {
		out[i] = convertArg(a)
	}
	return out
}

// SQLScan wraps scan destinations database/sql can't fill by itself: a *uuid.UUID / **uuid.UUID (from its text
// form), a slice, map or struct (from JSON), a type with only a text form, and times, which come back in the local
// time zone (MariaDB keeps them in UTC).
func SQLScan(dest ...any) []any {
	out := make([]any, len(dest))
	for i, d := range dest {
		out[i] = wrapDest(d)
	}
	return out
}

var (
	valuerType          = reflect.TypeFor[driver.Valuer]()
	scannerType         = reflect.TypeFor[sql.Scanner]()
	textMarshalerType   = reflect.TypeFor[encoding.TextMarshaler]()
	textUnmarshalerType = reflect.TypeFor[encoding.TextUnmarshaler]()
	timeType            = reflect.TypeFor[time.Time]()
	byteSliceType       = reflect.TypeFor[[]byte]()
)

// isUUIDType reports whether t is a 16-byte array with a text form, like uuid.UUID.
func isUUIDType(t reflect.Type) bool {
	return t.Kind() == reflect.Array && t.Len() == 16 && t.Elem().Kind() == reflect.Uint8 &&
		(t.Implements(textMarshalerType) || reflect.PointerTo(t).Implements(textUnmarshalerType))
}

// isTextType reports whether t (not a UUID, time or a type with its own database conversion) has a text form, like
// decimal types: it is stored as that text.
func isTextType(t reflect.Type) bool {
	if t == timeType || isUUIDType(t) || t.Implements(valuerType) || reflect.PointerTo(t).Implements(scannerType) {
		return false
	}
	if k := t.Kind(); k != reflect.Struct && k != reflect.Array {
		return false
	}
	return t.Implements(textMarshalerType) && reflect.PointerTo(t).Implements(textUnmarshalerType)
}

// isJSONType reports whether values of t are stored as JSON: slices (other than []byte), maps and structs other
// than time.Time, unless t handles its own conversion or has a text form.
func isJSONType(t reflect.Type) bool {
	if t == timeType || t == byteSliceType || isUUIDType(t) || isTextType(t) {
		return false
	}
	switch t.Kind() {
	case reflect.Slice, reflect.Map, reflect.Struct:
		return !t.Implements(valuerType) && !reflect.PointerTo(t).Implements(scannerType)
	}
	return false
}

// convertArg converts an argument: a UUID becomes its text form, a slice, map or struct becomes JSON, and anything
// else is left to database/sql (drivers, Valuers, numbers, strings, booleans, times and pointers to them).
func convertArg(a any) any {
	if a == nil {
		return nil
	}
	if _, ok := a.(driver.Valuer); ok {
		return a
	}
	v := reflect.ValueOf(a)
	for v.Kind() == reflect.Pointer {
		if v.IsNil() {
			return nil
		}
		if v.Type().Implements(valuerType) {
			return v.Interface()
		}
		v = v.Elem()
	}
	t := v.Type()
	switch {
	case isUUIDType(t):
		if m, ok := v.Interface().(encoding.TextMarshaler); ok {
			if b, err := m.MarshalText(); err == nil {
				return string(b)
			}
		}
		b := make([]byte, 16)
		reflect.Copy(reflect.ValueOf(b), v)
		return b
	case isTextType(t):
		if b, err := v.Interface().(encoding.TextMarshaler).MarshalText(); err == nil {
			return string(b)
		}
		return a
	case isJSONType(t):
		if (t.Kind() == reflect.Slice || t.Kind() == reflect.Map) && v.IsNil() {
			return "null"
		}
		b, err := json.Marshal(v.Interface())
		if err != nil {
			return a
		}
		return string(b)
	}
	return v.Interface()
}

// wrapDest returns a scanner for the destinations database/sql can't fill: UUIDs, JSON values and times (which come
// back in the local time zone, as pgx gave them).
func wrapDest(d any) any {
	switch p := d.(type) {
	case sql.Scanner:
		return d
	case *time.Time:
		return &timeDest{dst: p}
	case **time.Time:
		return &nullTimeDest{dst: p}
	}
	v := reflect.ValueOf(d)
	if v.Kind() != reflect.Pointer || v.IsNil() {
		return d
	}
	t := v.Type().Elem()
	switch {
	case isUUIDType(t), isTextType(t):
		return &textDest{dst: v}
	case isJSONType(t):
		return &jsonDest{dst: v}
	case t.Kind() == reflect.Pointer:
		inner := t.Elem()
		switch {
		case isUUIDType(inner), isTextType(inner):
			return &textDest{dst: v, nullable: true}
		case isJSONType(inner):
			return &jsonDest{dst: v, nullable: true}
		}
	}
	return d
}

func asBytes(src any) ([]byte, bool) {
	switch s := src.(type) {
	case []byte:
		return s, true
	case string:
		return []byte(s), true
	}
	return nil, false
}

// textDest fills a UUID or another type with a text form (or a pointer to one when nullable) from that text.
type textDest struct {
	dst      reflect.Value // pointer to the value, or to the pointer when nullable
	nullable bool
}

func (d *textDest) Scan(src any) error {
	target := d.dst.Elem()
	if src == nil {
		if d.nullable {
			target.SetZero()
			return nil
		}
		return fmt.Errorf("database: converting NULL to %s is unsupported", target.Type())
	}
	if d.nullable {
		target.Set(reflect.New(target.Type().Elem()))
		target = target.Elem()
	}
	b, ok := asBytes(src)
	if !ok {
		// numbers, for a decimal type read from a numeric column
		b = fmt.Append(b, src)
	}
	if len(b) == 16 && isUUIDType(target.Type()) {
		reflect.Copy(target, reflect.ValueOf(b))
		return nil
	}
	u, ok := target.Addr().Interface().(encoding.TextUnmarshaler)
	if !ok {
		return fmt.Errorf("database: %s has no text form", target.Type())
	}
	return u.UnmarshalText(b)
}

// jsonDest fills a slice, map or struct (or a pointer to one when nullable) from JSON.
type jsonDest struct {
	dst      reflect.Value
	nullable bool
}

func (d *jsonDest) Scan(src any) error {
	target := d.dst.Elem()
	if src == nil {
		target.SetZero()
		return nil
	}
	b, ok := asBytes(src)
	if !ok {
		return fmt.Errorf("database: converting %T to %s is unsupported", src, target.Type())
	}
	if d.nullable {
		target.Set(reflect.New(target.Type().Elem()))
		target = target.Elem()
	}
	return json.Unmarshal(b, target.Addr().Interface())
}

// timeLayout is how MariaDB writes DATETIME(6) when the driver returns it as text.
const timeLayout = "2006-01-02 15:04:05.999999"

func toTime(src any) (time.Time, error) {
	switch s := src.(type) {
	case time.Time:
		return s.In(time.Local), nil
	case []byte, string:
		b, _ := asBytes(s)
		t, err := time.ParseInLocation(timeLayout, string(b), time.UTC)
		if err != nil {
			return time.Time{}, err
		}
		return t.In(time.Local), nil
	}
	return time.Time{}, fmt.Errorf("database: converting %T to time.Time is unsupported", src)
}

type timeDest struct {
	dst *time.Time
}

func (d *timeDest) Scan(src any) error {
	if src == nil {
		return fmt.Errorf("database: converting NULL to time.Time is unsupported")
	}
	t, err := toTime(src)
	if err != nil {
		return err
	}
	*d.dst = t
	return nil
}

type nullTimeDest struct {
	dst **time.Time
}

func (d *nullTimeDest) Scan(src any) error {
	if src == nil {
		*d.dst = nil
		return nil
	}
	t, err := toTime(src)
	if err != nil {
		return err
	}
	*d.dst = &t
	return nil
}
