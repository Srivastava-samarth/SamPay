package main

import (
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"sort"
	"sync"
	"sync/atomic"
	"time"
)

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func main() {
	baseURL := flag.String(
		"url",
		"http://localhost:8080",
		"SamPay base URL",
	)

	concurrency := flag.Int(
		"concurrency",
		1,
		"Number of concurrent workers",
	)

	duration := flag.Duration(
		"duration",
		30*time.Second,
		"Test duration",
	)

	flag.Parse()

	email := os.Getenv("SAMPAY_TEST_EMAIL")
	password := os.Getenv("SAMPAY_TEST_PASSWORD")

	if email == "" || password == "" {
		fmt.Println("Missing SAMPAY_TEST_EMAIL or SAMPAY_TEST_PASSWORD")
		os.Exit(1)
	}

	runLoadTest(
		*baseURL,
		email,
		password,
		*concurrency,
		*duration,
	)
}

func runLoadTest(
	baseURL string,
	email string,
	password string,
	concurrency int,
	duration time.Duration,
) {
	url := baseURL + "/api/v1/login"

	fmt.Println("===================================")
	fmt.Println("       SamPay Login Load Test      ")
	fmt.Println("===================================")
	fmt.Printf("URL:         %s\n", url)
	fmt.Printf("Concurrency: %d\n", concurrency)
	fmt.Printf("Duration:    %s\n", duration)
	fmt.Println()

	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	payload := LoginRequest{
		Email:    email,
		Password: password,
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		fmt.Printf("Failed to marshal request: %v\n", err)
		return
	}

	// This context controls how long workers are allowed
	// to START new requests.
	testCtx, cancel := context.WithTimeout(
		context.Background(),
		duration,
	)
	defer cancel()

	var totalRequests atomic.Int64
	var successfulRequests atomic.Int64
	var failedRequests atomic.Int64

	var latencyMu sync.Mutex
	latencies := make([]time.Duration, 0, 10000)

	var statusMu sync.Mutex
	statusCounts := make(map[int]int)

	var firstFailureOnce sync.Once

	var wg sync.WaitGroup

	start := time.Now()

	for worker := 0; worker < concurrency; worker++ {
		wg.Add(1)

		go func() {
			defer wg.Done()

			for {
				// Stop creating new requests once the
				// test duration has elapsed.
				select {
				case <-testCtx.Done():
					return
				default:
				}

				// IMPORTANT:
				// Do NOT attach testCtx to the HTTP request.
				//
				// An already-running request should be allowed
				// to finish even when the load-test duration ends.
				req, err := http.NewRequest(
					http.MethodPost,
					url,
					bytes.NewReader(payloadBytes),
				)

				if err != nil {
					totalRequests.Add(1)
					failedRequests.Add(1)

					firstFailureOnce.Do(func() {
						fmt.Printf(
							"\nFIRST REQUEST CREATION ERROR: %v\n\n",
							err,
						)
					})

					continue
				}

				req.Header.Set(
					"Content-Type",
					"application/json",
				)

				requestStart := time.Now()

				resp, err := client.Do(req)

				latency := time.Since(requestStart)

				if err != nil {
					totalRequests.Add(1)
					failedRequests.Add(1)

					firstFailureOnce.Do(func() {
						fmt.Printf(
							"\nFIRST REQUEST ERROR: %v\n\n",
							err,
						)
					})

					latencyMu.Lock()
					latencies = append(latencies, latency)
					latencyMu.Unlock()

					continue
				}

				responseBody, _ := io.ReadAll(resp.Body)
				resp.Body.Close()

				totalRequests.Add(1)

				statusMu.Lock()
				statusCounts[resp.StatusCode]++
				statusMu.Unlock()

				if resp.StatusCode == http.StatusOK {
					successfulRequests.Add(1)
				} else {
					failedRequests.Add(1)

					firstFailureOnce.Do(func() {
						fmt.Println(
							"\n========== FIRST FAILED REQUEST ==========",
						)
						fmt.Printf(
							"HTTP Status: %d (%s)\n",
							resp.StatusCode,
							resp.Status,
						)
						fmt.Printf(
							"Response:    %s\n",
							string(responseBody),
						)
						fmt.Println(
							"===========================================\n",
						)
					})
				}

				latencyMu.Lock()
				latencies = append(latencies, latency)
				latencyMu.Unlock()
			}
		}()
	}

	// Workers stop starting new requests when testCtx expires.
	// Any request already in progress is allowed to finish.
	wg.Wait()

	elapsed := time.Since(start)

	total := totalRequests.Load()
	successful := successfulRequests.Load()
	failed := failedRequests.Load()

	var average time.Duration

	latencyMu.Lock()

	if len(latencies) > 0 {
		sort.Slice(
			latencies,
			func(i, j int) bool {
				return latencies[i] < latencies[j]
			},
		)

		var totalLatency time.Duration

		for _, latency := range latencies {
			totalLatency += latency
		}

		average = totalLatency / time.Duration(
			len(latencies),
		)
	}

	p50 := percentile(latencies, 0.50)
	p95 := percentile(latencies, 0.95)
	p99 := percentile(latencies, 0.99)

	latencyMu.Unlock()

	fmt.Println("============== Results ==============")
	fmt.Printf("Requests:     %d\n", total)
	fmt.Printf("Successful:   %d\n", successful)
	fmt.Printf("Failed:       %d\n", failed)

	if elapsed > 0 {
		fmt.Printf(
			"RPS:          %.2f\n",
			float64(total)/elapsed.Seconds(),
		)
	}

	fmt.Printf("Average:      %s\n", average)
	fmt.Printf("P50:          %s\n", p50)
	fmt.Printf("P95:          %s\n", p95)
	fmt.Printf("P99:          %s\n", p99)

	fmt.Println()
	fmt.Println("HTTP Status Codes:")

	statusMu.Lock()

	statusCodes := make([]int, 0, len(statusCounts))

	for statusCode := range statusCounts {
		statusCodes = append(statusCodes, statusCode)
	}

	sort.Ints(statusCodes)

	for _, statusCode := range statusCodes {
		fmt.Printf(
			"%d: %d\n",
			statusCode,
			statusCounts[statusCode],
		)
	}

	statusMu.Unlock()

	fmt.Println("====================================")
}

func percentile(
	latencies []time.Duration,
	percent float64,
) time.Duration {
	if len(latencies) == 0 {
		return 0
	}

	index := int(
		float64(len(latencies)-1) * percent,
	)

	return latencies[index]
}
