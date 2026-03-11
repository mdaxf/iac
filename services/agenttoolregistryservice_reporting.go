package services

// Reporting tools — allow agents to generate rich HTML reports with charts and tables.
//
// - generate_html_report: produces a self-contained HTML file (no server required)
//   using Chart.js (CDN) for bar/pie/line/doughnut charts and a sortable data table.
//   The returned HTML string can be sent via send_email (content_type: html),
//   saved to disk, or returned to the user directly.

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

func (s *AgentToolRegistryService) registerReportingTools() {
	s.registerHTMLReportTool()
	s.registerUIRenderTool()
	s.registerUISaveViewTool()
	s.registerUIComposePageTool()
}

func (s *AgentToolRegistryService) registerHTMLReportTool() {
	s.register(ToolDefinition{
		Type: "function",
		Function: ToolFunctionDef{
			Name: "generate_html_report",
			Description: "Generates a self-contained HTML report with optional Chart.js charts and a data table. " +
				"Returns the full HTML as a string that can be sent via send_email (content_type: html), " +
				"saved as a file, or shown directly to the user. " +
				"Charts support bar, pie, line, and doughnut types. " +
				"The table is auto-generated from the rows in table_data_json.",
			Parameters: ToolParameterSchema{
				Type: "object",
				Properties: map[string]ToolPropertySchema{
					"title": {
						Type:        "string",
						Description: "Report title shown in the header (required).",
					},
					"subtitle": {
						Type:        "string",
						Description: "Optional subtitle or description shown below the title.",
					},
					"charts_json": {
						Type: "string",
						Description: `JSON array of chart configurations. Each item: ` +
							`{"type":"bar|pie|line|doughnut","title":"Chart Title",` +
							`"labels":["A","B","C"],` +
							`"datasets":[{"label":"Series 1","data":[10,20,30],"backgroundColor":"#667eea"}]}. ` +
							`backgroundColor can be a single colour string or an array of colour strings (required for pie/doughnut).`,
					},
					"table_data_json": {
						Type: "string",
						Description: `JSON array of row objects for the data table. ` +
							`Example: [{"Name":"Alice","Score":95},{"Name":"Bob","Score":82}]. ` +
							`Column headers are derived from the object keys of the first row.`,
					},
					"table_title": {
						Type:        "string",
						Description: "Optional heading for the data table section (default: 'Data').",
					},
				},
				Required: []string{"title"},
			},
		},
	}, func(ctx context.Context, args map[string]interface{}) (string, error) {
		title := stringArg(args, "title")
		if title == "" {
			return "", fmt.Errorf("title is required")
		}
		subtitle := stringArg(args, "subtitle")
		tableTitle := stringArg(args, "table_title")
		if tableTitle == "" {
			tableTitle = "Data"
		}

		// Parse charts
		var charts []chartConfig
		if cj := stringArg(args, "charts_json"); cj != "" {
			if err := json.Unmarshal([]byte(cj), &charts); err != nil {
				return "", fmt.Errorf("invalid charts_json: %w", err)
			}
		}

		// Parse table data
		var tableRows []map[string]interface{}
		if tj := stringArg(args, "table_data_json"); tj != "" {
			if err := json.Unmarshal([]byte(tj), &tableRows); err != nil {
				return "", fmt.Errorf("invalid table_data_json: %w", err)
			}
		}

		html := buildHTMLReport(title, subtitle, tableTitle, charts, tableRows)
		out, _ := json.Marshal(map[string]string{"html": html})
		return string(out), nil
	})
}

// ─── ui_save_view tool ────────────────────────────────────────────────────────

func (s *AgentToolRegistryService) registerUISaveViewTool() {
	s.register(ToolDefinition{
		Type: "function",
		Function: ToolFunctionDef{
			Name: "ui_save_view",
			Description: `Saves a ui_render spec as a reusable UI_View template in the system.
Use this after rendering a ui_render panel to persist it so it can be embedded in pages or reused.
Inputs and outputs are automatically extracted from the spec:
  - form fields become inputs (id, datatype, description)
  - kpi_row items, table columns, chart datasets become outputs
Returns the view UUID, name, and the extracted inputs/outputs for reference.`,
			Parameters: ToolParameterSchema{
				Type: "object",
				Properties: map[string]ToolPropertySchema{
					"spec_json":   {Type: "string", Description: "The ui_render spec JSON (same format as ui_render). Accepts string or object."},
					"name":        {Type: "string", Description: "View name identifier (required). Use kebab-case, e.g. sales-dashboard-q1"},
					"title":       {Type: "string", Description: "Human-readable display title"},
					"description": {Type: "string", Description: "Brief description of what this view shows"},
					"version":     {Type: "string", Description: "Version string, e.g. '1.0' (default: '1.0')"},
				},
				Required: []string{"spec_json", "name"},
			},
		},
	}, func(ctx context.Context, args map[string]interface{}) (string, error) {
		if s.docDB == nil {
			return docNotAvailable()
		}
		spec, err := parseUISpec(args)
		if err != nil {
			return "", err
		}
		name := stringArg(args, "name")
		if name == "" {
			return "", fmt.Errorf("name is required")
		}
		version := stringArg(args, "version")
		if version == "" {
			version = "1.0"
		}

		specBytes, _ := json.Marshal(spec)
		inputs, outputs := extractViewIO(spec)

		doc := map[string]interface{}{
			"name":        name,
			"title":       stringArg(args, "title"),
			"description": stringArg(args, "description"),
			"type":        "custom",
			"subtype":     "ui_render",
			"version":     version,
			"uispec":      string(specBytes),
			"inputs":      inputs,
			"outputs":     outputs,
			"status":      0, // Design
		}

		id, err := s.entityInsert(ctx, collUIView, doc)
		if err != nil {
			return "", fmt.Errorf("failed to save view: %w", err)
		}

		result := map[string]interface{}{
			"view_id":   id,
			"view_name": name,
			"version":   version,
			"inputs":    inputs,
			"outputs":   outputs,
			"status":    "created",
		}
		out, _ := json.Marshal(result)
		return string(out), nil
	})
}

// ─── ui_compose_page tool ─────────────────────────────────────────────────────

func (s *AgentToolRegistryService) registerUIComposePageTool() {
	s.register(ToolDefinition{
		Type: "function",
		Function: ToolFunctionDef{
			Name: "ui_compose_page",
			Description: `Composes one or more saved UI_Views into a UI_Page.
Each view becomes a panel in the page layout. Use ui_save_view first to save views from ui_render,
then call this tool with their names to build a complete reusable page.
Use list_views to discover existing view names.
Returns the page UUID and name.`,
			Parameters: ToolParameterSchema{
				Type: "object",
				Properties: map[string]ToolPropertySchema{
					"name":        {Type: "string", Description: "Page name identifier (required). e.g. sales-dashboard-page"},
					"title":       {Type: "string", Description: "Page display title"},
					"description": {Type: "string", Description: "Page description"},
					"version":     {Type: "string", Description: "Version string (default: '1.0')"},
					"view_names":  {Type: "string", Description: `JSON array of view names to include as panels, in order. Example: ["kpi-overview","monthly-chart","product-table"]`},
					"orientation": {Type: "string", Description: "Layout direction: 'vertical' (default) | 'horizontal'"},
				},
				Required: []string{"name", "view_names"},
			},
		},
	}, func(ctx context.Context, args map[string]interface{}) (string, error) {
		if s.docDB == nil {
			return docNotAvailable()
		}
		name := stringArg(args, "name")
		if name == "" {
			return "", fmt.Errorf("name is required")
		}

		// view_names: accept JSON array string or raw []interface{}
		var viewNames []string
		switch v := args["view_names"].(type) {
		case string:
			if err := json.Unmarshal([]byte(v), &viewNames); err != nil {
				return "", fmt.Errorf("invalid view_names JSON: %w", err)
			}
		case []interface{}:
			for _, n := range v {
				if s, ok := n.(string); ok {
					viewNames = append(viewNames, s)
				}
			}
		default:
			return "", fmt.Errorf("view_names is required")
		}
		if len(viewNames) == 0 {
			return "", fmt.Errorf("view_names must contain at least one view name")
		}

		version := stringArg(args, "version")
		if version == "" {
			version = "1.0"
		}
		orientation := 1 // vertical default
		if stringArg(args, "orientation") == "horizontal" {
			orientation = 2
		}

		panels := make([]map[string]interface{}, 0, len(viewNames))
		for i, vn := range viewNames {
			panels = append(panels, map[string]interface{}{
				"name":  fmt.Sprintf("panel%d", i+1),
				"title": vn,
				"view": map[string]interface{}{
					"name": vn,
					"type": "document",
				},
			})
		}

		doc := map[string]interface{}{
			"name":        name,
			"title":       stringArg(args, "title"),
			"description": stringArg(args, "description"),
			"version":     version,
			"orientation": orientation,
			"isdefault":   false,
			"panels":      panels,
			"status":      0,
		}

		id, err := s.entityInsert(ctx, collUIPage, doc)
		if err != nil {
			return "", fmt.Errorf("failed to create page: %w", err)
		}

		result := map[string]interface{}{
			"page_id":     id,
			"page_name":   name,
			"version":     version,
			"panel_count": len(panels),
			"view_names":  viewNames,
			"status":      "created",
		}
		out, _ := json.Marshal(result)
		return string(out), nil
	})
}

// ─── internal types ──────────────────────────────────────────────────────────

type chartConfig struct {
	Type     string         `json:"type"`
	Title    string         `json:"title"`
	Labels   []string       `json:"labels"`
	Datasets []chartDataset `json:"datasets"`
}

type chartDataset struct {
	Label           string      `json:"label"`
	Data            []float64   `json:"data"`
	BackgroundColor interface{} `json:"backgroundColor"` // string or []string
}

// ─── HTML builder ─────────────────────────────────────────────────────────────

func buildHTMLReport(title, subtitle, tableTitle string, charts []chartConfig, rows []map[string]interface{}) string {
	subtitleHTML := ""
	if subtitle != "" {
		subtitleHTML = fmt.Sprintf(`<p class="report-subtitle">%s</p>`, htmlEscape(subtitle))
	}

	chartsHTML, chartScripts := buildChartsHTML(charts)
	tableHTML := buildTableHTML(tableTitle, rows)

	generatedAt := time.Now().UTC().Format("2006-01-02 15:04:05 UTC")

	return fmt.Sprintf(`<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width,initial-scale=1.0">
  <title>%s</title>
  <script src="https://cdn.jsdelivr.net/npm/chart.js@4/dist/chart.umd.min.js"></script>
  <style>
    *, *::before, *::after { box-sizing: border-box; }
    body { font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
           margin: 0; padding: 0; background: #f5f7fa; color: #1a1a1a; }
    .report-wrapper { max-width: 1200px; margin: 0 auto; padding: 2rem; }
    .report-header { background: linear-gradient(135deg, #667eea 0%%, #764ba2 100%%);
                     color: white; padding: 2rem 2.5rem; border-radius: 12px;
                     margin-bottom: 2rem; }
    .report-header h1 { margin: 0 0 0.4rem 0; font-size: 1.9rem; font-weight: 700; }
    .report-subtitle { margin: 0; opacity: 0.85; font-size: 1rem; }
    .charts-grid { display: grid;
                   grid-template-columns: repeat(auto-fit, minmax(440px, 1fr));
                   gap: 1.5rem; margin-bottom: 2rem; }
    .chart-card { background: white; border-radius: 12px; padding: 1.5rem;
                  box-shadow: 0 2px 8px rgba(0,0,0,0.08); }
    .chart-card h3 { margin: 0 0 1rem 0; font-size: 1rem; font-weight: 600;
                     color: #374151; }
    .chart-canvas-wrapper { position: relative; height: 280px; }
    .table-section { background: white; border-radius: 12px; padding: 1.5rem;
                     box-shadow: 0 2px 8px rgba(0,0,0,0.08);
                     margin-bottom: 2rem; overflow-x: auto; }
    .table-section h2 { margin: 0 0 1rem 0; font-size: 1.15rem; font-weight: 600;
                        color: #374151; }
    table { width: 100%%; border-collapse: collapse; font-size: 0.875rem; }
    thead th { text-align: left; padding: 0.7rem 1rem; background: #f9fafb;
               font-weight: 600; color: #6b7280; text-transform: uppercase;
               letter-spacing: 0.04em; font-size: 0.75rem;
               border-bottom: 2px solid #e5e7eb; white-space: nowrap; }
    tbody td { padding: 0.7rem 1rem; border-bottom: 1px solid #f3f4f6;
               color: #374151; }
    tbody tr:last-child td { border-bottom: none; }
    tbody tr:hover { background: #f9fafb; }
    .report-footer { text-align: center; color: #9ca3af; font-size: 0.8rem;
                     margin-top: 1rem; padding-bottom: 1rem; }
    @media (max-width: 640px) {
      .report-wrapper { padding: 1rem; }
      .charts-grid { grid-template-columns: 1fr; }
    }
  </style>
</head>
<body>
<div class="report-wrapper">
  <div class="report-header">
    <h1>%s</h1>
    %s
  </div>
  %s
  %s
  <div class="report-footer">Generated by IAC Agent &middot; %s</div>
</div>
<script>
%s
</script>
</body>
</html>`,
		htmlEscape(title),
		htmlEscape(title),
		subtitleHTML,
		chartsHTML,
		tableHTML,
		generatedAt,
		chartScripts,
	)
}

// defaultChartColors is a palette applied when the dataset has no explicit backgroundColor.
var defaultChartColors = []string{
	"#667eea", "#764ba2", "#f093fb", "#4facfe",
	"#43e97b", "#fa709a", "#fee140", "#30cfd0",
	"#a18cd1", "#fda085", "#84fab0", "#8fd3f4",
}

func buildChartsHTML(charts []chartConfig) (html string, scripts string) {
	if len(charts) == 0 {
		return "", ""
	}
	var htmlBuf, scriptBuf strings.Builder
	htmlBuf.WriteString(`<div class="charts-grid">`)
	for i, c := range charts {
		canvasID := fmt.Sprintf("iacChart%d", i)
		chartType := c.Type
		if chartType == "" {
			chartType = "bar"
		}

		htmlBuf.WriteString(fmt.Sprintf(`
  <div class="chart-card">
    <h3>%s</h3>
    <div class="chart-canvas-wrapper"><canvas id="%s"></canvas></div>
  </div>`, htmlEscape(c.Title), canvasID))

		// Build datasets JSON
		var dsJSON []string
		for di, ds := range c.Datasets {
			dataBytes, _ := json.Marshal(ds.Data)

			var bgJSON string
			if ds.BackgroundColor != nil {
				b, _ := json.Marshal(ds.BackgroundColor)
				bgJSON = string(b)
			} else {
				// Assign palette colours: single colour for bar/line, full palette for pie/doughnut
				if chartType == "pie" || chartType == "doughnut" {
					palette := make([]string, len(ds.Data))
					for pi := range ds.Data {
						palette[pi] = defaultChartColors[pi%len(defaultChartColors)]
					}
					b, _ := json.Marshal(palette)
					bgJSON = string(b)
				} else {
					colour := defaultChartColors[di%len(defaultChartColors)]
					bgJSON = fmt.Sprintf("%q", colour)
				}
			}

			labelBytes, _ := json.Marshal(ds.Label)
			dsJSON = append(dsJSON, fmt.Sprintf(
				`{"label":%s,"data":%s,"backgroundColor":%s,"borderColor":%s,"borderWidth":1}`,
				labelBytes, dataBytes, bgJSON, bgJSON,
			))
		}

		labelsBytes, _ := json.Marshal(c.Labels)
		indexAxis := ""
		if chartType == "horizontalBar" {
			chartType = "bar"
			indexAxis = `,"indexAxis":"y"`
		}

		scriptBuf.WriteString(fmt.Sprintf(`
(function(){
  var ctx = document.getElementById(%q).getContext('2d');
  new Chart(ctx, {
    type: %q,
    data: { labels: %s, datasets: [%s] },
    options: { responsive: true, maintainAspectRatio: false%s,
               plugins: { legend: { position: 'bottom' } } }
  });
})();`, canvasID, chartType, labelsBytes, strings.Join(dsJSON, ","), indexAxis))
	}
	htmlBuf.WriteString(`</div>`)
	return htmlBuf.String(), scriptBuf.String()
}

func buildTableHTML(tableTitle string, rows []map[string]interface{}) string {
	if len(rows) == 0 {
		return ""
	}

	// Collect column headers from first row (preserve insertion order)
	cols := make([]string, 0)
	seen := make(map[string]bool)
	for k := range rows[0] {
		if !seen[k] {
			cols = append(cols, k)
			seen[k] = true
		}
	}
	// Also pick up any columns from subsequent rows to handle sparse maps
	for _, row := range rows[1:] {
		for k := range row {
			if !seen[k] {
				cols = append(cols, k)
				seen[k] = true
			}
		}
	}

	var buf strings.Builder
	buf.WriteString(fmt.Sprintf(`<div class="table-section"><h2>%s</h2><table>`, htmlEscape(tableTitle)))
	buf.WriteString("<thead><tr>")
	for _, col := range cols {
		buf.WriteString(fmt.Sprintf("<th>%s</th>", htmlEscape(col)))
	}
	buf.WriteString("</tr></thead><tbody>")
	for _, row := range rows {
		buf.WriteString("<tr>")
		for _, col := range cols {
			val := ""
			if v, ok := row[col]; ok && v != nil {
				switch t := v.(type) {
				case string:
					val = t
				default:
					b, _ := json.Marshal(v)
					val = string(b)
				}
			}
			buf.WriteString(fmt.Sprintf("<td>%s</td>", htmlEscape(val)))
		}
		buf.WriteString("</tr>")
	}
	buf.WriteString("</tbody></table></div>")
	return buf.String()
}

// ─── shared spec helpers ──────────────────────────────────────────────────────

// parseUISpec parses spec_json from tool args — accepts both a JSON string and
// a raw object (LLMs often pass the spec as a nested object, not a string).
func parseUISpec(args map[string]interface{}) (map[string]interface{}, error) {
	switch v := args["spec_json"].(type) {
	case string:
		if v == "" {
			return nil, fmt.Errorf("spec_json is required")
		}
		var spec map[string]interface{}
		if err := json.Unmarshal([]byte(v), &spec); err != nil {
			return nil, fmt.Errorf("invalid spec_json: %w", err)
		}
		return spec, nil
	case map[string]interface{}:
		return v, nil
	default:
		return nil, fmt.Errorf("spec_json is required")
	}
}

// extractViewIO walks a ui_render spec and derives input/output parameter maps
// suitable for storing on a UI_View document.
//
// Inputs come from "form" section fields (each field → one input).
// Outputs come from "kpi_row" items, "table" columns, and "chart" datasets.
func extractViewIO(spec map[string]interface{}) (inputs, outputs map[string]interface{}) {
	inputs = map[string]interface{}{}
	outputs = map[string]interface{}{}

	sections, _ := spec["sections"].([]interface{})
	for _, raw := range sections {
		sec, ok := raw.(map[string]interface{})
		if !ok {
			continue
		}
		switch sec["type"] {
		case "form":
			fields, _ := sec["fields"].([]interface{})
			for _, rf := range fields {
				f, ok := rf.(map[string]interface{})
				if !ok {
					continue
				}
				id, _ := f["id"].(string)
				if id == "" {
					continue
				}
				label, _ := f["label"].(string)
				ftype, _ := f["type"].(string)
				datatype := "string"
				if ftype == "number" {
					datatype = "number"
				} else if ftype == "date" {
					datatype = "datetime"
				}
				inputs[id] = map[string]interface{}{
					"name":        id,
					"datatype":    datatype,
					"description": label,
				}
			}
		case "kpi_row":
			items, _ := sec["items"].([]interface{})
			for _, ri := range items {
				it, ok := ri.(map[string]interface{})
				if !ok {
					continue
				}
				label, _ := it["label"].(string)
				if label == "" {
					continue
				}
				id := toSnakeCase(label)
				outputs[id] = map[string]interface{}{
					"name":        id,
					"datatype":    "string",
					"description": label + " metric value",
				}
			}
		case "table":
			cols, _ := sec["columns"].([]interface{})
			for _, rc := range cols {
				col, _ := rc.(string)
				if col == "" {
					continue
				}
				id := toSnakeCase(col)
				outputs[id] = map[string]interface{}{
					"name":        id,
					"datatype":    "string",
					"isarray":     true,
					"description": col + " column values",
				}
			}
		case "chart":
			datasets, _ := sec["datasets"].([]interface{})
			for _, rd := range datasets {
				ds, ok := rd.(map[string]interface{})
				if !ok {
					continue
				}
				label, _ := ds["label"].(string)
				if label == "" {
					continue
				}
				id := toSnakeCase(label)
				outputs[id] = map[string]interface{}{
					"name":        id,
					"datatype":    "number",
					"isarray":     true,
					"description": label + " data series",
				}
			}
		}
	}
	return
}

// toSnakeCase converts a human label to a snake_case identifier.
func toSnakeCase(s string) string {
	s = strings.ToLower(s)
	s = strings.ReplaceAll(s, " ", "_")
	s = strings.ReplaceAll(s, "-", "_")
	return s
}

// ─── ui_render tool ───────────────────────────────────────────────────────────

func (s *AgentToolRegistryService) registerUIRenderTool() {
	s.register(ToolDefinition{
		Type: "function",
		Function: ToolFunctionDef{
			Name: "ui_render",
			Description: `Renders a rich, interactive UI panel (dashboards, KPI cards, charts, tables, forms) ` +
				`directly inside the chat window. Prefer this over plain text for any structured data or report. ` +
				`The spec_json parameter is a JSON object with a "sections" array — see the ui_render skill ` +
				`in the skill catalog for the full schema reference and examples.`,
			Parameters: ToolParameterSchema{
				Type: "object",
				Properties: map[string]ToolPropertySchema{
					"spec_json": {
						Type:        "string",
						Description: "JSON object with optional title, subtitle, and sections array defining the UI layout.",
					},
				},
				Required: []string{"spec_json"},
			},
		},
	}, func(ctx context.Context, args map[string]interface{}) (string, error) {
		spec, err := parseUISpec(args)
		if err != nil {
			return "", err
		}
		out, _ := json.Marshal(map[string]interface{}{"ui_json": spec})
		return string(out), nil
	})
}

// htmlEscape escapes the 5 HTML special characters.
func htmlEscape(s string) string {
	s = strings.ReplaceAll(s, "&", "&amp;")
	s = strings.ReplaceAll(s, "<", "&lt;")
	s = strings.ReplaceAll(s, ">", "&gt;")
	s = strings.ReplaceAll(s, `"`, "&#34;")
	s = strings.ReplaceAll(s, "'", "&#39;")
	return s
}
