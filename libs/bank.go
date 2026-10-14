package libs

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"

	models "github.com/Srivastava-samarth/sampay/database/models"
	"github.com/Srivastava-samarth/sampay/dto"
	"github.com/google/uuid"
)

func (l *Libs) GetBankAccount(
	ctx context.Context,
	bankAccountID uuid.UUID,
) (*models.BankAccount, error) {
	if bankAccountID == uuid.Nil {
		return nil, fmt.Errorf("bank account ID cannot be empty")
	}

	baseURL := strings.TrimRight(l.Config.SampayBankURL, "/")
	if baseURL == "" {
		return nil, fmt.Errorf("sampay bank URL is not configured")
	}

	endpoint := fmt.Sprintf(
		"%s/api/v1/bank-accounts/%s",
		baseURL,
		url.PathEscape(bankAccountID.String()),
	)

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodGet,
		endpoint,
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf("create bank account request: %w", err)
	}

	req.Header.Set("X-API-Key", l.Config.XAPIKey)
	req.Header.Set("Accept", "application/json")

	httpClient := &http.Client{
		Timeout: 30 * time.Second,
	}

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request bank account from sampay-bank: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read sampay-bank response: %w", err)
	}

	log.Printf(
		"sampay-bank response: status=%d content-type=%q body=%s",
		resp.StatusCode,
		resp.Header.Get("Content-Type"),
		string(body),
	)

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))

		return nil, fmt.Errorf(
			"sampay-bank returned status %d: %s",
			resp.StatusCode,
			strings.TrimSpace(string(body)),
		)
	}
	var bankAccount models.BankAccount

	if err := json.Unmarshal(body, &bankAccount); err != nil {
		return nil, fmt.Errorf("decode bank account response: %w", err)
	}
	return &bankAccount, nil
}

func (l *Libs) UpdateBankAccountStatus(
	ctx context.Context,
	bankAccountID uuid.UUID,
	request *dto.UpdateBankAccountStatus,
) (*models.BankAccount, error) {
	if bankAccountID == uuid.Nil {
		return nil, fmt.Errorf("bank account ID cannot be empty")
	}

	baseURL := strings.TrimRight(l.Config.SampayBankURL, "/")
	if baseURL == "" {
		return nil, fmt.Errorf("sampay bank URL is not configured")
	}

	endpoint := fmt.Sprintf(
		"%s/api/v1/bank-accounts/%s/status",
		baseURL,
		url.PathEscape(bankAccountID.String()),
	)

	requestBody, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("encode bank account status request: %w", err)
	}

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPatch,
		endpoint,
		strings.NewReader(string(requestBody)),
	)
	if err != nil {
		return nil, fmt.Errorf("create bank account request: %w", err)
	}

	req.Header.Set("X-API-Key", l.Config.XAPIKey)
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")

	httpClient := &http.Client{
		Timeout: 30 * time.Second,
	}

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request bank account from sampay-bank: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read sampay-bank response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))

		return nil, fmt.Errorf(
			"sampay-bank returned status %d: %s",
			resp.StatusCode,
			strings.TrimSpace(string(body)),
		)
	}
	var bankAccount models.BankAccount

	if err := json.Unmarshal(body, &bankAccount); err != nil {
		return nil, fmt.Errorf("decode bank account response: %w", err)
	}
	return &bankAccount, nil
}

func (l *Libs) UpdateBankAccountBalance(
	ctx context.Context,
	bankAccountID uuid.UUID,
	request *dto.UpdateBankAccountBalance,
) (*models.BankAccount, error) {
	if bankAccountID == uuid.Nil {
		return nil, fmt.Errorf("bank account ID cannot be empty")
	}

	baseURL := strings.TrimRight(l.Config.SampayBankURL, "/")
	if baseURL == "" {
		return nil, fmt.Errorf("sampay bank URL is not configured")
	}

	endpoint := fmt.Sprintf(
		"%s/api/v1/bank-accounts/%s/balance",
		baseURL,
		url.PathEscape(bankAccountID.String()),
	)

	requestBody, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("encode bank account status request: %w", err)
	}

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPatch,
		endpoint,
		strings.NewReader(string(requestBody)),
	)
	if err != nil {
		return nil, fmt.Errorf("create bank account request: %w", err)
	}

	req.Header.Set("X-API-Key", l.Config.XAPIKey)
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")

	httpClient := &http.Client{
		Timeout: 30 * time.Second,
	}

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request bank account from sampay-bank: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read sampay-bank response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))

		return nil, fmt.Errorf(
			"sampay-bank returned status %d: %s",
			resp.StatusCode,
			strings.TrimSpace(string(body)),
		)
	}
	var bankAccount models.BankAccount

	if err := json.Unmarshal(body, &bankAccount); err != nil {
		return nil, fmt.Errorf("decode bank account response: %w", err)
	}
	return &bankAccount, nil
}

