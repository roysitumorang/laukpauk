package config

import (
	"context"
	"errors"
	"os"
	"strconv"

	"github.com/jackc/pgx/v5/pgxpool"
)

func GetDbWriteOnly(ctx context.Context) (*pgxpool.Pool, error) {
	return createDbConnection(
		ctx,
		os.Getenv("DB_WRITE_DSN"),
	)
}

func GetDbReadOnly(ctx context.Context) (*pgxpool.Pool, error) {
	return createDbConnection(
		ctx,
		os.Getenv("DB_READ_DSN"),
	)
}

func createDbConnection(ctx context.Context, dataSourceName string) (*pgxpool.Pool, error) {
	envMaxConns, ok := os.LookupEnv("DB_MAX_CONNECTIONS")
	if !ok || envMaxConns == "" {
		return nil, errors.New("db: env DB_MAX_CONNECTIONS is required")
	}
	maxConns, err := strconv.Atoi(envMaxConns)
	if err != nil {
		return nil, err
	}
	if maxConns < 1 {
		return nil, errors.New("db: env DB_MAX_CONNECTIONS requires a positive integer")
	}
	config, err := pgxpool.ParseConfig(dataSourceName)
	if err != nil {
		return nil, err
	}
	config.MaxConns = int32(maxConns)
	db, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return nil, err
	}
	if err = db.Ping(ctx); err != nil {
		return nil, err
	}
	return db, nil
}
