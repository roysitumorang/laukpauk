package models

import (
	"errors"
	"time"
	"uuid"
)

var ErrSettingNotFound = errors.New("setting not found")

type Setting struct {
	ID                      uint64    `json:"-" form:"id"`
	MaintenanceMode         bool      `json:"maintenance_mode" form:"maintenance_mode"`
	SubscriptionTrialPeriod int       `json:"-" form:"subscription_trial_period"`
	MonthlyBillingAmount    int64     `json:"-" form:"monthly_billing_amount"`
	AnnualBillingAmount     int64     `json:"-" form:"annual_billing_amount"`
	BillingGracePeriod      int       `json:"-" form:"billing_grace_period"`
	CreatedBy               uuid.UUID `json:"-" form:"-"`
	CreatedAt               time.Time `json:"-" form:"-"`
	UpdatedBy               uuid.UUID `json:"-" form:"-"`
	UpdatedAt               time.Time `json:"-" form:"-"`
}

func (q *Setting) Validate() error {
	if q.SubscriptionTrialPeriod < 1 {
		return errors.New("subscription_trial_period: is required")
	}
	if q.MonthlyBillingAmount < 1 {
		return errors.New("monthly_billing_amount: is required")
	}
	if q.AnnualBillingAmount < 1 {
		return errors.New("annual_billing_amount: is required")
	}
	if q.BillingGracePeriod < 1 {
		return errors.New("billing_grace_period: is required")
	}
	return nil
}

func (q *Setting) Populate() *Setting {
	return q
}
