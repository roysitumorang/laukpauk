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

func FindLandingPageItems(ctx context.Context, c fiber.Ctx) (*models.LandingPageItemFilter, error) {
	ctxt := "LandingPageItemSanitizer-FindLandingPageItems"
	response := models.NewLandingPageItemFilter()
	originalURL, err := url.ParseRequestURI(c.OriginalURL())
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrParseRequestURI")
		return nil, err
	}
	urlValues := originalURL.Query()
	response.WithPaginationURL(helper.Sprintf("%s%s", c.BaseURL(), originalURL.Path))
	if keyword := strings.TrimSpace(c.Query("q")); keyword != "" {
		urlValues.Set("q", keyword)
		response.WithKeyword(keyword)
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

func ValidateLandingPageItem(
	ctx context.Context,
	c fiber.Ctx,
	landingPageName models.LandingPageName,
) (
	*models.LandingPageItem,
	*multipart.FileHeader,
	int,
	error,
) {
	ctxt := "LandingPageItemSanitizer-ValidateLandingPageItem"
	var response models.LandingPageItem
	err := c.Bind().Body(&response)
	if fiberErr, ok := errors.AsType[*fiber.Error](err); ok {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrBody")
		return nil, nil, fiberErr.Code, err
	}
	response.LandingPageName = landingPageName
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
		response.Photo = new(helper.Sprintf("%s%s", uuid.New().String(), mtype.Extension()))
	}
	if err = (&response).Validate(); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrValidate")
		return &response, photo, fiber.StatusBadRequest, err
	}
	return &response, photo, fiber.StatusOK, nil
}
