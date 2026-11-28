package services

import (
	"context"
	"errors"
	"log/slog"
	"uuid"

	"github.com/roysitumorang/laukpauk/database"
	"github.com/roysitumorang/laukpauk/helper"
	"github.com/roysitumorang/laukpauk/models"
	"github.com/roysitumorang/laukpauk/repositories"
)

type DeliveryAddressService interface {
	FindDeliveryAddresses(
		ctx context.Context,
		request *models.DeliveryAddressFilter,
	) ([]*models.DeliveryAddress, *models.Pagination, error)
	FindDeliveryAddress(
		ctx context.Context,
		request *models.DeliveryAddressFilter,
	) (*models.DeliveryAddress, error)
	CreateDeliveryAddress(
		ctx context.Context,
		tx database.Tx,
		request *models.DeliveryAddress,
	) (*models.DeliveryAddress, error)
	UpdateDeliveryAddress(
		ctx context.Context,
		tx database.Tx,
		request *models.DeliveryAddress,
	) (*models.DeliveryAddress, error)
	UnsetMain(
		ctx context.Context,
		tx database.Tx,
		userID uuid.UUID,
	) error
}

type deliveryAddressService struct {
	deliveryAddressRepository repositories.DeliveryAddressRepository
}

func NewDeliveryAddressService(
	deliveryAddressRepository repositories.DeliveryAddressRepository,
) DeliveryAddressService {
	return &deliveryAddressService{
		deliveryAddressRepository: deliveryAddressRepository,
	}
}

func (q *deliveryAddressService) FindDeliveryAddresses(
	ctx context.Context,
	request *models.DeliveryAddressFilter,
) ([]*models.DeliveryAddress, *models.Pagination, error) {
	ctxt := "DeliveryAddressService-FindDeliveryAddresses"
	deliveryAddresses, total, pages, err := q.deliveryAddressRepository.FindDeliveryAddresses(ctx, request)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindDeliveryAddresses")
		return nil, nil, err
	}
	n := len(deliveryAddresses)
	rows := make([]*models.DeliveryAddress, n)
	if n > 0 {
		copy(rows, deliveryAddresses)
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

func (q *deliveryAddressService) FindDeliveryAddress(
	ctx context.Context,
	request *models.DeliveryAddressFilter,
) (*models.DeliveryAddress, error) {
	ctxt := "DeliveryAddressService-FindDeliveryAddress"
	response, err := q.deliveryAddressRepository.FindDeliveryAddress(ctx, request)
	if err != nil && !errors.Is(err, models.ErrDeliveryAddressNotFound) {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindDeliveryAddress")
	}
	return response, err
}

func (q *deliveryAddressService) CreateDeliveryAddress(
	ctx context.Context,
	tx database.Tx,
	request *models.DeliveryAddress,
) (*models.DeliveryAddress, error) {
	ctxt := "DeliveryAddressService-CreateDeliveryAddress"
	response, err := q.deliveryAddressRepository.CreateDeliveryAddress(ctx, tx, request)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrCreateDeliveryAddress")
	}
	return response, err
}

func (q *deliveryAddressService) UpdateDeliveryAddress(
	ctx context.Context,
	tx database.Tx,
	request *models.DeliveryAddress,
) (*models.DeliveryAddress, error) {
	ctxt := "DeliveryAddressService-UpdateDeliveryAddress"
	response, err := q.deliveryAddressRepository.UpdateDeliveryAddress(ctx, tx, request)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrUpdateDeliveryAddress")
	}
	return response, err
}

func (q *deliveryAddressService) UnsetMain(
	ctx context.Context,
	tx database.Tx,
	userID uuid.UUID,
) error {
	ctxt := "DeliveryAddressService-UnsetMain"
	err := q.deliveryAddressRepository.UnsetMain(ctx, tx, userID)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrUnsetMain")
	}
	return err
}
