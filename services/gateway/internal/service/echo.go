package service

import (
	"context"

	gatewayv1 "api/gateway/v1"
	"platform/markers"
)

type EchoService struct {
	markers.NoCopy
}

var _ gatewayv1.GatewayServiceServer = (*EchoService)(nil)

func NewEchoService() EchoService {
	return EchoService{}
}

func (e *EchoService) Echo(_ context.Context, request *gatewayv1.EchoRequest) (*gatewayv1.EchoResponse, error) {
	return &gatewayv1.EchoResponse{
		Message: request.Message,
	}, nil
}
