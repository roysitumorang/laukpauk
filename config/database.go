package config

import (
	"context"
	"database/sql"
	"time"

	"github.com/go-sql-driver/mysql"
)

// sqlMode makes invalid or truncated values errors instead of warnings.
const sqlMode = "'STRICT_ALL_TABLES,NO_ZERO_IN_DATE,NO_ZERO_DATE,ERROR_FOR_DIVISION_BY_ZERO,NO_ENGINE_SUBSTITUTION'"

// CreateDbConnection connects to MariaDB with at most maxConns open connections and checks the connection.
// dataSourceName is a go-sql-driver/mysql DSN, e.g. user:password@tcp(localhost:3306)/laukpauk. Times are parsed and
// kept in UTC on the server.
func CreateDbConnection(ctx context.Context, dataSourceName string, maxConns int32) (*sql.DB, error) {
	cfg, err := mysql.ParseDSN(dataSourceName)
	if err != nil {
		return nil, err
	}
	cfg.ParseTime = true
	cfg.Loc = time.UTC
	if cfg.Params == nil {
		cfg.Params = map[string]string{}
	}
	cfg.Params["time_zone"] = "'+00:00'"
	cfg.Params["sql_mode"] = sqlMode
	connector, err := mysql.NewConnector(cfg)
	if err != nil {
		return nil, err
	}
	db := sql.OpenDB(connector)
	db.SetMaxOpenConns(int(maxConns))
	db.SetMaxIdleConns(int(maxConns))
	db.SetConnMaxIdleTime(5 * time.Minute)
	if err = db.PingContext(ctx); err != nil {
		_ = db.Close()
		return nil, err
	}
	return db, nil
}
