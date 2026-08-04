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
	adminTypes ...models.AdminType,
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
			return flash.Redirect(c, sess.Session, "/admin/auth/login")
		}
		currentAdmin, err := adminService.FindAdmin(
			ctx,
			models.NewAdminFilter().
				WithAdminIDs(userID),
		)
		if err != nil || currentAdmin == nil {
			_ = sess.ResetWithContext(ctx)
			if c.Method() == fiber.MethodGet {
				sess.Set("next", c.OriginalURL())
			}
			return flash.Redirect(c, sess.Session, "/admin/auth/login")
		}
		sess.Set(models.CurrentAdmin, currentAdmin)
		_ = sess.Session.SaveWithContext(ctx)
		if len(adminTypes) > 0 && currentAdmin.AdminType != adminTypes[0] {
			return flash.Danger(models.ErrAccessDenied.Error()).Redirect(c, sess.Session, "/admin")
		}
		return c.Next()
	}
}
