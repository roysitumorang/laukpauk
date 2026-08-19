package models

import (
	"errors"
	"regexp"
	"strings"
	"uuid"

	"github.com/roysitumorang/laukpauk/helper"
)

type Topic string

func (q Topic) String() string {
	return string(q)
}

type TopicEvent string

type TopicMessage struct {
	Topic Topic      `json:"topic"`
	Event TopicEvent `json:"event"`
	ID    uuid.UUID  `json:"id"`
}

type TopicRequest struct {
	Topic    Topic
	Messages []TopicMessage
}

type Map map[string]any

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

type RecordStatus int

func (q RecordStatus) String() string {
	if q == RecordStatusDeleted {
		return "deleted"
	}
	return "active"
}

type DiscountType int

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

type SortingDirection string

func (q SortingDirection) String() string {
	return string(q)
}

func (q SortingDirection) Equal(v string) bool {
	return strings.EqualFold(q.String(), v)
}

type Sorting struct {
	Field     string
	Direction SortingDirection
	Label     string
}

func (q Sorting) String() string {
	return helper.Sprintf("%s:%s", q.Field, strings.ToLower(q.Direction.String()))
}

const (
	RecordStatusActive RecordStatus = iota
	RecordStatusDeleted
)

const (
	DiscountTypeFixedAmount DiscountType = iota
	DiscountTypePercentage
)

const SortingFieldID = "id"
const SortingFieldName = "name"

const SortingDirectionAsc SortingDirection = "ASC"
const SortingDirectionDesc SortingDirection = "DESC"

var ErrAccessDenied = errors.New("access denied")
var MapLimits = map[int64]struct{}{10: {}, 25: {}, 50: {}, 100: {}}
var Limits = []int64{10, 25, 50, 100}
var PhoneNumberRegex = regexp.MustCompile(`^\+[1-9]\d{1,14}$`)
var UsernameRegex = regexp.MustCompile("[^a-z0-9]+")
var SlugRegex = regexp.MustCompile("^[a-z0-9-]+$")
var SliceTopics = []string{
	TopicTextMessage.String(),
}
var MapTopics = map[Topic]struct{}{
	TopicTextMessage: {},
}
var DefaultSortBy = Sorting{
	Field:     SortingFieldID,
	Direction: SortingDirectionDesc,
}
var DefaultAscendingSortBy = Sorting{
	Field:     SortingFieldID,
	Direction: SortingDirectionAsc,
}
var DiscountTypes = []DiscountType{
	DiscountTypeFixedAmount,
	DiscountTypePercentage,
}
var MapDiscountTypes = map[DiscountType]struct{}{
	DiscountTypeFixedAmount: {},
	DiscountTypePercentage:  {},
}
