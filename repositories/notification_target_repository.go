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

type NotificationTargetRepository interface {
	FindNotificationTargets(
		ctx context.Context,
		request *models.NotificationTargetFilter,
	) ([]*models.NotificationTarget, int64, int64, error)
	FindNotificationTarget(
		ctx context.Context,
		request *models.NotificationTargetFilter,
	) (*models.NotificationTarget, error)
	SaveNotificationTarget(
		ctx context.Context,
		tx pgx.Tx,
		request *models.NotificationTarget,
	) (*models.NotificationTarget, error)
	DeleteNotificationTarget(
		ctx context.Context,
		tx pgx.Tx,
		request *models.NotificationTargetFilter,
	) (int64, error)
}

type notificationTargetRepository struct {
	dbRead,
	dbWrite *pgxpool.Pool
}

func NewNotificationTargetRepository(
	dbRead,
	dbWrite *pgxpool.Pool,
) NotificationTargetRepository {
	return &notificationTargetRepository{
		dbRead:  dbRead,
		dbWrite: dbWrite,
	}
}

func (q *notificationTargetRepository) FindNotificationTargets(
	ctx context.Context,
	request *models.NotificationTargetFilter,
) ([]*models.NotificationTarget, int64, int64, error) {
	ctxt := "NotificationTargetRepository-FindNotificationTargets"
	sb := pools.BuilderPool.Get()
	defer func() {
		sb.Reset()
		pools.BuilderPool.Put(sb)
	}()
	sb.Reset()
	condition, params := q.buildCondition(request, sb)
	sb.Reset()
	_, _ = fmt.Fprintf(sb, "SELECT COUNT(1) FROM notification_targets %s", condition)
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
		, admin_id
		, seller_id
		, buyer_id
		, firebase_installation_id
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
	var response []*models.NotificationTarget
	for rows.Next() {
		var notificationTarget models.NotificationTarget
		if err = rows.Scan(
			&notificationTarget.ID,
			&notificationTarget.AdminID,
			&notificationTarget.SellerID,
			&notificationTarget.BuyerID,
			&notificationTarget.FirebaseInstallationID,
			&notificationTarget.CreatedAt,
			&notificationTarget.UpdatedAt,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrScan")
			return nil, 0, 0, err
		}
		offset++
		notificationTarget.RowNo = offset
		response = append(response, &notificationTarget)
	}
	return response, total, pages, nil
}

func (q *notificationTargetRepository) FindNotificationTarget(
	ctx context.Context,
	request *models.NotificationTargetFilter,
) (*models.NotificationTarget, error) {
	ctxt := "NotificationTargetRepository-FindNotificationTarget"
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
			, admin_id
			, seller_id
			, buyer_id
			, firebase_installation_id
			, created_at
			, updated_at
		FROM notification_targets %s`,
		condition,
	)
	query := sb.String()
	var response models.NotificationTarget
	err := q.dbRead.QueryRow(ctx, query, params...).Scan(
		&response.ID,
		&response.AdminID,
		&response.SellerID,
		&response.BuyerID,
		&response.FirebaseInstallationID,
		&response.CreatedAt,
		&response.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, models.ErrNotificationTargetNotFound
	}
	if err != nil {
		helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrScan")
		return nil, err
	}
	return &response, nil
}

func (q *notificationTargetRepository) SaveNotificationTarget(
	ctx context.Context,
	tx pgx.Tx,
	request *models.NotificationTarget,
) (*models.NotificationTarget, error) {
	ctxt := "NotificationTargetRepository-SaveNotificationTarget"
	var response models.NotificationTarget
	err := tx.QueryRow(
		ctx,
		`SELECT
			id
			, admin_id
			, seller_id
			, buyer_id
			, firebase_installation_id
			, created_at
			, updated_at
		FROM notification_targets
		WHERE firebase_installation_id = $1
		FOR UPDATE`,
		request.FirebaseInstallationID,
	).Scan(
		&response.ID,
		&response.AdminID,
		&response.SellerID,
		&response.BuyerID,
		&response.FirebaseInstallationID,
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
	if response.ID.Compare(uuid.Nil()) == 0 {
		if err = tx.QueryRow(
			ctx,
			`INSERT INTO notification_targets (
				admin_id
				, seller_id
				, buyer_id
				, firebase_installation_id
				, created_at
				, updated_at
			) VALUES ($1, $2, $3, $4, $5, $6)
			RETURNING id
				, admin_id
				, seller_id
				, buyer_id
				, firebase_installation_id
				, created_at
				, updated_at`,
			request.AdminID,
			request.SellerID,
			request.BuyerID,
			request.FirebaseInstallationID,
			request.CreatedAt,
			request.UpdatedAt,
		).Scan(
			&response.ID,
			&response.AdminID,
			&response.SellerID,
			&response.BuyerID,
			&response.FirebaseInstallationID,
			&response.CreatedAt,
			&response.UpdatedAt,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrScan")
			return nil, err
		}
	} else {
		var params []any
		var n int
		sb := pools.BuilderPool.Get()
		defer func() {
			sb.Reset()
			pools.BuilderPool.Put(sb)
		}()
		sb.Reset()
		_, _ = sb.WriteString("UPDATE notification_targets SET")
		if response.AdminID != request.AdminID {
			params = append(params, request.AdminID)
			if n = len(params); n > 1 {
				_, _ = sb.WriteString(",")
			}
			_, _ = fmt.Fprintf(sb, " admin_id = $%d", n)
		}
		if response.SellerID != request.SellerID {
			params = append(params, request.SellerID)
			if n = len(params); n > 1 {
				_, _ = sb.WriteString(",")
			}
			_, _ = fmt.Fprintf(sb, " seller_id = $%d", n)
		}
		if response.BuyerID != request.BuyerID {
			params = append(params, request.BuyerID)
			if n = len(params); n > 1 {
				_, _ = sb.WriteString(",")
			}
			_, _ = fmt.Fprintf(sb, " buyer_id = $%d", n)
		}
		if len(params) == 0 {
			return &response, nil
		}
		if !response.UpdatedAt.Equal(request.UpdatedAt) {
			params = append(params, request.UpdatedAt)
			if n = len(params); n > 1 {
				_, _ = sb.WriteString(",")
			}
			_, _ = fmt.Fprintf(sb, " updated_at = $%d", n)
		}
		params = append(params, response.ID)
		_, _ = fmt.Fprintf(
			sb,
			` WHERE id = $%d
			RETURNING id
				, admin_id
				, seller_id
				, buyer_id
				, firebase_installation_id
				, created_at
				, updated_at`,
			len(params),
		)
		if err = tx.QueryRow(ctx, sb.String(), params...).Scan(
			&response.ID,
			&response.AdminID,
			&response.SellerID,
			&response.BuyerID,
			&response.FirebaseInstallationID,
			&response.CreatedAt,
			&response.UpdatedAt,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrScan")
			return nil, err
		}
	}
	return &response, nil
}

func (q *notificationTargetRepository) DeleteNotificationTarget(
	ctx context.Context,
	tx pgx.Tx,
	request *models.NotificationTargetFilter,
) (int64, error) {
	ctxt := "NotificationTargetRepository-DeleteNotificationTarget"
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
	_, _ = fmt.Fprintf(sb, "DELETE FROM notification_targets %s", condition)
	result, err := tx.Exec(ctx, sb.String(), params...)
	if err != nil {
		helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
		return 0, err
	}
	return result.RowsAffected(), nil
}

func (q *notificationTargetRepository) buildCondition(request *models.NotificationTargetFilter, sb *strings.Builder) (condition string, params []any) {
	var conds []string
	if len(request.NotificationTargetIDs) > 0 {
		sb.Reset()
		_, _ = sb.WriteString("id IN (")
		for i, notificationTargetID := range request.NotificationTargetIDs {
			params = append(params, notificationTargetID)
			if i > 0 {
				_, _ = sb.WriteString(",")
			}
			_, _ = fmt.Fprintf(sb, "$%d", len(params))
		}
		_, _ = sb.WriteString(")")
		conds = append(conds, sb.String())
	}
	if len(request.AdminIDs) > 0 {
		sb.Reset()
		_, _ = sb.WriteString("admin_id IN (")
		for i, adminID := range request.AdminIDs {
			params = append(params, adminID)
			if i > 0 {
				_, _ = sb.WriteString(",")
			}
			_, _ = fmt.Fprintf(sb, "$%d", len(params))
		}
		_, _ = sb.WriteString(")")
		conds = append(conds, sb.String())
	}
	if len(request.SellerIDs) > 0 {
		sb.Reset()
		_, _ = sb.WriteString("seller_id IN (")
		for i, sellerID := range request.SellerIDs {
			params = append(params, sellerID)
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
	if len(request.FirebaseInstallationIDs) > 0 {
		sb.Reset()
		_, _ = sb.WriteString("firebase_installation_id IN (")
		for i, firebaseInstallationID := range request.FirebaseInstallationIDs {
			params = append(params, firebaseInstallationID)
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
