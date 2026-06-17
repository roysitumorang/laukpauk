package repositories

import (
	"context"
	"errors"
	"strconv"
	"strings"

	"github.com/govalues/decimal"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/roysitumorang/laukpauk/helper"
	"github.com/roysitumorang/laukpauk/models"
	"go.uber.org/zap"
)

type (
	WebsiteItemRepository interface {
		FindWebsiteItems(
			ctx context.Context,
			request *models.WebsiteItemFilter,
		) ([]*models.WebsiteItem, int64, int64, error)
		CreateWebsiteItem(
			ctx context.Context,
			tx pgx.Tx,
			request *models.WebsiteItem,
		) (*models.WebsiteItem, error)
		UpdateWebsiteItem(
			ctx context.Context,
			tx pgx.Tx,
			request *models.WebsiteItem,
		) (*models.WebsiteItem, error)
	}

	websiteItemRepository struct {
		dbRead,
		dbWrite *pgxpool.Pool
	}
)

func NewWebsiteItemRepository(
	dbRead,
	dbWrite *pgxpool.Pool,
) WebsiteItemRepository {
	return new(websiteItemRepository{
		dbRead:  dbRead,
		dbWrite: dbWrite,
	})
}

func (q *websiteItemRepository) FindWebsiteItems(
	ctx context.Context,
	request *models.WebsiteItemFilter,
) ([]*models.WebsiteItem, int64, int64, error) {
	ctxt := "WebsiteItemRepository-FindWebsiteItems"
	var (
		params []any
		conds  []string
		sb     strings.Builder
	)
	if len(request.WebsiteItemIDs) > 0 {
		sb.Reset()
		_, _ = sb.WriteString("id IN (")
		for i, websiteItemID := range request.WebsiteItemIDs {
			params = append(params, websiteItemID)
			if i > 0 {
				_, _ = sb.WriteString(",")
			}
			_, _ = sb.WriteString("$")
			_, _ = sb.WriteString(strconv.Itoa(len(params)))
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
			_, _ = sb.WriteString("$")
			_, _ = sb.WriteString(strconv.Itoa(len(params)))
		}
		_, _ = sb.WriteString(")")
		conds = append(conds, sb.String())
	}
	if len(request.WebsiteItemTypes) > 0 {
		sb.Reset()
		_, _ = sb.WriteString("website_item_type IN (")
		for i, websiteItemType := range request.WebsiteItemTypes {
			params = append(params, websiteItemType)
			if i > 0 {
				_, _ = sb.WriteString(",")
			}
			_, _ = sb.WriteString("$")
			_, _ = sb.WriteString(strconv.Itoa(len(params)))
		}
		_, _ = sb.WriteString(")")
		conds = append(conds, sb.String())
	}
	if request.Keyword != "" {
		sb.Reset()
		_, _ = sb.WriteString("%%")
		_, _ = sb.WriteString(strings.ToLower(request.Keyword))
		_, _ = sb.WriteString("%%")
		params = append(params, sb.String())
		n := strconv.Itoa(len(params))
		sb.Reset()
		_, _ = sb.WriteString("(LOWER(subject) LIKE $")
		_, _ = sb.WriteString(n)
		_, _ = sb.WriteString(" OR LOWER(body) LIKE $")
		_, _ = sb.WriteString(n)
		_, _ = sb.WriteString(")")
		conds = append(conds, sb.String())
	}
	sb.Reset()
	_, _ = sb.WriteString(
		`SELECT COUNT(1)
        	FROM website_items`,
	)
	if len(conds) > 0 {
		_, _ = sb.WriteString(" WHERE")
		for i, cond := range conds {
			if i > 0 {
				_, _ = sb.WriteString(" AND")
			}
			_, _ = sb.WriteString(" ")
			_, _ = sb.WriteString(cond)
		}
	}
	query := sb.String()
	var total int64
	err := q.dbRead.QueryRow(ctx, query, params...).Scan(&total)
	if err != nil {
		helper.Capture(ctx, zap.ErrorLevel, err, ctxt, "ErrExec")
		return nil, 0, 0, err
	}
	if total == 0 {
		return nil, 0, 0, nil
	}
	query = strings.ReplaceAll(
		query,
		"COUNT(1)",
		`ROW_NUMBER() OVER (ORDER BY id DESC) AS row_no
		, id
		, record_status
		, website_item_type
		, subject
		, body
		, photo
		, link_text
		, link_url
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
	if request.Limit > 0 {
		totalDecimal, err := decimal.New(total, 0)
		if err != nil {
			helper.Capture(ctx, zap.ErrorLevel, err, ctxt, "ErrNew")
			return nil, 0, 0, err
		}
		perPageDecimal, err := decimal.New(request.Limit, 0)
		if err != nil {
			helper.Capture(ctx, zap.ErrorLevel, err, ctxt, "ErrNew")
			return nil, 0, 0, err
		}
		pagesDecimal, err := totalDecimal.Quo(perPageDecimal)
		if err != nil {
			helper.Capture(ctx, zap.ErrorLevel, err, ctxt, "ErrQuo")
			return nil, 0, 0, err
		}
		pages, _, _ = pagesDecimal.Ceil(0).Int64(0)
		offset := request.Page * request.Limit
		_, _ = sb.WriteString(" LIMIT ")
		_, _ = sb.WriteString(strconv.FormatInt(request.Limit, 10))
		_, _ = sb.WriteString(" OFFSET ")
		_, _ = sb.WriteString(strconv.FormatInt(offset, 10))
	}
	rows, err := q.dbRead.Query(ctx, sb.String(), params...)
	if errors.Is(err, pgx.ErrNoRows) {
		err = nil
	}
	if err != nil {
		helper.Capture(ctx, zap.ErrorLevel, err, ctxt, "ErrQuery")
		return nil, 0, 0, err
	}
	defer rows.Close()
	var response []*models.WebsiteItem
	for rows.Next() {
		var websiteItem models.WebsiteItem
		if err = rows.Scan(
			&websiteItem.RowNo,
			&websiteItem.ID,
			&websiteItem.RecordStatus,
			&websiteItem.WebsiteItemType,
			&websiteItem.Subject,
			&websiteItem.Body,
			&websiteItem.Photo,
			&websiteItem.LinkText,
			&websiteItem.LinkURL,
			&websiteItem.CreatedBy,
			&websiteItem.CreatedAt,
			&websiteItem.UpdatedBy,
			&websiteItem.UpdatedAt,
			&websiteItem.DeletedBy,
			&websiteItem.DeletedAt,
			&websiteItem.DeletionReason,
		); err != nil {
			helper.Capture(ctx, zap.ErrorLevel, err, ctxt, "ErrScan")
			return nil, 0, 0, err
		}
		response = append(response, &websiteItem)
	}
	return response, total, pages, nil
}

func (q *websiteItemRepository) CreateWebsiteItem(
	ctx context.Context,
	tx pgx.Tx,
	request *models.WebsiteItem,
) (*models.WebsiteItem, error) {
	ctxt := "WebsiteItemRepository-CreateWebsiteItem"
	var response models.WebsiteItem
	if err := tx.QueryRow(
		ctx,
		`INSERT INTO website_items (
			record_status
			, website_item_type
			, subject
			, body
			, photo
			, link_text
			, link_url
			, created_by
			, created_at
			, updated_by
			, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		RETURNING id
			, record_status
			, website_item_type
			, subject
			, body
			, photo
			, link_text
			, link_url
			, created_by
			, created_at
			, updated_by
			, updated_at
			, deleted_by
			, deleted_at
			, deletion_reason`,
		request.RecordStatus,
		request.WebsiteItemType,
		request.Subject,
		request.Body,
		request.Photo,
		request.LinkText,
		request.LinkURL,
		request.CreatedBy,
		request.CreatedAt,
		request.UpdatedBy,
		request.UpdatedAt,
	).Scan(
		&response.ID,
		&response.RecordStatus,
		&response.WebsiteItemType,
		&response.Subject,
		&response.Body,
		&response.Photo,
		&response.LinkText,
		&response.LinkURL,
		&response.CreatedBy,
		&response.CreatedAt,
		&response.UpdatedBy,
		&response.UpdatedAt,
		&response.DeletedBy,
		&response.DeletedAt,
		&response.DeletionReason,
	); err != nil {
		if errRollback := tx.Rollback(ctx); errRollback != nil {
			helper.Capture(ctx, zap.ErrorLevel, errRollback, ctxt, "ErrRollback")
		}
		helper.Capture(ctx, zap.ErrorLevel, err, ctxt, "ErrScan")
		return nil, err
	}
	return new(response), nil
}

func (q *websiteItemRepository) UpdateWebsiteItem(
	ctx context.Context,
	tx pgx.Tx,
	request *models.WebsiteItem,
) (*models.WebsiteItem, error) {
	ctxt := "WebsiteItemRepository-UpdateWebsiteItem"
	var response models.WebsiteItem
	if err := tx.QueryRow(
		ctx,
		`UPDATE website_items SET
			record_status = $1
			, subject = $2
			, body = $3
			, photo = $4
			, link_text = $5
			, link_url = $6
			, updated_by = $7
			, updated_at = $8
			, deleted_by = $9
			, deleted_at = $10
			, deletion_reason = $11
		WHERE id = $12
		RETURNING id
			, record_status
			, website_item_type
			, subject
			, body
			, photo
			, link_text
			, link_url
			, created_by
			, created_at
			, updated_by
			, updated_at
			, deleted_by
			, deleted_at
			, deletion_reason`,
		request.RecordStatus,
		request.Subject,
		request.Body,
		request.Photo,
		request.LinkText,
		request.LinkURL,
		request.UpdatedBy,
		request.UpdatedAt,
		request.DeletedBy,
		request.DeletedAt,
		request.DeletionReason,
		request.ID,
	).Scan(
		&response.ID,
		&response.RecordStatus,
		&response.WebsiteItemType,
		&response.Subject,
		&response.Body,
		&response.Photo,
		&response.LinkText,
		&response.LinkURL,
		&response.CreatedBy,
		&response.CreatedAt,
		&response.UpdatedBy,
		&response.UpdatedAt,
		&response.DeletedBy,
		&response.DeletedAt,
		&response.DeletionReason,
	); err != nil {
		if errRollback := tx.Rollback(ctx); errRollback != nil {
			helper.Capture(ctx, zap.ErrorLevel, errRollback, ctxt, "ErrRollback")
		}
		helper.Capture(ctx, zap.ErrorLevel, err, ctxt, "ErrScan")
		return nil, err
	}
	return new(response), nil
}
