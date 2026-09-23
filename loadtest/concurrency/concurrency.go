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
	"time"

	"github.com/google/uuid"
)

const (
	defaultBaseURL      = "http://localhost:8080/api/v1"
	concurrentRequests  = 10
	paymentAmount       = "2000"
	paymentCurrency     = "INR"
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

type PaymentResult struct {
	RequestNumber  int
	IdempotencyKey string
	StatusCode     int
	Success        bool
	Latency        time.Duration
	ResponseBody   string
	Error          error
}

func main() {
	baseURL := os.Getenv("SAMPAY_BASE_URL")
	if baseURL == "" {
		baseURL = defaultBaseURL
	}

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

	accessToken, err := login(client, baseURL, email, password)
	if err != nil {
		log.Fatalf("login failed: %v", err)
	}

	fmt.Println("Login successful.")
	fmt.Printf(
		"Starting concurrency test: %d requests simultaneously\n",
		concurrentRequests,
	)
	fmt.Printf("Payment amount per request: %s %s\n", paymentAmount, paymentCurrency)
	fmt.Println()

	results := runConcurrencyTest(
		client,
		baseURL,
		accessToken,
		merchantID,
		receiverID,
	)

	printResults(results)
}

func login(
	client *http.Client,
	baseURL string,
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
		return "", fmt.Errorf("unmarshal login response: %w", err)
	}

	if authResponse.Data.AccessToken == "" {
		return "", fmt.Errorf("login response contains no access token")
	}

	return authResponse.Data.AccessToken, nil
}

func runConcurrencyTest(
	client *http.Client,
	baseURL string,
	accessToken string,
	merchantID string,
	receiverMerchantID uuid.UUID,
) []PaymentResult {
	start := make(chan struct{})
	ready := make(chan struct{}, concurrentRequests)
	results := make(chan PaymentResult, concurrentRequests)

	var wg sync.WaitGroup

	for i := 1; i <= concurrentRequests; i++ {
		wg.Add(1)

		requestNumber := i

		go func() {
			defer wg.Done()

			// Tell main goroutine that this worker is ready.
			ready <- struct{}{}

			// Wait until every worker is ready.
			<-start

			result := sendPayment(
				client,
				baseURL,
				accessToken,
				merchantID,
				receiverMerchantID,
				requestNumber,
			)

			results <- result
		}()
	}

	// Wait until all workers are ready.
	for i := 0; i < concurrentRequests; i++ {
		<-ready
	}

	fmt.Println("All workers ready.")
	fmt.Println("Releasing requests simultaneously...")
	fmt.Println()

	// Release all requests at approximately the same time.
	close(start)

	wg.Wait()
	close(results)

	var output []PaymentResult

	for result := range results {
		output = append(output, result)
	}

	return output
}

func sendPayment(
	client *http.Client,
	baseURL string,
	accessToken string,
	merchantID string,
	receiverMerchantID uuid.UUID,
	requestNumber int,
) PaymentResult {
	idempotencyKey := uuid.New().String()

	payload := CreatePaymentRequest{
		ReceiverMerchantID: receiverMerchantID,
		Amount:             paymentAmount,
		Currency:           paymentCurrency,
		Description:        "SamPay concurrency test",
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return PaymentResult{
			RequestNumber:  requestNumber,
			IdempotencyKey: idempotencyKey,
			Error:          err,
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
		bytes.NewReader(body),
	)
	if err != nil {
		return PaymentResult{
			RequestNumber:  requestNumber,
			IdempotencyKey: idempotencyKey,
			Error:          err,
		}
	}

	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Idempotency-Key", idempotencyKey)

	start := time.Now()

	resp, err := client.Do(req)

	latency := time.Since(start)

	if err != nil {
		return PaymentResult{
			RequestNumber:  requestNumber,
			IdempotencyKey: idempotencyKey,
			Latency:        latency,
			Error:          err,
		}
	}

	defer resp.Body.Close()

	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return PaymentResult{
			RequestNumber:  requestNumber,
			IdempotencyKey: idempotencyKey,
			StatusCode:     resp.StatusCode,
			Latency:        latency,
			Error:          err,
		}
	}

	// Try to understand the API's success field.
	var apiResponse struct {
		Success bool `json:"success"`
	}

	_ = json.Unmarshal(responseBody, &apiResponse)

	return PaymentResult{
		RequestNumber:  requestNumber,
		IdempotencyKey: idempotencyKey,
		StatusCode:     resp.StatusCode,
		Success:        apiResponse.Success,
		Latency:        latency,
		ResponseBody:   string(responseBody),
	}
}

func printResults(results []PaymentResult) {
	fmt.Println("========== CONCURRENCY TEST RESULT ==========")

	successCount := 0
	failureCount := 0

	for _, result := range results {
		if result.Error != nil {
			fmt.Printf(
				"Request #%d | ERROR | latency=%s | error=%v\n",
				result.RequestNumber,
				result.Latency,
				result.Error,
			)

			failureCount++
			continue
		}

		if result.Success {
			successCount++
		} else {
			failureCount++
		}

		fmt.Printf(
			"Request #%d | status=%d | success=%t | latency=%s\n",
			result.RequestNumber,
			result.StatusCode,
			result.Success,
			result.Latency,
		)

		fmt.Printf(
			"  Idempotency-Key: %s\n",
			result.IdempotencyKey,
		)

		fmt.Printf(
			"  Response: %s\n",
			result.ResponseBody,
		)

		fmt.Println()
	}

	fmt.Println("---------------------------------------------")
	fmt.Printf("Total Requests: %d\n", len(results))
	fmt.Printf("Successful:     %d\n", successCount)
	fmt.Printf("Failed:         %d\n", failureCount)
	fmt.Println("=============================================")
}