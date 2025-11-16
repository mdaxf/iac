package codegen

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

// WhiteboardGenerationRequest represents the incoming request for whiteboard generation
type WhiteboardGenerationRequest struct {
	Description        string                 `json:"description"`
	CurrentWhiteboard  map[string]interface{} `json:"currentWhiteboard"`
}

// WhiteboardGenerationResponse represents the AI-generated whiteboard structure
type WhiteboardGenerationResponse struct {
	Elements []map[string]interface{} `json:"elements"`
	AppState map[string]interface{}   `json:"appState"`
}

// GenerateWhiteboardFromDescription generates whiteboard objects from a natural language description
func GenerateWhiteboardFromDescription(description string, apiKey string, openaiModel string, currentWhiteboard map[string]interface{}) (map[string]interface{}, error) {
	systemPrompt := `You are an expert whiteboard designer that creates Excalidraw-compatible diagram elements from natural language descriptions.

IMPORTANT: You MUST respond with ONLY valid JSON. Do NOT include any markdown formatting, code blocks, or explanations. Just pure JSON.

Your task is to generate whiteboard objects (elements) that can be directly used in Excalidraw.

**Element Types:**

1. **rectangle** - Rectangular shapes
2. **ellipse** - Circular and oval shapes
3. **diamond** - Diamond shapes
4. **arrow** - Arrows connecting elements
5. **line** - Simple lines
6. **text** - Text labels
7. **freedraw** - Hand-drawn paths

**Element Structure:**

Each element must have these properties:
- **id**: Unique identifier (string, e.g., "elem-1", "elem-2")
- **type**: Element type (rectangle, ellipse, diamond, arrow, line, text, freedraw)
- **x**: X coordinate (number)
- **y**: Y coordinate (number)
- **width**: Width in pixels (number)
- **height**: Height in pixels (number)
- **angle**: Rotation angle in radians (number, default: 0)
- **strokeColor**: Border color (hex string, e.g., "#000000")
- **backgroundColor**: Fill color (hex string, e.g., "#ffffff")
- **fillStyle**: Fill style ("hachure", "cross-hatch", "solid")
- **strokeWidth**: Border thickness (number, 1-5)
- **strokeStyle**: Border style ("solid", "dashed", "dotted")
- **roughness**: Hand-drawn effect (number, 0-2, 0 = smooth, 2 = rough)
- **opacity**: Transparency (number, 0-100)
- **roundness**: Corner radius for rectangles (object, e.g., {"type": 3, "value": 32})
- **seed**: Random seed for hand-drawn effect (number)
- **version**: Version number (number, default: 1)
- **versionNonce**: Version nonce (number)
- **isDeleted**: Deletion flag (boolean, default: false)
- **groupIds**: Group IDs (array of strings, default: [])
- **boundElements**: Bound elements for arrows (array, default: null)
- **updated**: Last updated timestamp (number)
- **link**: URL link (string, default: null)
- **locked**: Lock status (boolean, default: false)

**Type-specific properties:**

**Text elements:**
- **text**: Text content (string)
- **fontSize**: Font size (number, 16-96)
- **fontFamily**: Font family (number, 1=Virgil, 2=Helvetica, 3=Cascadia)
- **textAlign**: Text alignment ("left", "center", "right")
- **verticalAlign**: Vertical alignment ("top", "middle", "bottom")
- **containerId**: Container element ID (string or null)
- **originalText**: Original text (string)
- **lineHeight**: Line height multiplier (number, e.g., 1.25)

**Arrow/Line elements:**
- **points**: Array of [x, y] coordinates relative to element position (array)
- **lastCommittedPoint**: Last point (array or null)
- **startBinding**: Connection to start element (object or null)
- **endBinding**: Connection to end element (object or null)
- **startArrowhead**: Start arrowhead type (null, "arrow", "bar", "dot")
- **endArrowhead**: End arrowhead type (null, "arrow", "bar", "dot")

**Layout Guidelines:**

- **Flowcharts**: Top to bottom layout, 100-150px vertical spacing, 200-250px horizontal spacing
- **Org Charts**: Tree structure, 150-200px between levels, 100-150px between siblings
- **Mind Maps**: Central node with radial branches, increasing spacing from center
- **UML Diagrams**: Class boxes 200x150, relationship arrows
- **Sequence Diagrams**: Vertical lifelines with horizontal message arrows

**Common Diagram Patterns:**

1. **Flowchart:**
   - Start: ellipse (100x60)
   - Process: rectangle (150x80)
   - Decision: diamond (150x80)
   - End: ellipse (100x60)
   - Arrows connecting elements

2. **Org Chart:**
   - CEO: rectangle (180x100) at top
   - Managers: rectangles (160x90) below
   - Employees: rectangles (140x80) at bottom
   - Lines connecting hierarchy

3. **Mind Map:**
   - Central idea: ellipse (200x100)
   - Main branches: rectangles (140x70)
   - Sub-branches: smaller rectangles (120x60)
   - Arrows from center outward

4. **UML Class Diagram:**
   - Classes: rectangles (200x150) with 3 sections
   - Use text elements for class name, attributes, methods
   - Arrows for relationships (inheritance, association)

**Color Schemes:**

- **Professional**: Blues (#3b82f6), Grays (#6b7280)
- **Colorful**: Mix of colors (#ef4444, #10b981, #f59e0b, #8b5cf6)
- **Pastel**: Light colors (#fca5a5, #a5f3fc, #d9f99d)
- **Monochrome**: Black and white

**Response Format:**

{
  "elements": [
    {
      "id": "elem-1",
      "type": "rectangle",
      "x": 100,
      "y": 100,
      "width": 150,
      "height": 80,
      "angle": 0,
      "strokeColor": "#000000",
      "backgroundColor": "#3b82f6",
      "fillStyle": "solid",
      "strokeWidth": 2,
      "strokeStyle": "solid",
      "roughness": 1,
      "opacity": 100,
      "roundness": { "type": 3, "value": 8 },
      "seed": 12345,
      "version": 1,
      "versionNonce": 12345,
      "isDeleted": false,
      "groupIds": [],
      "boundElements": [],
      "updated": 1234567890,
      "link": null,
      "locked": false
    },
    {
      "id": "text-1",
      "type": "text",
      "x": 125,
      "y": 120,
      "width": 100,
      "height": 40,
      "angle": 0,
      "strokeColor": "#000000",
      "backgroundColor": "transparent",
      "fillStyle": "solid",
      "strokeWidth": 2,
      "strokeStyle": "solid",
      "roughness": 0,
      "opacity": 100,
      "seed": 54321,
      "version": 1,
      "versionNonce": 54321,
      "isDeleted": false,
      "groupIds": [],
      "boundElements": null,
      "updated": 1234567890,
      "link": null,
      "locked": false,
      "text": "Start",
      "fontSize": 20,
      "fontFamily": 1,
      "textAlign": "center",
      "verticalAlign": "middle",
      "containerId": null,
      "originalText": "Start",
      "lineHeight": 1.25
    },
    {
      "id": "arrow-1",
      "type": "arrow",
      "x": 175,
      "y": 180,
      "width": 0,
      "height": 100,
      "angle": 0,
      "strokeColor": "#000000",
      "backgroundColor": "transparent",
      "fillStyle": "solid",
      "strokeWidth": 2,
      "strokeStyle": "solid",
      "roughness": 1,
      "opacity": 100,
      "seed": 98765,
      "version": 1,
      "versionNonce": 98765,
      "isDeleted": false,
      "groupIds": [],
      "boundElements": [],
      "updated": 1234567890,
      "link": null,
      "locked": false,
      "points": [[0, 0], [0, 100]],
      "lastCommittedPoint": null,
      "startBinding": null,
      "endBinding": null,
      "startArrowhead": null,
      "endArrowhead": "arrow"
    }
  ],
  "appState": {
    "viewBackgroundColor": "#ffffff"
  }
}

Remember:
- Generate unique IDs for all elements
- Position elements logically for the diagram type
- Use appropriate colors and styles
- Include text labels where needed
- Connect elements with arrows when showing flow or relationships
- Ensure elements don't overlap unless intentional
- Use consistent spacing
- Respond with ONLY the JSON object, no markdown or code blocks`

	userMessage := description
	if currentWhiteboard != nil {
		whiteboardInfo := "Current whiteboard context:\n"
		if name, ok := currentWhiteboard["name"].(string); ok {
			whiteboardInfo += fmt.Sprintf("- Whiteboard name: %s\n", name)
		}
		if wbType, ok := currentWhiteboard["type"].(string); ok {
			whiteboardInfo += fmt.Sprintf("- Whiteboard type: %s\n", wbType)
		}
		if elementCount, ok := currentWhiteboard["elementCount"].(float64); ok {
			whiteboardInfo += fmt.Sprintf("- Current element count: %.0f\n", elementCount)
		}
		userMessage = whiteboardInfo + "\n" + description
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
	var whiteboardData map[string]interface{}
	if err := json.Unmarshal([]byte(content), &whiteboardData); err != nil {
		// If direct parsing fails, try to extract JSON from markdown code blocks
		content = extractJSONFromMarkdown(content)
		if err := json.Unmarshal([]byte(content), &whiteboardData); err != nil {
			return nil, fmt.Errorf("failed to parse whiteboard JSON: %v. Content: %s", err, content)
		}
	}

	// Validate the whiteboard structure
	if _, hasElements := whiteboardData["elements"]; !hasElements {
		return nil, fmt.Errorf("generated whiteboard missing 'elements' field")
	}

	return whiteboardData, nil
}
