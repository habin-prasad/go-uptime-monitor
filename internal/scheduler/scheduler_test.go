package scheduler

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"
)

func TestRunPreserveInputOrder(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			switch r.URL.Path {

			case "/slow":
				time.Sleep(50 * time.Millisecond)
				w.WriteHeader(http.StatusCreated)
			case "/fast":
				w.WriteHeader(http.StatusNoContent)
			default:
				w.WriteHeader(http.StatusOK)
			}
		},
	))
	defer server.Close()

	urls := []string{
		server.URL + "/slow",
		server.URL + "/fast",
		server.URL + "/normal",
	}

	outcomes := Run(
		context.Background(),
		NewClient(time.Second),
		urls,
		3,
	)
	if len(outcomes) != len(urls) {
		t.Fatalf("got %d outcomes, want %d", len(outcomes), len(urls))
	}
	wantStatuses := []int{
		http.StatusCreated,
		http.StatusNoContent,
		http.StatusOK,
	}

	for i, outcome := range outcomes {
		if outcome.Err != nil {
			t.Errorf("outcome %d: unexpected error: %v", i, outcome.Err)
			continue
		}
		if outcome.Result.StatusCode != wantStatuses[i] {
			t.Errorf(
				"outcome %d: status = %d, want %d",
				i, outcome.Result.StatusCode, wantStatuses[i],
			)
		}
		if outcome.Result.URL != urls[i] {
			t.Errorf(
				"outcome %d: URL = %q, want %q",
				i, outcome.Result.URL, urls[i],
			)
		}
	}

}

func TestRunRespectsConcurrencyLimit(t *testing.T) {
	const limit = 3
	const total = 12

	var active atomic.Int32
	var maxActive atomic.Int32

	server := httptest.NewServer(http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			current := active.Add(1)
			defer active.Add(-1)

			for {
				previous := maxActive.Load()
				if current <= previous ||
					maxActive.CompareAndSwap(previous, current) {
					break
				}
			}

			time.Sleep(30 * time.Millisecond)
			w.WriteHeader(http.StatusOK)
		},
	))
	defer server.Close()

	urls := make([]string, total)
	for i := range urls {
		urls[i] = fmt.Sprintf("%s/check/%d", server.URL, i)
	}

	outcomes := Run(
		context.Background(),
		NewClient(2*time.Second),
		urls,
		limit,
	)

	for i, outcome := range outcomes {
		if outcome.Err != nil {
			t.Errorf("outcome %d: unexpected error: %v", i, outcome.Err)
		}
	}

	if got := maxActive.Load(); got > limit {
		t.Errorf("maximum concurrent requests = %d, want at most %d", got, limit)
	}
	if got := maxActive.Load(); got < 2 {
		t.Errorf("maximum concurrent requests = %d, expected concurrent work", got)
	}
}

func TestRunEmptyURLs(t *testing.T) {
	outcomes := Run(
		context.Background(),
		NewClient(time.Second),
		nil,
		4,
	)

	if len(outcomes) != 0 {
		t.Fatalf("got %d outcomes, want 0", len(outcomes))
	}
}

func TestRunWithCanceledContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	urls := []string{
		"https://example.com",
		"https://example.org",
		"https://example.net",
	}

	client := NewClient(2 * time.Second)

	outcomes := Run(ctx, client, urls, 2)

	if len(outcomes) != len(urls) {
		t.Fatalf(
			"expected %d outcomes, got %d",
			len(urls),
			len(outcomes),
		)
	}

	for i, outcome := range outcomes {
		if !errors.Is(outcome.Err, context.Canceled) {
			t.Errorf(
				"outcome %d: expected cancellation error, got %v",
				i,
				outcome.Err,
			)
		}
	}
}

func TestRunCancellationDuringActiveRequests(t *testing.T) {
	requestStarted := make(chan struct{}, 2)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestStarted <- struct{}{}

		// Keep the request active until the client cancels it.
		<-r.Context().Done()
	}))

	defer server.Close()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	urls := []string{
		server.URL + "/1",
		server.URL + "/2",
		server.URL + "/3",
		server.URL + "/4",
	}

	client := NewClient(5 * time.Second)

	done := make(chan []Outcome, 1)
	go func() {
		done <- Run(ctx, client, urls, 2)
	}()

	// Wait until both workers have started HTTP requests.
	for i := 0; i < 2; i++ {
		select {
		case <-requestStarted:
		case <-time.After(2 * time.Second):
			cancel()
			t.Fatal("timed out waiting for requests to start")
		}
	}

	cancel()

	select {
	case outcomes := <-done:
		if len(outcomes) != len(urls) {
			t.Fatalf("expected %d outcomes, got %d", len(urls), len(outcomes))
		}

		for i, outcome := range outcomes {
			if !errors.Is(outcome.Err, context.Canceled) {
				t.Errorf("outcome %d: expected cancellation error, got %v", i, outcome.Err)
			}
		}

	case <-time.After(3 * time.Second):
		t.Fatal("scheduler did not return after cancellation")
	}

}
