package codegen

import (
	"context"
	"fmt"

	"github.com/mdaxf/iac/llm"
)

// CallOpenAI makes a chat completion call via the unified LLM client.
// The vendor and model are resolved from aiconfig.json (code_generation use case).
// Falls back to OpenAI using apiKey/model if config is unavailable.
func CallOpenAI(apiKey string, model string, messages []map[string]interface{}, temperature float64) (string, error) {
	return llm.CallLLM(context.Background(), "code_generation", apiKey, model, messages, temperature)
}

// AssistantRequest represents the incoming request for AI assistant
type AssistantRequest struct {
	Question             string                   `json:"question"`
	PageContext          map[string]interface{}   `json:"pageContext"`
	ConversationHistory  []map[string]interface{} `json:"conversationHistory"`
}

// GenerateAssistantResponse generates context-aware responses for the AI assistant.
// Uses the chatbot use case from aiconfig.json, falling back to OpenAI.
func GenerateAssistantResponse(question string, apiKey string, openaiModel string, pageContext map[string]interface{}, conversationHistory []map[string]interface{}) (string, error) {
	systemPrompt := buildSystemPrompt(pageContext)

	messages := []map[string]interface{}{
		{"role": "system", "content": systemPrompt},
	}

	// Add conversation history (last 10 messages)
	if len(conversationHistory) > 0 {
		maxHistory := 10
		startIdx := 0
		if len(conversationHistory) > maxHistory {
			startIdx = len(conversationHistory) - maxHistory
		}
		for _, msg := range conversationHistory[startIdx:] {
			role, _ := msg["role"].(string)
			content, _ := msg["content"].(string)
			if role == "user" || role == "assistant" {
				messages = append(messages, map[string]interface{}{
					"role":    role,
					"content": content,
				})
			}
		}
	}

	messages = append(messages, map[string]interface{}{
		"role":    "user",
		"content": question,
	})

	return llm.CallLLM(context.Background(), "chatbot", apiKey, openaiModel, messages, 0.7)
}

func buildSystemPrompt(pageContext map[string]interface{}) string {
	pageName, _ := pageContext["pageName"].(string)
	pageType, _ := pageContext["pageType"].(string)
	route, _ := pageContext["route"].(string)

	if pageName == "" {
		pageName = "Unknown Page"
	}

	basePrompt := `You are an intelligent AI assistant for the IAC (Intelligent Application Composer) system. Your role is to help users understand and use the system effectively.

**Your Capabilities:**
1. Answer questions about the current page and its features
2. Explain how to use specific functionality
3. Provide step-by-step guidance for common tasks
4. Troubleshoot issues and suggest solutions
5. Explain system concepts and workflows
6. Guide users through complex processes

**Guidelines:**
- Be helpful, friendly, and concise
- Provide specific, actionable information
- Use examples when explaining concepts
- Break down complex tasks into simple steps
- Acknowledge when you don't know something
- Suggest relevant features the user might not know about
- Use markdown formatting for better readability (bold, lists, code blocks)
- Keep responses under 300 words unless detailed explanation is needed

**Current Context:**
- User is on: ` + pageName + `
- Page Type: ` + pageType + `
- Route: ` + route + `

`

	pageGuidance := getPageSpecificGuidance(pageName, pageType)
	if pageGuidance != "" {
		basePrompt += "\n" + pageGuidance
	}

	basePrompt += `

**System Features Overview:**
- **BPM Management**: Create and manage business process flows with function groups, routing, and data transformations
- **Page Editor**: Design pages with panels, views, actions, and datasources
- **View Editor**: Create UI views with HTML, JavaScript, CSS, and form definitions
- **Workflow Editor**: Build workflows with tasks, gateways, and decision points using React Flow
- **Whiteboard**: Create diagrams, flowcharts, and wireframes using Excalidraw
- **Database Diagram**: Visualize database schema and relationships
- **AI Generators**: Use AI to generate BPM flows, pages, views, workflows, and whiteboard diagrams from text descriptions or images

Now, answer the user's question with context-aware, helpful information.`

	return basePrompt
}

func getPageSpecificGuidance(pageName string, pageType string) string {
	guidance := make(map[string]string)

	guidance["BPM Editor"] = `**BPM Editor Help:**
- Create function groups to organize business logic
- Define routing between function groups based on conditions
- Add functions (JavaScript, Query, Table operations, etc.)
- Use AI chat bot to generate flows from descriptions
- Test functions with the test panel
- Save and version your TranCode`

	guidance["Page Editor"] = `**Page Editor Help:**
- Add panels to structure your page layout
- Assign views to panels for UI content
- Configure actions (buttons, events, workflows)
- Set up datasources for data binding
- Use AI chat bot to generate complete page structures
- Preview and test your page before saving`

	guidance["View Editor"] = `**View Editor Help:**
- Write HTML for your view structure
- Add JavaScript for interactivity and data binding
- Style with CSS for appearance
- Define fields, actions, inputs, and outputs
- Use AI chat bot to generate views from descriptions
- Test your view in the preview panel`

	guidance["Workflow Editor"] = `**Workflow Editor Help:**
- Drag nodes from the left toolbar (Start, Task, Gateway, Subflow, End)
- Connect nodes by dragging from one to another
- Click nodes to configure properties (name, page, trancode, users, roles)
- Set up routing tables for gateway nodes
- Use AI chat bot to generate workflow structures
- Validate before saving`

	guidance["Whiteboard Editor"] = `**Whiteboard Editor Help:**
- Use drawing tools to create shapes, arrows, text
- Right-click selected elements to "Generate BPM by AI" or "Generate View by AI"
- Use AI chat bot to generate complete diagrams
- Save to library for reuse
- Export as PNG or SVG
- Collaborate in real-time (if enabled)`

	if pageType == "list" {
		guidance[pageName] = `**` + pageName + ` Help:**
- Click rows to view/edit items
- Use search and filters to find specific items
- Click "New" or "Create" to add items
- Export data if needed
- Sort columns by clicking headers`
	}

	guidance["Home Dashboard"] = `**Dashboard Help:**
- Navigate using the menu on the left
- Check system health and notifications
- Access recent items and quick actions
- Use search to find features quickly
- AI Assistant (me!) is always available to help`

	guidance["Settings"] = `**Settings Help:**
- Configure user preferences
- Manage system settings
- Update API configurations
- View and restore backups
- Check system information`

	guidance["System Health Monitor"] = `**Health Monitor Help:**
- View real-time status of all system components
- Check backend servers, nodes, and web servers
- Monitor MongoDB, MySQL, SignalR health
- Auto-refreshes every 30 seconds
- Red = issues, Green = healthy`

	if help, exists := guidance[pageName]; exists {
		return help
	}
	return fmt.Sprintf("")
}
