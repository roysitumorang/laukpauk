package services

import (
	"context"
	"errors"
	"log/slog"
	"time"
	"uuid"

	"github.com/jackc/pgx/v5"
	"github.com/roysitumorang/laukpauk/helper"
	"github.com/roysitumorang/laukpauk/models"
	"github.com/roysitumorang/laukpauk/repositories"
)

type CompanyService interface {
	FindCompanies(
		ctx context.Context,
		request *models.CompanyFilter,
	) ([]*models.Company, *models.Pagination, error)
	FindCompany(
		ctx context.Context,
		request *models.CompanyFilter,
	) (*models.Company, error)
	CreateCompany(
		ctx context.Context,
		tx pgx.Tx,
		request *models.Company,
	) (*models.Company, error)
	UpdateCompany(
		ctx context.Context,
		tx pgx.Tx,
		request *models.Company,
	) (*models.Company, error)
	AddReferralPointsExpense(
		ctx context.Context,
		tx pgx.Tx,
		companyID uuid.UUID,
		addition int64,
		updatedBy uuid.UUID,
		updatedAt time.Time,
	) (*models.Company, error)
}

type companyService struct {
	companyRepository repositories.CompanyRepository
}

func NewCompanyService(
	companyRepository repositories.CompanyRepository,
) CompanyService {
	return &companyService{
		companyRepository: companyRepository,
	}
}

func (q *companyService) FindCompanies(
	ctx context.Context,
	request *models.CompanyFilter,
) ([]*models.Company, *models.Pagination, error) {
	ctxt := "CompanyService-FindCompanies"
	companies, total, pages, err := q.companyRepository.FindCompanies(ctx, request)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindCompanies")
		return nil, nil, err
	}
	n := len(companies)
	rows := make([]*models.Company, n)
	if n > 0 {
		copy(rows, companies)
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

func (q *companyService) FindCompany(
	ctx context.Context,
	request *models.CompanyFilter,
) (*models.Company, error) {
	ctxt := "CompanyService-FindCompany"
	response, err := q.companyRepository.FindCompany(ctx, request)
	if err != nil && !errors.Is(err, models.ErrCompanyNotFound) {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindCompany")
	}
	return response, err
}

func (q *companyService) CreateCompany(
	ctx context.Context,
	tx pgx.Tx,
	request *models.Company,
) (*models.Company, error) {
	ctxt := "CompanyService-CreateCompany"
	response, err := q.companyRepository.CreateCompany(ctx, tx, request)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrCreateCompany")
	}
	return response, err
}
func (q *companyService) UpdateCompany(
	ctx context.Context,
	tx pgx.Tx,
	request *models.Company,
) (*models.Company, error) {
	ctxt := "CompanyService-UpdateCompany"
	response, err := q.companyRepository.UpdateCompany(ctx, tx, request)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrUpdateCompany")
	}
	return response, err
}

func (q *companyService) AddReferralPointsExpense(
	ctx context.Context,
	tx pgx.Tx,
	companyID uuid.UUID,
	addition int64,
	updatedBy uuid.UUID,
	updatedAt time.Time,
) (*models.Company, error) {
	ctxt := "SettingService-AddReferralPointsExpense"
	response, err := q.companyRepository.AddReferralPointsExpense(
		ctx,
		tx,
		companyID,
		addition,
		updatedBy,
		updatedAt,
	)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrAddReferralPointsExpense")
	}
	return response, err
}
