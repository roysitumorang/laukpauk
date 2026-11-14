package models

import (
	"errors"
)

const CurrentJsonWebToken = "current_jwt"

var ErrLoginFailed = errors.New("login failed")
