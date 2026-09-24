package models

import (
	"time"
	"uuid"
)

type TransactionCounter struct {
	ID                uuid.UUID `json:"-" form:"-"`
	CompanyID         uuid.UUID `json:"-" form:"-"`
	TransactionWindow string    `json:"-" form:"-"`
	Hits              int64     `json:"-" form:"-"`
	CreatedAt         time.Time `json:"-" form:"-"`
	UpdatedAt         time.Time `json:"-" form:"-"`
}
