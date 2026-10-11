package scheduler

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func BenchmarkRun(b *testing.B) {
	concurrencyLevels := []int{1, 2, 5, 10}
	const targetCount = 20

	for _, concurrency := range concurrencyLevels {
		b.Run(fmt.Sprintf("concurrency_%d", concurrency), func(b *testing.B) {
			server := httptest.NewServer(http.HandlerFunc(func(
				w http.ResponseWriter,
				r *http.Request,
			) {
				// Simulate a small amount of network/server latency.
				time.Sleep(5 * time.Millisecond)
				w.WriteHeader(http.StatusOK)
			}))
			defer server.Close()

			urls := make([]string, targetCount)
			for i := range urls {
				urls[i] = server.URL
			}

			client := NewClient(2 * time.Second)
			ctx := context.Background()

			b.ReportAllocs()
			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				outcomes := Run(ctx, client, urls, concurrency)

				if len(outcomes) != targetCount {
					b.Fatalf(
						"expected %d outcomes, got %d",
						targetCount,
						len(outcomes),
					)
				}

				for _, outcome := range outcomes {
					if outcome.Err != nil {
						b.Fatalf("check failed: %v", outcome.Err)
					}
					if !outcome.Result.Healthy {
						b.Fatalf(
							"unexpected unhealthy result: status %d",
							outcome.Result.StatusCode,
						)
					}
				}
			}
		})
	}
}
