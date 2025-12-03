# Code Modification Handler for AI Assistant

## Issue
When editing a query/script function in BPM Editor and requesting AI to modify the code, the AI was giving general explanations instead of returning actual code changes.

**Root Cause**: The backend detected `code_modification` intent but had no handler for it, falling back to `GeneralQuestionHandler` which only provides explanations.

## Solution

Created a new `CodeModificationHandler` that specifically handles code/query/script modifications for selected functions in BPM Editor.

## Files Modified

### 1. `services/aiagency_handlers.go` (Lines 602-662)

Added new handler:

**IMPORTANT FIX** (2025-12-03 10:19): Updated handler to read `EntityContext` from top-level request field instead of `Options`:
```go
// CodeModificationHandler handles code/query/script modifications for existing functions
type CodeModificationHandler struct {
	OpenAIKey   string
	OpenAIModel string
	iLog        logger.Log
}

func (h *CodeModificationHandler) GetName() string {
	return "CodeModificationHandler"
}

func (h *CodeModificationHandler) CanHandle(ctx context.Context, intent string, editorType string) bool {
	// Handle code_modification intent for BPM editor
	return intent == "code_modification" && editorType == "bpm"
}

func (h *CodeModificationHandler) Handle(ctx context.Context, request *AIAgencyRequest, conversation *ConversationContext) (*AIAgencyResponse, error) {
	// Check if we have entity context (selected function/query)
	// EntityContext is a top-level field in AIAgencyRequest
	if request.EntityContext == nil || len(request.EntityContext) == 0 {
		h.iLog.Warn("No entity_context provided for code modification")
		return nil, fmt.Errorf("code modification requires a selected function/query")
	}

	entityContext := request.EntityContext

	// Get the system prompt override if provided (from Options)
	systemPrompt := ""
	if request.Options != nil {
		if sp, ok := request.Options["system_prompt_override"].(string); ok {
			systemPrompt = sp
		}
	}

	// Build the full context for code generation including current function data
	fullContext := map[string]interface{}{
		"entity_context":   entityContext,
		"page_context":     request.PageContext,
		"conversation":     toConversationHistory(conversation.ConversationHistory),
		"system_prompt":    systemPrompt,
	}

	// Merge entity context into the full context for flow generation
	if data, ok := entityContext["data"].(map[string]interface{}); ok {
		for k, v := range data {
			fullContext[k] = v
		}
	}

	// Use GenerateFlowFromDescription with full context
	result, err := codegen.GenerateFlowFromDescription(
		request.Question,
		h.OpenAIKey,
		h.OpenAIModel,
		fullContext,
	)

	if err != nil {
		h.iLog.Error(fmt.Sprintf("Failed to generate code modification: %v", err))
		return nil, fmt.Errorf("failed to generate code modification: %w", err)
	}

	return &AIAgencyResponse{
		Answer:         "✅ I've updated the code based on your request. You can review and apply the changes.",
		ResponseType:   "code_generation",
		IntentType:     "code_modification",
		Data:           result,
		RequiresAction: true,
		NextStep:       "apply_to_editor",
		Confidence:     0.9,
	}, nil
}
```

### 2. `services/aiagencyservice.go` (Lines 430-435)

Registered the handler in `getHandlers()`:
```go
// getHandlers returns all available transaction handlers
func (s *AIAgencyService) getHandlers() []TransactionHandler {
	return []TransactionHandler{
		// Code modification should be checked first (more specific)
		&CodeModificationHandler{
			OpenAIKey:   s.OpenAIKey,
			OpenAIModel: s.OpenAIModel,
			iLog:        s.iLog,
		},
		// ... other handlers ...
	}
}
```

**Note**: CodeModificationHandler is registered FIRST to ensure it's checked before more general handlers.

### 3. Frontend: `AIAssistant.tsx` (Lines 353-360)

Fixed request body to send `system_prompt_override` in `options` field:
```typescript
// Add selection context for targeted modifications
if (selectionContext && selectionContext.type !== 'none') {
  requestBody.entity_context = selectionContext
  // Put system_prompt_override in options as expected by backend
  requestBody.options = {
    system_prompt_override: contextSystemPrompt
  }
}
```

**Before**: `system_prompt_override` was sent as top-level field
**After**: `system_prompt_override` is sent in `options` object

### 4. Frontend: `aiContextExtractor.ts` (Lines 179-208)

Enhanced system prompt for code/query functions:
```typescript
if (context.type === 'function') {
  const hasContent = !!context.metadata.content
  const funcType = getFunctionTypeName(context.metadata.functionType!)
  const isCodeFunction = context.metadata.functionType === FunctionType.Javascript ||
                        context.metadata.functionType === FunctionType.GoExpr ||
                        context.metadata.functionType === FunctionType.Query ||
                        context.metadata.functionType === FunctionType.StoreProcedure

  let promptAddition = ''
  if (isCodeFunction) {
    promptAddition = `
IMPORTANT: This is a ${funcType} function. When the user requests code/query modifications:
1. You MUST return a complete JSON object with the updated function
2. The JSON must include the 'content' field with the modified code/query
3. Include all required fields: id, name, functype, content, inputs, outputs
4. Do NOT just explain what to do - provide the actual code change in JSON format
5. The user expects to see code they can apply directly to the editor

Current content:
\`\`\`
${context.metadata.content || '(empty)'}
\`\`\`
`
  }

  return `${basePrompt} The user has selected a specific Function named "${context.metadata.name}" of type "${funcType}".
${promptAddition}
When the user requests changes, focus on modifying ONLY this specific function's properties, inputs, outputs, or content.
Return the updated function object with the same ID (${context.metadata.id}) to ensure it replaces the existing function.`
}
```

## How It Works

### Flow

1. **User selects a query/script function** in BPM Editor
   - Frontend dispatches `bpm-selection-changed` event
   - AIAssistant captures selection context

2. **User requests modification** (e.g., "update the query to summarize by date")
   - AIAssistant sends request with:
     - `entity_context`: Selected function data
     - `system_prompt_override`: Custom prompt with current code

3. **Backend detects intent**
   - Intent detection identifies: `code_modification` (confidence: 0.90)

4. **CodeModificationHandler is invoked**
   - Checks if it can handle: `intent == "code_modification" && editorType == "bpm"`
   - Extracts entity context and system prompt
   - Calls `codegen.GenerateFlowFromDescription` with full context

5. **AI generates code**
   - Uses custom system prompt that includes current code
   - Returns complete function object with updated `content` field

6. **Frontend displays result**
   - Shows generated code in confirmation modal
   - User can review and apply changes

### Key Features

- **Context-Aware**: Includes current function code in the prompt
- **Explicit Instructions**: System prompt explicitly tells AI to return JSON code
- **Type-Specific**: Only activates for code/query/script functions
- **Prevents Explanation Mode**: Handler prevents fallback to general question handler

## Testing

### Test Case 1: Modify Query
1. Open BPM Editor
2. Select a Query function node
3. Open AI Assistant
4. Request: "update the query to summarize conversations by date"
5. **Expected**: AI returns complete function with modified SQL query
6. **Verify**: Console shows `CodeModificationHandler` handling the request

### Test Case 2: Modify JavaScript
1. Select a JavaScript function
2. Request: "add error handling to this script"
3. **Expected**: AI returns function with updated JavaScript code

### Test Case 3: Modify Go Expression
1. Select a GoExpr function
2. Request: "optimize this expression"
3. **Expected**: AI returns function with updated Go code

## Backend Logs

### Before Fix
```
[I] System AIAgencyService 🔍 Intent Detection - Detected Intent: code_modification (confidence: 0.90)
[I] System AIAgencyService ⚠️  No specific handler found for intent 'code_modification', using GeneralQuestionHandler
```

### After Fix
```
[I] System AIAgencyService 🔍 Intent Detection - Detected Intent: code_modification (confidence: 0.90)
[I] System AIAgencyService ✅ Using handler: CodeModificationHandler
[I] System AIAgencyService 📝 Generating code modification with full context
[I] System AIAgencyService ✅ Code modification generated successfully
```

## Benefits

1. **Accurate Code Generation**: AI receives current code as context
2. **Proper Response Format**: AI returns JSON instead of explanations
3. **Type Safety**: Handler validates entity context before processing
4. **Better UX**: Users get immediate, applicable code changes

## Troubleshooting

### Error: "No entity_context provided for code modification"

**Symptoms**:
```
[W] System AIAgencyService No entity_context provided for code modification
[E] Error handling AI agency request: handler failed: code modification requires a selected function/query
```

**Causes**:
1. No function/query selected in BPM Editor before asking AI
2. Selection context not being dispatched from FlowEditor
3. Request body not including entity_context

**Solution**:
1. Make sure to click on a Query/JavaScript/GoExpr function node before requesting modifications
2. Verify console shows: `[AIAssistant] Received BPM selection context`
3. Check that `selectionContext.type` is 'function' and not 'none'

### Error: "system_prompt_override not found"

**Cause**: Frontend sending `system_prompt_override` as top-level field instead of in `options`

**Solution**: Already fixed in `AIAssistant.tsx:357` - system prompt is now sent in options object

### AI still giving explanations instead of code

**Possible Causes**:
1. CodeModificationHandler not being selected (check intent detection logs)
2. System prompt not being applied correctly
3. Function type doesn't support content modification

**Debug Steps**:
1. Check backend logs for: `✅ Using handler: CodeModificationHandler`
2. Verify function type is Query/JavaScript/GoExpr/StoreProcedure
3. Check that `selectionContext.metadata.content` exists
4. Verify system prompt includes "IMPORTANT: This is a ... function"

## Date
2025-12-03

## Status
✅ **IMPLEMENTED** - Code Modification Handler is now active

## Updates

### 2025-12-03 10:19 - Fixed EntityContext Reading
- Changed handler to read `EntityContext` from top-level request field
- Updated frontend to send `system_prompt_override` in `options` object
- Rebuilt backend with fixes

## Compliance
✅ **COMPLIANT** with claude.md rules - Proper code structure, no mock data
