package common

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"

	"go.opentelemetry.io/contrib/bridges/otelslog"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploghttp"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/sdk/log"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
)

const serviceName = "open-iga-core"

// stdoutLogger logs JSON to stdout only, used when telemetry is disabled.
func stdoutLogger() *slog.Logger {
	return slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
}

// SetupOTel enables each OTLP signal whose endpoint URL is configured; signals
// with an empty URL are left off (no exporter or provider). Each URL is used
// as-is (scheme/TLS and path are read from it by the SDK). Logs go to the OTLP
// pipeline when their endpoint is set, otherwise to a JSON stdout logger. It
// returns the logger plus a shutdown func that flushes the enabled providers.
func SetupOTel(ctx context.Context, telemetry Telemetry) (*slog.Logger, func(context.Context) error, error) {
	otelResource, err := resource.Merge(resource.Default(),
		resource.NewSchemaless(semconv.ServiceName(serviceName)))
	if err != nil {
		return nil, nil, fmt.Errorf("otel resource: %w", err)
	}

	var shutdowns []func(context.Context) error

	if telemetry.TracesEndpoint != "" {
		traceExporter, err := otlptracehttp.New(ctx, otlptracehttp.WithEndpointURL(telemetry.TracesEndpoint))
		if err != nil {
			return nil, nil, fmt.Errorf("otel trace exporter: %w", err)
		}
		tracerProvider := trace.NewTracerProvider(trace.WithBatcher(traceExporter), trace.WithResource(otelResource))
		otel.SetTracerProvider(tracerProvider)
		shutdowns = append(shutdowns, tracerProvider.Shutdown)
	}

	if telemetry.MetricsEndpoint != "" {
		metricExporter, err := otlpmetrichttp.New(ctx, otlpmetrichttp.WithEndpointURL(telemetry.MetricsEndpoint))
		if err != nil {
			return nil, nil, fmt.Errorf("otel metric exporter: %w", err)
		}
		meterProvider := metric.NewMeterProvider(metric.WithReader(metric.NewPeriodicReader(metricExporter)), metric.WithResource(otelResource))
		otel.SetMeterProvider(meterProvider)
		shutdowns = append(shutdowns, meterProvider.Shutdown)
	}

	logger := stdoutLogger()
	if telemetry.LogsEndpoint != "" {
		logExporter, err := otlploghttp.New(ctx, otlploghttp.WithEndpointURL(telemetry.LogsEndpoint))
		if err != nil {
			return nil, nil, fmt.Errorf("otel log exporter: %w", err)
		}
		loggerProvider := log.NewLoggerProvider(log.WithProcessor(log.NewBatchProcessor(logExporter)), log.WithResource(otelResource))
		logger = slog.New(otelslog.NewHandler(serviceName, otelslog.WithLoggerProvider(loggerProvider)))
		shutdowns = append(shutdowns, loggerProvider.Shutdown)
	}

	shutdown := func(ctx context.Context) error {
		var shutdownErrors []error
		for _, shutdownProvider := range shutdowns {
			shutdownErrors = append(shutdownErrors, shutdownProvider(ctx))
		}
		return errors.Join(shutdownErrors...)
	}
	return logger, shutdown, nil
}
