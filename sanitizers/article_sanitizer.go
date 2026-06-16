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
	"github.com/roysitumorang/laukpauk/helper"
	"github.com/roysitumorang/laukpauk/models"
	"go.uber.org/zap"
)

func FindArticles(ctx context.Context, c fiber.Ctx) (*models.ArticleFilter, error) {
	ctxt := "ArticleSanitizer-FindArticles"
	response := models.NewArticleFilter()
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

func ValidateArticle(ctx context.Context, c fiber.Ctx) (*models.Article, *multipart.FileHeader, int, error) {
	ctxt := "ArticleSanitizer-ValidateArticle"
	var response models.Article
	err := c.Bind().Body(&response)
	if fiberErr, ok := errors.AsType[*fiber.Error](err); ok {
		helper.Log(ctx, zap.ErrorLevel, err.Error(), ctxt, "ErrBody")
		return nil, nil, fiberErr.Code, err
	}
	if err = (&response).Validate(); err != nil {
		helper.Log(ctx, zap.ErrorLevel, err.Error(), ctxt, "ErrValidate")
		return &response, nil, fiber.StatusBadRequest, err
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
		response.Photo = fmt.Sprintf(models.ArticlePhotoName, helper.GenerateSnowflakeID(), mtype.Extension())
	}
	return &response, photo, fiber.StatusOK, nil
}
