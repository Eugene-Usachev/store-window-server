package postgres_pool

import (
	"context"
	"fmt"

	"platform/logger"

	"github.com/caarlos0/env/v11"
	"github.com/jackc/pgx/v5/pgxpool"
)

type postgresConfig struct {
	Host     string `env:"POSTGRES_HOST, required, notEmpty"`
	Port     string `env:"POSTGRES_PORT, required, notEmpty"`
	UserName string `env:"POSTGRES_USERNAME, required, notEmpty"`
	UserPass string `env:"POSTGRES_PASSWORD, required, notEmpty"`
	DBName   string `env:"POSTGRES_DBNAME, required, notEmpty"`
	SSLMode  string `env:"POSTGRES_SSLMODE, required, notEmpty"`
}

func MustCreatePostgresPool() *pgxpool.Pool {
	cfg, err := env.ParseAs[postgresConfig]()
	if err != nil {
		logger.Fatal(err.Error())

		return nil
	}

	var config *pgxpool.Config

	url := fmt.Sprintf(
		"postgresql://%s:%s@%s:%s/%s",
		cfg.UserName,
		cfg.UserPass,
		cfg.Host,
		cfg.Port,
		cfg.DBName,
	)
	config, err = pgxpool.ParseConfig(url)
	if err != nil {
		logger.Fatal(fmt.Sprintf("failed to parse the config of pgxpool, %v", err))

		return nil
	}

	config.ConnConfig.Tracer = newPostgresLogger(logger.MainLogger())

	pool, err := pgxpool.NewWithConfig(context.Background(), config)
	if err != nil {
		logger.Fatal("failed to create pgx Pool")

		return nil
	}

	return pool
}
