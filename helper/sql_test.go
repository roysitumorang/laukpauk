package helper

import (
	"encoding/hex"
	"errors"
	"reflect"
	"strings"
	"testing"
	"time"
)

// testUUID stands in for uuid.UUID: a 16-byte array with a text form.
type testUUID [16]byte

func (u testUUID) MarshalText() ([]byte, error) {
	h := hex.EncodeToString(u[:])
	return []byte(h[0:8] + "-" + h[8:12] + "-" + h[12:16] + "-" + h[16:20] + "-" + h[20:]), nil
}

func (u *testUUID) UnmarshalText(b []byte) error {
	s := strings.ReplaceAll(string(b), "-", "")
	if len(s) != 32 {
		return errors.New("bad uuid")
	}
	_, err := hex.Decode(u[:], []byte(s))
	return err
}

// testDecimal stands in for a decimal type with a text form but no database conversion of its own.
type testDecimal struct{ s string }

func (d testDecimal) MarshalText() ([]byte, error) { return []byte(d.s), nil }

func (d *testDecimal) UnmarshalText(b []byte) error {
	d.s = string(b)
	return nil
}

const testUUIDText = "0193b3a4-7c1e-7cc2-9f3a-0123456789ab"

func mustUUID(t *testing.T) testUUID {
	t.Helper()
	var u testUUID
	if err := u.UnmarshalText([]byte(testUUIDText)); err != nil {
		t.Fatal(err)
	}
	return u
}

func TestConvertArg(t *testing.T) {
	u := mustUUID(t)
	var nilUUID *testUUID
	now := time.Now()
	cases := []struct {
		in   any
		want any
	}{
		{nil, nil},
		{u, testUUIDText},
		{&u, testUUIDText},
		{nilUUID, nil},
		{[]int64{1, 2, 3}, "[1,2,3]"},
		{[]int64(nil), "null"},
		{map[string]any{"a": 1}, `{"a":1}`},
		{"x", "x"},
		{int64(7), int64(7)},
		{true, true},
		{now, now},
		{[]byte("raw"), []byte("raw")},
		{testDecimal{"3.141"}, "3.141"},
	}
	for _, c := range cases {
		if got := convertArg(c.in); !reflect.DeepEqual(got, c.want) {
			t.Errorf("convertArg(%#v) = %#v; want %#v", c.in, got, c.want)
		}
	}
}

type scanner interface{ Scan(any) error }

func TestWrapDest(t *testing.T) {
	var u testUUID
	if err := wrapDest(&u).(scanner).Scan([]byte(testUUIDText)); err != nil || u != mustUUID(t) {
		t.Fatalf("uuid: %v %v", u, err)
	}
	if err := wrapDest(&u).(scanner).Scan(nil); err == nil {
		t.Fatal("uuid: want an error for NULL")
	}

	nu := &testUUID{}
	if err := wrapDest(&nu).(scanner).Scan(nil); err != nil || nu != nil {
		t.Fatalf("*uuid NULL: %v %v", nu, err)
	}
	if err := wrapDest(&nu).(scanner).Scan(testUUIDText); err != nil || nu == nil || *nu != mustUUID(t) {
		t.Fatalf("*uuid: %v %v", nu, err)
	}

	var slots []int64
	if err := wrapDest(&slots).(scanner).Scan([]byte("[10,20]")); err != nil || !reflect.DeepEqual(slots, []int64{10, 20}) {
		t.Fatalf("json: %v %v", slots, err)
	}
	var params map[string]any
	if err := wrapDest(&params).(scanner).Scan(nil); err != nil || params != nil {
		t.Fatalf("json NULL: %v %v", params, err)
	}

	utc := time.Date(2026, 11, 28, 3, 4, 5, 0, time.UTC)
	var ts time.Time
	if err := wrapDest(&ts).(scanner).Scan(utc); err != nil || !ts.Equal(utc) || ts.Location() != time.Local {
		t.Fatalf("time: %v %v", ts, err)
	}
	var pts *time.Time
	if err := wrapDest(&pts).(scanner).Scan(nil); err != nil || pts != nil {
		t.Fatalf("*time NULL: %v %v", pts, err)
	}
	if err := wrapDest(&pts).(scanner).Scan([]byte("2026-11-28 03:04:05.000000")); err != nil || pts == nil || !pts.Equal(utc) {
		t.Fatalf("*time text: %v %v", pts, err)
	}

	var d testDecimal
	if err := wrapDest(&d).(scanner).Scan([]byte("98.765")); err != nil || d.s != "98.765" {
		t.Fatalf("text type: %v %v", d, err)
	}
	var pd *testDecimal
	if err := wrapDest(&pd).(scanner).Scan(int64(12)); err != nil || pd == nil || pd.s != "12" {
		t.Fatalf("*text type: %v %v", pd, err)
	}

	var n int64
	if got := wrapDest(&n); got != any(&n) {
		t.Fatal("int64: want the destination as is")
	}
	var s *string
	if got := wrapDest(&s); got != any(&s) {
		t.Fatal("*string: want the destination as is")
	}
}
