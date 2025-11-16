package codegen

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/mdaxf/iac/logger"
)

var s_OPENAI_PAGE_SYSTEM_PROMPT = `You are an expert in UI/UX design and web application architecture. Your role is to analyze business requirements and generate complete page structures with panels, views, and backend logic integration.

When given a description of a page requirement, you should:
1. Design an appropriate page layout with panels
2. Identify what views are needed (list views, form views, detail views, etc.)
3. Determine what TranCodes (backend logic) are needed
4. Link everything together with appropriate actions
5. Suggest which views/trancodes already exist vs. need to be created

For each page requirement, generate a JSON response with the following structure:
{
  "page": {
    "name": "PageName",
    "title": "Page Title",
    "description": "Page description",
    "orientation": 1,
    "version": "1.0",
    "isdefault": false,
    "status": 0,
    "panels": [
      {
        "id": "panel1",
        "type": "grid|form|detail|custom",
        "title": "Panel Title",
        "view": "ViewName",
        "position": {"x": 0, "y": 0, "width": 12, "height": 6},
        "properties": {
          "showHeader": true,
          "collapsible": false
        }
      }
    ],
    "actions": [
      {
        "name": "ActionName",
        "type": "button|link|menu",
        "target": "ViewName or TranCodeName",
        "targetType": "view|trancode|url",
        "icon": "fas fa-icon",
        "label": "Button Label",
        "position": "toolbar|context|inline"
      }
    ]
  },
  "views": [
    {
      "action": "create|use_existing",
      "viewName": "ViewName",
      "viewConfig": {
        "name": "ViewName",
        "type": "list|form|detail|custom",
        "title": "View Title",
        "datasource": "TableName or TranCodeName",
        "fields": [
          {
            "name": "fieldName",
            "label": "Field Label",
            "type": "text|number|date|select|checkbox",
            "required": false,
            "editable": true
          }
        ],
        "actions": []
      }
    }
  ],
  "trancodes": [
    {
      "action": "create|use_existing",
      "trancodeName": "TranCodeName",
      "trancodeConfig": {
        "name": "TranCodeName",
        "trancodename": "TranCodeName",
        "description": "TranCode description",
        "version": "1.0",
        "status": 0,
        "inputs": [
          {
            "name": "inputName",
            "datatype": "string|integer|float|bool|datetime|object",
            "description": "Input description"
          }
        ],
        "outputs": [
          {
            "name": "outputName",
            "datatype": "string|integer|float|bool|datetime|object",
            "description": "Output description"
          }
        ],
        "functiongroups": [
          {
            "name": "FunctionGroupName",
            "description": "What this does",
            "functions": [
              {
                "name": "FunctionName",
                "functype": 8,
                "description": "Function description"
              }
            ]
          }
        ]
      }
    }
  ]
}

Panel Types:
- "grid": Data table/list panel
- "form": Input form panel
- "detail": Detail view panel
- "custom": Custom HTML/component panel

View Types:
- "list": Data list/table view
- "form": Input form view
- "detail": Detail/read-only view
- "custom": Custom view

Action Types:
- "button": Click button action
- "link": Navigation link
- "menu": Dropdown menu item

Target Types:
- "view": Opens a view (overlay or navigation)
- "trancode": Executes backend logic
- "url": Navigate to URL

Panel Orientation:
- 1: Portrait (default)
- 2: Landscape

Status Values:
- 0: Development
- 1: Testing
- 2: UAT
- 3: Production

Common Page Patterns:

1. **List-Detail Page**:
   - Top panel: Search/filter controls
   - Main panel: Data grid with list
   - Actions: Create, Edit, Delete, View Details

2. **CRUD Page**:
   - List view for browsing
   - Form view for create/edit
   - Detail view for read-only
   - TranCodes for CRUD operations

3. **Dashboard Page**:
   - Multiple panels with different views
   - Summary/statistics panels
   - Chart/graph panels
   - Recent activity panel

4. **Master-Detail Page**:
   - Master list panel on left
   - Detail panel on right
   - Actions to navigate between items

Best Practices:
- Use descriptive names for pages, views, and trancodes
- Follow naming convention: PageName, ViewName_List, ViewName_Form, TranCode_Action
- Reuse existing views and trancodes when possible
- Keep panels focused on single responsibility
- Link actions to appropriate views or trancodes
- Include necessary CRUD operations as trancodes

For "use_existing" actions:
- Set action to "use_existing" if the view/trancode likely already exists
- Common existing views: User_List, Product_List, Order_Form, etc.
- Common existing trancodes: User_CRUD, Data_Query, File_Upload, etc.

For "create" actions:
- Provide complete viewConfig or trancodeConfig
- Include all necessary fields, actions, and logic
- Follow the structure examples above

Return ONLY valid JSON. Do not include any explanations or markdown formatting.`

// GeneratePageFromDescription generates a complete page structure with views and trancodes from a text description using OpenAI
func GeneratePageFromDescription(description string, apiKey string, openaiModel string, currentPage map[string]interface{}) (map[string]interface{}, error) {
	iLog := logger.Log{ModuleName: logger.API, User: "System", ControllerName: "PageGenAI"}
	result := make(map[string]interface{})

	startTime := time.Now()
	defer func() {
		elapsed := time.Since(startTime)
		iLog.PerformanceWithDuration("codegen.GeneratePageFromDescription", elapsed)
	}()

	if apiKey == "" {
		iLog.Error("API key is required")
		return result, errors.New("API key is required")
	}

	if openaiModel == "" {
		iLog.Error("OpenAI model is required")
		return result, errors.New("OpenAI model is required")
	}

	if description == "" {
		iLog.Error("Description is required")
		return result, errors.New("Description is required")
	}

	// Build user prompt with context
	userPrompt := fmt.Sprintf("Generate a complete page structure for the following requirement:\n\n%s", description)

	// Add context from current page if provided
	if currentPage != nil {
		contextInfo := ""
		if name, ok := currentPage["name"].(string); ok && name != "" {
			contextInfo += fmt.Sprintf("\nCurrent page name: %s", name)
		}
		if panels, ok := currentPage["panels"].(float64); ok && panels > 0 {
			contextInfo += fmt.Sprintf("\nExisting panels: %.0f defined", panels)
		}
		if views, ok := currentPage["views"].([]interface{}); ok && len(views) > 0 {
			contextInfo += fmt.Sprintf("\nExisting views: %d defined", len(views))
			viewNames := make([]string, len(views))
			for i, v := range views {
				if viewName, ok := v.(string); ok {
					viewNames[i] = viewName
				}
			}
			if len(viewNames) > 0 {
				contextInfo += fmt.Sprintf(" (%v)", viewNames)
			}
		}

		if contextInfo != "" {
			userPrompt += fmt.Sprintf("\n\nContext from current page:%s", contextInfo)
		}
	}

	userPrompt += "\n\nGenerate the complete page structure including views and trancodes in JSON format as specified in the system prompt."

	// Create messages for OpenAI API
	messages := []map[string]interface{}{
		{
			"role":    "system",
			"content": s_OPENAI_PAGE_SYSTEM_PROMPT,
		},
		{
			"role":    "user",
			"content": userPrompt,
		},
	}

	// Create request
	requestBody := GPT4VCompletionRequest{
		Model:       openaiModel,
		Messages:    messages,
		Temperature: 0.7,
		MaxTokens:   4096,
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		iLog.Error(fmt.Sprintf("Error marshaling request: %v", err))
		return result, err
	}

	iLog.Debug(fmt.Sprintf("Request to OpenAI: %s", string(jsonBody)))

	// Make HTTP request to OpenAI
	req, err := http.NewRequestWithContext(context.Background(), "POST", url, bytes.NewBuffer(jsonBody))
	if err != nil {
		iLog.Error(fmt.Sprintf("Error creating request: %v", err))
		return result, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", apiKey))

	client := &http.Client{Timeout: 60 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		iLog.Error(fmt.Sprintf("Error making request: %v", err))
		return result, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		iLog.Error(fmt.Sprintf("Error reading response: %v", err))
		return result, err
	}

	iLog.Debug(fmt.Sprintf("Response from OpenAI: %s", string(body)))

	if resp.StatusCode != http.StatusOK {
		iLog.Error(fmt.Sprintf("OpenAI API error: %s", string(body)))
		return result, fmt.Errorf("OpenAI API error: %s", string(body))
	}

	// Parse OpenAI response
	var openAIResp GPT4VCompletionResponse
	err = json.Unmarshal(body, &openAIResp)
	if err != nil {
		iLog.Error(fmt.Sprintf("Error unmarshaling response: %v", err))
		return result, err
	}

	if len(openAIResp.Choices) == 0 {
		iLog.Error("No choices returned from OpenAI")
		return result, errors.New("No choices returned from OpenAI")
	}

	// Extract the generated page JSON from the response
	content := openAIResp.Choices[0].Message.Content

	iLog.Debug(fmt.Sprintf("Generated page content: %s", content))

	// Parse the generated JSON
	var pageData map[string]interface{}
	err = json.Unmarshal([]byte(content), &pageData)
	if err != nil {
		iLog.Error(fmt.Sprintf("Error parsing generated page JSON: %v", err))
		// Try to extract JSON from markdown code blocks
		content = extractJSONFromMarkdown(content)
		err = json.Unmarshal([]byte(content), &pageData)
		if err != nil {
			iLog.Error(fmt.Sprintf("Error parsing extracted JSON: %v", err))
			return result, fmt.Errorf("Error parsing generated page: %v", err)
		}
	}

	// Return the complete structure
	result["page"] = pageData["page"]
	result["views"] = pageData["views"]
	result["trancodes"] = pageData["trancodes"]

	iLog.Info("Successfully generated page structure from description")

	return result, nil
}
