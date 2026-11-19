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

type AddressService interface {
	FindAddresses(
		ctx context.Context,
		request *models.AddressFilter,
	) ([]*models.Address, *models.Pagination, error)
	FindAddress(
		ctx context.Context,
		request *models.AddressFilter,
	) (*models.Address, error)
	CreateAddress(
		ctx context.Context,
		tx pgx.Tx,
		request *models.Address,
	) (*models.Address, error)
	UpdateAddress(
		ctx context.Context,
		tx pgx.Tx,
		request *models.Address,
	) (*models.Address, error)
	UnsetMain(
		ctx context.Context,
		tx pgx.Tx,
		userID uuid.UUID,
	) error
}

type addressService struct {
	bannerRepository repositories.AddressRepository
}

func NewAddressService(
	bannerRepository repositories.AddressRepository,
) AddressService {
	return &addressService{
		bannerRepository: bannerRepository,
	}
}

func (q *addressService) FindAddresses(
	ctx context.Context,
	request *models.AddressFilter,
) ([]*models.Address, *models.Pagination, error) {
	ctxt := "AddressService-FindAddresses"
	addresses, total, pages, err := q.bannerRepository.FindAddresses(ctx, request)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindAddresses")
		return nil, nil, err
	}
	n := len(addresses)
	rows := make([]*models.Address, n)
	if n > 0 {
		copy(rows, addresses)
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

func (q *addressService) FindAddress(
	ctx context.Context,
	request *models.AddressFilter,
) (*models.Address, error) {
	ctxt := "AddressService-FindAddress"
	response, err := q.bannerRepository.FindAddress(ctx, request)
	if err != nil && !errors.Is(err, models.ErrAddressNotFound) {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindAddress")
	}
	return response, err
}

func (q *addressService) CreateAddress(
	ctx context.Context,
	tx pgx.Tx,
	request *models.Address,
) (*models.Address, error) {
	ctxt := "AddressService-CreateAddress"
	response, err := q.bannerRepository.CreateAddress(ctx, tx, request)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrCreateAddress")
	}
	return response, err
}

func (q *addressService) UpdateAddress(
	ctx context.Context,
	tx pgx.Tx,
	request *models.Address,
) (*models.Address, error) {
	ctxt := "AddressService-UpdateAddress"
	response, err := q.bannerRepository.UpdateAddress(ctx, tx, request)
	if err != nil && !errors.Is(err, models.ErrAddressNotFound) {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrUpdateAddress")
	}
	return response, err
}

func (q *addressService) UnsetMain(
	ctx context.Context,
	tx pgx.Tx,
	userID uuid.UUID,
) error {
	ctxt := "AddressService-UnsetMain"
	err := q.bannerRepository.UnsetMain(ctx, tx, userID)
	if err != nil && !errors.Is(err, models.ErrAddressNotFound) {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrUnsetMain")
	}
	return err
}
