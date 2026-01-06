package e2e_tests

import (
	gatewayv1 "api/gateway/v1"
	"net/http"
	"testing"
)

func TestEcho(t *testing.T) {
	AcquireTestLockShared()
	defer ReleaseTestLockShared()

	client := NewHTTPClient(t)

	body := &gatewayv1.EchoRequest{
		Message: "Hey",
	}

	client.
		POST("/echo").
		WithJSON(body).
		Expect().
		Status(http.StatusOK).
		JSON().
		Object().
		HasValue("message", "Hey")

	client.
		POST("/echo").
		Expect().
		Status(http.StatusBadRequest)
}
