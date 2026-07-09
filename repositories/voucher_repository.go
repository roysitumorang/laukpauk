package repositories

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"

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
		, code
		, photo
		, name
		, budget
		, discount
		, minimum_purchase
		, maximum_usage
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
	for rows.Next() {
		var voucher models.Voucher
		if err = rows.Scan(
			&voucher.ID,
			&voucher.RecordStatus,
			&voucher.VoucherType,
			&voucher.Code,
			&voucher.Photo,
			&voucher.Name,
			&voucher.Budget,
			&voucher.Discount,
			&voucher.MinimumPurchase,
			&voucher.MaximumUsage,
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
			, code
			, photo
			, name
			, budget
			, discount
			, minimum_purchase
			, maximum_usage
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
		&response.Code,
		&response.Photo,
		&response.Name,
		&response.Budget,
		&response.Discount,
		&response.MinimumPurchase,
		&response.MaximumUsage,
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
			, code
			, photo
			, name
			, budget
			, discount
			, minimum_purchase
			, maximum_usage
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
			, code
			, photo
			, name
			, budget
			, discount
			, minimum_purchase
			, maximum_usage
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
		request.Code,
		request.Photo,
		request.Name,
		request.Budget,
		request.Discount,
		request.MinimumPurchase,
		request.MaximumUsage,
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
		&response.Code,
		&response.Photo,
		&response.Name,
		&response.Budget,
		&response.Discount,
		&response.MinimumPurchase,
		&response.MaximumUsage,
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
			, code = $3
			, photo = $4
			, name = $5
			, budget = $6
			, discount = $7
			, minimum_purchase = $8
			, maximum_usage = $9
			, effective_start_time = $10
			, effective_end_time = $11
			, effective_start_date = $12
			, effective_end_date = $13
			, description = $14
			, updated_by = $15
			, updated_at = $16
			, deleted_by = $17
			, deleted_at = $18
			, deletion_reason = $19
		WHERE id = $20
		RETURNING id
			, record_status
			, voucher_type
			, code
			, photo
			, name
			, budget
			, discount
			, minimum_purchase
			, maximum_usage
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
		request.Code,
		request.Photo,
		request.Name,
		request.Budget,
		request.Discount,
		request.MinimumPurchase,
		request.MaximumUsage,
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
		&response.Code,
		&response.Photo,
		&response.Name,
		&response.Budget,
		&response.Discount,
		&response.MinimumPurchase,
		&response.MaximumUsage,
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
