package sanitizers

import (
	"context"
	"errors"
	"fmt"
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

func FindLandingPageBlocks(ctx context.Context, c fiber.Ctx) (*models.LandingPageBlockFilter, error) {
	ctxt := "LandingPageBlockSanitizer-FindLandingPageBlocks"
	response := models.NewLandingPageBlockFilter()
	originalURL, err := url.ParseRequestURI(helper.ByteSlice2String(c.Request().URI().FullURI()))
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrParseRequestURI")
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

func ValidateLandingPageBlock(ctx context.Context, c fiber.Ctx, parentName *models.LandingPageBlockName, landingPageBlockIDs ...uuid.UUID) (*models.LandingPageBlock, *multipart.FileHeader, int, error) {
	ctxt := "LandingPageBlockSanitizer-ValidateLandingPageBlock"
	var response models.LandingPageBlock
	err := c.Bind().Body(&response)
	if fiberErr, ok := errors.AsType[*fiber.Error](err); ok {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrBody")
		return nil, nil, fiberErr.Code, err
	}
	response.ParentName = parentName
	if len(landingPageBlockIDs) > 0 {
		response.ID = landingPageBlockIDs[0]
	}
	var photo *multipart.FileHeader
	if parentName != nil {
		if photo, err = c.FormFile("photo"); err == nil {
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
			response.Photo = new(fmt.Sprintf(models.LandingPageBlockPhotoName, *parentName, helper.GenerateSnowflakeID(), mtype.Extension()))
		}
	}
	if err = (&response).Validate(); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrValidate")
		return &response, photo, fiber.StatusBadRequest, err
	}
	return &response, photo, fiber.StatusOK, nil
}

func ValidateLandingPageBlockCustomAttributes(
	ctx context.Context,
	c fiber.Ctx,
	name models.LandingPageBlockName,
) (
	response map[string]string,
	photo *multipart.FileHeader,
	err error,
) {
	ctxt := "LandingPageBlockSanitizer-ValidateLandingPageBlockCustomAttributes"
	err = c.Bind().Body(&response)
	if _, ok := errors.AsType[*fiber.Error](err); ok {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrBody")
		return nil, nil, err
	}
	switch name {
	case models.LandingPageBlockNameFeature:
		subject := strings.TrimSpace(response["subject"])
		response = map[string]string{"subject": subject}
	case models.LandingPageBlockNameWhy:
		subject := strings.TrimSpace(response["subject"])
		linkText := strings.TrimSpace(response["link_text"])
		linkURL := strings.TrimSpace(response["link_url"])
		response = map[string]string{
			"subject":   subject,
			"link_text": linkText,
			"link_url":  linkURL,
		}
	case models.LandingPageBlockNameCommitment:
		subject := strings.TrimSpace(response["subject"])
		body := strings.TrimSpace(response["body"])
		response = map[string]string{
			"subject": subject,
			"body":    body,
		}
	case models.LandingPageBlockNameFreeShopping:
		subject := strings.TrimSpace(response["subject"])
		body := strings.TrimSpace(response["body"])
		var photoFilename string
		if photo, err = c.FormFile("photo"); err == nil {
			file, err := photo.Open()
			if err != nil {
				helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrOpen")
				return response, nil, err
			}
			mtype, err := mimetype.DetectReader(file)
			if err != nil {
				helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrDetectReader")
				return response, nil, err
			}
			if _, ok := mimeTypes[mtype.String()]; !ok {
				return response, nil, errors.New("photo: invalid JPG/PNG file")
			}
			photoFilename = fmt.Sprintf(models.LandingPageBlockPhotoName, name, helper.GenerateSnowflakeID(), mtype.Extension())
		}
		response = map[string]string{
			"subject": subject,
			"body":    body,
			"photo":   photoFilename,
		}
	case models.LandingPageBlockNameKitchenHero:
		body := strings.TrimSpace(response["body"])
		var photoFilename string
		if photo, err = c.FormFile("photo"); err == nil {
			file, err := photo.Open()
			if err != nil {
				helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrOpen")
				return response, nil, err
			}
			mtype, err := mimetype.DetectReader(file)
			if err != nil {
				helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrDetectReader")
				return response, nil, err
			}
			if _, ok := mimeTypes[mtype.String()]; !ok {
				return response, nil, errors.New("photo: invalid JPG/PNG file")
			}
			photoFilename = fmt.Sprintf(models.LandingPageBlockPhotoName, name, helper.GenerateSnowflakeID(), mtype.Extension())
		}
		response = map[string]string{
			"body":  body,
			"photo": photoFilename,
		}
	}
	return response, photo, nil
}
