package e2e_tests

import (
	"net/http"
	"testing"
	"time"

	"github.com/gavv/httpexpect/v2"
)

func NewHTTPClient(t *testing.T) *httpexpect.Expect {
	return httpexpect.WithConfig(httpexpect.Config{
		BaseURL: "http://localhost:8080/v1",
		Client: &http.Client{
			Timeout: 5 * time.Second,
		},
		Reporter: httpexpect.NewAssertReporter(t),
		Printers: []httpexpect.Printer{
			httpexpect.NewDebugPrinter(t, true),
		},
	})
}

type RawHTTPClient struct {
	inner http.Client
}

func NewRawHTTPClient() *RawHTTPClient {
	return &RawHTTPClient{
		inner: http.Client{
			Timeout: 5 * time.Second,
		},
	}
}

func (c *RawHTTPClient) Do(req *http.Request) (*http.Response, error) {
	req.URL.Host = "localhost:8080"
	req.URL.Scheme = "http"
	req.URL.Path = "/v1" + req.URL.Path

	return c.inner.Do(req)
}
