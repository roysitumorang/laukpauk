package repositories

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"

	"github.com/govalues/decimal"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/roysitumorang/laukpauk/helper"
	"github.com/roysitumorang/laukpauk/models"
	"github.com/roysitumorang/laukpauk/pools"
)

type (
	LandingPageItemRepository interface {
		FindLandingPageItems(
			ctx context.Context,
			request *models.LandingPageItemFilter,
		) ([]*models.LandingPageItem, int64, int64, error)
		FindLandingPageItem(
			ctx context.Context,
			request *models.LandingPageItemFilter,
		) (*models.LandingPageItem, error)
		CreateLandingPageItem(
			ctx context.Context,
			tx pgx.Tx,
			request *models.LandingPageItem,
		) (*models.LandingPageItem, error)
		UpdateLandingPageItem(
			ctx context.Context,
			tx pgx.Tx,
			request *models.LandingPageItem,
		) (*models.LandingPageItem, error)
	}

	landingPageItemRepository struct {
		dbRead,
		dbWrite *pgxpool.Pool
	}
)

func NewLandingPageItemRepository(
	dbRead,
	dbWrite *pgxpool.Pool,
) LandingPageItemRepository {
	return &landingPageItemRepository{
		dbRead:  dbRead,
		dbWrite: dbWrite,
	}
}

func (q *landingPageItemRepository) FindLandingPageItems(
	ctx context.Context,
	request *models.LandingPageItemFilter,
) ([]*models.LandingPageItem, int64, int64, error) {
	ctxt := "LandingPageItemRepository-FindLandingPageItems"
	sb := pools.BuilderPool.Get()
	defer func() {
		sb.Reset()
		pools.BuilderPool.Put(sb)
	}()
	sb.Reset()
	condition, params := q.buildCondition(request, sb)
	sb.Reset()
	_, _ = fmt.Fprintf(sb, "SELECT COUNT(1) FROM landing_page_items %s", condition)
	query := sb.String()
	var total int64
	err := q.dbRead.QueryRow(ctx, query, params...).Scan(&total)
	if err != nil {
		helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
		return nil, 0, 0, err
	}
	if total == 0 {
		return nil, 0, 0, nil
	}
	query = strings.ReplaceAll(
		query,
		"COUNT(1)",
		`id
		, record_status
		, block_name
		, subject
		, body
		, link_text
		, link_url
		, photo
		, created_by
		, created_at
		, updated_by
		, updated_at
		, deleted_by
		, deleted_at
		, deletion_reason`,
	)
	sb.Reset()
	_, _ = sb.WriteString(query)
	pages := int64(1)
	var offset int64
	if request.Limit > 0 {
		totalDecimal, err := decimal.New(total, 0)
		if err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrNew")
			return nil, 0, 0, err
		}
		perPageDecimal, err := decimal.New(request.Limit, 0)
		if err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrNew")
			return nil, 0, 0, err
		}
		pagesDecimal, err := totalDecimal.Quo(perPageDecimal)
		if err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrQuo")
			return nil, 0, 0, err
		}
		pages, _, _ = pagesDecimal.Ceil(0).Int64(0)
		offset = request.Page * request.Limit
		_, _ = fmt.Fprintf(sb, " LIMIT %d OFFSET %d", request.Limit, offset)
	}
	rows, err := q.dbRead.Query(ctx, sb.String(), params...)
	if errors.Is(err, pgx.ErrNoRows) {
		err = nil
	}
	if err != nil {
		helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrQuery")
		return nil, 0, 0, err
	}
	defer rows.Close()
	var response []*models.LandingPageItem
	for rows.Next() {
		var landingPageItem models.LandingPageItem
		if err = rows.Scan(
			&landingPageItem.ID,
			&landingPageItem.RecordStatus,
			&landingPageItem.LandingPageName,
			&landingPageItem.Subject,
			&landingPageItem.Body,
			&landingPageItem.LinkText,
			&landingPageItem.LinkURL,
			&landingPageItem.Photo,
			&landingPageItem.CreatedBy,
			&landingPageItem.CreatedAt,
			&landingPageItem.UpdatedBy,
			&landingPageItem.UpdatedAt,
			&landingPageItem.DeletedBy,
			&landingPageItem.DeletedAt,
			&landingPageItem.DeletionReason,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrScan")
			return nil, 0, 0, err
		}
		offset++
		landingPageItem.RowNo = offset
		response = append(response, &landingPageItem)
	}
	return response, total, pages, nil
}

func (q *landingPageItemRepository) FindLandingPageItem(
	ctx context.Context,
	request *models.LandingPageItemFilter,
) (*models.LandingPageItem, error) {
	ctxt := "LandingPageItemRepository-FindLandingPageItem"
	sb := pools.BuilderPool.Get()
	defer func() {
		sb.Reset()
		pools.BuilderPool.Put(sb)
	}()
	sb.Reset()
	condition, params := q.buildCondition(request, sb)
	sb.Reset()
	_, _ = fmt.Fprintf(
		sb,
		`SELECT
			id
			, record_status
			, block_name
			, subject
			, body
			, link_text
			, link_url
			, photo
			, created_by
			, created_at
			, updated_by
			, updated_at
			, deleted_by
			, deleted_at
			, deletion_reason
		FROM landing_page_items %s`,
		condition,
	)
	var response models.LandingPageItem
	err := q.dbRead.QueryRow(ctx, sb.String(), params...).Scan(
		&response.ID,
		&response.RecordStatus,
		&response.LandingPageName,
		&response.Subject,
		&response.Body,
		&response.LinkText,
		&response.LinkURL,
		&response.Photo,
		&response.CreatedBy,
		&response.CreatedAt,
		&response.UpdatedBy,
		&response.UpdatedAt,
		&response.DeletedBy,
		&response.DeletedAt,
		&response.DeletionReason,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, models.ErrLandingPageItemNotFound
	}
	if err != nil {
		helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrScan")
		return nil, err
	}
	return &response, nil
}

func (q *landingPageItemRepository) buildCondition(request *models.LandingPageItemFilter, sb *strings.Builder) (condition string, params []any) {
	var conds []string
	if len(request.LandingPageItemIDs) > 0 {
		sb.Reset()
		_, _ = sb.WriteString("id IN (")
		for i, landingPageItemID := range request.LandingPageItemIDs {
			params = append(params, landingPageItemID)
			if i > 0 {
				_, _ = sb.WriteString(",")
			}
			_, _ = fmt.Fprintf(sb, "$%d", len(params))
		}
		_, _ = sb.WriteString(")")
		conds = append(conds, sb.String())
	}
	if len(request.LandingPageNames) > 0 {
		sb.Reset()
		_, _ = sb.WriteString("landing_page_name IN (")
		for i, landingPageName := range request.LandingPageNames {
			params = append(params, landingPageName)
			if i > 0 {
				_, _ = sb.WriteString(",")
			}
			_, _ = fmt.Fprintf(sb, "$%d", len(params))
		}
		_, _ = sb.WriteString(")")
		conds = append(conds, sb.String())
	}
	if len(request.RecordStatus) > 0 {
		sb.Reset()
		_, _ = sb.WriteString("record_status IN (")
		for i, recordStatus := range request.RecordStatus {
			params = append(params, recordStatus)
			if i > 0 {
				_, _ = sb.WriteString(",")
			}
			_, _ = fmt.Fprintf(sb, "$%d", len(params))
		}
		_, _ = sb.WriteString(")")
		conds = append(conds, sb.String())
	}
	if request.Keyword != "" {
		sb.Reset()
		_, _ = fmt.Fprintf(sb, "%%%s%%", strings.ToLower(request.Keyword))
		params = append(params, sb.String())
		n := len(params)
		sb.Reset()
		_, _ = fmt.Fprintf(sb, "(LOWER(subject) LIKE $%d OR LOWER(body) LIKE $%d)", n, n)
		conds = append(conds, sb.String())
	}
	if len(conds) > 0 {
		sb.Reset()
		_, _ = sb.WriteString("WHERE")
		for i, cond := range conds {
			if i > 0 {
				_, _ = sb.WriteString(" AND")
			}
			_, _ = fmt.Fprintf(sb, " %s", cond)
		}
		condition = sb.String()
	}
	return
}

func (q *landingPageItemRepository) CreateLandingPageItem(
	ctx context.Context,
	tx pgx.Tx,
	request *models.LandingPageItem,
) (*models.LandingPageItem, error) {
	ctxt := "LandingPageItemRepository-CreateLandingPageItem"
	var response models.LandingPageItem
	if err := tx.QueryRow(
		ctx,
		`INSERT INTO block_items (
			record_status
			, block_name
			, subject
			, body
			, link_text
			, link_url
			, photo
			, created_by
			, created_at
			, updated_by
			, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		RETURNING id
			, record_status
			, block_name
			, subject
			, body
			, link_text
			, link_url
			, photo
			, created_by
			, created_at
			, updated_by
			, updated_at
			, deleted_by
			, deleted_at
			, deletion_reason`,
		request.RecordStatus,
		request.LandingPageName,
		request.Subject,
		request.Body,
		request.LinkText,
		request.LinkURL,
		request.Photo,
		request.CreatedBy,
		request.CreatedAt,
		request.UpdatedBy,
		request.UpdatedAt,
	).Scan(
		&response.ID,
		&response.RecordStatus,
		&response.LandingPageName,
		&response.Subject,
		&response.Body,
		&response.LinkText,
		&response.LinkURL,
		&response.Photo,
		&response.CreatedBy,
		&response.CreatedAt,
		&response.UpdatedBy,
		&response.UpdatedAt,
		&response.DeletedBy,
		&response.DeletedAt,
		&response.DeletionReason,
	); err != nil {
		helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrScan")
		return nil, err
	}
	return &response, nil
}

func (q *landingPageItemRepository) UpdateLandingPageItem(
	ctx context.Context,
	tx pgx.Tx,
	request *models.LandingPageItem,
) (*models.LandingPageItem, error) {
	ctxt := "LandingPageItemRepository-UpdateLandingPageItem"
	var response models.LandingPageItem
	err := tx.QueryRow(
		ctx,
		`SELECT
			id
			, record_status
			, block_name
			, subject
			, body
			, link_text
			, link_url
			, photo
			, created_by
			, created_at
			, updated_by
			, updated_at
			, deleted_by
			, deleted_at
			, deletion_reason
		FROM landing_page_items
		WHERE id = $1
		FOR NO KEY UPDATE`,
		request.ID,
	).Scan(
		&response.ID,
		&response.RecordStatus,
		&response.LandingPageName,
		&response.Subject,
		&response.Body,
		&response.LinkText,
		&response.LinkURL,
		&response.Photo,
		&response.CreatedBy,
		&response.CreatedAt,
		&response.UpdatedBy,
		&response.UpdatedAt,
		&response.DeletedBy,
		&response.DeletedAt,
		&response.DeletionReason,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, models.ErrLandingPageItemNotFound
	}
	if err != nil {
		helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrScan")
		return nil, err
	}
	var params []any
	var n int
	sb := pools.BufferPool.Get()
	defer func() {
		sb.Reset()
		pools.BufferPool.Put(sb)
	}()
	sb.Reset()
	_, _ = sb.WriteString("UPDATE landing_page_items SET")
	if response.RecordStatus != request.RecordStatus {
		params = append(params, request.RecordStatus)
		if n = len(params); n > 1 {
			_, _ = sb.WriteString(",")
		}
		_, _ = fmt.Fprintf(sb, " record_status = $%d", n)
	}
	if response.Subject != request.Subject {
		params = append(params, request.Subject)
		if n = len(params); n > 1 {
			_, _ = sb.WriteString(",")
		}
		_, _ = fmt.Fprintf(sb, " subject = $%d", n)
	}
	if response.Body != request.Body {
		params = append(params, request.Body)
		if n = len(params); n > 1 {
			_, _ = sb.WriteString(",")
		}
		_, _ = fmt.Fprintf(sb, " body = $%d", n)
	}
	if (response.LinkText == nil && request.LinkText != nil) ||
		(response.LinkText != nil && request.LinkText == nil) ||
		(response.LinkText != nil && request.LinkText != nil &&
			*response.LinkText != *request.LinkText) {
		params = append(params, request.LinkText)
		if n = len(params); n > 1 {
			_, _ = sb.WriteString(",")
		}
		_, _ = fmt.Fprintf(sb, " link_text = $%d", n)
	}
	if (response.LinkURL == nil && request.LinkURL != nil) ||
		(response.LinkURL != nil && request.LinkURL == nil) ||
		(response.LinkURL != nil && request.LinkURL != nil &&
			*response.LinkURL != *request.LinkURL) {
		params = append(params, request.LinkURL)
		if n = len(params); n > 1 {
			_, _ = sb.WriteString(",")
		}
		_, _ = fmt.Fprintf(sb, " link_url = $%d", n)
	}
	if (response.Photo == nil && request.Photo != nil) ||
		(response.Photo != nil && request.Photo == nil) ||
		(response.Photo != nil && request.Photo != nil &&
			*response.Photo != *request.Photo) {
		params = append(params, request.Photo)
		if n = len(params); n > 1 {
			_, _ = sb.WriteString(",")
		}
		_, _ = fmt.Fprintf(sb, " photo = $%d", n)
	}
	if len(params) == 0 {
		return &response, nil
	}
	if response.UpdatedBy.Compare(request.UpdatedBy) != 0 {
		params = append(params, request.UpdatedBy)
		if n = len(params); n > 1 {
			_, _ = sb.WriteString(",")
		}
		_, _ = fmt.Fprintf(sb, " updated_by = $%d", n)
	}
	if !response.UpdatedAt.Equal(request.UpdatedAt) {
		params = append(params, request.UpdatedAt)
		if n = len(params); n > 1 {
			_, _ = sb.WriteString(",")
		}
		_, _ = fmt.Fprintf(sb, " updated_at = $%d", n)
	}
	if response.DeletedBy == nil && request.DeletedBy != nil {
		params = append(params, request.DeletedBy)
		if n = len(params); n > 1 {
			_, _ = sb.WriteString(",")
		}
		_, _ = fmt.Fprintf(sb, " deleted_by = $%d", n)
	}
	if response.DeletedAt == nil && request.DeletedAt != nil {
		params = append(params, request.DeletedAt)
		if n = len(params); n > 1 {
			_, _ = sb.WriteString(",")
		}
		_, _ = fmt.Fprintf(sb, " deleted_at = $%d", n)
	}
	if response.DeletionReason == nil && request.DeletionReason != nil {
		params = append(params, request.DeletionReason)
		if n = len(params); n > 1 {
			_, _ = sb.WriteString(",")
		}
		_, _ = fmt.Fprintf(sb, " deletion_reason = $%d", n)
	}
	params = append(params, request.ID)
	_, _ = fmt.Fprintf(
		sb,
		` WHERE id = $%d
		RETURNING id
			, record_status
			, block_name
			, subject
			, body
			, link_text
			, link_url
			, photo
			, created_by
			, created_at
			, updated_by
			, updated_at
			, deleted_by
			, deleted_at
			, deletion_reason`,
		len(params),
	)
	if err = tx.QueryRow(ctx, sb.String(), params...).Scan(
		&response.ID,
		&response.RecordStatus,
		&response.LandingPageName,
		&response.Subject,
		&response.Body,
		&response.LinkText,
		&response.LinkURL,
		&response.Photo,
		&response.CreatedBy,
		&response.CreatedAt,
		&response.UpdatedBy,
		&response.UpdatedAt,
		&response.DeletedBy,
		&response.DeletedAt,
		&response.DeletionReason,
	); err != nil {
		helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrScan")
		return nil, err
	}
	return &response, nil
}
