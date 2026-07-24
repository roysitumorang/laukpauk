package repositories

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"uuid"

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
		UpdateCart(
			ctx context.Context,
			tx pgx.Tx,
			request *models.Cart,
		) (*models.Cart, error)
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
		, deleted_buyer_id
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
	var (
		response []*models.Cart
		n        int
	)
	params = []any{}
	sb.Reset()
	_, _ = sb.WriteString(
		`SELECT
			id
			, cart_id
			, product_variant_id
			, quantity
			, created_at
			, updated_at
		FROM cart_items
		WHERE cart_id IN (`,
	)
	mapCartIDs := map[uuid.UUID]int{}
	for rows.Next() {
		var cart models.Cart
		if err = rows.Scan(
			&cart.ID,
			&cart.BuyerID,
			&cart.DeletedBuyerID,
			&cart.CreatedAt,
			&cart.UpdatedAt,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrScan")
			return nil, 0, 0, err
		}
		offset++
		cart.RowNo = offset
		response = append(response, &cart)
		mapCartIDs[cart.ID] = len(response) - 1
		params = append(params, cart.ID)
		if n = len(params); n > 1 {
			_, _ = sb.WriteString(",")
		}
		_, _ = fmt.Fprintf(&sb, "$%d", n)
	}
	if len(params) == 0 {
		return response, total, pages, nil
	}
	_, _ = sb.WriteString(") ORDER BY id")
	rows, err = q.dbRead.Query(ctx, sb.String(), params...)
	if errors.Is(err, pgx.ErrNoRows) {
		err = nil
	}
	if err != nil {
		helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrQuery")
		return nil, 0, 0, err
	}
	defer rows.Close()
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
		i := mapCartIDs[cartItem.CartID]
		cart := response[i]
		cart.Quantity += cartItem.Quantity
		cart.Items = append(cart.Items, &cartItem)
		response[i] = cart
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
			, deleted_buyer_id
			, created_at
			, updated_at
		FROM carts %s`,
		condition,
	)
	var response models.Cart
	err := q.dbRead.QueryRow(ctx, sb.String(), params...).Scan(
		&response.ID,
		&response.BuyerID,
		&response.DeletedBuyerID,
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
	rows, err := q.dbRead.Query(
		ctx,
		`SELECT
			id
			, cart_id
			, product_variant_id
			, quantity
			, created_at
			, updated_at
		FROM cart_items
		WHERE cart_id = $1
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
			return nil, err
		}
		response.Items = append(response.Items, &cartItem)
		response.Quantity += cartItem.Quantity
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
	err := tx.QueryRow(
		ctx,
		`INSERT INTO carts (
			buyer_id
			, created_at
			, updated_at
		) VALUES ($1, $2, $3)
		ON CONFLICT (buyer_id) DO UPDATE SET
			buyer_id = EXCLUDED.buyer_id
		RETURNING id
			, buyer_id
			, deleted_buyer_id
			, created_at
			, updated_at`,
		request.BuyerID,
		request.CreatedAt,
		request.UpdatedAt,
	).Scan(
		&response.ID,
		&response.BuyerID,
		&response.DeletedBuyerID,
		&response.CreatedAt,
		&response.UpdatedAt,
	)
	if err != nil {
		helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrScan")
		return nil, err
	}
	var (
		sbInsert,
		sbDelete strings.Builder
		n int
	)
	var paramsInsert []any
	_, _ = sbInsert.WriteString(
		`INSERT INTO cart_items (
			cart_id
			, product_variant_id
			, quantity
			, created_at
			, updated_at
		) VALUES`,
	)
	paramsDelete := []any{response.ID}
	_, _ = sbDelete.WriteString("DELETE FROM cart_items WHERE cart_id = $1 AND product_variant_id NOT IN (")
	for _, cartItem := range request.Items {
		if cartItem.Quantity > 0 {
			paramsInsert = append(
				paramsInsert,
				response.ID,
				cartItem.ProductVariantID,
				cartItem.Quantity,
				cartItem.CreatedAt,
				cartItem.UpdatedAt,
			)
			if n = len(paramsInsert); n > 5 {
				_, _ = sbInsert.WriteString(",")
			}
			_, _ = fmt.Fprintf(&sbInsert, " ($%d, $%d, $%d, $%d, $%d)", n-4, n-3, n-2, n-1, n)
		} else {
			paramsDelete = append(paramsDelete, cartItem.ProductVariantID)
			if n = len(paramsDelete); n > 2 {
				_, _ = sbDelete.WriteString(",")
			}
			_, _ = fmt.Fprintf(&sbDelete, "$%d", n)
		}
	}
	if len(paramsDelete) > 1 {
		_, _ = sbDelete.WriteString(")")
		if _, err = tx.Exec(ctx, sbDelete.String(), paramsDelete...); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrScan")
			return nil, err
		}
	}
	if len(paramsInsert) > 0 {
		_, _ = sbInsert.WriteString(
			` ON CONFLICT (cart_id, product_variant_id) DO UPDATE SET
				quantity = EXCLUDED.quantity
				, updated_at = EXCLUDED.updated_at
			RETURNING id
				, cart_id
				, product_variant_id
				, quantity
				, created_at
				, updated_at`,
		)
		rows, err := tx.Query(ctx, sbInsert.String(), paramsInsert...)
		if errors.Is(err, pgx.ErrNoRows) {
			err = nil
		}
		if err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrQuery")
			return nil, err
		}
		defer rows.Close()
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
				return nil, err
			}
			response.Items = append(response.Items, &cartItem)
			response.Quantity += cartItem.Quantity
		}
	}
	return &response, nil
}

func (q *cartRepository) UpdateCart(
	ctx context.Context,
	tx pgx.Tx,
	request *models.Cart,
) (*models.Cart, error) {
	ctxt := "CartRepository-UpdateCart"
	var response models.Cart
	err := tx.QueryRow(
		ctx,
		`SELECT
			id
			, buyer_id
			, deleted_buyer_id
			, created_at
			, updated_at
		FROM carts
		WHERE id = $1
		FOR UPDATE`,
		request.ID,
	).Scan(
		&response.ID,
		&response.BuyerID,
		&response.DeletedBuyerID,
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
	var (
		params []any
		sb     strings.Builder
		n      int
	)
	_, _ = sb.WriteString("UPDATE carts SET")
	if response.BuyerID != nil && request.BuyerID == nil {
		params = append(params, request.BuyerID)
		if n = len(params); n > 1 {
			_, _ = sb.WriteString(",")
		}
		_, _ = fmt.Fprintf(&sb, " buyer_id = $%d", n)
	}
	if response.DeletedBuyerID == nil && request.DeletedBuyerID != nil {
		params = append(params, request.DeletedBuyerID)
		if n = len(params); n > 1 {
			_, _ = sb.WriteString(",")
		}
		_, _ = fmt.Fprintf(&sb, " deleted_buyer_id = $%d", n)
	}
	if len(params) == 0 {
		return &response, nil
	}
	params = append(params, request.UpdatedAt, request.ID)
	n = len(params)
	_, _ = fmt.Fprintf(
		&sb,
		`, updated_at = $%d
		WHERE id = $%d
		RETURNING id
			, buyer_id
			, deleted_buyer_id
			, created_at
			, updated_at`,
		n-1, n,
	)
	if err = tx.QueryRow(ctx, sb.String(), params...).Scan(
		&response.ID,
		&response.BuyerID,
		&response.DeletedBuyerID,
		&response.CreatedAt,
		&response.UpdatedAt,
	); err != nil {
		helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrScan")
		return nil, err
	}
	return &response, nil
}

func (q *cartRepository) buildCondition(request *models.CartFilter, sb *strings.Builder) (condition string, params []any) {
	var conds []string
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
