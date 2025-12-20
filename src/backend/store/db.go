package store

import (
	"context"
	"runtime"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// DB - ...
type DB struct {
	pool *pgxpool.Pool
}

// NewDB - ...
func NewDB(connString string) (*DB, error) {
	poolConfig, err := pgxpool.ParseConfig(connString)
	if err != nil {
		return nil, err
	}
	poolConfig.MinConns = 0
	poolConfig.MaxConns = int32(runtime.NumCPU())
	poolConfig.MaxConnLifetime = time.Minute * 10
	poolConfig.MaxConnIdleTime = time.Minute * 5
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		return nil, err
	}
	if err := pool.Ping(ctx); err != nil {
		return nil, err
	}
	return &DB{pool: pool}, nil
}

// Close - закрытие соединений
func (db *DB) Close() { db.pool.Close() }
