package middleware

import (
	"platform/metric"

	"github.com/go-kratos/kratos/v2/middleware"
	"github.com/go-kratos/kratos/v2/middleware/metrics"
)

func MetricForServer(serviceName string) middleware.Middleware {
	metricRequests, metricSeconds := metric.NewMetric(serviceName)

	return metrics.Server(
		metrics.WithSeconds(metricSeconds),
		metrics.WithRequests(metricRequests),
	)
}
