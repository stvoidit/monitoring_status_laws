package store

import (
	"context"
	"log/slog"
	"runtime"
	"strconv"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/tern/v2/migrate"
)

// DB - ...
type DB struct {
	pool *pgxpool.Pool
	m    *migrate.Migrator
}

type Logger struct {
}

func (l *Logger) TraceQueryStart(ctx context.Context, conn *pgx.Conn, data pgx.TraceQueryStartData) context.Context {
	var slogAttrs = make([]slog.Attr, len(data.Args))
	for i, v := range data.Args {
		if m, ok := v.(map[string]any); ok {
			for key, value := range m {
				slogAttrs[i] = slog.Any(key, value)
			}
		} else {
			slogAttrs[i] = slog.Any(strconv.Itoa(i), v)
		}
	}
	slog.DebugContext(ctx, "trace", slog.String("sql", data.SQL), slog.GroupAttrs("args", slogAttrs...))
	return ctx
}

func (l *Logger) TraceQueryEnd(ctx context.Context, conn *pgx.Conn, data pgx.TraceQueryEndData) {}

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
	poolConfig.ConnConfig.Tracer = &Logger{}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*15)
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

// func (db *DB) CheckMigrations(ctx context.Context) error {
// 	const versionTable = "monitoring_draft_laws."
// 	m, err := migrate.NewMigrator(ctx, conn, versionTable)
// 	if err != nil {
// 		return err
// 	}
// 	m.LoadMigrations()
// }
