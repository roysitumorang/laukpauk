package models

import (
	"errors"
	"strings"
	"uuid"

	"github.com/coregx/coregex"
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

var SliceTopics = []string{
	TopicTextMessage.String(),
}
var MapTopics = map[Topic]struct{}{
	TopicTextMessage: {},
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

type RecordStatus uint8

const (
	RecordStatusActive RecordStatus = iota
	RecordStatusDeleted
)

func (q RecordStatus) String() string {
	if q == RecordStatusDeleted {
		return "deleted"
	}
	return "active"
}

type DiscountType uint8

const (
	DiscountTypeFixedAmount DiscountType = iota
	DiscountTypePercentage
)

func (q DiscountType) Int() uint8 {
	return uint8(q)
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

var DiscountTypes = []DiscountType{
	DiscountTypeFixedAmount,
	DiscountTypePercentage,
}
var MapDiscountTypes = map[DiscountType]struct{}{
	DiscountTypeFixedAmount: {},
	DiscountTypePercentage:  {},
}

type SortingDirection string

const SortingDirectionAsc SortingDirection = "ASC"
const SortingDirectionDesc SortingDirection = "DESC"

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

const SortingFieldID = "id"
const SortingFieldName = "name"

var DefaultSortBy = Sorting{
	Field:     SortingFieldID,
	Direction: SortingDirectionDesc,
}
var DefaultAscendingSortBy = Sorting{
	Field:     SortingFieldID,
	Direction: SortingDirectionAsc,
}

var ErrAccessDenied = errors.New("access denied")
var MapLimits = map[int64]struct{}{10: {}, 25: {}, 50: {}, 100: {}}
var Limits = []int64{10, 25, 50, 100}
var PhoneNumberRegex = coregex.MustCompile(`^\+[1-9]\d{1,14}$`)
var UsernameRegex = coregex.MustCompile("[^a-z0-9]+")
var SlugRegex = coregex.MustCompile("^[a-z0-9-]+$")
