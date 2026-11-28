package repositories

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/govalues/decimal"
	"github.com/roysitumorang/laukpauk/helper"
	"github.com/roysitumorang/laukpauk/models"
	"github.com/roysitumorang/laukpauk/pools"
)

type VoucherRepository interface {
	FindVouchers(
		ctx context.Context,
		request *models.VoucherFilter,
	) ([]*models.Voucher, int64, int64, error)
	FindVoucher(
		ctx context.Context,
		request *models.VoucherFilter,
	) (*models.Voucher, error)
	CreateVoucher(
		ctx context.Context,
		tx *sql.Tx,
		request *models.Voucher,
	) (*models.Voucher, error)
	UpdateVoucher(
		ctx context.Context,
		tx *sql.Tx,
		request *models.Voucher,
	) (*models.Voucher, error)
}

type voucherRepository struct {
	dbRead,
	dbWrite *sql.DB
}

func NewVoucherRepository(
	dbRead,
	dbWrite *sql.DB,
) VoucherRepository {
	return &voucherRepository{
		dbRead:  dbRead,
		dbWrite: dbWrite,
	}
}

func (q *voucherRepository) FindVouchers(
	ctx context.Context,
	request *models.VoucherFilter,
) ([]*models.Voucher, int64, int64, error) {
	ctxt := "VoucherRepository-FindVouchers"
	sb := pools.BuilderPool.Get()
	defer func() {
		sb.Reset()
		pools.BuilderPool.Put(sb)
	}()
	sb.Reset()
	condition, params := q.buildCondition(request, sb)
	sb.Reset()
	_, _ = fmt.Fprintf(sb, "SELECT COUNT(1) FROM vouchers %s", condition)
	query := sb.String()
	var total int64
	err := q.dbRead.QueryRowContext(ctx, query, params...).Scan(&total)
	if err != nil {
		helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrScan")
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
		, voucher_type
		, target_audience
		, code
		, deleted_code
		, photo
		, name
		, discount
		, minimum_purchase
		, quota_per_buyer
		, effective_start_time
		, effective_end_time
		, effective_start_date
		, effective_end_date
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
	rows, err := q.dbRead.QueryContext(ctx, sb.String(), params...)
	if errors.Is(err, sql.ErrNoRows) {
		err = nil
	}
	if err != nil {
		helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrQueryContext")
		return nil, 0, 0, err
	}
	defer func() { _ = rows.Close() }()
	var response []*models.Voucher
	params = []any{}
	sb.Reset()
	_, _ = sb.WriteString("SELECT voucher_id, user_id FROM vouchers_users WHERE voucher_id IN (")
	var n int
	mapVoucherIDs := map[uuid.UUID]int{}
	for rows.Next() {
		var voucher models.Voucher
		if err = rows.Scan(
			&voucher.ID,
			&voucher.RecordStatus,
			&voucher.VoucherType,
			&voucher.TargetAudience,
			&voucher.Code,
			&voucher.DeletedCode,
			&voucher.Photo,
			&voucher.Name,
			&voucher.Discount,
			&voucher.MinimumPurchase,
			&voucher.QuotaPerBuyer,
			&voucher.EffectiveStartTime,
			&voucher.EffectiveEndTime,
			&voucher.EffectiveStartDate,
			&voucher.EffectiveEndDate,
			&voucher.Description,
			&voucher.CreatedBy,
			&voucher.CreatedAt,
			&voucher.UpdatedBy,
			&voucher.UpdatedAt,
			&voucher.DeletedBy,
			&voucher.DeletedAt,
			&voucher.DeletionReason,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrScan")
			return nil, 0, 0, err
		}
		offset++
		voucher.RowNo = offset
		response = append(response, (&voucher).Populate())
		mapVoucherIDs[voucher.ID] = len(response) - 1
		params = append(params, voucher.ID)
		if n = len(params); n > 1 {
			_, _ = sb.WriteString(",")
		}
		_, _ = sb.WriteString("?")
	}
	if err = rows.Err(); err != nil {
		helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrErr")
		return nil, 0, 0, err
	}
	_, _ = sb.WriteString(") ORDER BY id")
	if len(params) == 0 {
		return response, total, pages, nil
	}
	rows, err = q.dbRead.QueryContext(ctx, sb.String(), params...)
	if errors.Is(err, sql.ErrNoRows) {
		err = nil
	}
	if err != nil {
		helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrQueryContext")
		return nil, 0, 0, err
	}
	defer func() { _ = rows.Close() }()
	var (
		voucherID,
		buyerID uuid.UUID
		i int
	)
	for rows.Next() {
		if err = rows.Scan(
			&voucherID,
			&buyerID,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrScan")
			return nil, 0, 0, err
		}
		i = mapVoucherIDs[voucherID]
		voucher := response[i]
		voucher.UserIDs = append(voucher.UserIDs, buyerID)
		response[i] = voucher
	}
	if err = rows.Err(); err != nil {
		helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrErr")
		return nil, 0, 0, err
	}
	return response, total, pages, nil
}

func (q *voucherRepository) FindVoucher(
	ctx context.Context,
	request *models.VoucherFilter,
) (*models.Voucher, error) {
	ctxt := "VoucherRepository-FindVoucher"
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
			, voucher_type
			, target_audience
			, code
			, deleted_code
			, photo
			, name
			, discount
			, minimum_purchase
			, quota_per_buyer
			, effective_start_time
			, effective_end_time
			, effective_start_date
			, effective_end_date
			, description
			, created_by
			, created_at
			, updated_by
			, updated_at
			, deleted_by
			, deleted_at
			, deletion_reason
		FROM vouchers %s`,
		condition,
	)
	query := sb.String()
	var response models.Voucher
	err := q.dbRead.QueryRowContext(ctx, query, params...).Scan(
		&response.ID,
		&response.RecordStatus,
		&response.VoucherType,
		&response.TargetAudience,
		&response.Code,
		&response.DeletedCode,
		&response.Photo,
		&response.Name,
		&response.Discount,
		&response.MinimumPurchase,
		&response.QuotaPerBuyer,
		&response.EffectiveStartTime,
		&response.EffectiveEndTime,
		&response.EffectiveStartDate,
		&response.EffectiveEndDate,
		&response.Description,
		&response.CreatedBy,
		&response.CreatedAt,
		&response.UpdatedBy,
		&response.UpdatedAt,
		&response.DeletedBy,
		&response.DeletedAt,
		&response.DeletionReason,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, models.ErrVoucherNotFound
	}
	if err != nil {
		helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrScan")
		return nil, err
	}
	rows, err := q.dbRead.QueryContext(
		ctx,
		`SELECT user_id
		FROM vouchers_users
		WHERE voucher_id = ?
		ORDER BY id`,
		response.ID,
	)
	if errors.Is(err, sql.ErrNoRows) {
		err = nil
	}
	if err != nil {
		helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrQueryContext")
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	var buyerID uuid.UUID
	for rows.Next() {
		if err = rows.Scan(&buyerID); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrScan")
			return nil, err
		}
		response.UserIDs = append(response.UserIDs, buyerID)
	}
	if err = rows.Err(); err != nil {
		helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrErr")
		return nil, err
	}
	return (&response).Populate(), nil
}

func (q *voucherRepository) CreateVoucher(
	ctx context.Context,
	tx *sql.Tx,
	request *models.Voucher,
) (*models.Voucher, error) {
	ctxt := "VoucherRepository-CreateVoucher"
	var response models.Voucher
	if err := tx.QueryRowContext(
		ctx,
		`INSERT INTO vouchers (
			record_status
			, voucher_type
			, target_audience
			, code
			, photo
			, name
			, discount
			, minimum_purchase
			, quota_per_buyer
			, effective_start_time
			, effective_end_time
			, effective_start_date
			, effective_end_date
			, description
			, created_by
			, created_at
			, updated_by
			, updated_at
		) VALUES (
			?, ?, ?, ?, ?, ?, ?, ?, ?,
			?, ?, ?, ?, ?, ?, ?, ?, ?
		)
		RETURNING id
			, record_status
			, voucher_type
			, target_audience
			, code
			, deleted_code
			, photo
			, name
			, discount
			, minimum_purchase
			, quota_per_buyer
			, effective_start_time
			, effective_end_time
			, effective_start_date
			, effective_end_date
			, description
			, created_by
			, created_at
			, updated_by
			, updated_at
			, deleted_by
			, deleted_at
			, deletion_reason`,
		request.RecordStatus,
		request.VoucherType,
		request.TargetAudience,
		request.Code,
		request.Photo,
		request.Name,
		request.Discount,
		request.MinimumPurchase,
		request.QuotaPerBuyer,
		request.EffectiveStartTime,
		request.EffectiveEndTime,
		request.EffectiveStartDate,
		request.EffectiveEndDate,
		request.Description,
		request.CreatedBy,
		request.CreatedAt,
		request.UpdatedBy,
		request.UpdatedAt,
	).Scan(
		&response.ID,
		&response.RecordStatus,
		&response.VoucherType,
		&response.TargetAudience,
		&response.Code,
		&response.DeletedCode,
		&response.Photo,
		&response.Name,
		&response.Discount,
		&response.MinimumPurchase,
		&response.QuotaPerBuyer,
		&response.EffectiveStartTime,
		&response.EffectiveEndTime,
		&response.EffectiveStartDate,
		&response.EffectiveEndDate,
		&response.Description,
		&response.CreatedBy,
		&response.CreatedAt,
		&response.UpdatedBy,
		&response.UpdatedAt,
		&response.DeletedBy,
		&response.DeletedAt,
		&response.DeletionReason,
	); err != nil {
		if isUniqueViolation(err, "vouchers_unique_code_idx") {
			err = models.ErrUniqueVoucherCodeViolation
		} else {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrScan")
		}
		return nil, err
	}
	return (&response).Populate(), nil
}

func (q *voucherRepository) UpdateVoucher(
	ctx context.Context,
	tx *sql.Tx,
	request *models.Voucher,
) (*models.Voucher, error) {
	ctxt := "VoucherRepository-UpdateVoucher"
	var response models.Voucher
	err := tx.QueryRowContext(
		ctx,
		`SELECT
			id
			, record_status
			, voucher_type
			, target_audience
			, code
			, deleted_code
			, photo
			, name
			, discount
			, minimum_purchase
			, quota_per_buyer
			, effective_start_time
			, effective_end_time
			, effective_start_date
			, effective_end_date
			, description
			, created_by
			, created_at
			, updated_by
			, updated_at
			, deleted_by
			, deleted_at
			, deletion_reason
		FROM vouchers
		WHERE id = ?
		FOR UPDATE`,
		request.ID,
	).Scan(
		&response.ID,
		&response.RecordStatus,
		&response.VoucherType,
		&response.TargetAudience,
		&response.Code,
		&response.DeletedCode,
		&response.Photo,
		&response.Name,
		&response.Discount,
		&response.MinimumPurchase,
		&response.QuotaPerBuyer,
		&response.EffectiveStartTime,
		&response.EffectiveEndTime,
		&response.EffectiveStartDate,
		&response.EffectiveEndDate,
		&response.Description,
		&response.CreatedBy,
		&response.CreatedAt,
		&response.UpdatedBy,
		&response.UpdatedAt,
		&response.DeletedBy,
		&response.DeletedAt,
		&response.DeletionReason,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, models.ErrVoucherNotFound
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
	_, _ = sb.WriteString("UPDATE vouchers SET")
	if response.RecordStatus != request.RecordStatus {
		params = append(params, request.RecordStatus)
		if n = len(params); n > 1 {
			_, _ = sb.WriteString(",")
		}
		_, _ = sb.WriteString(" record_status = ?")
	}
	if response.VoucherType != request.VoucherType {
		params = append(params, request.VoucherType)
		if n = len(params); n > 1 {
			_, _ = sb.WriteString(",")
		}
		_, _ = sb.WriteString(" voucher_type = ?")
	}
	if response.TargetAudience != request.TargetAudience {
		params = append(params, request.TargetAudience)
		if n = len(params); n > 1 {
			_, _ = sb.WriteString(",")
		}
		_, _ = sb.WriteString(" target_audience = ?")
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
	if response.Photo != request.Photo {
		params = append(params, request.Photo)
		if n = len(params); n > 1 {
			_, _ = sb.WriteString(",")
		}
		_, _ = sb.WriteString(" photo = ?")
	}
	if response.Name != request.Name {
		params = append(params, request.Name)
		if n = len(params); n > 1 {
			_, _ = sb.WriteString(",")
		}
		_, _ = sb.WriteString(" name = ?")
	}
	if response.Discount != request.Discount {
		params = append(params, request.Discount)
		if n = len(params); n > 1 {
			_, _ = sb.WriteString(",")
		}
		_, _ = sb.WriteString(" discount = ?")
	}
	if response.MinimumPurchase != request.MinimumPurchase {
		params = append(params, request.MinimumPurchase)
		if n = len(params); n > 1 {
			_, _ = sb.WriteString(",")
		}
		_, _ = sb.WriteString(" minimum_purchase = ?")
	}
	if response.QuotaPerBuyer != request.QuotaPerBuyer {
		params = append(params, request.QuotaPerBuyer)
		if n = len(params); n > 1 {
			_, _ = sb.WriteString(",")
		}
		_, _ = sb.WriteString(" quota_per_buyer = ?")
	}
	if !response.EffectiveStartTime.Equal(request.EffectiveStartTime) {
		params = append(params, request.EffectiveStartTime)
		if n = len(params); n > 1 {
			_, _ = sb.WriteString(",")
		}
		_, _ = sb.WriteString(" effective_start_time = ?")
	}
	if !response.EffectiveEndTime.Equal(request.EffectiveEndTime) {
		params = append(params, request.EffectiveEndTime)
		if n = len(params); n > 1 {
			_, _ = sb.WriteString(",")
		}
		_, _ = sb.WriteString(" effective_end_time = ?")
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
	if response.Description != request.Description {
		params = append(params, request.Description)
		if n = len(params); n > 1 {
			_, _ = sb.WriteString(",")
		}
		_, _ = sb.WriteString(" description = ?")
	}
	if len(params) > 0 {
		if response.UpdatedBy != request.UpdatedBy {
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
				*response.DeletedBy != *request.DeletedBy) {
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
				, voucher_type
				, target_audience
				, code
				, deleted_code
				, photo
				, name
				, discount
				, minimum_purchase
				, quota_per_buyer
				, effective_start_time
				, effective_end_time
				, effective_start_date
				, effective_end_date
				, description
				, created_by
				, created_at
				, updated_by
				, updated_at
				, deleted_by
				, deleted_at
				, deletion_reason`)
		if err = tx.QueryRowContext(ctx, sb.String(), params...).Scan(
			&response.ID,
			&response.RecordStatus,
			&response.VoucherType,
			&response.TargetAudience,
			&response.Code,
			&response.DeletedCode,
			&response.Photo,
			&response.Name,
			&response.Discount,
			&response.MinimumPurchase,
			&response.QuotaPerBuyer,
			&response.EffectiveStartTime,
			&response.EffectiveEndTime,
			&response.EffectiveStartDate,
			&response.EffectiveEndDate,
			&response.Description,
			&response.CreatedBy,
			&response.CreatedAt,
			&response.UpdatedBy,
			&response.UpdatedAt,
			&response.DeletedBy,
			&response.DeletedAt,
			&response.DeletionReason,
		); err != nil {
			if isUniqueViolation(err, "vouchers_unique_code_idx") {
				err = models.ErrUniqueVoucherCodeViolation
			} else {
				helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrScan")
			}
			return nil, err
		}
	}
	rows, err := tx.QueryContext(
		ctx,
		`SELECT user_id
		FROM vouchers_users
		WHERE voucher_id = ?
		ORDER BY id`,
		request.ID,
	)
	if errors.Is(err, sql.ErrNoRows) {
		err = nil
	}
	if err != nil {
		helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrQueryContext")
		return nil, err
	}
	mapBuyerIDs := map[uuid.UUID]struct{}{}
	var buyerID uuid.UUID
	for rows.Next() {
		if err = rows.Scan(&buyerID); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrScan")
			return nil, err
		}
		mapBuyerIDs[buyerID] = struct{}{}
	}
	if err = rows.Err(); err != nil {
		helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrErr")
		return nil, err
	}
	params = params[:0]
	params = append(params, request.ID)
	sb.Reset()
	_, _ = sb.WriteString("DELETE FROM vouchers_users WHERE voucher_id = ?")
	if len(request.UserIDs) > 0 {
		_, _ = sb.WriteString(" AND user_id NOT IN (")
	}
	for _, buyerID := range request.UserIDs {
		params = append(params, buyerID)
		if n = len(params); n > 2 {
			_, _ = sb.WriteString(",")
		}
		_, _ = sb.WriteString(" ?")
		if _, ok := mapBuyerIDs[buyerID]; ok {
			continue
		}
		if _, err = tx.ExecContext(
			ctx,
			`INSERT INTO vouchers_users (
				voucher_id
				, user_id
				, created_at
			) VALUES (?, ?, ?)`,
			request.ID,
			buyerID,
			request.UpdatedAt,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExecContext")
			return nil, err
		}
	}
	if len(request.UserIDs) > 0 {
		_, _ = sb.WriteString(")")
	}
	if _, err = tx.ExecContext(ctx, sb.String(), params...); err != nil {
		helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExecContext")
		return nil, err
	}
	return (&response).Populate(), nil
}

func (q *voucherRepository) buildCondition(request *models.VoucherFilter, sb *strings.Builder) (condition string, params []any) {
	var conds []string
	if len(request.VoucherIDs) > 0 {
		sb.Reset()
		_, _ = sb.WriteString("id IN (")
		for i, voucherID := range request.VoucherIDs {
			params = append(params, voucherID)
			if i > 0 {
				_, _ = sb.WriteString(",")
			}
			_, _ = sb.WriteString("?")
		}
		_, _ = sb.WriteString(")")
		conds = append(conds, sb.String())
	}
	if len(request.ExcludedVoucherIDs) > 0 {
		sb.Reset()
		_, _ = sb.WriteString("id NOT IN (")
		for i, excludedVoucherID := range request.ExcludedVoucherIDs {
			params = append(params, excludedVoucherID)
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
	if len(request.VoucherTypes) > 0 {
		sb.Reset()
		_, _ = sb.WriteString("voucher_type IN (")
		for i, voucherType := range request.VoucherTypes {
			params = append(params, voucherType)
			if i > 0 {
				_, _ = sb.WriteString(",")
			}
			_, _ = sb.WriteString("?")
		}
		_, _ = sb.WriteString(")")
		conds = append(conds, sb.String())
	}
	if len(request.TargetAudiences) > 0 {
		sb.Reset()
		_, _ = sb.WriteString("target_audience IN (")
		for i, targetAudience := range request.TargetAudiences {
			params = append(params, targetAudience)
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
		_, _ = sb.WriteString("LOWER(code) IN (")
		for i, name := range request.Codes {
			params = append(params, strings.ToLower(name))
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
		_, _ = sb.WriteString("(LOWER(code) LIKE ? OR LOWER(name) LIKE ?)")
		conds = append(conds, sb.String())
	}
	if len(request.UserIDs) > 0 {
		params = append(params, models.VoucherTargetAudienceAllBuyers)
		sb.Reset()
		_, _ = sb.WriteString(`(target_audience = ? OR EXISTS(
				SELECT 1
				FROM vouchers_users
				WHERE vouchers_users.voucher_id = vouchers.id
					AND vouchers_users.user_id IN (`)
		for i, buyerID := range request.UserIDs {
			params = append(params, buyerID)
			if i > 0 {
				_, _ = sb.WriteString(",")
			}
			_, _ = sb.WriteString("?")
		}
		_, _ = sb.WriteString(")))")
		conds = append(conds, sb.String())
	}
	if !request.TransactionDatetime.IsZero() {
		t := time.Date(
			1,
			1,
			1,
			request.TransactionDatetime.Hour(),
			request.TransactionDatetime.Minute(),
			request.TransactionDatetime.Second(),
			request.TransactionDatetime.Nanosecond(),
			request.TransactionDatetime.Location(),
		)
		params = append(params, request.TransactionDatetime, t)
		sb.Reset()
		_, _ = sb.WriteString(`(? BETWEEN effective_start_date AND effective_end_date
			AND ? BETWEEN effective_start_time AND effective_end_time)`)
		conds = append(conds, sb.String())
	}
	if request.Purchase > 0 {
		params = append(params, request.Purchase)
		sb.Reset()
		_, _ = sb.WriteString("minimum_purchase <= ?")
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
