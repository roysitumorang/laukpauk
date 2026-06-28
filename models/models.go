package models

import (
	"regexp"
)

type (
	Map map[string]any

	RecordStatus int
	DiscountType int
)

const (
	RecordStatusActive RecordStatus = iota
	RecordStatusDeleted
)

const (
	DiscountTypeFlat DiscountType = iota
	DiscountTypePercentage
)

var (
	MapLimits        = map[int64]int{10: 1, 25: 1, 50: 1, 100: 1}
	Limits           = []int64{10, 25, 50, 100}
	PhoneNumberRegex = regexp.MustCompile(`^\+[1-9]\d{1,14}$`)
	UsernameRegex    = regexp.MustCompile("[^a-z0-9]+")
	SlugRegex        = regexp.MustCompile("^[a-z0-9-]+$")
	SliceTopics      = []string{}
	MapTopics        = map[string]struct{}{}
)

func NewMap() *Map {
	return new(Map{})
}

func (q *Map) Push(args ...any) *Map {
	n := len(args)
	for i := 0; i < len(args); i += 2 {
		if i+1 >= n {
			break
		}
		if key, ok := args[i].(string); ok {
			(*q)[key] = args[i+1]
		}
	}
	return q
}

func (q *Map) Delete(key string) *Map {
	delete(*q, key)
	return q
}

func (q *Map) Size() int {
	return len(*q)
}
