package models

import (
	"regexp"
	"strings"
	"uuid"

	"github.com/roysitumorang/laukpauk/helper"
)

type (
	Topic      string
	TopicEvent string

	TopicMessage struct {
		Topic Topic      `json:"topic"`
		Event TopicEvent `json:"event"`
		ID    uuid.UUID  `json:"id"`
	}

	TopicRequest struct {
		Topic    Topic
		Messages []TopicMessage
	}

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
	DiscountTypeFixedAmount DiscountType = iota
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
	SliceTopics      = []string{
		TopicSms.String(),
	}
	MapTopics = map[Topic]struct{}{
		TopicSms: {},
	}
	DefaultSortBy = Sorting{
		Field:     SortingFieldID,
		Direction: SortingDirectionDesc,
	}
	DefaultAscendingSortBy = Sorting{
		Field:     SortingFieldID,
		Direction: SortingDirectionAsc,
	}
	DiscountTypes    = []DiscountType{DiscountTypeFixedAmount, DiscountTypePercentage}
	MapDiscountTypes = map[DiscountType]struct{}{
		DiscountTypeFixedAmount: {},
		DiscountTypePercentage:  {},
	}
)

func (q Topic) String() string {
	return string(q)
}

func NewMap() Map {
	return Map{}
}

func (q Map) Push(args ...any) Map {
	n := len(args)
	for i := 0; i < len(args); i += 2 {
		if i+1 >= n {
			break
		}
		if key, ok := args[i].(string); ok {
			q[key] = args[i+1]
		}
	}
	return q
}

func (q Map) Delete(key string) Map {
	delete(q, key)
	return q
}

func (q Map) Size() int {
	return len(q)
}

func (q Map) Contains(key string) bool {
	_, ok := q[key]
	return ok
}

func (q Map) Clear() {
	clear(q)
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

func (q DiscountType) Int() int {
	return int(q)
}

func (q DiscountType) String() string {
	if q == DiscountTypePercentage {
		return "persentase"
	}
	return "cash"
}

func (q DiscountType) Equal(v DiscountType) bool {
	return q.Int() == v.Int()
}
