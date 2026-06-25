package services

import (
	"context"
	"log/slog"

	"github.com/jackc/pgx/v5"
	"github.com/roysitumorang/laukpauk/helper"
	"github.com/roysitumorang/laukpauk/models"
	"github.com/roysitumorang/laukpauk/repositories"
)

type (
	BlockService interface {
		FindBlocks(
			ctx context.Context,
			request *models.BlockFilter,
		) ([]*models.Block, *models.Pagination, error)
		UpdateBlock(
			ctx context.Context,
			tx pgx.Tx,
			request *models.Block,
		) (*models.Block, error)
		FindBlockItems(
			ctx context.Context,
			request *models.BlockItemFilter,
		) ([]*models.BlockItem, *models.Pagination, error)
		CreateBlockItem(
			ctx context.Context,
			tx pgx.Tx,
			request *models.BlockItem,
		) (*models.BlockItem, error)
		UpdateBlockItem(
			ctx context.Context,
			tx pgx.Tx,
			request *models.BlockItem,
		) (*models.BlockItem, error)
	}

	blockService struct {
		blockRepository repositories.BlockRepository
	}
)

func NewBlockService(
	blockRepository repositories.BlockRepository,
) BlockService {
	return new(blockService{
		blockRepository: blockRepository,
	})
}

func (q *blockService) FindBlocks(
	ctx context.Context,
	request *models.BlockFilter,
) ([]*models.Block, *models.Pagination, error) {
	ctxt := "BlockService-FindBlocks"
	blocks, total, pages, err := q.blockRepository.FindBlocks(ctx, request)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindBlocks")
		return nil, nil, err
	}
	n := len(blocks)
	rows := make([]*models.Block, n)
	if n > 0 {
		copy(rows, blocks)
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

func (q *blockService) UpdateBlock(
	ctx context.Context,
	tx pgx.Tx,
	request *models.Block,
) (*models.Block, error) {
	ctxt := "BlockService-UpdateBlock"
	response, err := q.blockRepository.UpdateBlock(ctx, tx, request)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrUpdateBlock")
		return nil, err
	}
	return response, nil
}

func (q *blockService) FindBlockItems(
	ctx context.Context,
	request *models.BlockItemFilter,
) ([]*models.BlockItem, *models.Pagination, error) {
	ctxt := "BlockService-FindBlockItems"
	blockItems, total, pages, err := q.blockRepository.FindBlockItems(ctx, request)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrFindBlockItems")
		return nil, nil, err
	}
	n := len(blockItems)
	rows := make([]*models.BlockItem, n)
	if n > 0 {
		copy(rows, blockItems)
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

func (q *blockService) CreateBlockItem(
	ctx context.Context,
	tx pgx.Tx,
	request *models.BlockItem,
) (*models.BlockItem, error) {
	ctxt := "BlockService-CreateBlockItem"
	response, err := q.blockRepository.CreateBlockItem(ctx, tx, request)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrCreateBlockItem")
		return nil, err
	}
	return response, nil
}

func (q *blockService) UpdateBlockItem(
	ctx context.Context,
	tx pgx.Tx,
	request *models.BlockItem,
) (*models.BlockItem, error) {
	ctxt := "BlockService-UpdateBlockItem"
	response, err := q.blockRepository.UpdateBlockItem(ctx, tx, request)
	if err != nil {
		helper.Log(ctx, slog.LevelError, err.Error(), ctxt, "ErrUpdateBlockItem")
		return nil, err
	}
	return response, nil
}
