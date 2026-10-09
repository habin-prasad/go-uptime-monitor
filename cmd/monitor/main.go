package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"os"
	"time"
)

func main() {
	fmt.Println("Go Uptime monitor starting...")
	targetURL := flag.String("url", "", "URL to check")
	timeout := flag.Duration("timeout", 5*time.Second, "Request timeout")
	flag.Parse()

	if *targetURL == "" {
		fmt.Fprintln(os.Stderr, "Usage: monitor -url https://example.com")
		os.Exit(2)
	}

	if *timeout <= 0 {
		fmt.Fprintln(os.Stderr, "Error: timeout must be positive")
		os.Exit(2)
	}

	ctx, cancel := context.WithTimeout(context.Background(), *timeout)
	defer cancel()

	client := &http.Client{
		Timeout: *timeout,
		CheckRedirect: func(
			req *http.Request,
			via []*http.Request,
		) error {
			if len(via) >= 10 {
				return http.ErrUseLastResponse
			}
			return nil
		},
	}

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodGet,
		*targetURL,
		nil,
	)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Invalide request: %v\n", err)
		os.Exit(2)
	}

	start := time.Now()
	resp, err := client.Do(req)
	duration := time.Since(start)

	if err != nil {
		fmt.Printf("UNHEALTHY %s | error %v | duration=%s\n",
			*targetURL, err, duration.Round(time.Millisecond))
		os.Exit(2)
	}

	defer resp.Body.Close()

	healthy := resp.StatusCode >= 200 && resp.StatusCode < 400
	status := "UNHEALTHY"
	if healthy {
		status = "HEALTHY"
	}

	fmt.Printf("%s %s | status=%d | duration=%s\n",
		status, *targetURL, resp.StatusCode, duration.Round(time.Millisecond))

	if !healthy {
		os.Exit(1)
	}

}
