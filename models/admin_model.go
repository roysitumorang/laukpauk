package models

import (
	"errors"
	"net/url"
	"strings"
	"uuid"

	"github.com/roysitumorang/laukpauk/helper"
)

var (
	ErrAdminNotFound = errors.New("admin not found")
)

const (
	AdminTypeAdmin AdminType = iota
	AdminTypeSuperAdmin
)

type (
	AdminType int

	Admin struct {
		user
		AdminType AdminType `json:"-" form:"-"`
	}

	AdminFilter struct {
		userFilter
	}
)

func NewAdminFilter() *AdminFilter {
	return new(AdminFilter{
		MapSortings: NewMap(),
		UrlValues:   url.Values{},
	})
}

func (q *AdminFilter) WithAdminIDs(adminIDs ...uuid.UUID) *AdminFilter {
	q.UserIDs = adminIDs
	return q
}

func (q *AdminFilter) WithExcludedAdminIDs(excludedAdminIDs ...uuid.UUID) *AdminFilter {
	q.ExcludedUserIDs = excludedAdminIDs
	return q
}

func (q *AdminFilter) WithUserStatuses(userStatuses ...UserStatus) *AdminFilter {
	q.UserStatuses = userStatuses
	return q
}

func (q *AdminFilter) WithLogin(login string) *AdminFilter {
	q.Login = login
	return q
}

func (q *AdminFilter) WithKeyword(keyword string) *AdminFilter {
	q.Keyword = keyword
	return q
}

func (q *AdminFilter) WithUsername(username string) *AdminFilter {
	q.Username = username
	return q
}

func (q *AdminFilter) WithEmail(email string) *AdminFilter {
	q.Email = email
	return q
}

func (q *AdminFilter) WithMobilePhone(mobilePhone string) *AdminFilter {
	q.MobilePhone = mobilePhone
	return q
}

func (q *AdminFilter) WithPasswordResetToken(passwordResetToken string) *AdminFilter {
	q.PasswordResetToken = passwordResetToken
	return q
}

func (q *AdminFilter) WithPaginationURL(paginationURL string) *AdminFilter {
	q.PaginationURL = paginationURL
	return q
}

func (q *AdminFilter) WithSortings(sortings ...Sorting) *AdminFilter {
	q.Sortings = sortings
	q.MapSortings.Clear()
	for _, sorting := range sortings {
		q.MapSortings.Push(strings.ToLower(helper.Sprintf("%s:%s", sorting.Field, sorting.Direction)), struct{}{})
	}
	return q
}

func (q *AdminFilter) WithLimit(limit int64) *AdminFilter {
	q.Limit = limit
	return q
}

func (q *AdminFilter) WithPage(page int64) *AdminFilter {
	q.Page = page
	return q
}

func (q *AdminFilter) WithUrlValues(urlValues url.Values) *AdminFilter {
	q.UrlValues = urlValues
	return q
}
