package storage

import (
	"context"
	"fmt"
	"sync/atomic"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	log "github.com/sirupsen/logrus"
)

type DB struct {
	pool   *pgxpool.Pool
	status atomic.Bool
}

type Config struct {
	Host     string
	Port     int
	Database string
	Username string
	Password string
	SSLMode  string
}

func InitDB(ctx context.Context, config Config) (*DB, error) {
	dsn := fmt.Sprintf(
		"postgres://%s:%s@%s:%d/%s?sslmode=%s",
		config.Username,
		config.Password,
		config.Host,
		config.Port,
		config.Database,
		config.SSLMode,
	)

	poolConfig, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to parse database config: %w", err)
	}

	poolConfig.MaxConns = 25
	poolConfig.MinConns = 5
	poolConfig.MaxConnLifetime = 5 * time.Minute
	poolConfig.MaxConnIdleTime = 30 * time.Second

	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create connection pool: %w", err)
	}

	db := &DB{
		pool:   pool,
		status: atomic.Bool{},
	}

	return db, nil
}

func (db *DB) Start(ctx context.Context) error {
	if err := db.pool.Ping(ctx); err != nil {
		return fmt.Errorf("failed to ping database: %w", err)
	}

	db.status.Store(true)
	log.Info("Database connection established")
	return nil
}

func (db *DB) Close() {
	if db.pool != nil {
		db.pool.Close()
	}
	db.status.Store(false)
	log.Info("Database connection closed")
}

func (db *DB) IsRunning() bool {
	return db.status.Load()
}

func (db *DB) GetPool() *pgxpool.Pool {
	return db.pool
}