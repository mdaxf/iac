package codegen

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

// WorkflowGenerationRequest represents the incoming request for workflow generation
type WorkflowGenerationRequest struct {
	Description     string                 `json:"description"`
	CurrentWorkflow map[string]interface{} `json:"currentWorkflow"`
}

// WorkflowGenerationResponse represents the AI-generated workflow structure
type WorkflowGenerationResponse struct {
	Nodes []map[string]interface{} `json:"nodes"`
	Edges []map[string]interface{} `json:"edges"`
}

// GenerateWorkflowFromDescription generates a complete workflow structure from a natural language description
func GenerateWorkflowFromDescription(description string, apiKey string, openaiModel string, currentWorkflow map[string]interface{}) (map[string]interface{}, error) {
	systemPrompt := `You are an expert workflow designer that creates workflow structures from natural language descriptions.

IMPORTANT: You MUST respond with ONLY valid JSON. Do NOT include any markdown formatting, code blocks, or explanations. Just pure JSON.

Your task is to generate a complete workflow structure with nodes and connections that follows React Flow format.

**Node Types:**
1. **start** - Start node (entry point of workflow)
   - Fields: id, type, name, description
   - Size: 24x24, circular

2. **task** - Task node (user action or activity)
   - Fields: id, type, name, description, page, trancode, users, roles, processdata
   - Size: 120x80, rectangular
   - Can have assigned page for UI
   - Can have assigned trancode for backend logic
   - Can be assigned to specific users or roles

3. **gateway** - Decision/Gateway node (routing based on conditions)
   - Fields: id, type, name, description, routingtables
   - Size: 100x100, diamond shape
   - routingtables format: [{ sequence: 1, data: "fieldName", value: "expectedValue", target: "targetNodeId", isDefault: false }]
   - One route can be marked as default (isDefault: true) for fallback

4. **subflow** - Subflow node (calls another workflow)
   - Fields: id, type, name, description, subflowName, subflowId
   - Size: 150x100, rectangular with border
   - References another workflow by name/ID

5. **end** - End node (completion point)
   - Fields: id, type, name, description
   - Size: 24x24, circular

**Edge Format:**
- id: "source-target" or unique identifier
- source: source node id
- target: target node id
- type: "default" (or "smoothstep", "step")
- data: { label: "connection label or condition" } (optional)

**Layout Guidelines:**
- Start nodes should be positioned at the top (x: 250, y: 50)
- Arrange nodes in logical flow from top to bottom or left to right
- Leave adequate spacing between nodes (150-200 units vertically, 250-300 units horizontally)
- Gateway branches should fan out horizontally
- End nodes should be at the bottom

**Common Workflow Patterns:**

1. **Linear Workflow:**
   Start -> Task 1 -> Task 2 -> Task 3 -> End

2. **Approval Workflow:**
   Start -> Submit Task -> Approval Gateway (Approved/Rejected) -> [Approved: Process Task, Rejected: Revision Task] -> End

3. **Multi-stage Workflow:**
   Start -> Data Entry -> Review Gateway -> [OK: Next Stage, Fix: Back to Data Entry] -> Final Approval -> End

4. **Parallel Tasks:**
   Start -> Split Gateway -> [Task A, Task B, Task C] -> Join Gateway -> End

**ProcessData:**
- Use processdata field in tasks to define data fields that are collected/updated
- Format: { "fieldName": { "type": "string/number/date", "label": "Field Label", "required": true } }

**Routing Tables (for gateways):**
- sequence: execution order (1, 2, 3...)
- data: field name to evaluate
- value: expected value for this route
- target: target node ID
- isDefault: true for default/fallback route (only one per gateway)

**Response Format:**
{
  "nodes": [
    {
      "id": "node1",
      "type": "start",
      "position": { "x": 250, "y": 50 },
      "data": {
        "name": "Start",
        "description": "Workflow start point",
        "label": "Start",
        "width": 24,
        "height": 24
      }
    },
    {
      "id": "node2",
      "type": "task",
      "position": { "x": 200, "y": 150 },
      "data": {
        "name": "Submit Request",
        "description": "User submits request",
        "label": "Submit Request",
        "page": "RequestForm",
        "trancode": "",
        "users": [],
        "roles": ["Requester"],
        "width": 120,
        "height": 80,
        "processdata": {
          "requestTitle": { "type": "string", "label": "Request Title", "required": true },
          "requestDetails": { "type": "string", "label": "Details", "required": true }
        }
      }
    },
    {
      "id": "node3",
      "type": "gateway",
      "position": { "x": 225, "y": 280 },
      "data": {
        "name": "Approval Decision",
        "description": "Manager approval",
        "label": "Approval Decision",
        "width": 100,
        "height": 100,
        "routingtables": [
          { "sequence": 1, "data": "approvalStatus", "value": "approved", "target": "node4" },
          { "sequence": 2, "data": "approvalStatus", "value": "rejected", "target": "node5" },
          { "sequence": 3, "data": "", "value": "", "target": "node5", "isDefault": true }
        ]
      }
    },
    {
      "id": "node4",
      "type": "end",
      "position": { "x": 250, "y": 420 },
      "data": {
        "name": "End",
        "description": "Workflow completion",
        "label": "End",
        "width": 24,
        "height": 24
      }
    }
  ],
  "edges": [
    {
      "id": "e1-2",
      "source": "node1",
      "target": "node2",
      "type": "default",
      "data": { "label": "" }
    },
    {
      "id": "e2-3",
      "source": "node2",
      "target": "node3",
      "type": "default",
      "data": { "label": "" }
    },
    {
      "id": "e3-4",
      "source": "node3",
      "target": "node4",
      "type": "default",
      "data": { "label": "approved" }
    }
  ]
}

Remember:
- Generate meaningful node IDs (e.g., "start1", "task-submit", "gateway-approval", "end1")
- Include descriptive names and labels
- Ensure all source/target references in edges match actual node IDs
- Position nodes for clear visual flow
- Use appropriate node types for each step
- Add routing tables for gateway nodes
- Respond with ONLY the JSON object, no markdown or code blocks`

	userMessage := description
	if currentWorkflow != nil {
		workflowInfo := "Current workflow context:\n"
		if name, ok := currentWorkflow["name"].(string); ok {
			workflowInfo += fmt.Sprintf("- Workflow name: %s\n", name)
		}
		if nodeCount, ok := currentWorkflow["nodeCount"].(float64); ok {
			workflowInfo += fmt.Sprintf("- Current node count: %.0f\n", nodeCount)
		}
		userMessage = workflowInfo + "\n" + description
	}

	// Prepare OpenAI request
	requestBody := map[string]interface{}{
		"model": openaiModel,
		"messages": []map[string]string{
			{"role": "system", "content": systemPrompt},
			{"role": "user", "content": userMessage},
		},
		"temperature": 0.7,
		"max_tokens":  4096,
	}

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %v", err)
	}

	// Make OpenAI API request
	req, err := http.NewRequest("POST", "https://api.openai.com/v1/chat/completions", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %v", err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("OpenAI API error (status %d): %s", resp.StatusCode, string(body))
	}

	// Parse OpenAI response
	var openAIResp struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}

	if err := json.Unmarshal(body, &openAIResp); err != nil {
		return nil, fmt.Errorf("failed to parse OpenAI response: %v", err)
	}

	if len(openAIResp.Choices) == 0 {
		return nil, fmt.Errorf("no choices in OpenAI response")
	}

	content := openAIResp.Choices[0].Message.Content

	// Try to parse the content as JSON
	var workflowData map[string]interface{}
	if err := json.Unmarshal([]byte(content), &workflowData); err != nil {
		// If direct parsing fails, try to extract JSON from markdown code blocks
		content = extractJSONFromMarkdown(content)
		if err := json.Unmarshal([]byte(content), &workflowData); err != nil {
			return nil, fmt.Errorf("failed to parse workflow JSON: %v. Content: %s", err, content)
		}
	}

	// Validate the workflow structure
	if _, hasNodes := workflowData["nodes"]; !hasNodes {
		return nil, fmt.Errorf("generated workflow missing 'nodes' field")
	}
	if _, hasEdges := workflowData["edges"]; !hasEdges {
		return nil, fmt.Errorf("generated workflow missing 'edges' field")
	}

	return workflowData, nil
}
