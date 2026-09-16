package controllers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Srivastava-samarth/sampay/dto"
	"github.com/gin-gonic/gin"
)

func TestTriggerRecon(t *testing.T) {
	t.Run("successfully triggers recon workflow", func(t *testing.T) {
		cleanTestDB(t)

		router := gin.New()
		router.POST("/recon", testController.TriggerRecon())

		req := httptest.NewRequest(
			http.MethodPost,
			"/recon",
			nil,
		)

		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf(
				"expected status 200, got %d. response: %s",
				rec.Code,
				rec.Body.String(),
			)
		}

		var response dto.Response
		if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if !response.Success {
			t.Errorf("expected success=true")
		}

		if response.Data == nil {
			t.Fatalf("expected workflow ID in response")
		}

		workflowID, ok := response.Data.(string)
		if !ok {
			t.Fatalf(
				"expected workflow ID to be string, got %T",
				response.Data,
			)
		}

		if workflowID == "" {
			t.Errorf("expected workflow ID to be non-empty")
		}

		if !strings.HasPrefix(workflowID, "recon-") {
			t.Errorf(
				"expected workflow ID to start with 'recon-', got %s",
				workflowID,
			)
		}
	})
}
