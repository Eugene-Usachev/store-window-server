package middleware

import (
	"context"
	"platform/logger"
	"sync/atomic"
	"time"

	"github.com/go-kratos/aegis/ratelimit"
	"github.com/go-kratos/aegis/ratelimit/bbr"
)

type serverRateLimiter struct {
	bbr                   *bbr.BBR
	requestsLeft          atomic.Int64
	requestPerSecondLimit int64
}

var _ ratelimit.Limiter = (*serverRateLimiter)(nil)

func ServerRateLimiter(ctx context.Context, requestPerSecondLimit uint64) ratelimit.Limiter {
	const Frequency = 10

	if requestPerSecondLimit < Frequency {
		logger.Fatalf("requestPerSecondLimit must not be less than %d", Frequency)
	}

	limit := int64(requestPerSecondLimit)

	s := &serverRateLimiter{
		bbr:                   bbr.NewLimiter(),
		requestsLeft:          atomic.Int64{},
		requestPerSecondLimit: limit,
	}

	s.requestsLeft.Store(limit)

	go func() {
		ticker := time.NewTicker(time.Second / Frequency)

		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				for {
					curr := s.requestsLeft.Load()
					canAdd := limit - curr
					maxAdd := limit / Frequency
					additional := min(canAdd, maxAdd)

					isSuccess := s.requestsLeft.CompareAndSwap(curr, curr+additional)
					if isSuccess {
						break
					}
				}
			}
		}
	}()

	return s
}

func (s *serverRateLimiter) Allow() (ratelimit.DoneFunc, error) {
	curr := s.requestsLeft.Add(-1)
	if curr < 0 {
		// It is a hot-path optimization: we believe that it is unlikely to exceed the limit,
		// so we use one atomic adding on a hot path and two ones on a cold one and CAS in the refilling loop.

		s.requestsLeft.Add(1)

		return nil, ratelimit.ErrLimitExceed
	}

	bbrFunc, err := s.bbr.Allow()
	if err != nil {
		return nil, err
	}

	return func(info ratelimit.DoneInfo) {
		bbrFunc(info)
	}, nil
}
