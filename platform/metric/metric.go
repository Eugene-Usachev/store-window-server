package metric

import (
	"platform/logger"

	"github.com/caarlos0/env/v11"
	"github.com/go-kratos/kratos/v2/middleware/metrics"
	"github.com/go-kratos/kratos/v2/transport/http"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.opentelemetry.io/otel/exporters/prometheus"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/metric/noop"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
)

type config struct {
	Enabled              bool   `env:"METRICS_ENABLED" envDefault:"true"`
	RequestsCounterName  string `env:"METRICS_REQUESTS_COUNTER_NAME" envDefault:"server_requests_total"`
	SecondsHistogramName string `env:"METRICS_SECONDS_HISTOGRAM_NAME" envDefault:"server_request_duration_seconds"`
}

func mustLoadConfig() config {
	cfg, err := env.ParseAs[config]()
	if err != nil {
		logger.Fatal(err.Error())
	}
	return cfg
}

func NewMetric(serviceName string) (metric.Int64Counter, metric.Float64Histogram) {
	cfg := mustLoadConfig()

	if !cfg.Enabled {
		return noop.Int64Counter{}, noop.Float64Histogram{}
	}

	exporter, err := prometheus.New()
	if err != nil {
		logger.Fatal(err.Error())
	}

	provider := sdkmetric.NewMeterProvider(
		sdkmetric.WithReader(exporter),
	)

	meter := provider.Meter(serviceName)

	requestsCounter, err := metrics.DefaultRequestsCounter(
		meter,
		cfg.RequestsCounterName,
	)
	if err != nil {
		logger.Fatal(err.Error())
	}

	secondsHistogram, err := metrics.DefaultSecondsHistogram(
		meter,
		cfg.SecondsHistogramName,
	)
	if err != nil {
		logger.Fatal(err.Error())
	}

	return requestsCounter, secondsHistogram
}

func RegisterMetricHTTPEndpoint(server *http.Server) {
	server.Handle("/metrics", promhttp.Handler())
}
