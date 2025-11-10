package postgres

import (
	"baize-monitor/pkg/config"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
)

type PostgresClient struct {
	*sqlx.DB
}

func NewPostgresClient(cfg *config.PostGresConfig) (*PostgresClient, error) {
	dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.Database, cfg.SSLMode)

	db, err := sqlx.Connect("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to postgres: %w", err)
	}

	db.SetMaxOpenConns(cfg.MaxConns)
	db.SetMaxIdleConns(cfg.IdleConns)
	db.SetConnMaxLifetime(30 * time.Minute)

	return &PostgresClient{db}, nil
}

func (c *PostgresClient) Close() error {
	return c.DB.Close()
}
