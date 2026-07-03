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
	SettingRepository interface {
		FindSettings(
			ctx context.Context,
			request *models.SettingFilter,
		) ([]*models.Setting, int64, int64, error)
		FindSetting(
			ctx context.Context,
			request *models.SettingFilter,
		) (*models.Setting, error)
		CreateSetting(
			ctx context.Context,
			tx pgx.Tx,
			request *models.Setting,
		) (*models.Setting, error)
		UpdateSetting(
			ctx context.Context,
			tx pgx.Tx,
			request *models.Setting,
		) (*models.Setting, error)
	}

	settingRepository struct {
		dbRead,
		dbWrite *pgxpool.Pool
	}
)

func NewSettingRepository(
	dbRead,
	dbWrite *pgxpool.Pool,
) SettingRepository {
	return new(settingRepository{
		dbRead:  dbRead,
		dbWrite: dbWrite,
	})
}

func (q *settingRepository) FindSettings(
	ctx context.Context,
	request *models.SettingFilter,
) ([]*models.Setting, int64, int64, error) {
	ctxt := "SettingRepository-FindSettings"
	var sb strings.Builder
	condition, params := q.buildCondition(request, &sb)
	sb.Reset()
	_, _ = fmt.Fprintf(&sb, "SELECT COUNT(1) FROM settings %s", condition)
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
		, active
		, key
		, value
		, description
		, created_by
		, created_at
		, updated_by
		, updated_at`,
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
	var response []*models.Setting
	for rows.Next() {
		var setting models.Setting
		if err = rows.Scan(
			&setting.ID,
			&setting.Active,
			&setting.Key,
			&setting.Value,
			&setting.Description,
			&setting.CreatedBy,
			&setting.CreatedAt,
			&setting.UpdatedBy,
			&setting.UpdatedAt,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrScan")
			return nil, 0, 0, err
		}
		offset++
		setting.RowNo = offset
		response = append(response, &setting)
	}
	return response, total, pages, nil
}

func (q *settingRepository) FindSetting(
	ctx context.Context,
	request *models.SettingFilter,
) (*models.Setting, error) {
	ctxt := "SettingRepository-FindSetting"
	var sb strings.Builder
	condition, params := q.buildCondition(request, &sb)
	sb.Reset()
	_, _ = fmt.Fprintf(
		&sb,
		`SELECT
			id
			, active
			, key
			, value
			, description
			, created_by
			, created_at
			, updated_by
			, updated_at
		FROM settings %s`,
		condition,
	)
	query := sb.String()
	var response models.Setting
	err := q.dbRead.QueryRow(ctx, query, params...).Scan(
		&response.ID,
		&response.Active,
		&response.Key,
		&response.Value,
		&response.Description,
		&response.CreatedBy,
		&response.CreatedAt,
		&response.UpdatedBy,
		&response.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, models.ErrSettingNotFound
	}
	if err != nil {
		helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrScan")
		return nil, err
	}
	return new(response), nil
}

func (q *settingRepository) CreateSetting(
	ctx context.Context,
	tx pgx.Tx,
	request *models.Setting,
) (*models.Setting, error) {
	ctxt := "SettingRepository-CreateSetting"
	var response models.Setting
	if err := tx.QueryRow(
		ctx,
		`INSERT INTO settings (
			active
			, key
			, value
			, description
			, created_by
			, created_at
			, updated_by
			, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id
			, active
			, key
			, value
			, description
			, created_by
			, created_at
			, updated_by
			, updated_at`,
		request.Active,
		request.Key,
		request.Value,
		request.Description,
		request.CreatedBy,
		request.CreatedAt,
		request.UpdatedBy,
		request.UpdatedAt,
	).Scan(
		&response.ID,
		&response.Active,
		&response.Key,
		&response.Value,
		&response.Description,
		&response.CreatedBy,
		&response.CreatedAt,
		&response.UpdatedBy,
		&response.UpdatedAt,
	); err != nil {
		if errRollback := tx.Rollback(ctx); errRollback != nil {
			helper.Capture(ctx, slog.LevelError, errRollback, ctxt, "ErrRollback")
		}
		if pgxErr, ok := errors.AsType[*pgconn.PgError](err); ok &&
			pgxErr.Code == pgerrcode.UniqueViolation &&
			pgxErr.ConstraintName == "settings_unique_key_idx" {
			err = models.ErrUniqueSettingKeyViolation
		} else {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrScan")
		}
		return nil, err
	}
	return new(response), nil
}

func (q *settingRepository) UpdateSetting(
	ctx context.Context,
	tx pgx.Tx,
	request *models.Setting,
) (*models.Setting, error) {
	ctxt := "SettingRepository-UpdateSetting"
	var response models.Setting
	if err := tx.QueryRow(
		ctx,
		`UPDATE settings SET
			active = $1
			, key = $2
			, value = $3
			, description = $4
			, updated_by = $5
			, updated_at = $6
		WHERE id = $7
		RETURNING id
			, active
			, key
			, value
			, description
			, created_by
			, created_at
			, updated_by
			, updated_at`,
		request.Active,
		request.Key,
		request.Value,
		request.Description,
		request.UpdatedBy,
		request.UpdatedAt,
		request.ID,
	).Scan(
		&response.ID,
		&response.Active,
		&response.Key,
		&response.Value,
		&response.Description,
		&response.CreatedBy,
		&response.CreatedAt,
		&response.UpdatedBy,
		&response.UpdatedAt,
	); err != nil {
		if errRollback := tx.Rollback(ctx); errRollback != nil {
			helper.Capture(ctx, slog.LevelError, errRollback, ctxt, "ErrRollback")
		}
		if pgxErr, ok := errors.AsType[*pgconn.PgError](err); ok &&
			pgxErr.Code == pgerrcode.UniqueViolation &&
			pgxErr.ConstraintName == "settings_unique_key_idx" {
			err = models.ErrUniqueSettingKeyViolation
		} else {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrScan")
		}
		return nil, err
	}
	return new(response), nil
}

func (q *settingRepository) buildCondition(request *models.SettingFilter, sb *strings.Builder) (condition string, params []any) {
	var conds []string
	if len(request.SettingIDs) > 0 {
		sb.Reset()
		_, _ = sb.WriteString("id IN (")
		for i, settingID := range request.SettingIDs {
			params = append(params, settingID)
			if i > 0 {
				_, _ = sb.WriteString(",")
			}
			_, _ = fmt.Fprintf(sb, "$%d", len(params))
		}
		_, _ = sb.WriteString(")")
		conds = append(conds, sb.String())
	}
	if len(request.ExcludedSettingIDs) > 0 {
		sb.Reset()
		_, _ = sb.WriteString("id NOT IN (")
		for i, excludedSettingID := range request.ExcludedSettingIDs {
			params = append(params, excludedSettingID)
			if i > 0 {
				_, _ = sb.WriteString(",")
			}
			_, _ = fmt.Fprintf(sb, "$%d", len(params))
		}
		_, _ = sb.WriteString(")")
		conds = append(conds, sb.String())
	}
	if request.Active != nil {
		if *request.Active {
			conds = append(conds, "active")
		} else {
			conds = append(conds, "NOT active")
		}
	}
	if len(request.Keys) > 0 {
		sb.Reset()
		_, _ = sb.WriteString("LOWER(key) IN (")
		for i, key := range request.Keys {
			params = append(params, strings.ToLower(key))
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
		sb.Reset()
		_, _ = fmt.Fprintf(sb, "LOWER(key) LIKE $%d", len(params))
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
