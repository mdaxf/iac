// Copyright 2023 IAC. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package services

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/mdaxf/iac/config"
	"github.com/mdaxf/iac/logger"
	openai "github.com/sashabaranov/go-openai"
)

// PlantStudioAIService handles AI-powered 3D model generation for Plant Studio
type PlantStudioAIService struct {
	iLog logger.Log
}

// PrimitivePart represents a geometric primitive in 3D space
type PrimitivePart struct {
	Type      string    `json:"type"`      // BOX, CYLINDER, SPHERE, CONE, TORUS, TRIANGLE
	Position  []float64 `json:"position"`  // [x, y, z]
	Rotation  []float64 `json:"rotation"`  // [x, y, z] in radians
	Scale     []float64 `json:"scale"`     // [x, y, z]
	Color     string    `json:"color"`     // hex color
	Operation string    `json:"operation"` // ADD or SUBTRACT
	Opacity   float64   `json:"opacity"`
	Wireframe bool      `json:"wireframe"`
	Emissive  string    `json:"emissive"`
	Name      string    `json:"name"`
}

// PlantObject represents a complete 3D object
type PlantObject struct {
	Type     string          `json:"type"`
	Name     string          `json:"name"`
	Position []float64       `json:"position"`
	Parts    []PrimitivePart `json:"parts"`
	Width    float64         `json:"width"`
	Height   float64         `json:"height"`
	Depth    float64         `json:"depth"`
	Status   string          `json:"status"`
}

// TextTo3DRequest represents a text-to-3D generation request
type TextTo3DRequest struct {
	Description string `json:"description"`
}

// TextTo3DResponse represents a text-to-3D generation response
type TextTo3DResponse struct {
	Assets []PlantObject `json:"assets"`
}

// ImageTo3DRequest represents an image-to-3D generation request
type ImageTo3DRequest struct {
	Image string `json:"image"` // base64 encoded image
}

// ImageTo3DResponse represents an image-to-3D generation response
type ImageTo3DResponse struct {
	Asset PlantObject `json:"asset"`
}

// Google Gemini API structures
type GeminiRequest struct {
	Contents         []GeminiContent         `json:"contents"`
	GenerationConfig GeminiGenerationConfig  `json:"generationConfig,omitempty"`
	SystemInstruction *GeminiSystemInstruction `json:"systemInstruction,omitempty"`
}

type GeminiSystemInstruction struct {
	Parts []GeminiPart `json:"parts"`
}

type GeminiContent struct {
	Parts []GeminiPart `json:"parts"`
	Role  string       `json:"role,omitempty"`
}

type GeminiPart struct {
	Text       string          `json:"text,omitempty"`
	InlineData *GeminiInlineData `json:"inlineData,omitempty"`
}

type GeminiInlineData struct {
	MimeType string `json:"mimeType"`
	Data     string `json:"data"`
}

type GeminiGenerationConfig struct {
	Temperature      float64              `json:"temperature,omitempty"`
	MaxOutputTokens  int                  `json:"maxOutputTokens,omitempty"`
	ResponseMimeType string               `json:"responseMimeType,omitempty"`
	ResponseSchema   *GeminiResponseSchema `json:"responseSchema,omitempty"`
}

type GeminiResponseSchema struct {
	Type       string                        `json:"type"`
	Properties map[string]GeminiSchemaProperty `json:"properties,omitempty"`
	Items      *GeminiSchemaProperty         `json:"items,omitempty"`
}

type GeminiSchemaProperty struct {
	Type       string                        `json:"type"`
	Properties map[string]GeminiSchemaProperty `json:"properties,omitempty"`
	Items      *GeminiSchemaProperty         `json:"items,omitempty"`
	Enum       []string                      `json:"enum,omitempty"`
}

type GeminiResponse struct {
	Candidates []GeminiCandidate `json:"candidates"`
}

type GeminiCandidate struct {
	Content GeminiContent `json:"content"`
}

// NewPlantStudioAIService creates a new Plant Studio AI service
func NewPlantStudioAIService() *PlantStudioAIService {
	return &PlantStudioAIService{
		iLog: logger.Log{
			ModuleName:     logger.Framework,
			User:           "System",
			ControllerName: "PlantStudioAIService",
		},
	}
}

// GenerateFromText generates 3D models from text description
func (s *PlantStudioAIService) GenerateFromText(ctx context.Context, req TextTo3DRequest) (*TextTo3DResponse, error) {
	s.iLog.Info(fmt.Sprintf("GenerateFromText START - Description: %s", req.Description))

	// Get AI configuration for plant_studio_text_to_3d use case
	aiConfig := config.GetAIConfig()
	if aiConfig == nil {
		return nil, fmt.Errorf("AI configuration not loaded")
	}

	useCaseName := "plant_studio_text_to_3d"
	useCaseConfig, exists := aiConfig.UseCases[useCaseName]
	if !exists {
		return nil, fmt.Errorf("use case '%s' not found in AI configuration", useCaseName)
	}

	vendor := useCaseConfig.Vendor
	vendorConfig, exists := aiConfig.AIVendors[vendor]
	if !exists || !vendorConfig.Enabled {
		return nil, fmt.Errorf("vendor '%s' not found or disabled", vendor)
	}

	// Get model for this use case
	model := useCaseConfig.ModelOverride
	if model == "" {
		model, exists = vendorConfig.Models[useCaseName]
		if !exists {
			return nil, fmt.Errorf("no model configured for use case %s", useCaseName)
		}
	}

	s.iLog.Info(fmt.Sprintf("Using vendor: %s, model: %s", vendor, model))

	// Call appropriate AI vendor
	var assets []PlantObject
	var err error

	switch vendor {
	case "google":
		assets, err = s.generateFromTextGoogle(ctx, vendorConfig, model, useCaseConfig.SystemPrompt, req)
	case "openai":
		assets, err = s.generateFromTextOpenAI(ctx, vendorConfig, model, useCaseConfig.SystemPrompt, req)
	default:
		return nil, fmt.Errorf("vendor '%s' not supported for Plant Studio 3D generation", vendor)
	}

	if err != nil {
		s.iLog.Error(fmt.Sprintf("Failed to generate 3D from text: %v", err))
		return nil, err
	}

	s.iLog.Info(fmt.Sprintf("GenerateFromText COMPLETE - Generated %d assets", len(assets)))

	return &TextTo3DResponse{
		Assets: assets,
	}, nil
}

// GenerateFromImage generates 3D model from image
func (s *PlantStudioAIService) GenerateFromImage(ctx context.Context, req ImageTo3DRequest) (*ImageTo3DResponse, error) {
	s.iLog.Info("GenerateFromImage START")

	// Get AI configuration for plant_studio_image_to_3d use case
	aiConfig := config.GetAIConfig()
	if aiConfig == nil {
		return nil, fmt.Errorf("AI configuration not loaded")
	}

	useCaseName := "plant_studio_image_to_3d"
	useCaseConfig, exists := aiConfig.UseCases[useCaseName]
	if !exists {
		return nil, fmt.Errorf("use case '%s' not found in AI configuration", useCaseName)
	}

	vendor := useCaseConfig.Vendor
	vendorConfig, exists := aiConfig.AIVendors[vendor]
	if !exists || !vendorConfig.Enabled {
		return nil, fmt.Errorf("vendor '%s' not found or disabled", vendor)
	}

	// Get model for this use case
	model := useCaseConfig.ModelOverride
	if model == "" {
		model, exists = vendorConfig.Models[useCaseName]
		if !exists {
			return nil, fmt.Errorf("no model configured for use case %s", useCaseName)
		}
	}

	s.iLog.Info(fmt.Sprintf("Using vendor: %s, model: %s", vendor, model))

	// Call appropriate AI vendor
	var asset PlantObject
	var err error

	switch vendor {
	case "google":
		asset, err = s.generateFromImageGoogle(ctx, vendorConfig, model, useCaseConfig.SystemPrompt, req)
	case "openai":
		asset, err = s.generateFromImageOpenAI(ctx, vendorConfig, model, useCaseConfig.SystemPrompt, req)
	default:
		return nil, fmt.Errorf("vendor '%s' not supported for Plant Studio 3D generation", vendor)
	}

	if err != nil {
		s.iLog.Error(fmt.Sprintf("Failed to generate 3D from image: %v", err))
		return nil, err
	}

	s.iLog.Info("GenerateFromImage COMPLETE")

	return &ImageTo3DResponse{
		Asset: asset,
	}, nil
}

// generateFromTextGoogle generates 3D models using Google Gemini from text
func (s *PlantStudioAIService) generateFromTextGoogle(ctx context.Context, vendorConfig config.AIVendorConfig, model, systemPrompt string, req TextTo3DRequest) ([]PlantObject, error) {
	// Build the request prompt
	userPrompt := fmt.Sprintf(`Design a 3D industrial asset based on this description: "%s".
Return a complex mechanical assembly using 40-80 primitive parts.
The assembly should look professional and highly detailed.
Return JSON: { "assets": [ { "name": string, "position": [x,y,z], "parts": [...] } ] }`, req.Description)

	// Build Gemini API request with structured output
	geminiReq := GeminiRequest{
		SystemInstruction: &GeminiSystemInstruction{
			Parts: []GeminiPart{
				{Text: systemPrompt},
			},
		},
		Contents: []GeminiContent{
			{
				Parts: []GeminiPart{
					{Text: userPrompt},
				},
			},
		},
		GenerationConfig: GeminiGenerationConfig{
			Temperature:      0.7,
			MaxOutputTokens:  4000,
			ResponseMimeType: "application/json",
			ResponseSchema: &GeminiResponseSchema{
				Type: "object",
				Properties: map[string]GeminiSchemaProperty{
					"assets": {
						Type: "array",
						Items: &GeminiSchemaProperty{
							Type: "object",
							Properties: map[string]GeminiSchemaProperty{
								"name": {Type: "string"},
								"position": {
									Type: "array",
									Items: &GeminiSchemaProperty{Type: "number"},
								},
								"parts": {
									Type: "array",
									Items: &GeminiSchemaProperty{
										Type: "object",
										Properties: map[string]GeminiSchemaProperty{
											"type":      {Type: "string", Enum: []string{"BOX", "CYLINDER", "SPHERE", "CONE", "TORUS", "TRIANGLE"}},
											"position":  {Type: "array", Items: &GeminiSchemaProperty{Type: "number"}},
											"rotation":  {Type: "array", Items: &GeminiSchemaProperty{Type: "number"}},
											"scale":     {Type: "array", Items: &GeminiSchemaProperty{Type: "number"}},
											"color":     {Type: "string"},
											"operation": {Type: "string", Enum: []string{"ADD", "SUBTRACT"}},
											"opacity":   {Type: "number"},
											"wireframe": {Type: "boolean"},
											"emissive":  {Type: "string"},
											"name":      {Type: "string"},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}

	// Call Gemini API
	responseText, err := s.callGeminiAPI(ctx, vendorConfig, model, geminiReq)
	if err != nil {
		return nil, err
	}

	// Parse response
	var result struct {
		Assets []struct {
			Name     string          `json:"name"`
			Position []float64       `json:"position"`
			Parts    []PrimitivePart `json:"parts"`
		} `json:"assets"`
	}

	if err := json.Unmarshal([]byte(responseText), &result); err != nil {
		s.iLog.Error(fmt.Sprintf("Failed to parse Gemini response: %v", err))
		return nil, fmt.Errorf("failed to parse AI response: %w", err)
	}

	// Convert to PlantObject format and normalize
	assets := make([]PlantObject, 0, len(result.Assets))
	for _, asset := range result.Assets {
		// Sanitize and normalize parts
		sanitized := s.sanitizeParts(asset.Parts)
		normalized := s.normalizeAsset(sanitized)

		plantObj := PlantObject{
			Type:     "GENERATED",
			Name:     asset.Name,
			Position: asset.Position,
			Parts:    normalized.Parts,
			Width:    normalized.Width,
			Height:   normalized.Height,
			Depth:    normalized.Depth,
			Status:   "IDLE",
		}
		assets = append(assets, plantObj)
	}

	return assets, nil
}

// generateFromImageGoogle generates 3D model using Google Gemini from image
func (s *PlantStudioAIService) generateFromImageGoogle(ctx context.Context, vendorConfig config.AIVendorConfig, model, systemPrompt string, req ImageTo3DRequest) (PlantObject, error) {
	// Build the request prompt
	userPrompt := `As a World-Class Mechanical Engineer, analyze the industrial asset in this image.
Reconstruct it as a high-fidelity 3D Digital Twin using a complex assembly of 30-60 geometric primitives.

CONSTRAINTS:
1. USE HIERARCHY: Combine BOX, CYLINDER, TORUS, and TRIANGLE primitives.
2. COLORS: Use realistic industrial palettes (grays, safety oranges, deep blues).
3. DETAIL: Add small details like control panels (small BOXES), pipes (TORUS/CYLINDER), and structural supports (TRIANGLE).
4. CENTER: Ensure the root of the assembly is at [0,0,0].
5. FORMAT: Return ONLY a valid JSON object with a "parts" array.`

	// Build Gemini API request with image and structured output
	geminiReq := GeminiRequest{
		SystemInstruction: &GeminiSystemInstruction{
			Parts: []GeminiPart{
				{Text: systemPrompt},
			},
		},
		Contents: []GeminiContent{
			{
				Parts: []GeminiPart{
					{
						InlineData: &GeminiInlineData{
							MimeType: "image/jpeg",
							Data:     req.Image,
						},
					},
					{Text: userPrompt},
				},
			},
		},
		GenerationConfig: GeminiGenerationConfig{
			Temperature:      0.7,
			MaxOutputTokens:  4000,
			ResponseMimeType: "application/json",
			ResponseSchema: &GeminiResponseSchema{
				Type: "object",
				Properties: map[string]GeminiSchemaProperty{
					"parts": {
						Type: "array",
						Items: &GeminiSchemaProperty{
							Type: "object",
							Properties: map[string]GeminiSchemaProperty{
								"type":      {Type: "string", Enum: []string{"BOX", "CYLINDER", "SPHERE", "CONE", "TORUS", "TRIANGLE"}},
								"position":  {Type: "array", Items: &GeminiSchemaProperty{Type: "number"}},
								"rotation":  {Type: "array", Items: &GeminiSchemaProperty{Type: "number"}},
								"scale":     {Type: "array", Items: &GeminiSchemaProperty{Type: "number"}},
								"color":     {Type: "string"},
								"operation": {Type: "string", Enum: []string{"ADD", "SUBTRACT"}},
								"opacity":   {Type: "number"},
								"wireframe": {Type: "boolean"},
								"emissive":  {Type: "string"},
								"name":      {Type: "string"},
							},
						},
					},
				},
			},
		},
	}

	// Call Gemini API
	responseText, err := s.callGeminiAPI(ctx, vendorConfig, model, geminiReq)
	if err != nil {
		return PlantObject{}, err
	}

	// Parse response
	var result struct {
		Parts []PrimitivePart `json:"parts"`
	}

	if err := json.Unmarshal([]byte(responseText), &result); err != nil {
		s.iLog.Error(fmt.Sprintf("Failed to parse Gemini response: %v", err))
		return PlantObject{}, fmt.Errorf("failed to parse AI response: %w", err)
	}

	// Sanitize and normalize parts
	sanitized := s.sanitizeParts(result.Parts)
	normalized := s.normalizeAsset(sanitized)

	return PlantObject{
		Type:   "GENERATED",
		Name:   "AI Vision Asset",
		Parts:  normalized.Parts,
		Width:  normalized.Width,
		Height: normalized.Height,
		Depth:  normalized.Depth,
		Status: "IDLE",
	}, nil
}

// callGeminiAPI makes a request to the Google Gemini API
func (s *PlantStudioAIService) callGeminiAPI(ctx context.Context, vendorConfig config.AIVendorConfig, model string, req GeminiRequest) (string, error) {
	// Build API URL
	apiURL := fmt.Sprintf("https://generativelanguage.googleapis.com/v1beta/models/%s:generateContent?key=%s",
		model, vendorConfig.APIKey)

	// Marshal request
	reqBody, err := json.Marshal(req)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create HTTP request
	httpReq, err := http.NewRequestWithContext(ctx, "POST", apiURL, bytes.NewBuffer(reqBody))
	if err != nil {
		return "", fmt.Errorf("failed to create HTTP request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")

	// Make HTTP request
	client := &http.Client{Timeout: 60 * time.Second}
	resp, err := client.Do(httpReq)
	if err != nil {
		return "", fmt.Errorf("failed to call Gemini API: %w", err)
	}
	defer resp.Body.Close()

	// Read response body
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		s.iLog.Error(fmt.Sprintf("Gemini API error: %s", string(respBody)))
		return "", fmt.Errorf("Gemini API returned status %d: %s", resp.StatusCode, string(respBody))
	}

	// Parse Gemini response
	var geminiResp GeminiResponse
	if err := json.Unmarshal(respBody, &geminiResp); err != nil {
		return "", fmt.Errorf("failed to parse Gemini response: %w", err)
	}

	if len(geminiResp.Candidates) == 0 || len(geminiResp.Candidates[0].Content.Parts) == 0 {
		return "", fmt.Errorf("no response from Gemini API")
	}

	return geminiResp.Candidates[0].Content.Parts[0].Text, nil
}

// sanitizeParts validates and sanitizes primitive parts
func (s *PlantStudioAIService) sanitizeParts(parts []PrimitivePart) []PrimitivePart {
	if len(parts) == 0 {
		// Return fallback default parts
		return []PrimitivePart{
			{Type: "BOX", Position: []float64{0, 0.25, 0}, Rotation: []float64{0, 0, 0}, Scale: []float64{1, 0.5, 1}, Color: "#34495e", Name: "Base"},
			{Type: "CYLINDER", Position: []float64{0, 0.75, 0}, Rotation: []float64{0, 0, 0}, Scale: []float64{0.2, 1, 0.2}, Color: "#7f8c8d", Name: "Mast"},
		}
	}

	sanitized := make([]PrimitivePart, 0, len(parts))
	validTypes := map[string]bool{"BOX": true, "CYLINDER": true, "SPHERE": true, "CONE": true, "TORUS": true, "TRIANGLE": true}

	for _, p := range parts {
		part := PrimitivePart{
			Type:      "BOX",
			Position:  []float64{0, 0, 0},
			Rotation:  []float64{0, 0, 0},
			Scale:     []float64{1, 1, 1},
			Color:     "#717171",
			Operation: "ADD",
			Opacity:   1.0,
			Wireframe: false,
			Emissive:  "black",
			Name:      "Component",
		}

		// Validate and set type
		if validTypes[p.Type] {
			part.Type = p.Type
		}

		// Validate and set position
		if len(p.Position) == 3 {
			part.Position = p.Position
		}

		// Validate and set rotation
		if len(p.Rotation) == 3 {
			part.Rotation = p.Rotation
		}

		// Validate and set scale
		if len(p.Scale) == 3 {
			part.Scale = p.Scale
		}

		// Set other properties
		if p.Color != "" {
			part.Color = p.Color
		}
		if p.Operation == "SUBTRACT" {
			part.Operation = "SUBTRACT"
		}
		if p.Opacity > 0 && p.Opacity <= 1 {
			part.Opacity = p.Opacity
		}
		part.Wireframe = p.Wireframe
		if p.Emissive != "" {
			part.Emissive = p.Emissive
		}
		if p.Name != "" {
			part.Name = p.Name
		}

		sanitized = append(sanitized, part)
	}

	return sanitized
}

// normalizeAsset normalizes asset dimensions and centers it
func (s *PlantStudioAIService) normalizeAsset(parts []PrimitivePart) struct {
	Parts  []PrimitivePart
	Width  float64
	Height float64
	Depth  float64
} {
	if len(parts) == 0 {
		return struct {
			Parts  []PrimitivePart
			Width  float64
			Height float64
			Depth  float64
		}{Parts: parts, Width: 1, Height: 1, Depth: 1}
	}

	// Calculate bounding box
	minX, minY, minZ := 1e10, 1e10, 1e10
	maxX, maxY, maxZ := -1e10, -1e10, -1e10

	for _, p := range parts {
		px, py, pz := p.Position[0], p.Position[1], p.Position[2]
		sx, sy, sz := p.Scale[0]/2, p.Scale[1]/2, p.Scale[2]/2

		if px-sx < minX {
			minX = px - sx
		}
		if px+sx > maxX {
			maxX = px + sx
		}
		if py-sy < minY {
			minY = py - sy
		}
		if py+sy > maxY {
			maxY = py + sy
		}
		if pz-sz < minZ {
			minZ = pz - sz
		}
		if pz+sz > maxZ {
			maxZ = pz + sz
		}
	}

	// Handle flat or zero-size objects
	rawWidth := maxX - minX
	if rawWidth < 0.1 {
		rawWidth = 0.1
	}
	rawHeight := maxY - minY
	if rawHeight < 0.1 {
		rawHeight = 0.1
	}
	rawDepth := maxZ - minZ
	if rawDepth < 0.1 {
		rawDepth = 0.1
	}

	// Center offset: Move center-bottom to [0,0,0]
	offsetX := -(minX + rawWidth/2)
	offsetY := -minY
	offsetZ := -(minZ + rawDepth/2)

	// Apply offset to all parts
	normalizedParts := make([]PrimitivePart, len(parts))
	for i, p := range parts {
		normalizedParts[i] = p
		normalizedParts[i].Position = []float64{
			p.Position[0] + offsetX,
			p.Position[1] + offsetY,
			p.Position[2] + offsetZ,
		}
	}

	return struct {
		Parts  []PrimitivePart
		Width  float64
		Height float64
		Depth  float64
	}{
		Parts:  normalizedParts,
		Width:  float64(int(rawWidth*100)) / 100, // Round to 2 decimals
		Height: float64(int(rawHeight*100)) / 100,
		Depth:  float64(int(rawDepth*100)) / 100,
	}
}

// generateFromTextOpenAI generates 3D models using OpenAI from text
func (s *PlantStudioAIService) generateFromTextOpenAI(ctx context.Context, vendorConfig config.AIVendorConfig, model, systemPrompt string, req TextTo3DRequest) ([]PlantObject, error) {
	client := openai.NewClient(vendorConfig.APIKey)

	// Build user prompt
	userPrompt := fmt.Sprintf(`Design a 3D industrial asset based on this description: "%s".
Return a complex mechanical assembly using 40-80 primitive parts.
The assembly should look professional and highly detailed.
Return JSON: { "assets": [ { "name": string, "position": [x,y,z], "parts": [...] } ] }

Each part should have:
- type: one of BOX, CYLINDER, SPHERE, CONE, TORUS, TRIANGLE
- position: [x, y, z] coordinates
- rotation: [x, y, z] in radians
- scale: [x, y, z] dimensions
- color: hex color string
- operation: ADD or SUBTRACT
- opacity: 0-1
- wireframe: boolean
- emissive: hex color string
- name: descriptive name`, req.Description)

	messages := []openai.ChatCompletionMessage{
		{
			Role:    openai.ChatMessageRoleSystem,
			Content: systemPrompt,
		},
		{
			Role:    openai.ChatMessageRoleUser,
			Content: userPrompt,
		},
	}

	// Get temperature from vendor config or default
	temperature := 0.7
	if temp, ok := vendorConfig.Parameters["temperature"].(float64); ok {
		temperature = temp
	}

	// Get max tokens from vendor config or default
	maxTokens := 4000
	if tokens, ok := vendorConfig.Parameters["max_tokens"].(float64); ok {
		maxTokens = int(tokens)
	}

	// Create chat completion request
	resp, err := client.CreateChatCompletion(
		ctx,
		openai.ChatCompletionRequest{
			Model:       model,
			Messages:    messages,
			Temperature: float32(temperature),
			MaxTokens:   maxTokens,
		},
	)

	if err != nil {
		return nil, fmt.Errorf("OpenAI API error: %w", err)
	}

	if len(resp.Choices) == 0 {
		return nil, fmt.Errorf("no response from OpenAI")
	}

	// Parse response
	content := resp.Choices[0].Message.Content
	s.iLog.Debug(fmt.Sprintf("OpenAI response: %s", content))

	// Extract JSON from markdown code blocks if present
	content = s.extractJSON(content)

	var result struct {
		Assets []struct {
			Name     string          `json:"name"`
			Position []float64       `json:"position"`
			Parts    []PrimitivePart `json:"parts"`
		} `json:"assets"`
	}

	if err := json.Unmarshal([]byte(content), &result); err != nil {
		s.iLog.Error(fmt.Sprintf("Failed to parse OpenAI response: %v", err))
		return nil, fmt.Errorf("failed to parse AI response: %w", err)
	}

	// Convert to PlantObject format and normalize
	assets := make([]PlantObject, 0, len(result.Assets))
	for _, asset := range result.Assets {
		// Sanitize and normalize parts
		sanitized := s.sanitizeParts(asset.Parts)
		normalized := s.normalizeAsset(sanitized)

		plantObj := PlantObject{
			Type:     "GENERATED",
			Name:     asset.Name,
			Position: asset.Position,
			Parts:    normalized.Parts,
			Width:    normalized.Width,
			Height:   normalized.Height,
			Depth:    normalized.Depth,
			Status:   "IDLE",
		}
		assets = append(assets, plantObj)
	}

	return assets, nil
}

// generateFromImageOpenAI generates 3D model using OpenAI Vision from image
func (s *PlantStudioAIService) generateFromImageOpenAI(ctx context.Context, vendorConfig config.AIVendorConfig, model, systemPrompt string, req ImageTo3DRequest) (PlantObject, error) {
	client := openai.NewClient(vendorConfig.APIKey)

	// Build user prompt
	userPrompt := `As a World-Class Mechanical Engineer, analyze the industrial asset in this image.
Reconstruct it as a high-fidelity 3D Digital Twin using a complex assembly of 30-60 geometric primitives.

CONSTRAINTS:
1. USE HIERARCHY: Combine BOX, CYLINDER, TORUS, and TRIANGLE primitives.
2. COLORS: Use realistic industrial palettes (grays, safety oranges, deep blues).
3. DETAIL: Add small details like control panels (small BOXES), pipes (TORUS/CYLINDER), and structural supports (TRIANGLE).
4. CENTER: Ensure the root of the assembly is at [0,0,0].
5. FORMAT: Return ONLY a valid JSON object with a "parts" array.

Each part should have:
- type: one of BOX, CYLINDER, SPHERE, CONE, TORUS, TRIANGLE
- position: [x, y, z] coordinates
- rotation: [x, y, z] in radians
- scale: [x, y, z] dimensions
- color: hex color string
- operation: ADD or SUBTRACT
- opacity: 0-1
- wireframe: boolean
- emissive: hex color string
- name: descriptive name

Return JSON: { "parts": [...] }`

	// Prepare image URL - OpenAI expects base64 images as data URLs
	imageURL := fmt.Sprintf("data:image/jpeg;base64,%s", req.Image)

	messages := []openai.ChatCompletionMessage{
		{
			Role:    openai.ChatMessageRoleSystem,
			Content: systemPrompt,
		},
		{
			Role: openai.ChatMessageRoleUser,
			MultiContent: []openai.ChatMessagePart{
				{
					Type: openai.ChatMessagePartTypeImageURL,
					ImageURL: &openai.ChatMessageImageURL{
						URL:    imageURL,
						Detail: openai.ImageURLDetailHigh,
					},
				},
				{
					Type: openai.ChatMessagePartTypeText,
					Text: userPrompt,
				},
			},
		},
	}

	// Get temperature from vendor config or default
	temperature := 0.7
	if temp, ok := vendorConfig.Parameters["temperature"].(float64); ok {
		temperature = temp
	}

	// Get max tokens from vendor config or default
	maxTokens := 4000
	if tokens, ok := vendorConfig.Parameters["max_tokens"].(float64); ok {
		maxTokens = int(tokens)
	}

	// Create chat completion request
	resp, err := client.CreateChatCompletion(
		ctx,
		openai.ChatCompletionRequest{
			Model:       model,
			Messages:    messages,
			Temperature: float32(temperature),
			MaxTokens:   maxTokens,
		},
	)

	if err != nil {
		return PlantObject{}, fmt.Errorf("OpenAI API error: %w", err)
	}

	if len(resp.Choices) == 0 {
		return PlantObject{}, fmt.Errorf("no response from OpenAI")
	}

	// Parse response
	content := resp.Choices[0].Message.Content
	s.iLog.Debug(fmt.Sprintf("OpenAI response: %s", content))

	// Extract JSON from markdown code blocks if present
	content = s.extractJSON(content)

	var result struct {
		Parts []PrimitivePart `json:"parts"`
	}

	if err := json.Unmarshal([]byte(content), &result); err != nil {
		s.iLog.Error(fmt.Sprintf("Failed to parse OpenAI response: %v", err))
		return PlantObject{}, fmt.Errorf("failed to parse AI response: %w", err)
	}

	// Sanitize and normalize parts
	sanitized := s.sanitizeParts(result.Parts)
	normalized := s.normalizeAsset(sanitized)

	return PlantObject{
		Type:   "GENERATED",
		Name:   "AI Vision Asset",
		Parts:  normalized.Parts,
		Width:  normalized.Width,
		Height: normalized.Height,
		Depth:  normalized.Depth,
		Status: "IDLE",
	}, nil
}

// extractJSON removes markdown code blocks from JSON response
func (s *PlantStudioAIService) extractJSON(content string) string {
	// Remove markdown code blocks if present
	if idx := strings.Index(content, "```json"); idx >= 0 {
		content = content[idx+7:]
		if idx := strings.Index(content, "```"); idx >= 0 {
			content = content[:idx]
		}
	} else if idx := strings.Index(content, "```"); idx >= 0 {
		content = content[idx+3:]
		if idx := strings.Index(content, "```"); idx >= 0 {
			content = content[:idx]
		}
	}
	return strings.TrimSpace(content)
}
