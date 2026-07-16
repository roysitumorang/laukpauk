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
)

type (
	IdempotencyKeyRepository interface {
		FindIdempotencyKeys(
			ctx context.Context,
			request *models.IdempotencyKeyFilter,
		) ([]*models.IdempotencyKey, int64, int64, error)
		FindIdempotencyKey(
			ctx context.Context,
			request *models.IdempotencyKeyFilter,
		) (*models.IdempotencyKey, error)
		CreateIdempotencyKey(
			ctx context.Context,
			tx pgx.Tx,
			request *models.IdempotencyKey,
		) (*models.IdempotencyKey, error)
		UpdateIdempotencyKey(
			ctx context.Context,
			tx pgx.Tx,
			request *models.IdempotencyKey,
		) (*models.IdempotencyKey, error)
		DeleteIdempotencyKey(
			ctx context.Context,
			tx pgx.Tx,
			request *models.IdempotencyKeyFilter,
		) (int64, error)
	}

	idempotencyKeyRepository struct {
		dbRead,
		dbWrite *pgxpool.Pool
	}
)

func NewIdempotencyKeyRepository(
	dbRead,
	dbWrite *pgxpool.Pool,
) IdempotencyKeyRepository {
	return &idempotencyKeyRepository{
		dbRead:  dbRead,
		dbWrite: dbWrite,
	}
}

func (q *idempotencyKeyRepository) FindIdempotencyKeys(
	ctx context.Context,
	request *models.IdempotencyKeyFilter,
) ([]*models.IdempotencyKey, int64, int64, error) {
	ctxt := "IdempotencyKeyRepository-FindIdempotencyKeys"
	var sb strings.Builder
	condition, params := q.buildCondition(request, &sb)
	sb.Reset()
	_, _ = fmt.Fprintf(&sb, "SELECT COUNT(1) FROM idempotency_keys %s", condition)
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
		, buyer_id
		, occupied
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
	var response []*models.IdempotencyKey
	for rows.Next() {
		var idempotencyKey models.IdempotencyKey
		if err = rows.Scan(
			&idempotencyKey.ID,
			&idempotencyKey.BuyerID,
			&idempotencyKey.Occupied,
			&idempotencyKey.CreatedAt,
			&idempotencyKey.ExpiredAt,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrScan")
			return nil, 0, 0, err
		}
		offset++
		idempotencyKey.RowNo = offset
		response = append(response, &idempotencyKey)
	}
	return response, total, pages, nil
}

func (q *idempotencyKeyRepository) FindIdempotencyKey(
	ctx context.Context,
	request *models.IdempotencyKeyFilter,
) (*models.IdempotencyKey, error) {
	ctxt := "IdempotencyKeyRepository-FindIdempotencyKey"
	var sb strings.Builder
	condition, params := q.buildCondition(request, &sb)
	sb.Reset()
	_, _ = fmt.Fprintf(
		&sb,
		`SELECT
			id
			, buyer_id
			, occupied
			, created_at
			, expired_at
		FROM idempotency_keys %s`,
		condition,
	)
	query := sb.String()
	var response models.IdempotencyKey
	err := q.dbRead.QueryRow(ctx, query, params...).Scan(
		&response.ID,
		&response.BuyerID,
		&response.Occupied,
		&response.CreatedAt,
		&response.ExpiredAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, models.ErrIdempotencyKeyNotFound
	}
	if err != nil {
		helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrScan")
		return nil, err
	}
	return &response, nil
}

func (q *idempotencyKeyRepository) CreateIdempotencyKey(
	ctx context.Context,
	tx pgx.Tx,
	request *models.IdempotencyKey,
) (*models.IdempotencyKey, error) {
	ctxt := "IdempotencyKeyRepository-CreateIdempotencyKey"
	var response models.IdempotencyKey
	if err := tx.QueryRow(
		ctx,
		`INSERT INTO idempotency_keys (
			buyer_id
			, occupied
			, created_at
			, expired_at
		) VALUES ($1, $2, $3, $4)
		RETURNING id
			, buyer_id
			, occupied
			, created_at
			, expired_at`,
		request.BuyerID,
		request.Occupied,
		request.CreatedAt,
		request.ExpiredAt,
	).Scan(
		&response.ID,
		&response.BuyerID,
		&response.Occupied,
		&response.CreatedAt,
		&response.ExpiredAt,
	); err != nil {
		helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrScan")
		return nil, err
	}
	return &response, nil
}

func (q *idempotencyKeyRepository) UpdateIdempotencyKey(
	ctx context.Context,
	tx pgx.Tx,
	request *models.IdempotencyKey,
) (*models.IdempotencyKey, error) {
	ctxt := "IdempotencyKeyRepository-UpdateIdempotencyKey"
	var response models.IdempotencyKey
	if err := tx.QueryRow(
		ctx,
		`UPDATE idempotency_keys SET
			occupied = $1
		WHERE id = $2
		RETURNING id
			, buyer_id
			, occupied
			, created_at
			, expired_at`,
		request.Occupied,
		request.ID,
	).Scan(
		&response.ID,
		&response.BuyerID,
		&response.Occupied,
		&response.CreatedAt,
		&response.ExpiredAt,
	); err != nil {
		helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrScan")
		return nil, err
	}
	return &response, nil
}

func (q *idempotencyKeyRepository) DeleteIdempotencyKey(
	ctx context.Context,
	tx pgx.Tx,
	request *models.IdempotencyKeyFilter,
) (int64, error) {
	ctxt := "IdempotencyKeyRepository-DeleteIdempotencyKey"
	var sb strings.Builder
	condition, params := q.buildCondition(request, &sb)
	if len(params) == 0 {
		return 0, nil
	}
	sb.Reset()
	_, _ = fmt.Fprintf(&sb, "DELETE FROM idempotency_keys %s", condition)
	result, err := tx.Exec(ctx, sb.String(), params...)
	if err != nil {
		helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
		return 0, err
	}
	return result.RowsAffected(), nil
}

func (q *idempotencyKeyRepository) buildCondition(request *models.IdempotencyKeyFilter, sb *strings.Builder) (condition string, params []any) {
	var conds []string
	if len(request.IdempotencyKeyIDs) > 0 {
		sb.Reset()
		_, _ = sb.WriteString("id IN (")
		for i, idempotencyKeyID := range request.IdempotencyKeyIDs {
			params = append(params, idempotencyKeyID)
			if i > 0 {
				_, _ = sb.WriteString(",")
			}
			_, _ = fmt.Fprintf(sb, "$%d", len(params))
		}
		_, _ = sb.WriteString(")")
		conds = append(conds, sb.String())
	}
	if len(request.BuyerIDs) > 0 {
		sb.Reset()
		_, _ = sb.WriteString("buyer_id IN (")
		for i, buyerID := range request.BuyerIDs {
			params = append(params, buyerID)
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
		_, _ = fmt.Fprintf(sb, " $%d", len(params))
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
