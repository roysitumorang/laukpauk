package migration

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
)

func init() {
	Migrations[1698302382225052479] = func(ctx context.Context, tx pgx.Tx) (err error) {
		if _, err = tx.Exec(
			ctx,
			`UPDATE banners SET
				thumbnails = NULL
			WHERE thumbnails IN ('null', '[]');`,
		); err != nil {
			return
		}
		if _, err = tx.Exec(
			ctx,
			`UPDATE banners SET
				updated_by = created_by
				, updated_at = created_at
			WHERE updated_by IS NULL
			OR updated_at IS NULL;`,
		); err != nil {
			return
		}
		if _, err = tx.Exec(
			ctx,
			`ALTER TABLE banners
				ALTER COLUMN published TYPE bool USING published::int::bool
				, ALTER COLUMN file TYPE char varying USING file::char varying
				, ALTER file SET NOT NULL
				, ALTER created_by SET NOT NULL
				, ALTER COLUMN created_at TYPE timestamp with time zone USING created_at::timestamp with time zone
				, ALTER created_at SET NOT NULL
				, ALTER updated_by SET NOT NULL
				, ALTER COLUMN updated_at TYPE timestamp with time zone USING updated_at::timestamp with time zone
				, ALTER updated_at SET NOT NULL
				, ADD parent_id bigint REFERENCES banners (id) ON UPDATE RESTRICT ON DELETE RESTRICT;`,
		); err != nil {
			return
		}
		rows, err := tx.Query(
			ctx,
			`SELECT
				id
				, thumbnails
			FROM banners
			WHERE thumbnails IS NOT NULL`,
		)
		if errors.Is(err, pgx.ErrNoRows) {
			err = nil
		}
		if err != nil {
			return
		}
		defer rows.Close()
		mapThumbnails := map[int64][]string{}
		for rows.Next() {
			var (
				bannerID   int64
				thumbnails string
			)
			if err = rows.Scan(&bannerID, &thumbnails); err != nil {
				return
			}
			mapThumbnails[bannerID] = strings.Split(thumbnails, ",")
		}
		now := time.Now().UTC()
		for bannerID, thumbnails := range mapThumbnails {
			var (
				params       = []interface{}{true, 1, now}
				placeholders []string
			)
			for _, thumbnail := range thumbnails {
				params = append(params, bannerID, thumbnail)
				n := len(params)
				placeholders = append(placeholders, fmt.Sprintf("($%d, $%d, $1, $2, $3, $2, $3)", n-1, n))
			}
			if _, err = tx.Exec(
				ctx,
				fmt.Sprintf(
					`INSERT INTO banners (
						parent_id
						, file
						, published
						, created_by
						, created_at
						, updated_by
						, updated_at
					) VALUES %s`,
					strings.Join(placeholders, ","),
				),
				params...,
			); err != nil {
				return
			}
		}
		_, err = tx.Exec(
			ctx,
			`ALTER TABLE banners
				DROP COLUMN thumbnails;`,
		)
		return
	}
}
