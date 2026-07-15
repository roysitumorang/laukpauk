package models

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/roysitumorang/laukpauk/helper"
)

const (
	CurrentAdmin = "current_admin"
	CurrentBuyer = "current_buyer"
	UserID       = "user_id"
)

var (
	ErrLoginFailed = errors.New("login failed")
)

type (
	LoginRequest struct {
		Login          string `json:"login" form:"login"`
		Base64Password string `json:"password" form:"password"`
		Password       string `json:"-" form:"-"`
	}

	BuyerLoginResponse struct {
		IDToken   string    `json:"id_token"`
		ExpiredAt time.Time `json:"expired_at"`
		Buyer     *Buyer    `json:"buyer"`
	}
)

func (q *LoginRequest) Validate() error {
	if q.Login = strings.TrimSpace(q.Login); q.Login == "" {
		return errors.New("login: is required")
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
