package models

import (
	"errors"
	"net/url"
	"strings"
	"uuid"

	"github.com/roysitumorang/laukpauk/helper"
)

var (
	ErrBuyerNotFound = errors.New("buyer not found")
)

type (
	Buyer struct {
		user
		ReferralCode              string     `json:"-" form:"-"`
		Address                   string     `json:"-" form:"-"`
		ReferrerID                *uuid.UUID `json:"-" form:"-"`
		ReferralPoints            int64      `json:"-" form:"-"`
		UnspendableReferralPoints int64      `json:"-" form:"-"`
		SpendableReferralPoints   int64      `json:"-" form:"-"`
		RewardPoints              int64      `json:"-" form:"-"`
		SpendablePoints           int64      `json:"-" form:"-"`
		Wallet                    int64      `json:"-" form:"-"`
	}

	BuyerFilter struct {
		userFilter
	}
)

func NewBuyerFilter() *BuyerFilter {
	return new(BuyerFilter{
		MapSortings: NewMap(),
		UrlValues:   url.Values{},
	})
}

func (q *BuyerFilter) WithBuyerIDs(buyerIDs ...uuid.UUID) *BuyerFilter {
	q.UserIDs = buyerIDs
	return q
}

func (q *BuyerFilter) WithLogin(login string) *BuyerFilter {
	q.Login = login
	return q
}

func (q *BuyerFilter) WithKeyword(keyword string) *BuyerFilter {
	q.Keyword = keyword
	return q
}

func (q *BuyerFilter) WithUsername(username string) *BuyerFilter {
	q.Username = username
	return q
}

func (q *BuyerFilter) WithEmail(email string) *BuyerFilter {
	q.Email = email
	return q
}

func (q *BuyerFilter) WithMobilePhone(mobilePhone string) *BuyerFilter {
	q.MobilePhone = mobilePhone
	return q
}

func (q *BuyerFilter) WithPasswordResetToken(passwordResetToken string) *BuyerFilter {
	q.PasswordResetToken = passwordResetToken
	return q
}

func (q *BuyerFilter) WithPaginationURL(paginationURL string) *BuyerFilter {
	q.PaginationURL = paginationURL
	return q
}

func (q *BuyerFilter) WithSortings(sortings ...Sorting) *BuyerFilter {
	q.Sortings = sortings
	q.MapSortings.Clear()
	for _, sorting := range sortings {
		q.MapSortings.Push(strings.ToLower(helper.Sprintf("%s:%s", sorting.Field, sorting.Direction)), struct{}{})
	}
	return q
}

func (q *BuyerFilter) WithLimit(limit int64) *BuyerFilter {
	q.Limit = limit
	return q
}

func (q *BuyerFilter) WithPage(page int64) *BuyerFilter {
	q.Page = page
	return q
}

func (q *BuyerFilter) WithUrlValues(urlValues url.Values) *BuyerFilter {
	q.UrlValues = urlValues
	return q
}
