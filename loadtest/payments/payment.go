package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"sync"
	"sync/atomic"
	"time"

	"github.com/google/uuid"
)

const (
	baseURL  = "http://localhost:8080/api/v1"
	rps      = 2
	duration = 30 * time.Second
)

type AuthRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type AuthResponse struct {
    Success bool `json:"success"`
    Data    struct {
        AccessToken  string `json:"access_token"`
        RefreshToken string `json:"refresh_token"`
        ExpiresIn    int    `json:"expires_in"`
    } `json:"data"`
}

type CreatePaymentRequest struct {
	ReceiverMerchantID uuid.UUID `json:"receiver_merchant_id"`
	Amount             string    `json:"amount"`
	Currency           string    `json:"currency"`
	Description        string    `json:"description"`
}

var (
	totalRequests uint64
	successCount  uint64
	errorCount    uint64

	mu        sync.Mutex
	latencies []time.Duration
)

func main() {
	email := os.Getenv("SAMPAY_EMAIL")
	password := os.Getenv("SAMPAY_PASSWORD")
	merchantID := os.Getenv("SAMPAY_MERCHANT_ID")
	receiverMerchantID := os.Getenv("SAMPAY_RECEIVER_MERCHANT_ID")

	if email == "" || password == "" {
		log.Fatal("SAMPAY_EMAIL and SAMPAY_PASSWORD are required")
	}

	if merchantID == "" || receiverMerchantID == "" {
		log.Fatal("SAMPAY_MERCHANT_ID and SAMPAY_RECEIVER_MERCHANT_ID are required")
	}

	receiverID, err := uuid.Parse(receiverMerchantID)
	if err != nil {
		log.Fatalf("invalid receiver merchant ID: %v", err)
	}

	client := &http.Client{
		Timeout: 60 * time.Second,
	}

	// Authenticate once.
	accessToken, err := login(client, email, password)
	if err != nil {
		log.Fatalf("login failed: %v", err)
	}

	fmt.Println("Login successful.")
	fmt.Printf("Starting payment load test: %d RPS for %s\n", rps, duration)

	runPaymentLoadTest(
		client,
		accessToken,
		merchantID,
		receiverID,
	)

	printResults()
}

func login(
	client *http.Client,
	email string,
	password string,
) (string, error) {
	payload := AuthRequest{
		Email:    email,
		Password: password,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("marshal auth request: %w", err)
	}

	req, err := http.NewRequest(
		http.MethodPost,
		baseURL+"/login",
		bytes.NewReader(body),
	)
	if err != nil {
		return "", fmt.Errorf("create login request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("execute login request: %w", err)
	}

	defer resp.Body.Close()

	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("read login response: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf(
			"login returned status %d: %s",
			resp.StatusCode,
			string(responseBody),
		)
	}

	var authResponse AuthResponse

	if err := json.Unmarshal(responseBody, &authResponse); err != nil {
		return "", fmt.Errorf(
			"unmarshal login response: %w",
			err,
		)
	}

	if authResponse.Data.AccessToken == "" {
		return "", fmt.Errorf("login response contains no access token")
	}

	return authResponse.Data.AccessToken, nil
}

func runPaymentLoadTest(
	client *http.Client,
	accessToken string,
	merchantID string,
	receiverMerchantID uuid.UUID,
) {
	start := time.Now()
	end := start.Add(duration)

	interval := time.Second / time.Duration(rps)

	var wg sync.WaitGroup

	for time.Now().Before(end) {
		wg.Add(1)

		go func() {
			defer wg.Done()

			runPayment(
				client,
				accessToken,
				merchantID,
				receiverMerchantID,
			)
		}()

		time.Sleep(interval)
	}

	wg.Wait()

	fmt.Printf(
		"Load test finished in %s\n",
		time.Since(start),
	)
}

func runPayment(
	client *http.Client,
	accessToken string,
	merchantID string,
	receiverMerchantID uuid.UUID,
) {
	atomic.AddUint64(&totalRequests, 1)

	payload := CreatePaymentRequest{
		ReceiverMerchantID: receiverMerchantID,
		Amount:             "10",
		Currency:           "INR",
		Description:        "SamPay load test",
	}

	body, err := json.Marshal(payload)
	if err != nil {
		atomic.AddUint64(&errorCount, 1)
		return
	}

	url := fmt.Sprintf(
		"%s/%s/payment",
		baseURL,
		merchantID,
	)

	req, err := http.NewRequest(
		http.MethodPost,
		url,
		bytes.NewReader(body),
	)
	if err != nil {
		atomic.AddUint64(&errorCount, 1)
		return
	}

	req.Header.Set(
		"Authorization",
		"Bearer "+accessToken,
	)

	req.Header.Set(
		"Content-Type",
		"application/json",
	)

	// SamPay requires UUID-only idempotency keys.
	req.Header.Set(
		"Idempotency-Key",
		uuid.New().String(),
	)

	start := time.Now()

	resp, err := client.Do(req)

	latency := time.Since(start)

	mu.Lock()
	latencies = append(latencies, latency)
	mu.Unlock()

	if err != nil {
		atomic.AddUint64(&errorCount, 1)

		fmt.Printf(
			"ERROR latency=%s error=%v\n",
			latency,
			err,
		)

		return
	}

	defer resp.Body.Close()

	_, _ = io.Copy(io.Discard, resp.Body)

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		atomic.AddUint64(&successCount, 1)

		fmt.Printf(
			"SUCCESS status=%d latency=%s\n",
			resp.StatusCode,
			latency,
		)

		return
	}

	atomic.AddUint64(&errorCount, 1)

	fmt.Printf(
		"FAILED status=%d latency=%s\n",
		resp.StatusCode,
		latency,
	)
}

func printResults() {
	mu.Lock()
	defer mu.Unlock()

	fmt.Println()
	fmt.Println("========== LOAD TEST RESULT ==========")

	fmt.Printf("Total Requests: %d\n", totalRequests)
	fmt.Printf("Successful:     %d\n", successCount)
	fmt.Printf("Failed:         %d\n", errorCount)

	if len(latencies) == 0 {
		fmt.Println("No latency data.")
		return
	}

	fmt.Printf("Min Latency:    %s\n", minLatency())
	fmt.Printf("Max Latency:    %s\n", maxLatency())
	fmt.Printf("Avg Latency:    %s\n", avgLatency())

	fmt.Println("======================================")
}

func minLatency() time.Duration {
	min := latencies[0]

	for _, latency := range latencies {
		if latency < min {
			min = latency
		}
	}

	return min
}

func maxLatency() time.Duration {
	max := latencies[0]

	for _, latency := range latencies {
		if latency > max {
			max = latency
		}
	}

	return max
}

func avgLatency() time.Duration {
	var total time.Duration

	for _, latency := range latencies {
		total += latency
	}

	return total / time.Duration(len(latencies))
}