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

type PromotionRepository interface {
	FindPromotions(
		ctx context.Context,
		request *models.PromotionFilter,
	) ([]*models.Promotion, int64, int64, error)
	FindPromotion(
		ctx context.Context,
		request *models.PromotionFilter,
	) (*models.Promotion, error)
	CreatePromotion(
		ctx context.Context,
		tx *sql.Tx,
		request *models.Promotion,
	) (*models.Promotion, error)
	UpdatePromotion(
		ctx context.Context,
		tx *sql.Tx,
		request *models.Promotion,
	) (*models.Promotion, error)
}

type promotionRepository struct {
	dbRead,
	dbWrite *sql.DB
}

func NewPromotionRepository(
	dbRead,
	dbWrite *sql.DB,
) PromotionRepository {
	return &promotionRepository{
		dbRead:  dbRead,
		dbWrite: dbWrite,
	}
}

func (q *promotionRepository) FindPromotions(
	ctx context.Context,
	request *models.PromotionFilter,
) ([]*models.Promotion, int64, int64, error) {
	ctxt := "PromotionRepository-FindPromotions"
	sb := pools.BuilderPool.Get()
	defer func() {
		sb.Reset()
		pools.BuilderPool.Put(sb)
	}()
	sb.Reset()
	condition, params := q.buildCondition(request, sb)
	sb.Reset()
	_, _ = fmt.Fprintf(sb, "SELECT COUNT(1) FROM promotions %s", condition)
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
		, name
		, code
		, deleted_code
		, description
		, discount
		, quota_per_buyer
		, effective_start_date
		, effective_end_date
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
	defer func() { _ = rows.Close() }()
	var response []*models.Promotion
	for rows.Next() {
		var promotion models.Promotion
		if err = rows.Scan(
			helper.SQLScan(
				&promotion.ID,
				&promotion.RecordStatus,
				&promotion.Name,
				&promotion.Code,
				&promotion.DeletedCode,
				&promotion.Description,
				&promotion.Discount,
				&promotion.QuotaPerBuyer,
				&promotion.EffectiveStartDate,
				&promotion.EffectiveEndDate,
				&promotion.CreatedBy,
				&promotion.CreatedAt,
				&promotion.UpdatedBy,
				&promotion.UpdatedAt,
				&promotion.DeletedBy,
				&promotion.DeletedAt,
				&promotion.DeletionReason,
			)...,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrScan")
			return nil, 0, 0, err
		}
		offset++
		promotion.RowNo = offset
		response = append(response, &promotion)
	}
	if err = rows.Err(); err != nil {
		helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrErr")
		return nil, 0, 0, err
	}
	return response, total, pages, nil
}

func (q *promotionRepository) FindPromotion(
	ctx context.Context,
	request *models.PromotionFilter,
) (*models.Promotion, error) {
	ctxt := "PromotionRepository-FindPromotion"
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
			, name
			, code
			, deleted_code
			, description
			, discount
			, quota_per_buyer
			, effective_start_date
			, effective_end_date
			, created_by
			, created_at
			, updated_by
			, updated_at
			, deleted_by
			, deleted_at
			, deletion_reason
		FROM promotions %s`,
		condition,
	)
	var response models.Promotion
	err := q.dbRead.QueryRowContext(ctx, sb.String(), helper.SQLArgs(params...)...).Scan(
		helper.SQLScan(
			&response.ID,
			&response.RecordStatus,
			&response.Name,
			&response.Code,
			&response.DeletedCode,
			&response.Description,
			&response.Discount,
			&response.QuotaPerBuyer,
			&response.EffectiveStartDate,
			&response.EffectiveEndDate,
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
		return nil, models.ErrPromotionNotFound
	}
	if err != nil {
		helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrScan")
		return nil, err
	}
	return &response, nil
}

func (q *promotionRepository) CreatePromotion(
	ctx context.Context,
	tx *sql.Tx,
	request *models.Promotion,
) (*models.Promotion, error) {
	ctxt := "PromotionRepository-CreatePromotion"
	var response models.Promotion
	if err := tx.QueryRowContext(
		ctx,
		`INSERT INTO promotions (
			record_status
			, name
			, code
			, description
			, discount
			, quota_per_buyer
			, effective_start_date
			, effective_end_date
			, created_by
			, created_at
			, updated_by
			, updated_at
		) VALUES (
			?, ?, ?, ?, ?, ?, ?, ?, ?,
			?, ?, ?
		)
		RETURNING id
			, record_status
			, name
			, code
			, deleted_code
			, description
			, discount
			, quota_per_buyer
			, effective_start_date
			, effective_end_date
			, created_by
			, created_at
			, updated_by
			, updated_at
			, deleted_by
			, deleted_at
			, deletion_reason`,
		helper.SQLArgs(
			request.RecordStatus,
			request.Name,
			request.Code,
			request.Description,
			request.Discount,
			request.QuotaPerBuyer,
			request.EffectiveStartDate,
			request.EffectiveEndDate,
			request.CreatedBy,
			request.CreatedAt,
			request.UpdatedBy,
			request.UpdatedAt,
		)...,
	).Scan(
		helper.SQLScan(
			&response.ID,
			&response.RecordStatus,
			&response.Name,
			&response.Code,
			&response.DeletedCode,
			&response.Description,
			&response.Discount,
			&response.QuotaPerBuyer,
			&response.EffectiveStartDate,
			&response.EffectiveEndDate,
			&response.CreatedBy,
			&response.CreatedAt,
			&response.UpdatedBy,
			&response.UpdatedAt,
			&response.DeletedBy,
			&response.DeletedAt,
			&response.DeletionReason,
		)...,
	); err != nil {
		if isUniqueViolation(err, "promotions_code_key") {
			err = models.ErrUniquePromotionCodeViolation
		} else {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrScan")
		}
		return nil, err
	}
	return &response, nil
}

func (q *promotionRepository) UpdatePromotion(
	ctx context.Context,
	tx *sql.Tx,
	request *models.Promotion,
) (*models.Promotion, error) {
	ctxt := "PromotionRepository-UpdatePromotion"
	var response models.Promotion
	err := tx.QueryRowContext(
		ctx,
		`SELECT
			id
			, record_status
			, name
			, code
			, deleted_code
			, description
			, discount
			, quota_per_buyer
			, effective_start_date
			, effective_end_date
			, created_by
			, created_at
			, updated_by
			, updated_at
			, deleted_by
			, deleted_at
			, deletion_reason
		FROM promotions
		WHERE id = ?
		FOR UPDATE`,
		helper.SQLArgs(request.ID)...,
	).Scan(
		helper.SQLScan(
			&response.ID,
			&response.RecordStatus,
			&response.Name,
			&response.Code,
			&response.DeletedCode,
			&response.Description,
			&response.Discount,
			&response.QuotaPerBuyer,
			&response.EffectiveStartDate,
			&response.EffectiveEndDate,
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
		return nil, models.ErrPromotionNotFound
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
	_, _ = sb.WriteString("UPDATE promotions SET")
	if response.RecordStatus != request.RecordStatus {
		params = append(params, request.RecordStatus)
		if n = len(params); n > 1 {
			_, _ = sb.WriteString(",")
		}
		_, _ = sb.WriteString(" record_status = ?")
	}
	if response.Name != request.Name {
		params = append(params, request.Name)
		if n = len(params); n > 1 {
			_, _ = sb.WriteString(",")
		}
		_, _ = sb.WriteString(" name = ?")
	}
	if (response.Code == nil && request.Code != nil) ||
		(response.Code != nil && request.Code == nil) ||
		(response.Code != nil && request.Code != nil &&
			*response.Code != *request.Code) {
		params = append(params, request.Code)
		if n = len(params); n > 1 {
			_, _ = sb.WriteString(",")
		}
		_, _ = sb.WriteString(" code = ?")
	}
	if (response.DeletedCode == nil && request.DeletedCode != nil) ||
		(response.DeletedCode != nil && request.DeletedCode == nil) ||
		(response.DeletedCode != nil && request.DeletedCode != nil &&
			*response.DeletedCode != *request.DeletedCode) {
		params = append(params, request.DeletedCode)
		if n = len(params); n > 1 {
			_, _ = sb.WriteString(",")
		}
		_, _ = sb.WriteString(" deleted_code = ?")
	}
	if response.Description != request.Description {
		params = append(params, request.Description)
		if n = len(params); n > 1 {
			_, _ = sb.WriteString(",")
		}
		_, _ = sb.WriteString(" description = ?")
	}
	if response.Discount != request.Discount {
		params = append(params, request.Discount)
		if n = len(params); n > 1 {
			_, _ = sb.WriteString(",")
		}
		_, _ = sb.WriteString(" discount = ?")
	}
	if response.QuotaPerBuyer != request.QuotaPerBuyer {
		params = append(params, request.QuotaPerBuyer)
		if n = len(params); n > 1 {
			_, _ = sb.WriteString(",")
		}
		_, _ = sb.WriteString(" quota_per_buyer = ?")
	}
	if !response.EffectiveStartDate.Equal(request.EffectiveStartDate) {
		params = append(params, request.EffectiveStartDate)
		if n = len(params); n > 1 {
			_, _ = sb.WriteString(",")
		}
		_, _ = sb.WriteString(" effective_start_date = ?")
	}
	if !response.EffectiveEndDate.Equal(request.EffectiveEndDate) {
		params = append(params, request.EffectiveEndDate)
		if n = len(params); n > 1 {
			_, _ = sb.WriteString(",")
		}
		_, _ = sb.WriteString(" effective_end_date = ?")
	}
	if len(params) == 0 {
		return &response, nil
	}
	if response.UpdatedBy.Compare(request.UpdatedBy) != 0 {
		params = append(params, request.UpdatedBy)
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
			, name
			, code
			, deleted_code
			, description
			, discount
			, quota_per_buyer
			, effective_start_date
			, effective_end_date
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
			&response.Name,
			&response.Code,
			&response.DeletedCode,
			&response.Description,
			&response.Discount,
			&response.QuotaPerBuyer,
			&response.EffectiveStartDate,
			&response.EffectiveEndDate,
			&response.CreatedBy,
			&response.CreatedAt,
			&response.UpdatedBy,
			&response.UpdatedAt,
			&response.DeletedBy,
			&response.DeletedAt,
			&response.DeletionReason,
		)...,
	); err != nil {
		if isUniqueViolation(err, "promotions_code_key") {
			err = models.ErrUniquePromotionCodeViolation
		} else {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrScan")
		}
		return nil, err
	}
	return &response, nil
}

func (q *promotionRepository) buildCondition(request *models.PromotionFilter, sb *strings.Builder) (condition string, params []any) {
	var conds []string
	if len(request.PromotionIDs) > 0 {
		sb.Reset()
		_, _ = sb.WriteString("id IN (")
		for i, promotionID := range request.PromotionIDs {
			params = append(params, promotionID)
			if i > 0 {
				_, _ = sb.WriteString(",")
			}
			_, _ = sb.WriteString("?")
		}
		_, _ = sb.WriteString(")")
		conds = append(conds, sb.String())
	}
	if len(request.ExcludedPromotionIDs) > 0 {
		sb.Reset()
		_, _ = sb.WriteString("id NOT IN (")
		for i, excludedPromotionID := range request.ExcludedPromotionIDs {
			params = append(params, excludedPromotionID)
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
	if len(request.Codes) > 0 {
		sb.Reset()
		_, _ = sb.WriteString("code IN (")
		for i, code := range request.Codes {
			params = append(params, code)
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
		params = append(params, keyword, keyword)
		sb.Reset()
		_, _ = sb.WriteString("(code LIKE ? OR name LIKE ?)")
		conds = append(conds, sb.String())
	}
	if !request.TransactionDatetime.IsZero() {
		params = append(params, request.TransactionDatetime)
		sb.Reset()
		_, _ = sb.WriteString(`(? BETWEEN effective_start_date AND effective_end_date)`)
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
