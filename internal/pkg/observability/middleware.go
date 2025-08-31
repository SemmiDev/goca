package observability

import (
	"context"
	"errors"
	"runtime"
	"time"

	"github.com/gofiber/fiber/v2"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/propagation"
	semconv "go.opentelemetry.io/otel/semconv/v1.24.0"
	"go.opentelemetry.io/otel/trace"
)

// Observability holds all instruments and configs.
type Observability struct {
	RequestCounter   metric.Int64Counter
	RequestDuration  metric.Float64Histogram
	RequestsInFlight metric.Int64UpDownCounter
}

// NewObservability creates and registers all metrics.
func NewObservability(appName string) (*Observability, error) {
	meter := otel.GetMeterProvider().Meter(appName)
	var errs error

	reqCounter, err := meter.Int64Counter(
		"http_requests_total",
		metric.WithDescription("Total number of HTTP requests."),
		metric.WithUnit("{request}"),
	)
	if err != nil {
		errs = errors.Join(errs, err)
	}

	reqDuration, err := meter.Float64Histogram(
		"http_request_duration_seconds",
		metric.WithDescription("Latency of HTTP requests."),
		metric.WithUnit("s"),
	)
	if err != nil {
		errs = errors.Join(errs, err)
	}

	reqInFlight, err := meter.Int64UpDownCounter(
		"http_requests_in_flight",
		metric.WithDescription("Number of active HTTP requests."),
		metric.WithUnit("{request}"),
	)
	if err != nil {
		errs = errors.Join(errs, err)
	}

	_, err = meter.Int64ObservableGauge(
		"go_goroutines",
		metric.WithDescription("Number of running goroutines."),
		metric.WithUnit("1"),
		metric.WithInt64Callback(func(_ context.Context, obs metric.Int64Observer) error {
			obs.Observe(int64(runtime.NumGoroutine()))
			return nil
		}),
	)
	if err != nil {
		errs = errors.Join(errs, err)
	}

	if errs != nil {
		return nil, errs
	}

	return &Observability{
		RequestCounter:   reqCounter,
		RequestDuration:  reqDuration,
		RequestsInFlight: reqInFlight,
	}, nil
}

// MetricsMiddleware returns a Fiber handler for metrics.
func (o *Observability) MetricsMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		start := time.Now()

		requestAttrs := attribute.NewSet(
			semconv.HTTPRouteKey.String(c.Route().Path), // Gunakan route untuk kardinalitas rendah
			semconv.HTTPMethodKey.String(c.Method()),
		)

		o.RequestsInFlight.Add(c.Context(), 1, metric.WithAttributeSet(requestAttrs))
		defer o.RequestsInFlight.Add(c.Context(), -1, metric.WithAttributeSet(requestAttrs))

		err := c.Next()

		responseAttrs := attribute.NewSet(
			semconv.HTTPRouteKey.String(c.Route().Path),
			semconv.HTTPMethodKey.String(c.Method()),
			semconv.HTTPStatusCodeKey.Int(c.Response().StatusCode()),
		)

		o.RequestCounter.Add(c.Context(), 1, metric.WithAttributeSet(responseAttrs))
		o.RequestDuration.Record(c.Context(), time.Since(start).Seconds(), metric.WithAttributeSet(responseAttrs))

		return err
	}
}

// TracingMiddleware returns a Fiber handler for tracing.
func TracingMiddleware(appName string) fiber.Handler {
	tracer := otel.Tracer(appName)
	propagator := otel.GetTextMapPropagator()

	return func(c *fiber.Ctx) error {
		ctx := propagator.Extract(c.UserContext(), propagation.HeaderCarrier(c.GetReqHeaders()))
		spanName := c.Method() + " " + c.Path()

		ctx, span := tracer.Start(ctx, spanName, trace.WithSpanKind(trace.SpanKindServer))
		defer span.End()

		c.SetUserContext(ctx)

		span.SetName(spanName)

		span.SetAttributes(
			semconv.HTTPMethodKey.String(c.Method()),
			semconv.HTTPRouteKey.String(c.Route().Path),
			semconv.HTTPURLKey.String(c.OriginalURL()),
			semconv.NetHostNameKey.String(c.Hostname()),
		)

		err := c.Next()
		statusCode := c.Response().StatusCode()
		span.SetAttributes(semconv.HTTPStatusCodeKey.Int(statusCode))
		if err != nil {
			span.RecordError(err)
		}

		return err
	}
}
