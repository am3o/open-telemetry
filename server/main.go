package main

import (
	"context"
	"fmt"
	"github.com/am3o/OpenTelemetry/server/pkg/metrics"
	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/otel/attribute"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"time"
)

func main() {
	env, ok := os.LookupEnv("OPEN_TELEMETRY_COLLECTOR_URL")
	if !ok {
		panic("open telemetry collector url must be set")
	}

	collectorURL, err := url.Parse(env)
	if err != nil {
		panic(fmt.Sprintf("could not parse the collector url: %w", err))
	}

	ctx := context.TODO()
	collector, err := metrics.New(ctx, *collectorURL, "country_service")
	if err != nil {
		panic(err)
	}
	defer collector.Shutdown(ctx)

	router := gin.Default()
	router.GET("/api/v1/examples", func(c *gin.Context) {
		start := time.Now()
		defer collector.TotalRequestDuration.Record(
			c.Request.Context(),
			time.Since(start).Milliseconds(),
		)

		collector.TotalRequestCounter.Add(c.Request.Context(), 1)
		collector.TotalOpenConnections.Add(c.Request.Context(), rand.Int63(), attribute.String("agent", c.Request.UserAgent()))

		if rand.Int()%2 == 0 {
			collector.TotalErrorCounter.Add(c.Request.Context(), 1,
				attribute.String("host", c.Request.Host),
				attribute.String("method", c.Request.Method),
				attribute.String("path", c.Request.URL.Path),
			)

			c.Writer.WriteHeader(http.StatusNotFound)
			return
		}
	})

	if err := router.Run("0.0.0.0:8080"); err != nil {
		panic(err)
	}
}
