package services

// Planning and scheduling optimization tools — expose PlanSchedulerProfileService
// to agents so they can load plan data and reason over optimization opportunities.

import (
	"context"
	"encoding/json"
	"fmt"
)

// registerPlanningTools registers plan profile listing and optimization tools
func (s *AgentToolRegistryService) registerPlanningTools() {
	s.register(ToolDefinition{
		Type: "function",
		Function: ToolFunctionDef{
			Name:        "list_plan_profiles",
			Description: "Lists all plan scheduler profiles. Use this to discover available profiles before calling plan_optimize.",
			Parameters:  ToolParameterSchema{Type: "object", Properties: map[string]ToolPropertySchema{}},
		},
	}, func(ctx context.Context, args map[string]interface{}) (string, error) {
		svc := NewPlanSchedulerProfileService(s.sqlDB, s.db)
		profiles, err := svc.GetAllProfiles(ctx)
		if err != nil {
			return "", fmt.Errorf("failed to list plan profiles: %w", err)
		}

		type item struct {
			ID          string `json:"id"`
			Name        string `json:"name"`
			Description string `json:"description,omitempty"`
			IsDefault   bool   `json:"is_default"`
		}
		items := make([]item, 0, len(profiles))
		for _, p := range profiles {
			items = append(items, item{
				ID:          p.ID,
				Name:        p.Name,
				Description: p.Description,
				IsDefault:   p.IsDefault,
			})
		}
		b, _ := json.Marshal(items)
		return string(b), nil
	})

	s.register(ToolDefinition{
		Type: "function",
		Function: ToolFunctionDef{
			Name: "plan_optimize",
			Description: "Loads a plan scheduler profile's full data (tasks, resources, constraints, settings) for AI-driven optimization analysis. " +
				"Returns the plan data so the agent can reason over scheduling conflicts, resource under-utilization, overtime risks, or priority mismatches " +
				"and suggest specific improvements such as task reordering, resource reassignment, or constraint adjustments. " +
				"Use list_plan_profiles first to get available profile IDs.",
			Parameters: ToolParameterSchema{
				Type: "object",
				Properties: map[string]ToolPropertySchema{
					"profile_id": {
						Type:        "string",
						Description: "Plan scheduler profile ID. Use list_plan_profiles to discover IDs. Omit to use the default profile.",
					},
					"parameters_json": {
						Type:        "string",
						Description: "Optional JSON object of parameters passed to profile data sources (e.g. '{\"date\":\"2026-03-01\",\"plant\":\"Plant-A\"}').",
					},
				},
			},
		},
	}, func(ctx context.Context, args map[string]interface{}) (string, error) {
		svc := NewPlanSchedulerProfileService(s.sqlDB, s.db)

		profileID, _ := args["profile_id"].(string)

		// Resolve profile — use default if no ID given
		if profileID == "" {
			profile, err := svc.GetDefaultProfile(ctx)
			if err != nil {
				return "", fmt.Errorf("no default plan profile found; specify a profile_id from list_plan_profiles: %w", err)
			}
			profileID = profile.ID
		}

		// Parse caller-supplied parameters
		parameters := map[string]interface{}{}
		if v := stringArg(args, "parameters_json"); v != "" {
			json.Unmarshal([]byte(v), &parameters) //nolint:errcheck
		}

		// Load full profile data (tasks, resources, constraints, settings)
		data, err := svc.LoadProfileData(ctx, profileID, parameters)
		if err != nil {
			return "", fmt.Errorf("failed to load plan profile data for %s: %w", profileID, err)
		}

		out := map[string]interface{}{
			"profile_id": profileID,
			"data":       data,
			"hint": "Analyze the tasks and resources above to identify: " +
				"(1) scheduling conflicts or overlaps, " +
				"(2) resource under-utilization or bottlenecks, " +
				"(3) overtime or capacity risks, " +
				"(4) priority mismatches. " +
				"Suggest specific reordering, resource reassignments, or constraint adjustments with reasoning.",
		}
		b, _ := json.Marshal(out)
		return string(b), nil
	})
}
