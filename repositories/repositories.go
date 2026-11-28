package repositories

import (
	"context"
	"database/sql"
	"log/slog"

	"github.com/roysitumorang/laukpauk/helper"
	"github.com/roysitumorang/laukpauk/pools"
)

var dbWrite *sql.DB

func SetDbWrite(db *sql.DB) {
	dbWrite = db
}

func BeginTx(ctx context.Context) (*sql.Tx, error) {
	ctxt := "Repositories-BeginTx"
	tx, err := dbWrite.BeginTx(ctx, nil)
	if err != nil {
		helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrBegin")
	}
	return tx, err
}

// naturalSortColumns are the text columns PostgreSQL sorted with its natural_sort collation ("Item 2" before
// "Item 10"); MariaDB has no such collation, so ORDER BY sorts them by NATURAL_SORT_KEY.
var naturalSortColumns = map[string]struct{}{
	"name":                  {},
	"product_name":          {},
	"product_category_name": {},
	"stock_unit":            {},
}

// sortExpression is what ORDER BY sorts a sorting field by.
func sortExpression(field string) string {
	if _, ok := naturalSortColumns[field]; !ok {
		return field
	}
	sb := pools.BuilderPool.Get()
	defer func() {
		sb.Reset()
		pools.BuilderPool.Put(sb)
	}()
	sb.Reset()
	_, _ = sb.WriteString("NATURAL_SORT_KEY(")
	_, _ = sb.WriteString(field)
	_, _ = sb.WriteString(")")
	return sb.String()
}
