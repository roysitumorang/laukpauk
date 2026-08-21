package repositories

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/govalues/decimal"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/roysitumorang/laukpauk/helper"
	"github.com/roysitumorang/laukpauk/models"
	"github.com/roysitumorang/laukpauk/pools"
)

type JsonWebTokenRepository interface {
	FindJsonWebTokens(
		ctx context.Context,
		request *models.JsonWebTokenFilter,
	) ([]*models.JsonWebToken, int64, int64, error)
	FindJsonWebToken(
		ctx context.Context,
		request *models.JsonWebTokenFilter,
	) (*models.JsonWebToken, error)
	CreateJsonWebToken(
		ctx context.Context,
		tx pgx.Tx,
		request *models.JsonWebToken,
	) (*models.JsonWebToken, error)
	DeleteJsonWebToken(
		ctx context.Context,
		tx pgx.Tx,
		request *models.JsonWebTokenFilter,
	) (int64, error)
}

type jsonWebTokenRepository struct {
	dbRead,
	dbWrite *pgxpool.Pool
}

func NewJsonWebTokenRepository(
	dbRead,
	dbWrite *pgxpool.Pool,
) JsonWebTokenRepository {
	return &jsonWebTokenRepository{
		dbRead:  dbRead,
		dbWrite: dbWrite,
	}
}

func (q *jsonWebTokenRepository) FindJsonWebTokens(
	ctx context.Context,
	request *models.JsonWebTokenFilter,
) ([]*models.JsonWebToken, int64, int64, error) {
	ctxt := "JsonWebTokenRepository-FindJsonWebTokens"
	sb := pools.BuilderPool.Get()
	defer func() {
		sb.Reset()
		pools.BuilderPool.Put(sb)
	}()
	sb.Reset()
	condition, params := q.buildCondition(request, sb)
	sb.Reset()
	_, _ = fmt.Fprintf(sb, "SELECT COUNT(1) FROM json_web_tokens %s", condition)
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
		, user_id
		, created_at
		, expired_at`,
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
		_, _ = fmt.Fprintf(sb, " %s %s", item.Field, item.Direction)
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
	rows, err := q.dbRead.Query(ctx, sb.String(), params...)
	if errors.Is(err, pgx.ErrNoRows) {
		err = nil
	}
	if err != nil {
		helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrQuery")
		return nil, 0, 0, err
	}
	defer rows.Close()
	var response []*models.JsonWebToken
	for rows.Next() {
		var jsonWebToken models.JsonWebToken
		if err = rows.Scan(
			&jsonWebToken.ID,
			&jsonWebToken.UserID,
			&jsonWebToken.CreatedAt,
			&jsonWebToken.ExpiredAt,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrScan")
			return nil, 0, 0, err
		}
		offset++
		jsonWebToken.RowNo = offset
		response = append(response, &jsonWebToken)
	}
	return response, total, pages, nil
}

func (q *jsonWebTokenRepository) FindJsonWebToken(
	ctx context.Context,
	request *models.JsonWebTokenFilter,
) (*models.JsonWebToken, error) {
	ctxt := "JsonWebTokenRepository-FindJsonWebToken"
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
			, user_id
			, created_at
			, expired_at
		FROM json_web_tokens %s`,
		condition,
	)
	query := sb.String()
	var response models.JsonWebToken
	err := q.dbRead.QueryRow(ctx, query, params...).Scan(
		&response.ID,
		&response.UserID,
		&response.CreatedAt,
		&response.ExpiredAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, models.ErrJsonWebTokenNotFound
	}
	if err != nil {
		helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrScan")
		return nil, err
	}
	return &response, nil
}

func (q *jsonWebTokenRepository) CreateJsonWebToken(
	ctx context.Context,
	tx pgx.Tx,
	request *models.JsonWebToken,
) (*models.JsonWebToken, error) {
	ctxt := "JsonWebTokenRepository-CreateJsonWebToken"
	var response models.JsonWebToken
	if err := tx.QueryRow(
		ctx,
		`INSERT INTO json_web_tokens (
			user_id
			, created_at
			, expired_at
		) VALUES ($1, $2, $3)
		RETURNING id
			, user_id
			, created_at
			, expired_at`,
		request.UserID,
		request.CreatedAt,
		request.ExpiredAt,
	).Scan(
		&response.ID,
		&response.UserID,
		&response.CreatedAt,
		&response.ExpiredAt,
	); err != nil {
		helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrScan")
		return nil, err
	}
	return &response, nil
}

func (q *jsonWebTokenRepository) DeleteJsonWebToken(
	ctx context.Context,
	tx pgx.Tx,
	request *models.JsonWebTokenFilter,
) (int64, error) {
	ctxt := "JsonWebTokenRepository-DeleteJsonWebToken"
	sb := pools.BuilderPool.Get()
	defer func() {
		sb.Reset()
		pools.BuilderPool.Put(sb)
	}()
	sb.Reset()
	condition, params := q.buildCondition(request, sb)
	if len(params) == 0 {
		return 0, nil
	}
	sb.Reset()
	_, _ = fmt.Fprintf(sb, "DELETE FROM json_web_tokens %s", condition)
	result, err := tx.Exec(ctx, sb.String(), params...)
	if err != nil {
		helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
		return 0, err
	}
	return result.RowsAffected(), nil
}

func (q *jsonWebTokenRepository) buildCondition(request *models.JsonWebTokenFilter, sb *strings.Builder) (condition string, params []any) {
	var conds []string
	if len(request.JsonWebTokenIDs) > 0 {
		sb.Reset()
		_, _ = sb.WriteString("id IN (")
		for i, jsonWebTokenID := range request.JsonWebTokenIDs {
			params = append(params, jsonWebTokenID)
			if i > 0 {
				_, _ = sb.WriteString(",")
			}
			_, _ = fmt.Fprintf(sb, "$%d", len(params))
		}
		_, _ = sb.WriteString(")")
		conds = append(conds, sb.String())
	}
	if len(request.UserIDs) > 0 {
		sb.Reset()
		_, _ = sb.WriteString("user_id IN (")
		for i, userID := range request.UserIDs {
			params = append(params, userID)
			if i > 0 {
				_, _ = sb.WriteString(",")
			}
			_, _ = fmt.Fprintf(sb, "$%d", len(params))
		}
		_, _ = sb.WriteString(")")
		conds = append(conds, sb.String())
	}
	if request.Expired != nil {
		params = append(params, time.Now())
		sb.Reset()
		_, _ = sb.WriteString("expired_at ")
		if *request.Expired {
			_, _ = sb.WriteString("<=")
		} else {
			_, _ = sb.WriteString(">")
		}
		_, _ = fmt.Fprintf(sb, "$%d", len(params))
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
