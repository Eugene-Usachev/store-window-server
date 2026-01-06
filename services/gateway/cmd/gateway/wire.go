//go:build wireinject
// +build wireinject

package main

import (
	"context"

	"gateway/internal/server"
	"gateway/internal/service"

	"github.com/go-kratos/kratos/v2"
	"github.com/google/wire"
)

func provideContext() (context.Context, func(), error) {
	ctx, cancel := context.WithCancel(context.Background())
	return ctx, cancel, nil
}

func wireApp() (*kratos.App, func(), error) {
	panic(wire.Build(provideContext, server.ProviderSet, service.ProviderSet, newApp))
}
