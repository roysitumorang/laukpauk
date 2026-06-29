package models

import (
	"regexp"
	"strings"

	"github.com/roysitumorang/laukpauk/helper"
)

type (
	Map map[string]any

	RecordStatus     int
	DiscountType     int
	SortingDirection string

	Sorting struct {
		Field     string
		Direction SortingDirection
		Label     string
	}
)

const (
	RecordStatusActive RecordStatus = iota
	RecordStatusDeleted
)

const (
	DiscountTypeFlat DiscountType = iota
	DiscountTypePercentage
)

const (
	SortingFieldID   = "id"
	SortingFieldName = "name"

	SortingDirectionAsc  SortingDirection = "ASC"
	SortingDirectionDesc SortingDirection = "DESC"
)

var (
	MapLimits        = map[int64]struct{}{10: {}, 25: {}, 50: {}, 100: {}}
	Limits           = []int64{10, 25, 50, 100}
	PhoneNumberRegex = regexp.MustCompile(`^\+[1-9]\d{1,14}$`)
	UsernameRegex    = regexp.MustCompile("[^a-z0-9]+")
	SlugRegex        = regexp.MustCompile("^[a-z0-9-]+$")
	SliceTopics      = []string{}
	MapTopics        = map[string]struct{}{}
	DefaultSortBy    = Sorting{
		Field:     SortingFieldID,
		Direction: SortingDirectionDesc,
	}
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

func (q *Map) Contains(key string) bool {
	_, ok := (*q)[key]
	return ok
}

func (q *Map) Clear() {
	clear(*q)
}

func (q RecordStatus) String() string {
	if q == RecordStatusDeleted {
		return "deleted"
	}
	return "active"
}

func (q Sorting) String() string {
	return helper.Sprintf("%s:%s", q.Field, strings.ToLower(q.Direction.String()))
}

func (q SortingDirection) String() string {
	return string(q)
}

func (q SortingDirection) Equal(v string) bool {
	return strings.EqualFold(q.String(), v)
}
