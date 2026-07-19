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
	VariableRepository interface {
		FindVariables(
			ctx context.Context,
			request *models.VariableFilter,
		) ([]*models.Variable, int64, int64, error)
		FindVariable(
			ctx context.Context,
			request *models.VariableFilter,
		) (*models.Variable, error)
		SaveVariable(
			ctx context.Context,
			tx pgx.Tx,
			request *models.Variable,
		) (*models.Variable, error)
	}

	variableRepository struct {
		dbRead,
		dbWrite *pgxpool.Pool
	}
)

func NewVariableRepository(
	dbRead,
	dbWrite *pgxpool.Pool,
) VariableRepository {
	return &variableRepository{
		dbRead:  dbRead,
		dbWrite: dbWrite,
	}
}

func (q *variableRepository) FindVariables(
	ctx context.Context,
	request *models.VariableFilter,
) ([]*models.Variable, int64, int64, error) {
	ctxt := "VariableRepository-FindVariables"
	var sb strings.Builder
	condition, params := q.buildCondition(request, &sb)
	sb.Reset()
	_, _ = fmt.Fprintf(&sb, "SELECT COUNT(1) FROM variables %s", condition)
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
		, name
		, value
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
	var response []*models.Variable
	for rows.Next() {
		var variable models.Variable
		if err = rows.Scan(
			&variable.ID,
			&variable.Name,
			&variable.Value,
			&variable.CreatedBy,
			&variable.CreatedAt,
			&variable.UpdatedBy,
			&variable.UpdatedAt,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrScan")
			return nil, 0, 0, err
		}
		offset++
		variable.RowNo = offset
		response = append(response, &variable)
	}
	return response, total, pages, nil
}

func (q *variableRepository) FindVariable(
	ctx context.Context,
	request *models.VariableFilter,
) (*models.Variable, error) {
	ctxt := "VariableRepository-FindVariable"
	var sb strings.Builder
	condition, params := q.buildCondition(request, &sb)
	sb.Reset()
	_, _ = fmt.Fprintf(
		&sb,
		`SELECT
			id
			, name
			, value
			, created_by
			, created_at
			, updated_by
			, updated_at
		FROM variables %s`,
		condition,
	)
	var response models.Variable
	err := q.dbRead.QueryRow(ctx, sb.String(), params...).Scan(
		&response.ID,
		&response.Name,
		&response.Value,
		&response.CreatedBy,
		&response.CreatedAt,
		&response.UpdatedBy,
		&response.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, models.ErrVariableNotFound
	}
	if err != nil {
		helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrScan")
		return nil, err
	}
	return &response, nil
}

func (q *variableRepository) SaveVariable(
	ctx context.Context,
	tx pgx.Tx,
	request *models.Variable,
) (*models.Variable, error) {
	ctxt := "VariableRepository-SaveVariable"
	var response models.Variable
	if err := tx.QueryRow(
		ctx,
		`INSERT INTO variables (
			name
			, value
			, created_by
			, created_at
			, updated_by
			, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (LOWER(name)) DO UPDATE SET
			value = $2
			, updated_by = $5
			, updated_at = $6
		RETURNING id
			, name
			, value
			, created_by
			, created_at
			, updated_by
			, updated_at`,
		request.Name,
		request.Value,
		request.CreatedBy,
		request.CreatedAt,
		request.UpdatedBy,
		request.UpdatedAt,
	).Scan(
		&response.ID,
		&response.Name,
		&response.Value,
		&response.CreatedBy,
		&response.CreatedAt,
		&response.UpdatedBy,
		&response.UpdatedAt,
	); err != nil {
		if pgxErr, ok := errors.AsType[*pgconn.PgError](err); ok &&
			pgxErr.Code == pgerrcode.UniqueViolation &&
			pgxErr.ConstraintName == "variables_unique_name_idx" {
			err = models.ErrUniqueVariableNameViolation
		} else {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrScan")
		}
		return nil, err
	}
	return &response, nil
}

func (q *variableRepository) buildCondition(request *models.VariableFilter, sb *strings.Builder) (condition string, params []any) {
	var conds []string
	if len(request.VariableIDs) > 0 {
		sb.Reset()
		_, _ = sb.WriteString("id IN (")
		for i, variableID := range request.VariableIDs {
			params = append(params, variableID)
			if i > 0 {
				_, _ = sb.WriteString(",")
			}
			_, _ = fmt.Fprintf(sb, "$%d", len(params))
		}
		_, _ = sb.WriteString(")")
		conds = append(conds, sb.String())
	}
	if len(request.ExcludedVariableIDs) > 0 {
		sb.Reset()
		_, _ = sb.WriteString("id NOT IN (")
		for i, excludedVariableID := range request.ExcludedVariableIDs {
			params = append(params, excludedVariableID)
			if i > 0 {
				_, _ = sb.WriteString(",")
			}
			_, _ = fmt.Fprintf(sb, "$%d", len(params))
		}
		_, _ = sb.WriteString(")")
		conds = append(conds, sb.String())
	}
	if len(request.Names) > 0 {
		sb.Reset()
		_, _ = sb.WriteString("LOWER(name) IN (")
		for i, name := range request.Names {
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
		sb.Reset()
		_, _ = fmt.Fprintf(sb, "LOWER(name) LIKE $%d", len(params))
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
