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
		deliveryPlanRepository repositories.DeliveryPlanRepository
	}
)

func NewTransactionDayService(
	dayRepository repositories.TransactionDayRepository,
	deliveryPlanRepository repositories.DeliveryPlanRepository,
) TransactionDayService {
	return &dayService{
		dayRepository:          dayRepository,
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
			WithDeliveryWeekdays(weekday),
	)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindDeliveryPlans")
		return nil, err
	}
	var t time.Time
	request.TransactionWindows = make([]*models.TransactionWindow, len(deliveryPlans))
	for i, deliveryPlan := range deliveryPlans {
		var deliverySchedules []*models.DeliverySchedule
		if deliveryPlan.SameDay {
			for _, deliveryTime := range deliveryPlan.SameDayDeliveryTimes {
				t = deliveryTime.Timestamp
				deliverySchedules = append(deliverySchedules, &models.DeliverySchedule{
					DeliverySpeed: deliveryTime.DeliverySpeed,
					DeliveryTime:  time.Date(now.Year(), now.Month(), now.Day(), t.Hour(), t.Minute(), t.Second(), 0, t.Location()),
					CreatedAt:     now,
				})
			}
		}
		if deliveryPlan.NextDay {
			for _, deliveryTime := range deliveryPlan.NextDayDeliveryTimes {
				t = deliveryTime.Timestamp
				deliverySchedules = append(deliverySchedules, &models.DeliverySchedule{
					DeliverySpeed: deliveryTime.DeliverySpeed,
					DeliveryTime:  time.Date(now.Year(), now.Month(), now.Day()+1, t.Hour(), t.Minute(), t.Second(), 0, t.Location()),
					CreatedAt:     now,
				})
			}
		}
		if deliveryPlan.NextTwoDay {
			for _, deliveryTime := range deliveryPlan.NextTwoDayDeliveryTimes {
				t = deliveryTime.Timestamp
				deliverySchedules = append(deliverySchedules, &models.DeliverySchedule{
					DeliverySpeed: deliveryTime.DeliverySpeed,
					DeliveryTime:  time.Date(now.Year(), now.Month(), now.Day()+2, t.Hour(), t.Minute(), t.Second(), 0, t.Location()),
					CreatedAt:     now,
				})
			}
		}
		min := deliveryPlan.TransactionWindowMin
		max := deliveryPlan.TransactionWindowMax
		request.TransactionWindows[i] = &models.TransactionWindow{
			DeliveryPlanID:       deliveryPlan.ID,
			TransactionWindowMin: time.Date(now.Year(), now.Month(), now.Day(), min.Hour(), min.Minute(), min.Second(), 0, t.Location()),
			TransactionWindowMax: time.Date(now.Year(), now.Month(), now.Day(), max.Hour(), max.Minute(), max.Second(), 0, t.Location()),
			CreatedAt:            now,
			DeliverySchedules:    deliverySchedules,
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
