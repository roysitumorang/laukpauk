package sanitizer

import (
	"context"
	"encoding/base64"
	"errors"
	"strings"
	"unsafe"

	"github.com/gofiber/fiber/v2"
	"github.com/roysitumorang/laukpauk/helper"
	"github.com/roysitumorang/laukpauk/modules/auth/model"
	"go.uber.org/zap"
)

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
	statusCode = fiber.StatusOK
	request.Password = unsafe.String(unsafe.SliceData(password), len(password))
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
	statusCode = fiber.StatusOK
	request.OldPassword = unsafe.String(unsafe.SliceData(oldPassword), len(oldPassword))
	request.NewPassword = unsafe.String(unsafe.SliceData(newPassword), len(newPassword))
	return
}
