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
	response map[string]string,
	photo *multipart.FileHeader,
	err error,
) {
	ctxt := "LandingPageSanitizer-ValidateLandingPageCustomAttributes"
	err = c.Bind().Body(&response)
	if _, ok := errors.AsType[*fiber.Error](err); ok {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrBody")
		return nil, nil, err
	}
	switch landingPageName {
	case models.LandingPageNameFeature:
		subject := strings.TrimSpace(response["subject"])
		if subject == "" {
			return nil, nil, errors.New("subject: is required")
		}
		response = map[string]string{"subject": subject}
	case models.LandingPageNameWhy:
		subject := strings.TrimSpace(response["subject"])
		linkText := strings.TrimSpace(response["link_text"])
		linkURL := strings.TrimSpace(response["link_url"])
		response = map[string]string{
			"subject":   subject,
			"link_text": linkText,
			"link_url":  linkURL,
		}
	case models.LandingPageNameCommitment:
		subject := strings.TrimSpace(response["subject"])
		body := strings.TrimSpace(response["body"])
		response = map[string]string{
			"subject": subject,
			"body":    body,
		}
	case models.LandingPageNameFreeShopping:
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
			photoFilename = helper.Sprintf("%s%s", uuid.New().String(), mtype.Extension())
		}
		response = map[string]string{
			"subject": subject,
			"body":    body,
			"photo":   photoFilename,
		}
	case models.LandingPageNameKitchenHero:
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
			photoFilename = helper.Sprintf("%s%s", uuid.New().String(), mtype.Extension())
		}
		response = map[string]string{
			"body":  body,
			"photo": photoFilename,
		}
	}
	return response, photo, nil
}
