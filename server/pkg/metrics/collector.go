package metrics

import (
	"context"
	"fmt"
	otlp "go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp"
	"go.opentelemetry.io/otel/metric/instrument"
	"go.opentelemetry.io/otel/metric/instrument/syncint64"
	"go.opentelemetry.io/otel/sdk/metric"
	"net/url"
)

type Collector struct {
	namespace string
	provider  *metric.MeterProvider

	TotalRequestCounter  syncint64.Counter
	TotalErrorCounter    syncint64.Counter
	TotalOpenConnections syncint64.UpDownCounter
	TotalRequestDuration syncint64.Histogram
}

func New(ctx context.Context, url url.URL, namespace string) (Collector, error) {
	exporter, err := otlp.New(
		ctx,
		otlp.WithInsecure(),
		otlp.WithEndpoint(url.Host),
	)
	if err != nil {
		return Collector{}, fmt.Errorf("could not initialize collector: %w", err)
	}

	var collector = Collector{
		namespace: namespace,
		provider: metric.NewMeterProvider(
			metric.WithReader(metric.NewPeriodicReader(exporter)),
		),
	}

	return collector, collector.initialize(namespace)
}

func (c *Collector) initialize(namespace string) (err error) {
	Meter := c.provider.Meter(namespace)
	c.TotalRequestCounter, err = Meter.SyncInt64().Counter(
		"http_requests_total",
		instrument.WithUnit("1"),
		instrument.WithDescription("Total number of requests"),
	)
	if err != nil {
		return fmt.Errorf("could not initialize total http request counter: %w", err)
	}

	c.TotalErrorCounter, err = Meter.SyncInt64().Counter(
		"errors_total",
		instrument.WithUnit("1"),
		instrument.WithDescription("Total number of errors"),
	)
	if err != nil {
		return fmt.Errorf("could not initialize total http request counter: %w", err)
	}

	c.TotalOpenConnections, err = Meter.SyncInt64().UpDownCounter(
		"open_connections_total",
		instrument.WithUnit("1"),
		instrument.WithDescription("Total open connections"),
	)
	if err != nil {
		return fmt.Errorf("could not initialize total open connections: %w", err)
	}

	c.TotalRequestDuration, err = Meter.SyncInt64().Histogram(
		"request_duration_millisecond",
		instrument.WithUnit("ms"),
		instrument.WithDescription("Total duration of requests"),
	)
	if err != nil {
		return fmt.Errorf("could not initialize total duration of requests")
	}

	return nil
}

func (c *Collector) Shutdown(ctx context.Context) error {
	return c.provider.Shutdown(ctx)
}
