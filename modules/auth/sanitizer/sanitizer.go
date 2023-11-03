package sanitizer

import (
	"context"
	"encoding/base64"
	"errors"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/nyaruka/phonenumbers"
	"github.com/roysitumorang/laukpauk/helper"
	"github.com/roysitumorang/laukpauk/modules/auth/model"
	"go.uber.org/zap"
)

func Register(ctx context.Context, c *fiber.Ctx) (request model.RegisterRequest, statusCode int, err error) {
	ctxt := "AuthSanitizer-Register"
	statusCode = fiber.StatusBadRequest
	err = c.BodyParser(&request)
	var fiberErr *fiber.Error
	if errors.As(err, &fiberErr) {
		statusCode = fiberErr.Code
		helper.Log(ctx, zap.ErrorLevel, err.Error(), ctxt, "ErrBodyParser")
		return
	}
	if request.MobilePhone = strings.TrimSpace(request.MobilePhone); request.MobilePhone == "" {
		err = errors.New("mobile phone is required")
		return
	}
	phoneNumber, err := phonenumbers.Parse(request.MobilePhone, "ID")
	if err != nil {
		helper.Log(ctx, zap.ErrorLevel, err.Error(), ctxt, "ErrParse")
		return
	}
	request.MobilePhone = phonenumbers.Format(phoneNumber, phonenumbers.E164)
	if request.Password = strings.TrimSpace(request.Password); request.Password == "" {
		err = errors.New("password is required")
		return
	}
	password, err := base64.StdEncoding.DecodeString(request.Password)
	if err != nil {
		err = errors.New("invalid password")
		return
	}
	request.Password = helper.ByteSlice2String(password)
	if request.VillageID == 0 {
		err = errors.New("village_id is required")
		return
	}
	if request.Address = strings.TrimSpace(request.Address); request.Address == "" {
		err = errors.New("address is required")
		return
	}
	request.IpAddress = helper.GetIPAdress(c.Request())
	statusCode = fiber.StatusOK
	return
}

func Login(ctx context.Context, c *fiber.Ctx) (request model.LoginRequest, statusCode int, err error) {
	ctxt := "AuthSanitizer-Login"
	statusCode = fiber.StatusBadRequest
	err = c.BodyParser(&request)
	var fiberErr *fiber.Error
	if errors.As(err, &fiberErr) {
		statusCode = fiberErr.Code
		helper.Log(ctx, zap.ErrorLevel, err.Error(), ctxt, "ErrBodyParser")
		return
	}
	if request.MobilePhone = strings.TrimSpace(request.MobilePhone); request.MobilePhone == "" {
		err = errors.New("mobile phone is required")
		return
	}
	if request.Password = strings.TrimSpace(request.Password); request.Password == "" {
		err = errors.New("password is required")
		return
	}
	password, err := base64.StdEncoding.DecodeString(request.Password)
	if err != nil {
		err = errors.New("invalid password")
		return
	}
	request.Password = helper.ByteSlice2String(password)
	statusCode = fiber.StatusOK
	return
}

func ChangePassword(ctx context.Context, c *fiber.Ctx) (request model.ChangePassword, statusCode int, err error) {
	ctxt := "AuthSanitizer-ChangePassword"
	statusCode = fiber.StatusBadRequest
	err = c.BodyParser(&request)
	var fiberErr *fiber.Error
	if errors.As(err, &fiberErr) {
		statusCode = fiberErr.Code
		helper.Log(ctx, zap.ErrorLevel, err.Error(), ctxt, "ErrBodyParser")
		return
	}
	if request.OldPassword = strings.TrimSpace(request.OldPassword); request.OldPassword == "" {
		err = errors.New("old password is required")
		return
	}
	if request.NewPassword = strings.TrimSpace(request.NewPassword); request.NewPassword == "" {
		err = errors.New("new password is required")
		return
	}
	oldPassword, err := base64.StdEncoding.DecodeString(request.OldPassword)
	if err != nil {
		err = errors.New("invalid old password")
		return
	}
	newPassword, err := base64.StdEncoding.DecodeString(request.NewPassword)
	if err != nil {
		err = errors.New("invalid new password")
		return
	}
	request.OldPassword = helper.ByteSlice2String(oldPassword)
	request.NewPassword = helper.ByteSlice2String(newPassword)
	statusCode = fiber.StatusOK
	return
}
