package checker

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type Result struct {
	URL        string
	StatusCode int
	Duration   time.Duration
	Healthy    bool
}

func Check(
	ctx context.Context,
	client *http.Client,
	targetURL string,
) (Result, error) {
	start := time.Now()

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodGet,
		targetURL,
		nil,
	)
	if err != nil {
		return Result{}, fmt.Errorf("create request: %w", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		return Result{}, fmt.Errorf("send request: %w", err)
	}
	defer resp.Body.Close()

	result := Result{
		URL:        targetURL,
		StatusCode: resp.StatusCode,
		Duration:   time.Since(start),
		Healthy:    resp.StatusCode >= 200 && resp.StatusCode < 400,
	}
	return result, nil
}

func ValidateURL(rawURL string) error {
	parsed, err := url.ParseRequestURI(rawURL)
	if err != nil {
		return fmt.Errorf("invalid URL %q: %w", rawURL, err)
	}

	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return fmt.Errorf("unsupported URL scheme in %q: use http or https", rawURL)
	}

	if parsed.Hostname() == "" {
		return fmt.Errorf("URL %q must include a hostname", rawURL)
	}

	if strings.TrimSpace(rawURL) != rawURL {
		return fmt.Errorf("URL must not contain leading or trailing whitespace")
	}
	return nil
}
