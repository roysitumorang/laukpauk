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

type DeliveryAddressRepository interface {
	FindDeliveryAddresses(
		ctx context.Context,
		request *models.DeliveryAddressFilter,
	) ([]*models.DeliveryAddress, int64, int64, error)
	FindDeliveryAddress(
		ctx context.Context,
		request *models.DeliveryAddressFilter,
	) (*models.DeliveryAddress, error)
	CreateDeliveryAddress(
		ctx context.Context,
		tx pgx.Tx,
		request *models.DeliveryAddress,
	) (*models.DeliveryAddress, error)
	UpdateDeliveryAddress(
		ctx context.Context,
		tx pgx.Tx,
		request *models.DeliveryAddress,
	) (*models.DeliveryAddress, error)
	UnsetMain(
		ctx context.Context,
		tx pgx.Tx,
		userID uuid.UUID,
	) error
}

type deliveryAddressRepository struct {
	dbRead,
	dbWrite *pgxpool.Pool
}

func NewDeliveryAddressRepository(
	dbRead,
	dbWrite *pgxpool.Pool,
) DeliveryAddressRepository {
	return &deliveryAddressRepository{
		dbRead:  dbRead,
		dbWrite: dbWrite,
	}
}

func (q *deliveryAddressRepository) FindDeliveryAddresses(
	ctx context.Context,
	request *models.DeliveryAddressFilter,
) ([]*models.DeliveryAddress, int64, int64, error) {
	ctxt := "DeliveryAddressRepository-FindDeliveryAddresses"
	sb := pools.BuilderPool.Get()
	defer func() {
		sb.Reset()
		pools.BuilderPool.Put(sb)
	}()
	sb.Reset()
	condition, params := q.buildCondition(request, sb)
	sb.Reset()
	_, _ = fmt.Fprintf(sb, "SELECT COUNT(1) FROM delivery_addresses %s", condition)
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
		, origin_longitude
		, origin_latitude
		, destination_longitude
		, destination_latitude
		, distance_meters
		, distance_km
		, street_address
		, street_address_note
		, label
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
		if item.Field == models.DeliveryAddressSortingFieldMainFirst {
			_, _ = fmt.Fprintf(sb, " %s DESC NULLS LAST", item.Field)
		} else {
			_, _ = fmt.Fprintf(sb, " %s %s", item.Field, item.Direction)
		}
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
	var response []*models.DeliveryAddress
	for rows.Next() {
		var deliveryAddress models.DeliveryAddress
		if err = rows.Scan(
			&deliveryAddress.ID,
			&deliveryAddress.RecordStatus,
			&deliveryAddress.UserID,
			&deliveryAddress.OriginLongitude,
			&deliveryAddress.OriginLatitude,
			&deliveryAddress.DestinationLongitude,
			&deliveryAddress.DestinationLatitude,
			&deliveryAddress.DistanceMeters,
			&deliveryAddress.DistanceKM,
			&deliveryAddress.StreetAddress,
			&deliveryAddress.StreetAddressNote,
			&deliveryAddress.Label,
			&deliveryAddress.Main,
			&deliveryAddress.CreatedAt,
			&deliveryAddress.UpdatedAt,
			&deliveryAddress.DeletedAt,
			&deliveryAddress.DeletionReason,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrScan")
			return nil, 0, 0, err
		}
		offset++
		deliveryAddress.RowNo = offset
		response = append(response, &deliveryAddress)
	}
	if err = rows.Err(); err != nil {
		helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrQuery")
		return nil, 0, 0, err
	}
	return response, total, pages, nil
}

func (q *deliveryAddressRepository) FindDeliveryAddress(
	ctx context.Context,
	request *models.DeliveryAddressFilter,
) (*models.DeliveryAddress, error) {
	ctxt := "DeliveryAddressRepository-FindDeliveryAddress"
	sb := pools.BuilderPool.Get()
	defer func() {
		sb.Reset()
		pools.BuilderPool.Put(sb)
	}()
	sb.Reset()
	condition, params := q.buildCondition(request, sb)
	sb.Reset()
	_, _ = fmt.Fprintf(
		sb,
		`SELECT
			id
			, record_status
			, user_id
			, origin_longitude
			, origin_latitude
			, destination_longitude
			, destination_latitude
			, distance_meters
			, distance_km
			, street_address
			, street_address_note
			, label
			, main
			, created_at
			, updated_at
			, deleted_at
			, deletion_reason
		FROM delivery_addresses %s`,
		condition,
	)
	var response models.DeliveryAddress
	err := q.dbRead.QueryRow(ctx, sb.String(), params...).Scan(
		&response.ID,
		&response.RecordStatus,
		&response.UserID,
		&response.OriginLongitude,
		&response.OriginLatitude,
		&response.DestinationLongitude,
		&response.DestinationLatitude,
		&response.DistanceMeters,
		&response.DistanceKM,
		&response.StreetAddress,
		&response.StreetAddressNote,
		&response.Label,
		&response.Main,
		&response.CreatedAt,
		&response.UpdatedAt,
		&response.DeletedAt,
		&response.DeletionReason,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, models.ErrDeliveryAddressNotFound
	}
	if err != nil {
		helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrScan")
		return nil, err
	}
	return &response, nil
}

func (q *deliveryAddressRepository) CreateDeliveryAddress(
	ctx context.Context,
	tx pgx.Tx,
	request *models.DeliveryAddress,
) (*models.DeliveryAddress, error) {
	ctxt := "DeliveryAddressRepository-CreateDeliveryAddress"
	var response models.DeliveryAddress
	if err := tx.QueryRow(
		ctx,
		`INSERT INTO delivery_addresses (
			record_status
			, user_id
			, origin_longitude
			, origin_latitude
			, destination_longitude
			, destination_latitude
			, distance_meters
			, distance_km
			, street_address
			, street_address_note
			, label
			, main
			, created_at
			, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
		RETURNING id
			record_status
			, user_id
			, origin_longitude
			, origin_latitude
			, destination_longitude
			, destination_latitude
			, distance_meters
			, distance_km
			, street_address
			, street_address_note
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
		request.OriginLongitude,
		request.OriginLatitude,
		request.DestinationLongitude,
		request.DestinationLatitude,
		request.DistanceMeters,
		request.DistanceKM,
		request.StreetAddress,
		request.StreetAddressNote,
		request.Label,
		request.Main,
		request.CreatedAt,
		request.UpdatedAt,
	).Scan(
		&response.ID,
		&response.RecordStatus,
		&response.UserID,
		&response.OriginLongitude,
		&response.OriginLatitude,
		&response.DestinationLongitude,
		&response.DestinationLatitude,
		&response.DistanceMeters,
		&response.DistanceKM,
		&response.StreetAddress,
		&response.StreetAddressNote,
		&response.Label,
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

func (q *deliveryAddressRepository) UpdateDeliveryAddress(
	ctx context.Context,
	tx pgx.Tx,
	request *models.DeliveryAddress,
) (*models.DeliveryAddress, error) {
	ctxt := "DeliveryAddressRepository-UpdateDeliveryAddress"
	var response models.DeliveryAddress
	err := tx.QueryRow(
		ctx,
		`SELECT
			id
			, record_status
			, user_id
			, origin_longitude
			, origin_latitude
			, destination_longitude
			, destination_latitude
			, distance_meters
			, distance_km
			, street_address
			, street_address_note
			, label
			, main
			, created_at
			, updated_at
			, deleted_at
			, deletion_reason
		FROM delivery_addresses
		WHERE id = $1
		FOR UPDATE`,
		request.ID,
	).Scan(
		&response.ID,
		&response.RecordStatus,
		&response.UserID,
		&response.OriginLongitude,
		&response.OriginLatitude,
		&response.DestinationLongitude,
		&response.DestinationLatitude,
		&response.DistanceMeters,
		&response.DistanceKM,
		&response.StreetAddress,
		&response.StreetAddressNote,
		&response.Label,
		&response.Main,
		&response.CreatedAt,
		&response.UpdatedAt,
		&response.DeletedAt,
		&response.DeletionReason,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, models.ErrDeliveryAddressNotFound
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
	_, _ = sb.WriteString("UPDATE delivery_addresses SET")
	if response.RecordStatus != request.RecordStatus {
		params = append(params, request.RecordStatus)
		if n = len(params); n > 1 {
			_, _ = sb.WriteString(",")
		}
		_, _ = fmt.Fprintf(sb, " record_status = $%d", n)
	}
	if response.DestinationLongitude != request.DestinationLongitude {
		params = append(params, request.DestinationLongitude)
		if n = len(params); n > 1 {
			_, _ = sb.WriteString(",")
		}
		_, _ = fmt.Fprintf(sb, " destination_longitude = $%d", n)
	}
	if response.DestinationLatitude != request.DestinationLatitude {
		params = append(params, request.DestinationLatitude)
		if n = len(params); n > 1 {
			_, _ = sb.WriteString(",")
		}
		_, _ = fmt.Fprintf(sb, " destination_latitude = $%d", n)
	}
	if response.DistanceMeters != request.DistanceMeters {
		params = append(params, request.DistanceMeters)
		if n = len(params); n > 1 {
			_, _ = sb.WriteString(",")
		}
		_, _ = fmt.Fprintf(sb, " distance_meters = $%d", n)
	}
	if response.DistanceKM != request.DistanceKM {
		params = append(params, request.DistanceKM)
		if n = len(params); n > 1 {
			_, _ = sb.WriteString(",")
		}
		_, _ = fmt.Fprintf(sb, " distance_km = $%d", n)
	}
	if response.StreetAddress != request.StreetAddress {
		params = append(params, request.StreetAddress)
		if n = len(params); n > 1 {
			_, _ = sb.WriteString(",")
		}
		_, _ = fmt.Fprintf(sb, " street_address = $%d", n)
	}
	if response.StreetAddressNote != request.StreetAddressNote {
		params = append(params, request.StreetAddressNote)
		if n = len(params); n > 1 {
			_, _ = sb.WriteString(",")
		}
		_, _ = fmt.Fprintf(sb, " street_address_note = $%d", n)
	}
	if response.Label != request.Label {
		params = append(params, request.Label)
		if n = len(params); n > 1 {
			_, _ = sb.WriteString(",")
		}
		_, _ = fmt.Fprintf(sb, " label = $%d", n)
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
			, origin_longitude
			, origin_latitude
			, destination_longitude
			, destination_latitude
			, distance_meters
			, distance_km
			, street_address
			, street_address_note
			, label
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
		&response.OriginLongitude,
		&response.OriginLatitude,
		&response.DestinationLongitude,
		&response.DestinationLatitude,
		&response.DistanceMeters,
		&response.DistanceKM,
		&response.StreetAddress,
		&response.StreetAddressNote,
		&response.Label,
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

func (q *deliveryAddressRepository) UnsetMain(
	ctx context.Context,
	tx pgx.Tx,
	userID uuid.UUID,
) error {
	ctxt := "DeliveryAddressRepository-UnsetMain"
	_, err := tx.Exec(
		ctx,
		`UPDATE delivery_addresses SET
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

func (q *deliveryAddressRepository) buildCondition(request *models.DeliveryAddressFilter, sb *strings.Builder) (condition string, params []any) {
	var conds []string
	if len(request.DeliveryAddressIDs) > 0 {
		sb.Reset()
		_, _ = sb.WriteString("id IN (")
		for i, deliveryAddressID := range request.DeliveryAddressIDs {
			params = append(params, deliveryAddressID)
			if i > 0 {
				_, _ = sb.WriteString(",")
			}
			_, _ = fmt.Fprintf(sb, "$%d", len(params))
		}
		_, _ = sb.WriteString(")")
		conds = append(conds, sb.String())
	}
	if len(request.ExcludedDeliveryAddressIDs) > 0 {
		sb.Reset()
		_, _ = sb.WriteString("id NOT IN (")
		for i, excludedDeliveryAddressID := range request.ExcludedDeliveryAddressIDs {
			params = append(params, excludedDeliveryAddressID)
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
	if len(request.OriginLongitudes) > 0 {
		sb.Reset()
		_, _ = sb.WriteString("origin_longitude IN (")
		for i, originLongitude := range request.OriginLongitudes {
			params = append(params, originLongitude)
			if i > 0 {
				_, _ = sb.WriteString(",")
			}
			_, _ = fmt.Fprintf(sb, "$%d", len(params))
		}
		_, _ = sb.WriteString(")")
		conds = append(conds, sb.String())
	}
	if len(request.OriginLatitudes) > 0 {
		sb.Reset()
		_, _ = sb.WriteString("origin_latitude IN (")
		for i, originLatitude := range request.OriginLatitudes {
			params = append(params, originLatitude)
			if i > 0 {
				_, _ = sb.WriteString(",")
			}
			_, _ = fmt.Fprintf(sb, "$%d", len(params))
		}
		_, _ = sb.WriteString(")")
		conds = append(conds, sb.String())
	}
	if len(request.DestinationLongitudes) > 0 {
		sb.Reset()
		_, _ = sb.WriteString("destination_longitude IN (")
		for i, destinationLongitude := range request.DestinationLongitudes {
			params = append(params, destinationLongitude)
			if i > 0 {
				_, _ = sb.WriteString(",")
			}
			_, _ = fmt.Fprintf(sb, "$%d", len(params))
		}
		_, _ = sb.WriteString(")")
		conds = append(conds, sb.String())
	}
	if len(request.DestinationLatitudes) > 0 {
		sb.Reset()
		_, _ = sb.WriteString("destination_latitude IN (")
		for i, destinationLatitude := range request.DestinationLatitudes {
			params = append(params, destinationLatitude)
			if i > 0 {
				_, _ = sb.WriteString(",")
			}
			_, _ = fmt.Fprintf(sb, "$%d", len(params))
		}
		_, _ = sb.WriteString(")")
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
