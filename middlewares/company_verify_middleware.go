package middlewares

import (
	"github.com/gofiber/fiber/v3"
	"github.com/roysitumorang/laukpauk/helper"
	"github.com/roysitumorang/laukpauk/models"
	"github.com/roysitumorang/laukpauk/services"
)

func CompanyVerifyMiddleware(
	companyService services.CompanyService,
) fiber.Handler {
	return func(c fiber.Ctx) error {
		ctx := c.Context()
		company, err := companyService.FindCompany(
			ctx,
			models.NewCompanyFilter().
				WithSlugs(c.Params("slug")),
		)
		if err != nil {
			return helper.NewResponse(fiber.StatusBadRequest).
				SetMessage(err.Error()).
				WriteResponse(c)
		}
		if company == nil || !company.Android {
			return helper.NewResponse(fiber.StatusBadRequest).
				SetMessage(models.ErrCompanyNotFound.Error()).
				WriteResponse(c)
		}
		c.Locals(models.CurrentCompany, company)
		return c.Next()
	}
}
