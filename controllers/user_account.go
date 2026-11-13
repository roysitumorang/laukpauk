package controllers

import (
	"context"
	"errors"
	"uuid"

	"github.com/gofiber/fiber/v3"
	"github.com/roysitumorang/laukpauk/models"
	"github.com/roysitumorang/laukpauk/services"
)

// checkUniqueLogins rejects a username, mobile phone or email that another
// user (seller or buyer) already has. excludedUserID is the account being
// updated, or uuid.Nil() when creating one.
func checkUniqueLogins(
	ctx context.Context,
	userService services.UserService,
	request *models.UserAccount,
	excludedUserID uuid.UUID,
) (int, error) {
	type check struct {
		filter *models.UserFilter
		err    error
	}
	checks := []check{
		{models.NewUserFilter().WithUsername(request.Username), models.ErrUniqueUserUsernameViolation},
		{models.NewUserFilter().WithMobilePhone(request.MobilePhone), models.ErrUniqueUserMobilePhoneViolation},
	}
	if request.Email != nil {
		checks = append(checks, check{models.NewUserFilter().WithEmail(*request.Email), models.ErrUniqueUserEmailViolation})
	}
	for _, check := range checks {
		if excludedUserID.Compare(uuid.Nil()) != 0 {
			check.filter.WithExcludedUserIDs(excludedUserID)
		}
		other, err := userService.FindUser(ctx, check.filter)
		if err != nil && !errors.Is(err, models.ErrUserNotFound) {
			return fiber.StatusUnprocessableEntity, err
		}
		if other != nil {
			return fiber.StatusUnprocessableEntity, check.err
		}
	}
	return fiber.StatusOK, nil
}

// userNotFoundStatusCode answers 404 when an account lookup found nothing.
func userNotFoundStatusCode(err error) int {
	if errors.Is(err, models.ErrUserNotFound) {
		return fiber.StatusNotFound
	}
	return fiber.StatusUnprocessableEntity
}
