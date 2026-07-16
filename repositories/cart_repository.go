package repositories

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"

	"github.com/govalues/decimal"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/roysitumorang/laukpauk/helper"
	"github.com/roysitumorang/laukpauk/models"
)

type (
	CartRepository interface {
		FindCarts(
			ctx context.Context,
			request *models.CartFilter,
		) ([]*models.Cart, int64, int64, error)
		FindCart(
			ctx context.Context,
			request *models.CartFilter,
		) (*models.Cart, error)
		SaveCart(
			ctx context.Context,
			tx pgx.Tx,
			request *models.Cart,
		) (*models.Cart, error)
		DeleteCart(
			ctx context.Context,
			tx pgx.Tx,
			request *models.CartFilter,
		) (int64, error)
	}

	cartRepository struct {
		dbRead,
		dbWrite *pgxpool.Pool
	}
)

func NewCartRepository(
	dbRead,
	dbWrite *pgxpool.Pool,
) CartRepository {
	return &cartRepository{
		dbRead:  dbRead,
		dbWrite: dbWrite,
	}
}

func (q *cartRepository) FindCarts(
	ctx context.Context,
	request *models.CartFilter,
) ([]*models.Cart, int64, int64, error) {
	ctxt := "CartRepository-FindCarts"
	var sb strings.Builder
	condition, params := q.buildCondition(request, &sb)
	sb.Reset()
	_, _ = fmt.Fprintf(&sb, "SELECT COUNT(1) FROM carts %s", condition)
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
		, quantity
		, created_at
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
	var response []*models.Cart
	for rows.Next() {
		var cart models.Cart
		if err = rows.Scan(
			&cart.ID,
			&cart.BuyerID,
			&cart.Quantity,
			&cart.CreatedAt,
			&cart.UpdatedAt,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrScan")
			return nil, 0, 0, err
		}
		offset++
		cart.RowNo = offset
		response = append(response, &cart)
	}
	return response, total, pages, nil
}

func (q *cartRepository) FindCart(
	ctx context.Context,
	request *models.CartFilter,
) (*models.Cart, error) {
	ctxt := "CartRepository-FindCart"
	var sb strings.Builder
	condition, params := q.buildCondition(request, &sb)
	sb.Reset()
	_, _ = fmt.Fprintf(
		&sb,
		`SELECT
			id
			, buyer_id
			, quantity
			, created_at
			, updated_at
		FROM carts %s`,
		condition,
	)
	query := sb.String()
	var response models.Cart
	err := q.dbRead.QueryRow(ctx, query, params...).Scan(
		&response.ID,
		&response.BuyerID,
		&response.Quantity,
		&response.CreatedAt,
		&response.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, models.ErrCartNotFound
	}
	if err != nil {
		helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrScan")
		return nil, err
	}
	return &response, nil
}

func (q *cartRepository) SaveCart(
	ctx context.Context,
	tx pgx.Tx,
	request *models.Cart,
) (*models.Cart, error) {
	ctxt := "CartRepository-SaveCart"
	var response models.Cart
	if err := tx.QueryRow(
		ctx,
		`INSERT INTO carts (
			buyer_id
			, quantity
			, created_at
			, updated_at
		) VALUES ($1, $2, $3, $4)
		ON CONFLICT (buyer_id) DO UPDATE SET
			quantity = $2
			, updated_at = $4
		RETURNING id
			, buyer_id
			, quantity
			, created_at
			, updated_at`,
		request.BuyerID,
		request.Quantity,
		request.CreatedAt,
		request.UpdatedAt,
	).Scan(
		&response.ID,
		&response.BuyerID,
		&response.Quantity,
		&response.CreatedAt,
		&response.UpdatedAt,
	); err != nil {
		helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrScan")
		return nil, err
	}
	return &response, nil
}

func (q *cartRepository) DeleteCart(
	ctx context.Context,
	tx pgx.Tx,
	request *models.CartFilter,
) (int64, error) {
	ctxt := "CartRepository-DeleteCart"
	var sb strings.Builder
	condition, params := q.buildCondition(request, &sb)
	if len(params) == 0 {
		return 0, nil
	}
	sb.Reset()
	_, _ = fmt.Fprintf(&sb, "DELETE FROM carts %s", condition)
	result, err := tx.Exec(ctx, sb.String(), params...)
	if err != nil {
		helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
		return 0, err
	}
	return result.RowsAffected(), nil
}

func (q *cartRepository) buildCondition(request *models.CartFilter, sb *strings.Builder) (condition string, params []any) {
	var conds []string
	if len(request.CartIDs) > 0 {
		sb.Reset()
		_, _ = sb.WriteString("(id IN (")
		j := len(params)
		for i, cartID := range request.CartIDs {
			params = append(params, cartID)
			if i > 0 {
				_, _ = sb.WriteString(",")
			}
			_, _ = fmt.Fprintf(sb, "$%d", len(params))
		}
		_, _ = sb.WriteString(") OR buyer_id IN (")
		for i := range request.CartIDs {
			if i > 0 {
				_, _ = sb.WriteString(",")
			}
			_, _ = fmt.Fprintf(sb, "$%d", j+1)
		}
		_, _ = sb.WriteString("))")
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
