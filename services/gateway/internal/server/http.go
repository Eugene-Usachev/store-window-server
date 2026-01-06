package server

import (
	gatewayv1 "api/gateway/v1"
	"context"
	"gateway/internal/service"
	"platform/metric"
	"platform/middleware"
	"time"

	"platform/logger"

	"github.com/caarlos0/env/v11"
	"github.com/go-kratos/kratos/contrib/middleware/validate/v2"
	"github.com/go-kratos/kratos/v2/middleware/logging"
	"github.com/go-kratos/kratos/v2/middleware/ratelimit"
	"github.com/go-kratos/kratos/v2/middleware/recovery"
	"github.com/go-kratos/kratos/v2/transport/http"
)

func mustCreateServer(ctx context.Context) *http.Server {
	type config struct {
		Addr    string `env:"HTTP_ADDR"`
		Network string `env:"HTTP_NETWORK"`
	}

	cfg, err := env.ParseAs[config]()
	if err != nil {
		logger.Fatal(err.Error())

		return nil
	}

	var opts = []http.ServerOption{
		http.Middleware(
			recovery.Recovery(),
			logging.Server(logger.MainLogger().Logger()),
			ratelimit.Server(ratelimit.WithLimiter(middleware.ServerRateLimiter(ctx, 50))),
			middleware.MetricForServer("gateway"),
			validate.ProtoValidate(),
		),
	}

	if cfg.Network == "" {
		logger.Info("HTTP network is not set, using default 'tcp'")
		cfg.Network = "tcp"
	}

	opts = append(opts, http.Network(cfg.Network))

	if cfg.Addr == "" {
		logger.Info("HTTP addr is not set, using default port 8080")
		cfg.Addr = ":8080"
	}

	opts = append(opts, http.Address(cfg.Addr))

	opts = append(opts, http.Timeout(1*time.Second))

	return http.NewServer(opts...)
}

// NewHTTPServer new an HTTP server.
func NewHTTPServer(ctx context.Context, s *service.Services) *http.Server {
	srv := mustCreateServer(ctx)

	gatewayv1.RegisterGatewayServiceHTTPServer(srv, s)

	metric.RegisterMetricHTTPEndpoint(srv)

	return srv
}
