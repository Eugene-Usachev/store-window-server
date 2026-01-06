package e2e_tests

import (
	gatewayv1 "api/gateway/v1"
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestRateLimit(t *testing.T) {
	AcquireTestLockExclusive()
	defer ReleaseTestLockExclusive()

	const (
		Limit      = 50
		BurstCount = Limit * 3
		MaxRetries = 5
	)

	wg := sync.WaitGroup{}

	sendEcho := func() (int, error) {
		client := NewRawHTTPClient()
		body, _ := json.Marshal(&gatewayv1.EchoRequest{Message: "Hey"})

		req, err := http.NewRequest(http.MethodPost, "/echo", bytes.NewReader(body))
		if err != nil {
			return 0, err
		}
		req.Header.Set("Content-Type", "application/json")

		res, err := client.Do(req)
		if err != nil {
			return 0, err
		}

		defer func(Body io.ReadCloser) {
			err = Body.Close()
			if err != nil {
				t.Errorf("Failed to close response body: %v", err)
			}
		}(res.Body)

		return res.StatusCode, nil
	}
	rateLimitedCount := atomic.Uint64{}

	for attempt := 0; attempt < MaxRetries; attempt++ {
		start := time.Now()

		for range BurstCount {
			wg.Go(func() {
				status, err := sendEcho()
				if err != nil {
					t.Errorf("Request failed: %v", err)
					return
				}

				if status == http.StatusTooManyRequests {
					rateLimitedCount.Add(1)
				} else if status != http.StatusOK {
					t.Errorf("Unexpected status: %d", status)
				}
			})
		}

		wg.Wait()

		if time.Since(start) > 1500*time.Millisecond {
			rateLimitedCount.Store(0)

			assert.Less(t, attempt, MaxRetries, "server is too slow to test the rate limit")

			continue
		}

		break
	}

	assert.GreaterOrEqual(t, rateLimitedCount.Load(), uint64(Limit))
}
