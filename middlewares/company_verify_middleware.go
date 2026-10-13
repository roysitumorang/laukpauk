package middlewares

import (
	"os"
	"strings"

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
		subdomain := strings.ReplaceAll(c.Hostname(), helper.Sprintf(".%s", os.Getenv("DOMAIN")), "")
		company, err := companyService.FindCompany(
			ctx,
			models.NewCompanyFilter().
				WithSubdomains(subdomain),
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
