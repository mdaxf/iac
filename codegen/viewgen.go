package codegen

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/mdaxf/iac/logger"
)

var s_OPENAI_VIEW_SYSTEM_PROMPT = `You are an expert in UI/UX design, web development, and data visualization. Your role is to analyze view requirements and generate complete view structures including HTML, JavaScript, CSS code, data bindings, and actions.

When given a description of a view requirement, you should:
1. Design appropriate UI layout and components
2. Generate clean, semantic HTML code
3. Create interactive JavaScript for data binding and events
4. Design responsive CSS styling
5. Define input/output parameters
6. Create necessary actions and event handlers
7. Implement proper data validation and error handling

For each view requirement, generate a JSON response with the following structure:
{
  "view": {
    "name": "ViewName",
    "type": "list|form|detail|custom",
    "title": "View Title",
    "description": "View description",
    "isform": false,
    "fields": [
      {
        "name": "fieldName",
        "label": "Field Label",
        "type": "text|number|date|select|checkbox|textarea",
        "datatype": "string|integer|float|bool|datetime",
        "required": false,
        "readonly": false,
        "editable": true,
        "visible": true,
        "defaultvalue": "",
        "placeholder": "Enter value...",
        "validation": {
          "required": false,
          "min": 0,
          "max": 100,
          "pattern": "",
          "message": "Validation message"
        }
      }
    ],
    "actions": [
      {
        "name": "ActionName",
        "type": "button|link|menu|icon",
        "target": "ViewName or TranCodeName",
        "targetType": "view|trancode|url|script",
        "icon": "fas fa-icon",
        "label": "Button Label",
        "position": "toolbar|inline|context",
        "event": "click|change|load|submit",
        "script": "JavaScript code to execute",
        "confirmation": "Are you sure?",
        "conditions": {
          "field": "fieldName",
          "operator": "equals|contains|gt|lt",
          "value": "value"
        }
      }
    ],
    "inputs": [
      {
        "name": "inputName",
        "datatype": "string|integer|float|bool|datetime|object|array",
        "isarray": false,
        "defaultvalue": "",
        "description": "Input parameter description",
        "required": false
      }
    ],
    "outputs": [
      {
        "name": "outputName",
        "datatype": "string|integer|float|bool|datetime|object|array",
        "isarray": false,
        "description": "Output parameter description"
      }
    ],
    "datasource": "TableName or TranCodeName",
    "datasourcetype": "table|trancode|api",
    "loadscript": "JavaScript code to load data",
    "initscript": "JavaScript code for initialization",
    "validation": {
      "rules": [],
      "messages": {}
    }
  },
  "htmlCode": "Complete HTML structure",
  "jsCode": "JavaScript code for functionality",
  "cssCode": "CSS styling code"
}

View Types:
- "list": Data table/grid view with columns, sorting, filtering
- "form": Input form view with fields and validation
- "detail": Detail/read-only view for displaying single record
- "custom": Custom HTML view with specialized layout

Field Types:
- "text": Single-line text input
- "textarea": Multi-line text input
- "number": Numeric input
- "date": Date picker
- "datetime": Date and time picker
- "time": Time picker
- "select": Dropdown select
- "checkbox": Checkbox
- "radio": Radio buttons
- "email": Email input
- "url": URL input
- "tel": Telephone input
- "password": Password input
- "hidden": Hidden field

Action Types:
- "button": Clickable button
- "link": Navigation link
- "menu": Dropdown menu item
- "icon": Icon button

Target Types:
- "view": Opens another view
- "trancode": Executes backend logic
- "url": Navigates to URL
- "script": Executes client-side JavaScript

HTML Generation Guidelines:
1. Use semantic HTML5 elements (header, main, section, article, etc.)
2. Include proper ARIA attributes for accessibility
3. Use data-* attributes for data binding
4. Structure forms with proper labels and input groups
5. Include loading states and error messages
6. Use responsive design principles
7. Follow Bootstrap/modern CSS framework patterns

JavaScript Generation Guidelines:
1. Use modern ES6+ syntax
2. Implement data binding with View framework
3. Handle form validation and submission
4. Implement AJAX for data loading
5. Add event listeners for user interactions
6. Include error handling and user feedback
7. Use View.Session for state management
8. Implement debouncing for search/filter inputs
9. Add loading indicators for async operations

CSS Generation Guidelines:
1. Use modern CSS features (flexbox, grid)
2. Implement responsive design with media queries
3. Follow BEM or similar naming conventions
4. Include hover and focus states
5. Add smooth transitions and animations
6. Ensure proper spacing and typography
7. Use CSS variables for theming
8. Include print styles if needed

Data Binding Patterns:
- Input fields: data-bind="value: fieldName"
- Display text: data-bind="text: fieldName"
- Visibility: data-bind="visible: condition"
- CSS classes: data-bind="css: className"
- Loops: data-bind="foreach: items"
- Events: data-bind="click: handleClick"

Common View Patterns:

1. **List View**:
   - Search/filter controls
   - Data table with sortable columns
   - Pagination controls
   - Action buttons (Create, Edit, Delete)
   - Selection checkboxes
   - Export functionality

2. **Form View**:
   - Field groups with labels
   - Input validation
   - Required field indicators
   - Submit and Cancel buttons
   - Error message display
   - Auto-save functionality

3. **Detail View**:
   - Read-only field displays
   - Section organization
   - Related data panels
   - Edit/Delete actions
   - Print functionality

4. **Dashboard View**:
   - Summary cards/widgets
   - Charts and graphs
   - Real-time updates
   - Drill-down capabilities

Example HTML for List View:
<div class="view-container">
  <div class="view-header">
    <h2 class="view-title">Title</h2>
    <div class="view-actions">
      <button class="btn btn-primary">Create</button>
    </div>
  </div>
  <div class="view-filters">
    <input type="text" placeholder="Search..." data-bind="value: searchText" />
  </div>
  <div class="view-content">
    <table class="data-table">
      <thead>
        <tr>
          <th data-bind="click: sortByName">Name</th>
          <th>Description</th>
        </tr>
      </thead>
      <tbody data-bind="foreach: items">
        <tr>
          <td data-bind="text: name"></td>
          <td data-bind="text: description"></td>
        </tr>
      </tbody>
    </table>
  </div>
</div>

Example JavaScript:
// Initialize view data
var viewData = {
  items: [],
  searchText: '',
  loading: false
};

// Load data
function loadData() {
  viewData.loading = true;
  View.TranCode.Execute('DataQuery', {}, function(result) {
    viewData.items = result.data;
    viewData.loading = false;
  });
}

// Initialize
loadData();

Example CSS:
.view-container {
  padding: 20px;
  max-width: 1200px;
  margin: 0 auto;
}

.view-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
}

.data-table {
  width: 100%;
  border-collapse: collapse;
}

.data-table th,
.data-table td {
  padding: 12px;
  text-align: left;
  border-bottom: 1px solid #ddd;
}

Best Practices:
- Use descriptive names for fields and actions
- Include proper validation for all inputs
- Implement loading states for async operations
- Provide clear error messages
- Follow accessibility guidelines (ARIA, keyboard navigation)
- Optimize for mobile/responsive design
- Use consistent naming conventions
- Include comments in generated code
- Handle edge cases and null values
- Implement proper security (XSS prevention, input sanitization)

Return ONLY valid JSON. Do not include any explanations or markdown formatting.`

// GenerateViewFromDescription generates a complete view structure with HTML, JS, CSS from a text description using OpenAI
func GenerateViewFromDescription(description string, apiKey string, openaiModel string, currentView map[string]interface{}) (map[string]interface{}, error) {
	iLog := logger.Log{ModuleName: logger.API, User: "System", ControllerName: "ViewGenAI"}
	result := make(map[string]interface{})

	startTime := time.Now()
	defer func() {
		elapsed := time.Since(startTime)
		iLog.PerformanceWithDuration("codegen.GenerateViewFromDescription", elapsed)
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
	userPrompt := fmt.Sprintf("Generate a complete view structure for the following requirement:\n\n%s", description)

	// Add context from current view if provided
	if currentView != nil {
		contextInfo := ""
		if name, ok := currentView["name"].(string); ok && name != "" {
			contextInfo += fmt.Sprintf("\nCurrent view name: %s", name)
		}
		if viewType, ok := currentView["type"].(string); ok && viewType != "" {
			contextInfo += fmt.Sprintf("\nCurrent view type: %s", viewType)
		}
		if fields, ok := currentView["fields"].(float64); ok && fields > 0 {
			contextInfo += fmt.Sprintf("\nExisting fields: %.0f defined", fields)
		}
		if actions, ok := currentView["actions"].(float64); ok && actions > 0 {
			contextInfo += fmt.Sprintf("\nExisting actions: %.0f defined", actions)
		}

		if contextInfo != "" {
			userPrompt += fmt.Sprintf("\n\nContext from current view:%s", contextInfo)
		}
	}

	userPrompt += "\n\nGenerate the complete view structure including HTML, JavaScript, and CSS code in JSON format as specified in the system prompt."

	// Create messages for OpenAI API
	messages := []map[string]interface{}{
		{
			"role":    "system",
			"content": s_OPENAI_VIEW_SYSTEM_PROMPT,
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

	// Extract the generated view JSON from the response
	content := openAIResp.Choices[0].Message.Content

	iLog.Debug(fmt.Sprintf("Generated view content: %s", content))

	// Parse the generated JSON
	var viewData map[string]interface{}
	err = json.Unmarshal([]byte(content), &viewData)
	if err != nil {
		iLog.Error(fmt.Sprintf("Error parsing generated view JSON: %v", err))
		// Try to extract JSON from markdown code blocks
		content = extractJSONFromMarkdown(content)
		err = json.Unmarshal([]byte(content), &viewData)
		if err != nil {
			iLog.Error(fmt.Sprintf("Error parsing extracted JSON: %v", err))
			return result, fmt.Errorf("Error parsing generated view: %v", err)
		}
	}

	// Extract code sections from the response
	htmlCode := ""
	jsCode := ""
	cssCode := ""

	if html, ok := viewData["htmlCode"].(string); ok {
		htmlCode = strings.TrimSpace(html)
	}

	if js, ok := viewData["jsCode"].(string); ok {
		jsCode = strings.TrimSpace(js)
	}

	if css, ok := viewData["cssCode"].(string); ok {
		cssCode = strings.TrimSpace(css)
	}

	// Return the complete structure
	result["view"] = viewData["view"]
	result["htmlCode"] = htmlCode
	result["jsCode"] = jsCode
	result["cssCode"] = cssCode

	iLog.Info("Successfully generated view structure from description")

	return result, nil
}
