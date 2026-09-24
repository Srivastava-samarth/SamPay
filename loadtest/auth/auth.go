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

	"github.com/Srivastava-samarth/sampay/config"
	models "github.com/Srivastava-samarth/sampay/database/models"
	"github.com/Srivastava-samarth/sampay/middlewares"
	"github.com/Srivastava-samarth/sampay/utils"
	"github.com/google/uuid"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type LoginRequest struct {
	Email    *string `json:"email"`
	Password *string `json:"password"`
}

type RefreshTokenRequest struct {
	RefreshToken *string `json:"refresh_token"`
}

type LoginResponse struct {
	Data AuthResponse `json:"data"`
}

type AuthResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int    `json:"expires_in"`
}

type ForgotPasswordRequest struct {
	Email *string `json:"email"`
}

type ResetPasswordRequest struct {
	Email           *string `json:"email"`
	ResetToken      *string `json:"reset_token"`
	NewPassword     *string `json:"new_password"`
	ConfirmPassword *string `json:"confirm_password"`
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

	wrongPassword := flag.Bool(
		"wrong-password",
		false,
		"Use an intentionally incorrect password",
	)

	refreshTokenMode := flag.Bool(
		"refresh-token",
		false,
		"Run refresh-token load test",
	)

	forgotPasswordMode := flag.Bool(
		"forgot-password",
		false,
		"Run forgot-password load test",
	)

	resetPasswordMode := flag.Bool(
		"reset-password",
		false,
		"Run reset-password load test",
	)

	flag.Parse()

	email := os.Getenv("SAMPAY_TEST_EMAIL")
	password := os.Getenv("SAMPAY_TEST_PASSWORD")

	if email == "" || password == "" {
		fmt.Println("Missing SAMPAY_TEST_EMAIL or SAMPAY_TEST_PASSWORD")
		os.Exit(1)
	}

	if *refreshTokenMode {
		runRefreshTokenLoadTest(
			*baseURL,
			email,
			password,
			*concurrency,
			*duration,
		)
		return
	}

	if *forgotPasswordMode {
		runForgotPasswordLoadTest(
			*baseURL,
			email,
			*concurrency,
			*duration,
		)
		return
	}

	if *resetPasswordMode {
		runResetPasswordLoadTest(
			*baseURL,
			email,
			password,
			*concurrency,
			*duration,
		)
		return
	}

	testPassword := password

	if *wrongPassword {
		testPassword = password + "-wrong"
	}

	runLoginLoadTest(
		*baseURL,
		email,
		testPassword,
		*concurrency,
		*duration,
		*wrongPassword,
	)
}

func runLoginLoadTest(
	baseURL string,
	email string,
	password string,
	concurrency int,
	duration time.Duration,
	wrongPassword bool,
) {
	url := baseURL + "/api/v1/login"

	fmt.Println("===================================")
	fmt.Println("       SamPay Login Load Test      ")
	fmt.Println("===================================")
	fmt.Printf("URL:         %s\n", url)
	fmt.Printf("Concurrency: %d\n", concurrency)
	fmt.Printf("Duration:    %s\n", duration)

	if wrongPassword {
		fmt.Println("Mode:        WRONG PASSWORD")
	} else {
		fmt.Println("Mode:        VALID PASSWORD")
	}

	fmt.Println()

	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	payload := LoginRequest{
		Email:    &email,
		Password: &password,
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		fmt.Printf("Failed to marshal request: %v\n", err)
		return
	}

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
				select {
				case <-testCtx.Done():
					return
				default:
				}

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
				if err := resp.Body.Close(); err != nil {
					totalRequests.Add(1)
					failedRequests.Add(1)

					latencyMu.Lock()
					latencies = append(latencies, latency)
					latencyMu.Unlock()

					continue
				}

				totalRequests.Add(1)

				statusMu.Lock()
				statusCounts[resp.StatusCode]++
				statusMu.Unlock()

				expectedStatus := http.StatusOK

				if wrongPassword {
					expectedStatus = http.StatusUnauthorized
				}

				if resp.StatusCode == expectedStatus {
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
					})
				}

				latencyMu.Lock()
				latencies = append(latencies, latency)
				latencyMu.Unlock()
			}
		}()
	}

	wg.Wait()

	elapsed := time.Since(start)

	printResults(
		totalRequests.Load(),
		successfulRequests.Load(),
		failedRequests.Load(),
		elapsed,
		latencies,
		&latencyMu,
		&statusMu,
		statusCounts,
	)
}

func runRefreshTokenLoadTest(
	baseURL string,
	email string,
	password string,
	concurrency int,
	duration time.Duration,
) {
	fmt.Println("===================================")
	fmt.Println("   SamPay Refresh Token Load Test  ")
	fmt.Println("===================================")

	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	// Generate one valid refresh token before the benchmark.
	refreshToken, err := getRefreshToken(
		baseURL,
		email,
		password,
	)

	if err != nil {
		fmt.Printf("Failed to generate refresh token: %v\n", err)
		return
	}

	url := baseURL + "/api/v1/refresh-token"

	fmt.Printf("URL:         %s\n", url)
	fmt.Printf("Concurrency: %d\n", concurrency)
	fmt.Printf("Duration:    %s\n", duration)
	fmt.Println("Mode:        VALID REFRESH TOKEN")
	fmt.Println("Token:       ONE SHARED TOKEN")
	fmt.Println()

	payload := RefreshTokenRequest{
		RefreshToken: &refreshToken,
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		fmt.Printf("Failed to marshal request: %v\n", err)
		return
	}

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
				select {
				case <-testCtx.Done():
					return
				default:
				}

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
				if err := resp.Body.Close(); err != nil {
					totalRequests.Add(1)
					failedRequests.Add(1)

					latencyMu.Lock()
					latencies = append(latencies, latency)
					latencyMu.Unlock()

					continue
				}

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
					})
				}

				latencyMu.Lock()
				latencies = append(latencies, latency)
				latencyMu.Unlock()
			}
		}()
	}

	wg.Wait()

	elapsed := time.Since(start)

	printResults(
		totalRequests.Load(),
		successfulRequests.Load(),
		failedRequests.Load(),
		elapsed,
		latencies,
		&latencyMu,
		&statusMu,
		statusCounts,
	)
}

func prepareResetTokens(
	email string,
	count int,
) ([]string, error) {
	cfg, err := config.Load()
	if err != nil {
		return nil, fmt.Errorf(
			"failed to load config: %w",
			err,
		)
	}

	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s TimeZone=%s",
		cfg.Database.Host,
		cfg.Database.Port,
		cfg.Database.User,
		cfg.Database.Password,
		cfg.Database.Name,
		cfg.Database.SSLMode,
		cfg.Database.TimeZone,
	)

	db, err := gorm.Open(
		postgres.Open(dsn),
		&gorm.Config{},
	)
	if err != nil {
		return nil, fmt.Errorf(
			"failed to connect to database: %w",
			err,
		)
	}

	var user models.User

	if err := db.
		Where("email = ?", email).
		First(&user).
		Error; err != nil {
		return nil, fmt.Errorf(
			"failed to find test user: %w",
			err,
		)
	}

	jwtService := &middlewares.Jwt{
		Config: &cfg.JWT,
	}

	tokens := make([]string, 0, count)

	for i := 0; i < count; i++ {
		token, err := jwtService.GenerateResetPasswordToken(
			user.ID,
			user.Email,
		)
		if err != nil {
			return nil, fmt.Errorf(
				"failed to generate reset token %d: %w",
				i,
				err,
			)
		}

		hashedToken := utils.HashToken(token)

		passwordReset := &models.PasswordResetToken{
			ID:        uuid.New(),
			UserID:    user.ID,
			TokenHash: hashedToken,
			ExpiresAt: time.Now().Add(time.Hour),
			CreatedAt: time.Now(),
		}

		if err := db.Create(passwordReset).Error; err != nil {
			return nil, fmt.Errorf(
				"failed to create reset token %d: %w",
				i,
				err,
			)
		}

		tokens = append(tokens, token)
	}

	return tokens, nil
}

func getRefreshToken(baseURL, email, password string) (string, error) {
	loginRequest := LoginRequest{
		Email:    &email,
		Password: &password,
	}

	body, err := json.Marshal(loginRequest)
	if err != nil {
		return "", fmt.Errorf("failed to marshal login request: %w", err)
	}

	req, err := http.NewRequest(
		http.MethodPost,
		baseURL+"/api/v1/login",
		bytes.NewReader(body),
	)
	if err != nil {
		return "", fmt.Errorf("failed to create login request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("login request failed: %w", err)
	}
	if err := resp.Body.Close(); err != nil {
		return "", fmt.Errorf("login request failed: %w", err)
	}

	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read login response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf(
			"login failed with status %d: %s",
			resp.StatusCode,
			string(responseBody),
		)
	}

	var loginResponse LoginResponse

	if err := json.Unmarshal(responseBody, &loginResponse); err != nil {
		return "", fmt.Errorf(
			"failed to decode login response: %w; body=%s",
			err,
			string(responseBody),
		)
	}

	if loginResponse.Data.RefreshToken == "" {
		return "", fmt.Errorf(
			"login response did not contain a refresh token; body=%s",
			string(responseBody),
		)
	}

	return loginResponse.Data.RefreshToken, nil
}

func printResults(
	total int64,
	successful int64,
	failed int64,
	elapsed time.Duration,
	latencies []time.Duration,
	latencyMu *sync.Mutex,
	statusMu *sync.Mutex,
	statusCounts map[int]int,
) {
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

		average = totalLatency / time.Duration(len(latencies))
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

func runForgotPasswordLoadTest(
	baseURL string,
	email string,
	concurrency int,
	duration time.Duration,
) {
	url := baseURL + "/api/v1/forgot-password"

	fmt.Println("===================================")
	fmt.Println("   SamPay Forgot Password Load Test")
	fmt.Println("===================================")
	fmt.Printf("URL:         %s\n", url)
	fmt.Printf("Concurrency: %d\n", concurrency)
	fmt.Printf("Duration:    %s\n", duration)
	fmt.Println("Mode:        VALID EMAIL")
	fmt.Println()

	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	payload := ForgotPasswordRequest{
		Email: &email,
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		fmt.Printf("Failed to marshal request: %v\n", err)
		return
	}

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
				select {
				case <-testCtx.Done():
					return
				default:
				}

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
				if err := resp.Body.Close(); err != nil {
					totalRequests.Add(1)
					failedRequests.Add(1)

					latencyMu.Lock()
					latencies = append(latencies, latency)
					latencyMu.Unlock()

					continue
				}

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
					})
				}

				latencyMu.Lock()
				latencies = append(latencies, latency)
				latencyMu.Unlock()
			}
		}()
	}

	wg.Wait()

	elapsed := time.Since(start)

	printResults(
		totalRequests.Load(),
		successfulRequests.Load(),
		failedRequests.Load(),
		elapsed,
		latencies,
		&latencyMu,
		&statusMu,
		statusCounts,
	)
}

func runResetPasswordLoadTest(
	baseURL string,
	email string,
	password string,
	concurrency int,
	duration time.Duration,
) {
	url := baseURL + "/api/v1/reset-password"

	fmt.Println("===================================")
	fmt.Println("   SamPay Reset Password Load Test ")
	fmt.Println("===================================")
	fmt.Printf("URL:         %s\n", url)
	fmt.Printf("Concurrency: %d\n", concurrency)
	fmt.Printf("Duration:    %s\n", duration)
	fmt.Println("Mode:        UNIQUE VALID RESET TOKENS")
	fmt.Println()

	tokenCount := concurrency * int(duration.Seconds()) * 10

	if tokenCount < 100 {
		tokenCount = 100
	}

	fmt.Printf(
		"Preparing %d reset tokens...\n",
		tokenCount,
	)

	tokens, err := prepareResetTokens(
		email,
		tokenCount,
	)
	if err != nil {
		fmt.Printf(
			"Failed to prepare reset tokens: %v\n",
			err,
		)
		return
	}

	fmt.Printf(
		"Prepared %d reset tokens\n\n",
		len(tokens),
	)

	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	testCtx, cancel := context.WithTimeout(
		context.Background(),
		duration,
	)
	defer cancel()

	var tokenIndex atomic.Int64

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
				select {
				case <-testCtx.Done():
					return

				default:
				}

				index := int(
					tokenIndex.Add(1) - 1,
				)

				if index >= len(tokens) {
					return
				}

				token := tokens[index]

				// Use the original test password so
				// the test account remains usable.
				newPassword := password

				payload := ResetPasswordRequest{
					Email:           &email,
					ResetToken:      &token,
					NewPassword:     &newPassword,
					ConfirmPassword: &newPassword,
				}

				payloadBytes, err := json.Marshal(payload)

				if err != nil {
					totalRequests.Add(1)
					failedRequests.Add(1)

					firstFailureOnce.Do(func() {
						fmt.Printf(
							"\nFIRST JSON ERROR: %v\n\n",
							err,
						)
					})

					continue
				}

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

					latencies = append(
						latencies,
						latency,
					)

					latencyMu.Unlock()

					continue
				}

				responseBody, _ := io.ReadAll(
					resp.Body,
				)

				if err := resp.Body.Close(); err != nil {
					totalRequests.Add(1)
					failedRequests.Add(1)

					latencyMu.Lock()
					latencies = append(latencies, latency)
					latencyMu.Unlock()

					continue
				}

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
					})
				}

				latencyMu.Lock()

				latencies = append(
					latencies,
					latency,
				)

				latencyMu.Unlock()
			}
		}()
	}

	wg.Wait()

	elapsed := time.Since(start)

	printResults(
		totalRequests.Load(),
		successfulRequests.Load(),
		failedRequests.Load(),
		elapsed,
		latencies,
		&latencyMu,
		&statusMu,
		statusCounts,
	)
}
