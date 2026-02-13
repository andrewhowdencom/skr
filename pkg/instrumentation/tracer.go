package instrumentation

import (
	"context"
	"fmt"
	"runtime/debug"

	"github.com/google/uuid"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.24.0"
)

// InitTracer initializes the OpenTelemetry tracer provider
func InitTracer(ctx context.Context, serviceName, provider, endpoint string) (func(context.Context) error, error) {
	var exporter sdktrace.SpanExporter
	var err error
	var opts []sdktrace.TracerProviderOption

	switch provider {
	case "stdout":
		exporter, err = stdouttrace.New(
			stdouttrace.WithPrettyPrint(),
		)
		// Use Syncer for stdout to ensure immediate printing
		opts = append(opts, sdktrace.WithSyncer(exporter))
	case "otlp":
		otlpOpts := []otlptracehttp.Option{
			otlptracehttp.WithInsecure(), // For local simplicity
		}
		if endpoint != "" {
			otlpOpts = append(otlpOpts, otlptracehttp.WithEndpoint(endpoint))
		}
		exporter, err = otlptracehttp.New(ctx, otlpOpts...)
		opts = append(opts, sdktrace.WithBatcher(exporter))
	case "none", "":
		return func(_ context.Context) error { return nil }, nil
	default:
		return nil, fmt.Errorf("unknown trace provider: %s", provider)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to create exporter: %w", err)
	}

	// Determine version
	version := "dev"
	if info, ok := debug.ReadBuildInfo(); ok {
		if info.Main.Version != "" && info.Main.Version != "(devel)" {
			version = info.Main.Version
		}
	}

	// Generate Instance ID
	instanceID := uuid.New().String()

	res, err := resource.New(ctx,
		resource.WithAttributes(
			semconv.ServiceNameKey.String(serviceName),
			semconv.ServiceVersionKey.String(version),
			semconv.ServiceInstanceIDKey.String(instanceID),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create resource: %w", err)
	}

	opts = append(opts, sdktrace.WithResource(res))

	tp := sdktrace.NewTracerProvider(opts...)

	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, propagation.Baggage{}))

	return tp.Shutdown, nil
}
