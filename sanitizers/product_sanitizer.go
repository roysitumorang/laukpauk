package sanitizers

import (
	"context"
	"errors"
	"log/slog"
	"mime/multipart"
	"net/url"
	"strconv"
	"strings"

	"github.com/gabriel-vasile/mimetype"
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/utils/v2"
	"github.com/roysitumorang/laukpauk/helper"
	"github.com/roysitumorang/laukpauk/models"
)

func FindProducts(ctx context.Context, c fiber.Ctx) (*models.ProductFilter, error) {
	ctxt := "ProductSanitizer-FindProducts"
	response := models.NewProductFilter()
	originalURL, err := url.ParseRequestURI(c.OriginalURL())
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrParseRequestURI")
		return nil, err
	}
	var sb strings.Builder
	_, _ = sb.WriteString(c.BaseURL())
	_, _ = sb.WriteString(originalURL.Path)
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
				if _, ok := models.MapProductSortings[sorting]; ok {
					sortings = append(sortings, models.Sorting{
						Field:     sorting,
						Direction: models.SortingDirectionAsc,
					})
				}
			} else if items := strings.Split(sorting, ":"); len(items) == 2 {
				field, order := items[0], items[1]
				if _, ok := models.MapProductSortings[field]; ok &&
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

func ValidateProduct(ctx context.Context, c fiber.Ctx) (*models.Product, *multipart.FileHeader, int, error) {
	ctxt := "ProductSanitizer-ValidateProduct"
	var response models.Product
	err := c.Bind().Body(&response)
	if fiberErr, ok := errors.AsType[*fiber.Error](err); ok {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrBody")
		return nil, nil, fiberErr.Code, err
	}
	if err = (&response).Validate(); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrValidate")
		return &response, nil, fiber.StatusBadRequest, err
	}
	photo, err := c.FormFile("photo")
	if err == nil {
		file, err := photo.Open()
		if err != nil {
			helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrOpen")
			return &response, nil, fiber.StatusBadRequest, err
		}
		mtype, err := mimetype.DetectReader(file)
		if err != nil {
			helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrDetectReader")
			return &response, nil, fiber.StatusBadRequest, err
		}
		if _, ok := mimeTypes[mtype.String()]; !ok {
			return &response, nil, fiber.StatusBadRequest, errors.New("photo: invalid JPG/PNG file")
		}
		response.Photo = helper.Sprintf(models.ProductPhotoName, helper.GenerateSnowflakeID(), mtype.Extension())
	}
	return &response, photo, fiber.StatusOK, nil
}
