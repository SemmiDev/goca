package observability

import (
	"context"
	"errors"
	"time"

	"github.com/sammidev/goca/internal/config"
	"github.com/sammidev/goca/internal/pkg/logger"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.24.0"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// SetupOTelProvider menginisialisasi dan mendaftarkan provider OTel untuk metrics dan traces.
// Semua data telemetri akan dikirim ke OTel Collector melalui OTLP/gRPC.
func SetupOTelProvider(ctx context.Context, cfg *config.Config, log logger.Logger) (shutdown func(context.Context) error, err error) {
	// Buat resource yang akan dilampirkan ke semua data telemetri.
	// Ini memberikan konteks, seperti nama layanan.
	res, err := resource.New(ctx,
		resource.WithAttributes(
			semconv.ServiceName(cfg.AppName),
		),
	)
	if err != nil {
		return nil, err
	}

	// Buat koneksi gRPC ke OTel Collector.
	// cfg.OtelExporterOtlpEndpoint harus menunjuk ke Collector, misalnya "otel-collector:4317".
	conn, err := grpc.NewClient(cfg.OtelExporterOtlpEndpoint,
		grpc.WithTransportCredentials(insecure.NewCredentials()), // Gunakan koneksi insecure untuk dev lokal.
	)
	if err != nil {
		return nil, err
	}

	// ---- KONFIGURASI TRACER PROVIDER ----
	// Buat exporter trace OTLP yang akan mengirim data melalui koneksi gRPC.
	traceExporter, err := otlptracegrpc.New(ctx, otlptracegrpc.WithGRPCConn(conn))
	if err != nil {
		return nil, err
	}

	// BatchSpanProcessor mengelompokkan span sebelum mengirimkannya untuk efisiensi.
	bsp := sdktrace.NewBatchSpanProcessor(traceExporter)
	tracerProvider := sdktrace.NewTracerProvider(
		sdktrace.WithSpanProcessor(bsp), // Perbaikan: Menggunakan WithSpanProcessor
		sdktrace.WithResource(res),
	)
	otel.SetTracerProvider(tracerProvider) // Atur sebagai provider trace global.

	// ---- KONFIGURASI METER PROVIDER ----
	// Buat exporter metrik OTLP.
	metricExporter, err := otlpmetricgrpc.New(ctx, otlpmetricgrpc.WithGRPCConn(conn))
	if err != nil {
		return nil, err
	}

	// MeterProvider adalah pabrik untuk Meter.
	// PeriodicReader secara berkala membaca metrik dan mengirimkannya ke exporter.
	meterProvider := sdkmetric.NewMeterProvider(
		sdkmetric.WithReader(sdkmetric.NewPeriodicReader(metricExporter, sdkmetric.WithInterval(5*time.Second))),
		sdkmetric.WithResource(res),
	)
	otel.SetMeterProvider(meterProvider) // Atur sebagai provider metrik global.

	// Atur propagator untuk W3C Trace Context.
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, propagation.Baggage{}))

	log.Info("OpenTelemetry provider (metrics & traces) berhasil diinisialisasi untuk OTel Collector.")

	// Fungsi shutdown untuk membersihkan semua koneksi saat aplikasi berhenti.
	shutdown = func(ctx context.Context) error {
		return errors.Join(
			tracerProvider.Shutdown(ctx),
			meterProvider.Shutdown(ctx),
			conn.Close(),
		)
	}
	return
}
