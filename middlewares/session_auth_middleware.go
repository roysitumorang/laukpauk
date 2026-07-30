package middlewares

import (
	"uuid"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/session"
	"github.com/roysitumorang/laukpauk/models"
	"github.com/roysitumorang/laukpauk/services"
)

func AdminSessionAuthMiddleware(
	adminService services.AdminService,
) fiber.Handler {
	return func(c fiber.Ctx) error {
		ctx := c.Context()
		sess := session.FromContext(c)
		userID, ok := sess.Get(models.UserID).(uuid.UUID)
		if !ok {
			if c.Method() == fiber.MethodGet {
				sess.Set("next", c.OriginalURL())
				_ = sess.Session.SaveWithContext(ctx)
			}
			return c.Redirect().To("/admin/auth/login")
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
			_ = sess.Session.SaveWithContext(ctx)
			return c.Redirect().To("/admin/auth/login")
		}
		sess.Set(models.CurrentAdmin, currentAdmin)
		_ = sess.Session.SaveWithContext(ctx)
		return c.Next()
	}
}
