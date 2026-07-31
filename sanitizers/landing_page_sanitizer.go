package sanitizers

import (
	"context"
	"errors"
	"log/slog"
	"mime/multipart"
	"strings"
	"uuid"

	"github.com/gabriel-vasile/mimetype"
	"github.com/gofiber/fiber/v3"
	"github.com/roysitumorang/laukpauk/helper"
	"github.com/roysitumorang/laukpauk/models"
)

func ValidateLandingPageCustomAttributes(
	ctx context.Context,
	c fiber.Ctx,
	landingPageName models.LandingPageName,
) (
	map[string]string,
	*multipart.FileHeader,
	*string,
	error,
) {
	ctxt := "LandingPageSanitizer-ValidateLandingPageCustomAttributes"
	var response map[string]string
	err := c.Bind().Body(&response)
	if _, ok := errors.AsType[*fiber.Error](err); ok {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrBody")
		return response, nil, nil, err
	}
	var photo *multipart.FileHeader
	var mimeType *string
	switch landingPageName {
	case models.LandingPageNameFeature:
		subject := strings.TrimSpace(response["subject"])
		if subject == "" {
			return response, nil, nil, errors.New("subject: is required")
		}
		response = map[string]string{"subject": subject}
	case models.LandingPageNameWhy:
		subject := strings.TrimSpace(response["subject"])
		if subject == "" {
			return response, nil, nil, errors.New("subject: is required")
		}
		linkText := strings.TrimSpace(response["link_text"])
		if linkText == "" {
			return response, nil, nil, errors.New("link_text: is required")
		}
		linkURL := strings.TrimSpace(response["link_url"])
		if linkURL == "" {
			return response, nil, nil, errors.New("link_url: is required")
		}
		response = map[string]string{
			"subject":   subject,
			"link_text": linkText,
			"link_url":  linkURL,
		}
	case models.LandingPageNameCommitment:
		subject := strings.TrimSpace(response["subject"])
		if subject == "" {
			return response, nil, nil, errors.New("subject: is required")
		}
		body := strings.TrimSpace(response["body"])
		if body == "" {
			return response, nil, nil, errors.New("body: is required")
		}
		response = map[string]string{
			"subject": subject,
			"body":    body,
		}
	case models.LandingPageNameFreeShopping:
		subject := strings.TrimSpace(response["subject"])
		if subject == "" {
			return response, nil, nil, errors.New("subject: is required")
		}
		body := strings.TrimSpace(response["body"])
		if body == "" {
			return response, nil, nil, errors.New("body: is required")
		}
		var photoFilename string
		if photo, err = c.FormFile("photo"); err == nil {
			file, err := photo.Open()
			if err != nil {
				helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrOpen")
				return response, nil, nil, err
			}
			mtype, err := mimetype.DetectReader(file)
			if err != nil {
				helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrDetectReader")
				return response, nil, nil, err
			}
			if _, ok := mimeTypes[mtype.String()]; !ok {
				return response, nil, nil, errors.New("photo: invalid JPG/PNG/WEBP file")
			}
			mimeType = new(mtype.String())
			photoFilename = helper.Sprintf("%s.webp", uuid.NewV7().String())
		}
		response = map[string]string{
			"subject": subject,
			"body":    body,
			"photo":   photoFilename,
		}
	case models.LandingPageNameKitchenHero:
		body := strings.TrimSpace(response["body"])
		if body == "" {
			return response, nil, nil, errors.New("body: is required")
		}
		var photoFilename string
		if photo, err = c.FormFile("photo"); err == nil {
			file, err := photo.Open()
			if err != nil {
				helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrOpen")
				return response, nil, nil, err
			}
			mtype, err := mimetype.DetectReader(file)
			if err != nil {
				helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrDetectReader")
				return response, nil, nil, err
			}
			if _, ok := mimeTypes[mtype.String()]; !ok {
				return response, nil, nil, errors.New("photo: invalid JPG/PNG/WEBP file")
			}
			mimeType = new(mtype.String())
			photoFilename = helper.Sprintf("%s.webp", uuid.NewV7().String())
		}
		response = map[string]string{
			"body":  body,
			"photo": photoFilename,
		}
	}
	return response, photo, mimeType, nil
}
