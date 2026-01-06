package main

import (
	"context"
	"os"
	"platform/logger"

	"github.com/go-kratos/kratos/v2"
	"github.com/go-kratos/kratos/v2/transport/http"
	_ "go.uber.org/automaxprocs"
)

// go build -ldflags "-X main.Version=x.y.z"
var (
	Name    string
	Version string
	id, _   = os.Hostname()
)

func newApp(ctx context.Context, hs *http.Server) *kratos.App {
	return kratos.New(
		kratos.ID(id),
		kratos.Name(Name),
		kratos.Version(Version),
		kratos.Metadata(map[string]string{}),
		kratos.Logger(logger.MainLogger().Logger()),
		kratos.Server(hs),
		kratos.Context(ctx),
	)
}

func main() {
	app, cleanup, err := wireApp()
	if err != nil {
		panic(err)
	}

	defer cleanup()

	logger.MainLogger().Info("Gateway started")

	if err = app.Run(); err != nil {
		panic(err)
	}
}
