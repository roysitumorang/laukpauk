package models

import (
	"time"
	"uuid"
)

type TransactionCounter struct {
	ID                uuid.UUID `json:"-"`
	TransactionWindow string    `json:"-"`
	Hits              int64     `json:"-"`
	CreatedAt         time.Time `json:"-"`
	UpdatedAt         time.Time `json:"-"`
}
