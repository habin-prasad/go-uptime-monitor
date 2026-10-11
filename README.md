# Go Uptime Monitor

A concurrent HTTP health-check engine written in Go, demonstrating bounded worker pools, context-aware cancellation, deterministic result collection, structured reporting, and performance benchmarking.

The project explores how to build a lightweight I/O-bound system with explicit concurrency limits, predictable behavior, and measurable trade-offs.

## Features

- **HTTP health checks:** Capture HTTP status codes and request durations.
- **Concurrent execution:** Run multiple checks through a bounded worker pool.
- **Configurable concurrency:** Control the maximum number of simultaneous checks.
- **Timeout handling:** Configure HTTP client timeouts.
- **Cancellation support:** Propagate context cancellation to in-flight HTTP requests and stop dispatching additional work.
- **URL validation:** Reject invalid target URLs before execution.
- **File-based targets:** Load multiple URLs from a file, ignoring blank lines and comments.
- **Deterministic result ordering:** Return outcomes in the same order as the input targets.
- **Text and JSON reporting:** Support human-readable output and structured reports.
- **Automated testing:** Exercise checker, loader, and scheduler behavior.
- **Performance benchmarks:** Compare different concurrency levels using a controlled local HTTP server.

## Architecture

```text
                         CLI
                          |
             +------------+------------+
             |                         |
       URL / File Input          Configuration
             |                  (timeout, workers)
             +------------+------------+
                          |
                 Input Validation
                          |
                    Scheduler
                          |
                    Job Channel
                          |
             +------------+------------+
             |            |            |
          Worker 1     Worker 2      Worker N
             |            |            |
             +------------+------------+
                          |
                     HTTP Checker
                          |
                  Ordered Outcomes
                          |
                   Report Generation
                          |
                    Text / JSON
```

### Component responsibilities

| Component | Responsibility |
|---|---|
| `cmd/monitor` | CLI flags, validation, orchestration, exit status, output selection |
| `internal/checker` | HTTP requests, status classification, duration measurement, URL validation |
| `internal/targets` | Read and normalize targets from files |
| `internal/scheduler` | Bounded worker pool, job dispatch, cancellation, outcome collection |
| `internal/monitor` | Structured report and summary data models |

The scheduler and checker are separated so that execution policy can evolve without coupling the HTTP-checking logic to command-line presentation.

## Prerequisites

- Go installed at a version compatible with the project's `go.mod`.
- Git, if cloning the repository.

## Getting started

Clone the repository and enter the project directory:

```bash
git clone <YOUR_REPOSITORY_URL>
cd go-uptime-monitor
```

Run the test suite:

```bash
go test ./...
```

Build all packages:

```bash
go build ./...
```

## Usage

### Check a single URL

```bash
go run ./cmd/monitor -url https://example.com
```

### Configure the HTTP timeout

```bash
go run ./cmd/monitor \
  -url https://example.com \
  -timeout 5s
```

### Check multiple URLs

Create a file named `urls.txt`:

```text
# Public endpoints for local testing
https://example.com
https://www.google.com
https://www.github.com
```

Run the monitor:

```bash
go run ./cmd/monitor -file urls.txt
```

### Configure concurrency

```bash
go run ./cmd/monitor \
  -file urls.txt \
  -concurrency 5 \
  -timeout 5s
```

The concurrency setting controls how many checks may execute simultaneously. Higher concurrency can improve throughput for I/O-bound workloads, but it can also increase resource usage and load on monitored endpoints.

### Generate JSON output

```bash
go run ./cmd/monitor \
  -file urls.txt \
  -format json
```

JSON output is intended for automation and future integrations. It includes per-target results and aggregate counts.

The exact output schema is defined by the project's report model.

## Health classification

The checker classifies received HTTP status codes as follows:

| Condition | Classification |
|---|---|
| HTTP status 200–399 | Healthy |
| HTTP status outside 200–399 | Unhealthy |
| Request or transport failure | Error |

An HTTP response with an unhealthy status is different from a request that fails before a usable response is received. The report preserves this distinction.

## Exit codes

The CLI's intended exit-code contract is:

| Exit code | Meaning |
|---:|---|
| `0` | All checks are healthy |
| `1` | At least one check is unhealthy or has a request error |
| `2` | Invalid usage or configuration |

Use these exit codes in shell scripts or CI jobs to automate basic endpoint checks.

## Concurrency model

The scheduler uses a fixed worker pool and a channel-based job queue.

Instead of creating an unrestricted goroutine for every target:

1. The scheduler prepares indexed jobs.
2. Workers consume jobs from the queue.
3. Each worker executes an HTTP check.
4. Outcomes are stored against the original target index.
5. The caller receives the outcomes in input order.

This design separates execution order from presentation order. A slow request does not require subsequent results to be displayed out of sequence.

Context cancellation is propagated to HTTP requests. The scheduler also stops dispatching work when cancellation is observed. Cancellation is cooperative, so its exact behavior depends on whether a job has already been dispatched or a request is in flight.

## Testing

Run all tests:

```bash
go test ./...
```

Run tests with Go's race detector:

```bash
go test -race ./...
```

Run static analysis:

```bash
go vet ./...
```

Format the code:

```bash
gofmt -w .
```

Build all packages:

```bash
go build ./...
```

The test suite uses Go's testing facilities and local HTTP test servers to exercise behavior without depending on public endpoints for every test.

## Performance benchmarking

The scheduler benchmark compares concurrency levels of 1, 2, 5, and 10 workers. Each benchmark operation checks 20 URLs against a local HTTP test server with an artificial 5 ms delay.

Run the benchmark:

```bash
go test -run='^$' \
  -bench=BenchmarkRun \
  -benchmem \
  -benchtime=3s \
  ./internal/scheduler
```

### Baseline results

Environment: Apple M1, macOS, Go benchmark output from the local development environment.

| Concurrency | Batch time | Allocated bytes/op | Allocations/op |
|---:|---:|---:|---:|
| 1 | 121.73 ms | 116,949 B | 1,291 |
| 2 | 60.04 ms | 116,625 B | 1,293 |
| 5 | 25.14 ms | 169,087 B | 1,556 |
| 10 | 16.21 ms | 252,273 B | 1,982 |

In this controlled workload, moving from one worker to ten reduced batch execution time by approximately 7.5 times. Allocations per operation increased as concurrency rose.

These results illustrate the relationship between concurrency, latency, and resource consumption. They are a local benchmark baseline, not a guarantee of production performance. Real results depend on network latency, server behavior, connection reuse, workload size, and resource limits.

## Design decisions

### Why a bounded worker pool?

It makes concurrency an explicit resource policy instead of allowing the number of goroutines to grow with the number of targets.

### Why use contexts?

Contexts provide a standard way to propagate cancellation through request execution and allow the scheduler to stop dispatching additional work when cancellation is observed.

### Why preserve input order?

Stable ordering makes CLI output and JSON reports easier to read, compare, test, and consume from scripts—even when requests complete in a different order.

### Why distinguish HTTP failures from transport errors?

An HTTP response is a completed network exchange, even when the status indicates an unhealthy service. A transport error means the request could not produce a usable HTTP response. Keeping them separate improves diagnosis and reporting.

### Why benchmark instead of assuming?

Concurrency introduces scheduling and allocation overhead. Measuring a controlled workload helps identify the point at which additional parallelism provides diminishing returns.

## Current scope and limitations

This project is a CLI-based HTTP checker, not a full monitoring platform.

The current scope focuses on one-shot checks, concurrency control, cancellation, result reporting, and performance measurement.

Potential future extensions include:

- TLS certificate-expiration monitoring.
- Periodic checks and historical metrics.
- Bounded retries with backoff.
- Prometheus-compatible metrics.
- Persistent result storage and alerting.
- A restricted interactive demo using predefined targets.

## Project goals

This project is part of a hands-on Go engineering portfolio focused on:

- Concurrency and goroutine lifecycle management.
- Context propagation and timeout behavior.
- Separation of concerns and package design.
- Automated testing and race detection.
- Benchmark-driven engineering decisions.
- Reliable command-line interfaces and machine-readable output.

The emphasis is on making trade-offs explicit, validating behavior through tests, and establishing a foundation that can be extended without prematurely introducing distributed-system complexity.
