package models

import (
	"errors"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/google/uuid"
)

const (
	LandingPageBlockPhotoName = "%s-%d%s"
	LandingPageBlockPhotoPath = "images/landing-page/%s/%s"

	LandingPageBlockNameHero         LandingPageBlockName = "hero"
	LandingPageBlockNameFeature      LandingPageBlockName = "feature"
	LandingPageBlockNameWhy          LandingPageBlockName = "why"
	LandingPageBlockNameFreeShopping LandingPageBlockName = "free-shopping"
	LandingPageBlockNameCommitment   LandingPageBlockName = "commitment"
	LandingPageBlockNameKitchenHero  LandingPageBlockName = "kitchen-hero"
	LandingPageBlockNameFAQ          LandingPageBlockName = "faq"
)

type (
	LandingPageBlockName string
	LandingPageBlock     struct {
		RowNo            int64                 `json:"-" form:"-"`
		ID               uuid.UUID             `json:"-" form:"-"`
		RecordStatus     RecordStatus          `json:"-" form:"-"`
		ParentName       *LandingPageBlockName `json:"-" form:"-"`
		Name             *LandingPageBlockName `json:"-" form:"-"`
		Subject          *string               `json:"-" form:"subject"`
		Body             *string               `json:"-" form:"body"`
		Photo            *string               `json:"-" form:"photo"`
		LinkText         *string               `json:"-" form:"link_text"`
		LinkURL          *string               `json:"-" form:"link_url"`
		CustomAttributes map[string]string     `json:"-" form:"custom_attributes"`
		CreatedBy        uuid.UUID             `json:"-" form:"-"`
		CreatedAt        time.Time             `json:"-" form:"-"`
		UpdatedBy        uuid.UUID             `json:"-" form:"-"`
		UpdatedAt        time.Time             `json:"-" form:"-"`
		DeletedBy        *uuid.UUID            `json:"-" form:"-"`
		DeletedAt        *time.Time            `json:"-" form:"-"`
		DeletionReason   *string               `json:"-" form:"-"`
	}

	LandingPageBlockFilter struct {
		LandingPageBlockIDs []uuid.UUID
		RecordStatus        []RecordStatus
		ParentNames,
		Names []LandingPageBlockName
		Keyword,
		PaginationURL string
		Limit,
		Page int64
		UrlValues url.Values
	}
)

func NewLandingPageBlockFilter() *LandingPageBlockFilter {
	return new(LandingPageBlockFilter{UrlValues: url.Values{}})
}

func (q *LandingPageBlockFilter) WithLandingPageBlockIDs(LandingPageBlockItemIDs ...uuid.UUID) *LandingPageBlockFilter {
	q.LandingPageBlockIDs = LandingPageBlockItemIDs
	return q
}

func (q *LandingPageBlockFilter) WithRecordStatus(recordStatus ...RecordStatus) *LandingPageBlockFilter {
	q.RecordStatus = recordStatus
	return q
}

func (q *LandingPageBlockFilter) WithParentNames(parentNames ...LandingPageBlockName) *LandingPageBlockFilter {
	q.ParentNames = parentNames
	return q
}

func (q *LandingPageBlockFilter) WithNames(names ...LandingPageBlockName) *LandingPageBlockFilter {
	q.Names = names
	return q
}

func (q *LandingPageBlockFilter) WithKeyword(keyword string) *LandingPageBlockFilter {
	q.Keyword = keyword
	return q
}

func (q *LandingPageBlockFilter) WithPaginationURL(paginationURL string) *LandingPageBlockFilter {
	q.PaginationURL = paginationURL
	return q
}

func (q *LandingPageBlockFilter) WithLimit(limit int64) *LandingPageBlockFilter {
	q.Limit = limit
	return q
}

func (q *LandingPageBlockFilter) WithPage(page int64) *LandingPageBlockFilter {
	q.Page = page
	return q
}

func (q *LandingPageBlockFilter) WithUrlValues(urlValues url.Values) *LandingPageBlockFilter {
	q.UrlValues = urlValues
	return q
}

func (q *LandingPageBlock) Validate() error {
	if q.Subject != nil {
		*q.Subject = strings.TrimSpace(*q.Subject)
	}
	if q.Body != nil {
		*q.Body = strings.TrimSpace(*q.Body)
	}
	if q.LinkText != nil {
		*q.LinkText = strings.TrimSpace(*q.LinkText)
	}
	if q.LinkURL != nil {
		*q.LinkURL = strings.TrimSpace(*q.LinkURL)
	}
	if q.ParentName == nil {
		return nil
	}
	switch *q.ParentName {
	case LandingPageBlockNameHero:
		if q.Subject == nil ||
			*q.Subject == "" {
			return errors.New("subject: is required")
		}
		if q.Body == nil ||
			*q.Body == "" {
			return errors.New("body: is required")
		}
		if q.LinkText == nil ||
			*q.LinkText == "" {
			return errors.New("link_text: is required")
		}
		if q.LinkURL == nil ||
			*q.LinkURL == "" {
			return errors.New("link_url: is required")
		}
		if _, err := url.ParseRequestURI(*q.LinkURL); err != nil {
			return errors.New("link_url: invalid URL")
		}
		if q.ID == uuid.Nil &&
			(q.Photo == nil ||
				*q.Photo == "") {
			return errors.New("photo: is required")
		}
	case LandingPageBlockNameFeature, LandingPageBlockNameWhy:
		if q.Subject == nil ||
			*q.Subject == "" {
			return errors.New("subject: is required")
		}
		if q.Body == nil ||
			*q.Body == "" {
			return errors.New("body: is required")
		}
		q.LinkText = nil
		q.LinkURL = nil
		if q.ID == uuid.Nil &&
			(q.Photo == nil ||
				*q.Photo == "") {
			return errors.New("photo: is required")
		}
	}
	return nil
}

func (q *LandingPageBlock) SetPhoto(photo *string) *LandingPageBlock {
	q.Photo = photo
	return q
}

func (q *LandingPageBlock) PhotoPath() *string {
	if q.ParentName == nil ||
		q.Photo == nil {
		return nil
	}
	var sb strings.Builder
	_, _ = fmt.Fprintf(&sb, LandingPageBlockPhotoPath, *q.ParentName, *q.Photo)
	return new(sb.String())
}
