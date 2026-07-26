package repositories

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"
	"uuid"

	"github.com/govalues/decimal"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/roysitumorang/laukpauk/helper"
	"github.com/roysitumorang/laukpauk/models"
)

type (
	VoucherRepository interface {
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
			tx pgx.Tx,
			request *models.Voucher,
		) (*models.Voucher, error)
		UpdateVoucher(
			ctx context.Context,
			tx pgx.Tx,
			request *models.Voucher,
		) (*models.Voucher, error)
	}

	voucherRepository struct {
		dbRead,
		dbWrite *pgxpool.Pool
	}
)

func NewVoucherRepository(
	dbRead,
	dbWrite *pgxpool.Pool,
) VoucherRepository {
	return new(voucherRepository{
		dbRead:  dbRead,
		dbWrite: dbWrite,
	})
}

func (q *voucherRepository) FindVouchers(
	ctx context.Context,
	request *models.VoucherFilter,
) ([]*models.Voucher, int64, int64, error) {
	ctxt := "VoucherRepository-FindVouchers"
	var sb strings.Builder
	condition, params := q.buildCondition(request, &sb)
	sb.Reset()
	_, _ = fmt.Fprintf(&sb, "SELECT COUNT(1) FROM vouchers %s", condition)
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
		_, _ = fmt.Fprintf(&sb, " %s %s", item.Field, item.Direction)
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
		_, _ = fmt.Fprintf(&sb, " LIMIT %d OFFSET %d", request.Limit, offset)
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
	var response []*models.Voucher
	params = []any{}
	sb.Reset()
	_, _ = sb.WriteString("SELECT voucher_id, buyer_id FROM vouchers_buyers WHERE voucher_id IN (")
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
		_, _ = fmt.Fprintf(&sb, "$%d", n)
	}
	_, _ = sb.WriteString(") ORDER BY id")
	if len(params) == 0 {
		return response, total, pages, nil
	}
	rows, err = q.dbRead.Query(ctx, sb.String(), params...)
	if errors.Is(err, pgx.ErrNoRows) {
		err = nil
	}
	if err != nil {
		helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrQuery")
		return nil, 0, 0, err
	}
	defer rows.Close()
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
		voucher.BuyerIDs = append(voucher.BuyerIDs, buyerID)
		response[i] = voucher
	}
	return response, total, pages, nil
}

func (q *voucherRepository) FindVoucher(
	ctx context.Context,
	request *models.VoucherFilter,
) (*models.Voucher, error) {
	ctxt := "VoucherRepository-FindVoucher"
	var sb strings.Builder
	condition, params := q.buildCondition(request, &sb)
	sb.Reset()
	_, _ = fmt.Fprintf(
		&sb,
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
	err := q.dbRead.QueryRow(ctx, query, params...).Scan(
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
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, models.ErrVoucherNotFound
	}
	if err != nil {
		helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrScan")
		return nil, err
	}
	rows, err := q.dbRead.Query(
		ctx,
		`SELECT buyer_id
		FROM vouchers_buyers
		WHERE voucher_id = $1
		ORDER BY id`,
		response.ID,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		err = nil
	}
	if err != nil {
		helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrQuery")
		return nil, err
	}
	defer rows.Close()
	var buyerID uuid.UUID
	for rows.Next() {
		if err = rows.Scan(&buyerID); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrScan")
			return nil, err
		}
		response.BuyerIDs = append(response.BuyerIDs, buyerID)
	}
	return (&response).Populate(), nil
}

func (q *voucherRepository) CreateVoucher(
	ctx context.Context,
	tx pgx.Tx,
	request *models.Voucher,
) (*models.Voucher, error) {
	ctxt := "VoucherRepository-CreateVoucher"
	var response models.Voucher
	if err := tx.QueryRow(
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
			$1, $2, $3, $4, $5, $6, $7, $8, $9,
			$10, $11, $12, $13, $14, $15, $16, $17, $18
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
		if pgxErr, ok := errors.AsType[*pgconn.PgError](err); ok &&
			pgxErr.Code == pgerrcode.UniqueViolation &&
			pgxErr.ConstraintName == "vouchers_unique_code_idx" {
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
	tx pgx.Tx,
	request *models.Voucher,
) (*models.Voucher, error) {
	ctxt := "VoucherRepository-UpdateVoucher"
	var response models.Voucher
	if err := tx.QueryRow(
		ctx,
		`UPDATE vouchers SET
			record_status = $1
			, voucher_type = $2
			, target_audience = $3
			, code = $4
			, deleted_code = $5
			, photo = $6
			, name = $7
			, discount = $8
			, minimum_purchase = $9
			, quota_per_buyer = $10
			, effective_start_time = $11
			, effective_end_time = $12
			, effective_start_date = $13
			, effective_end_date = $14
			, description = $15
			, updated_by = $16
			, updated_at = $17
			, deleted_by = $18
			, deleted_at = $19
			, deletion_reason = $20
		WHERE id = $21
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
		request.DeletedCode,
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
		request.UpdatedBy,
		request.UpdatedAt,
		request.DeletedBy,
		request.DeletedAt,
		request.DeletionReason,
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
	); err != nil {
		if pgxErr, ok := errors.AsType[*pgconn.PgError](err); ok &&
			pgxErr.Code == pgerrcode.UniqueViolation &&
			pgxErr.ConstraintName == "vouchers_unique_code_idx" {
			err = models.ErrUniqueVoucherCodeViolation
		} else {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrScan")
		}
		return nil, err
	}
	n := len(request.BuyerIDs)
	if n == 0 {
		return (&response).Populate(), nil
	}
	paramsCreate := make([]any, n+2)
	paramsCreate[0] = response.ID
	paramsCreate[1] = response.UpdatedAt
	paramsDelete := make([]any, n+1)
	paramsDelete[0] = response.ID
	var sbCreate,
		sbDelete strings.Builder
	_, _ = sbCreate.WriteString("INSERT INTO vouchers_buyers (voucher_id, buyer_id, created_at) VALUES")
	_, _ = sbDelete.WriteString("DELETE FROM vouchers_buyers WHERE voucher_id = $1 AND buyer_id NOT IN (")
	for i := range request.BuyerIDs {
		paramsCreate[i+2] = request.BuyerIDs[i]
		paramsDelete[i+1] = request.BuyerIDs[i]
		if i > 0 {
			_, _ = sbCreate.WriteString(",")
			_, _ = sbDelete.WriteString(",")
		}
		_, _ = fmt.Fprintf(&sbCreate, " ($1, $%d, $2)", i+3)
		_, _ = fmt.Fprintf(&sbDelete, "$%d", i+2)
	}
	_, _ = sbCreate.WriteString(" ON CONFLICT (voucher_id, buyer_id) DO UPDATE SET id = EXCLUDED.id RETURNING buyer_id")
	_, _ = sbDelete.WriteString(")")
	rows, err := tx.Query(ctx, sbCreate.String(), paramsCreate...)
	if errors.Is(err, pgx.ErrNoRows) {
		err = nil
	}
	if err != nil {
		helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrQuery")
		return nil, err
	}
	var buyerID uuid.UUID
	for rows.Next() {
		if err = rows.Scan(&buyerID); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrScan")
			return nil, err
		}
		response.BuyerIDs = append(response.BuyerIDs, buyerID)
	}
	if err = rows.Err(); err != nil {
		helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrErr")
		return nil, err
	}
	if _, err = tx.Exec(ctx, sbDelete.String(), paramsDelete...); err != nil {
		helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
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
			_, _ = fmt.Fprintf(sb, "$%d", len(params))
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
			_, _ = fmt.Fprintf(sb, "$%d", len(params))
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
			_, _ = fmt.Fprintf(sb, "$%d", len(params))
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
			_, _ = fmt.Fprintf(sb, "$%d", len(params))
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
			_, _ = fmt.Fprintf(sb, "$%d", len(params))
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
		_, _ = fmt.Fprintf(sb, "(LOWER(code) LIKE $%d OR LOWER(name) LIKE $%d)", n, n)
		conds = append(conds, sb.String())
	}
	if len(request.BuyerIDs) > 0 {
		params = append(params, models.VoucherTargetAudienceAllBuyers)
		sb.Reset()
		_, _ = fmt.Fprintf(
			sb,
			`(target_audience = $%d OR EXISTS(
				SELECT 1
				FROM vouchers_buyers
				WHERE vouchers_buyers.voucher_id = vouchers.id
					AND vouchers_buyers.buyer_id IN (`,
			len(params),
		)
		for i, buyerID := range request.BuyerIDs {
			params = append(params, buyerID)
			if i > 0 {
				_, _ = sb.WriteString(",")
			}
			_, _ = fmt.Fprintf(sb, "$%d", len(params))
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
		n := len(params)
		sb.Reset()
		_, _ = fmt.Fprintf(
			sb,
			`($%d BETWEEN effective_start_date AND effective_end_date
			AND $%d BETWEEN effective_start_time AND effective_end_time)`,
			n-1, n,
		)
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
