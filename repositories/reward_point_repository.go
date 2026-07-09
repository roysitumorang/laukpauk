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
	RewardPointRepository interface {
		FindRewardPoints(
			ctx context.Context,
			request *models.RewardPointFilter,
		) ([]*models.RewardPoint, int64, int64, error)
		FindRewardPoint(
			ctx context.Context,
			request *models.RewardPointFilter,
		) (*models.RewardPoint, error)
		CreateRewardPoint(
			ctx context.Context,
			tx pgx.Tx,
			request *models.RewardPoint,
		) (*models.RewardPoint, error)
		UpdateRewardPoint(
			ctx context.Context,
			tx pgx.Tx,
			request *models.RewardPoint,
		) (*models.RewardPoint, error)
	}

	rewardPointRepository struct {
		dbRead,
		dbWrite *pgxpool.Pool
	}
)

func NewRewardPointRepository(
	dbRead,
	dbWrite *pgxpool.Pool,
) RewardPointRepository {
	return new(rewardPointRepository{
		dbRead:  dbRead,
		dbWrite: dbWrite,
	})
}

func (q *rewardPointRepository) FindRewardPoints(
	ctx context.Context,
	request *models.RewardPointFilter,
) ([]*models.RewardPoint, int64, int64, error) {
	ctxt := "RewardPointRepository-FindRewardPoints"
	var sb strings.Builder
	condition, params := q.buildCondition(request, &sb)
	sb.Reset()
	_, _ = fmt.Fprintf(
		&sb,
		`SELECT COUNT(1)
		FROM reward_points p
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
		, p.record_status
		, p.reward_point_type
		, p.buyer_id
		, u.name
		, u.username
		, u.mobile_phone
		, b.referral_code
		, p.amount
		, p.description
		, p.created_by
		, p.created_at
		, p.updated_by
		, p.updated_at
		, p.deleted_by
		, p.deleted_at
		, p.deletion_reason`,
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
	var response []*models.RewardPoint
	for rows.Next() {
		var point models.RewardPoint
		if err = rows.Scan(
			&point.ID,
			&point.RecordStatus,
			&point.RewardPointType,
			&point.BuyerID,
			&point.BuyerName,
			&point.BuyerUsername,
			&point.BuyerMobilePhone,
			&point.BuyerReferralCode,
			&point.Amount,
			&point.Description,
			&point.CreatedBy,
			&point.CreatedAt,
			&point.UpdatedBy,
			&point.UpdatedAt,
			&point.DeletedBy,
			&point.DeletedAt,
			&point.DeletionReason,
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

func (q *rewardPointRepository) FindRewardPoint(
	ctx context.Context,
	request *models.RewardPointFilter,
) (*models.RewardPoint, error) {
	ctxt := "RewardPointRepository-FindRewardPoint"
	var sb strings.Builder
	condition, params := q.buildCondition(request, &sb)
	sb.Reset()
	_, _ = fmt.Fprintf(
		&sb,
		`SELECT
			p.id
			, p.record_status
			, p.reward_point_type
			, p.buyer_id
			, u.name
			, u.username
			, u.mobile_phone
			, b.referral_code
			, p.amount
			, p.description
			, p.created_by
			, p.created_at
			, p.updated_by
			, p.updated_at
			, p.deleted_by
			, p.deleted_at
			, p.deletion_reason
		FROM reward_points p
		JOIN users u ON p.buyer_id = u.id
		JOIN buyers b ON u.id = b.user_id %s`,
		condition,
	)
	query := sb.String()
	var response models.RewardPoint
	err := q.dbRead.QueryRow(ctx, query, params...).Scan(
		&response.ID,
		&response.RecordStatus,
		&response.RewardPointType,
		&response.BuyerID,
		&response.BuyerName,
		&response.BuyerUsername,
		&response.BuyerMobilePhone,
		&response.BuyerReferralCode,
		&response.Amount,
		&response.Description,
		&response.CreatedBy,
		&response.CreatedAt,
		&response.UpdatedBy,
		&response.UpdatedAt,
		&response.DeletedBy,
		&response.DeletedAt,
		&response.DeletionReason,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, models.ErrRewardPointNotFound
	}
	if err != nil {
		helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrScan")
		return nil, err
	}
	return &response, nil
}

func (q *rewardPointRepository) CreateRewardPoint(
	ctx context.Context,
	tx pgx.Tx,
	request *models.RewardPoint,
) (*models.RewardPoint, error) {
	ctxt := "RewardPointRepository-CreateRewardPoint"
	var response models.RewardPoint
	err := tx.QueryRow(
		ctx,
		`WITH p AS (
			INSERT INTO reward_points (
				record_status
				, reward_point_type
				, buyer_id
				, amount
				, description
				, created_by
				, created_at
				, updated_by
				, updated_at
			) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
			RETURNING *
		)
		SELECT
			p.id
			, p.record_status
			, p.reward_point_type
			, p.buyer_id
			, u.name
			, u.username
			, u.mobile_phone
			, b.referral_code
			, p.amount
			, p.description
			, p.created_by
			, p.created_at
			, p.updated_by
			, p.updated_at
			, p.deleted_by
			, p.deleted_at
			, p.deletion_reason
		FROM p
		JOIN users u ON p.buyer_id = u.id
		JOIN buyers b ON u.id = b.user_id`,
		request.RecordStatus,
		request.RewardPointType,
		request.BuyerID,
		request.Amount,
		request.Description,
		request.CreatedBy,
		request.CreatedAt,
		request.UpdatedBy,
		request.UpdatedAt,
	).Scan(
		&response.ID,
		&response.RecordStatus,
		&response.RewardPointType,
		&response.BuyerID,
		&response.BuyerName,
		&response.BuyerUsername,
		&response.BuyerMobilePhone,
		&response.BuyerReferralCode,
		&response.Amount,
		&response.Description,
		&response.CreatedBy,
		&response.CreatedAt,
		&response.UpdatedBy,
		&response.UpdatedAt,
		&response.DeletedBy,
		&response.DeletedAt,
		&response.DeletionReason,
	)
	if err != nil {
		helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrScan")
		return nil, err
	}
	return &response, nil
}

func (q *rewardPointRepository) UpdateRewardPoint(
	ctx context.Context,
	tx pgx.Tx,
	request *models.RewardPoint,
) (*models.RewardPoint, error) {
	ctxt := "RewardPointRepository-UpdateRewardPoint"
	var response models.RewardPoint
	err := tx.QueryRow(
		ctx,
		`WITH p AS (
			UPDATE reward_points SET
				record_status = $1
				, amount = $2
				, description = $3
				, updated_by = $4
				, updated_at = $5
				, deleted_by = $6
				, deleted_at = $7
				, deletion_reason = $8
			WHERE id = $9
			RETURNING *
		)
		SELECT
			p.id
			, p.record_status
			, p.reward_point_type
			, p.buyer_id
			, u.name
			, u.username
			, u.mobile_phone
			, b.referral_code
			, p.amount
			, p.description
			, p.created_by
			, p.created_at
			, p.updated_by
			, p.updated_at
			, p.deleted_by
			, p.deleted_at
			, p.deletion_reason
		FROM p
		JOIN users u ON p.buyer_id = u.id
		JOIN buyers b ON u.id = b.user_id`,
		request.RecordStatus,
		request.Amount,
		request.Description,
		request.UpdatedBy,
		request.UpdatedAt,
		request.DeletedBy,
		request.DeletedAt,
		request.DeletionReason,
		request.ID,
	).Scan(
		&response.ID,
		&response.RecordStatus,
		&response.RewardPointType,
		&response.BuyerID,
		&response.BuyerName,
		&response.BuyerUsername,
		&response.BuyerMobilePhone,
		&response.BuyerReferralCode,
		&response.Amount,
		&response.Description,
		&response.CreatedBy,
		&response.CreatedAt,
		&response.UpdatedBy,
		&response.UpdatedAt,
		&response.DeletedBy,
		&response.DeletedAt,
		&response.DeletionReason,
	)
	if err != nil {
		helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrScan")
		return nil, err
	}
	return &response, nil
}

func (q *rewardPointRepository) buildCondition(request *models.RewardPointFilter, sb *strings.Builder) (condition string, params []any) {
	var conds []string
	if len(request.RewardPointIDs) > 0 {
		sb.Reset()
		_, _ = sb.WriteString("p.id IN (")
		for i, pointID := range request.RewardPointIDs {
			params = append(params, pointID)
			if i > 0 {
				_, _ = sb.WriteString(",")
			}
			_, _ = fmt.Fprintf(sb, "$%d", len(params))
		}
		_, _ = sb.WriteString(")")
		conds = append(conds, sb.String())
	}
	if len(request.RecordStatuses) > 0 {
		sb.Reset()
		_, _ = sb.WriteString("p.record_status IN (")
		for i, recordStatus := range request.RecordStatuses {
			params = append(params, recordStatus)
			if i > 0 {
				_, _ = sb.WriteString(",")
			}
			_, _ = fmt.Fprintf(sb, "$%d", len(params))
		}
		_, _ = sb.WriteString(")")
		conds = append(conds, sb.String())
	}
	if len(request.RewardPointTypes) > 0 {
		sb.Reset()
		_, _ = sb.WriteString("p.reward_point_type IN (")
		for i, pointType := range request.RewardPointTypes {
			params = append(params, pointType)
			if i > 0 {
				_, _ = sb.WriteString(",")
			}
			_, _ = fmt.Fprintf(sb, "$%d", len(params))
		}
		_, _ = sb.WriteString(")")
		conds = append(conds, sb.String())
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
