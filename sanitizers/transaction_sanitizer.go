package sanitizers

import (
	"context"
	"errors"
	"log/slog"
	"net/url"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/utils/v2"
	"github.com/roysitumorang/laukpauk/helper"
	"github.com/roysitumorang/laukpauk/models"
)

func FindTransactions(ctx context.Context, c fiber.Ctx) (*models.TransactionFilter, error) {
	ctxt := "TransactionSanitizer-FindTransactions"
	response := models.NewTransactionFilter()
	originalURL, err := url.ParseRequestURI(c.OriginalURL())
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrParseRequestURI")
		return nil, err
	}
	urlValues := originalURL.Query()
	response = response.WithPaginationURL(helper.Sprintf("%s%s", c.BaseURL(), originalURL.Path))
	if keyword := strings.TrimSpace(c.Query("q")); keyword != "" {
		urlValues.Set("q", keyword)
		response = response.WithKeyword(keyword)
	}
	if sortBy := urlValues["sort_by"]; len(sortBy) > 0 {
		urlValues.Set("sort_by", c.Query("sort_by"))
		var sortings []models.Sorting
		for _, sorting := range sortBy {
			if !strings.Contains(sorting, ":") {
				if _, ok := models.MapTransactionSortings[sorting]; ok {
					sortings = append(sortings, models.Sorting{
						Field:     sorting,
						Direction: models.SortingDirectionAsc,
					})
				}
			} else if items := strings.Split(sorting, ":"); len(items) == 2 {
				field, order := items[0], items[1]
				if _, ok := models.MapTransactionSortings[field]; ok &&
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
	response = response.WithLimit(limit)
	page, _ := utils.ParseInt(c.Query("page"))
	page = max(page, 0)
	response = response.WithPage(page).
		WithUrlValues(urlValues)
	return response, nil
}

func ValidateTransaction(ctx context.Context, c fiber.Ctx) (*models.Transaction, int, error) {
	ctxt := "TransactionSanitizer-ValidateTransaction"
	var response models.Transaction
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

func ValidateTransactionPayment(ctx context.Context, c fiber.Ctx) (*models.TransactionPaymentWebhookNotification, int, error) {
	ctxt := "TransactionSanitizer-ValidateTransactionPayment"
	var response models.TransactionPaymentWebhookNotification
	err := c.Bind().Body(&response)
	if fiberErr, ok := errors.AsType[*fiber.Error](err); ok {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrBody")
		return &response, fiberErr.Code, err
	}
	return &response, fiber.StatusOK, nil
}

func ValidateTransactionWithdrawal(ctx context.Context, c fiber.Ctx) (*models.TransactionWithdrawal, int, error) {
	ctxt := "TransactionSanitizer-ValidateTransactionWithdrawal"
	var response models.TransactionWithdrawal
	err := c.Bind().Body(&response)
	if fiberErr, ok := errors.AsType[*fiber.Error](err); ok {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrBody")
		return &response, fiberErr.Code, err
	}
	return &response, fiber.StatusOK, nil
}

func ValidateTransactionShipment(ctx context.Context, c fiber.Ctx) (*models.TransactionShipment, int, error) {
	ctxt := "TransactionSanitizer-ValidateTransactionShipment"
	var response models.TransactionShipment
	err := c.Bind().Body(&response)
	if fiberErr, ok := errors.AsType[*fiber.Error](err); ok {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrBody")
		return &response, fiberErr.Code, err
	}
	return &response, fiber.StatusOK, nil
}
