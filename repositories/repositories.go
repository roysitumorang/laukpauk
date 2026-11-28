package repositories

import (
	"context"
	"log/slog"

	"github.com/roysitumorang/laukpauk/database"
	"github.com/roysitumorang/laukpauk/helper"
)

var dbWrite *database.Pool

func SetDbWrite(db *database.Pool) {
	dbWrite = db
}

func BeginTx(ctx context.Context) (database.Tx, error) {
	ctxt := "Repositories-BeginTx"
	tx, err := dbWrite.Begin(ctx)
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
	if _, ok := naturalSortColumns[field]; ok {
		return "NATURAL_SORT_KEY(" + field + ")"
	}
	return field
}
