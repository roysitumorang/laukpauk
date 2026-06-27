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
	MetricUnitRepository interface {
		FindMetricUnits(
			ctx context.Context,
			request *models.MetricUnitFilter,
		) ([]*models.MetricUnit, int64, int64, error)
		CreateMetricUnit(
			ctx context.Context,
			tx pgx.Tx,
			request *models.MetricUnit,
		) (*models.MetricUnit, error)
		UpdateMetricUnit(
			ctx context.Context,
			tx pgx.Tx,
			request *models.MetricUnit,
		) (*models.MetricUnit, error)
	}

	metricUnitRepository struct {
		dbRead,
		dbWrite *pgxpool.Pool
	}
)

func NewMetricUnitRepository(
	dbRead,
	dbWrite *pgxpool.Pool,
) MetricUnitRepository {
	return new(metricUnitRepository{
		dbRead:  dbRead,
		dbWrite: dbWrite,
	})
}

func (q *metricUnitRepository) FindMetricUnits(
	ctx context.Context,
	request *models.MetricUnitFilter,
) ([]*models.MetricUnit, int64, int64, error) {
	ctxt := "MetricUnitRepository-FindMetricUnits"
	var (
		params []any
		conds  []string
		sb     strings.Builder
	)
	if len(request.MetricUnitIDs) > 0 {
		sb.Reset()
		_, _ = sb.WriteString("id IN (")
		for i, metricUnitID := range request.MetricUnitIDs {
			params = append(params, metricUnitID)
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
		_, _ = sb.WriteString("LOWER(name) LIKE $")
		_, _ = sb.WriteString(n)
		conds = append(conds, sb.String())
	}
	sb.Reset()
	_, _ = sb.WriteString(
		`SELECT COUNT(1)
		FROM metric_units`,
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
	var response []*models.MetricUnit
	for rows.Next() {
		var metricUnit models.MetricUnit
		if err = rows.Scan(
			&metricUnit.RowNo,
			&metricUnit.ID,
			&metricUnit.RecordStatus,
			&metricUnit.Name,
			&metricUnit.CreatedBy,
			&metricUnit.CreatedAt,
			&metricUnit.UpdatedBy,
			&metricUnit.UpdatedAt,
			&metricUnit.DeletedBy,
			&metricUnit.DeletedAt,
			&metricUnit.DeletionReason,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrScan")
			return nil, 0, 0, err
		}
		response = append(response, &metricUnit)
	}
	return response, total, pages, nil
}

func (q *metricUnitRepository) CreateMetricUnit(
	ctx context.Context,
	tx pgx.Tx,
	request *models.MetricUnit,
) (*models.MetricUnit, error) {
	ctxt := "MetricUnitRepository-CreateMetricUnit"
	var response models.MetricUnit
	if err := tx.QueryRow(
		ctx,
		`INSERT INTO metric_units (
			record_status
			, name
			, created_by
			, created_at
			, updated_by
			, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id
			, record_status
			, name
			, created_by
			, created_at
			, updated_by
			, updated_at
			, deleted_by
			, deleted_at
			, deletion_reason`,
		request.RecordStatus,
		request.Name,
		request.CreatedBy,
		request.CreatedAt,
		request.UpdatedBy,
		request.UpdatedAt,
	).Scan(
		&response.ID,
		&response.RecordStatus,
		&response.Name,
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

func (q *metricUnitRepository) UpdateMetricUnit(
	ctx context.Context,
	tx pgx.Tx,
	request *models.MetricUnit,
) (*models.MetricUnit, error) {
	ctxt := "MetricUnitRepository-UpdateMetricUnit"
	var response models.MetricUnit
	if err := tx.QueryRow(
		ctx,
		`UPDATE metric_units SET
			record_status = $1
			, name = $2
			, updated_by = $3
			, updated_at = $4
			, deleted_by = $5
			, deleted_at = $6
			, deletion_reason = $7
		WHERE id = $8
		RETURNING id
			, record_status
			, name
			, created_by
			, created_at
			, updated_by
			, updated_at
			, deleted_by
			, deleted_at
			, deletion_reason`,
		request.RecordStatus,
		request.Name,
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
