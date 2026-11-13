package models

import (
	"errors"
)

var ErrLoginNotFound = errors.New("login: not found")

const CurrentJsonWebToken = "current_jwt"
const CurrentSubdomain = "current_subdomain"
const CurrentNotifications = "current_notifications"

var ErrLoginFailed = errors.New("login failed")
