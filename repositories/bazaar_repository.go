package repositories

import (
	"context"
	"errors"
	"log/slog"
	"strconv"
	"strings"

	"github.com/govalues/decimal"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/roysitumorang/laukpauk/helper"
	"github.com/roysitumorang/laukpauk/models"
)

type (
	BazaarRepository interface {
		FindBazaars(
			ctx context.Context,
			request *models.BazaarFilter,
		) ([]*models.Bazaar, int64, int64, error)
		CreateBazaar(
			ctx context.Context,
			tx pgx.Tx,
			request *models.Bazaar,
		) (*models.Bazaar, error)
		UpdateBazaar(
			ctx context.Context,
			tx pgx.Tx,
			request *models.Bazaar,
		) (*models.Bazaar, error)
	}

	bazaarRepository struct {
		dbRead,
		dbWrite *pgxpool.Pool
	}
)

func NewBazaarRepository(
	dbRead,
	dbWrite *pgxpool.Pool,
) BazaarRepository {
	return new(bazaarRepository{
		dbRead:  dbRead,
		dbWrite: dbWrite,
	})
}

func (q *bazaarRepository) FindBazaars(
	ctx context.Context,
	request *models.BazaarFilter,
) ([]*models.Bazaar, int64, int64, error) {
	ctxt := "BazaarRepository-FindBazaars"
	var (
		params []any
		conds  []string
		sb     strings.Builder
	)
	if len(request.BazaarIDs) > 0 {
		sb.Reset()
		_, _ = sb.WriteString("id IN (")
		for i, bazaarID := range request.BazaarIDs {
			params = append(params, bazaarID)
			if i > 0 {
				_, _ = sb.WriteString(",")
			}
			_, _ = sb.WriteString("$")
			_, _ = sb.WriteString(strconv.Itoa(len(params)))
		}
		_, _ = sb.WriteString(")")
		conds = append(conds, sb.String())
	}
	if len(request.RecordStatus) > 0 {
		sb.Reset()
		_, _ = sb.WriteString("record_status IN (")
		for i, recordStatus := range request.RecordStatus {
			params = append(params, recordStatus)
			if i > 0 {
				_, _ = sb.WriteString(",")
			}
			_, _ = sb.WriteString("$")
			_, _ = sb.WriteString(strconv.Itoa(len(params)))
		}
		_, _ = sb.WriteString(")")
		conds = append(conds, sb.String())
	}
	if request.Keyword != "" {
		sb.Reset()
		_, _ = sb.WriteString("%%")
		_, _ = sb.WriteString(strings.ToLower(request.Keyword))
		_, _ = sb.WriteString("%%")
		params = append(params, sb.String())
		n := strconv.Itoa(len(params))
		sb.Reset()
		_, _ = sb.WriteString("LOWER(name) LIKE $")
		_, _ = sb.WriteString(n)
		conds = append(conds, sb.String())
	}
	sb.Reset()
	_, _ = sb.WriteString(
		`SELECT COUNT(1)
		FROM bazaars`,
	)
	if len(conds) > 0 {
		_, _ = sb.WriteString(" WHERE")
		for i, cond := range conds {
			if i > 0 {
				_, _ = sb.WriteString(" AND")
			}
			_, _ = sb.WriteString(" ")
			_, _ = sb.WriteString(cond)
		}
	}
	query := sb.String()
	var total int64
	err := q.dbRead.QueryRow(ctx, query, params...).Scan(&total)
	if err != nil {
		helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
		return nil, 0, 0, err
	}
	if total == 0 {
		return nil, 0, 0, nil
	}
	query = strings.ReplaceAll(
		query,
		"COUNT(1)",
		`ROW_NUMBER() OVER (ORDER BY id DESC) AS row_no
		, id
		, record_status
		, name
		, address
		, photo
		, coordinates[0]
		, coordinates[1]
		, bazaar_type
		, created_by
		, created_at
		, updated_by
		, updated_at
		, deleted_by
		, deleted_at
		, deletion_reason`,
	)
	sb.Reset()
	_, _ = sb.WriteString(query)
	pages := int64(1)
	if request.Limit > 0 {
		totalDecimal, err := decimal.New(total, 0)
		if err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrNew")
			return nil, 0, 0, err
		}
		perPageDecimal, err := decimal.New(request.Limit, 0)
		if err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrNew")
			return nil, 0, 0, err
		}
		pagesDecimal, err := totalDecimal.Quo(perPageDecimal)
		if err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrQuo")
			return nil, 0, 0, err
		}
		pages, _, _ = pagesDecimal.Ceil(0).Int64(0)
		offset := request.Page * request.Limit
		_, _ = sb.WriteString(" LIMIT ")
		_, _ = sb.WriteString(strconv.FormatInt(request.Limit, 10))
		_, _ = sb.WriteString(" OFFSET ")
		_, _ = sb.WriteString(strconv.FormatInt(offset, 10))
	}
	rows, err := q.dbRead.Query(ctx, sb.String(), params...)
	if errors.Is(err, pgx.ErrNoRows) {
		err = nil
	}
	if err != nil {
		helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrQuery")
		return nil, 0, 0, err
	}
	defer rows.Close()
	var response []*models.Bazaar
	for rows.Next() {
		var bazaar models.Bazaar
		if err = rows.Scan(
			&bazaar.RowNo,
			&bazaar.ID,
			&bazaar.RecordStatus,
			&bazaar.Name,
			&bazaar.Address,
			&bazaar.Photo,
			&bazaar.Longitude,
			&bazaar.Latitude,
			&bazaar.BazaarType,
			&bazaar.CreatedBy,
			&bazaar.CreatedAt,
			&bazaar.UpdatedBy,
			&bazaar.UpdatedAt,
			&bazaar.DeletedBy,
			&bazaar.DeletedAt,
			&bazaar.DeletionReason,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrScan")
			return nil, 0, 0, err
		}
		response = append(response, &bazaar)
	}
	return response, total, pages, nil
}

func (q *bazaarRepository) CreateBazaar(
	ctx context.Context,
	tx pgx.Tx,
	request *models.Bazaar,
) (*models.Bazaar, error) {
	ctxt := "BazaarRepository-CreateBazaar"
	var response models.Bazaar
	if err := tx.QueryRow(
		ctx,
		`INSERT INTO bazaars (
			record_status
			, name
			, address
			, photo
			, coordinates
			, bazaar_type
			, created_by
			, created_at
			, updated_by
			, updated_at
		) VALUES ($1, $2, $3, $4, POINT($5, $6), $7, $8, $9, $10, $11)
		RETURNING id
			, record_status
			, name
			, address
			, photo
			, coordinates[0]
			, coordinates[1]
			, bazaar_type
			, created_by
			, created_at
			, updated_by
			, updated_at
			, deleted_by
			, deleted_at
			, deletion_reason`,
		request.RecordStatus,
		request.Name,
		request.Address,
		request.Photo,
		request.Longitude,
		request.Latitude,
		request.BazaarType,
		request.CreatedBy,
		request.CreatedAt,
		request.UpdatedBy,
		request.UpdatedAt,
	).Scan(
		&response.ID,
		&response.RecordStatus,
		&response.Name,
		&response.Address,
		&response.Photo,
		&response.Longitude,
		&response.Latitude,
		&response.CreatedBy,
		&response.CreatedAt,
		&response.UpdatedBy,
		&response.UpdatedAt,
		&response.DeletedBy,
		&response.DeletedAt,
		&response.DeletionReason,
	); err != nil {
		if errRollback := tx.Rollback(ctx); errRollback != nil {
			helper.Capture(ctx, slog.LevelError, errRollback, ctxt, "ErrRollback")
		}
		helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrScan")
		return nil, err
	}
	return new(response), nil
}

func (q *bazaarRepository) UpdateBazaar(
	ctx context.Context,
	tx pgx.Tx,
	request *models.Bazaar,
) (*models.Bazaar, error) {
	ctxt := "BazaarRepository-UpdateBazaar"
	var response models.Bazaar
	if err := tx.QueryRow(
		ctx,
		`UPDATE bazaars SET
			record_status = $1
			, name = $2
			, address = $3
			, photo = $4
			, coordinates = POINT($5, $6)
			, bazaar_type = $7
			, updated_by = $8
			, updated_at = $9
			, deleted_by = $10
			, deleted_at = $11
			, deletion_reason = $12
		WHERE id = $13
		RETURNING id
			, record_status
			, name
			, address
			, photo
			, coordinates[0]
			, coordinates[1]
			, bazaar_type
			, created_by
			, created_at
			, updated_by
			, updated_at
			, deleted_by
			, deleted_at
			, deletion_reason`,
		request.RecordStatus,
		request.Name,
		request.Address,
		request.Photo,
		request.Longitude,
		request.Latitude,
		request.BazaarType,
		request.UpdatedBy,
		request.UpdatedAt,
		request.DeletedBy,
		request.DeletedAt,
		request.DeletionReason,
		request.ID,
	).Scan(
		&response.ID,
		&response.RecordStatus,
		&response.Name,
		&response.Address,
		&response.Photo,
		&response.Longitude,
		&response.Latitude,
		&response.CreatedBy,
		&response.CreatedAt,
		&response.UpdatedBy,
		&response.UpdatedAt,
		&response.DeletedBy,
		&response.DeletedAt,
		&response.DeletionReason,
	); err != nil {
		if errRollback := tx.Rollback(ctx); errRollback != nil {
			helper.Capture(ctx, slog.LevelError, errRollback, ctxt, "ErrRollback")
		}
		helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrScan")
		return nil, err
	}
	return new(response), nil
}
