package query

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/roysitumorang/laukpauk/helper"
	"github.com/roysitumorang/laukpauk/modules/region/model"
	"go.uber.org/zap"
)

type (
	regionQuery struct {
		dbRead, dbWrite *pgxpool.Pool
	}
)

func NewRegionQuery(
	dbRead,
	dbWrite *pgxpool.Pool,
) RegionQuery {
	return &regionQuery{
		dbRead:  dbRead,
		dbWrite: dbWrite,
	}
}

func (q *regionQuery) FindProvinces(ctx context.Context) (response []model.Region, err error) {
	ctxt := "RegionQuery-FindProvinces"
	response = []model.Region{}
	rows, err := q.dbRead.Query(
		ctx,
		`SELECT
			a.id
			, a.name
		FROM provinces a
		WHERE EXISTS(
			SELECT 1
			FROM cities b
			JOIN subdistricts c ON b.id = c.city_id
			JOIN villages d ON c.id = d.subdistrict_id
			JOIN coverage_area e ON d.id = e.village_id
			JOIN users f ON e.user_id = f.id
			WHERE b.province_id = a.id
		)
		ORDER BY a.name`,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		err = nil
	}
	if err != nil {
		helper.Capture(ctx, zap.ErrorLevel, err, ctxt, "ErrQuery")
		return
	}
	defer rows.Close()
	for rows.Next() {
		var province model.Region
		if err = rows.Scan(
			&province.ID,
			&province.Name,
		); err != nil {
			helper.Capture(ctx, zap.ErrorLevel, err, ctxt, "ErrScan")
			return
		}
		response = append(response, province)
	}
	if err = rows.Err(); err != nil {
		helper.Capture(ctx, zap.ErrorLevel, err, ctxt, "ErrErr")
	}
	return
}

func (q *regionQuery) FindCitiesByProvinceID(ctx context.Context, provinceID int64) (response []model.Region, err error) {
	ctxt := "RegionQuery-FindCitiesByProvinceID"
	response = []model.Region{}
	rows, err := q.dbRead.Query(
		ctx,
		`SELECT
			b.id
			, b.type || ' ' || b.name AS name
		FROM cities b
		WHERE EXISTS(
			SELECT 1
			FROM subdistricts c
			JOIN villages d ON c.id = d.subdistrict_id
			JOIN coverage_area e ON d.id = e.village_id
			JOIN users f ON e.user_id = f.id
			WHERE c.city_id = b.id
		)
		AND b.province_id = $1
		ORDER BY name`,
		provinceID,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		err = nil
	}
	if err != nil {
		helper.Capture(ctx, zap.ErrorLevel, err, ctxt, "ErrQuery")
		return
	}
	defer rows.Close()
	for rows.Next() {
		var city model.Region
		if err = rows.Scan(&city.ID, &city.Name); err != nil {
			helper.Capture(ctx, zap.ErrorLevel, err, ctxt, "ErrScan")
			return
		}
		response = append(response, city)
	}
	if err = rows.Err(); err != nil {
		helper.Capture(ctx, zap.ErrorLevel, err, ctxt, "ErrErr")
	}
	return
}

func (q *regionQuery) FindSubdistrictsByCityID(ctx context.Context, cityID int64) (response []model.Region, err error) {
	ctxt := "RegionQuery-FindSubdistrictsByCityID"
	response = []model.Region{}
	rows, err := q.dbRead.Query(
		ctx,
		`SELECT
			c.id
			, c.name
		FROM subdistricts c
		WHERE EXISTS(
			SELECT 1
			FROM villages d
			JOIN coverage_area e ON d.id = e.village_id
			JOIN users f ON e.user_id = f.id
			WHERE d.subdistrict_id = c.id
		)
		AND c.city_id = $1
		ORDER BY c.name`,
		cityID,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		err = nil
	}
	if err != nil {
		helper.Capture(ctx, zap.ErrorLevel, err, ctxt, "ErrQuery")
		return
	}
	defer rows.Close()
	for rows.Next() {
		var subdistrict model.Region
		if err = rows.Scan(&subdistrict.ID, &subdistrict.Name); err != nil {
			helper.Capture(ctx, zap.ErrorLevel, err, ctxt, "ErrScan")
			return
		}
		response = append(response, subdistrict)
	}
	if err = rows.Err(); err != nil {
		helper.Capture(ctx, zap.ErrorLevel, err, ctxt, "ErrErr")
	}
	return
}

func (q *regionQuery) FindVillagesBySubdistrictID(ctx context.Context, subdistrictID int64) (response []model.Region, err error) {
	ctxt := "RegionQuery-FindVillagesBySubdistrictID"
	response = []model.Region{}
	rows, err := q.dbRead.Query(
		ctx,
		`SELECT
			d.id
			, d.name
		FROM villages d
		WHERE EXISTS(
			SELECT 1
			FROM coverage_area e
			JOIN users f ON e.user_id = f.id
			WHERE e.village_id = d.id
		)
		AND d.subdistrict_id = $1
		ORDER BY d.name`,
		subdistrictID,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		err = nil
	}
	if err != nil {
		helper.Capture(ctx, zap.ErrorLevel, err, ctxt, "ErrQuery")
		return
	}
	defer rows.Close()
	for rows.Next() {
		var village model.Region
		if err = rows.Scan(&village.ID, &village.Name); err != nil {
			helper.Capture(ctx, zap.ErrorLevel, err, ctxt, "ErrScan")
			return
		}
		response = append(response, village)
	}
	if err = rows.Err(); err != nil {
		helper.Capture(ctx, zap.ErrorLevel, err, ctxt, "ErrErr")
	}
	return
}

func (q *regionQuery) FindVillageByID(ctx context.Context, villageID int64) (*model.Village, error) {
	ctxt := "RegionQuery-FindVillageByID"
	var response model.Village
	err := q.dbRead.QueryRow(
		ctx,
		`SELECT
			id
			, subdistrict_id
			, name
		FROM villages
		WHERE id = $1`,
		villageID,
	).Scan(
		&response.ID,
		&response.SubdistrictID,
		&response.Name,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		err = nil
	}
	if err != nil {
		helper.Capture(ctx, zap.ErrorLevel, err, ctxt, "ErrScan")
		return nil, err
	}
	return &response, nil
}
