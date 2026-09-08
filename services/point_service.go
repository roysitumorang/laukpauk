package services

import (
	"context"
	"errors"
	"log/slog"
	"uuid"

	"github.com/jackc/pgx/v5"
	"github.com/roysitumorang/laukpauk/helper"
	"github.com/roysitumorang/laukpauk/models"
	"github.com/roysitumorang/laukpauk/repositories"
)

type PointService interface {
	FindPoints(
		ctx context.Context,
		request *models.PointFilter,
	) ([]*models.Point, *models.Pagination, error)
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

type pointService struct {
	pointRepository       repositories.PointRepository
	buyerRepository       repositories.BuyerRepository
	transactionRepository repositories.TransactionRepository
}

func NewPointService(
	pointRepository repositories.PointRepository,
	buyerRepository repositories.BuyerRepository,
	transactionRepository repositories.TransactionRepository,
) PointService {
	return &pointService{
		pointRepository:       pointRepository,
		buyerRepository:       buyerRepository,
		transactionRepository: transactionRepository,
	}
}

func (q *pointService) FindPoints(
	ctx context.Context,
	request *models.PointFilter,
) ([]*models.Point, *models.Pagination, error) {
	ctxt := "PointService-FindPoints"
	points, total, pages, err := q.pointRepository.FindPoints(ctx, request)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindPoints")
		return nil, nil, err
	}
	n := len(points)
	rows := make([]*models.Point, n)
	if n > 0 {
		var buyerIDs []uuid.UUID
		mapBuyerIDs := map[uuid.UUID]struct{}{}
		var transactionIDs []uuid.UUID
		mapTransactionIDs := map[uuid.UUID]struct{}{}
		var ok bool
		for _, point := range points {
			if _, ok = mapBuyerIDs[point.BuyerID]; !ok {
				buyerIDs = append(buyerIDs, point.BuyerID)
				mapBuyerIDs[point.BuyerID] = struct{}{}
			}
			if point.TransactionID != nil {
				if _, ok = mapTransactionIDs[*point.TransactionID]; !ok {
					transactionIDs = append(transactionIDs, *point.TransactionID)
					mapTransactionIDs[*point.TransactionID] = struct{}{}
				}
			}
		}
		mapBuyers := map[uuid.UUID]*models.Buyer{}
		if len(buyerIDs) > 0 {
			buyers, _, _, err := q.buyerRepository.FindBuyers(
				ctx,
				models.NewBuyerFilter().
					WithBuyerIDs(buyerIDs...),
			)
			if err != nil {
				helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindBuyers")
				return nil, nil, err
			}
			for _, buyer := range buyers {
				mapBuyers[buyer.ID] = buyer
			}
		}
		mapTransactions := map[uuid.UUID]*models.Transaction{}
		if len(transactionIDs) > 0 {
			transactions, _, _, err := q.transactionRepository.FindTransactions(
				ctx,
				models.NewTransactionFilter().
					WithTransactionIDs(transactionIDs...),
			)
			if err != nil {
				helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindTransactions")
				return nil, nil, err
			}
			for _, transaction := range transactions {
				mapTransactions[transaction.ID] = transaction
			}
		}
		for i, point := range points {
			if buyer, ok := mapBuyers[point.BuyerID]; ok {
				point.BuyerName = buyer.Name
				point.BuyerUsername = buyer.Username
				point.BuyerMobilePhone = buyer.MobilePhone
				point.BuyerReferralCode = buyer.ReferralCode
			}
			if point.TransactionID != nil {
				if transaction, ok := mapTransactions[*point.TransactionID]; ok {
					point.TransactionReferenceNumber = &transaction.ReferenceNumber
				}
			}
			points[i] = point
		}
		copy(rows, points)
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

func (q *pointService) FindPoint(
	ctx context.Context,
	request *models.PointFilter,
) (*models.Point, error) {
	ctxt := "PointService-FindPoint"
	response, err := q.pointRepository.FindPoint(ctx, request)
	if errors.Is(err, models.ErrPointNotFound) {
		return nil, err
	}
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindPoint")
		return nil, err
	}
	buyer, err := q.buyerRepository.FindBuyer(
		ctx,
		models.NewBuyerFilter().
			WithBuyerIDs(response.BuyerID),
	)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindBuyer")
		return nil, err
	}
	response.BuyerName = buyer.Name
	response.BuyerUsername = buyer.Username
	response.BuyerMobilePhone = buyer.MobilePhone
	response.BuyerReferralCode = buyer.ReferralCode
	if response.TransactionID != nil {
		transaction, err := q.transactionRepository.FindTransaction(
			ctx,
			models.NewTransactionFilter().
				WithTransactionIDs(*response.TransactionID),
		)
		if err != nil {
			helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindTransaction")
			return nil, err
		}
		response.TransactionReferenceNumber = &transaction.ReferenceNumber
	}
	return response, nil
}

func (q *pointService) CreatePoint(
	ctx context.Context,
	tx pgx.Tx,
	request *models.Point,
) (*models.Point, error) {
	ctxt := "PointService-CreatePoint"
	response, err := q.pointRepository.CreatePoint(ctx, tx, request)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrCreatePoint")
		return nil, err
	}
	buyer, err := q.buyerRepository.FindBuyer(
		ctx,
		models.NewBuyerFilter().
			WithBuyerIDs(response.BuyerID),
	)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindBuyer")
		return nil, err
	}
	response.BuyerName = buyer.Name
	response.BuyerUsername = buyer.Username
	response.BuyerMobilePhone = buyer.MobilePhone
	response.BuyerReferralCode = buyer.ReferralCode
	if response.TransactionID != nil {
		transaction, err := q.transactionRepository.FindTransaction(
			ctx,
			models.NewTransactionFilter().
				WithTransactionIDs(*response.TransactionID),
		)
		if err != nil {
			helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindTransaction")
			return nil, err
		}
		response.TransactionReferenceNumber = &transaction.ReferenceNumber
	}
	return response, nil
}

func (q *pointService) UpdatePoint(
	ctx context.Context,
	tx pgx.Tx,
	request *models.Point,
) (*models.Point, error) {
	ctxt := "PointService-UpdatePoint"
	response, err := q.pointRepository.UpdatePoint(ctx, tx, request)
	if errors.Is(err, models.ErrPointNotFound) {
		return nil, err
	}
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrUpdatePoint")
		return nil, err
	}
	buyer, err := q.buyerRepository.FindBuyer(
		ctx,
		models.NewBuyerFilter().
			WithBuyerIDs(response.BuyerID),
	)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindBuyer")
		return nil, err
	}
	response.BuyerName = buyer.Name
	response.BuyerUsername = buyer.Username
	response.BuyerMobilePhone = buyer.MobilePhone
	response.BuyerReferralCode = buyer.ReferralCode
	if response.TransactionID != nil {
		transaction, err := q.transactionRepository.FindTransaction(
			ctx,
			models.NewTransactionFilter().
				WithTransactionIDs(*response.TransactionID),
		)
		if err != nil {
			helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindTransaction")
			return nil, err
		}
		response.TransactionReferenceNumber = &transaction.ReferenceNumber
	}
	return response, nil
}
