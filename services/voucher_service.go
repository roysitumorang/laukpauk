package services

import (
	"context"
	"errors"
	"log/slog"

	"github.com/jackc/pgx/v5"
	"github.com/roysitumorang/laukpauk/helper"
	"github.com/roysitumorang/laukpauk/models"
	"github.com/roysitumorang/laukpauk/repositories"
)

type VoucherService interface {
	FindVouchers(
		ctx context.Context,
		request *models.VoucherFilter,
	) ([]*models.Voucher, *models.Pagination, error)
	FindVoucher(
		ctx context.Context,
		request *models.VoucherFilter,
	) (*models.Voucher, error)
	CreateVoucher(
		ctx context.Context,
		tx pgx.Tx,
		request *models.Voucher,
	) (*models.Voucher, error)
	UpdateVoucher(
		ctx context.Context,
		tx pgx.Tx,
		request *models.Voucher,
	) (*models.Voucher, error)
}

type voucherService struct {
	voucherRepository repositories.VoucherRepository
}

func NewVoucherService(
	voucherRepository repositories.VoucherRepository,
) VoucherService {
	return &voucherService{
		voucherRepository: voucherRepository,
	}
}

func (q *voucherService) FindVouchers(
	ctx context.Context,
	request *models.VoucherFilter,
) ([]*models.Voucher, *models.Pagination, error) {
	ctxt := "VoucherService-FindVouchers"
	vouchers, total, pages, err := q.voucherRepository.FindVouchers(ctx, request)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindVouchers")
		return nil, nil, err
	}
	n := len(vouchers)
	rows := make([]*models.Voucher, n)
	if n > 0 {
		copy(rows, vouchers)
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

func (q *voucherService) FindVoucher(
	ctx context.Context,
	request *models.VoucherFilter,
) (*models.Voucher, error) {
	ctxt := "VoucherService-FindVoucher"
	response, err := q.voucherRepository.FindVoucher(ctx, request)
	if errors.Is(err, models.ErrVoucherNotFound) {
		return nil, err
	}
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindVoucher")
	}
	return response, err
}

func (q *voucherService) CreateVoucher(
	ctx context.Context,
	tx pgx.Tx,
	request *models.Voucher,
) (*models.Voucher, error) {
	ctxt := "VoucherService-CreateVoucher"
	response, err := q.voucherRepository.CreateVoucher(ctx, tx, request)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrCreateVoucher")
	}
	return response, nil
}
func (q *voucherService) UpdateVoucher(
	ctx context.Context,
	tx pgx.Tx,
	request *models.Voucher,
) (*models.Voucher, error) {
	ctxt := "VoucherService-UpdateVoucher"
	response, err := q.voucherRepository.UpdateVoucher(ctx, tx, request)
	if errors.Is(err, models.ErrVoucherNotFound) {
		return nil, err
	}
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrUpdateVoucher")
	}
	return response, nil
}
