package sanitizers

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/url"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/utils/v2"
	"github.com/roysitumorang/laukpauk/helper"
	"github.com/roysitumorang/laukpauk/models"
)

func FindCustomDeliveryDays(ctx context.Context, c fiber.Ctx) (*models.CustomDeliveryDayFilter, error) {
	ctxt := "CustomDeliveryDaySanitizer-FindCustomDeliveryDays"
	response := models.NewCustomDeliveryDayFilter()
	originalURL, err := url.ParseRequestURI(c.OriginalURL())
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrParseRequestURI")
		return response, err
	}
	var sb strings.Builder
	_, _ = fmt.Fprintf(&sb, "%s%s", c.BaseURL(), originalURL.Path)
	urlValues := originalURL.Query()
	response.WithPaginationURL(sb.String())
	if keyword := strings.TrimSpace(c.Query("q")); keyword != "" {
		urlValues.Set("q", keyword)
		response.WithKeyword(keyword)
	}
	if sortBy := urlValues["sort_by"]; len(sortBy) > 0 {
		urlValues.Set("sort_by", c.Query("sort_by"))
		var sortings []models.Sorting
		for _, sorting := range sortBy {
			if !strings.Contains(sorting, ":") {
				if _, ok := models.MapCustomDeliveryDaySortings[sorting]; ok {
					sortings = append(sortings, models.Sorting{
						Field:     sorting,
						Direction: models.SortingDirectionAsc,
					})
				}
			} else if items := strings.Split(sorting, ":"); len(items) == 2 {
				field, order := items[0], items[1]
				if _, ok := models.MapCustomDeliveryDaySortings[field]; ok &&
					(models.SortingDirectionAsc.Equal(order) ||
						models.SortingDirectionDesc.Equal(order)) {
					sortings = append(sortings, models.Sorting{
						Field:     field,
						Direction: models.SortingDirection(order),
					})
				}
			}
		}
		response.WithSortings(sortings...)
	} else {
		response.WithSortings(models.DefaultSortBy)
	}
	limit, _ := utils.ParseInt(c.Query("limit"))
	if _, ok := models.MapLimits[limit]; !ok {
		limit = models.Limits[0]
	}
	urlValues.Set("limit", strconv.FormatInt(limit, 10))
	response.WithLimit(limit)
	page, _ := utils.ParseInt(c.Query("page"))
	page = max(page, 0)
	response.WithPage(page).
		WithUrlValues(urlValues)
	return response, nil
}

func ValidateCustomDeliveryDay(ctx context.Context, c fiber.Ctx) (*models.CustomDeliveryDay, int, error) {
	ctxt := "CustomDeliveryDaySanitizer-ValidateCustomDeliveryDay"
	var response models.CustomDeliveryDay
	err := c.Bind().Body(&response)
	if fiberErr, ok := errors.AsType[*fiber.Error](err); ok {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrBody")
		return &response, fiberErr.Code, err
	}
	if err = (&response).Validate(); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrValidate")
		return &response, fiber.StatusBadRequest, err
	}
	return &response, fiber.StatusOK, nil
}
