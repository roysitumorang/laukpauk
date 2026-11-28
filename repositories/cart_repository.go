package repositories

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"uuid"

	"github.com/govalues/decimal"
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
		tx *sql.Tx,
		request *models.Cart,
	) (*models.Cart, error)
}

type cartRepository struct {
	dbRead,
	dbWrite *sql.DB
}

func NewCartRepository(
	dbRead,
	dbWrite *sql.DB,
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
	err := q.dbRead.QueryRowContext(ctx, query, helper.SQLArgs(params...)...).Scan(helper.SQLScan(&total)...)
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
		, deleted_user_id
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
		_, _ = fmt.Fprintf(sb, " %s %s", sortExpression(item.Field), item.Direction)
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
	rows, err := q.dbRead.QueryContext(ctx, sb.String(), helper.SQLArgs(params...)...)
	if errors.Is(err, sql.ErrNoRows) {
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
			, product_category_id
			, product_category_name
			, product_id
			, product_name
			, product_photo
			, product_ready_stock
			, stock_unit
			, base_price
			, selling_price
			, discount_type
			, discount_rate
			, discount_minimum_quantity
			, discount_amount
			, sale_price
			, description
			, quantity
			, note
			, created_at
			, updated_at
		FROM cart_items
		WHERE cart_id IN (`,
	)
	mapCartIDs := map[uuid.UUID]int{}
	for rows.Next() {
		var cart models.Cart
		if err = rows.Scan(
			helper.SQLScan(
				&cart.ID,
				&cart.UserID,
				&cart.DeletedUserID,
				&cart.CreatedAt,
				&cart.UpdatedAt,
			)...,
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
		_, _ = sb.WriteString("?")
	}
	if len(params) == 0 {
		return response, total, pages, nil
	}
	_, _ = sb.WriteString(") ORDER BY id")
	rows, err = q.dbRead.QueryContext(ctx, sb.String(), helper.SQLArgs(params...)...)
	if errors.Is(err, sql.ErrNoRows) {
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
			helper.SQLScan(
				&cartItem.ID,
				&cartItem.CartID,
				&cartItem.ProductVariantID,
				&cartItem.ProductCategoryID,
				&cartItem.ProductCategoryName,
				&cartItem.ProductID,
				&cartItem.ProductName,
				&cartItem.ProductPhoto,
				&cartItem.ProductReadyStock,
				&cartItem.StockUnit,
				&cartItem.BasePrice,
				&cartItem.SellingPrice,
				&cartItem.DiscountType,
				&cartItem.DiscountRate,
				&cartItem.DiscountMinimumQuantity,
				&cartItem.DiscountAmount,
				&cartItem.SalePrice,
				&cartItem.Description,
				&cartItem.Quantity,
				&cartItem.Note,
				&cartItem.CreatedAt,
				&cartItem.UpdatedAt,
			)...,
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
			, user_id
			, deleted_user_id
			, created_at
			, updated_at
		FROM carts %s`,
		condition,
	)
	var response models.Cart
	err := q.dbRead.QueryRowContext(ctx, sb.String(), helper.SQLArgs(params...)...).Scan(
		helper.SQLScan(
			&response.ID,
			&response.UserID,
			&response.DeletedUserID,
			&response.CreatedAt,
			&response.UpdatedAt,
		)...,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, models.ErrCartNotFound
	}
	if err != nil {
		helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrScan")
		return nil, err
	}
	rows, err := q.dbRead.QueryContext(
		ctx,
		`SELECT
			id
			, cart_id
			, product_variant_id
			, product_category_id
			, product_category_name
			, product_id
			, product_name
			, product_photo
			, product_ready_stock
			, stock_unit
			, base_price
			, selling_price
			, discount_type
			, discount_rate
			, discount_minimum_quantity
			, discount_amount
			, sale_price
			, description
			, quantity
			, note
			, created_at
			, updated_at
		FROM cart_items
		WHERE cart_id = ?
		ORDER BY id`,
		helper.SQLArgs(response.ID)...,
	)
	if errors.Is(err, sql.ErrNoRows) {
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
			helper.SQLScan(
				&cartItem.ID,
				&cartItem.CartID,
				&cartItem.ProductVariantID,
				&cartItem.ProductCategoryID,
				&cartItem.ProductCategoryName,
				&cartItem.ProductID,
				&cartItem.ProductName,
				&cartItem.ProductPhoto,
				&cartItem.ProductReadyStock,
				&cartItem.StockUnit,
				&cartItem.BasePrice,
				&cartItem.SellingPrice,
				&cartItem.DiscountType,
				&cartItem.DiscountRate,
				&cartItem.DiscountMinimumQuantity,
				&cartItem.DiscountAmount,
				&cartItem.SalePrice,
				&cartItem.Description,
				&cartItem.Quantity,
				&cartItem.Note,
				&cartItem.CreatedAt,
				&cartItem.UpdatedAt,
			)...,
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
	tx *sql.Tx,
	request *models.Cart,
) (*models.Cart, error) {
	ctxt := "CartRepository-SaveCart"
	var (
		response models.Cart
		buyerID  uuid.UUID
	)
	if request.UserID != nil {
		buyerID = *request.UserID
	} else if request.DeletedUserID != nil {
		buyerID = *request.DeletedUserID
	}
	err := tx.QueryRowContext(
		ctx,
		`SELECT
			id
			, user_id
			, deleted_user_id
			, created_at
			, updated_at
		FROM carts
		WHERE id = ? AND user_id = ?
		FOR UPDATE`,
		helper.SQLArgs(
			request.ID,
			buyerID,
		)...,
	).Scan(
		helper.SQLScan(
			&response.ID,
			&response.UserID,
			&response.DeletedUserID,
			&response.CreatedAt,
			&response.UpdatedAt,
		)...,
	)
	if errors.Is(err, sql.ErrNoRows) {
		err = nil
	}
	if err != nil {
		helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrScan")
		return nil, err
	}
	var row *sql.Row
	if response.ID.Compare(uuid.Nil()) == 0 {
		row = tx.QueryRowContext(
			ctx,
			`INSERT INTO carts (
				user_id
				, created_at
				, updated_at
			) VALUES (?, ?, ?)
			ON DUPLICATE KEY UPDATE
				user_id = VALUES(user_id)
			RETURNING id
				, user_id
				, deleted_user_id
				, created_at
				, updated_at`,
			helper.SQLArgs(
				request.UserID,
				request.CreatedAt,
				request.UpdatedAt,
			)...,
		)
	} else {
		row = tx.QueryRowContext(
			ctx,
			`UPDATE carts SET
				user_id = ?
				, deleted_user_id = ?
				, updated_at = ?
			WHERE id = ?
			RETURNING id
				, user_id
				, deleted_user_id
				, created_at
				, updated_at`,
			helper.SQLArgs(
				request.UserID,
				request.DeletedUserID,
				request.UpdatedAt,
				request.ID,
			)...,
		)
	}
	if err = row.Scan(
		helper.SQLScan(
			&response.ID,
			&response.UserID,
			&response.DeletedUserID,
			&response.CreatedAt,
			&response.UpdatedAt,
		)...,
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
			, product_category_id
			, product_category_name
			, product_id
			, product_name
			, product_photo
			, product_ready_stock
			, stock_unit
			, base_price
			, selling_price
			, discount_type
			, discount_rate
			, discount_minimum_quantity
			, discount_amount
			, sale_price
			, description
			, quantity
			, note
			, created_at
			, updated_at
		) VALUES (
			?, ?, ?, ?, ?, ?, ?, ?, ?, ?
			, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?
			, ?
		) ON DUPLICATE KEY UPDATE
			product_category_id = VALUES(product_category_id)
			, product_category_name = VALUES(product_category_name)
			, product_id = VALUES(product_id)
			, product_name = VALUES(product_name)
			, product_photo = VALUES(product_photo)
			, product_ready_stock = VALUES(product_ready_stock)
			, stock_unit = VALUES(stock_unit)
			, base_price = VALUES(base_price)
			, selling_price = VALUES(selling_price)
			, discount_type = VALUES(discount_type)
			, discount_rate = VALUES(discount_rate)
			, discount_minimum_quantity = VALUES(discount_minimum_quantity)
			, discount_amount = VALUES(discount_amount)
			, sale_price = VALUES(sale_price)
			, description = VALUES(description)
			, quantity = VALUES(quantity)
			, note = VALUES(note)
			, updated_at = VALUES(updated_at)
		RETURNING id
			, cart_id
			, product_variant_id
			, product_category_id
			, product_category_name
			, product_id
			, product_name
			, product_photo
			, product_ready_stock
			, stock_unit
			, base_price
			, selling_price
			, discount_type
			, discount_rate
			, discount_minimum_quantity
			, discount_amount
			, sale_price
			, description
			, quantity
			, note
			, created_at
			, updated_at`,
	)
	for _, oldCartItem := range request.Items {
		if oldCartItem.Quantity < 1 {
			continue
		}
		var newCartItem models.CartItem
		if err = tx.QueryRowContext(
			ctx,
			sb.String(),
			helper.SQLArgs(
				response.ID,
				oldCartItem.ProductVariantID,
				oldCartItem.ProductCategoryID,
				oldCartItem.ProductCategoryName,
				oldCartItem.ProductID,
				oldCartItem.ProductName,
				oldCartItem.ProductPhoto,
				oldCartItem.ProductReadyStock,
				oldCartItem.StockUnit,
				oldCartItem.BasePrice,
				oldCartItem.SellingPrice,
				oldCartItem.DiscountType,
				oldCartItem.DiscountRate,
				oldCartItem.DiscountMinimumQuantity,
				oldCartItem.DiscountAmount,
				oldCartItem.SalePrice,
				oldCartItem.Description,
				oldCartItem.Quantity,
				oldCartItem.Note,
				oldCartItem.CreatedAt,
				oldCartItem.UpdatedAt,
			)...,
		).Scan(
			helper.SQLScan(
				&newCartItem.ID,
				&newCartItem.CartID,
				&newCartItem.ProductVariantID,
				&newCartItem.ProductCategoryID,
				&newCartItem.ProductCategoryName,
				&newCartItem.ProductID,
				&newCartItem.ProductName,
				&newCartItem.ProductPhoto,
				&newCartItem.ProductReadyStock,
				&newCartItem.StockUnit,
				&newCartItem.BasePrice,
				&newCartItem.SellingPrice,
				&newCartItem.DiscountType,
				&newCartItem.DiscountRate,
				&newCartItem.DiscountMinimumQuantity,
				&newCartItem.DiscountAmount,
				&newCartItem.SalePrice,
				&newCartItem.Description,
				&newCartItem.Quantity,
				&newCartItem.Note,
				&newCartItem.CreatedAt,
				&newCartItem.UpdatedAt,
			)...,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrScan")
			return nil, err
		}
		response.Items = append(response.Items, &newCartItem)
		response.Quantity += newCartItem.Quantity
		params = append(params, newCartItem.ID)
	}
	params = append([]any{response.ID}, params...)
	sb.Reset()
	_, _ = sb.WriteString("DELETE FROM cart_items WHERE cart_id = ?")
	if n = len(params) - 1; n > 0 {
		_, _ = sb.WriteString(" AND id NOT IN (")
		for i := range n {
			if i > 0 {
				_, _ = sb.WriteString(",")
			}
			_, _ = sb.WriteString("?")
		}
		_, _ = sb.WriteString(")")
	}
	if _, err = tx.ExecContext(ctx, sb.String(), helper.SQLArgs(params...)...); err != nil {
		helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
		return nil, err
	}
	return &response, nil
}

func (q *cartRepository) buildCondition(request *models.CartFilter, sb *strings.Builder) (condition string, params []any) {
	var conds []string
	if len(request.UserIDs) > 0 {
		sb.Reset()
		_, _ = sb.WriteString("user_id IN (")
		for i, buyerID := range request.UserIDs {
			params = append(params, buyerID)
			if i > 0 {
				_, _ = sb.WriteString(",")
			}
			_, _ = sb.WriteString("?")
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
