package config

import (
	"context"

	"github.com/roysitumorang/laukpauk/database"
)

// CreateDbConnection connects to MariaDB. dataSourceName is a go-sql-driver/mysql DSN, e.g.
// user:password@tcp(localhost:3306)/laukpauk.
func CreateDbConnection(ctx context.Context, dataSourceName string, maxConns int32) (*database.Pool, error) {
	return database.Open(ctx, dataSourceName, int(maxConns))
}
