package services

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/roysitumorang/laukpauk/helper"
	"github.com/roysitumorang/laukpauk/models"
	"github.com/roysitumorang/laukpauk/repositories"
)

type (
	TransactionDayService interface {
		FindTransactionDays(
			ctx context.Context,
			request *models.TransactionDayFilter,
		) ([]*models.TransactionDay, *models.Pagination, error)
		FindTransactionDay(
			ctx context.Context,
			request *models.TransactionDayFilter,
		) (*models.TransactionDay, error)
		CreateTransactionDay(
			ctx context.Context,
			tx pgx.Tx,
			request *models.TransactionDay,
		) (*models.TransactionDay, error)
		FindOrCreateTransactionDay(
			ctx context.Context,
			now time.Time,
		) (*models.TransactionDay, error)
		UpdateTransactionDay(
			ctx context.Context,
			tx pgx.Tx,
			request *models.TransactionDay,
		) (*models.TransactionDay, error)
	}

	dayService struct {
		dayRepository          repositories.TransactionDayRepository
		settingRepository      repositories.SettingRepository
		deliveryPlanRepository repositories.DeliveryPlanRepository
	}
)

func NewTransactionDayService(
	dayRepository repositories.TransactionDayRepository,
	settingRepository repositories.SettingRepository,
	deliveryPlanRepository repositories.DeliveryPlanRepository,
) TransactionDayService {
	return &dayService{
		dayRepository:          dayRepository,
		settingRepository:      settingRepository,
		deliveryPlanRepository: deliveryPlanRepository,
	}
}

func (q *dayService) FindTransactionDays(
	ctx context.Context,
	request *models.TransactionDayFilter,
) ([]*models.TransactionDay, *models.Pagination, error) {
	ctxt := "TransactionDayService-FindTransactionDays"
	days, total, pages, err := q.dayRepository.FindTransactionDays(ctx, request)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindTransactionDays")
		return nil, nil, err
	}
	n := len(days)
	rows := make([]*models.TransactionDay, n)
	if n > 0 {
		copy(rows, days)
	}
	pagination, err := models.NewPagination(
		total,
		pages,
		request.Limit,
		request.Page,
		request.PaginationURL,
		request.UrlValues,
	)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrNewPagination")
		return nil, nil, err
	}
	return rows, pagination, nil
}

func (q *dayService) FindTransactionDay(
	ctx context.Context,
	request *models.TransactionDayFilter,
) (*models.TransactionDay, error) {
	ctxt := "TransactionDayService-FindTransactionDay"
	response, err := q.dayRepository.FindTransactionDay(ctx, request)
	if err != nil && !errors.Is(err, models.ErrTransactionDayNotFound) {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindTransactionDay")
	}
	return response, err
}

func (q *dayService) CreateTransactionDay(
	ctx context.Context,
	tx pgx.Tx,
	request *models.TransactionDay,
) (*models.TransactionDay, error) {
	ctxt := "TransactionDayService-CreateTransactionDay"
	response, err := q.dayRepository.CreateTransactionDay(ctx, tx, request)
	if err != nil && !errors.Is(err, models.ErrUniqueTransactionDayDateViolation) {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrCreateTransactionDay")
	}
	return response, err
}

func (q *dayService) FindOrCreateTransactionDay(
	ctx context.Context,
	now time.Time,
) (*models.TransactionDay, error) {
	ctxt := "TransactionDayService-FindOrCreateTransactionDay"
	date := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	request, err := q.dayRepository.FindTransactionDay(
		ctx,
		models.NewTransactionDayFilter().
			WithDates(date),
	)
	if err != nil && !errors.Is(err, models.ErrTransactionDayNotFound) {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindTransactionDay")
		return nil, err
	}
	if request != nil {
		return request, nil
	}
	setting, err := q.settingRepository.FindSetting(ctx)
	if err != nil && !errors.Is(err, models.ErrSettingNotFound) {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindSetting")
		return nil, err
	}
	if setting == nil {
		return nil, models.ErrSettingNotFound
	}
	request = &models.TransactionDay{
		Date:      date,
		Today:     new(true),
		CreatedAt: now,
		UpdatedAt: now,
	}
	var weekday models.DeliveryWeekday
	switch now.Weekday() {
	case time.Sunday:
		weekday = models.DeliveryWeekdaySunday
	case time.Friday:
		weekday = models.DeliveryWeekdayFriday
	case time.Saturday:
		weekday = models.DeliveryWeekdaySaturday
	default:
		weekday = models.DeliveryWeekdayWeekdays
	}
	deliveryPlans, _, _, err := q.deliveryPlanRepository.FindDeliveryPlans(
		ctx,
		models.NewDeliveryPlanFilter().
			WithRecordStatuses(models.RecordStatusActive).
			WithDeliveryWeekdays(weekday).
			WithDeliveryTimeInterval(setting.DeliveryTimeInterval),
	)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindDeliveryPlans")
		return nil, err
	}
	var t time.Time
	request.TransactionWindows = make([]*models.TransactionWindow, len(deliveryPlans))
	for i, deliveryPlan := range deliveryPlans {
		deliveryTimes := make([]*models.DeliveryTime, len(deliveryPlan.DeliveryTimes))
		for i, deliveryTime := range deliveryPlan.DeliveryTimes {
			t = deliveryTime.Datetime
			deliveryTimes[i] = &models.DeliveryTime{
				DeliverySpeed: deliveryTime.DeliverySpeed,
				Datetime: time.Date(
					now.Year(),
					now.Month(),
					now.Day(),
					t.Hour(),
					t.Minute(),
					t.Second(),
					t.Nanosecond(),
					t.Location(),
				),
				CreatedAt: now,
			}
		}
		min := deliveryPlan.TransactionWindowMin
		max := deliveryPlan.TransactionWindowMax
		request.TransactionWindows[i] = &models.TransactionWindow{
			DeliveryPlanID: deliveryPlan.ID,
			TransactionWindowMin: time.Date(
				now.Year(),
				now.Month(),
				now.Day(),
				min.Hour(),
				min.Minute(),
				min.Second(),
				min.Nanosecond(),
				t.Location(),
			),
			TransactionWindowMax: time.Date(
				now.Year(),
				now.Month(),
				now.Day(),
				max.Hour(),
				max.Minute(),
				max.Second(),
				max.Nanosecond(),
				t.Location(),
			),
			CreatedAt:     now,
			DeliveryTimes: deliveryTimes,
		}
	}
	tx, err := helper.BeginTx(ctx)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrBeginTx")
		return nil, err
	}
	defer func(ctx context.Context) {
		if err = tx.Rollback(ctx); errors.Is(err, pgx.ErrTxClosed) {
			err = nil
		}
		if err != nil {
			helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrRollback")
		}
	}(ctx)
	response, err := q.dayRepository.CreateTransactionDay(ctx, tx, request)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrCreateTransactionDay")
		return nil, err
	}
	if err = tx.Commit(ctx); err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrCommit")
		return nil, err
	}
	return response, err
}

func (q *dayService) UpdateTransactionDay(
	ctx context.Context,
	tx pgx.Tx,
	request *models.TransactionDay,
) (*models.TransactionDay, error) {
	ctxt := "TransactionDayService-UpdateTransactionDay"
	response, err := q.dayRepository.UpdateTransactionDay(ctx, tx, request)
	if err != nil && !errors.Is(err, models.ErrTransactionDayNotFound) {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrUpdateTransactionDay")
	}
	return response, err
}
