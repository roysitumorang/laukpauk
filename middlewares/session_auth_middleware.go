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
			return flash.Redirect(c, sess.Session, "/admin/sessions/new")
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
			return flash.Redirect(c, sess.Session, "/admin/sessions/new")
		}
		c.Locals(models.CurrentAdmin, currentAdmin)
		_ = sess.Session.SaveWithContext(ctx)
		if len(adminLevels) > 0 && currentAdmin.AdminLevel != adminLevels[0] {
			return flash.Danger(models.ErrAccessDenied.Error()).Redirect(c, sess.Session, "/admin")
		}
		return c.Next()
	}
}

func SellerSessionAuthMiddleware(
	sellerService services.SellerService,
	notificationService services.NotificationService,
	sellerLevels ...models.SellerLevel,
) fiber.Handler {
	return func(c fiber.Ctx) error {
		ctx := c.Context()
		sess := session.FromContext(c)
		flash, ok := sess.Get(helper.Flash).(*helper.FlashMessage)
		if !ok {
			flash = helper.NewFlashMessage()
		}
		sellerID, ok := sess.Get(models.SellerID).(uuid.UUID)
		if !ok {
			if c.Method() == fiber.MethodGet {
				sess.Set("next", c.OriginalURL())
			}
			return flash.Redirect(c, sess.Session, "/seller/otp/new")
		}
		currentSeller, err := sellerService.FindSeller(
			ctx,
			models.NewSellerFilter().
				WithSellerIDs(sellerID),
		)
		if err != nil || currentSeller == nil {
			_ = sess.ResetWithContext(ctx)
			if c.Method() == fiber.MethodGet {
				sess.Set("next", c.OriginalURL())
			}
			return flash.Redirect(c, sess.Session, "/seller/otp/new")
		}
		notifications, _, err := notificationService.FindNotifications(
			ctx,
			models.NewNotificationFilter().
				WithSellerIDs(currentSeller.ID).
				WithReadAtBlank(new(true)),
		)
		if err != nil {
			_ = sess.ResetWithContext(ctx)
			if c.Method() == fiber.MethodGet {
				sess.Set("next", c.OriginalURL())
			}
			return flash.Redirect(c, sess.Session, "/seller/otp/new")
		}
		c.Locals(models.CurrentSeller, currentSeller)
		c.Locals(models.CurrentNotifications, notifications)
		_ = sess.Session.SaveWithContext(ctx)
		if len(sellerLevels) > 0 && currentSeller.SellerLevel != sellerLevels[0] {
			return flash.Danger(models.ErrAccessDenied.Error()).Redirect(c, sess.Session, "/seller")
		}
		return c.Next()
	}
}
