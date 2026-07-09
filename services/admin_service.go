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
	AdminService interface {
		FindAdmins(
			ctx context.Context,
			request *models.AdminFilter,
		) ([]*models.Admin, *models.Pagination, error)
		FindAdmin(
			ctx context.Context,
			request *models.AdminFilter,
		) (*models.Admin, error)
	}

	adminService struct {
		adminRepository repositories.AdminRepository
	}
)

func NewAdminService(
	adminRepository repositories.AdminRepository,
) AdminService {
	return new(adminService{
		adminRepository: adminRepository,
	})
}

func (q *adminService) FindAdmins(
	ctx context.Context,
	request *models.AdminFilter,
) ([]*models.Admin, *models.Pagination, error) {
	ctxt := "AdminService-FindAdmins"
	admins, total, pages, err := q.adminRepository.FindAdmins(ctx, request)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindAdmins")
		return nil, nil, err
	}
	n := len(admins)
	rows := make([]*models.Admin, n)
	if n > 0 {
		copy(rows, admins)
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

func (q *adminService) FindAdmin(
	ctx context.Context,
	request *models.AdminFilter,
) (*models.Admin, error) {
	ctxt := "AdminService-FindAdmin"
	response, err := q.adminRepository.FindAdmin(ctx, request)
	if err != nil && !errors.Is(err, models.ErrAdminNotFound) {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindAdmin")
	}
	return response, err
}
