package database

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/sammidev/goca/internal/config"
	"github.com/sammidev/goca/internal/pkg/logger"
)

// Database is the high-level abstraction over any database (SQL/NoSQL).
type Database interface {
	GetSQLExecutor(ctx context.Context) (SQLExecutor, error)
	WithTransaction(ctx context.Context, fn UnitOfWorkFunc) error
	Close()
}

// Executor is a marker interface.
// Concrete executors: SQLExecutor (pgx) or MongoExecutor (*mongo.Database).
type Executor interface{}

// UnitOfWorkFunc defines a function executed within a transaction/session.
type UnitOfWorkFunc func(ctx context.Context) error

// SQLExecutor defines query execution for SQL databases.
type SQLExecutor interface {
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
	Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error)
}

func maskDSN(dsn string) string {
	const mask = "****"
	if len(dsn) < 10 {
		return mask
	}
	return dsn[:5] + mask + dsn[len(dsn)-5:]
}

func NewDatabase(cfg *config.Config, log logger.Logger) (Database, error) {
	switch cfg.DatabaseDriver {
	case "postgres":
		return NewPostgreSQLDatabase(cfg, log)
	case "mongodb":
		// Placeholder for MongoDB implementation
		return nil, fmt.Errorf("MongoDB not implemented yet")
	default:
		return nil, fmt.Errorf("unsupported database driver: %s", cfg.DatabaseDriver)
	}
}
