package service

import (
	"platform/markers"

	"github.com/google/wire"
)

type Services struct {
	markers.NoCopy
	EchoService
}

func NewServices() *Services {
	return &Services{
		EchoService: NewEchoService(),
	}
}

// ProviderSet is service providers.
var ProviderSet = wire.NewSet(NewServices)
