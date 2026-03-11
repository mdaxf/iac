package services

// Schedule and batch-schedule tools — wrap AIScheduleService and
// AIScheduleBatchService so agents can generate, optimize, and manage
// production schedules through the ReAct loop.
//
// Tools registered here:
//   generate_schedule          — NL prompt → task list
//   optimize_schedule          — inline optimization (≤ 30 tasks, synchronous)
//   optimize_schedule_batch    — async batch optimization for large task lists
//   get_batch_job_status       — poll async batch job progress
//   get_batch_job_result       — retrieve completed batch results
//   schedule_chat              — conversational AI about an existing schedule

import (
	"context"
	"encoding/json"
	"fmt"
)

func (s *AgentToolRegistryService) registerScheduleTools() {

	// ── generate_schedule ────────────────────────────────────────────────────
	s.register(ToolDefinition{
		Type: "function",
		Function: ToolFunctionDef{
			Name: "generate_schedule",
			Description: "Generates a production schedule (list of tasks with start/end times, resources, " +
				"dependencies, status) from a natural language description. " +
				"Returns tasks as JSON. Use optimize_schedule afterward to further refine the result.",
			Parameters: ToolParameterSchema{
				Type: "object",
				Properties: map[string]ToolPropertySchema{
					"prompt": {
						Type:        "string",
						Description: "Natural language description of the production schedule to generate. " +
							"Include relevant details: product type, quantity, deadline, shift patterns, priorities.",
					},
					"resources_json": {
						Type:        "string",
						Description: "Optional JSON array of available resources. " +
							"Format: [{\"id\":\"r1\",\"name\":\"Machine A\",\"type\":\"machine\",\"role\":\"CNC\"},...]. " +
							"Omit to let AI infer resources from the prompt.",
					},
				},
				Required: []string{"prompt"},
			},
		},
	}, func(ctx context.Context, args map[string]interface{}) (string, error) {
		prompt := stringArg(args, "prompt")
		if prompt == "" {
			return "", fmt.Errorf("prompt is required")
		}

		req := GenerateScheduleRequest{Prompt: prompt}

		if resourcesJSON := stringArg(args, "resources_json"); resourcesJSON != "" {
			if err := json.Unmarshal([]byte(resourcesJSON), &req.AvailableResources); err != nil {
				return "", fmt.Errorf("invalid resources_json: %w", err)
			}
		}

		svc := NewAIScheduleService()
		resp, err := svc.GenerateSchedule(ctx, req)
		if err != nil {
			return "", fmt.Errorf("schedule generation failed: %w", err)
		}

		out := map[string]interface{}{
			"tasks":    resp.Tasks,
			"metadata": resp.Metadata,
			"hint":     "Review the tasks and use optimize_schedule to improve resource utilization, compress the timeline, or rebalance workloads.",
		}
		b, _ := json.Marshal(out)
		return string(b), nil
	})

	// ── optimize_schedule ────────────────────────────────────────────────────
	s.register(ToolDefinition{
		Type: "function",
		Function: ToolFunctionDef{
			Name: "optimize_schedule",
			Description: "Optimizes an existing production schedule synchronously. Best for task lists with ≤ 30 tasks. " +
				"For larger lists use optimize_schedule_batch instead. " +
				"Objectives: 'compress' (shorten overall duration), 'balance' (level resource load), " +
				"'cost' (minimize cost by assigning cheaper resources), 'priority' (reorder by priority).",
			Parameters: ToolParameterSchema{
				Type: "object",
				Properties: map[string]ToolPropertySchema{
					"tasks_json": {
						Type:        "string",
						Description: "JSON array of current tasks to optimize. " +
							"Each task: {\"id\",\"name\",\"resourceIds\":[],\"start\":\"ISO\",\"end\":\"ISO\",\"progress\":0,\"status\":\"todo\"}. " +
							"Copy the tasks array from generate_schedule output.",
					},
					"resources_json": {
						Type:        "string",
						Description: "JSON array of available resources. Same format as generate_schedule.",
					},
					"objective": {
						Type:        "string",
						Description: "Optimization objective: 'compress', 'balance', 'cost', or 'priority'. " +
							"Or describe a custom goal, e.g. 'minimize overtime for resource R1'.",
					},
					"constraints_json": {
						Type:        "string",
						Description: "Optional JSON object of additional constraints, " +
							"e.g. '{\"max_overtime_hours\":2,\"deadline\":\"2026-03-15\"}'.",
					},
				},
				Required: []string{"tasks_json", "objective"},
			},
		},
	}, func(ctx context.Context, args map[string]interface{}) (string, error) {
		tasksJSON := stringArg(args, "tasks_json")
		if tasksJSON == "" {
			return "", fmt.Errorf("tasks_json is required")
		}
		objective := stringArg(args, "objective")
		if objective == "" {
			return "", fmt.Errorf("objective is required")
		}

		var tasks []Task
		if err := json.Unmarshal([]byte(tasksJSON), &tasks); err != nil {
			return "", fmt.Errorf("invalid tasks_json: %w", err)
		}

		req := OptimizeScheduleRequest{
			CurrentTasks: tasks,
			Objective:    objective,
		}

		if resourcesJSON := stringArg(args, "resources_json"); resourcesJSON != "" {
			if err := json.Unmarshal([]byte(resourcesJSON), &req.Resources); err != nil {
				return "", fmt.Errorf("invalid resources_json: %w", err)
			}
		}

		if constraintsJSON := stringArg(args, "constraints_json"); constraintsJSON != "" {
			if err := json.Unmarshal([]byte(constraintsJSON), &req.Constraints); err != nil {
				return "", fmt.Errorf("invalid constraints_json: %w", err)
			}
		}

		if len(tasks) > 30 {
			return "", fmt.Errorf("task list has %d tasks which exceeds the synchronous limit of 30. "+
				"Use optimize_schedule_batch instead for large task lists", len(tasks))
		}

		svc := NewAIScheduleService()
		resp, err := svc.OptimizeSchedule(ctx, req)
		if err != nil {
			return "", fmt.Errorf("schedule optimization failed: %w", err)
		}

		out := map[string]interface{}{
			"tasks":    resp.Tasks,
			"changes":  resp.Changes,
			"metadata": resp.Metadata,
		}
		b, _ := json.Marshal(out)
		return string(b), nil
	})

	// ── optimize_schedule_batch ──────────────────────────────────────────────
	s.register(ToolDefinition{
		Type: "function",
		Function: ToolFunctionDef{
			Name: "optimize_schedule_batch",
			Description: "Starts an async batch optimization for large schedules (30+ tasks). " +
				"Splits the task list into chunks, queues each as a background job, and returns a job_id. " +
				"Poll get_batch_job_status with the job_id until status is 'completed', then call get_batch_job_result.",
			Parameters: ToolParameterSchema{
				Type: "object",
				Properties: map[string]ToolPropertySchema{
					"tasks_json": {
						Type:        "string",
						Description: "JSON array of current tasks to optimize (same format as optimize_schedule).",
					},
					"resources_json": {
						Type:        "string",
						Description: "JSON array of available resources.",
					},
					"objective": {
						Type:        "string",
						Description: "Optimization objective: 'compress', 'balance', 'cost', or 'priority'.",
					},
					"batch_size": {
						Type:        "string",
						Description: "Number of tasks per batch (10–100, default 30). " +
							"Smaller batches complete faster individually but create more jobs.",
					},
					"constraints_json": {
						Type:        "string",
						Description: "Optional JSON object of additional constraints.",
					},
				},
				Required: []string{"tasks_json", "objective"},
			},
		},
	}, func(ctx context.Context, args map[string]interface{}) (string, error) {
		tasksJSON := stringArg(args, "tasks_json")
		if tasksJSON == "" {
			return "", fmt.Errorf("tasks_json is required")
		}
		objective := stringArg(args, "objective")
		if objective == "" {
			return "", fmt.Errorf("objective is required")
		}

		var tasks []Task
		if err := json.Unmarshal([]byte(tasksJSON), &tasks); err != nil {
			return "", fmt.Errorf("invalid tasks_json: %w", err)
		}

		req := OptimizeScheduleBatchRequest{
			CurrentTasks: tasks,
			Objective:    objective,
		}

		if resourcesJSON := stringArg(args, "resources_json"); resourcesJSON != "" {
			if err := json.Unmarshal([]byte(resourcesJSON), &req.Resources); err != nil {
				return "", fmt.Errorf("invalid resources_json: %w", err)
			}
		}

		if constraintsJSON := stringArg(args, "constraints_json"); constraintsJSON != "" {
			if err := json.Unmarshal([]byte(constraintsJSON), &req.Constraints); err != nil {
				return "", fmt.Errorf("invalid constraints_json: %w", err)
			}
		}

		if v := stringArg(args, "batch_size"); v != "" {
			var sz int
			if err := json.Unmarshal([]byte(v), &sz); err == nil {
				req.BatchSize = sz
			}
		} else if v, ok := args["batch_size"].(float64); ok {
			req.BatchSize = int(v)
		}

		svc := NewAIScheduleBatchService(s.sqlDB, nil)
		resp, err := svc.OptimizeScheduleBatch(ctx, req, "agent")
		if err != nil {
			return "", fmt.Errorf("batch optimization failed: %w", err)
		}

		out := map[string]interface{}{
			"job_id":        resp.JobID,
			"total_batches": resp.TotalBatches,
			"total_tasks":   resp.TotalTasks,
			"batch_size":    resp.BatchSize,
			"status":        resp.Status,
			"message":       resp.Message,
			"hint":          "Poll get_batch_job_status with job_id until status = 'completed', then call get_batch_job_result.",
		}
		b, _ := json.Marshal(out)
		return string(b), nil
	})

	// ── get_batch_job_status ─────────────────────────────────────────────────
	s.register(ToolDefinition{
		Type: "function",
		Function: ToolFunctionDef{
			Name: "get_batch_job_status",
			Description: "Returns the current status and progress of an async batch schedule optimization job " +
				"started by optimize_schedule_batch. Poll this until status is 'completed', 'partial', or 'failed'.",
			Parameters: ToolParameterSchema{
				Type: "object",
				Properties: map[string]ToolPropertySchema{
					"job_id": {
						Type:        "string",
						Description: "Batch job ID returned by optimize_schedule_batch.",
					},
				},
				Required: []string{"job_id"},
			},
		},
	}, func(ctx context.Context, args map[string]interface{}) (string, error) {
		jobID := stringArg(args, "job_id")
		if jobID == "" {
			return "", fmt.Errorf("job_id is required")
		}

		svc := NewAIScheduleBatchService(s.sqlDB, nil)
		status, err := svc.GetBatchJobStatus(ctx, jobID)
		if err != nil {
			return "", fmt.Errorf("failed to get batch job status: %w", err)
		}

		b, _ := json.Marshal(status)
		return string(b), nil
	})

	// ── get_batch_job_result ─────────────────────────────────────────────────
	s.register(ToolDefinition{
		Type: "function",
		Function: ToolFunctionDef{
			Name: "get_batch_job_result",
			Description: "Retrieves the aggregated optimized tasks and changes from a completed batch job. " +
				"Only call this after get_batch_job_status reports status 'completed' or 'partial'. " +
				"A 'partial' status means some batches failed; successfully optimized tasks are still returned.",
			Parameters: ToolParameterSchema{
				Type: "object",
				Properties: map[string]ToolPropertySchema{
					"job_id": {
						Type:        "string",
						Description: "Batch job ID returned by optimize_schedule_batch.",
					},
				},
				Required: []string{"job_id"},
			},
		},
	}, func(ctx context.Context, args map[string]interface{}) (string, error) {
		jobID := stringArg(args, "job_id")
		if jobID == "" {
			return "", fmt.Errorf("job_id is required")
		}

		svc := NewAIScheduleBatchService(s.sqlDB, nil)
		result, err := svc.GetBatchJobResult(ctx, jobID)
		if err != nil {
			return "", fmt.Errorf("failed to get batch job result: %w", err)
		}

		b, _ := json.Marshal(result)
		return string(b), nil
	})

	// ── schedule_chat ────────────────────────────────────────────────────────
	s.register(ToolDefinition{
		Type: "function",
		Function: ToolFunctionDef{
			Name: "schedule_chat",
			Description: "Asks the scheduling AI a conversational question about an existing schedule. " +
				"Use this to explain schedule decisions, answer 'what if' questions, " +
				"identify conflicts, or get recommendations without performing a full optimization.",
			Parameters: ToolParameterSchema{
				Type: "object",
				Properties: map[string]ToolPropertySchema{
					"question": {
						Type:        "string",
						Description: "The question to ask about the schedule, e.g. " +
							"'Why is Machine A overloaded on Monday?' or 'What tasks are on the critical path?'",
					},
					"context_json": {
						Type:        "string",
						Description: "Optional JSON object providing schedule context for the question, " +
							"e.g. '{\"tasks\":[...],\"resources\":[...],\"objective\":\"minimize makespan\"}'.",
					},
				},
				Required: []string{"question"},
			},
		},
	}, func(ctx context.Context, args map[string]interface{}) (string, error) {
		question := stringArg(args, "question")
		if question == "" {
			return "", fmt.Errorf("question is required")
		}

		req := ChatRequest{Question: question}

		if ctxJSON := stringArg(args, "context_json"); ctxJSON != "" {
			if err := json.Unmarshal([]byte(ctxJSON), &req.Context); err != nil {
				return "", fmt.Errorf("invalid context_json: %w", err)
			}
		}

		svc := NewAIScheduleService()
		resp, err := svc.Chat(ctx, req)
		if err != nil {
			return "", fmt.Errorf("schedule chat failed: %w", err)
		}

		b, _ := json.Marshal(map[string]string{"answer": resp.Answer})
		return string(b), nil
	})
}
