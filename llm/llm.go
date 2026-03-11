package llm

// Multi-vendor LLM client — unified interface for OpenAI, Anthropic,
// Azure OpenAI, Google Gemini, and Ollama (local LLMs).
//
// Use NewLLMClient(useCase, fallbackAPIKey, fallbackModel) to get a client
// for a specific use case. Configuration is read from aiconfig.json via
// config.GetAIConfig(). Falls back to OpenAI with the supplied key/model
// when config is unavailable or the use case is not configured.
//
// To add a local Ollama model, set "enabled": true on the "ollama" vendor
// in aiconfig.json and point "api_base_url" to your Ollama server
// (default: "http://localhost:11434").

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/mdaxf/iac/config"
)

// LLMMessage is a single chat turn.
type LLMMessage struct {
	Role    string `json:"role"`    // "system" | "user" | "assistant"
	Content string `json:"content"`
}

// LLMOptions controls per-call generation parameters.
// Nil fields fall back to aiconfig.json or provider defaults.
type LLMOptions struct {
	Temperature *float64
	MaxTokens   *int
}

// LLMClient is a vendor-agnostic chat completion client.
type LLMClient interface {
	Chat(ctx context.Context, messages []LLMMessage, opts *LLMOptions) (string, error)
	Vendor() string
	Model() string
}

// NewLLMClient returns the best client for useCase by reading aiconfig.json.
// If aiconfig is not loaded or the use case is not configured, it falls back
// to an OpenAI client using fallbackAPIKey and fallbackModel.
func NewLLMClient(useCase, fallbackAPIKey, fallbackModel string) (LLMClient, error) {
	cfg := config.GetAIConfig()
	if cfg != nil {
		if ucCfg, ok := cfg.UseCases[useCase]; ok {
			vendor := ucCfg.Vendor
			if vcCfg, ok := cfg.AIVendors[vendor]; ok && vcCfg.Enabled {
				model := ucCfg.ModelOverride
				if model == "" {
					model = vcCfg.Models[useCase]
				}
				switch vendor {
				case "openai":
					baseURL := vcCfg.APIBaseURL
					if baseURL == "" {
						baseURL = "https://api.openai.com/v1"
					}
					return &openAIClient{apiKey: vcCfg.APIKey, baseURL: baseURL, model: model}, nil
				case "anthropic":
					baseURL := vcCfg.APIBaseURL
					if baseURL == "" {
						baseURL = "https://api.anthropic.com/v1"
					}
					return &anthropicClient{apiKey: vcCfg.APIKey, baseURL: baseURL, model: model}, nil
				case "azure_openai":
					return &azureOpenAIClient{
						apiKey:     vcCfg.APIKey,
						baseURL:    vcCfg.APIBaseURL,
						deployment: vcCfg.Deployment,
						apiVersion: vcCfg.APIVersion,
						model:      model,
					}, nil
				case "google":
					return &googleClient{apiKey: vcCfg.APIKey, model: model}, nil
				case "ollama":
					baseURL := vcCfg.APIBaseURL
					if baseURL == "" {
						baseURL = "http://localhost:11434"
					}
					return &ollamaClient{baseURL: baseURL, model: model}, nil
				}
			}
		}
	}

	// Fallback: OpenAI with supplied key/model
	if fallbackAPIKey == "" {
		return nil, fmt.Errorf("no AI configuration available for use case %q and no fallback API key provided", useCase)
	}
	if fallbackModel == "" {
		fallbackModel = "gpt-4o"
	}
	return &openAIClient{
		apiKey:  fallbackAPIKey,
		baseURL: "https://api.openai.com/v1",
		model:   fallbackModel,
	}, nil
}

// CallLLM is a convenience wrapper that creates a client and calls Chat in one step.
// messages uses the generic map format accepted by existing callers in codegen.
func CallLLM(ctx context.Context, useCase, fallbackAPIKey, fallbackModel string,
	messages []map[string]interface{}, temperature float64) (string, error) {

	client, err := NewLLMClient(useCase, fallbackAPIKey, fallbackModel)
	if err != nil {
		return "", err
	}

	llmMsgs := make([]LLMMessage, 0, len(messages))
	for _, m := range messages {
		role, _ := m["role"].(string)
		content, _ := m["content"].(string)
		llmMsgs = append(llmMsgs, LLMMessage{Role: role, Content: content})
	}

	opts := &LLMOptions{Temperature: &temperature}
	return client.Chat(ctx, llmMsgs, opts)
}

// httpClient is the shared HTTP client with a generous timeout.
var httpClient = &http.Client{Timeout: 120 * time.Second}

// ─── OpenAI ──────────────────────────────────────────────────────────────────

type openAIClient struct {
	apiKey  string
	baseURL string
	model   string
}

func (c *openAIClient) Vendor() string { return "openai" }
func (c *openAIClient) Model() string  { return c.model }

func (c *openAIClient) Chat(ctx context.Context, messages []LLMMessage, opts *LLMOptions) (string, error) {
	temp := 0.7
	maxTok := 2000
	if opts != nil {
		if opts.Temperature != nil {
			temp = *opts.Temperature
		}
		if opts.MaxTokens != nil {
			maxTok = *opts.MaxTokens
		}
	}

	body := map[string]interface{}{
		"model":       c.model,
		"messages":    messages,
		"temperature": temp,
		"max_tokens":  maxTok,
	}
	data, err := json.Marshal(body)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/chat/completions", bytes.NewReader(data))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.apiKey)

	resp, err := httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("openai request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, _ := ioutil.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("openai error %d: %s", resp.StatusCode, string(respBody))
	}

	var out struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}
	if err := json.Unmarshal(respBody, &out); err != nil {
		return "", fmt.Errorf("openai response parse error: %w", err)
	}
	if len(out.Choices) == 0 {
		return "", fmt.Errorf("openai returned no choices")
	}
	return out.Choices[0].Message.Content, nil
}

// ─── Anthropic ────────────────────────────────────────────────────────────────

type anthropicClient struct {
	apiKey  string
	baseURL string
	model   string
}

func (c *anthropicClient) Vendor() string { return "anthropic" }
func (c *anthropicClient) Model() string  { return c.model }

func (c *anthropicClient) Chat(ctx context.Context, messages []LLMMessage, opts *LLMOptions) (string, error) {
	maxTok := 4000
	if opts != nil && opts.MaxTokens != nil {
		maxTok = *opts.MaxTokens
	}

	// Anthropic separates the system message from the conversation turns
	var system string
	var turns []map[string]string
	for _, m := range messages {
		if m.Role == "system" {
			system = m.Content
		} else {
			turns = append(turns, map[string]string{"role": m.Role, "content": m.Content})
		}
	}

	reqBody := map[string]interface{}{
		"model":      c.model,
		"messages":   turns,
		"max_tokens": maxTok,
	}
	if system != "" {
		reqBody["system"] = system
	}
	if opts != nil && opts.Temperature != nil {
		reqBody["temperature"] = *opts.Temperature
	}

	data, err := json.Marshal(reqBody)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/messages", bytes.NewReader(data))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", c.apiKey)
	req.Header.Set("anthropic-version", "2023-06-01")

	resp, err := httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("anthropic request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, _ := ioutil.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("anthropic error %d: %s", resp.StatusCode, string(respBody))
	}

	var out struct {
		Content []struct {
			Type string `json:"type"`
			Text string `json:"text"`
		} `json:"content"`
	}
	if err := json.Unmarshal(respBody, &out); err != nil {
		return "", fmt.Errorf("anthropic response parse error: %w", err)
	}
	for _, block := range out.Content {
		if block.Type == "text" {
			return block.Text, nil
		}
	}
	return "", fmt.Errorf("anthropic returned no text content")
}

// ─── Azure OpenAI ─────────────────────────────────────────────────────────────

type azureOpenAIClient struct {
	apiKey     string
	baseURL    string
	deployment string
	apiVersion string
	model      string
}

func (c *azureOpenAIClient) Vendor() string { return "azure_openai" }
func (c *azureOpenAIClient) Model() string  { return c.model }

func (c *azureOpenAIClient) Chat(ctx context.Context, messages []LLMMessage, opts *LLMOptions) (string, error) {
	temp := 0.7
	maxTok := 2000
	if opts != nil {
		if opts.Temperature != nil {
			temp = *opts.Temperature
		}
		if opts.MaxTokens != nil {
			maxTok = *opts.MaxTokens
		}
	}

	deployment := c.deployment
	if deployment == "" {
		deployment = c.model
	}
	apiVersion := c.apiVersion
	if apiVersion == "" {
		apiVersion = "2024-02-15-preview"
	}

	url := fmt.Sprintf("%s/openai/deployments/%s/chat/completions?api-version=%s",
		strings.TrimRight(c.baseURL, "/"), deployment, apiVersion)

	body := map[string]interface{}{
		"messages":    messages,
		"temperature": temp,
		"max_tokens":  maxTok,
	}
	data, err := json.Marshal(body)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(data))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("api-key", c.apiKey)

	resp, err := httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("azure openai request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, _ := ioutil.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("azure openai error %d: %s", resp.StatusCode, string(respBody))
	}

	var out struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}
	if err := json.Unmarshal(respBody, &out); err != nil {
		return "", fmt.Errorf("azure openai response parse error: %w", err)
	}
	if len(out.Choices) == 0 {
		return "", fmt.Errorf("azure openai returned no choices")
	}
	return out.Choices[0].Message.Content, nil
}

// ─── Google Gemini ────────────────────────────────────────────────────────────

type googleClient struct {
	apiKey string
	model  string
}

func (c *googleClient) Vendor() string { return "google" }
func (c *googleClient) Model() string  { return c.model }

func (c *googleClient) Chat(ctx context.Context, messages []LLMMessage, opts *LLMOptions) (string, error) {
	// Convert messages to Gemini format
	type part struct {
		Text string `json:"text"`
	}
	type content struct {
		Role  string `json:"role"`
		Parts []part `json:"parts"`
	}

	var systemText string
	var contents []content
	for _, m := range messages {
		if m.Role == "system" {
			systemText = m.Content
			continue
		}
		role := m.Role
		if role == "assistant" {
			role = "model"
		}
		contents = append(contents, content{Role: role, Parts: []part{{Text: m.Content}}})
	}

	// Prepend system message as a user turn if present
	if systemText != "" && len(contents) > 0 {
		contents[0].Parts[0].Text = systemText + "\n\n" + contents[0].Parts[0].Text
	}

	reqBody := map[string]interface{}{
		"contents": contents,
	}
	if opts != nil {
		gc := map[string]interface{}{}
		if opts.Temperature != nil {
			gc["temperature"] = *opts.Temperature
		}
		if opts.MaxTokens != nil {
			gc["maxOutputTokens"] = *opts.MaxTokens
		}
		if len(gc) > 0 {
			reqBody["generationConfig"] = gc
		}
	}

	data, err := json.Marshal(reqBody)
	if err != nil {
		return "", err
	}

	url := fmt.Sprintf("https://generativelanguage.googleapis.com/v1beta/models/%s:generateContent?key=%s",
		c.model, c.apiKey)

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(data))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("google gemini request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, _ := ioutil.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("google gemini error %d: %s", resp.StatusCode, string(respBody))
	}

	var out struct {
		Candidates []struct {
			Content struct {
				Parts []struct {
					Text string `json:"text"`
				} `json:"parts"`
			} `json:"content"`
		} `json:"candidates"`
	}
	if err := json.Unmarshal(respBody, &out); err != nil {
		return "", fmt.Errorf("google gemini response parse error: %w", err)
	}
	if len(out.Candidates) == 0 || len(out.Candidates[0].Content.Parts) == 0 {
		return "", fmt.Errorf("google gemini returned no candidates")
	}
	return out.Candidates[0].Content.Parts[0].Text, nil
}

// ─── Ollama (local LLM) ───────────────────────────────────────────────────────

type ollamaClient struct {
	baseURL string
	model   string
}

func (c *ollamaClient) Vendor() string { return "ollama" }
func (c *ollamaClient) Model() string  { return c.model }

func (c *ollamaClient) Chat(ctx context.Context, messages []LLMMessage, opts *LLMOptions) (string, error) {
	type ollamaMsg struct {
		Role    string `json:"role"`
		Content string `json:"content"`
	}

	turns := make([]ollamaMsg, len(messages))
	for i, m := range messages {
		turns[i] = ollamaMsg{Role: m.Role, Content: m.Content}
	}

	reqBody := map[string]interface{}{
		"model":    c.model,
		"messages": turns,
		"stream":   false,
	}
	if opts != nil {
		options := map[string]interface{}{}
		if opts.Temperature != nil {
			options["temperature"] = *opts.Temperature
		}
		if opts.MaxTokens != nil {
			options["num_predict"] = *opts.MaxTokens
		}
		if len(options) > 0 {
			reqBody["options"] = options
		}
	}

	data, err := json.Marshal(reqBody)
	if err != nil {
		return "", err
	}

	baseURL := strings.TrimRight(c.baseURL, "/")
	req, err := http.NewRequestWithContext(ctx, "POST", baseURL+"/api/chat", bytes.NewReader(data))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("ollama request failed — is Ollama running at %s? error: %w", c.baseURL, err)
	}
	defer resp.Body.Close()

	respBody, _ := ioutil.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("ollama error %d: %s", resp.StatusCode, string(respBody))
	}

	var out struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	}
	if err := json.Unmarshal(respBody, &out); err != nil {
		return "", fmt.Errorf("ollama response parse error: %w", err)
	}
	return out.Message.Content, nil
}
