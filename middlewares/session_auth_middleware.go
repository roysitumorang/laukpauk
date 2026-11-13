package middlewares

import (
	"uuid"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/session"
	"github.com/roysitumorang/laukpauk/helper"
	"github.com/roysitumorang/laukpauk/models"
	"github.com/roysitumorang/laukpauk/services"
)

func AdminSessionAuthMiddleware(
	adminService services.AdminService,
	adminLevels ...models.AdminLevel,
) fiber.Handler {
	return func(c fiber.Ctx) error {
		ctx := c.Context()
		sess := session.FromContext(c)
		flash, ok := sess.Get(helper.Flash).(*helper.FlashMessage)
		if !ok {
			flash = helper.NewFlashMessage()
		}
		adminID, ok := sess.Get(models.AdminID).(uuid.UUID)
		if !ok {
			if c.Method() == fiber.MethodGet {
				sess.Set("next", c.OriginalURL())
			}
			return flash.Redirect(c, sess.Session, "/admin/auth/sessions/new")
		}
		currentAdmin, err := adminService.FindAdmin(
			ctx,
			models.NewAdminFilter().
				WithAdminIDs(adminID),
		)
		if err != nil || currentAdmin == nil {
			_ = sess.ResetWithContext(ctx)
			if c.Method() == fiber.MethodGet {
				sess.Set("next", c.OriginalURL())
			}
			return flash.Redirect(c, sess.Session, "/admin/auth/sessions/new")
		}
		c.Locals(models.CurrentAdmin, currentAdmin)
		_ = sess.Session.SaveWithContext(ctx)
		if len(adminLevels) > 0 && currentAdmin.AdminLevel != adminLevels[0] {
			return flash.Danger(models.ErrAccessDenied.Error()).Redirect(c, sess.Session, "/admin")
		}
		return c.Next()
	}
}

func SessionAuthMiddleware(
	role string,
	userService services.UserService,
	notificationService services.NotificationService,
	userLevels ...string,
) fiber.Handler {
	return func(c fiber.Ctx) error {
		ctx := c.Context()
		sess := session.FromContext(c)
		flash, ok := sess.Get(helper.Flash).(*helper.FlashMessage)
		if !ok {
			flash = helper.NewFlashMessage()
		}
		userID, ok := sess.Get(models.UserID).(uuid.UUID)
		if !ok {
			if c.Method() == fiber.MethodGet {
				sess.Set("next", c.OriginalURL())
			}
			return flash.Redirect(c, sess.Session, "/auth/sessions/new")
		}
		currentUser, err := userService.FindUser(
			ctx,
			models.NewUserFilter().
				WithRoles(role).
				WithUserIDs(userID),
		)
		if err != nil || currentUser == nil {
			_ = sess.ResetWithContext(ctx)
			if c.Method() == fiber.MethodGet {
				sess.Set("next", c.OriginalURL())
			}
			return flash.Redirect(c, sess.Session, "/auth/sessions/new")
		}
		notifications, _, err := notificationService.FindNotifications(
			ctx,
			models.NewNotificationFilter().
				WithUserIDs(currentUser.ID).
				WithReadAtBlank(new(true)),
		)
		if err != nil {
			_ = sess.ResetWithContext(ctx)
			if c.Method() == fiber.MethodGet {
				sess.Set("next", c.OriginalURL())
			}
			return flash.Redirect(c, sess.Session, "/otp/new")
		}
		c.Locals(models.CurrentUser, currentUser)
		c.Locals(models.CurrentNotifications, notifications)
		_ = sess.Session.SaveWithContext(ctx)
		if len(userLevels) > 0 && currentUser.UserLevel != userLevels[0] {
			return flash.Danger(models.ErrAccessDenied.Error()).Redirect(c, sess.Session, "/admin")
		}
		return c.Next()
	}
}
