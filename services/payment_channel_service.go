package services

import (
	"context"
	"errors"
	"log/slog"

	"github.com/roysitumorang/laukpauk/helper"
	"github.com/roysitumorang/laukpauk/models"
	"github.com/roysitumorang/laukpauk/repositories"
)

type (
	PaymentChannelService interface {
		FindPaymentChannels(
			ctx context.Context,
			request *models.PaymentChannelFilter,
		) ([]*models.PaymentChannel, *models.Pagination, error)
		FindPaymentChannel(
			ctx context.Context,
			request *models.PaymentChannelFilter,
		) (*models.PaymentChannel, error)
	}

	paymentChannelService struct {
		paymentChannelRepository repositories.PaymentChannelRepository
	}
)

func NewPaymentChannelService(
	paymentChannelRepository repositories.PaymentChannelRepository,
) PaymentChannelService {
	return &paymentChannelService{
		paymentChannelRepository: paymentChannelRepository,
	}
}

func (q *paymentChannelService) FindPaymentChannels(
	ctx context.Context,
	request *models.PaymentChannelFilter,
) ([]*models.PaymentChannel, *models.Pagination, error) {
	ctxt := "PaymentChannelService-FindPaymentChannels"
	paymentChannels, total, pages, err := q.paymentChannelRepository.FindPaymentChannels(ctx, request)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindPaymentChannels")
		return nil, nil, err
	}
	n := len(paymentChannels)
	rows := make([]*models.PaymentChannel, n)
	if n > 0 {
		copy(rows, paymentChannels)
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

func (q *paymentChannelService) FindPaymentChannel(
	ctx context.Context,
	request *models.PaymentChannelFilter,
) (*models.PaymentChannel, error) {
	ctxt := "PaymentChannelService-FindPaymentChannel"
	response, err := q.paymentChannelRepository.FindPaymentChannel(ctx, request)
	if err != nil && !errors.Is(err, models.ErrPaymentChannelNotFound) {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindPaymentChannel")
	}
	return response, err
}
