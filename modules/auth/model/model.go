package model

import (
	"github.com/gofiber/fiber/v2"
	"github.com/roysitumorang/laukpauk/errors"
	userModel "github.com/roysitumorang/laukpauk/modules/user/model"
)

var ErrLoginFailed = errors.New(fiber.StatusBadRequest, "login failed")

type (
	LoginRequest struct {
		MobilePhone string `json:"mobile_phone"`
		Password    string `json:"password"`
	}

	LoginResponse struct {
		IDToken   string         `json:"id_token"`
		ExpiresIn int64          `json:"expires_in"`
		Profile   userModel.User `json:"profile"`
	}
)
