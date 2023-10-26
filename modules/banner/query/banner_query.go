package query

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/roysitumorang/laukpauk/helper"
	"github.com/roysitumorang/laukpauk/modules/banner/model"
	"go.uber.org/zap"
)

type (
	bannerQuery struct {
		dbRead, dbWrite *pgxpool.Pool
	}
)

func NewBannerQuery(
	dbRead,
	dbWrite *pgxpool.Pool,
) BannerQuery {
	return &bannerQuery{
		dbRead:  dbRead,
		dbWrite: dbWrite,
	}
}

func (q *bannerQuery) FindBanners(ctx context.Context) (response []model.Banner, err error) {
	ctxt := "BannerQuery-FindBanners"
	response = []model.Banner{}
	rows, err := q.dbRead.Query(
		ctx,
		`SELECT
			id
			, user_id
			, published
			, file
			, created_by
			, created_at
			, updated_by
			, updated_at
		FROM banners
		WHERE parent_id IS NULL
		AND published
		ORDER BY -id`,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		err = nil
	}
	if err != nil {
		helper.Capture(ctx, zap.ErrorLevel, err, ctxt, "ErrQuery")
		return
	}
	defer rows.Close()
	var (
		params       []interface{}
		placeholders []string
		mapIndex     = map[int64]int{}
	)
	for rows.Next() {
		var banner model.Banner
		if err = rows.Scan(
			&banner.ID,
			&banner.UserID,
			&banner.Published,
			&banner.File,
			&banner.CreatedBy,
			&banner.CreatedAt,
			&banner.UpdatedBy,
			&banner.UpdatedAt,
		); err != nil {
			helper.Capture(ctx, zap.ErrorLevel, err, ctxt, "ErrScan")
			return
		}
		params = append(params, banner.ID)
		placeholders = append(placeholders, fmt.Sprintf("$%d", len(params)))
		response = append(response, banner)
		mapIndex[banner.ID] = len(params) - 1
	}
	if err = rows.Err(); err != nil {
		helper.Capture(ctx, zap.ErrorLevel, err, ctxt, "ErrErr")
		return
	}
	if len(params) == 0 {
		return
	}
	rows, err = q.dbRead.Query(
		ctx,
		fmt.Sprintf(
			`SELECT
				id
				, parent_id
				, user_id
				, published
				, file
				, created_by
				, created_at
				, updated_by
				, updated_at
			FROM banners
			WHERE parent_id IN (%s)
			AND published
			ORDER BY -id`,
			strings.Join(placeholders, ","),
		),
		params...,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		err = nil
	}
	if err != nil {
		helper.Capture(ctx, zap.ErrorLevel, err, ctxt, "ErrQuery")
		return
	}
	defer rows.Close()
	for rows.Next() {
		var thumbnail model.Banner
		if err = rows.Scan(
			&thumbnail.ID,
			&thumbnail.ParentID,
			&thumbnail.UserID,
			&thumbnail.Published,
			&thumbnail.File,
			&thumbnail.CreatedBy,
			&thumbnail.CreatedAt,
			&thumbnail.UpdatedBy,
			&thumbnail.UpdatedAt,
		); err != nil {
			helper.Capture(ctx, zap.ErrorLevel, err, ctxt, "ErrScan")
			return
		}
		if thumbnail.ParentID != nil {
			index := mapIndex[*thumbnail.ParentID]
			banner := response[index]
			banner.Thumbnails = append(banner.Thumbnails, thumbnail)
			response[index] = banner
		}
	}
	if err = rows.Err(); err != nil {
		helper.Capture(ctx, zap.ErrorLevel, err, ctxt, "ErrErr")
	}
	return
}
