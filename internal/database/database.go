package database

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type DB struct{ Pool *pgxpool.Pool }

func Open(ctx context.Context, databaseURL string) (*DB, error) {
	config, err := pgxpool.ParseConfig(databaseURL)
	if err != nil { return nil, fmt.Errorf("parse database configuration: %w", err) }
	config.MaxConns = 20
	config.MinConns = 2
	config.MaxConnLifetime = time.Hour
	config.MaxConnIdleTime = 15 * time.Minute
	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil { return nil, fmt.Errorf("open database: %w", err) }
	if err := pool.Ping(ctx); err != nil { pool.Close(); return nil, fmt.Errorf("ping database: %w", err) }
	return &DB{Pool: pool}, nil
}

func (db *DB) Close() { db.Pool.Close() }
func (db *DB) Ping(ctx context.Context) error { return db.Pool.Ping(ctx) }

func (db *DB) InTx(ctx context.Context, fn func(pgx.Tx) error) error {
	tx, err := db.Pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil { return fmt.Errorf("begin transaction: %w", err) }
	defer func() { _ = tx.Rollback(ctx) }()
	if err := fn(tx); err != nil { return err }
	if err := tx.Commit(ctx); err != nil { return fmt.Errorf("commit transaction: %w", err) }
	return nil
}

