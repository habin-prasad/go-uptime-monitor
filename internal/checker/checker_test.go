package checker

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestCheckStatusCode(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
		healthy    bool
	}{
		{"OK", http.StatusOK, true},
		{"Not Found", http.StatusNotFound, false},
		{"Server Error", http.StatusInternalServerError, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(
				w http.ResponseWriter,
				r *http.Request,
			) {
				w.WriteHeader(tt.statusCode)
			}),
			)
			defer server.Close()
			result, err := Check(
				context.Background(),
				server.Client(),
				server.URL,
			)
			if err != nil {
				t.Fatalf("Check returned unexpected error %v", err)
			}
			if result.StatusCode != tt.statusCode {
				t.Errorf("got status %d, wnt %d", result.StatusCode, tt.statusCode)
			}
			if result.Healthy != tt.healthy {
				t.Errorf("got Healthy=%t, want %t",
					result.Healthy,
					tt.healthy,
				)
			}
		})
	}
}

func TestInvalidURL(t *testing.T) {
	client := &http.Client{}
	_, err := Check(
		context.Background(),
		client,
		":/invalid-url",
	)

	if err == nil {
		t.Fatal("expected an error for an invalid URL")
	}
}

func TestCheckTimeout(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(
		w http.ResponseWriter, r *http.Request,
	) {
		<-r.Context().Done()
	},
	))
	defer server.Close()
	ctx, cancel := context.WithTimeout(
		context.Background(),
		50*time.Millisecond,
	)

	defer cancel()

	_, err := Check(ctx, server.Client(), server.URL)

	if err == nil {
		t.Fatal("expected an error when the request times out")
	}
}
