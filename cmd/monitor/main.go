package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"time"

	"github.com/habin-prasad/go-uptime-monitor/internal/checker"
	"github.com/habin-prasad/go-uptime-monitor/internal/scheduler"
	"github.com/habin-prasad/go-uptime-monitor/internal/targets"

	"github.com/habin-prasad/go-uptime-monitor/internal/monitor"
)

func run() int {

	targetURL := flag.String("url", "", "URL to check")
	timeout := flag.Duration("timeout", 5*time.Second, "Request timeout")
	filePath := flag.String("file", "", "Path to a file containing URLs, one per line")
	concurrency := flag.Int("concurrency", 5, "maximum number of concurrent URL checks")
	outputFormat := flag.String("format", "text", "Output format: text or json")
	flag.Parse()

	if *outputFormat != "text" && *outputFormat != "json" {
		fmt.Fprintln(os.Stderr, "Invalid format: use text or json")
		return 2
	}
	if *outputFormat == "text" {
		fmt.Println("Go Uptime monitor starting...")
	}
	var urls []string
	if *targetURL != "" {
		urls = append(urls, *targetURL)
	}

	if *filePath != "" {
		fileURLs, err := targets.LoadFile(*filePath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error encountered: %v\n", err)
			return 2
		}
		urls = append(urls, fileURLs...)
	}

	if len(urls) == 0 {
		fmt.Fprintln(os.Stderr, "Usage: monitor -url https://example.com [-timeout 5s] | -file urls.txt")
		return 2
	}

	for _, target := range urls {
		if err := checker.ValidateURL(target); err != nil {
			fmt.Fprintln(os.Stderr, err)
			return 2
		}
	}

	ctx, stop := signal.NotifyContext(
		context.Background(),
		os.Interrupt,
	)
	defer stop()

	client := scheduler.NewClient(*timeout)
	outcomes := scheduler.Run(
		ctx,
		client,
		urls,
		*concurrency,
	)

	hasFailures := false

	report := monitor.Report{
		Results: make([]monitor.CheckResult, 0, len(outcomes)),
	}

	for i, outcome := range outcomes {
		item := monitor.CheckResult{
			URL: urls[i],
		}

		if outcome.Err != nil {
			message := outcome.Err.Error()
			item.Error = &message
			hasFailures = true
			report.Summary.Errors++
		} else {
			result := outcome.Result

			item.URL = result.URL
			item.StatusCode = result.StatusCode
			item.DurationMS = float64(result.Duration) / float64(time.Millisecond)
			item.Healthy = result.Healthy

			if result.Healthy {
				report.Summary.Healthy++
			} else {
				report.Summary.Unhealthy++
				hasFailures = true
			}
		}

		report.Results = append(report.Results, item)
	}

	report.Summary.Total = len(report.Results)

	if *outputFormat == "json" {
		if err := json.NewEncoder(os.Stdout).Encode(report); err != nil {
			fmt.Fprintf(os.Stderr, "Error encoding JSON: %v\n", err)
			return 2
		}
	} else {
		for _, item := range report.Results {
			if item.Error != nil {
				fmt.Printf("ERROR     %s: %s\n", item.URL, *item.Error)
				continue
			}

			status := "UNHEALTHY"
			if item.Healthy {
				status = "HEALTHY"
			}

			fmt.Printf(
				"%-10s %s (status: %d, duration: %.2fms)\n",
				status,
				item.URL,
				item.StatusCode,
				item.DurationMS,
			)
		}
	}

	if hasFailures {
		return 1
	}

	return 0
}

func main() {
	os.Exit(run())
}
