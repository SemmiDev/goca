package observability

import (
	"context"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

// TraceOperator executes a function with OpenTelemetry tracing, supporting any return type
func TraceOperation[T any](ctx context.Context, tracer trace.Tracer, operation string, fn func(context.Context) (T, error), attrs ...attribute.KeyValue) (T, error) {
	ctx, span := tracer.Start(ctx, operation)
	defer span.End()

	if len(attrs) > 0 {
		span.SetAttributes(attrs...)
	}

	result, err := fn(ctx)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
	}

	return result, err
}
