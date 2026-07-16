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
	FcmRegistrationRepository interface {
		FindFcmRegistrations(
			ctx context.Context,
			request *models.FcmRegistrationFilter,
		) ([]*models.FcmRegistration, int64, int64, error)
		FindFcmRegistration(
			ctx context.Context,
			request *models.FcmRegistrationFilter,
		) (*models.FcmRegistration, error)
		SaveFcmRegistration(
			ctx context.Context,
			tx pgx.Tx,
			request *models.FcmRegistration,
		) (*models.FcmRegistration, error)
		DeleteFcmRegistration(
			ctx context.Context,
			tx pgx.Tx,
			request *models.FcmRegistrationFilter,
		) (int64, error)
	}

	fcmRegistrationRepository struct {
		dbRead,
		dbWrite *pgxpool.Pool
	}
)

func NewFcmRegistrationRepository(
	dbRead,
	dbWrite *pgxpool.Pool,
) FcmRegistrationRepository {
	return &fcmRegistrationRepository{
		dbRead:  dbRead,
		dbWrite: dbWrite,
	}
}

func (q *fcmRegistrationRepository) FindFcmRegistrations(
	ctx context.Context,
	request *models.FcmRegistrationFilter,
) ([]*models.FcmRegistration, int64, int64, error) {
	ctxt := "FcmRegistrationRepository-FindFcmRegistrations"
	var sb strings.Builder
	condition, params := q.buildCondition(request, &sb)
	sb.Reset()
	_, _ = fmt.Fprintf(&sb, "SELECT COUNT(1) FROM fcm_registrations %s", condition)
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
		, token
		, created_at
		, updated_at
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
	var response []*models.FcmRegistration
	for rows.Next() {
		var fcmRegistration models.FcmRegistration
		if err = rows.Scan(
			&fcmRegistration.ID,
			&fcmRegistration.BuyerID,
			&fcmRegistration.Token,
			&fcmRegistration.CreatedAt,
			&fcmRegistration.UpdatedAt,
			&fcmRegistration.ExpiredAt,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrScan")
			return nil, 0, 0, err
		}
		offset++
		fcmRegistration.RowNo = offset
		response = append(response, &fcmRegistration)
	}
	return response, total, pages, nil
}

func (q *fcmRegistrationRepository) FindFcmRegistration(
	ctx context.Context,
	request *models.FcmRegistrationFilter,
) (*models.FcmRegistration, error) {
	ctxt := "FcmRegistrationRepository-FindFcmRegistration"
	var sb strings.Builder
	condition, params := q.buildCondition(request, &sb)
	sb.Reset()
	_, _ = fmt.Fprintf(
		&sb,
		`SELECT
			id
			, user_id
			, token
			, created_at
			, updated_at
			, expired_at
		FROM fcm_registrations %s`,
		condition,
	)
	query := sb.String()
	var response models.FcmRegistration
	err := q.dbRead.QueryRow(ctx, query, params...).Scan(
		&response.ID,
		&response.BuyerID,
		&response.Token,
		&response.CreatedAt,
		&response.UpdatedAt,
		&response.ExpiredAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, models.ErrFcmRegistrationNotFound
	}
	if err != nil {
		helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrScan")
		return nil, err
	}
	return &response, nil
}

func (q *fcmRegistrationRepository) SaveFcmRegistration(
	ctx context.Context,
	tx pgx.Tx,
	request *models.FcmRegistration,
) (*models.FcmRegistration, error) {
	ctxt := "FcmRegistrationRepository-SaveFcmRegistration"
	var response models.FcmRegistration
	if err := tx.QueryRow(
		ctx,
		`INSERT INTO fcm_registrations (
			user_id
			, token
			, created_at
			, updated_at
			, expired_at
		) VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (token) DO UPDATE SET
			user_id = $1
			, updated_at = $4
			, expired_at = $5
		RETURNING id
			, user_id
			, token
			, created_at
			, updated_at
			, expired_at`,
		request.BuyerID,
		request.Token,
		request.CreatedAt,
		request.UpdatedAt,
		request.ExpiredAt,
	).Scan(
		&response.ID,
		&response.BuyerID,
		&response.Token,
		&response.CreatedAt,
		&response.UpdatedAt,
		&response.ExpiredAt,
	); err != nil {
		helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrScan")
		return nil, err
	}
	return &response, nil
}

func (q *fcmRegistrationRepository) DeleteFcmRegistration(
	ctx context.Context,
	tx pgx.Tx,
	request *models.FcmRegistrationFilter,
) (int64, error) {
	ctxt := "FcmRegistrationRepository-DeleteFcmRegistration"
	var sb strings.Builder
	condition, params := q.buildCondition(request, &sb)
	if len(params) == 0 {
		return 0, nil
	}
	sb.Reset()
	_, _ = fmt.Fprintf(&sb, "DELETE FROM fcm_registrations %s", condition)
	result, err := tx.Exec(ctx, sb.String(), params...)
	if err != nil {
		helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
		return 0, err
	}
	return result.RowsAffected(), nil
}

func (q *fcmRegistrationRepository) buildCondition(request *models.FcmRegistrationFilter, sb *strings.Builder) (condition string, params []any) {
	var conds []string
	if len(request.FcmRegistrationIDs) > 0 {
		sb.Reset()
		_, _ = sb.WriteString("id IN (")
		for i, fcmRegistrationID := range request.FcmRegistrationIDs {
			params = append(params, fcmRegistrationID)
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
		for i, buyerID := range request.UserIDs {
			params = append(params, buyerID)
			if i > 0 {
				_, _ = sb.WriteString(",")
			}
			_, _ = fmt.Fprintf(sb, "$%d", len(params))
		}
		_, _ = sb.WriteString(")")
		conds = append(conds, sb.String())
	}
	if len(request.Tokens) > 0 {
		sb.Reset()
		_, _ = sb.WriteString("token IN (")
		for i, token := range request.Tokens {
			params = append(params, token)
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
