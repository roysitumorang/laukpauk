package models

import (
	"errors"
	"fmt"
	"strings"

	"github.com/roysitumorang/laukpauk/helper"
)

const (
	CurrentAdmin = "current_admin"
	UserID       = "user_id"
)

var (
	ErrLoginFailed = errors.New("login failed")
)

type (
	AdminLoginRequest struct {
		Email          string `json:"email" form:"login"`
		Base64Password string `json:"password" form:"password"`
		Password       string `json:"-" form:"-"`
	}
)

func (q *AdminLoginRequest) Validate() error {
	if q.Email = strings.TrimSpace(q.Email); q.Email == "" {
		return errors.New("email: is required")
	}
	if q.Base64Password = strings.TrimSpace(q.Base64Password); q.Base64Password == "" {
		return errors.New("password: is required")
	}
	password, err := helper.Base64Decode(q.Base64Password)
	if err != nil {
		return fmt.Errorf("password: %s", err.Error())
	}
	q.Password = password
	return nil
}
