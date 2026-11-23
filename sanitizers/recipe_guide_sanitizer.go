package sanitizers

import (
	"context"
	"errors"
	"log/slog"

	"github.com/gofiber/fiber/v3"
	"github.com/roysitumorang/laukpauk/helper"
	"github.com/roysitumorang/laukpauk/models"
)

func ValidateRecipeGuide(ctx context.Context, c fiber.Ctx) (*models.RecipeGuide, int, error) {
	ctxt := "RecipeGuideSanitizer-ValidateRecipeGuide"
	var response models.RecipeGuide
	err := c.Bind().Body(&response)
	if fiberErr, ok := errors.AsType[*fiber.Error](err); ok {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrBody")
		return nil, fiberErr.Code, err
	}
	if err = (&response).Validate(); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrValidate")
		return &response, fiber.StatusBadRequest, err
	}
	return &response, fiber.StatusOK, nil
}
