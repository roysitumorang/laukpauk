package models

import (
	"errors"
	"strings"
)

var ErrLoginNotFound = errors.New("login: not found")

type OtpRequest struct {
	Login string `json:"login" form:"login"`
}

func (q *OtpRequest) Validate() error {
	if q.Login = strings.TrimSpace(q.Login); q.Login == "" {
		return errors.New("login: is required")
	}
	return nil
}
