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
	WebsiteBlockRepository interface {
		FindWebsiteBlocks(
			ctx context.Context,
			request *models.WebsiteBlockFilter,
		) ([]*models.WebsiteBlock, int64, int64, error)
		CreateWebsiteBlock(
			ctx context.Context,
			tx pgx.Tx,
			request *models.WebsiteBlock,
		) (*models.WebsiteBlock, error)
		UpdateWebsiteBlock(
			ctx context.Context,
			tx pgx.Tx,
			request *models.WebsiteBlock,
		) (*models.WebsiteBlock, error)
	}

	websiteBlockRepository struct {
		dbRead,
		dbWrite *pgxpool.Pool
	}
)

func NewWebsiteBlockRepository(
	dbRead,
	dbWrite *pgxpool.Pool,
) WebsiteBlockRepository {
	return new(websiteBlockRepository{
		dbRead:  dbRead,
		dbWrite: dbWrite,
	})
}

func (q *websiteBlockRepository) FindWebsiteBlocks(
	ctx context.Context,
	request *models.WebsiteBlockFilter,
) ([]*models.WebsiteBlock, int64, int64, error) {
	ctxt := "WebsiteBlockRepository-FindWebsiteBlocks"
	var (
		params []any
		conds  []string
		sb     strings.Builder
	)
	if len(request.WebsiteBlockIDs) > 0 {
		sb.Reset()
		_, _ = sb.WriteString("id IN (")
		for i, websiteBlockID := range request.WebsiteBlockIDs {
			params = append(params, websiteBlockID)
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
	if len(request.Names) > 0 {
		sb.Reset()
		_, _ = sb.WriteString("name IN (")
		for i, name := range request.Names {
			params = append(params, name)
			if i > 0 {
				_, _ = sb.WriteString(",")
			}
			_, _ = sb.WriteString("$")
			_, _ = sb.WriteString(strconv.Itoa(len(params)))
		}
		_, _ = sb.WriteString(")")
		conds = append(conds, sb.String())
	}
	sb.Reset()
	_, _ = sb.WriteString(
		`SELECT COUNT(1)
        	FROM website_blocks`,
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
		, name
		, attributes
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
	var response []*models.WebsiteBlock
	for rows.Next() {
		var websiteBlock models.WebsiteBlock
		if err = rows.Scan(
			&websiteBlock.RowNo,
			&websiteBlock.ID,
			&websiteBlock.RecordStatus,
			&websiteBlock.Name,
			&websiteBlock.Attributes,
			&websiteBlock.CreatedBy,
			&websiteBlock.CreatedAt,
			&websiteBlock.UpdatedBy,
			&websiteBlock.UpdatedAt,
			&websiteBlock.DeletedBy,
			&websiteBlock.DeletedAt,
			&websiteBlock.DeletionReason,
		); err != nil {
			helper.Capture(ctx, zap.ErrorLevel, err, ctxt, "ErrScan")
			return nil, 0, 0, err
		}
		response = append(response, &websiteBlock)
	}
	return response, total, pages, nil
}

func (q *websiteBlockRepository) CreateWebsiteBlock(
	ctx context.Context,
	tx pgx.Tx,
	request *models.WebsiteBlock,
) (*models.WebsiteBlock, error) {
	ctxt := "WebsiteBlockRepository-CreateWebsiteBlock"
	var response models.WebsiteBlock
	if err := tx.QueryRow(
		ctx,
		`INSERT INTO website_blocks (
			record_status
			, name
			, attributes
			, created_by
			, created_at
			, updated_by
			, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id
			, record_status
			, name
			, attributes
			, created_by
			, created_at
			, updated_by
			, updated_at
			, deleted_by
			, deleted_at
			, deletion_reason`,
		request.RecordStatus,
		request.Name,
		request.Attributes,
		request.CreatedBy,
		request.CreatedAt,
		request.UpdatedBy,
		request.UpdatedAt,
	).Scan(
		&response.ID,
		&response.RecordStatus,
		&response.Name,
		&response.Attributes,
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

func (q *websiteBlockRepository) UpdateWebsiteBlock(
	ctx context.Context,
	tx pgx.Tx,
	request *models.WebsiteBlock,
) (*models.WebsiteBlock, error) {
	ctxt := "WebsiteBlockRepository-UpdateWebsiteBlock"
	var response models.WebsiteBlock
	if err := tx.QueryRow(
		ctx,
		`UPDATE website_blocks SET
			record_status = $1
			, attributes = $2
			, updated_by = $3
			, updated_at = $4
			, deleted_by = $5
			, deleted_at = $6
			, deletion_reason = $7
		WHERE id = $8
		RETURNING id
			, record_status
			, name
			, attributes
			, created_by
			, created_at
			, updated_by
			, updated_at
			, deleted_by
			, deleted_at
			, deletion_reason`,
		request.RecordStatus,
		request.Name,
		request.Attributes,
		request.UpdatedBy,
		request.UpdatedAt,
		request.DeletedBy,
		request.DeletedAt,
		request.DeletionReason,
		request.ID,
	).Scan(
		&response.ID,
		&response.RecordStatus,
		&response.Name,
		&response.Attributes,
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
