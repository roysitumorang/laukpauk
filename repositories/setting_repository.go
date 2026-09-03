package repositories

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/roysitumorang/laukpauk/helper"
	"github.com/roysitumorang/laukpauk/models"
	"github.com/roysitumorang/laukpauk/pools"
)

type SettingRepository interface {
	FindSetting(
		ctx context.Context,
	) (*models.Setting, error)
	UpdateSetting(
		ctx context.Context,
		tx pgx.Tx,
		request *models.Setting,
	) (*models.Setting, error)
}

type settingRepository struct {
	dbRead,
	dbWrite *pgxpool.Pool
}

func NewSettingRepository(
	dbRead,
	dbWrite *pgxpool.Pool,
) SettingRepository {
	return &settingRepository{
		dbRead:  dbRead,
		dbWrite: dbWrite,
	}
}

func (q *settingRepository) FindSetting(
	ctx context.Context,
) (*models.Setting, error) {
	ctxt := "SettingRepository-FindSetting"
	var response models.Setting
	err := q.dbRead.QueryRow(
		ctx,
		`SELECT
			id
			, maintenance_mode
			, subscription_trial_period
			, monthly_billing_amount
			, annual_billing_amount
			, billing_grace_period
			, created_by
			, created_at
			, updated_by
			, updated_at
		FROM setting`,
	).Scan(
		&response.ID,
		&response.MaintenanceMode,
		&response.SubscriptionTrialPeriod,
		&response.MonthlyBillingAmount,
		&response.AnnualBillingAmount,
		&response.BillingGracePeriod,
		&response.CreatedBy,
		&response.CreatedAt,
		&response.UpdatedBy,
		&response.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, models.ErrSettingNotFound
	}
	if err != nil {
		helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrScan")
		return nil, err
	}
	return (&response).Populate(), nil
}

func (q *settingRepository) UpdateSetting(
	ctx context.Context,
	tx pgx.Tx,
	request *models.Setting,
) (*models.Setting, error) {
	ctxt := "SettingRepository-UpdateSetting"
	var response models.Setting
	err := tx.QueryRow(
		ctx,
		`SELECT
			id
			, maintenance_mode
			, subscription_trial_period
			, monthly_billing_amount
			, annual_billing_amount
			, billing_grace_period
			, created_by
			, created_at
			, updated_by
			, updated_at
		FROM setting
		WHERE id = 1
		FOR NO KEY UPDATE`,
	).Scan(
		&response.ID,
		&response.MaintenanceMode,
		&response.SubscriptionTrialPeriod,
		&response.MonthlyBillingAmount,
		&response.AnnualBillingAmount,
		&response.BillingGracePeriod,
		&response.CreatedBy,
		&response.CreatedAt,
		&response.UpdatedBy,
		&response.UpdatedAt,
	)
	if err != nil {
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
	_, _ = sb.WriteString("UPDATE setting SET")
	if response.MaintenanceMode != request.MaintenanceMode {
		params = append(params, request.MaintenanceMode)
		if n = len(params); n > 1 {
			_, _ = sb.WriteString(",")
		}
		_, _ = fmt.Fprintf(sb, " maintenance_mode = $%d", n)
	}
	if response.SubscriptionTrialPeriod != request.SubscriptionTrialPeriod {
		params = append(params, request.SubscriptionTrialPeriod)
		if n = len(params); n > 1 {
			_, _ = sb.WriteString(",")
		}
		_, _ = fmt.Fprintf(sb, " subscription_trial_period = $%d", n)
	}
	if response.MonthlyBillingAmount != request.MonthlyBillingAmount {
		params = append(params, request.MonthlyBillingAmount)
		if n = len(params); n > 1 {
			_, _ = sb.WriteString(",")
		}
		_, _ = fmt.Fprintf(sb, " monthly_billing_amount = $%d", n)
	}
	if response.AnnualBillingAmount != request.AnnualBillingAmount {
		params = append(params, request.AnnualBillingAmount)
		if n = len(params); n > 1 {
			_, _ = sb.WriteString(",")
		}
		_, _ = fmt.Fprintf(sb, " annual_billing_amount = $%d", n)
	}
	if response.BillingGracePeriod != request.BillingGracePeriod {
		params = append(params, request.BillingGracePeriod)
		if n = len(params); n > 1 {
			_, _ = sb.WriteString(",")
		}
		_, _ = fmt.Fprintf(sb, " billing_grace_period = $%d", n)
	}
	if len(params) == 0 {
		return &response, nil
	}
	params = append(params, request.UpdatedBy, request.UpdatedAt, request.ID)
	n = len(params)
	_, _ = fmt.Fprintf(
		sb,
		`, updated_by = $%d
		, updated_at = $%d
		WHERE id = $%d
		RETURNING id
			, maintenance_mode
			, subscription_trial_period
			, monthly_billing_amount
			, annual_billing_amount
			, billing_grace_period
			, created_by
			, created_at
			, updated_by
			, updated_at`,
		n-2, n-1, n,
	)
	if err = tx.QueryRow(ctx, sb.String(), params...).Scan(
		&response.ID,
		&response.MaintenanceMode,
		&response.SubscriptionTrialPeriod,
		&response.MonthlyBillingAmount,
		&response.AnnualBillingAmount,
		&response.BillingGracePeriod,
		&response.CreatedBy,
		&response.CreatedAt,
		&response.UpdatedBy,
		&response.UpdatedAt,
	); err != nil {
		helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrScan")
		return nil, err
	}
	return (&response).Populate(), nil
}
