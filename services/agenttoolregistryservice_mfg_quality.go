package services

// Manufacturing Quality tools — AI-powered visual inspection and quality defect analysis.
//
// - inspect_image_quality : vision LLM identifies and annotates defects in a product image,
//                           optionally comparing against a reference/golden sample image.
//                           Returns structured defect list with bounding-box coordinates.
// - analyze_quality_defects: LLM-driven analysis of quality measurement data, SPC trends,
//                            and inspection history to identify defect root causes and
//                            corrective actions.
//
// Vision API note: uses OpenAI chat-completion with image_url content parts directly,
// since the shared LLMClient (llm.go) carries string-only content. Falls back to the
// global OpenAiKey / system aiconfig.json image_analysis use-case config.

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/mdaxf/iac/config"
	"github.com/mdaxf/iac/llm"
)

func (s *AgentToolRegistryService) registerMfgQualityTools() {
	s.registerImageQualityInspectionTool()
	s.registerQualityDefectAnalysisTool()
}

// ═══════════════════════════════════════════════════════════════════════════════
// VISUAL QUALITY INSPECTION
// ═══════════════════════════════════════════════════════════════════════════════

func (s *AgentToolRegistryService) registerImageQualityInspectionTool() {
	s.register(ToolDefinition{
		Type: "function",
		Function: ToolFunctionDef{
			Name: "inspect_image_quality",
			Description: "Performs AI-powered visual quality inspection on a product image. " +
				"Identifies and annotates manufacturing defects (scratches, cracks, voids, contamination, dimensional deviations, surface finish issues, etc.). " +
				"Optionally compares against a reference/golden-sample image to highlight deviations. " +
				"Returns a structured defect report with severity classification, location coordinates (as % of image), and overall pass/fail result. " +
				"Supports image URLs and base64-encoded images (JPEG/PNG/WEBP). " +
				"The coordinate data can be used to overlay visual annotations in the UI.",
			Parameters: ToolParameterSchema{
				Type: "object",
				Properties: map[string]ToolPropertySchema{
					"image_url": {
						Type:        "string",
						Description: "Public URL of the product image to inspect (use this OR image_base64).",
					},
					"image_base64": {
						Type:        "string",
						Description: "Base64-encoded product image. Include a data URI prefix: 'data:image/jpeg;base64,...' (use this OR image_url).",
					},
					"reference_image_url": {
						Type:        "string",
						Description: "Optional URL of a golden-sample/reference image. When provided, deviations from the reference are highlighted.",
					},
					"reference_image_base64": {
						Type:        "string",
						Description: "Optional base64-encoded reference image (data URI format).",
					},
					"product_type": {
						Type:        "string",
						Description: "Product or part type for context (e.g. 'PCB board', 'injection-molded housing', 'machined shaft', 'weld seam').",
					},
					"defect_types_json": {
						Type: "string",
						Description: `Optional JSON array of defect types to specifically look for. ` +
							`Example: ["scratch","crack","void","contamination","burr","delamination","oxidation"]. ` +
							`If omitted, the AI performs a general defect scan.`,
					},
					"inspection_standard": {
						Type:        "string",
						Description: "Optional description of the acceptance standard or AQL level (e.g. 'IPC Class 2', 'ISO 9001 visual criteria', 'no scratch > 2mm on functional surfaces').",
					},
					"inspection_level": {
						Type:        "string",
						Description: "Inspection thoroughness: 'standard' (default) | 'strict' (flag marginal defects) | 'critical' (flag any anomaly).",
					},
					"save_result": {
						Type:        "boolean",
						Description: "If true, saves the inspection result to the mfg_quality_inspections database table (default: false).",
					},
					"product_id": {
						Type:        "string",
						Description: "Optional product or lot ID to associate with the saved inspection record.",
					},
				},
			},
		},
	}, func(ctx context.Context, args map[string]interface{}) (string, error) {
		// Resolve image inputs
		imageURL := stringArg(args, "image_url")
		imageBase64 := stringArg(args, "image_base64")
		if imageURL == "" && imageBase64 == "" {
			return "", fmt.Errorf("either image_url or image_base64 is required")
		}

		refURL := stringArg(args, "reference_image_url")
		refBase64 := stringArg(args, "reference_image_base64")
		productType := stringArg(args, "product_type")
		inspectionStandard := stringArg(args, "inspection_standard")
		inspectionLevel := stringArg(args, "inspection_level")
		if inspectionLevel == "" {
			inspectionLevel = "standard"
		}

		var defectTypes []string
		if dtj := stringArg(args, "defect_types_json"); dtj != "" {
			json.Unmarshal([]byte(dtj), &defectTypes) //nolint:errcheck
		}

		hasReference := refURL != "" || refBase64 != ""

		// Build system prompt
		systemPrompt := buildVisionSystemPrompt(productType, defectTypes, inspectionStandard, inspectionLevel, hasReference)

		// Build vision messages
		var imageParts []visionImageInput
		if imageURL != "" {
			imageParts = append(imageParts, visionImageInput{URL: imageURL, Detail: "high"})
		} else {
			imageParts = append(imageParts, visionImageInput{URL: imageBase64, Detail: "high"})
		}
		if refURL != "" {
			imageParts = append(imageParts, visionImageInput{URL: refURL, Detail: "high", Label: "REFERENCE IMAGE"})
		} else if refBase64 != "" {
			imageParts = append(imageParts, visionImageInput{URL: refBase64, Detail: "high", Label: "REFERENCE IMAGE"})
		}

		// Call vision LLM
		resultJSON, err := s.callVisionLLM(ctx, systemPrompt, imageParts)
		if err != nil {
			return "", fmt.Errorf("vision inspection failed: %w", err)
		}

		// Parse result for DB save + enrichment
		var inspectionResult map[string]interface{}
		json.Unmarshal([]byte(resultJSON), &inspectionResult) //nolint:errcheck

		// Enrich with metadata
		enriched := map[string]interface{}{
			"inspection_id":       uuid.New().String(),
			"timestamp":           time.Now().UTC().Format(time.RFC3339),
			"product_type":        productType,
			"inspection_level":    inspectionLevel,
			"has_reference_image": hasReference,
			"result":              inspectionResult,
		}
		if productID := stringArg(args, "product_id"); productID != "" {
			enriched["product_id"] = productID
		}

		// Optionally save to DB
		if boolArg(args, "save_result") && s.sqlDB != nil {
			saveQualityInspection(ctx, s, enriched, inspectionResult)
		}

		out, _ := json.Marshal(enriched)
		return string(out), nil
	})
}

func buildVisionSystemPrompt(productType string, defectTypes []string, standard, level string, hasRef bool) string {
	var sb strings.Builder
	sb.WriteString("You are a precision manufacturing quality control expert with expertise in visual defect inspection. ")
	sb.WriteString("Analyze the provided product image(s) with meticulous attention to surface quality, dimensional accuracy, and manufacturing anomalies.\n\n")

	if hasRef {
		sb.WriteString("TWO IMAGES PROVIDED: The first is the INSPECTION IMAGE (product to evaluate). The second is the REFERENCE/GOLDEN SAMPLE image. Compare them carefully — highlight any differences, deviations, or anomalies present in the inspection image that are absent in the reference.\n\n")
	}

	if productType != "" {
		sb.WriteString(fmt.Sprintf("PRODUCT TYPE: %s\n", productType))
	}
	if standard != "" {
		sb.WriteString(fmt.Sprintf("INSPECTION STANDARD: %s\n", standard))
	}
	sb.WriteString(fmt.Sprintf("INSPECTION LEVEL: %s\n\n", level))

	if len(defectTypes) > 0 {
		sb.WriteString("DEFECT TYPES TO CHECK:\n")
		for _, d := range defectTypes {
			sb.WriteString(fmt.Sprintf("  - %s\n", d))
		}
		sb.WriteString("\n")
	} else {
		sb.WriteString("PERFORM A GENERAL SCAN for all types of manufacturing defects including but not limited to: scratches, cracks, voids, porosity, contamination, burrs, delamination, oxidation, dimensional deviations, surface finish issues, missing features, misalignment.\n\n")
	}

	sb.WriteString(`SEVERITY CLASSIFICATION:
  - critical: defect that makes the product unsafe or non-functional; must reject
  - major: defect that affects function or reliability; likely reject
  - minor: cosmetic defect within borderline tolerance; review required
  - observation: very slight anomaly; may be acceptable depending on standard

LOCATION FORMAT: Express bounding boxes as percentage of image dimensions.
  bbox_pct: {"x": left_pct, "y": top_pct, "width": width_pct, "height": height_pct}
  where 0,0 is the top-left corner and values are 0–100.

RESPOND ONLY with a JSON object in this exact format (no markdown, no extra text):
{
  "overall_result": "pass|fail|review",
  "confidence_overall": <0.0-1.0>,
  "defects_found": <integer count>,
  "defects": [
    {
      "id": <integer>,
      "type": "<defect type>",
      "severity": "critical|major|minor|observation",
      "confidence": <0.0-1.0>,
      "bbox_pct": {"x": <num>, "y": <num>, "width": <num>, "height": <num>},
      "location_description": "<human-readable location>",
      "description": "<detailed defect description>",
      "vs_reference": "<how it differs from reference, or 'N/A - no reference provided'>",
      "recommended_action": "<reject|rework|review|accept>"
    }
  ],
  "summary": "<concise overall assessment>",
  "recommendation": "<overall action: PASS / REJECT / REWORK / REVIEW>",
  "quality_notes": "<additional quality observations>",
  "inspection_completeness": "<areas that could not be assessed due to image angle/quality>"
}`)

	return sb.String()
}

// ─── vision LLM types ─────────────────────────────────────────────────────────

type visionImageInput struct {
	URL    string
	Detail string
	Label  string // optional prefix label in text content
}

type openAIVisionRequest struct {
	Model     string           `json:"model"`
	Messages  []visionMsgPart  `json:"messages"`
	MaxTokens int              `json:"max_tokens"`
}

type visionMsgPart struct {
	Role    string        `json:"role"`
	Content interface{}   `json:"content"` // string or []interface{}
}

// callVisionLLM sends images + system prompt to a vision-capable LLM and returns the response.
// Uses the image_analysis use case from aiconfig.json, or falls back to gpt-4o with the system OpenAI key.
func (s *AgentToolRegistryService) callVisionLLM(ctx context.Context, systemPrompt string, images []visionImageInput) (string, error) {
	// Resolve API key: tool registry → global config
	apiKey := s.openAIKey
	if apiKey == "" {
		apiKey = config.OpenAiKey
	}
	if apiKey == "" {
		return "", fmt.Errorf("OpenAI API key not configured: set openai_key in configuration.json")
	}

	// Use gpt-4o as it is vision-capable; check aiconfig for override
	model := "gpt-4o"
	if cfg := config.GetAIConfig(); cfg != nil {
		if uc, ok := cfg.UseCases["image_analysis"]; ok {
			if uc.ModelOverride != "" {
				model = uc.ModelOverride
			}
		}
	}

	// Build user content: text instruction + images
	var userContent []interface{}

	// Prefix label for multi-image scenarios
	for _, img := range images {
		if img.Label != "" {
			userContent = append(userContent, map[string]interface{}{
				"type": "text",
				"text": fmt.Sprintf("[%s]", img.Label),
			})
		}
		userContent = append(userContent, map[string]interface{}{
			"type": "image_url",
			"image_url": map[string]string{
				"url":    img.URL,
				"detail": img.Detail,
			},
		})
	}
	userContent = append(userContent, map[string]interface{}{
		"type": "text",
		"text": "Please inspect the image(s) above and respond with the JSON report as instructed.",
	})

	reqBody := map[string]interface{}{
		"model": model,
		"messages": []interface{}{
			map[string]interface{}{
				"role":    "system",
				"content": systemPrompt,
			},
			map[string]interface{}{
				"role":    "user",
				"content": userContent,
			},
		},
		"max_tokens": 2000,
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal vision request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost,
		"https://api.openai.com/v1/chat/completions", bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("failed to build vision request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)

	client := &http.Client{Timeout: 60 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("vision API call failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		preview := string(respBody)
		if len(preview) > 300 {
			preview = preview[:300]
		}
		return "", fmt.Errorf("vision API HTTP %d: %s", resp.StatusCode, preview)
	}

	// Parse OpenAI response envelope
	var oaiResp struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
		Error *struct {
			Message string `json:"message"`
		} `json:"error"`
	}
	if err := json.Unmarshal(respBody, &oaiResp); err != nil {
		return "", fmt.Errorf("failed to parse vision response: %w", err)
	}
	if oaiResp.Error != nil {
		return "", fmt.Errorf("vision API error: %s", oaiResp.Error.Message)
	}
	if len(oaiResp.Choices) == 0 {
		return "", fmt.Errorf("vision API returned no choices")
	}

	content := strings.TrimSpace(oaiResp.Choices[0].Message.Content)
	// Strip markdown code fences if the LLM wrapped the JSON
	content = strings.TrimPrefix(content, "```json")
	content = strings.TrimPrefix(content, "```")
	content = strings.TrimSuffix(content, "```")
	return strings.TrimSpace(content), nil
}

// saveQualityInspection persists the inspection result to mfg_quality_inspections.
func saveQualityInspection(ctx context.Context, s *AgentToolRegistryService, enriched, result map[string]interface{}) {
	inspID, _ := enriched["inspection_id"].(string)
	productID, _ := enriched["product_id"].(string)
	productType, _ := enriched["product_type"].(string)

	overallResult := ""
	defectsFound := 0
	if result != nil {
		overallResult, _ = result["overall_result"].(string)
		if df, ok := result["defects_found"].(float64); ok {
			defectsFound = int(df)
		}
	}
	resultJSON, _ := json.Marshal(result)

	s.sqlDB.ExecContext(ctx, //nolint:errcheck
		`INSERT INTO mfg_quality_inspections
		 (id, product_id, product_type, overall_result, defects_found, result_json, createdon)
		 VALUES ($1,$2,$3,$4,$5,$6,$7)
		 ON CONFLICT DO NOTHING`,
		inspID, productID, productType, overallResult, defectsFound, string(resultJSON), time.Now().UTC())
}

// ═══════════════════════════════════════════════════════════════════════════════
// QUALITY DEFECT ANALYSIS
// ═══════════════════════════════════════════════════════════════════════════════

func (s *AgentToolRegistryService) registerQualityDefectAnalysisTool() {
	s.register(ToolDefinition{
		Type: "function",
		Function: ToolFunctionDef{
			Name: "analyze_quality_defects",
			Description: "AI-powered analysis of quality measurement data, SPC trends, and inspection history " +
				"to identify defect patterns, determine root causes, and recommend corrective actions. " +
				"Accepts measurement data as JSON — the agent can first query the database to gather " +
				"quality records, then pass the results to this tool for in-depth analysis. " +
				"Returns structured findings with Pareto-ranked defects, root cause hypotheses, and prioritized action plan.",
			Parameters: ToolParameterSchema{
				Type: "object",
				Properties: map[string]ToolPropertySchema{
					"product_id": {
						Type:        "string",
						Description: "Product, part, or SKU identifier (used to query inspection history if available).",
					},
					"batch_id": {
						Type:        "string",
						Description: "Batch, lot, or work order ID for traceability.",
					},
					"measurement_data_json": {
						Type: "string",
						Description: `JSON array of measurement records. ` +
							`Example: [{"parameter":"diameter_mm","measured":25.12,"nominal":25.00,"tolerance_plus":0.05,"tolerance_minus":0.05,"timestamp":"2026-03-02T08:00:00Z"}]. ` +
							`Include as many parameters as available: dimensions, hardness, surface roughness, visual scores, etc.`,
					},
					"specification_json": {
						Type: "string",
						Description: `JSON object defining product quality specifications and critical-to-quality (CTQ) parameters. ` +
							`Example: {"max_defect_rate_pct":0.5,"critical_parameters":["diameter_mm","surface_Ra"],"aql_level":"2.5"}`,
					},
					"inspection_history_json": {
						Type: "string",
						Description: `JSON array of recent inspection records (from mfg_quality_inspections or your own table). ` +
							`If omitted, the tool will attempt to query mfg_quality_inspections for product_id history.`,
					},
					"spc_data_json": {
						Type: "string",
						Description: `Statistical Process Control data: control chart values, Cp/Cpk indices, run chart data. ` +
							`Example: [{"parameter":"diameter_mm","mean":25.01,"std_dev":0.02,"cp":1.1,"cpk":0.95}]`,
					},
					"defect_log_json": {
						Type: "string",
						Description: `Historical defect log data. Example: [{"date":"2026-03-01","defect_type":"scratch","count":5,"affected_unit":"LOT-A123"}]`,
					},
					"process_context": {
						Type:        "string",
						Description: "Optional context about the manufacturing process, recent changes, or suspected causes.",
					},
					"analysis_depth": {
						Type:        "string",
						Description: "Analysis depth: 'summary' (quick overview) | 'standard' (default) | 'deep' (full 8D / DMAIC framework)",
					},
				},
			},
		},
	}, func(ctx context.Context, args map[string]interface{}) (string, error) {
		productID := stringArg(args, "product_id")
		batchID := stringArg(args, "batch_id")
		analysisDepth := stringArg(args, "analysis_depth")
		if analysisDepth == "" {
			analysisDepth = "standard"
		}

		// Fetch inspection history from DB if not provided
		historyJSON := stringArg(args, "inspection_history_json")
		if historyJSON == "" && productID != "" && s.sqlDB != nil {
			rows, err := s.sqlDB.QueryContext(ctx,
				`SELECT id, overall_result, defects_found, result_json, createdon
				 FROM mfg_quality_inspections WHERE product_id = $1
				 ORDER BY createdon DESC LIMIT 20`, productID)
			if err == nil {
				defer rows.Close()
				if h, err := s.rowsToJSON(rows); err == nil {
					historyJSON = h
				}
			}
		}

		// Assemble analysis prompt
		prompt := buildQualityAnalysisPrompt(productID, batchID, args, historyJSON, analysisDepth)

		// Call text LLM
		messages := []map[string]interface{}{
			{"role": "system", "content": qualityAnalysisSystemPrompt(analysisDepth)},
			{"role": "user", "content": prompt},
		}
		apiKey, model := s.openAIKey, "gpt-4o"
		resultText, err := llm.CallLLM(ctx, "quality_inspection", apiKey, model, messages, 0.2)
		if err != nil {
			return "", fmt.Errorf("quality analysis LLM call failed: %w", err)
		}

		resultText = cleanLLMJSON(resultText)

		// Enrich with metadata
		var analysisResult map[string]interface{}
		json.Unmarshal([]byte(resultText), &analysisResult) //nolint:errcheck

		out, _ := json.Marshal(map[string]interface{}{
			"analysis_id":    uuid.New().String(),
			"timestamp":      time.Now().UTC().Format(time.RFC3339),
			"product_id":     productID,
			"batch_id":       batchID,
			"analysis_depth": analysisDepth,
			"analysis":       analysisResult,
		})
		return string(out), nil
	})
}

func qualityAnalysisSystemPrompt(depth string) string {
	base := "You are a manufacturing quality engineer with expertise in Statistical Process Control (SPC), Six Sigma, and root cause analysis (8D, Ishikawa, FMEA). " +
		"Analyze quality data to identify defect patterns, determine root causes, and provide actionable corrective/preventive actions.\n" +
		"Always respond with valid JSON only — no markdown, no extra text.\n\n"
	switch depth {
	case "deep":
		base += "Perform a full DMAIC / 8D analysis: Define-Measure-Analyze-Improve-Control structure with detailed statistical reasoning."
	case "summary":
		base += "Provide a concise summary of key findings and top 3 immediate actions."
	default:
		base += "Provide standard quality analysis with ranked defects, Pareto insight, root causes, and prioritized action plan."
	}
	return base
}

func buildQualityAnalysisPrompt(productID, batchID string, args map[string]interface{}, historyJSON, depth string) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("## Quality Defect Analysis Request\nProduct ID: %s | Batch ID: %s | Depth: %s\n\n", productID, batchID, depth))

	appendSection := func(title, key string) {
		if v := stringArg(args, key); v != "" {
			sb.WriteString(fmt.Sprintf("### %s\n%s\n\n", title, v))
		}
	}
	appendSection("Measurement Data", "measurement_data_json")
	appendSection("Product Specifications", "specification_json")
	appendSection("SPC Data (Cp/Cpk/Control Charts)", "spc_data_json")
	appendSection("Defect Log History", "defect_log_json")
	appendSection("Process Context", "process_context")

	if historyJSON != "" && historyJSON != `{"rows":[]}` {
		sb.WriteString(fmt.Sprintf("### Past Inspection Records\n%s\n\n", historyJSON))
	}

	sb.WriteString(`Respond with this JSON structure:
{
  "defect_summary": {
    "total_defects": <int>,
    "defect_rate_pct": <float>,
    "critical_count": <int>,
    "major_count": <int>,
    "minor_count": <int>
  },
  "pareto_analysis": [
    {"rank": 1, "defect_type": "<type>", "count": <int>, "pct_of_total": <float>, "cumulative_pct": <float>}
  ],
  "out_of_spec_parameters": [
    {"parameter": "<name>", "measured": <value>, "nominal": <value>, "deviation_pct": <float>, "severity": "critical|major|minor"}
  ],
  "root_causes": [
    {"hypothesis": "<cause>", "category": "Machine|Method|Material|Man|Measurement|Environment", "evidence": "<supporting data>", "probability": "high|medium|low"}
  ],
  "spc_findings": {
    "process_capable": <bool>,
    "cp": <float or null>,
    "cpk": <float or null>,
    "control_chart_signals": ["<signal description>"]
  },
  "corrective_actions": [
    {"priority": 1, "action": "<description>", "owner": "<role>", "timeline": "<immediate|1-week|1-month>", "expected_impact": "<outcome>"}
  ],
  "preventive_actions": ["<action>"],
  "overall_quality_status": "acceptable|marginal|unacceptable",
  "summary": "<executive summary>"
}`)
	return sb.String()
}

// cleanLLMJSON strips markdown code fences from LLM output.
func cleanLLMJSON(s string) string {
	s = strings.TrimSpace(s)
	s = strings.TrimPrefix(s, "```json")
	s = strings.TrimPrefix(s, "```")
	s = strings.TrimSuffix(s, "```")
	return strings.TrimSpace(s)
}
