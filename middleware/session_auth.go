package middleware

import (
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/session"
	"github.com/google/uuid"
	"github.com/roysitumorang/laukpauk/models"
	"github.com/roysitumorang/laukpauk/services"
)

func AdminSessionAuth(
	userService services.UserService,
) fiber.Handler {
	return func(c fiber.Ctx) error {
		ctx := c.Context()
		sess := session.FromContext(c)
		userID, ok := sess.Get(models.UserID).(uuid.UUID)
		if !ok {
			return c.Redirect().To("/admin/auth/login")
		}
		admins, _, err := userService.FindAdmins(
			ctx,
			models.NewUserFilter().WithUserIDs(userID),
		)
		if err != nil || len(admins) == 0 {
			_ = sess.Reset()
			_ = sess.Session.Save()
			return c.Redirect().To("/admin/auth/login")
		}
		currentAdmin := admins[0]
		sess.Set(models.CurrentAdmin, currentAdmin)
		_ = sess.Session.Save()
		return c.Next()
	}
}
