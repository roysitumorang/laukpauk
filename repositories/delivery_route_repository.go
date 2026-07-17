package repositories

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"

	"github.com/govalues/decimal"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/roysitumorang/laukpauk/helper"
	"github.com/roysitumorang/laukpauk/models"
)

type (
	DeliveryRouteRepository interface {
		FindDeliveryRoutes(
			ctx context.Context,
			request *models.DeliveryRouteFilter,
		) ([]*models.DeliveryRoute, int64, int64, error)
		FindDeliveryRoute(
			ctx context.Context,
			request *models.DeliveryRouteFilter,
		) (*models.DeliveryRoute, error)
		SaveDeliveryRoute(
			ctx context.Context,
			tx pgx.Tx,
			request *models.DeliveryRoute,
		) (*models.DeliveryRoute, error)
	}

	deliveryRouteRepository struct {
		dbRead,
		dbWrite *pgxpool.Pool
	}
)

func NewDeliveryRouteRepository(
	dbRead,
	dbWrite *pgxpool.Pool,
) DeliveryRouteRepository {
	return &deliveryRouteRepository{
		dbRead:  dbRead,
		dbWrite: dbWrite,
	}
}

func (q *deliveryRouteRepository) FindDeliveryRoutes(
	ctx context.Context,
	request *models.DeliveryRouteFilter,
) ([]*models.DeliveryRoute, int64, int64, error) {
	ctxt := "DeliveryRouteRepository-FindDeliveryRoutes"
	var sb strings.Builder
	condition, params := q.buildCondition(request, &sb)
	sb.Reset()
	_, _ = fmt.Fprintf(&sb, "SELECT COUNT(1) FROM delivery_routes %s", condition)
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
		, origin_longitude
		, origin_latitude
		, destination_longitude
		, destination_latitude
		, distance_meters
		, distance_km
		, created_at
		, updated_at`,
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
		_, _ = fmt.Fprintf(&sb, " %s %s", item.Field, item.Direction)
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
		_, _ = fmt.Fprintf(&sb, " LIMIT %d OFFSET %d", request.Limit, offset)
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
	var response []*models.DeliveryRoute
	for rows.Next() {
		var deliveryRoute models.DeliveryRoute
		if err = rows.Scan(
			&deliveryRoute.ID,
			&deliveryRoute.OriginLongitude,
			&deliveryRoute.OriginLatitude,
			&deliveryRoute.DestinationLongitude,
			&deliveryRoute.DestinationLatitude,
			&deliveryRoute.DistanceMeters,
			&deliveryRoute.DistanceKM,
			&deliveryRoute.CreatedAt,
			&deliveryRoute.UpdatedAt,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrScan")
			return nil, 0, 0, err
		}
		offset++
		deliveryRoute.RowNo = offset
		response = append(response, &deliveryRoute)
	}
	return response, total, pages, nil
}

func (q *deliveryRouteRepository) FindDeliveryRoute(
	ctx context.Context,
	request *models.DeliveryRouteFilter,
) (*models.DeliveryRoute, error) {
	ctxt := "DeliveryRouteRepository-FindDeliveryRoute"
	var sb strings.Builder
	condition, params := q.buildCondition(request, &sb)
	sb.Reset()
	_, _ = fmt.Fprintf(
		&sb,
		`SELECT
			id
			, origin_longitude
			, origin_latitude
			, destination_longitude
			, destination_latitude
			, distance_meters
			, distance_km
			, created_at
			, updated_at
		FROM delivery_routes %s`,
		condition,
	)
	query := sb.String()
	var response models.DeliveryRoute
	err := q.dbRead.QueryRow(ctx, query, params...).Scan(
		&response.ID,
		&response.OriginLongitude,
		&response.OriginLatitude,
		&response.DestinationLongitude,
		&response.DestinationLatitude,
		&response.DistanceMeters,
		&response.DistanceKM,
		&response.CreatedAt,
		&response.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, models.ErrDeliveryRouteNotFound
	}
	if err != nil {
		helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrScan")
		return nil, err
	}
	return &response, nil
}

func (q *deliveryRouteRepository) SaveDeliveryRoute(
	ctx context.Context,
	tx pgx.Tx,
	request *models.DeliveryRoute,
) (*models.DeliveryRoute, error) {
	ctxt := "DeliveryRouteRepository-SaveDeliveryRoute"
	var response models.DeliveryRoute
	if err := tx.QueryRow(
		ctx,
		`INSERT INTO delivery_routes (
			origin_longitude
			, origin_latitude
			, destination_longitude
			, destination_latitude
			, distance_meters
			, distance_km
			, created_at
			, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		ON CONFLICT (
			origin_longitude
			, origin_latitude
			, destination_longitude
			, destination_latitude
		)
		DO UPDATE SET
			distance_meters = $5
			, distance_km = $6
			, updated_at = $8
		RETURNING id
			, origin_longitude
			, origin_latitude
			, destination_longitude
			, destination_latitude
			, distance_meters
			, distance_km
			, created_at
			, updated_at`,
		request.OriginLongitude,
		request.OriginLatitude,
		request.DestinationLongitude,
		request.DestinationLatitude,
		request.DistanceMeters,
		request.DistanceKM,
		request.CreatedAt,
		request.UpdatedAt,
	).Scan(
		&response.ID,
		&response.OriginLongitude,
		&response.OriginLatitude,
		&response.DestinationLongitude,
		&response.DestinationLatitude,
		&response.DistanceMeters,
		&response.DistanceKM,
		&response.CreatedAt,
		&response.UpdatedAt,
	); err != nil {
		helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrScan")
		return nil, err
	}
	return &response, nil
}

func (q *deliveryRouteRepository) buildCondition(request *models.DeliveryRouteFilter, sb *strings.Builder) (condition string, params []any) {
	var conds []string
	if len(request.DeliveryRouteIDs) > 0 {
		sb.Reset()
		_, _ = sb.WriteString("id IN (")
		for i, deliveryRouteID := range request.DeliveryRouteIDs {
			params = append(params, deliveryRouteID)
			if i > 0 {
				_, _ = sb.WriteString(",")
			}
			_, _ = fmt.Fprintf(sb, "$%d", len(params))
		}
		_, _ = sb.WriteString(")")
		conds = append(conds, sb.String())
	}
	if request.OriginLongitude != nil {
		sb.Reset()
		params = append(params, request.OriginLongitude)
		_, _ = fmt.Fprintf(sb, "origin_longitude = $%d", len(params))
	}
	if request.OriginLatitude != nil {
		sb.Reset()
		params = append(params, request.OriginLatitude)
		_, _ = fmt.Fprintf(sb, "origin_latitude = $%d", len(params))
	}
	if request.DestinationLongitude != nil {
		sb.Reset()
		params = append(params, request.DestinationLongitude)
		_, _ = fmt.Fprintf(sb, "destination_longitude = $%d", len(params))
	}
	if request.DestinationLatitude != nil {
		sb.Reset()
		params = append(params, request.DestinationLatitude)
		_, _ = fmt.Fprintf(sb, "destination_latitude = $%d", len(params))
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
