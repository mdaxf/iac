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

func GetMockupFromUserStory(apiKey string, openaimodel string, userStory string) (map[string]interface{}, error) {

	iLog := logger.Log{ModuleName: logger.API, User: "System", ControllerName: "CodeAi"}
	result := make(map[string]interface{})
	startTime := time.Now()

	defer func() {
		elapsed := time.Since(startTime)
		iLog.PerformanceWithDuration("codegeneration.GetHtmlCodeFromImage", elapsed)
	}()

	if apiKey == "" {
		iLog.Error("API key is required")
		return result, errors.New("API key is required")
	}

	if openaimodel == "" {
		iLog.Error("OpenAI model is required")
		return result, errors.New("OpenAI model is required")
	}

	if userStory == "" {
		iLog.Error("User story is required")
		return result, errors.New("User story is required")
	}

	// Get mockup from user story

	system_prompt := `You are a helpful assistant, professional UI/UX expert, and manufacturing IT solution exper, that generates design mockup UIs and business logic flow charts in Excalidraw format. The output should be a JSON object that can be directly used in Excalidraw. The generated JSON object should be set in the data node.

	If the total token count exceeds 4096, regenerate the UI and business logic to reduce the content length.
	
	If the user story requires creating a UI, please create the UI mockup. If there is business logic in the user story, please create the business logic flow chart in Excalidraw format.
	
	Please separate the UI and the business logic flow chart into different blocks. However, both should be in the same Excalidraw object.
	
	The business logic blocks must be connected with arrows and vertically arranged.
	
	Additionally, if there is a UI, include the related business logic, such as loading data for the UI initialization and logic for each button click to represent the business flow. In some cases, the UI may include logic to call the backend API to get the data.
	
	The UI should have a rectangle outline which includes all elements of the UI.
	
	The UI should include elements like titles, labels, buttons, lists, tables, charts, etc. Each block should have a caption or description to present what it is.
	
	UI titles, labels, descriptions, and notes use the text type. Buttons use the rectangle type with text inside.
	
	Use arrow lines to link the business logic flow between the UI elements. Each arrow line should have a caption or description to represent the business logic flow and should indicate the flow direction.
	
	Please omit the elements' common or unimportant attributes, such as color, alignment, etc. Only include required attributes like type, x, y, width, height, seed, id, and text (if it is a text type), boundElements (if the element is part of another element) for each element. id and seed must be unique for each element.
	
	For the arrow type, the boundElements, startBinding, endBinding, endArrowhead, and startArrowhead nodes are required. The boundElements is an array of objects which has id and type, and id represents the bounded element's id. The startBinding and endBinding have the attributes of elementId, focus, and gap.
	
	No comments or // in the Excalidraw object format.`
	//

	user_prompt := `Given the following user story or description:
	` + userStory
	// /
	//
	//If there is the business logic in the user story, please create the business logic flow chart in Excalidraw format. The business logic flow chart should include the flow of the business logic, the decision points, the data flow, etc.

	messages := make([]map[string]interface{}, 2)
	messages[0] = map[string]interface{}{"role": "system", "content": system_prompt}
	messages[1] = map[string]interface{}{"role": "user", "content": user_prompt}
	request := GPT4VCompletionRequest{
		Model:       openaimodel, //"gpt-4o",
		Messages:    messages,
		MaxTokens:   4096,
		Temperature: 0,
		Seed:        42,
		N:           1,
	}

	requestJson, err := json.Marshal(request)
	if err != nil {
		iLog.Error(fmt.Sprintf("Error marshalling request: %v", err))
		return result, err
	}

	client := &http.Client{}
	ctx, cancel := context.WithTimeout(context.Background(), 3600*time.Second) // Set a timeout
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(requestJson))
	if err != nil {
		iLog.Error(fmt.Sprintf("Error creating request: %s", err))
		return result, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)

	resp, err := client.Do(req)
	if err != nil {
		iLog.Error(fmt.Sprintf("Error in calling OpenAi API: %s", err))
		return result, err
	}
	defer resp.Body.Close()

	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		iLog.Error(fmt.Sprintf("Error reading response body: %s", err))
		return result, err
	}

	err = json.Unmarshal(respBody, &result)
	if err != nil {
		iLog.Error(fmt.Sprintf("Error unmarshalling response: %s", err))
		return result, err
	}
	//	data := result.Choices[0].Message.Content
	iLog.Debug(fmt.Sprintf("Response data: %v", result))
	return result, nil
}
