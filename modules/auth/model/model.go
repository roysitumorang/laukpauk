package model

import (
	"errors"

	userModel "github.com/roysitumorang/laukpauk/modules/user/model"
)

var ErrLoginFailed = errors.New("login failed")

type (
	SignInRequest struct {
		MobilePhone string `json:"mobile_phone"`
		Password    string `json:"password"`
	}

	SignInResponse struct {
		IDToken   string         `json:"id_token"`
		ExpiresIn int64          `json:"expires_in"`
		Profile   userModel.User `json:"profile"`
	}
)
