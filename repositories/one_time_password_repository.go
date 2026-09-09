package repositories

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"
	"uuid"

	"github.com/govalues/decimal"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/roysitumorang/laukpauk/helper"
	"github.com/roysitumorang/laukpauk/models"
	"github.com/roysitumorang/laukpauk/pools"
)

type OneTimePasswordRepository interface {
	FindOneTimePasswords(
		ctx context.Context,
		request *models.OneTimePasswordFilter,
	) ([]*models.OneTimePassword, int64, int64, error)
	FindOneTimePassword(
		ctx context.Context,
		request *models.OneTimePasswordFilter,
	) (*models.OneTimePassword, error)
	CreateOneTimePassword(
		ctx context.Context,
		tx pgx.Tx,
		request *models.OneTimePassword,
	) (*models.OneTimePassword, error)
	UpdateOneTimePassword(
		ctx context.Context,
		tx pgx.Tx,
		request *models.OneTimePassword,
	) (*models.OneTimePassword, error)
	AutoDeleteInvalidOneTimePasswords(
		ctx context.Context,
	) error
}

type oneTimePasswordRepository struct {
	dbRead,
	dbWrite *pgxpool.Pool
}

func NewOneTimePasswordRepository(
	dbRead,
	dbWrite *pgxpool.Pool,
) OneTimePasswordRepository {
	return &oneTimePasswordRepository{
		dbRead:  dbRead,
		dbWrite: dbWrite,
	}
}

func (q *oneTimePasswordRepository) FindOneTimePasswords(
	ctx context.Context,
	request *models.OneTimePasswordFilter,
) ([]*models.OneTimePassword, int64, int64, error) {
	ctxt := "OneTimePasswordRepository-FindOneTimePasswords"
	sb := pools.BuilderPool.Get()
	defer func() {
		sb.Reset()
		pools.BuilderPool.Put(sb)
	}()
	sb.Reset()
	condition, params := q.buildCondition(request, sb)
	sb.Reset()
	_, _ = fmt.Fprintf(sb, "SELECT COUNT(1) FROM one_time_passwords %s", condition)
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
		, seller_id
		, buyer_id
		, hash
		, tracking_id
		, attempts_count
		, valid
		, created_at
		, expired_at`,
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
	var response []*models.OneTimePassword
	for rows.Next() {
		var oneTimePassword models.OneTimePassword
		if err = rows.Scan(
			&oneTimePassword.ID,
			&oneTimePassword.SellerID,
			&oneTimePassword.BuyerID,
			&oneTimePassword.Hash,
			&oneTimePassword.TrackingID,
			&oneTimePassword.AttemptsCount,
			&oneTimePassword.Valid,
			&oneTimePassword.CreatedAt,
			&oneTimePassword.ExpiredAt,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrScan")
			return nil, 0, 0, err
		}
		offset++
		oneTimePassword.RowNo = offset
		response = append(response, &oneTimePassword)
	}
	return response, total, pages, nil
}

func (q *oneTimePasswordRepository) FindOneTimePassword(
	ctx context.Context,
	request *models.OneTimePasswordFilter,
) (*models.OneTimePassword, error) {
	ctxt := "OneTimePasswordRepository-FindOneTimePassword"
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
			, seller_id
			, buyer_id
			, hash
			, tracking_id
			, attempts_count
			, valid
			, created_at
			, expired_at
		FROM one_time_passwords %s`,
		condition,
	)
	var response models.OneTimePassword
	err := q.dbRead.QueryRow(ctx, sb.String(), params...).Scan(
		&response.ID,
		&response.SellerID,
		&response.BuyerID,
		&response.Hash,
		&response.TrackingID,
		&response.AttemptsCount,
		&response.Valid,
		&response.CreatedAt,
		&response.ExpiredAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, models.ErrOneTimePasswordNotFound
	}
	if err != nil {
		helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrScan")
		return nil, err
	}
	return &response, nil
}

func (q *oneTimePasswordRepository) CreateOneTimePassword(
	ctx context.Context,
	tx pgx.Tx,
	request *models.OneTimePassword,
) (*models.OneTimePassword, error) {
	ctxt := "OneTimePasswordRepository-CreateOneTimePassword"
	var response models.OneTimePassword
	var trackingID uuid.UUID
	var count int
	for {
		trackingID = uuid.NewV4()
		if err := tx.QueryRow(
			ctx,
			`SELECT COUNT(1)
			FROM one_time_passwords
			WHERE tracking_id = $1`,
			trackingID,
		).Scan(&count); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrScan")
			return nil, err
		}
		if count == 0 {
			break
		}
	}
	if err := tx.QueryRow(
		ctx,
		`INSERT INTO one_time_passwords (
			seller_id
			, buyer_id
			, hash
			, tracking_id
			, attempts_count
			, valid
			, created_at
			, expired_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id
			, seller_id
			, buyer_id
			, hash
			, tracking_id
			, attempts_count
			, valid
			, created_at
			, expired_at`,
		request.SellerID,
		request.BuyerID,
		request.Hash,
		trackingID,
		request.AttemptsCount,
		request.Valid,
		request.CreatedAt,
		request.ExpiredAt,
	).Scan(
		&response.ID,
		&response.SellerID,
		&response.BuyerID,
		&response.Hash,
		&response.TrackingID,
		&response.AttemptsCount,
		&response.Valid,
		&response.CreatedAt,
		&response.ExpiredAt,
	); err != nil {
		if pgxErr, ok := errors.AsType[*pgconn.PgError](err); ok &&
			pgxErr.Code == pgerrcode.UniqueViolation &&
			pgxErr.ConstraintName == "one_time_passwords_tracking_id_key" {
			err = models.ErrUniqueOneTimePasswordTrackingIdViolation
		} else {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrScan")
		}
		return nil, err
	}
	return &response, nil
}

func (q *oneTimePasswordRepository) UpdateOneTimePassword(
	ctx context.Context,
	tx pgx.Tx,
	request *models.OneTimePassword,
) (*models.OneTimePassword, error) {
	ctxt := "OneTimePasswordRepository-UpdateOneTimePassword"
	var response models.OneTimePassword
	err := tx.QueryRow(
		ctx,
		`SELECT
			id
			, seller_id
			, buyer_id
			, hash
			, tracking_id
			, attempts_count
			, valid
			, created_at
			, expired_at
		FROM one_time_passwords
		WHERE id = $1
		FOR UPDATE`,
		request.ID,
	).Scan(
		&response.ID,
		&response.SellerID,
		&response.BuyerID,
		&response.Hash,
		&response.TrackingID,
		&response.AttemptsCount,
		&response.Valid,
		&response.CreatedAt,
		&response.ExpiredAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, models.ErrOneTimePasswordNotFound
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
	sb.Reset()
	_, _ = sb.WriteString("UPDATE one_time_passwords SET")
	if response.AttemptsCount != request.AttemptsCount {
		params = append(params, request.AttemptsCount)
		if n = len(params); n > 1 {
			_, _ = sb.WriteString(",")
		}
		_, _ = fmt.Fprintf(sb, " attempts_count = $%d", n)
	}
	if response.Valid != request.Valid {
		params = append(params, request.Valid)
		if n = len(params); n > 1 {
			_, _ = sb.WriteString(",")
		}
		_, _ = fmt.Fprintf(sb, " valid = $%d", n)
	}
	if len(params) == 0 {
		return &response, nil
	}
	params = append(params, request.ID)
	_, _ = fmt.Fprintf(
		sb,
		` WHERE id = $%d
		RETURNING id
			, seller_id
			, buyer_id
			, hash
			, tracking_id
			, attempts_count
			, valid
			, created_at
			, expired_at`,
		len(params),
	)
	if err = tx.QueryRow(ctx, sb.String(), params...).Scan(
		&response.ID,
		&response.SellerID,
		&response.BuyerID,
		&response.Hash,
		&response.TrackingID,
		&response.AttemptsCount,
		&response.Valid,
		&response.CreatedAt,
		&response.ExpiredAt,
	); err != nil {
		helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrScan")
		return nil, err
	}
	return &response, nil
}

func (q *oneTimePasswordRepository) AutoDeleteInvalidOneTimePasswords(
	ctx context.Context,
) error {
	ctxt := "OneTimePasswordRepository-AutoDeleteInvalidOneTimePasswords"
	tx, err := BeginTx(ctx)
	if err != nil {
		helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
		return err
	}
	defer func(ctx context.Context) {
		if err = tx.Rollback(ctx); errors.Is(err, pgx.ErrTxClosed) {
			err = nil
		}
		if err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrRollback")
		}
	}(ctx)
	if _, err = tx.Exec(
		ctx,
		`DELETE
		FROM one_time_passwords
		WHERE NOT valid OR expired_at < $1`,
		time.Now(),
	); err != nil {
		helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
		return err
	}
	if err = tx.Commit(ctx); err != nil {
		helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrCommit")
	}
	return err
}

func (q *oneTimePasswordRepository) buildCondition(request *models.OneTimePasswordFilter, sb *strings.Builder) (condition string, params []any) {
	var conds []string
	if len(request.OneTimePasswordIDs) > 0 {
		sb.Reset()
		_, _ = sb.WriteString("id IN (")
		for i, oneTimePasswordID := range request.OneTimePasswordIDs {
			params = append(params, oneTimePasswordID)
			if i > 0 {
				_, _ = sb.WriteString(",")
			}
			_, _ = fmt.Fprintf(sb, "$%d", len(params))
		}
		_, _ = sb.WriteString(")")
		conds = append(conds, sb.String())
	}
	if len(request.SellerIDs) > 0 {
		sb.Reset()
		_, _ = sb.WriteString("seller_id IN (")
		for i, sellerID := range request.SellerIDs {
			params = append(params, sellerID)
			if i > 0 {
				_, _ = sb.WriteString(",")
			}
			_, _ = fmt.Fprintf(sb, "$%d", len(params))
		}
		_, _ = sb.WriteString(")")
		conds = append(conds, sb.String())
	}
	if len(request.BuyerIDs) > 0 {
		sb.Reset()
		_, _ = sb.WriteString("buyer_id IN (")
		for i, buyerID := range request.BuyerIDs {
			params = append(params, buyerID)
			if i > 0 {
				_, _ = sb.WriteString(",")
			}
			_, _ = fmt.Fprintf(sb, "$%d", len(params))
		}
		_, _ = sb.WriteString(")")
		conds = append(conds, sb.String())
	}
	if len(request.TrackingIDs) > 0 {
		sb.Reset()
		_, _ = sb.WriteString("tracking_id IN (")
		for i, trackingID := range request.TrackingIDs {
			params = append(params, trackingID)
			if i > 0 {
				_, _ = sb.WriteString(",")
			}
			_, _ = fmt.Fprintf(sb, "$%d", len(params))
		}
		_, _ = sb.WriteString(")")
		conds = append(conds, sb.String())
	}
	if request.Valid != nil {
		if *request.Valid {
			conds = append(conds, "valid")
		} else {
			conds = append(conds, "NOT valid")
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
