package database

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/sammidev/goca/internal/pkg/logger"
)

type (
	txKey        struct{}
	traceDataKey struct{}
)

// queryTracer implements pgx.QueryTracer to log query execution details.
type queryTracer struct {
	logger logger.Logger
}

type traceData struct {
	startTime time.Time
	queryType string // e.g., SELECT, INSERT, UPDATE
	SQL       string
	Args      []any
}

// TraceQueryStart records the query start time and type in the context.
func (t *queryTracer) TraceQueryStart(ctx context.Context, _ *pgx.Conn, data pgx.TraceQueryStartData) context.Context {
	queryType := extractQueryType(data.SQL)
	return context.WithValue(ctx, traceDataKey{}, &traceData{
		startTime: time.Now(),
		queryType: queryType,
		SQL:       data.SQL,
		Args:      data.Args,
	})
}

// TraceQueryEnd logs query execution details upon completion.

func (t *queryTracer) TraceQueryEnd(ctx context.Context, _ *pgx.Conn, data pgx.TraceQueryEndData) {
	td, ok := ctx.Value(traceDataKey{}).(*traceData)
	if !ok {
		t.logger.Warn("Trace data not found in context; skipping query logging")
		return
	}

	duration := time.Since(td.startTime)
	logCtx := t.logger.WithContext(ctx)
	logFields := []interface{}{
		"query", td.SQL,
		"args", td.Args,
		"duration_ms", duration.Milliseconds(),
		"query_type", td.queryType,
	}

	if data.Err != nil {
		if errors.Is(data.Err, pgx.ErrNoRows) {
			logCtx.Debug("Query returned no rows", logFields...)
		} else {
			logFields = append(logFields, "error", data.Err.Error())
			logCtx.Error("Query failed", logFields...)
		}
	} else {
		logFields = append(logFields, "rows_affected", data.CommandTag.RowsAffected())
		logCtx.Debug("Query executed successfully", logFields...)
	}
}

// extractQueryType determines the type of SQL query (e.g., SELECT, INSERT).
func extractQueryType(sql string) string {
	if len(sql) == 0 {
		return "UNKNOWN"
	}

	// Simple heuristic: take the first word of the query
	var firstWord string
	_, _ = fmt.Sscanf(sql, "%s", &firstWord)
	switch firstWord {
	case "SELECT", "INSERT", "UPDATE", "DELETE":
		return firstWord
	default:
		return "OTHER"
	}
}
