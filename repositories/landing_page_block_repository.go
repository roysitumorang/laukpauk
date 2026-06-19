package repositories

import (
	"context"
	"errors"
	"log/slog"
	"strconv"
	"strings"

	"github.com/govalues/decimal"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/roysitumorang/laukpauk/helper"
	"github.com/roysitumorang/laukpauk/models"
)

type (
	LandingPageBlockRepository interface {
		FindLandingPageBlocks(
			ctx context.Context,
			request *models.LandingPageBlockFilter,
		) ([]*models.LandingPageBlock, int64, int64, error)
		CreateLandingPageBlock(
			ctx context.Context,
			tx pgx.Tx,
			request *models.LandingPageBlock,
		) (*models.LandingPageBlock, error)
		UpdateLandingPageBlock(
			ctx context.Context,
			tx pgx.Tx,
			request *models.LandingPageBlock,
		) (*models.LandingPageBlock, error)
	}

	landingPageBlockRepository struct {
		dbRead,
		dbWrite *pgxpool.Pool
	}
)

func NewLandingPageBlockRepository(
	dbRead,
	dbWrite *pgxpool.Pool,
) LandingPageBlockRepository {
	return new(landingPageBlockRepository{
		dbRead:  dbRead,
		dbWrite: dbWrite,
	})
}

func (q *landingPageBlockRepository) FindLandingPageBlocks(
	ctx context.Context,
	request *models.LandingPageBlockFilter,
) ([]*models.LandingPageBlock, int64, int64, error) {
	ctxt := "LandingPageBlockRepository-FindLandingPageBlocks"
	var (
		params []any
		conds  []string
		sb     strings.Builder
	)
	if len(request.LandingPageBlockIDs) > 0 {
		sb.Reset()
		_, _ = sb.WriteString("id IN (")
		for i, landingPageBlockID := range request.LandingPageBlockIDs {
			params = append(params, landingPageBlockID)
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
	if len(request.ParentNames) > 0 {
		sb.Reset()
		_, _ = sb.WriteString("parent_name IN (")
		for i, parentName := range request.ParentNames {
			params = append(params, parentName)
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
        	FROM landing_page_blocks`,
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
		helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
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
		, parent_name
		, name
		, subject
		, body
		, link_text
		, link_url
		, photo
		, custom_attributes
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
		helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrQuery")
		return nil, 0, 0, err
	}
	defer rows.Close()
	var response []*models.LandingPageBlock
	for rows.Next() {
		var landingPageBlock models.LandingPageBlock
		if err = rows.Scan(
			&landingPageBlock.RowNo,
			&landingPageBlock.ID,
			&landingPageBlock.RecordStatus,
			&landingPageBlock.ParentName,
			&landingPageBlock.Name,
			&landingPageBlock.Subject,
			&landingPageBlock.Body,
			&landingPageBlock.LinkText,
			&landingPageBlock.LinkURL,
			&landingPageBlock.Photo,
			&landingPageBlock.CustomAttributes,
			&landingPageBlock.CreatedBy,
			&landingPageBlock.CreatedAt,
			&landingPageBlock.UpdatedBy,
			&landingPageBlock.UpdatedAt,
			&landingPageBlock.DeletedBy,
			&landingPageBlock.DeletedAt,
			&landingPageBlock.DeletionReason,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrScan")
			return nil, 0, 0, err
		}
		response = append(response, &landingPageBlock)
	}
	return response, total, pages, nil
}

func (q *landingPageBlockRepository) CreateLandingPageBlock(
	ctx context.Context,
	tx pgx.Tx,
	request *models.LandingPageBlock,
) (*models.LandingPageBlock, error) {
	ctxt := "LandingPageBlockRepository-CreateLandingPageBlock"
	var response models.LandingPageBlock
	if err := tx.QueryRow(
		ctx,
		`INSERT INTO landing_page_blocks (
			record_status
			, parent_name
			, name
			, subject
			, body
			, link_text
			, link_url
			, photo
			, custom_attributes
			, created_by
			, created_at
			, updated_by
			, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
		RETURNING id
			, record_status
			, parent_name
			, name
			, subject
			, body
			, link_text
			, link_url
			, photo
			, custom_attributes
			, created_by
			, created_at
			, updated_by
			, updated_at
			, deleted_by
			, deleted_at
			, deletion_reason`,
		request.RecordStatus,
		request.ParentName,
		request.Name,
		request.Subject,
		request.Body,
		request.LinkText,
		request.LinkURL,
		request.Photo,
		request.CustomAttributes,
		request.CreatedBy,
		request.CreatedAt,
		request.UpdatedBy,
		request.UpdatedAt,
	).Scan(
		&response.ID,
		&response.RecordStatus,
		&response.ParentName,
		&response.Name,
		&response.Subject,
		&response.Body,
		&response.LinkText,
		&response.LinkURL,
		&response.Photo,
		&response.CustomAttributes,
		&response.CreatedBy,
		&response.CreatedAt,
		&response.UpdatedBy,
		&response.UpdatedAt,
		&response.DeletedBy,
		&response.DeletedAt,
		&response.DeletionReason,
	); err != nil {
		if errRollback := tx.Rollback(ctx); errRollback != nil {
			helper.Capture(ctx, slog.LevelError, errRollback, ctxt, "ErrRollback")
		}
		helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrScan")
		return nil, err
	}
	return new(response), nil
}

func (q *landingPageBlockRepository) UpdateLandingPageBlock(
	ctx context.Context,
	tx pgx.Tx,
	request *models.LandingPageBlock,
) (*models.LandingPageBlock, error) {
	ctxt := "LandingPageBlockRepository-UpdateLandingPageBlock"
	var response models.LandingPageBlock
	if err := tx.QueryRow(
		ctx,
		`UPDATE landing_page_blocks SET
			record_status = $1
			, subject = $2
			, body = $3
			, link_text = $4
			, link_url = $5
			, photo = $6
			, custom_attributes = $7
			, updated_by = $8
			, updated_at = $9
			, deleted_by = $10
			, deleted_at = $11
			, deletion_reason = $12
		WHERE id = $13
		RETURNING id
			, record_status
			, parent_name
			, name
			, subject
			, body
			, link_text
			, link_url
			, photo
			, custom_attributes
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
		request.LinkText,
		request.LinkURL,
		request.Photo,
		request.CustomAttributes,
		request.UpdatedBy,
		request.UpdatedAt,
		request.DeletedBy,
		request.DeletedAt,
		request.DeletionReason,
		request.ID,
	).Scan(
		&response.ID,
		&response.RecordStatus,
		&response.ParentName,
		&response.Name,
		&response.Subject,
		&response.Body,
		&response.LinkText,
		&response.LinkURL,
		&response.Photo,
		&response.CustomAttributes,
		&response.CreatedBy,
		&response.CreatedAt,
		&response.UpdatedBy,
		&response.UpdatedAt,
		&response.DeletedBy,
		&response.DeletedAt,
		&response.DeletionReason,
	); err != nil {
		if errRollback := tx.Rollback(ctx); errRollback != nil {
			helper.Capture(ctx, slog.LevelError, errRollback, ctxt, "ErrRollback")
		}
		helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrScan")
		return nil, err
	}
	return new(response), nil
}
