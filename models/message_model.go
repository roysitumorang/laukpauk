package models

import (
	"errors"
	"fmt"
	"strings"
	"time"
	"uuid"

	"github.com/nyaruka/phonenumbers"
)

type (
	Message struct {
		RowNo       int64     `json:"-" form:"-"`
		ID          uuid.UUID `json:"-" form:"-"`
		Name        string    `json:"-" form:"name"`
		MobilePhone string    `json:"-" form:"mobile_phone"`
		Body        string    `json:"-" form:"body"`
		CreatedAt   time.Time `json:"-" form:"-"`
	}
)

func (q *Message) Validate() error {
	if q.Name = strings.TrimSpace(q.Name); q.Name == "" {
		return errors.New("name: is required")
	}
	if q.MobilePhone = strings.TrimSpace(q.MobilePhone); q.MobilePhone == "" {
		return errors.New("mobile_phone: is required")
	}
	mobilePhone, err := phonenumbers.Parse(q.MobilePhone, "ID")
	if err != nil {
		return fmt.Errorf("mobile_phone: %w", err)
	}
	if q.Body = strings.TrimSpace(q.Body); q.Body == "" {
		return errors.New("body: is required")
	}
	q.MobilePhone = phonenumbers.Format(mobilePhone, phonenumbers.E164)
	return nil
}
