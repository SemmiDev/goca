package database

import (
	"context"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sammidev/goca/internal/config"
	"github.com/sammidev/goca/internal/pkg/logger"
)

// PoolManager mendefinisikan antarmuka untuk connection pool PostgreSQL.
// Hal ini memungkinkan untuk melakukan mocking pada pool saat pengujian.
type PoolManager interface {
	SQLExecutor
	Begin(ctx context.Context) (pgx.Tx, error)
	Close()
}

// PostgreSQLDatabase mengimplementasikan Database untuk Postgres.
type PostgreSQLDatabase struct {
	pool   PoolManager
	logger logger.Logger
}

var _ Database = (*PostgreSQLDatabase)(nil)

// NewPostgreSQLDatabase menginisialisasi connection pool Postgres.
func NewPostgreSQLDatabase(cfg *config.Config, log logger.Logger) (*PostgreSQLDatabase, error) {
	if cfg.DSN() == "" {
		return nil, fmt.Errorf("database DSN is empty")
	}

	poolConfig, err := pgxpool.ParseConfig(cfg.DSN())
	if err != nil {
		return nil, fmt.Errorf("failed to parse database DSN: %w", err)
	}

	poolConfig.ConnConfig.Tracer = &queryTracer{logger: log.WithComponent("database.tracer")}
	poolConfig.MaxConns = int32(cfg.DatabaseMaxOpenConns)
	poolConfig.MinConns = int32(cfg.DatabaseMaxIdleConns)
	poolConfig.MaxConnLifetime = cfg.DatabaseMaxLifetime
	poolConfig.MaxConnIdleTime = cfg.DatabaseMaxIdleTime

	pool, err := pgxpool.NewWithConfig(context.Background(), poolConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create connection pool: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), cfg.DatabaseMaxIdleTime)
	defer cancel()

	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	log.Info("Postgres connection established",
		"dsn", maskDSN(cfg.DSN()),
		"max_conns", poolConfig.MaxConns,
		"min_conns", poolConfig.MinConns,
	)

	return &PostgreSQLDatabase{
		pool:   pool,
		logger: log.WithComponent("database"),
	}, nil
}

func (db *PostgreSQLDatabase) Close() {
	if db.pool != nil {
		db.pool.Close()
		db.logger.Info("Postgres connection pool closed")
	}
}

func (db *PostgreSQLDatabase) WithTransaction(ctx context.Context, fn UnitOfWorkFunc) error {
	tx, err := db.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	defer func() {
		if p := recover(); p != nil {
			_ = tx.Rollback(ctx)
			panic(p)
		}
	}()

	txCtx := context.WithValue(ctx, txKey{}, tx)
	if err := fn(txCtx); err != nil {
		_ = tx.Rollback(ctx)
		return err
	}
	return tx.Commit(ctx)
}

func (db *PostgreSQLDatabase) GetSQLExecutor(ctx context.Context) (SQLExecutor, error) {
	if tx, ok := ctx.Value(txKey{}).(SQLExecutor); ok {
		return tx, nil
	}
	return db.pool, nil
}

func IsUniqueViolation(err error) bool {
	return err != nil && (strings.Contains(err.Error(), "unique_violation") ||
		strings.Contains(err.Error(), "duplicate key value violates unique constraint"))
}
