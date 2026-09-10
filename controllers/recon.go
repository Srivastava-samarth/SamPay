package controllers

import (
	"fmt"
	"net/http"
	"time"

	"github.com/Srivastava-samarth/sampay/dto"
	"github.com/Srivastava-samarth/sampay/temporal"
	"github.com/Srivastava-samarth/sampay/temporal/workflows"
	"github.com/gin-gonic/gin"
	"go.temporal.io/sdk/client"
)

type ReconController struct {
	TemporalClient client.Client
}

func NewReconController(
	temporalClient client.Client,
) *ReconController{
	return &ReconController{
		TemporalClient: temporalClient,
	}
}

func (rc *ReconController) TriggerRecon() gin.HandlerFunc{
	return func(c *gin.Context) {
		workflowOptions := client.StartWorkflowOptions{
			ID: fmt.Sprintf(
				"recon-%s",
				time.Now().UTC().Format("20060102-150405"),
			),
			TaskQueue: temporal.ReconFlowTaskQueue,
		}

		_, errW := rc.TemporalClient.ExecuteWorkflow(
			c.Request.Context(),
			workflowOptions,
			workflows.ReconFlow,
		)

		if errW != nil {
			dto.Fail(
				c,
				http.StatusInternalServerError,
				"RECON_FAILED",
				errW.Error(),
			)
			return
		}

		dto.Respond(
			c,
			http.StatusOK,
			workflowOptions.ID,
		)
	}
}