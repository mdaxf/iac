package services

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/mdaxf/iac/models"
)

// RegisterInstalledSkill registers all tools from an installed skill manifest
// into the tool registry so agents can call them via the ReAct loop.
// Each tool is backed by a subprocess handler that executes the skill's entry point.
func (s *AgentToolRegistryService) RegisterInstalledSkill(manifest *models.SkillManifest, installDir string) error {
	if len(manifest.Tools) == 0 {
		// Register a single default tool named after the skill
		return s.registerInstalledTool(manifest.Name, manifest.Name,
			manifest.Description, manifest, installDir, nil)
	}
	for _, td := range manifest.Tools {
		if err := s.registerInstalledTool(td.Name, manifest.Name,
			td.Description, manifest, installDir, td.Parameters); err != nil {
			return err
		}
	}
	return nil
}

// UnregisterInstalledSkill removes all tools registered for a skill name.
// It searches for tools whose names match registered installed skill tool names
// by looking for the skill tag prefix.
func (s *AgentToolRegistryService) UnregisterInstalledSkill(skillName string) {
	// Tools registered for an installed skill are tagged in the description.
	// We track them via a naming convention: the handler key matches the tool name.
	// We can simply remove the defs/handlers that are installed-skill tools for this skill.
	toDelete := []string{}
	for name := range s.handlers {
		if isInstalledSkillTool(name, skillName) {
			toDelete = append(toDelete, name)
		}
	}
	for _, name := range toDelete {
		delete(s.handlers, name)
		delete(s.defs, name)
	}
}

// isInstalledSkillTool returns true when toolName belongs to a given skillName.
// Convention: tool names are either exactly the skillName or have the prefix "skillName_".
func isInstalledSkillTool(toolName, skillName string) bool {
	return toolName == skillName || strings.HasPrefix(toolName, skillName+"_")
}

// registerInstalledTool registers a single tool that executes via an external process.
func (s *AgentToolRegistryService) registerInstalledTool(
	toolName, skillName, description string,
	manifest *models.SkillManifest,
	installDir string,
	rawParams map[string]interface{},
) error {
	params := buildToolParams(rawParams)
	def := ToolDefinition{
		Type: "function",
		Function: ToolFunctionDef{
			Name:        toolName,
			Description: description,
			Parameters:  params,
		},
	}

	// Read SKILL.md body for context injection
	skillMDBody := readSkillMDBody(installDir)

	handler := func(ctx context.Context, args map[string]interface{}) (string, error) {
		return executeInstalledSkillTool(ctx, manifest, installDir, toolName, args, skillMDBody)
	}

	s.register(def, handler)
	s.iLog.Info(fmt.Sprintf("Registered installed skill tool: %s (skill: %s)", toolName, skillName))
	return nil
}

// executeInstalledSkillTool runs the skill's entry-point script and returns its stdout.
// Args are passed as JSON via stdin. The skill context from SKILL.md is injected as
// the SKILL_CONTEXT environment variable for the subprocess.
func executeInstalledSkillTool(
	ctx context.Context,
	manifest *models.SkillManifest,
	installDir, toolName string,
	args map[string]interface{},
	skillMDBody string,
) (string, error) {
	if manifest.EntryPoint == "" {
		return "", fmt.Errorf("skill '%s' has no entry_point defined", manifest.Name)
	}

	entryAbs := filepath.Join(installDir, manifest.EntryPoint)
	if _, err := os.Stat(entryAbs); err != nil {
		return "", fmt.Errorf("skill entry point not found: %s", entryAbs)
	}

	// Build input payload
	input := map[string]interface{}{
		"tool":    toolName,
		"args":    args,
		"skill":   manifest.Name,
		"version": manifest.Version,
	}
	inputJSON, err := json.Marshal(input)
	if err != nil {
		return "", fmt.Errorf("failed to marshal tool input: %w", err)
	}

	// Build the command based on runtime
	cmd, err := buildSkillCommand(ctx, manifest.Runtime, entryAbs, installDir)
	if err != nil {
		return "", err
	}

	cmd.Stdin = bytes.NewReader(inputJSON)
	cmd.Env = append(os.Environ(),
		"SKILL_DIR="+installDir,
		"SKILL_NAME="+manifest.Name,
		"SKILL_TOOL="+toolName,
		"SKILL_CONTEXT="+skillMDBody,
	)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	// Timeout: use context deadline or default 60s
	if _, hasDeadline := ctx.Deadline(); !hasDeadline {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, 60*time.Second)
		defer cancel()
		_ = ctx
	}

	if err := cmd.Run(); err != nil {
		errMsg := stderr.String()
		if errMsg == "" {
			errMsg = err.Error()
		}
		return "", fmt.Errorf("skill '%s' tool '%s' execution failed: %s", manifest.Name, toolName, errMsg)
	}

	result := stdout.String()
	if result == "" {
		result = fmt.Sprintf(`{"status":"ok","tool":"%s"}`, toolName)
	}
	return result, nil
}

// buildSkillCommand constructs an exec.Cmd for the given runtime and entry point.
func buildSkillCommand(ctx context.Context, skillRuntime, entryAbs, workDir string) (*exec.Cmd, error) {
	var cmd *exec.Cmd
	switch strings.ToLower(skillRuntime) {
	case "python", "python3":
		pyBin := "python3"
		if runtime.GOOS == "windows" {
			pyBin = "python"
		}
		cmd = exec.CommandContext(ctx, pyBin, entryAbs)
	case "node", "nodejs":
		cmd = exec.CommandContext(ctx, "node", entryAbs)
	case "bash", "sh":
		if runtime.GOOS == "windows" {
			cmd = exec.CommandContext(ctx, "bash", entryAbs)
		} else {
			cmd = exec.CommandContext(ctx, "bash", entryAbs)
		}
	case "none", "":
		return nil, fmt.Errorf("skill runtime is 'none' — tool has no executable entry point")
	default:
		// Try to run directly (e.g. compiled binary)
		cmd = exec.CommandContext(ctx, entryAbs)
	}
	cmd.Dir = workDir
	return cmd, nil
}

// readSkillMDBody reads SKILL.md from the install directory and returns the markdown body
// (everything after the YAML frontmatter). Returns empty string if file not found.
func readSkillMDBody(installDir string) string {
	data, err := os.ReadFile(filepath.Join(installDir, "SKILL.md"))
	if err != nil {
		return ""
	}
	_, body := parseSkillMD(data)
	return body
}

// buildToolParams converts a raw JSON-schema parameters map to a ToolParameterSchema.
// Falls back to an empty object schema if nil or malformed.
func buildToolParams(raw map[string]interface{}) ToolParameterSchema {
	if raw == nil {
		return ToolParameterSchema{Type: "object", Properties: map[string]ToolPropertySchema{}}
	}

	// Re-encode and decode through ToolParameterSchema
	b, err := json.Marshal(raw)
	if err != nil {
		return ToolParameterSchema{Type: "object", Properties: map[string]ToolPropertySchema{}}
	}

	// Try to extract top-level properties
	var schema struct {
		Type       string                     `json:"type"`
		Properties map[string]json.RawMessage `json:"properties"`
		Required   []string                   `json:"required"`
	}
	if err := json.Unmarshal(b, &schema); err != nil {
		return ToolParameterSchema{Type: "object", Properties: map[string]ToolPropertySchema{}}
	}

	props := make(map[string]ToolPropertySchema)
	for name, rawProp := range schema.Properties {
		var prop struct {
			Type        string `json:"type"`
			Description string `json:"description"`
		}
		if err := json.Unmarshal(rawProp, &prop); err != nil {
			continue
		}
		props[name] = ToolPropertySchema{Type: prop.Type, Description: prop.Description}
	}

	return ToolParameterSchema{
		Type:       "object",
		Properties: props,
		Required:   schema.Required,
	}
}

// ReadSkillContext reads the SKILL.md body for the named skill from its install path.
// Returns empty string if skill is not installed or SKILL.md is missing.
func ReadSkillContext(skillInstallPath string) string {
	if skillInstallPath == "" {
		return ""
	}
	data, err := os.ReadFile(filepath.Join(skillInstallPath, "SKILL.md"))
	if err != nil {
		return ""
	}
	content := string(data)
	// Strip YAML frontmatter
	if strings.HasPrefix(content, "---") {
		end := strings.Index(content[3:], "---")
		if end >= 0 {
			return strings.TrimSpace(content[end+6:])
		}
	}
	return content
}
