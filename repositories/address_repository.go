package repositories

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"uuid"

	"github.com/govalues/decimal"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/roysitumorang/laukpauk/helper"
	"github.com/roysitumorang/laukpauk/models"
	"github.com/roysitumorang/laukpauk/pools"
)

type AddressRepository interface {
	FindAddresses(
		ctx context.Context,
		request *models.AddressFilter,
	) ([]*models.Address, int64, int64, error)
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

type addressRepository struct {
	dbRead,
	dbWrite *pgxpool.Pool
}

func NewAddressRepository(
	dbRead,
	dbWrite *pgxpool.Pool,
) AddressRepository {
	return &addressRepository{
		dbRead:  dbRead,
		dbWrite: dbWrite,
	}
}

func (q *addressRepository) FindAddresses(
	ctx context.Context,
	request *models.AddressFilter,
) ([]*models.Address, int64, int64, error) {
	ctxt := "AddressRepository-FindAddresses"
	sb := pools.BuilderPool.Get()
	defer func() {
		sb.Reset()
		pools.BuilderPool.Put(sb)
	}()
	condition, params := q.buildCondition(request, sb)
	sb.Reset()
	_, _ = fmt.Fprintf(sb, "SELECT COUNT(1) FROM addresses %s", condition)
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
		`id
		, record_status
		, user_id
		, longitude
		, latitude
		, street_address
		, delivery_note
		, label
		, recipient_name
		, recipient_mobile_phone
		, main
		, created_at
		, updated_at
		, deleted_at
		, deletion_reason`,
	)
	sb.Reset()
	_, _ = sb.WriteString(query)
	sortBy := request.Sortings
	if len(sortBy) == 0 {
		sortBy = append(sortBy, models.DefaultSortBy)
	}
	_, _ = sb.WriteString(" ORDER BY")
	for i, item := range sortBy {
		if i > 0 {
			_, _ = sb.WriteString(",")
		}
		_, _ = fmt.Fprintf(sb, " %s %s", item.Field, item.Direction)
	}
	pages := int64(1)
	var offset int64
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
		offset = request.Page * request.Limit
		_, _ = fmt.Fprintf(sb, " LIMIT %d OFFSET %d", request.Limit, offset)
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
	var response []*models.Address
	for rows.Next() {
		var address models.Address
		if err = rows.Scan(
			&address.ID,
			&address.RecordStatus,
			&address.UserID,
			&address.Longitude,
			&address.Latitude,
			&address.StreetAddress,
			&address.DeliveryNote,
			&address.Label,
			&address.RecipientName,
			&address.RecipientMobilePhone,
			&address.Main,
			&address.CreatedAt,
			&address.UpdatedAt,
			&address.DeletedAt,
			&address.DeletionReason,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrScan")
			return nil, 0, 0, err
		}
		offset++
		address.RowNo = offset
		response = append(response, &address)
	}
	return response, total, pages, nil
}

func (q *addressRepository) FindAddress(
	ctx context.Context,
	request *models.AddressFilter,
) (*models.Address, error) {
	ctxt := "AddressRepository-FindAddress"
	sb := pools.BuilderPool.Get()
	defer func() {
		sb.Reset()
		pools.BuilderPool.Put(sb)
	}()
	condition, params := q.buildCondition(request, sb)
	sb.Reset()
	_, _ = fmt.Fprintf(
		sb,
		`SELECT
			id
			, record_status
			, user_id
			, longitude
			, latitude
			, street_address
			, delivery_note
			, label
			, recipient_name
			, recipient_mobile_phone
			, main
			, created_at
			, updated_at
			, deleted_at
			, deletion_reason
		FROM addresses %s`,
		condition,
	)
	query := sb.String()
	var response models.Address
	err := q.dbRead.QueryRow(ctx, query, params...).Scan(
		&response.ID,
		&response.RecordStatus,
		&response.UserID,
		&response.Longitude,
		&response.Latitude,
		&response.StreetAddress,
		&response.DeliveryNote,
		&response.Label,
		&response.RecipientName,
		&response.RecipientMobilePhone,
		&response.Main,
		&response.CreatedAt,
		&response.UpdatedAt,
		&response.DeletedAt,
		&response.DeletionReason,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, models.ErrAddressNotFound
	}
	if err != nil {
		helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrScan")
		return nil, err
	}
	return &response, nil
}

func (q *addressRepository) CreateAddress(
	ctx context.Context,
	tx pgx.Tx,
	request *models.Address,
) (*models.Address, error) {
	ctxt := "AddressRepository-CreateAddress"
	var response models.Address
	if err := tx.QueryRow(
		ctx,
		`INSERT INTO addresses (
			record_status
			, user_id
			, longitude
			, latitude
			, street_address
			, delivery_note
			, label
			, recipient_name
			, recipient_mobile_phone
			, main
			, created_at
			, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
		RETURNING id
			, record_status
			, user_id
			, longitude
			, latitude
			, street_address
			, delivery_note
			, label
			, recipient_name
			, recipient_mobile_phone
			, main
			, created_at
			, updated_at
			, deleted_at
			, deletion_reason`,
		request.RecordStatus,
		request.UserID,
		request.Longitude,
		request.Latitude,
		request.StreetAddress,
		request.DeliveryNote,
		request.Label,
		request.RecipientName,
		request.RecipientMobilePhone,
		request.Main,
		request.CreatedAt,
		request.UpdatedAt,
	).Scan(
		&response.ID,
		&response.RecordStatus,
		&response.UserID,
		&response.Longitude,
		&response.Latitude,
		&response.StreetAddress,
		&response.DeliveryNote,
		&response.Label,
		&response.RecipientName,
		&response.RecipientMobilePhone,
		&response.Main,
		&response.CreatedAt,
		&response.UpdatedAt,
		&response.DeletedAt,
		&response.DeletionReason,
	); err != nil {
		helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrScan")
		return nil, err
	}
	return &response, nil
}

func (q *addressRepository) UpdateAddress(
	ctx context.Context,
	tx pgx.Tx,
	request *models.Address,
) (*models.Address, error) {
	ctxt := "AddressRepository-UpdateAddress"
	var response models.Address
	err := tx.QueryRow(
		ctx,
		`SELECT
			id
			, record_status
			, user_id
			, longitude
			, latitude
			, street_address
			, delivery_note
			, label
			, recipient_name
			, recipient_mobile_phone
			, main
			, created_at
			, updated_at
			, deleted_at
			, deletion_reason
		FROM addresses
		WHERE id = $1
		FOR UPDATE`,
		request.ID,
	).Scan(
		&response.ID,
		&response.RecordStatus,
		&response.UserID,
		&response.Longitude,
		&response.Latitude,
		&response.StreetAddress,
		&response.DeliveryNote,
		&response.Label,
		&response.RecipientName,
		&response.RecipientMobilePhone,
		&response.Main,
		&response.CreatedAt,
		&response.UpdatedAt,
		&response.DeletedAt,
		&response.DeletionReason,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, models.ErrAddressNotFound
	}
	if err != nil {
		helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrScan")
		return nil, err
	}
	var params []any
	var n int
	sb := pools.BuilderPool.Get()
	defer func() {
		sb.Reset()
		pools.BuilderPool.Put(sb)
	}()
	_, _ = sb.WriteString("UPDATE addresses SET")
	if response.RecordStatus != request.RecordStatus {
		params = append(params, request.RecordStatus)
		if n = len(params); n > 1 {
			_, _ = sb.WriteString(",")
		}
		_, _ = fmt.Fprintf(sb, " record_status = $%d", n)
	}
	if response.Longitude != request.Longitude {
		params = append(params, request.Longitude)
		if n = len(params); n > 1 {
			_, _ = sb.WriteString(",")
		}
		_, _ = fmt.Fprintf(sb, " longitude = $%d", n)
	}
	if response.Latitude != request.Latitude {
		params = append(params, request.Latitude)
		if n = len(params); n > 1 {
			_, _ = sb.WriteString(",")
		}
		_, _ = fmt.Fprintf(sb, " latitude = $%d", n)
	}
	if response.StreetAddress != request.StreetAddress {
		params = append(params, request.StreetAddress)
		if n = len(params); n > 1 {
			_, _ = sb.WriteString(",")
		}
		_, _ = fmt.Fprintf(sb, " street_address = $%d", n)
	}
	if response.DeliveryNote != request.DeliveryNote {
		params = append(params, request.DeliveryNote)
		if n = len(params); n > 1 {
			_, _ = sb.WriteString(",")
		}
		_, _ = fmt.Fprintf(sb, " delivery_note = $%d", n)
	}
	if response.Label != request.Label {
		params = append(params, request.Label)
		if n = len(params); n > 1 {
			_, _ = sb.WriteString(",")
		}
		_, _ = fmt.Fprintf(sb, " label = $%d", n)
	}
	if response.RecipientName != request.RecipientName {
		params = append(params, request.RecipientName)
		if n = len(params); n > 1 {
			_, _ = sb.WriteString(",")
		}
		_, _ = fmt.Fprintf(sb, " recipient_name = $%d", n)
	}
	if response.RecipientMobilePhone != request.RecipientMobilePhone {
		params = append(params, request.RecipientMobilePhone)
		if n = len(params); n > 1 {
			_, _ = sb.WriteString(",")
		}
		_, _ = fmt.Fprintf(sb, " recipient_mobile_phone = $%d", n)
	}
	if response.Main != request.Main {
		params = append(params, request.Main)
		if n = len(params); n > 1 {
			_, _ = sb.WriteString(",")
		}
		_, _ = fmt.Fprintf(sb, " main = $%d", n)
	}
	if len(params) == 0 {
		return &response, nil
	}
	if !response.UpdatedAt.Equal(request.UpdatedAt) {
		params = append(params, request.UpdatedAt)
		if n = len(params); n > 1 {
			_, _ = sb.WriteString(",")
		}
		_, _ = fmt.Fprintf(sb, " updated_at = $%d", n)
	}
	if response.DeletedAt == nil && request.DeletedAt != nil {
		params = append(params, request.DeletedAt)
		if n = len(params); n > 1 {
			_, _ = sb.WriteString(",")
		}
		_, _ = fmt.Fprintf(sb, " deleted_at = $%d", n)
	}
	if response.DeletionReason == nil && request.DeletionReason != nil {
		params = append(params, request.DeletionReason)
		if n = len(params); n > 1 {
			_, _ = sb.WriteString(",")
		}
		_, _ = fmt.Fprintf(sb, " deletion_reason = $%d", n)
	}
	params = append(params, request.ID)
	_, _ = fmt.Fprintf(
		sb,
		` WHERE id = $%d
		RETURNING id
			, record_status
			, user_id
			, longitude
			, latitude
			, street_address
			, delivery_note
			, label
			, recipient_name
			, recipient_mobile_phone
			, main
			, created_at
			, updated_at
			, deleted_at
			, deletion_reason`,
		len(params),
	)
	if err = tx.QueryRow(ctx, sb.String(), params...).Scan(
		&response.ID,
		&response.RecordStatus,
		&response.UserID,
		&response.Longitude,
		&response.Latitude,
		&response.StreetAddress,
		&response.DeliveryNote,
		&response.Label,
		&response.RecipientName,
		&response.RecipientMobilePhone,
		&response.Main,
		&response.CreatedAt,
		&response.UpdatedAt,
		&response.DeletedAt,
		&response.DeletionReason,
	); err != nil {
		helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrScan")
		return nil, err
	}
	return &response, nil
}

func (q *addressRepository) UnsetMain(
	ctx context.Context,
	tx pgx.Tx,
	userID uuid.UUID,
) error {
	ctxt := "AddressRepository-UnsetMain"
	_, err := tx.Exec(
		ctx,
		`UPDATE addresses SET
			main = NOT main
		WHERE user_id = $1
			AND main`,
		userID,
	)
	if err != nil {
		helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
	}
	return nil
}

func (q *addressRepository) buildCondition(request *models.AddressFilter, sb *strings.Builder) (condition string, params []any) {
	var conds []string
	if len(request.AddressIDs) > 0 {
		sb.Reset()
		_, _ = sb.WriteString("id IN (")
		for i, addressID := range request.AddressIDs {
			params = append(params, addressID)
			if i > 0 {
				_, _ = sb.WriteString(",")
			}
			_, _ = fmt.Fprintf(sb, "$%d", len(params))
		}
		_, _ = sb.WriteString(")")
		conds = append(conds, sb.String())
	}
	if len(request.UserIDs) > 0 {
		sb.Reset()
		_, _ = sb.WriteString("user_id IN (")
		for i, userID := range request.UserIDs {
			params = append(params, userID)
			if i > 0 {
				_, _ = sb.WriteString(",")
			}
			_, _ = fmt.Fprintf(sb, "$%d", len(params))
		}
		_, _ = sb.WriteString(")")
		conds = append(conds, sb.String())
	}
	if len(request.RecordStatuses) > 0 {
		sb.Reset()
		_, _ = sb.WriteString("record_status IN (")
		for i, recordStatus := range request.RecordStatuses {
			params = append(params, recordStatus)
			if i > 0 {
				_, _ = sb.WriteString(",")
			}
			_, _ = fmt.Fprintf(sb, "$%d", len(params))
		}
		_, _ = sb.WriteString(")")
		conds = append(conds, sb.String())
	}
	if request.Longitude != "" {
		params = append(params, request.Longitude)
		sb.Reset()
		_, _ = fmt.Fprintf(sb, "longitude = $%d", len(params))
		conds = append(conds, sb.String())
	}
	if request.Latitude != "" {
		params = append(params, request.Latitude)
		sb.Reset()
		_, _ = fmt.Fprintf(sb, "latitude = $%d", len(params))
		conds = append(conds, sb.String())
	}
	if request.Label != "" {
		params = append(params, strings.ToLower(request.Label))
		sb.Reset()
		_, _ = fmt.Fprintf(sb, "LOWER(label) = $%d", len(params))
		conds = append(conds, sb.String())
	}
	if request.Main != nil {
		if *request.Main {
			conds = append(conds, "main")
		} else {
			conds = append(conds, "NOT main")
		}
	}
	if len(conds) > 0 {
		sb.Reset()
		_, _ = sb.WriteString("WHERE")
		for i, cond := range conds {
			if i > 0 {
				_, _ = sb.WriteString(" AND")
			}
			_, _ = fmt.Fprintf(sb, " %s", cond)
		}
		condition = sb.String()
	}
	return
}
