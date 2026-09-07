package models

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/roysitumorang/laukpauk/helper"
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

const CurrentAdmin = "current_admin"
const CurrentSeller = "current_seller"
const CurrentBuyer = "current_buyer"
const CurrentJsonWebToken = "current_jwt"
const AdminID = "admin_id"
const SellerID = "seller_id"
const BuyerID = "buyer_id"

var ErrLoginFailed = errors.New("login failed")

type SellerLoginResponse struct {
	IDToken   string    `json:"id_token"`
	ExpiredAt time.Time `json:"expired_at"`
	Seller    *Seller   `json:"seller"`
}

type BuyerLoginResponse struct {
	IDToken   string    `json:"id_token"`
	ExpiredAt time.Time `json:"expired_at"`
	Buyer     *Buyer    `json:"buyer"`
}

type LoginRequest struct {
	Login          string `json:"login" form:"login"`
	Base64Password string `json:"password" form:"password"`
	Password       string `json:"-" form:"-"`
}

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

type SellerLoginRequest struct {
	Login          string `json:"login" form:"login"`
	Base64OtpToken string `json:"otp_token" form:"otp_token"`
	OtpToken       string `json:"-" form:"-"`
}

func (q *SellerLoginRequest) Validate() error {
	if q.Login = strings.TrimSpace(q.Login); q.Login == "" {
		return errors.New("login: is required")
	}
	if q.Base64OtpToken = strings.TrimSpace(q.Base64OtpToken); q.Base64OtpToken == "" {
		return nil
	}
	otpToken, err := helper.Base64Decode(q.Base64OtpToken)
	if err != nil {
		return fmt.Errorf("otp_token: %s", err.Error())
	}
	q.OtpToken = strings.TrimSpace(otpToken)
	return nil
}
