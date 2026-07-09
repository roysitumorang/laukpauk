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
	PointRepository interface {
		FindPoints(
			ctx context.Context,
			request *models.PointFilter,
		) ([]*models.Point, int64, int64, error)
		FindPoint(
			ctx context.Context,
			request *models.PointFilter,
		) (*models.Point, error)
		CreatePoint(
			ctx context.Context,
			tx pgx.Tx,
			request *models.Point,
		) (*models.Point, error)
		UpdatePoint(
			ctx context.Context,
			tx pgx.Tx,
			request *models.Point,
		) (*models.Point, error)
	}

	pointRepository struct {
		dbRead,
		dbWrite *pgxpool.Pool
	}
)

func NewPointRepository(
	dbRead,
	dbWrite *pgxpool.Pool,
) PointRepository {
	return new(pointRepository{
		dbRead:  dbRead,
		dbWrite: dbWrite,
	})
}

func (q *pointRepository) FindPoints(
	ctx context.Context,
	request *models.PointFilter,
) ([]*models.Point, int64, int64, error) {
	ctxt := "PointRepository-FindPoints"
	var sb strings.Builder
	condition, params := q.buildCondition(request, &sb)
	sb.Reset()
	_, _ = fmt.Fprintf(
		&sb,
		`SELECT COUNT(1)
		FROM points p
		JOIN users u ON p.buyer_id = u.id
		JOIN buyers b ON u.id = b.user_id %s`,
		condition,
	)
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
		`p.id
		, p.point_type
		, p.buyer_id
		, u.name
		, u.username
		, u.mobile_phone
		, b.referral_code
		, p.amount
		, p.transaction_id
		, p.spendable
		, p.description
		, p.created_at
		, p.updated_at`,
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
		_, _ = fmt.Fprintf(&sb, " p.%s %s", item.Field, item.Direction)
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
	var response []*models.Point
	for rows.Next() {
		var point models.Point
		if err = rows.Scan(
			&point.ID,
			&point.PointType,
			&point.BuyerID,
			&point.BuyerName,
			&point.BuyerUsername,
			&point.BuyerMobilePhone,
			&point.BuyerReferralCode,
			&point.Amount,
			&point.TransactionID,
			&point.Spendable,
			&point.Description,
			&point.CreatedAt,
			&point.UpdatedAt,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrScan")
			return nil, 0, 0, err
		}
		offset++
		point.RowNo = offset
		response = append(response, &point)
	}
	return response, total, pages, nil
}

func (q *pointRepository) FindPoint(
	ctx context.Context,
	request *models.PointFilter,
) (*models.Point, error) {
	ctxt := "PointRepository-FindPoint"
	var sb strings.Builder
	condition, params := q.buildCondition(request, &sb)
	sb.Reset()
	_, _ = fmt.Fprintf(
		&sb,
		`SELECT
			p.id
			, p.point_type
			, p.buyer_id
			, u.name
			, u.username
			, u.mobile_phone
			, b.referral_code
			, p.amount
			, p.transaction_id
			, p.spendable
			, p.description
			, p.created_at
			, p.updated_at
		FROM points p
		JOIN users u ON p.buyer_id = u.id
		JOIN buyers b ON u.id = b.user_id %s`,
		condition,
	)
	query := sb.String()
	var response models.Point
	err := q.dbRead.QueryRow(ctx, query, params...).Scan(
		&response.ID,
		&response.PointType,
		&response.BuyerID,
		&response.BuyerName,
		&response.BuyerUsername,
		&response.BuyerMobilePhone,
		&response.BuyerReferralCode,
		&response.Amount,
		&response.TransactionID,
		&response.Spendable,
		&response.Description,
		&response.CreatedAt,
		&response.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, models.ErrPointNotFound
	}
	if err != nil {
		helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrScan")
		return nil, err
	}
	return &response, nil
}

func (q *pointRepository) CreatePoint(
	ctx context.Context,
	tx pgx.Tx,
	request *models.Point,
) (*models.Point, error) {
	ctxt := "PointRepository-CreatePoint"
	var response models.Point
	err := tx.QueryRow(
		ctx,
		`WITH p AS (
			INSERT INTO points (
				point_type
				, buyer_id
				, amount
				, transaction_id
				, spendable
				, description
				, created_at
				, updated_at
			) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
			RETURNING *
		)
		SELECT
			p.id
			, p.point_type
			, p.buyer_id
			, u.name
			, u.username
			, u.mobile_phone
			, b.referral_code
			, p.amount
			, p.transaction_id
			, p.spendable
			, p.description
			, p.created_at
			, p.updated_at
		FROM p
		JOIN users u ON p.buyer_id = u.id
		JOIN buyers b ON u.id = b.user_id`,
		request.PointType,
		request.BuyerID,
		request.Amount,
		request.TransactionID,
		request.Spendable,
		request.Description,
		request.CreatedAt,
		request.UpdatedAt,
	).Scan(
		&response.ID,
		&response.PointType,
		&response.BuyerID,
		&response.BuyerName,
		&response.BuyerUsername,
		&response.BuyerMobilePhone,
		&response.BuyerReferralCode,
		&response.Amount,
		&response.TransactionID,
		&response.Spendable,
		&response.Description,
		&response.CreatedAt,
		&response.UpdatedAt,
	)
	if err != nil {
		helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrScan")
		return nil, err
	}
	return &response, nil
}

func (q *pointRepository) UpdatePoint(
	ctx context.Context,
	tx pgx.Tx,
	request *models.Point,
) (*models.Point, error) {
	ctxt := "PointRepository-UpdatePoint"
	var response models.Point
	err := tx.QueryRow(
		ctx,
		`WITH p AS (
			UPDATE points SET
				spendable = $1
				, updated_at = $2
			WHERE id = $3
			RETURNING *
		)
		SELECT
			p.id
			, p.point_type
			, p.buyer_id
			, u.name
			, u.username
			, u.mobile_phone
			, b.referral_code
			, p.amount
			, p.transaction_id
			, p.spendable
			, p.description
			, p.created_at
			, p.updated_at
		FROM p
		JOIN users u ON p.buyer_id = u.id
		JOIN buyers b ON u.id = b.user_id`,
		request.Spendable,
		request.UpdatedAt,
		request.ID,
	).Scan(
		&response.ID,
		&response.PointType,
		&response.BuyerID,
		&response.BuyerName,
		&response.BuyerUsername,
		&response.BuyerMobilePhone,
		&response.BuyerReferralCode,
		&response.Amount,
		&response.TransactionID,
		&response.Spendable,
		&response.Description,
		&response.CreatedAt,
		&response.UpdatedAt,
	)
	if err != nil {
		helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrScan")
		return nil, err
	}
	return &response, nil
}

func (q *pointRepository) buildCondition(request *models.PointFilter, sb *strings.Builder) (condition string, params []any) {
	var conds []string
	if len(request.PointIDs) > 0 {
		sb.Reset()
		_, _ = sb.WriteString("p.id IN (")
		for i, pointID := range request.PointIDs {
			params = append(params, pointID)
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
		_, _ = sb.WriteString("p.buyer_id IN (")
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
	if len(request.TransactionIDs) > 0 {
		sb.Reset()
		_, _ = sb.WriteString("p.transaction_id IN (")
		for i, transactionID := range request.TransactionIDs {
			params = append(params, transactionID)
			if i > 0 {
				_, _ = sb.WriteString(",")
			}
			_, _ = fmt.Fprintf(sb, "$%d", len(params))
		}
		_, _ = sb.WriteString(")")
		conds = append(conds, sb.String())
	}
	if len(request.PointTypes) > 0 {
		sb.Reset()
		_, _ = sb.WriteString("p.point_type IN (")
		for i, pointType := range request.PointTypes {
			params = append(params, pointType)
			if i > 0 {
				_, _ = sb.WriteString(",")
			}
			_, _ = fmt.Fprintf(sb, "$%d", len(params))
		}
		_, _ = sb.WriteString(")")
		conds = append(conds, sb.String())
	}
	if request.Spendable != nil {
		if *request.Spendable {
			conds = append(conds, "p.spendable")
		} else {
			conds = append(conds, "NOT p.spendable")
		}
	}
	if request.Keyword != "" {
		sb.Reset()
		_, _ = fmt.Fprintf(sb, "%%%s%%", strings.ToLower(request.Keyword))
		params = append(params, sb.String())
		n := len(params)
		sb.Reset()
		_, _ = fmt.Fprintf(
			sb,
			`(LOWER(u.name) LIKE $%d
				OR u.username LIKE $%d
				OR u.mobile_phone LIKE $%d
				OR LOWER(p.description) LIKE $%d)`,
			n, n, n, n,
		)
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
