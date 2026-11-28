package repositories

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"strings"

	"github.com/govalues/decimal"
	"github.com/roysitumorang/laukpauk/helper"
	"github.com/roysitumorang/laukpauk/models"
	"github.com/roysitumorang/laukpauk/pools"
)

type PointRepository interface {
	FindPoints(
		ctx context.Context,
		request *models.PointFilter,
	) ([]*models.Point, int64, int64, error)
	FindPoint(
		ctx context.Context,
		request *models.PointFilter,
	) (*models.Point, error)
	CreatePoint(
		ctx context.Context,
		tx *sql.Tx,
		request *models.Point,
	) (*models.Point, error)
	UpdatePoint(
		ctx context.Context,
		tx *sql.Tx,
		request *models.Point,
	) (*models.Point, error)
}

type pointRepository struct {
	dbRead,
	dbWrite *sql.DB
}

func NewPointRepository(
	dbRead,
	dbWrite *sql.DB,
) PointRepository {
	return &pointRepository{
		dbRead:  dbRead,
		dbWrite: dbWrite,
	}
}

func (q *pointRepository) FindPoints(
	ctx context.Context,
	request *models.PointFilter,
) ([]*models.Point, int64, int64, error) {
	ctxt := "PointRepository-FindPoints"
	sb := pools.BuilderPool.Get()
	defer func() {
		sb.Reset()
		pools.BuilderPool.Put(sb)
	}()
	sb.Reset()
	condition, params := q.buildCondition(request, sb)
	sb.Reset()
	_, _ = fmt.Fprintf(sb, "SELECT COUNT(1) FROM points %s", condition)
	query := sb.String()
	var total int64
	err := q.dbRead.QueryRowContext(ctx, query, helper.SQLArgs(params...)...).Scan(helper.SQLScan(&total)...)
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
		, point_type
		, user_id
		, amount
		, transaction_id
		, description
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
	sortBy := request.Sortings
	if len(sortBy) == 0 {
		sortBy = append(sortBy, models.DefaultSortBy)
	}
	_, _ = sb.WriteString(" ORDER BY")
	for i, item := range sortBy {
		if i > 0 {
			_, _ = sb.WriteString(",")
		}
		_, _ = fmt.Fprintf(sb, " %s %s", sortExpression(item.Field), item.Direction)
	}
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
	rows, err := q.dbRead.QueryContext(ctx, sb.String(), helper.SQLArgs(params...)...)
	if errors.Is(err, sql.ErrNoRows) {
		err = nil
	}
	if err != nil {
		helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrQuery")
		return nil, 0, 0, err
	}
	defer rows.Close()
	var response []*models.Point
	for rows.Next() {
		var point models.Point
		if err = rows.Scan(
			helper.SQLScan(
				&point.ID,
				&point.RecordStatus,
				&point.PointType,
				&point.UserID,
				&point.Amount,
				&point.TransactionID,
				&point.Description,
				&point.CreatedBy,
				&point.CreatedAt,
				&point.UpdatedBy,
				&point.UpdatedAt,
				&point.DeletedBy,
				&point.DeletedAt,
				&point.DeletionReason,
			)...,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrScan")
			return nil, 0, 0, err
		}
		offset++
		point.RowNo = offset
		response = append(response, &point)
	}
	return response, total, pages, nil
}

func (q *pointRepository) FindPoint(
	ctx context.Context,
	request *models.PointFilter,
) (*models.Point, error) {
	ctxt := "PointRepository-FindPoint"
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
			, point_type
			, user_id
			, amount
			, transaction_id
			, description
			, created_by
			, created_at
			, updated_by
			, updated_at
			, deleted_by
			, deleted_at
			, deletion_reason
		FROM points %s`,
		condition,
	)
	query := sb.String()
	var response models.Point
	err := q.dbRead.QueryRowContext(ctx, query, helper.SQLArgs(params...)...).Scan(
		helper.SQLScan(
			&response.ID,
			&response.RecordStatus,
			&response.PointType,
			&response.UserID,
			&response.Amount,
			&response.TransactionID,
			&response.Description,
			&response.CreatedBy,
			&response.CreatedAt,
			&response.UpdatedBy,
			&response.UpdatedAt,
			&response.DeletedBy,
			&response.DeletedAt,
			&response.DeletionReason,
		)...,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, models.ErrPointNotFound
	}
	if err != nil {
		helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrScan")
		return nil, err
	}
	return &response, nil
}

func (q *pointRepository) CreatePoint(
	ctx context.Context,
	tx *sql.Tx,
	request *models.Point,
) (*models.Point, error) {
	ctxt := "PointRepository-CreatePoint"
	var response models.Point
	err := tx.QueryRowContext(
		ctx,
		`INSERT INTO points (
			record_status
			, point_type
			, user_id
			, amount
			, transaction_id
			, description
			, created_by
			, created_at
			, updated_by
			, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		RETURNING id
			, record_status
			, point_type
			, user_id
			, amount
			, transaction_id
			, description
			, created_by
			, created_at
			, updated_by
			, updated_at
			, deleted_by
			, deleted_at
			, deletion_reason`,
		helper.SQLArgs(
			request.RecordStatus,
			request.PointType,
			request.UserID,
			request.Amount,
			request.TransactionID,
			request.Description,
			request.CreatedBy,
			request.CreatedAt,
			request.UpdatedBy,
			request.UpdatedAt,
		)...,
	).Scan(
		helper.SQLScan(
			&response.ID,
			&response.RecordStatus,
			&response.PointType,
			&response.UserID,
			&response.Amount,
			&response.TransactionID,
			&response.Description,
			&response.CreatedBy,
			&response.CreatedAt,
			&response.UpdatedBy,
			&response.UpdatedAt,
			&response.DeletedBy,
			&response.DeletedAt,
			&response.DeletionReason,
		)...,
	)
	if err != nil {
		helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrScan")
		return nil, err
	}
	return &response, nil
}

func (q *pointRepository) UpdatePoint(
	ctx context.Context,
	tx *sql.Tx,
	request *models.Point,
) (*models.Point, error) {
	ctxt := "PointRepository-UpdatePoint"
	var response models.Point
	err := tx.QueryRowContext(
		ctx,
		`SELECT
			id
			, record_status
			, point_type
			, user_id
			, amount
			, transaction_id
			, description
			, created_by
			, created_at
			, updated_by
			, updated_at
			, deleted_by
			, deleted_at
			, deletion_reason
		FROM points
		WHERE id = ?
		FOR UPDATE`,
		helper.SQLArgs(request.ID)...,
	).Scan(
		helper.SQLScan(
			&response.ID,
			&response.RecordStatus,
			&response.PointType,
			&response.UserID,
			&response.Amount,
			&response.TransactionID,
			&response.Description,
			&response.CreatedBy,
			&response.CreatedAt,
			&response.UpdatedBy,
			&response.UpdatedAt,
			&response.DeletedBy,
			&response.DeletedAt,
			&response.DeletionReason,
		)...,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, models.ErrPointNotFound
	}
	if err != nil {
		helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrScan")
		return nil, err
	}
	var params []any
	var n int
	sb := pools.BuilderPool.Get()
	defer func() {
		sb.Reset()
		pools.BuilderPool.Put(sb)
	}()
	sb.Reset()
	_, _ = sb.WriteString("UPDATE points SET")
	if response.RecordStatus != request.RecordStatus {
		params = append(params, request.RecordStatus)
		if n = len(params); n > 1 {
			_, _ = sb.WriteString(",")
		}
		_, _ = sb.WriteString(" record_status = ?")
	}
	if response.Amount != request.Amount {
		params = append(params, request.Amount)
		if n = len(params); n > 1 {
			_, _ = sb.WriteString(",")
		}
		_, _ = sb.WriteString(" amount = ?")
	}
	if response.Description != request.Description {
		params = append(params, request.Description)
		if n = len(params); n > 1 {
			_, _ = sb.WriteString(",")
		}
		_, _ = sb.WriteString(" description = ?")
	}
	if len(params) == 0 {
		return &response, nil
	}
	if (response.UpdatedBy == nil && request.UpdatedBy != nil) ||
		(response.UpdatedBy != nil && request.UpdatedBy == nil) ||
		(response.UpdatedBy != nil && request.UpdatedBy != nil &&
			response.UpdatedBy.Compare(*request.UpdatedBy) != 0) {
		params = append(params, request.DeletedBy)
		if n = len(params); n > 1 {
			_, _ = sb.WriteString(",")
		}
		_, _ = sb.WriteString(" updated_by = ?")
	}
	if !response.UpdatedAt.Equal(request.UpdatedAt) {
		params = append(params, request.UpdatedAt)
		if n = len(params); n > 1 {
			_, _ = sb.WriteString(",")
		}
		_, _ = sb.WriteString(" updated_at = ?")
	}
	if (response.DeletedBy == nil && request.DeletedBy != nil) ||
		(response.DeletedBy != nil && request.DeletedBy == nil) ||
		(response.DeletedBy != nil && request.DeletedBy != nil &&
			response.DeletedBy.Compare(*request.DeletedBy) != 0) {
		params = append(params, request.DeletedBy)
		if n = len(params); n > 1 {
			_, _ = sb.WriteString(",")
		}
		_, _ = sb.WriteString(" deleted_by = ?")
	}
	if (response.DeletedAt == nil && request.DeletedAt != nil) ||
		(response.DeletedAt != nil && request.DeletedAt == nil) ||
		(response.DeletedAt != nil && request.DeletedAt != nil &&
			!response.DeletedAt.Equal(*request.DeletedAt)) {
		params = append(params, request.DeletedAt)
		if n = len(params); n > 1 {
			_, _ = sb.WriteString(",")
		}
		_, _ = sb.WriteString(" deleted_at = ?")
	}
	if (response.DeletionReason == nil && request.DeletionReason != nil) ||
		(response.DeletionReason != nil && request.DeletionReason == nil) ||
		(response.DeletionReason != nil && request.DeletionReason != nil &&
			*response.DeletionReason != *request.DeletionReason) {
		params = append(params, request.DeletionReason)
		if n = len(params); n > 1 {
			_, _ = sb.WriteString(",")
		}
		_, _ = sb.WriteString(" deletion_reason = ?")
	}
	params = append(params, request.ID)
	_, _ = sb.WriteString(` WHERE id = ?
		RETURNING id
			, record_status
			, point_type
			, user_id
			, amount
			, transaction_id
			, description
			, created_by
			, created_at
			, updated_by
			, updated_at
			, deleted_by
			, deleted_at
			, deletion_reason`)
	if err = tx.QueryRowContext(ctx, sb.String(), helper.SQLArgs(params...)...).Scan(
		helper.SQLScan(
			&response.ID,
			&response.RecordStatus,
			&response.PointType,
			&response.UserID,
			&response.Amount,
			&response.TransactionID,
			&response.Description,
			&response.CreatedBy,
			&response.CreatedAt,
			&response.UpdatedBy,
			&response.UpdatedAt,
			&response.DeletedBy,
			&response.DeletedAt,
			&response.DeletionReason,
		)...,
	); err != nil {
		helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrScan")
		return nil, err
	}
	return &response, nil
}

func (q *pointRepository) buildCondition(request *models.PointFilter, sb *strings.Builder) (condition string, params []any) {
	var conds []string
	if len(request.PointIDs) > 0 {
		sb.Reset()
		_, _ = sb.WriteString("id IN (")
		for i, pointID := range request.PointIDs {
			params = append(params, pointID)
			if i > 0 {
				_, _ = sb.WriteString(",")
			}
			_, _ = sb.WriteString("?")
		}
		_, _ = sb.WriteString(")")
		conds = append(conds, sb.String())
	}
	if len(request.RecordStatuses) > 0 {
		sb.Reset()
		_, _ = sb.WriteString("record_status IN (")
		for i, recordStatus := range request.RecordStatuses {
			params = append(params, recordStatus)
			if i > 0 {
				_, _ = sb.WriteString(",")
			}
			_, _ = sb.WriteString("?")
		}
		_, _ = sb.WriteString(")")
		conds = append(conds, sb.String())
	}
	if len(request.UserIDs) > 0 {
		sb.Reset()
		_, _ = sb.WriteString("user_id IN (")
		for i, buyerID := range request.UserIDs {
			params = append(params, buyerID)
			if i > 0 {
				_, _ = sb.WriteString(",")
			}
			_, _ = sb.WriteString("?")
		}
		_, _ = sb.WriteString(")")
		conds = append(conds, sb.String())
	}
	if len(request.TransactionIDs) > 0 {
		sb.Reset()
		_, _ = sb.WriteString("transaction_id IN (")
		for i, transactionID := range request.TransactionIDs {
			params = append(params, transactionID)
			if i > 0 {
				_, _ = sb.WriteString(",")
			}
			_, _ = sb.WriteString("?")
		}
		_, _ = sb.WriteString(")")
		conds = append(conds, sb.String())
	}
	if len(request.PointTypes) > 0 {
		sb.Reset()
		_, _ = sb.WriteString("point_type IN (")
		for i, pointType := range request.PointTypes {
			params = append(params, pointType)
			if i > 0 {
				_, _ = sb.WriteString(",")
			}
			_, _ = sb.WriteString("?")
		}
		_, _ = sb.WriteString(")")
		conds = append(conds, sb.String())
	}
	if request.Keyword != "" {
		sb.Reset()
		_, _ = fmt.Fprintf(sb, "%%%s%%", strings.ToLower(request.Keyword))
		keyword := sb.String()
		params = append(params, keyword, keyword, keyword, keyword)
		sb.Reset()
		_, _ = sb.WriteString(`(EXISTS(
				SELECT 1
				FROM users u
				WHERE u.id = points.user_id
					AND (
						LOWER(u.name) LIKE ?
						OR u.username LIKE ?
						OR u.mobile_phone LIKE ?
					)
			) OR LOWER(points.description) LIKE ?)`)
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
