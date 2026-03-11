package services

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/mdaxf/iac/config"
	xhtml "golang.org/x/net/html"
)

// registerWebTools registers internet-access and HTTP tools.
//
// Tool registration strategy:
//   - web_fetch       — always registered (no key needed)
//   - http_request    — always registered (no key needed)
//   - search_duckduckgo — always registered (free, no key)
//   - search_brave    — registered when brave.api_key is configured
//   - search_tavily   — registered when tavily.api_key is configured
//   - search_serpapi  — registered when serpapi.api_key is configured
//   - search_google   — registered when google_cse.api_key + cx are configured
//   - web_search      — always registered; auto-routes to the configured primary
//                       provider with fallback; useful when the agent should not
//                       care which provider is used
func (s *AgentToolRegistryService) registerWebTools() {
	sc := config.GetSearchConfig()

	// Common search parameter schema (reused across all provider tools)
	baseSearchParams := ToolParameterSchema{
		Type: "object",
		Properties: map[string]ToolPropertySchema{
			"query":   {Type: "string", Description: "The search query"},
			"count":   {Type: "integer", Description: "Number of results to return (1–10, default 5)"},
			"country": {Type: "string", Description: "Two-letter country code for localised results (e.g. US, GB, AU)"},
		},
		Required: []string{"query"},
	}

	// ── web_fetch ────────────────────────────────────────────────────────────
	s.register(ToolDefinition{
		Type: "function",
		Function: ToolFunctionDef{
			Name:        "web_fetch",
			Description: "Fetches a public web page and returns its readable text content. Use this to retrieve information from a specific URL after a search, or to read documentation and web APIs.",
			Parameters: ToolParameterSchema{
				Type: "object",
				Properties: map[string]ToolPropertySchema{
					"url":       {Type: "string", Description: "The HTTP or HTTPS URL to fetch"},
					"max_chars": {Type: "integer", Description: "Maximum characters to return (default 5000, max 20000)"},
				},
				Required: []string{"url"},
			},
		},
	}, s.webFetchHandler)

	// ── http_request ─────────────────────────────────────────────────────────
	s.register(ToolDefinition{
		Type: "function",
		Function: ToolFunctionDef{
			Name:        "http_request",
			Description: "Makes an HTTP request to an external API or service and returns the status code and response body.",
			Parameters: ToolParameterSchema{
				Type: "object",
				Properties: map[string]ToolPropertySchema{
					"url":     {Type: "string", Description: "The URL to send the request to"},
					"method":  {Type: "string", Description: "HTTP method: GET, POST, PUT, DELETE, PATCH (default GET)"},
					"body":    {Type: "string", Description: "Request body string (for POST/PUT/PATCH requests)"},
					"headers": {Type: "string", Description: `JSON object of additional request headers, e.g. {"Authorization":"Bearer token","Content-Type":"application/json"}`},
				},
				Required: []string{"url"},
			},
		},
	}, s.httpRequestHandler)

	// ── search_duckduckgo — always available, no key required ─────────────────
	s.register(ToolDefinition{
		Type: "function",
		Function: ToolFunctionDef{
			Name: "search_duckduckgo",
			Description: "Searches the web using DuckDuckGo Instant Answers API (free, no API key required). " +
				"Returns Wikipedia summaries and related topics. " +
				"Best for: quick factual lookups, definitions, and public knowledge. " +
				"Limitation: not a full web index — for broad web searches use search_brave, search_tavily, or search_google if configured.",
			Parameters: baseSearchParams,
		},
	}, func(ctx context.Context, args map[string]interface{}) (string, error) {
		return s.providerSearchHandler(ctx, args, "duckduckgo", sc)
	})

	// ── search_brave — always registered; returns error if api_key not set ──────
	s.register(ToolDefinition{
		Type: "function",
		Function: ToolFunctionDef{
			Name: "search_brave",
			Description: "Searches the web using the Brave Search API — an independent, privacy-focused search index. " +
				"Returns titles, URLs, and descriptions for the top results. " +
				"Best for: general web search, news, privacy-conscious queries. " +
				"Free tier: 2,000 queries/month. Requires brave.api_key in aiconfig.json.",
			Parameters: baseSearchParams,
		},
	}, func(ctx context.Context, args map[string]interface{}) (string, error) {
		return s.providerSearchHandler(ctx, args, "brave", sc)
	})

	// ── search_tavily — always registered; returns error if api_key not set ──
	tavilyParams := ToolParameterSchema{
		Type: "object",
		Properties: map[string]ToolPropertySchema{
			"query":        {Type: "string", Description: "The search query"},
			"count":        {Type: "integer", Description: "Number of results to return (1–10, default 5)"},
			"search_depth": {Type: "string", Description: "Search depth: 'basic' (faster, default) or 'advanced' (deeper, slower, uses more quota)"},
		},
		Required: []string{"query"},
	}
	s.register(ToolDefinition{
		Type: "function",
		Function: ToolFunctionDef{
			Name: "search_tavily",
			Description: "Searches the web using Tavily — an AI-optimised search API purpose-built for language model agents. " +
				"Returns pre-cleaned, summarised content ideal for LLM consumption. " +
				"Best for: research tasks, summarisation, fact-finding where content quality matters. " +
				"Free tier: 1,000 queries/month. Requires tavily.api_key in aiconfig.json.",
			Parameters: tavilyParams,
		},
	}, func(ctx context.Context, args map[string]interface{}) (string, error) {
		// Allow per-call search_depth override
		if depth, ok := args["search_depth"].(string); ok && depth != "" {
			localSC := sc
			localSC.Tavily.SearchDepth = depth
			return s.providerSearchHandler(ctx, args, "tavily", localSC)
		}
		return s.providerSearchHandler(ctx, args, "tavily", sc)
	})

	// ── search_serpapi — always registered; returns error if api_key not set ──
	serpParams := ToolParameterSchema{
		Type: "object",
		Properties: map[string]ToolPropertySchema{
			"query":   {Type: "string", Description: "The search query"},
			"count":   {Type: "integer", Description: "Number of results to return (1–10, default 5)"},
			"country": {Type: "string", Description: "Two-letter country code for localised results (e.g. US, GB, AU)"},
			"engine":  {Type: "string", Description: "Search engine to use: google (default), bing, yahoo, duckduckgo, baidu, yandex"},
		},
		Required: []string{"query"},
	}
	s.register(ToolDefinition{
		Type: "function",
		Function: ToolFunctionDef{
			Name: "search_serpapi",
			Description: "Searches the web via SerpAPI, which provides real Google, Bing, and other engine results. " +
				"Specify 'engine' parameter to choose: google (default), bing, yahoo, duckduckgo, baidu, yandex. " +
				"Best for: when Google-quality results are required, multi-engine comparison. " +
				"Free tier: 100 queries/month. Requires serpapi.api_key in aiconfig.json.",
			Parameters: serpParams,
		},
	}, func(ctx context.Context, args map[string]interface{}) (string, error) {
		// Allow per-call engine override
		if engine, ok := args["engine"].(string); ok && engine != "" {
			localSC := sc
			localSC.SerpAPI.Engine = engine
			return s.providerSearchHandler(ctx, args, "serpapi", localSC)
		}
		return s.providerSearchHandler(ctx, args, "serpapi", sc)
	})

	// ── search_google — always registered; returns error if not fully configured
	s.register(ToolDefinition{
		Type: "function",
		Function: ToolFunctionDef{
			Name: "search_google",
			Description: "Searches the web using Google Custom Search Engine (CSE). " +
				"Returns authoritative Google results with titles, URLs, and snippets. " +
				"Best for: when Google search quality is required and a custom search engine is configured. " +
				"Free tier: 100 queries/day. Requires google_cse.api_key and google_cse.cx in aiconfig.json.",
			Parameters: baseSearchParams,
		},
	}, func(ctx context.Context, args map[string]interface{}) (string, error) {
		return s.providerSearchHandler(ctx, args, "google_cse", sc)
	})

	// ── web_search — auto-router using configured primary + fallback ──────────
	primaryDesc := buildSearchToolDescription(sc.Provider)
	s.register(ToolDefinition{
		Type: "function",
		Function: ToolFunctionDef{
			Name:        "web_search",
			Description: primaryDesc,
			Parameters:  baseSearchParams,
		},
	}, s.webSearchHandler)
}

// providerSearchHandler dispatches a search to a specific provider and returns JSON results.
func (s *AgentToolRegistryService) providerSearchHandler(ctx context.Context, args map[string]interface{}, provider string, sc config.SearchConfig) (string, error) {
	query, _ := args["query"].(string)
	if query == "" {
		return "", fmt.Errorf("query is required")
	}
	count := 5
	if v, ok := args["count"].(float64); ok && v > 0 {
		count = int(v)
		if count > 10 {
			count = 10
		}
	}
	country, _ := args["country"].(string)

	results, err := dispatchSearch(ctx, provider, sc, query, count, country)
	if err != nil {
		return "", err
	}
	if results == nil {
		results = []searchResult{}
	}
	b, _ := json.Marshal(results)
	return string(b), nil
}

// buildSearchToolDescription returns a description that names the configured provider.
func buildSearchToolDescription(provider string) string {
	names := map[string]string{
		"brave":      "Brave Search",
		"tavily":     "Tavily (AI-optimised)",
		"serpapi":    "SerpAPI (Google/Bing)",
		"google_cse": "Google Custom Search",
		"duckduckgo": "DuckDuckGo Instant Answers",
	}
	providerName, ok := names[provider]
	if !ok {
		providerName = "the configured search engine"
	}
	return fmt.Sprintf("Searches the web using %s and returns titles, descriptions, and URLs for the top results. "+
		"Configure search providers in aiconfig.json under the 'search' section.", providerName)
}

// ─── web_fetch ────────────────────────────────────────────────────────────────

func (s *AgentToolRegistryService) webFetchHandler(ctx context.Context, args map[string]interface{}) (string, error) {
	rawURL, _ := args["url"].(string)
	if rawURL == "" {
		return "", fmt.Errorf("url is required")
	}

	maxChars := 5000
	if v, ok := args["max_chars"].(float64); ok && v > 0 {
		maxChars = int(v)
		if maxChars > 20000 {
			maxChars = 20000
		}
	}

	parsed, err := url.Parse(rawURL)
	if err != nil {
		return "", fmt.Errorf("invalid URL: %w", err)
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return "", fmt.Errorf("only http and https URLs are supported")
	}
	if isPrivateHost(parsed.Hostname()) {
		return "", fmt.Errorf("access to private or local addresses is not allowed")
	}

	// ── Attempt 1: direct HTTP fetch ──────────────────────────────────────────
	text, fetchErr := s.doHTTPFetch(ctx, rawURL, maxChars)
	if fetchErr == nil {
		return text, nil
	}
	s.iLog.Info(fmt.Sprintf("web_fetch: direct fetch failed for %s (%v) — trying search fallback", rawURL, fetchErr))

	// ── Attempt 2: use available search providers ─────────────────────────────
	// Build a search query from the URL: use the page path as context.
	sc := config.GetSearchConfig()
	providers := getAvailableProviders(sc)

	// Build a query that targets the specific page: "site:hostname path keywords"
	searchQuery := rawURL
	if parsed.Host != "" {
		// Prefer "site:host path" style for better relevance
		path := strings.Trim(parsed.Path, "/")
		if path != "" {
			searchQuery = fmt.Sprintf("site:%s %s", parsed.Host, strings.ReplaceAll(path, "/", " "))
		} else {
			searchQuery = fmt.Sprintf("site:%s", parsed.Host)
		}
	}

	for _, provider := range providers {
		results, err := dispatchSearch(ctx, provider, sc, searchQuery, 3, "")
		if err != nil || len(results) == 0 {
			continue
		}
		// Found search results for the page — return them with a note
		var sb strings.Builder
		sb.WriteString(fmt.Sprintf("[Direct fetch failed: %v. Showing search results for %s]\n\n", fetchErr, rawURL))
		for i, r := range results {
			sb.WriteString(fmt.Sprintf("%d. %s\n   %s\n   %s\n\n", i+1, r.Title, r.URL, r.Description))
		}
		result := sb.String()
		if len(result) > maxChars {
			result = result[:maxChars] + "...[truncated]"
		}
		return result, nil
	}

	// All fallbacks exhausted — return the original fetch error
	return "", fmt.Errorf("web_fetch failed: %w", fetchErr)
}

// doHTTPFetch performs a direct HTTP GET and returns extracted text.
func (s *AgentToolRegistryService) doHTTPFetch(ctx context.Context, rawURL string, maxChars int) (string, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", rawURL, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("User-Agent", "IAC-Agent/1.0 (compatible)")
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,text/plain;q=0.8")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return "", fmt.Errorf("HTTP %d: %s", resp.StatusCode, http.StatusText(resp.StatusCode))
	}

	body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20)) // 1 MB cap
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	var text string
	ct := resp.Header.Get("Content-Type")
	if strings.Contains(ct, "text/html") {
		text = extractHTMLText(string(body))
	} else {
		text = string(body)
	}

	if len(text) > maxChars {
		text = text[:maxChars] + "...[truncated]"
	}
	return text, nil
}

// ─── web_search ───────────────────────────────────────────────────────────────

// searchResult is the unified result type returned by all providers.
type searchResult struct {
	Title       string `json:"title"`
	URL         string `json:"url"`
	Description string `json:"description"`
	Source      string `json:"source,omitempty"` // provider name
}

// getAvailableProviders returns the search providers that are actually configured
// (have API keys set), in preference order: tavily → brave → serpapi → google_cse → duckduckgo.
// DuckDuckGo is always appended last as a keyless fallback.
func getAvailableProviders(sc config.SearchConfig) []string {
	var providers []string
	// If an explicit primary is set and it's configured, put it first
	if sc.Provider != "" && sc.Provider != "duckduckgo" {
		if isProviderConfigured(sc, sc.Provider) {
			providers = append(providers, sc.Provider)
		}
	}
	// Then add remaining configured providers in preference order
	for _, p := range []string{"tavily", "brave", "serpapi", "google_cse"} {
		if p == sc.Provider {
			continue // already added above
		}
		if isProviderConfigured(sc, p) {
			providers = append(providers, p)
		}
	}
	// DuckDuckGo is always available as the last resort
	providers = append(providers, "duckduckgo")
	return providers
}

// isProviderConfigured returns true when a provider has the minimum required credentials set.
func isProviderConfigured(sc config.SearchConfig, provider string) bool {
	switch provider {
	case "brave":
		return sc.Brave.APIKey != ""
	case "tavily":
		return sc.Tavily.APIKey != ""
	case "serpapi":
		return sc.SerpAPI.APIKey != ""
	case "google_cse":
		return sc.GoogleCSE.APIKey != "" && sc.GoogleCSE.CX != ""
	case "duckduckgo":
		return true
	}
	return false
}

// webSearchHandler backs the generic "web_search" tool.
// It tries each available (configured) search provider in preference order until
// one succeeds and returns results. DuckDuckGo is always the final fallback.
func (s *AgentToolRegistryService) webSearchHandler(ctx context.Context, args map[string]interface{}) (string, error) {
	query, _ := args["query"].(string)
	if query == "" {
		return "", fmt.Errorf("query is required")
	}
	count := 5
	if v, ok := args["count"].(float64); ok && v > 0 {
		count = int(v)
		if count > 10 {
			count = 10
		}
	}
	country, _ := args["country"].(string)
	sc := config.GetSearchConfig()

	providers := getAvailableProviders(sc)
	var results []searchResult
	for _, provider := range providers {
		var err error
		results, err = dispatchSearch(ctx, provider, sc, query, count, country)
		if err != nil {
			s.iLog.Error(fmt.Sprintf("web_search: provider '%s' failed: %v", provider, err))
			continue
		}
		if len(results) > 0 {
			break
		}
		// Provider returned no results — try next (e.g. DDG returns empty for current events)
		s.iLog.Info(fmt.Sprintf("web_search: provider '%s' returned no results, trying next", provider))
	}

	if results == nil {
		results = []searchResult{}
	}
	b, _ := json.Marshal(results)
	return string(b), nil
}

// dispatchSearch routes to the selected provider implementation.
func dispatchSearch(ctx context.Context, provider string, sc config.SearchConfig, query string, count int, country string) ([]searchResult, error) {
	switch provider {
	case "brave":
		return searchBrave(ctx, sc.Brave.APIKey, query, count, country)
	case "tavily":
		return searchTavily(ctx, sc.Tavily.APIKey, sc.Tavily.SearchDepth, query, count)
	case "serpapi":
		engine := sc.SerpAPI.Engine
		if engine == "" {
			engine = "google"
		}
		return searchSerpAPI(ctx, sc.SerpAPI.APIKey, engine, query, count, country)
	case "google_cse":
		return searchGoogleCSE(ctx, sc.GoogleCSE.APIKey, sc.GoogleCSE.CX, query, count, country)
	case "duckduckgo", "":
		return searchDuckDuckGo(ctx, query)
	default:
		return nil, fmt.Errorf("unknown search provider: %s", provider)
	}
}

// ── Brave Search ──────────────────────────────────────────────────────────────
// Docs: https://api.search.brave.com/app/documentation/web-search/get-started
// Free tier: 2,000 queries/month. Endpoint returns web.results[].

func searchBrave(ctx context.Context, apiKey, query string, count int, country string) ([]searchResult, error) {
	if apiKey == "" {
		return nil, fmt.Errorf("Brave API key not configured (set brave.api_key in aiconfig.json or BRAVE_API_KEY env)")
	}

	params := url.Values{}
	params.Set("q", query)
	params.Set("count", fmt.Sprintf("%d", count))
	if country != "" {
		params.Set("country", strings.ToUpper(country))
	}

	req, err := http.NewRequestWithContext(ctx, "GET",
		"https://api.search.brave.com/res/v1/web/search?"+params.Encode(), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("X-Subscription-Token", apiKey)
	req.Header.Set("Accept", "application/json")

	resp, err := (&http.Client{Timeout: 15 * time.Second}).Do(req)
	if err != nil {
		return nil, fmt.Errorf("brave request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("brave returned HTTP %d", resp.StatusCode)
	}

	body, err := io.ReadAll(io.LimitReader(resp.Body, 512*1024))
	if err != nil {
		return nil, err
	}

	var raw map[string]interface{}
	if err := json.Unmarshal(body, &raw); err != nil {
		return nil, err
	}

	var results []searchResult
	if web, ok := raw["web"].(map[string]interface{}); ok {
		if items, ok := web["results"].([]interface{}); ok {
			for _, item := range items {
				if m, ok := item.(map[string]interface{}); ok {
					r := searchResult{Source: "brave"}
					if t, ok := m["title"].(string); ok {
						r.Title = t
					}
					if u, ok := m["url"].(string); ok {
						r.URL = u
					}
					if d, ok := m["description"].(string); ok {
						r.Description = d
					}
					results = append(results, r)
				}
			}
		}
	}
	return results, nil
}

// ── Tavily Search ─────────────────────────────────────────────────────────────
// Docs: https://docs.tavily.com/docs/tavily-api/rest_api
// Free tier: 1,000 queries/month. AI-optimised results, ideal for agents.
// POST https://api.tavily.com/search

func searchTavily(ctx context.Context, apiKey, searchDepth, query string, count int) ([]searchResult, error) {
	if apiKey == "" {
		return nil, fmt.Errorf("Tavily API key not configured (set tavily.api_key in aiconfig.json or TAVILY_API_KEY env)")
	}
	if searchDepth == "" {
		searchDepth = "basic"
	}

	payload := map[string]interface{}{
		"api_key":      apiKey,
		"query":        query,
		"search_depth": searchDepth,
		"max_results":  count,
		"include_answer":       false,
		"include_raw_content":  false,
		"include_images":       false,
	}
	b, _ := json.Marshal(payload)

	req, err := http.NewRequestWithContext(ctx, "POST", "https://api.tavily.com/search",
		strings.NewReader(string(b)))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := (&http.Client{Timeout: 20 * time.Second}).Do(req)
	if err != nil {
		return nil, fmt.Errorf("tavily request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		return nil, fmt.Errorf("tavily returned HTTP %d: %s", resp.StatusCode, string(body))
	}

	body, err := io.ReadAll(io.LimitReader(resp.Body, 512*1024))
	if err != nil {
		return nil, err
	}

	var raw struct {
		Results []struct {
			Title   string  `json:"title"`
			URL     string  `json:"url"`
			Content string  `json:"content"`
			Score   float64 `json:"score"`
		} `json:"results"`
	}
	if err := json.Unmarshal(body, &raw); err != nil {
		return nil, err
	}

	results := make([]searchResult, 0, len(raw.Results))
	for _, r := range raw.Results {
		results = append(results, searchResult{
			Title:       r.Title,
			URL:         r.URL,
			Description: r.Content,
			Source:      "tavily",
		})
	}
	return results, nil
}

// ── SerpAPI ───────────────────────────────────────────────────────────────────
// Docs: https://serpapi.com/search-api
// Free tier: 100 queries/month. Covers Google, Bing, Yahoo, DuckDuckGo, etc.
// GET https://serpapi.com/search.json?engine=google&q=...&api_key=...

func searchSerpAPI(ctx context.Context, apiKey, engine, query string, count int, country string) ([]searchResult, error) {
	if apiKey == "" {
		return nil, fmt.Errorf("SerpAPI key not configured (set serpapi.api_key in aiconfig.json or SERPAPI_API_KEY env)")
	}

	params := url.Values{}
	params.Set("api_key", apiKey)
	params.Set("engine", engine)
	params.Set("q", query)
	params.Set("num", fmt.Sprintf("%d", count))
	if country != "" {
		params.Set("gl", strings.ToLower(country)) // Google geo-location code
	}

	req, err := http.NewRequestWithContext(ctx, "GET",
		"https://serpapi.com/search.json?"+params.Encode(), nil)
	if err != nil {
		return nil, err
	}

	resp, err := (&http.Client{Timeout: 20 * time.Second}).Do(req)
	if err != nil {
		return nil, fmt.Errorf("serpapi request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("serpapi returned HTTP %d", resp.StatusCode)
	}

	body, err := io.ReadAll(io.LimitReader(resp.Body, 512*1024))
	if err != nil {
		return nil, err
	}

	var raw struct {
		OrganicResults []struct {
			Title   string `json:"title"`
			Link    string `json:"link"`
			Snippet string `json:"snippet"`
		} `json:"organic_results"`
	}
	if err := json.Unmarshal(body, &raw); err != nil {
		return nil, err
	}

	results := make([]searchResult, 0, len(raw.OrganicResults))
	for _, r := range raw.OrganicResults {
		results = append(results, searchResult{
			Title:       r.Title,
			URL:         r.Link,
			Description: r.Snippet,
			Source:      "serpapi/" + engine,
		})
	}
	return results, nil
}

// ── Google Custom Search ──────────────────────────────────────────────────────
// Docs: https://developers.google.com/custom-search/v1/reference/rest/v1/cse/list
// Free tier: 100 queries/day. Requires a Programmable Search Engine (cx) ID.
// GET https://www.googleapis.com/customsearch/v1?key=...&cx=...&q=...

func searchGoogleCSE(ctx context.Context, apiKey, cx, query string, count int, country string) ([]searchResult, error) {
	if apiKey == "" {
		return nil, fmt.Errorf("Google CSE API key not configured (set google_cse.api_key in aiconfig.json or GOOGLE_CSE_API_KEY env)")
	}
	if cx == "" {
		return nil, fmt.Errorf("Google CSE search engine ID (cx) not configured (set google_cse.cx in aiconfig.json or GOOGLE_CSE_CX env)")
	}

	params := url.Values{}
	params.Set("key", apiKey)
	params.Set("cx", cx)
	params.Set("q", query)
	params.Set("num", fmt.Sprintf("%d", minInt(count, 10))) // CSE max is 10
	if country != "" {
		params.Set("gl", strings.ToLower(country))
	}

	req, err := http.NewRequestWithContext(ctx, "GET",
		"https://www.googleapis.com/customsearch/v1?"+params.Encode(), nil)
	if err != nil {
		return nil, err
	}

	resp, err := (&http.Client{Timeout: 15 * time.Second}).Do(req)
	if err != nil {
		return nil, fmt.Errorf("google cse request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		return nil, fmt.Errorf("google cse returned HTTP %d: %s", resp.StatusCode, string(body))
	}

	body, err := io.ReadAll(io.LimitReader(resp.Body, 512*1024))
	if err != nil {
		return nil, err
	}

	var raw struct {
		Items []struct {
			Title   string `json:"title"`
			Link    string `json:"link"`
			Snippet string `json:"snippet"`
		} `json:"items"`
	}
	if err := json.Unmarshal(body, &raw); err != nil {
		return nil, err
	}

	results := make([]searchResult, 0, len(raw.Items))
	for _, r := range raw.Items {
		results = append(results, searchResult{
			Title:       r.Title,
			URL:         r.Link,
			Description: r.Snippet,
			Source:      "google_cse",
		})
	}
	return results, nil
}

// ── DuckDuckGo Instant Answers ────────────────────────────────────────────────
// Docs: https://api.duckduckgo.com/api
// Free, no key required. Returns Wikipedia abstracts and related topics.
// Not a full web index — best used as a free fallback.
// GET https://api.duckduckgo.com/?q=...&format=json&no_html=1&skip_disambig=1

func searchDuckDuckGo(ctx context.Context, query string) ([]searchResult, error) {
	params := url.Values{}
	params.Set("q", query)
	params.Set("format", "json")
	params.Set("no_html", "1")
	params.Set("skip_disambig", "1")

	req, err := http.NewRequestWithContext(ctx, "GET",
		"https://api.duckduckgo.com/?"+params.Encode(), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "IAC-Agent/1.0")

	resp, err := (&http.Client{Timeout: 10 * time.Second}).Do(req)
	if err != nil {
		return nil, fmt.Errorf("duckduckgo request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(io.LimitReader(resp.Body, 256*1024))
	if err != nil {
		return nil, err
	}

	var raw struct {
		AbstractText  string `json:"AbstractText"`
		AbstractURL   string `json:"AbstractURL"`
		AbstractTitle string `json:"AbstractTitle"`
		RelatedTopics []struct {
			Text      string `json:"Text"`
			FirstURL  string `json:"FirstURL"`
		} `json:"RelatedTopics"`
		Results []struct {
			Text     string `json:"Text"`
			FirstURL string `json:"FirstURL"`
		} `json:"Results"`
	}
	if err := json.Unmarshal(body, &raw); err != nil {
		return nil, err
	}

	var results []searchResult

	// Abstract (Wikipedia summary)
	if raw.AbstractText != "" {
		results = append(results, searchResult{
			Title:       raw.AbstractTitle,
			URL:         raw.AbstractURL,
			Description: raw.AbstractText,
			Source:      "duckduckgo",
		})
	}

	// Direct results
	for _, r := range raw.Results {
		if r.Text != "" && r.FirstURL != "" {
			results = append(results, searchResult{
				Title:       r.Text,
				URL:         r.FirstURL,
				Description: r.Text,
				Source:      "duckduckgo",
			})
		}
	}

	// Related topics (up to 4)
	added := 0
	for _, t := range raw.RelatedTopics {
		if added >= 4 {
			break
		}
		if t.Text != "" && t.FirstURL != "" {
			results = append(results, searchResult{
				Title:       t.Text,
				URL:         t.FirstURL,
				Description: t.Text,
				Source:      "duckduckgo",
			})
			added++
		}
	}

	return results, nil
}

// ─── http_request ─────────────────────────────────────────────────────────────

func (s *AgentToolRegistryService) httpRequestHandler(ctx context.Context, args map[string]interface{}) (string, error) {
	rawURL, _ := args["url"].(string)
	if rawURL == "" {
		return "", fmt.Errorf("url is required")
	}

	method := "GET"
	if m, ok := args["method"].(string); ok && m != "" {
		method = strings.ToUpper(m)
	}

	parsed, err := url.Parse(rawURL)
	if err != nil {
		return "", fmt.Errorf("invalid URL: %w", err)
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return "", fmt.Errorf("only http and https URLs are supported")
	}
	if isPrivateHost(parsed.Hostname()) {
		return "", fmt.Errorf("access to private or local addresses is not allowed")
	}

	var bodyReader io.Reader
	if bodyStr, ok := args["body"].(string); ok && bodyStr != "" {
		bodyReader = strings.NewReader(bodyStr)
	}

	req, err := http.NewRequestWithContext(ctx, method, rawURL, bodyReader)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("User-Agent", "IAC-Agent/1.0")

	if headersStr, ok := args["headers"].(string); ok && headersStr != "" {
		var headers map[string]string
		if err := json.Unmarshal([]byte(headersStr), &headers); err == nil {
			for k, v := range headers {
				req.Header.Set(k, v)
			}
		}
	}

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(io.LimitReader(resp.Body, 512*1024))
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	result := map[string]interface{}{
		"status": resp.StatusCode,
		"body":   string(body),
	}
	b, _ := json.Marshal(result)
	return string(b), nil
}

// ─── HTML utilities ───────────────────────────────────────────────────────────

// extractHTMLText parses HTML and extracts readable text, skipping scripts, styles, nav etc.
func extractHTMLText(htmlContent string) string {
	doc, err := xhtml.Parse(strings.NewReader(htmlContent))
	if err != nil {
		return stripHTMLTags(htmlContent)
	}
	var buf strings.Builder
	var traverse func(*xhtml.Node)
	traverse = func(n *xhtml.Node) {
		if n.Type == xhtml.ElementNode {
			switch n.Data {
			case "script", "style", "head", "nav", "footer", "aside", "noscript":
				return
			}
		}
		if n.Type == xhtml.TextNode {
			text := strings.TrimSpace(n.Data)
			if text != "" {
				buf.WriteString(text)
				buf.WriteString(" ")
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			traverse(c)
		}
	}
	traverse(doc)
	return strings.Join(strings.Fields(buf.String()), " ")
}

// stripHTMLTags is a fallback that removes HTML tags with a simple character scan.
func stripHTMLTags(s string) string {
	var b strings.Builder
	inTag := false
	for _, r := range s {
		switch {
		case r == '<':
			inTag = true
		case r == '>':
			inTag = false
		case !inTag:
			b.WriteRune(r)
		}
	}
	return strings.Join(strings.Fields(b.String()), " ")
}

// isPrivateHost returns true if the hostname looks like a private or local address.
// This is a best-effort SSRF guard on the raw hostname string before DNS resolution.
func isPrivateHost(host string) bool {
	host = strings.TrimPrefix(host, "[")
	host = strings.TrimSuffix(host, "]")
	lower := strings.ToLower(strings.TrimSpace(host))

	privatePatterns := []string{
		"localhost", "::1", "0.0.0.0",
		"127.", "10.", "192.168.",
		"172.16.", "172.17.", "172.18.", "172.19.",
		"172.20.", "172.21.", "172.22.", "172.23.",
		"172.24.", "172.25.", "172.26.", "172.27.",
		"172.28.", "172.29.", "172.30.", "172.31.",
		"169.254.",
	}
	for _, p := range privatePatterns {
		if lower == p || strings.HasPrefix(lower, p) {
			return true
		}
	}
	return false
}

// minInt returns the smaller of two ints.
func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}
