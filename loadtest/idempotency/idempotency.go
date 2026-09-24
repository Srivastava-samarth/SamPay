package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"sync"
	"time"
)

const (
	baseURL      = "http://localhost:8080/api/v1"
	concurrentRequests = 10
	paymentAmount      = 200
)

type AuthResponse struct {
	Success bool `json:"success"`
	Data    struct {
		AccessToken string `json:"access_token"`
	} `json:"data"`
}

type PaymentResponse struct {
	Success bool            `json:"success"`
	Data    json.RawMessage `json:"data"`
}

type Result struct {
	RequestNumber int
	StatusCode    int
	Success       bool
	Latency       time.Duration
	Body          string
	Error         error
}

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

	client := &http.Client{}

	accessToken, err := login(
		client,
		baseURL,
		email,
		password,
	)
	if err != nil {
		panic(fmt.Sprintf("login failed: %v", err))
	}

	fmt.Println("Login successful")
	fmt.Printf("Running %d concurrent requests with the SAME idempotency key\n", concurrentRequests)
	fmt.Printf("Payment amount: ₹%d\n", paymentAmount)

	// IMPORTANT:
	// Every request intentionally uses this exact same key.
	idempotencyKey := fmt.Sprintf(
		"concurrency-idempotency-test-%d",
		time.Now().UnixNano(),
	)

	fmt.Printf("Idempotency-Key: %s\n\n", idempotencyKey)

	start := make(chan struct{})
	ready := make(chan struct{}, concurrentRequests)

	results := make(chan Result, concurrentRequests)

	var wg sync.WaitGroup
	wg.Add(concurrentRequests)

	for i := 1; i <= concurrentRequests; i++ {
		requestNumber := i

		go func() {
			defer wg.Done()

			ready <- struct{}{}

			// Wait until all workers are ready.
			<-start

			result := executePayment(
				client,
				baseURL,
				accessToken,
				merchantID,
				receiverMerchantID,
				idempotencyKey,
				requestNumber,
			)

			results <- result
		}()
	}

	// Make sure all goroutines have reached the ready state.
	for i := 0; i < concurrentRequests; i++ {
		<-ready
	}

	fmt.Println("All requests ready. Starting simultaneously...")
	close(start)

	wg.Wait()
	close(results)

	fmt.Println("\n========== RESULTS ==========")

	successCount := 0
	failureCount := 0

	for result := range results {
		fmt.Printf(
			"\nRequest #%d\n",
			result.RequestNumber,
		)

		if result.Error != nil {
			fmt.Printf("Error: %v\n", result.Error)
			failureCount++
			continue
		}

		fmt.Printf("Status:  %d\n", result.StatusCode)
		fmt.Printf("Success: %v\n", result.Success)
		fmt.Printf("Latency: %v\n", result.Latency)
		fmt.Printf("Response: %s\n", result.Body)

		if result.Success {
			successCount++
		} else {
			failureCount++
		}
	}

	fmt.Println("\n========== SUMMARY ==========")
	fmt.Printf("Total requests: %d\n", concurrentRequests)
	fmt.Printf("Successful responses: %d\n", successCount)
	fmt.Printf("Failed responses: %d\n", failureCount)
	fmt.Printf("Same Idempotency-Key: %s\n", idempotencyKey)

	fmt.Println("\nExpected:")
	fmt.Println("- Exactly ONE actual payment should be created.")
	fmt.Println("- Duplicate requests should NOT create additional transactions.")
	fmt.Println("- Wallet should be debited only ONCE.")
	fmt.Println("- Ledger entries should be created only ONCE.")
}

func login(
	client *http.Client,
	baseURL string,
	email string,
	password string,
) (string, error) {
	payload := map[string]string{
		"email":    email,
		"password": password,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequest(
		http.MethodPost,
		baseURL+"/login",
		bytes.NewBuffer(body),
	)
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf(
			"login returned status %d",
			resp.StatusCode,
		)
	}

	var authResponse AuthResponse

	if err := json.NewDecoder(resp.Body).Decode(&authResponse); err != nil {
		return "", err
	}

	if !authResponse.Success {
		return "", fmt.Errorf("login response was unsuccessful")
	}

	if authResponse.Data.AccessToken == "" {
		return "", fmt.Errorf("access token missing from login response")
	}

	return authResponse.Data.AccessToken, nil
}

func executePayment(
	client *http.Client,
	baseURL string,
	accessToken string,
	merchantID string,
	receiverMerchantID string,
	idempotencyKey string,
	requestNumber int,
) Result {
	payload := map[string]interface{}{
		"amount":               paymentAmount,
		"currency":             "INR",
		"receiver_merchant_id": receiverMerchantID,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return Result{
			RequestNumber: requestNumber,
			Error:         err,
		}
	}

	url := fmt.Sprintf(
		"%s/%s/payment",
		baseURL,
		merchantID,
	)

	req, err := http.NewRequest(
		http.MethodPost,
		url,
		bytes.NewBuffer(body),
	)
	if err != nil {
		return Result{
			RequestNumber: requestNumber,
			Error:         err,
		}
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+accessToken)

	// CRITICAL:
	// Every concurrent request uses the SAME key.
	req.Header.Set("Idempotency-Key", idempotencyKey)

	start := time.Now()

	resp, err := client.Do(req)

	latency := time.Since(start)

	if err != nil {
		return Result{
			RequestNumber: requestNumber,
			Latency:       latency,
			Error:         err,
		}
	}

	defer resp.Body.Close()

	var responseBody bytes.Buffer

	_, err = responseBody.ReadFrom(resp.Body)
	if err != nil {
		return Result{
			RequestNumber: requestNumber,
			StatusCode:    resp.StatusCode,
			Latency:       latency,
			Error:         err,
		}
	}

	bodyString := responseBody.String()

	var paymentResponse PaymentResponse

	success := false

	if err := json.Unmarshal(
		[]byte(bodyString),
		&paymentResponse,
	); err == nil {
		success = paymentResponse.Success
	}

	return Result{
		RequestNumber: requestNumber,
		StatusCode:    resp.StatusCode,
		Success:       success,
		Latency:       latency,
		Body:          bodyString,
	}
}