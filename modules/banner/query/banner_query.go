package query

import (
	"context"
	"errors"

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
			, thumbnails
			, created_by
			, created_at
			, updated_by
			, updated_at
		FROM banners
		WHERE published = 1
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
	for rows.Next() {
		var banner model.Banner
		if err = rows.Scan(
			&banner.ID,
			&banner.UserID,
			&banner.Published,
			&banner.File,
			&banner.Thumbnails,
			&banner.CreatedBy,
			&banner.CreatedAt,
			&banner.UpdatedBy,
			&banner.UpdatedAt,
		); err != nil {
			helper.Capture(ctx, zap.ErrorLevel, err, ctxt, "ErrScan")
			return
		}
		response = append(response, banner)
	}
	if err = rows.Err(); err != nil {
		helper.Capture(ctx, zap.ErrorLevel, err, ctxt, "ErrErr")
	}
	return
}
