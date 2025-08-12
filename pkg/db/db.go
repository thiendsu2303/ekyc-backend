package db

import (
	"context"
	"fmt"
	"time"

	"github.com/ekyc-backend/pkg/config"
	"github.com/ekyc-backend/pkg/logger"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type DB struct {
	pool   *pgxpool.Pool
	logger *logger.Logger
}

func New(cfg *config.Config, logger *logger.Logger) (*DB, error) {
	config, err := pgxpool.ParseConfig(cfg.GetDBConnString())
	if err != nil {
		return nil, fmt.Errorf("failed to parse database config: %w", err)
	}

	// Set connection pool settings
	config.MaxConns = 10
	config.MinConns = 2
	config.MaxConnLifetime = time.Hour
	config.MaxConnIdleTime = 30 * time.Minute

	// Create connection pool
	pool, err := pgxpool.NewWithConfig(context.Background(), config)
	if err != nil {
		return nil, fmt.Errorf("failed to create connection pool: %w", err)
	}

	// Test connection
	if err := pool.Ping(context.Background()); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	logger.Info("Database connection established")

	return &DB{
		pool:   pool,
		logger: logger,
	}, nil
}

func (db *DB) Close() {
	if db.pool != nil {
		db.pool.Close()
	}
}

func (db *DB) GetPool() *pgxpool.Pool {
	return db.pool
}

func (db *DB) Exec(ctx context.Context, sql string, arguments ...interface{}) error {
	_, err := db.pool.Exec(ctx, sql, arguments...)
	return err
}

func (db *DB) QueryRow(ctx context.Context, sql string, arguments ...interface{}) pgx.Row {
	return db.pool.QueryRow(ctx, sql, arguments...)
}

func (db *DB) Query(ctx context.Context, sql string, arguments ...interface{}) (pgx.Rows, error) {
	return db.pool.Query(ctx, sql, arguments...)
}

func (db *DB) Begin(ctx context.Context) (pgx.Tx, error) {
	return db.pool.Begin(ctx)
}

func (db *DB) BeginTx(ctx context.Context, txOptions pgx.TxOptions) (pgx.Tx, error) {
	return db.pool.BeginTx(ctx, txOptions)
}
