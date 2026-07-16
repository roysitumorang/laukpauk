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
	CartItemRepository interface {
		FindCartItems(
			ctx context.Context,
			request *models.CartItemFilter,
		) ([]*models.CartItem, int64, int64, error)
		FindCartItem(
			ctx context.Context,
			request *models.CartItemFilter,
		) (*models.CartItem, error)
		SaveCartItem(
			ctx context.Context,
			tx pgx.Tx,
			request *models.CartItem,
		) (*models.CartItem, error)
		DeleteCartItem(
			ctx context.Context,
			tx pgx.Tx,
			request *models.CartItemFilter,
		) (int64, error)
	}

	cartItemRepository struct {
		dbRead,
		dbWrite *pgxpool.Pool
	}
)

func NewCartItemRepository(
	dbRead,
	dbWrite *pgxpool.Pool,
) CartItemRepository {
	return &cartItemRepository{
		dbRead:  dbRead,
		dbWrite: dbWrite,
	}
}

func (q *cartItemRepository) FindCartItems(
	ctx context.Context,
	request *models.CartItemFilter,
) ([]*models.CartItem, int64, int64, error) {
	ctxt := "CartItemRepository-FindCartItems"
	var sb strings.Builder
	condition, params := q.buildCondition(request, &sb)
	sb.Reset()
	_, _ = fmt.Fprintf(&sb, "SELECT COUNT(1) FROM cart_items %s", condition)
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
		, cart_id
		, product_variant_id
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
	var response []*models.CartItem
	for rows.Next() {
		var cartItem models.CartItem
		if err = rows.Scan(
			&cartItem.ID,
			&cartItem.CartID,
			&cartItem.ProductVariantID,
			&cartItem.Quantity,
			&cartItem.CreatedAt,
			&cartItem.UpdatedAt,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrScan")
			return nil, 0, 0, err
		}
		offset++
		cartItem.RowNo = offset
		response = append(response, &cartItem)
	}
	return response, total, pages, nil
}

func (q *cartItemRepository) FindCartItem(
	ctx context.Context,
	request *models.CartItemFilter,
) (*models.CartItem, error) {
	ctxt := "CartItemRepository-FindCartItem"
	var sb strings.Builder
	condition, params := q.buildCondition(request, &sb)
	sb.Reset()
	_, _ = fmt.Fprintf(
		&sb,
		`SELECT
			id
			, cart_id
			, product_variant_id
			, quantity
			, created_at
			, updated_at
		FROM cart_items %s`,
		condition,
	)
	query := sb.String()
	var response models.CartItem
	err := q.dbRead.QueryRow(ctx, query, params...).Scan(
		&response.ID,
		&response.CartID,
		&response.ProductVariantID,
		&response.Quantity,
		&response.CreatedAt,
		&response.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, models.ErrCartItemNotFound
	}
	if err != nil {
		helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrScan")
		return nil, err
	}
	return &response, nil
}

func (q *cartItemRepository) SaveCartItem(
	ctx context.Context,
	tx pgx.Tx,
	request *models.CartItem,
) (*models.CartItem, error) {
	ctxt := "CartItemRepository-SaveCartItem"
	var response models.CartItem
	if err := tx.QueryRow(
		ctx,
		`INSERT INTO cart_items (
			cart_id
			, product_variant_id
			, quantity
			, created_at
			, updated_at
		) VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (cart_id, product_variant_id) DO UPDATE SET
			quantity = $3
			, updated_at = $5
		RETURNING id
			, cart_id
			, product_variant_id
			, quantity
			, created_at
			, updated_at`,
		request.CartID,
		request.ProductVariantID,
		request.Quantity,
		request.CreatedAt,
		request.UpdatedAt,
	).Scan(
		&response.ID,
		&response.CartID,
		&response.ProductVariantID,
		&response.Quantity,
		&response.CreatedAt,
		&response.UpdatedAt,
	); err != nil {
		helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrScan")
		return nil, err
	}
	return &response, nil
}

func (q *cartItemRepository) DeleteCartItem(
	ctx context.Context,
	tx pgx.Tx,
	request *models.CartItemFilter,
) (int64, error) {
	ctxt := "CartItemRepository-DeleteCartItem"
	var sb strings.Builder
	condition, params := q.buildCondition(request, &sb)
	if len(params) == 0 {
		return 0, nil
	}
	sb.Reset()
	_, _ = fmt.Fprintf(&sb, "DELETE FROM cart_items %s", condition)
	result, err := tx.Exec(ctx, sb.String(), params...)
	if err != nil {
		helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
		return 0, err
	}
	return result.RowsAffected(), nil
}

func (q *cartItemRepository) buildCondition(request *models.CartItemFilter, sb *strings.Builder) (condition string, params []any) {
	var conds []string
	if len(request.CartItemIDs) > 0 {
		sb.Reset()
		_, _ = sb.WriteString("(id IN (")
		j := len(params)
		for i, cartItemID := range request.CartItemIDs {
			params = append(params, cartItemID)
			if i > 0 {
				_, _ = sb.WriteString(",")
			}
			_, _ = fmt.Fprintf(sb, "$%d", len(params))
		}
		_, _ = sb.WriteString(") OR product_variant_id IN (")
		for i := range request.CartItemIDs {
			if i > 0 {
				_, _ = sb.WriteString(",")
			}
			_, _ = fmt.Fprintf(sb, "$%d", j+1)
		}
		_, _ = sb.WriteString("))")
		conds = append(conds, sb.String())
	}
	if len(request.CartIDs) > 0 {
		sb.Reset()
		_, _ = sb.WriteString("cart_id IN (")
		for i, cartID := range request.CartIDs {
			params = append(params, cartID)
			if i > 0 {
				_, _ = sb.WriteString(",")
			}
			_, _ = fmt.Fprintf(sb, "$%d", len(params))
		}
		_, _ = sb.WriteString(")")
		conds = append(conds, sb.String())
	}
	if len(request.ProductVariantIDs) > 0 {
		sb.Reset()
		_, _ = sb.WriteString("product_variant_id IN (")
		for i, productVariantID := range request.ProductVariantIDs {
			params = append(params, productVariantID)
			if i > 0 {
				_, _ = sb.WriteString(",")
			}
			_, _ = fmt.Fprintf(sb, "$%d", len(params))
		}
		_, _ = sb.WriteString(")")
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
