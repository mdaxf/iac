package jobqueue

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/mdaxf/iac/logger"
	"github.com/mdaxf/iac/models"
	"github.com/mdaxf/iac/services"
)

// AgentRunnerHandler processes agent runner jobs from the job queue
type AgentRunnerHandler struct {
	iLog logger.Log
}

// NewAgentRunnerHandler creates a new agent runner handler
func NewAgentRunnerHandler() *AgentRunnerHandler {
	return &AgentRunnerHandler{
		iLog: logger.Log{
			ModuleName:     logger.Framework,
			User:           "System",
			ControllerName: "AgentRunnerHandler",
		},
	}
}

// Handle processes an agent runner queue job
func (h *AgentRunnerHandler) Handle(ctx context.Context, job *models.QueueJob) error {
	h.iLog.Info(fmt.Sprintf("AgentRunnerHandler START - JobID: %s", job.ID))

	var payload models.AgentJobPayload
	if err := json.Unmarshal([]byte(job.Payload), &payload); err != nil {
		h.iLog.Error(fmt.Sprintf("Failed to parse agent job payload: %v", err))
		return fmt.Errorf("failed to parse agent job payload: %w", err)
	}

	runner := services.GetGlobalAgentRunnerService()
	if runner == nil {
		return fmt.Errorf("agent runner service not initialized")
	}

	if err := runner.RunFromJob(ctx, payload); err != nil {
		h.iLog.Error(fmt.Sprintf("Agent job %s failed: %v", job.ID, err))
		return err
	}

	h.iLog.Info(fmt.Sprintf("AgentRunnerHandler COMPLETE - JobID: %s", job.ID))
	return nil
}

// RegisterAgentRunnerHandler registers the handler with the global job worker
func RegisterAgentRunnerHandler() {
	handler := NewAgentRunnerHandler()
	if GlobalJobWorker != nil {
		GlobalJobWorker.RegisterHandler("agent_runner", handler.Handle)
	}
}
