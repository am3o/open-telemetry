package main

import (
	"context"
	"log"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp"
	stdout "go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/metric/global"
	"go.opentelemetry.io/otel/metric/instrument"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/metric"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

func main() {
	ctx := context.Background()

	exp, err := otlpmetrichttp.New(ctx, otlpmetrichttp.WithInsecure(), otlpmetrichttp.WithEndpoint("localhost:4318"))
	if err != nil {
		panic(err)
	}

	meterProvider := metric.NewMeterProvider(metric.WithReader(metric.NewPeriodicReader(exp)))
	defer func() {
		if err := meterProvider.Shutdown(ctx); err != nil {
			panic(err)
		}
	}()
	global.SetMeterProvider(meterProvider)

	// From here, the meterProvider can be used by instrumentation to collect
	// telemetry.

	//creating a meter
	Meter := global.MeterProvider().Meter("server")
	counter, err := Meter.SyncInt64().Counter(
		"test_my_counter",
		instrument.WithUnit("1"),
		instrument.WithDescription("Just a test counter"),
	)

	if err != nil {
		panic(err)
	}

	for range time.NewTicker(time.Second).C {
		counter.Add(ctx, 1, attribute.String("type", "example"))
	}
}

func InitStdoutTracerProvider() *sdktrace.TracerProvider {
	// Create stdout exporter to be able to retrieve
	// the collected spans.
	exporter, err := stdout.New(stdout.WithPrettyPrint())
	if err != nil {
		log.Fatal(err)
	}

	// For the demonstration, use sdktrace.AlwaysSample sampler to sample all traces.
	// In a production application, use sdktrace.ProbabilitySampler with a desired probability.
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
		sdktrace.WithBatcher(exporter),
	)
	otel.SetTracerProvider(tp)

	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, propagation.Baggage{}))
	return tp
}
