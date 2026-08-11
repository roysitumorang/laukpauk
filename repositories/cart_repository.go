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
	"github.com/roysitumorang/laukpauk/pools"
)

type CartRepository interface {
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
}

type cartRepository struct {
	dbRead,
	dbWrite *pgxpool.Pool
}

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
	sb := pools.BuilderPool.Get()
	defer func() {
		sb.Reset()
		pools.BuilderPool.Put(sb)
	}()
	sb.Reset()
	condition, params := q.buildCondition(request, sb)
	sb.Reset()
	_, _ = fmt.Fprintf(sb, "SELECT COUNT(1) FROM carts %s", condition)
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
	var response []*models.Cart
	var n int
	params = params[:0]
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
		_, _ = fmt.Fprintf(sb, "$%d", n)
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
	var (
		response models.Cart
		buyerID  uuid.UUID
	)
	if request.BuyerID != nil {
		buyerID = *request.BuyerID
	} else if request.DeletedBuyerID != nil {
		buyerID = *request.DeletedBuyerID
	}
	err := tx.QueryRow(
		ctx,
		`SELECT
			id
			, buyer_id
			, deleted_buyer_id
			, created_at
			, updated_at
		FROM carts
		WHERE id = $1 AND buyer_id = $2
		FOR UPDATE`,
		request.ID,
		buyerID,
	).Scan(
		&response.ID,
		&response.BuyerID,
		&response.DeletedBuyerID,
		&response.CreatedAt,
		&response.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		err = nil
	}
	if err != nil {
		helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrScan")
		return nil, err
	}
	var row pgx.Row
	if response.ID.Compare(uuid.Nil()) == 0 {
		row = tx.QueryRow(
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
		)
	} else {
		row = tx.QueryRow(
			ctx,
			`UPDATE carts SET
				buyer_id = $1
				, deleted_buyer_id = $2
				, updated_at = $3
			WHERE id = $4
			RETURNING id
				, buyer_id
				, deleted_buyer_id
				, created_at
				, updated_at`,
			request.BuyerID,
			request.DeletedBuyerID,
			request.UpdatedAt,
			request.ID,
		)
	}
	if err = row.Scan(
		&response.ID,
		&response.BuyerID,
		&response.DeletedBuyerID,
		&response.CreatedAt,
		&response.UpdatedAt,
	); err != nil {
		helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrScan")
		return nil, err
	}
	var params []any
	var n int
	sb := pools.BuilderPool.Get()
	defer func() {
		sb.Reset()
		pools.BuilderPool.Put(sb)
	}()
	sb.Reset()
	_, _ = sb.WriteString(
		`INSERT INTO cart_items (
			cart_id
			, product_variant_id
			, quantity
			, created_at
			, updated_at
		) VALUES`,
	)
	for _, cartItem := range request.Items {
		if cartItem.Quantity > 0 {
			params = append(
				params,
				response.ID,
				cartItem.ProductVariantID,
				cartItem.Quantity,
				cartItem.CreatedAt,
				cartItem.UpdatedAt,
			)
			if n = len(params); n > 5 {
				_, _ = sb.WriteString(",")
			}
			_, _ = fmt.Fprintf(sb, " ($%d, $%d, $%d, $%d, $%d)", n-4, n-3, n-2, n-1, n)
		}
	}
	if len(params) > 0 {
		_, _ = sb.WriteString(
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
		rows, err := tx.Query(ctx, sb.String(), params...)
		if errors.Is(err, pgx.ErrNoRows) {
			err = nil
		}
		if err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrQuery")
			return nil, err
		}
		defer rows.Close()
		params = params[:0]
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
			params = append(params, cartItem.ID)
		}
	}
	params = append([]any{response.ID}, params...)
	sb.Reset()
	_, _ = sb.WriteString("DELETE FROM cart_items WHERE cart_id = $1")
	if n = len(params) - 1; n > 0 {
		_, _ = sb.WriteString(" AND id NOT IN (")
		for i := range n {
			if i > 0 {
				_, _ = sb.WriteString(",")
			}
			_, _ = fmt.Fprintf(sb, "$%d", i+2)
		}
		_, _ = sb.WriteString(")")
	}
	if _, err = tx.Exec(ctx, sb.String(), params...); err != nil {
		helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
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
