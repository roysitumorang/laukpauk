package model

import (
	"net"

	"github.com/gofiber/fiber/v2"
	"github.com/roysitumorang/laukpauk/errors"
	userModel "github.com/roysitumorang/laukpauk/modules/user/model"
)

var (
	ErrLoginFailed      = errors.New(fiber.StatusBadRequest, "login failed")
	ErrActivationFailed = errors.New(fiber.StatusBadRequest, "activation failed")
)

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

	ChangePassword struct {
		OldPassword string `json:"old_password"`
		NewPassword string `json:"new_password"`
	}

	RegisterRequest struct {
		RoleID        int64  `json:"-"`
		Address       string `json:"address"`
		VillageID     int64  `json:"village_id"`
		SubdistrictID int64  `json:"-"`
		Name          string `json:"name"`
		Password      string `json:"password"`
		MobilePhone   string `json:"mobile_phone"`
		IpAddress     net.IP `json:"-"`
	}

	RegisterResponse struct {
		ActivationToken string `json:"activation_token"`
	}
)
