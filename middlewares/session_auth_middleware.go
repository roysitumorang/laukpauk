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
		if len(adminLevels) > 0 && currentAdmin.AdminLevel != adminLevels[0] {
			return flash.Danger(models.ErrAccessDenied.Error()).Redirect(c, sess.Session, "/admin")
		}
		return c.Next()
	}
}

func SellerSessionAuthMiddleware(
	sellerService services.SellerService,
	sellerLevels ...models.SellerLevel,
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
			return flash.Redirect(c, sess.Session, "/seller/auth/login")
		}
		currentSeller, err := sellerService.FindSeller(
			ctx,
			models.NewSellerFilter().
				WithSellerIDs(userID),
		)
		if err != nil || currentSeller == nil {
			_ = sess.ResetWithContext(ctx)
			if c.Method() == fiber.MethodGet {
				sess.Set("next", c.OriginalURL())
			}
			return flash.Redirect(c, sess.Session, "/seller/auth/login")
		}
		sess.Set(models.CurrentSeller, currentSeller)
		_ = sess.Session.SaveWithContext(ctx)
		if len(sellerLevels) > 0 && currentSeller.SellerLevel != sellerLevels[0] {
			return flash.Danger(models.ErrAccessDenied.Error()).Redirect(c, sess.Session, "/seller")
		}
		return c.Next()
	}
}
