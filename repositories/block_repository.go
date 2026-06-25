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
	BlockRepository interface {
		FindBlocks(
			ctx context.Context,
			request *models.BlockFilter,
		) ([]*models.Block, int64, int64, error)
		UpdateBlock(
			ctx context.Context,
			tx pgx.Tx,
			request *models.Block,
		) (*models.Block, error)
		FindBlockItems(
			ctx context.Context,
			request *models.BlockItemFilter,
		) ([]*models.BlockItem, int64, int64, error)
		CreateBlockItem(
			ctx context.Context,
			tx pgx.Tx,
			request *models.BlockItem,
		) (*models.BlockItem, error)
		UpdateBlockItem(
			ctx context.Context,
			tx pgx.Tx,
			request *models.BlockItem,
		) (*models.BlockItem, error)
	}

	blockRepository struct {
		dbRead,
		dbWrite *pgxpool.Pool
	}
)

func NewBlockRepository(
	dbRead,
	dbWrite *pgxpool.Pool,
) BlockRepository {
	return new(blockRepository{
		dbRead:  dbRead,
		dbWrite: dbWrite,
	})
}

func (q *blockRepository) FindBlocks(
	ctx context.Context,
	request *models.BlockFilter,
) ([]*models.Block, int64, int64, error) {
	ctxt := "BlockRepository-FindBlocks"
	var (
		params []any
		conds  []string
		sb     strings.Builder
	)
	if len(request.BlockIDs) > 0 {
		sb.Reset()
		_, _ = sb.WriteString("id IN (")
		for i, blockID := range request.BlockIDs {
			params = append(params, blockID)
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
	if request.Keyword != "" {
		sb.Reset()
		_, _ = sb.WriteString("%%")
		_, _ = sb.WriteString(strings.ToLower(request.Keyword))
		_, _ = sb.WriteString("%%")
		params = append(params, sb.String())
		n := strconv.Itoa(len(params))
		sb.Reset()
		_, _ = sb.WriteString("name LIKE $")
		_, _ = sb.WriteString(n)
		conds = append(conds, sb.String())
	}
	sb.Reset()
	_, _ = sb.WriteString(
		`SELECT COUNT(1)
        	FROM blocks`,
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
		, name
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
	var response []*models.Block
	for rows.Next() {
		var block models.Block
		if err = rows.Scan(
			&block.RowNo,
			&block.ID,
			&block.RecordStatus,
			&block.Name,
			&block.CustomAttributes,
			&block.CreatedBy,
			&block.CreatedAt,
			&block.UpdatedBy,
			&block.UpdatedAt,
			&block.DeletedBy,
			&block.DeletedAt,
			&block.DeletionReason,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrScan")
			return nil, 0, 0, err
		}
		response = append(response, &block)
	}
	return response, total, pages, nil
}

func (q *blockRepository) UpdateBlock(
	ctx context.Context,
	tx pgx.Tx,
	request *models.Block,
) (*models.Block, error) {
	ctxt := "BlockRepository-UpdateBlock"
	var response models.Block
	if err := tx.QueryRow(
		ctx,
		`UPDATE blocks SET
			record_status = $1
			, custom_attributes = $2
			, updated_by = $3
			, updated_at = $4
			, deleted_by = $5
			, deleted_at = $6
			, deletion_reason = $7
		WHERE id = $8
		RETURNING id
			, record_status
			, name
			, custom_attributes
			, created_by
			, created_at
			, updated_by
			, updated_at
			, deleted_by
			, deleted_at
			, deletion_reason`,
		request.RecordStatus,
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
		&response.Name,
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

func (q *blockRepository) FindBlockItems(
	ctx context.Context,
	request *models.BlockItemFilter,
) ([]*models.BlockItem, int64, int64, error) {
	ctxt := "BlockRepository-FindBlockItems"
	var (
		params []any
		conds  []string
		sb     strings.Builder
	)
	if len(request.BlockItemIDs) > 0 {
		sb.Reset()
		_, _ = sb.WriteString("id IN (")
		for i, blockItemID := range request.BlockItemIDs {
			params = append(params, blockItemID)
			if i > 0 {
				_, _ = sb.WriteString(",")
			}
			_, _ = sb.WriteString("$")
			_, _ = sb.WriteString(strconv.Itoa(len(params)))
		}
		_, _ = sb.WriteString(")")
		conds = append(conds, sb.String())
	}
	if len(request.BlockNames) > 0 {
		sb.Reset()
		_, _ = sb.WriteString("block_name IN (")
		for i, blockName := range request.BlockNames {
			params = append(params, blockName)
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
		FROM block_items`,
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
	var response []*models.BlockItem
	for rows.Next() {
		var blockItem models.BlockItem
		if err = rows.Scan(
			&blockItem.RowNo,
			&blockItem.ID,
			&blockItem.RecordStatus,
			&blockItem.BlockName,
			&blockItem.Subject,
			&blockItem.Body,
			&blockItem.LinkText,
			&blockItem.LinkURL,
			&blockItem.Photo,
			&blockItem.CreatedBy,
			&blockItem.CreatedAt,
			&blockItem.UpdatedBy,
			&blockItem.UpdatedAt,
			&blockItem.DeletedBy,
			&blockItem.DeletedAt,
			&blockItem.DeletionReason,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrScan")
			return nil, 0, 0, err
		}
		response = append(response, &blockItem)
	}
	return response, total, pages, nil
}

func (q *blockRepository) CreateBlockItem(
	ctx context.Context,
	tx pgx.Tx,
	request *models.BlockItem,
) (*models.BlockItem, error) {
	ctxt := "BlockRepository-CreateBlockItem"
	var response models.BlockItem
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
		request.BlockName,
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
		&response.BlockName,
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
		if errRollback := tx.Rollback(ctx); errRollback != nil {
			helper.Capture(ctx, slog.LevelError, errRollback, ctxt, "ErrRollback")
		}
		helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrScan")
		return nil, err
	}
	return new(response), nil
}

func (q *blockRepository) UpdateBlockItem(
	ctx context.Context,
	tx pgx.Tx,
	request *models.BlockItem,
) (*models.BlockItem, error) {
	ctxt := "BlockRepository-UpdateBlockItem"
	var response models.BlockItem
	if err := tx.QueryRow(
		ctx,
		`UPDATE block_items SET
			record_status = $1
			, subject = $2
			, body = $3
			, link_text = $4
			, link_url = $5
			, photo = $6
			, updated_by = $7
			, updated_at = $8
			, deleted_by = $9
			, deleted_at = $10
			, deletion_reason = $11
		WHERE id = $12
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
		request.Subject,
		request.Body,
		request.LinkText,
		request.LinkURL,
		request.Photo,
		request.UpdatedBy,
		request.UpdatedAt,
		request.DeletedBy,
		request.DeletedAt,
		request.DeletionReason,
		request.ID,
	).Scan(
		&response.ID,
		&response.RecordStatus,
		&response.BlockName,
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
		if errRollback := tx.Rollback(ctx); errRollback != nil {
			helper.Capture(ctx, slog.LevelError, errRollback, ctxt, "ErrRollback")
		}
		helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrScan")
		return nil, err
	}
	return new(response), nil
}
