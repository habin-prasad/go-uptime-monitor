package scheduler

import (
	"context"
	"net/http"
	"sync"
	"time"

	"github.com/habin-prasad/go-uptime-monitor/internal/checker"
)

type Outcome struct {
	Result checker.Result
	Err    error
}

func Run(
	ctx context.Context,
	client *http.Client,
	urls []string,
	concurrency int,
) []Outcome {
	if concurrency < 1 {
		concurrency = 1
	}
	if concurrency > len(urls) {
		concurrency = len(urls)
	}

	outcomes := make([]Outcome, len(urls))
	if len(urls) == 0 {
		return outcomes
	}

	jobs := make(chan int)

	var wg sync.WaitGroup

	for range concurrency {
		wg.Add(1)
		go func() {
			defer wg.Done()

			for index := range jobs {
				if err := ctx.Err(); err != nil {
					outcomes[index].Err = err
					continue
				}
				result, err := checker.Check(ctx, client, urls[index])

				outcomes[index] = Outcome{
					Result: result,
					Err:    err,
				}
			}
		}()
	}
dispatch:
	for index := range urls {
		select {
		case <-ctx.Done():
			for remaining := index; remaining < len(urls); remaining++ {
				outcomes[remaining].Err = ctx.Err()
			}
			break dispatch

		case jobs <- index:
		}
	}

	close(jobs)
	wg.Wait()
	return outcomes

}

func NewClient(timeout time.Duration) *http.Client {
	return &http.Client{
		Timeout: timeout,
	}
}
