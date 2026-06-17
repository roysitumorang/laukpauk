package sanitizers

import (
	"context"
	"errors"
	"fmt"
	"mime/multipart"
	"net/url"
	"strconv"
	"strings"

	"github.com/gabriel-vasile/mimetype"
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/utils/v2"
	"github.com/google/uuid"
	"github.com/roysitumorang/laukpauk/helper"
	"github.com/roysitumorang/laukpauk/models"
	"go.uber.org/zap"
)

func FindWebsiteBlockItems(ctx context.Context, c fiber.Ctx) (*models.WebsiteBlockItemFilter, error) {
	ctxt := "WebsiteBlockItemSanitizer-FindWebsiteBlockItems"
	response := models.NewWebsiteBlockItemFilter()
	originalURL, err := url.ParseRequestURI(helper.ByteSlice2String(c.Request().URI().FullURI()))
	if err != nil {
		helper.Log(ctx, zap.ErrorLevel, err.Error(), ctxt, "ErrParseRequestURI")
		return nil, err
	}
	var sb strings.Builder
	_, _ = sb.WriteString(c.BaseURL())
	_, _ = sb.WriteString(originalURL.Path)
	urlValues := originalURL.Query()
	response = response.WithPaginationURL(sb.String())
	if keyword := strings.TrimSpace(c.Query("q")); keyword != "" {
		urlValues.Set("q", keyword)
		response = response.WithKeyword(keyword)
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

func ValidateWebsiteBlockItem(ctx context.Context, c fiber.Ctx, websiteBlockName models.WebsiteBlockName, websiteBlockItemIDs ...uuid.UUID) (*models.WebsiteBlockItem, *multipart.FileHeader, int, error) {
	ctxt := "WebsiteBlockItemSanitizer-ValidateWebsiteBlockItem"
	var response models.WebsiteBlockItem
	err := c.Bind().Body(&response)
	if fiberErr, ok := errors.AsType[*fiber.Error](err); ok {
		helper.Log(ctx, zap.ErrorLevel, err.Error(), ctxt, "ErrBody")
		return nil, nil, fiberErr.Code, err
	}
	if len(websiteBlockItemIDs) > 0 {
		response.ID = websiteBlockItemIDs[0]
	}
	photo, err := c.FormFile("photo")
	if err == nil {
		file, err := photo.Open()
		if err != nil {
			helper.Log(ctx, zap.ErrorLevel, err.Error(), ctxt, "ErrOpen")
			return &response, nil, fiber.StatusBadRequest, err
		}
		mtype, err := mimetype.DetectReader(file)
		if err != nil {
			helper.Log(ctx, zap.ErrorLevel, err.Error(), ctxt, "ErrDetectReader")
			return &response, nil, fiber.StatusBadRequest, err
		}
		if _, ok := mimeTypes[mtype.String()]; !ok {
			return &response, nil, fiber.StatusBadRequest, errors.New("photo: invalid JPG/PNG file")
		}
		response.Photo = fmt.Sprintf(models.WebsiteBlockHeroPhotoName, helper.GenerateSnowflakeID(), mtype.Extension())
	}
	if err = (&response).Validate(websiteBlockName); err != nil {
		helper.Log(ctx, zap.ErrorLevel, err.Error(), ctxt, "ErrValidate")
		return &response, photo, fiber.StatusBadRequest, err
	}
	return &response, photo, fiber.StatusOK, nil
}
