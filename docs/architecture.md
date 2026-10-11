# Architecture: Go Uptime Monitor

## 1. Purpose

Go Uptime Monitor is a command-line tool that checks the health of multiple HTTP endpoints. It demonstrates bounded concurrency, context-based cancellation, HTTP client timeouts, deterministic result ordering, automated testing, and structured reporting in Go.

## 2. High-level design

The application separates command-line orchestration, target loading, HTTP checking, scheduling, and report formatting.

```text
CLI / cmd/monitor
       |
       +---- Target loading and validation
       |
       +---- HTTP client configuration
       |
       +---- Scheduler
       |       |
       |       +---- Worker 1 ---- HTTP checker
       |       +---- Worker 2 ---- HTTP checker
       |       +---- Worker N ---- HTTP checker
       |
       +---- Ordered outcomes
       |
       +---- Text or JSON report
```

## 3. Component responsibilities

* **CLI (`cmd/monitor`)**: parses flags, validates configuration, coordinates execution, and selects the output format.
* **Target loader (`internal/targets`)**: reads URLs from a file, trimming whitespace and ignoring blank lines and comments.
* **Checker (`internal/checker`)**: performs an HTTP request, measures its duration, and classifies the HTTP status.
* **Scheduler (`internal/scheduler`)**: distributes checks across a bounded worker pool and collects outcomes in input order.
* **Report model (`internal/monitor`)**: represents check results and aggregate counts for structured output.

## 4. Concurrency model

The scheduler uses a fixed worker pool rather than launching an unrestricted goroutine for every target.

The concurrency setting limits how many checks can be processed simultaneously by the scheduler. This helps control resource consumption and the load generated against monitored services.

Results are stored by their original input index, preserving target order even when requests complete in a different order.

## 5. Cancellation and timeouts

The application uses contexts to propagate cancellation to HTTP requests. The HTTP client also has a configurable timeout.

These mechanisms serve different purposes:

* **Context cancellation** allows execution to stop when the caller cancels the operation or the process receives an interrupt.
* **HTTP client timeout** bounds the duration of an HTTP request.

Cancellation behavior and the handling of checks that have not yet started should be covered by automated tests.

## 6. Health classification

The checker classifies HTTP status codes from 200 through 399 as healthy. Other received HTTP statuses are classified as unhealthy.

Request failures, such as connection errors and timeouts, are represented separately from HTTP responses.

## 7. Output and exit behavior

The CLI supports human-readable text and machine-readable JSON output.

The intended exit-code contract is:

* `0`: all checks are healthy.
* `1`: one or more checks are unhealthy or encounter request errors.
* `2`: invalid command-line usage or configuration.

The implementation and tests should remain the source of truth for this contract.

## 8. Performance measurements

The scheduler has been benchmarked using a local HTTP test server with an artificial 5 ms response delay and 20 targets.

On an Apple M1, the initial benchmark showed that increasing concurrency from 1 to 10 reduced batch execution time from approximately 121.7 ms to 16.2 ms, while allocations per batch increased from approximately 1,291 to 1,982.

These measurements describe a controlled local workload. They should not be treated as production throughput guarantees.

## 9. Current limitations

This project currently focuses on periodic or manually invoked HTTP checks rather than a full monitoring platform.

Potential future extensions include:

* Periodic scheduling and historical metrics
* TLS certificate-expiration monitoring
* Retries with bounded backoff
* Prometheus-compatible metrics
* Persistent storage and alerting

Future extensions should be driven by a clear use case and explicit reliability requirements.
