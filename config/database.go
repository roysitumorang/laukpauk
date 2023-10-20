package config

import (
	"context"
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	dataSourceName = "host=%s user=%s password=%s dbname=%s %s"
)

func GetDbWriteOnly() *pgxpool.Pool {
	return createDbConnection(
		os.Getenv("DB_WRITE_HOST"),
		os.Getenv("DB_WRITE_USERNAME"),
		os.Getenv("DB_WRITE_PASSWORD"),
		os.Getenv("DB_WRITE_NAME"),
		os.Getenv("DB_WRITE_PARAM"),
	)
}

func GetDbReadOnly() *pgxpool.Pool {
	return createDbConnection(
		os.Getenv("DB_READ_HOST"),
		os.Getenv("DB_READ_USERNAME"),
		os.Getenv("DB_READ_PASSWORD"),
		os.Getenv("DB_READ_NAME"),
		os.Getenv("DB_READ_PARAM"),
	)
}

func createDbConnection(host, user, password, dbName, param string) *pgxpool.Pool {
	descriptor := fmt.Sprintf(dataSourceName, host, user, password, dbName, param)
	envMaxConns, ok := os.LookupEnv("DB_MAX_CONNECTIONS")
	if !ok {
		log.Fatalln("db: env DB_MAX_CONNECTIONS is not present")
	}
	maxConns, _ := strconv.Atoi(envMaxConns)
	if maxConns < 1 {
		log.Fatalln("db: env DB_MAX_CONNECTIONS value should be an integer greater than 0")
	}
	config, err := pgxpool.ParseConfig(descriptor)
	if err != nil {
		log.Fatalln(err.Error())
	}
	config.MaxConns = int32(maxConns)
	ctx := context.Background()
	db, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		log.Fatalln(err)
	}
	if err = db.Ping(ctx); err != nil {
		log.Fatalln(err)
	}
	return db
}
