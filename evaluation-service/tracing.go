package main

import (
	"context"
	"log"
	"os"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/propagation"
	sdkresource "go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

func setupTracing(ctx context.Context) (func(), error) {
	endpoint := os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT")
	tracesEndpoint := os.Getenv("OTEL_EXPORTER_OTLP_TRACES_ENDPOINT")

	if endpoint == "" && tracesEndpoint == "" {
		log.Println("OpenTelemetry tracing desabilitado: endpoint nao configurado")
		return func() {}, nil
	}

	exporter, err := otlptracehttp.New(ctx)
	if err != nil {
		return nil, err
	}

	provider := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithSampler(
			sdktrace.ParentBased(sdktrace.AlwaysSample()),
		),
		sdktrace.WithResource(
			sdkresource.NewWithAttributes(
				"",
				attribute.String("service.name", "evaluation-service"),
				attribute.String("deployment.environment.name", "fiap-lab"),
			),
		),
	)

	otel.SetTracerProvider(provider)
	otel.SetTextMapPropagator(
		propagation.NewCompositeTextMapPropagator(
			propagation.TraceContext{},
			propagation.Baggage{},
		),
	)

	log.Println("OpenTelemetry tracing habilitado")

	shutdown := func() {
		timeoutCtx, cancel := context.WithTimeout(
			context.Background(), 5*time.Second,
		)
		defer cancel()

		if err := provider.Shutdown(timeoutCtx); err != nil {
			log.Printf("Erro ao finalizar OpenTelemetry: %v", err)
		}
	}

	return shutdown, nil
}
