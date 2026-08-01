package sanitizers

import (
	"context"
	"errors"
	"log/slog"
	"mime/multipart"
	"net/url"
	"strconv"
	"strings"
	"uuid"

	"github.com/gabriel-vasile/mimetype"
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/utils/v2"
	"github.com/roysitumorang/laukpauk/helper"
	"github.com/roysitumorang/laukpauk/models"
)

func FindAdmins(ctx context.Context, c fiber.Ctx) (*models.AdminFilter, error) {
	ctxt := "AdminSanitizer-FindAdmins"
	response := models.NewAdminFilter()
	originalURL, err := url.ParseRequestURI(c.OriginalURL())
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrParseRequestURI")
		return response, err
	}
	urlValues := originalURL.Query()
	response.WithPaginationURL(helper.Sprintf("%s%s", c.BaseURL(), originalURL.Path))
	if keyword := strings.TrimSpace(c.Query("q")); keyword != "" {
		urlValues.Set("q", keyword)
		response.WithKeyword(keyword)
	}
	if sortBy := urlValues["sort_by"]; len(sortBy) > 0 {
		urlValues.Set("sort_by", c.Query("sort_by"))
		var sortings []models.Sorting
		for _, sorting := range sortBy {
			if !strings.Contains(sorting, ":") {
				if _, ok := models.MapAdminSortings[sorting]; ok {
					sortings = append(sortings, models.Sorting{
						Field:     sorting,
						Direction: models.SortingDirectionAsc,
					})
				}
			} else if items := strings.Split(sorting, ":"); len(items) == 2 {
				field, order := items[0], items[1]
				if _, ok := models.MapAdminSortings[field]; ok &&
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

func ValidateAdmin(ctx context.Context, c fiber.Ctx) (*models.Admin, *multipart.FileHeader, *string, int, error) {
	ctxt := "AdminSanitizer-ValidateAdmin"
	var response models.Admin
	err := c.Bind().Body(&response)
	if fiberErr, ok := errors.AsType[*fiber.Error](err); ok {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrBody")
		return &response, nil, nil, fiberErr.Code, err
	}
	if err = (&response).Validate(); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrValidate")
		return &response, nil, nil, fiber.StatusBadRequest, err
	}
	var mimeType *string
	photo, err := c.FormFile("photo")
	if err == nil {
		file, err := photo.Open()
		if err != nil {
			helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrOpen")
			return &response, nil, nil, fiber.StatusBadRequest, err
		}
		mtype, err := mimetype.DetectReader(file)
		if err != nil {
			helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrDetectReader")
			return &response, nil, nil, fiber.StatusBadRequest, err
		}
		if _, ok := mimeTypes[mtype.String()]; !ok {
			return &response, nil, nil, fiber.StatusBadRequest, errors.New("photo: invalid JPG/PNG/WEBP file")
		}
		mimeType = new(mtype.String())
		response.Photo = helper.Sprintf("%s.webp", uuid.NewV7().String())
	}
	return &response, photo, mimeType, fiber.StatusOK, nil
}
