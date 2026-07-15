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
		ReferralCode   string     `json:"referral_code" form:"-"`
		Address        string     `json:"address" form:"-"`
		ReferrerID     *uuid.UUID `json:"referrer_id" form:"-"`
		ReferralPoints int64      `json:"referral_points" form:"-"`
		RewardPoints   int64      `json:"reward_points" form:"-"`
		Wallet         int64      `json:"wallet" form:"-"`
	}

	BuyerFilter struct {
		userFilter
		ReferralCodes []string
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

func (q *BuyerFilter) WithExcludedBuyerIDs(excludedBuyerIDs ...uuid.UUID) *BuyerFilter {
	q.ExcludedUserIDs = excludedBuyerIDs
	return q
}

func (q *BuyerFilter) WithUserStatuses(userStatuses ...UserStatus) *BuyerFilter {
	q.UserStatuses = userStatuses
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

func (q *BuyerFilter) WithReferralCodes(referralCodes ...string) *BuyerFilter {
	q.ReferralCodes = referralCodes
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
