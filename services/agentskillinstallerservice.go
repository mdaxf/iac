package services

import (
	"archive/zip"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/mdaxf/iac/logger"
	"github.com/mdaxf/iac/models"
	"gopkg.in/yaml.v3"
)

// SkillsBaseDir is the root directory for installed skills.
// Override at startup (e.g. from config) before any install call.
var SkillsBaseDir = "./skills"

// InstallSkillRequest carries the installation URL and requester identity.
type InstallSkillRequest struct {
	URL         string `json:"url"`
	InstalledBy string `json:"installed_by"`
}

// InstallSkillFromURL downloads a skill package from the given URL, extracts it
// to SkillsBaseDir/{name}/, registers its tools in the tool registry, and
// persists the skill record in MongoDB.
//
// Supported URL formats:
//   - ZIP / .skill file → extracted into SkillsBaseDir/{name}/
//   - Raw skill.json → saved as manifest; script download not attempted
func (s *AgentDefinitionService) InstallSkillFromURL(req InstallSkillRequest) (*models.AgentSkill, error) {
	iLog := logger.Log{ModuleName: logger.Framework, User: req.InstalledBy, ControllerName: "SkillInstaller"}

	if req.URL == "" {
		return nil, fmt.Errorf("URL is required")
	}
	if err := s.ensureDocDB(); err != nil {
		return nil, err
	}

	iLog.Info(fmt.Sprintf("Installing skill from URL: %s", req.URL))

	// Download raw bytes
	data, contentType, err := downloadSkillURL(req.URL)
	if err != nil {
		return nil, fmt.Errorf("download failed: %w", err)
	}

	var manifest *models.SkillManifest
	var installDir string

	if isSkillZip(data, contentType, req.URL) {
		manifest, installDir, err = extractZipSkill(data)
	} else {
		manifest, installDir, err = extractJSONSkill(data)
	}
	if err != nil {
		return nil, fmt.Errorf("skill extraction failed: %w", err)
	}
	if manifest.Name == "" {
		return nil, fmt.Errorf("skill manifest missing 'name' field")
	}

	// Register tools in the running tool registry
	registry := GetGlobalToolRegistry()
	if registry != nil {
		if regErr := registry.RegisterInstalledSkill(manifest, installDir); regErr != nil {
			iLog.Error(fmt.Sprintf("Tool registry registration warning: %v", regErr))
		}
	}

	// Serialise tool definitions for storage
	toolDefsJSON, _ := json.Marshal(manifest.Tools)

	displayName := manifest.DisplayName
	if displayName == "" {
		displayName = manifest.Name
	}
	category := manifest.Category
	if category == "" {
		category = "installed"
	}

	ctx := context.Background()
	now := time.Now()

	// Update existing record or create new one
	existing, _ := s.GetSkillByName(manifest.Name)
	if existing != nil {
		filter := map[string]interface{}{"name": manifest.Name}
		update := map[string]interface{}{
			"$set": map[string]interface{}{
				"display_name": displayName,
				"description":  manifest.Description,
				"category":     category,
				"skill_type":   "installed",
				"source_url":   req.URL,
				"install_path": installDir,
				"runtime":      manifest.Runtime,
				"entry_point":  manifest.EntryPoint,
				"tool_defs":    string(toolDefsJSON),
				"version":      manifest.Version,
				"active":       true,
				"modifiedby":   req.InstalledBy,
				"modifiedon":   now.Format(time.RFC3339),
			},
			"$inc": map[string]interface{}{"rowversionstamp": 1},
		}
		if err := s.docDB.UpdateOne(ctx, models.CollectionAgentSkills, filter, update); err != nil {
			return nil, fmt.Errorf("failed to update skill record: %w", err)
		}
		return s.GetSkill(existing.ID)
	}

	// Create new record
	doc := models.AgentSkillDoc{
		ID:              uuid.New().String(),
		Name:            manifest.Name,
		DisplayName:     displayName,
		Description:     manifest.Description,
		Category:        category,
		Active:          true,
		SkillType:       "installed",
		SourceURL:       req.URL,
		InstallPath:     installDir,
		Runtime:         manifest.Runtime,
		EntryPoint:      manifest.EntryPoint,
		ToolDefs:        string(toolDefsJSON),
		Version:         manifest.Version,
		CreatedBy:       req.InstalledBy,
		ModifiedBy:      req.InstalledBy,
		CreatedOn:       now,
		ModifiedOn:      now,
		RowVersionStamp: 1,
	}
	m, err := docToMap(doc)
	if err != nil {
		return nil, err
	}
	if _, err := s.docDB.InsertOne(ctx, models.CollectionAgentSkills, m); err != nil {
		return nil, fmt.Errorf("failed to persist skill: %w", err)
	}

	iLog.Info(fmt.Sprintf("Skill '%s' installed to %s", manifest.Name, installDir))
	return models.AgentSkillDocToModel(&doc), nil
}

// UninstallSkill removes skill files from disk and soft-deletes the MongoDB record.
func (s *AgentDefinitionService) UninstallSkill(id string, deletedBy string) error {
	if err := s.ensureDocDB(); err != nil {
		return err
	}
	skill, err := s.GetSkill(id)
	if err != nil {
		return err
	}
	if skill.SkillType == "installed" {
		registry := GetGlobalToolRegistry()
		if registry != nil {
			registry.UnregisterInstalledSkill(skill.Name)
		}
		if skill.InstallPath != "" {
			if rmErr := os.RemoveAll(skill.InstallPath); rmErr != nil {
				iLog := logger.Log{ModuleName: logger.Framework, User: deletedBy, ControllerName: "SkillInstaller"}
				iLog.Error(fmt.Sprintf("Failed to remove skill files at %s: %v", skill.InstallPath, rmErr))
			}
		}
	}
	return s.DeleteSkill(id, deletedBy)
}

// GetSkillByName looks up a skill by name (active or not).
func (s *AgentDefinitionService) GetSkillByName(name string) (*models.AgentSkill, error) {
	if err := s.ensureDocDB(); err != nil {
		return nil, err
	}
	ctx := context.Background()
	result, err := s.docDB.FindOne(ctx, models.CollectionAgentSkills, map[string]interface{}{"name": name})
	if err != nil {
		return nil, err
	}
	var doc models.AgentSkillDoc
	if err := mapToDoc(result, &doc); err != nil {
		return nil, err
	}
	return models.AgentSkillDocToModel(&doc), nil
}

// LoadInstalledSkillsFromDisk scans SkillsBaseDir and re-registers all installed
// skill tools with the tool registry. Called at startup so tools survive restarts.
func (s *AgentDefinitionService) LoadInstalledSkillsFromDisk() {
	iLog := logger.Log{ModuleName: logger.Framework, User: "System", ControllerName: "SkillInstaller"}

	entries, err := os.ReadDir(SkillsBaseDir)
	if err != nil {
		if os.IsNotExist(err) {
			return
		}
		iLog.Error(fmt.Sprintf("Failed to read skills dir %s: %v", SkillsBaseDir, err))
		return
	}

	registry := GetGlobalToolRegistry()
	if registry == nil {
		return
	}

	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		skillDir := filepath.Join(SkillsBaseDir, e.Name())
		manifest, err := readSkillManifest(skillDir)
		if err != nil {
			iLog.Error(fmt.Sprintf("Skipping skill dir %s: %v", skillDir, err))
			continue
		}
		if err := registry.RegisterInstalledSkill(manifest, skillDir); err != nil {
			iLog.Error(fmt.Sprintf("Failed to register skill %s: %v", manifest.Name, err))
		} else {
			iLog.Info(fmt.Sprintf("Loaded installed skill: %s", manifest.Name))
		}
	}
}

// ---- internal helpers ----

func downloadSkillURL(rawURL string) ([]byte, string, error) {
	resp, err := http.Get(rawURL) //nolint:gosec
	if err != nil {
		return nil, "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, "", fmt.Errorf("HTTP %d from %s", resp.StatusCode, rawURL)
	}
	data, err := io.ReadAll(resp.Body)
	return data, resp.Header.Get("Content-Type"), err
}

func isSkillZip(data []byte, contentType, rawURL string) bool {
	lower := strings.ToLower(rawURL)
	if strings.HasSuffix(lower, ".zip") || strings.HasSuffix(lower, ".skill") {
		return true
	}
	if strings.Contains(contentType, "zip") || strings.Contains(contentType, "octet-stream") {
		return true
	}
	return len(data) >= 4 && data[0] == 'P' && data[1] == 'K' && data[2] == 0x03 && data[3] == 0x04
}

func extractZipSkill(data []byte) (*models.SkillManifest, string, error) {
	r, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return nil, "", fmt.Errorf("not a valid zip: %w", err)
	}

	// Read manifest first
	var manifest *models.SkillManifest
	for _, f := range r.File {
		if filepath.Base(f.Name) == "skill.json" && !f.FileInfo().IsDir() {
			rc, err := f.Open()
			if err != nil {
				return nil, "", err
			}
			b, err := io.ReadAll(rc)
			rc.Close()
			if err != nil {
				return nil, "", err
			}
			manifest, err = parseSkillManifest(b)
			if err != nil {
				return nil, "", err
			}
			break
		}
	}
	if manifest == nil {
		return nil, "", fmt.Errorf("skill.json not found in archive")
	}

	skillDir := filepath.Join(SkillsBaseDir, manifest.Name)
	if err := os.MkdirAll(skillDir, 0755); err != nil {
		return nil, "", err
	}

	prefix := detectZipPrefix(r.File)
	for _, f := range r.File {
		rel := strings.TrimPrefix(f.Name, prefix)
		if rel == "" {
			continue
		}
		destPath := filepath.Join(skillDir, rel)
		if !strings.HasPrefix(filepath.Clean(destPath), filepath.Clean(skillDir)) {
			continue // path traversal guard
		}
		if f.FileInfo().IsDir() {
			os.MkdirAll(destPath, 0755) //nolint:errcheck
			continue
		}
		os.MkdirAll(filepath.Dir(destPath), 0755) //nolint:errcheck
		rc, err := f.Open()
		if err != nil {
			return nil, "", err
		}
		out, err := os.Create(destPath)
		if err != nil {
			rc.Close()
			return nil, "", err
		}
		io.Copy(out, rc) //nolint:errcheck
		out.Close()
		rc.Close()
	}

	return manifest, skillDir, nil
}

func extractJSONSkill(data []byte) (*models.SkillManifest, string, error) {
	manifest, err := parseSkillManifest(data)
	if err != nil {
		return nil, "", fmt.Errorf("invalid skill.json: %w", err)
	}
	if manifest.Name == "" {
		return nil, "", fmt.Errorf("skill.json must have a 'name' field")
	}
	skillDir := filepath.Join(SkillsBaseDir, manifest.Name)
	if err := os.MkdirAll(skillDir, 0755); err != nil {
		return nil, "", err
	}
	if err := os.WriteFile(filepath.Join(skillDir, "skill.json"), data, 0644); err != nil {
		return nil, "", err
	}
	return manifest, skillDir, nil
}

func parseSkillManifest(data []byte) (*models.SkillManifest, error) {
	var m models.SkillManifest
	if err := json.Unmarshal(data, &m); err != nil {
		return nil, err
	}
	return &m, nil
}

func readSkillManifest(skillDir string) (*models.SkillManifest, error) {
	data, err := os.ReadFile(filepath.Join(skillDir, "skill.json"))
	if err != nil {
		return nil, err
	}
	return parseSkillManifest(data)
}

func detectZipPrefix(files []*zip.File) string {
	if len(files) == 0 {
		return ""
	}
	first := files[0].Name
	slash := strings.Index(first, "/")
	if slash < 0 {
		return ""
	}
	prefix := first[:slash+1]
	for _, f := range files[1:] {
		if !strings.HasPrefix(f.Name, prefix) {
			return ""
		}
	}
	return prefix
}

// skillFrontmatter holds the YAML front-matter from a SKILL.md file.
type skillFrontmatter struct {
	Name        string `yaml:"name"`
	Description string `yaml:"description"`
}

// parseSkillMD parses YAML front-matter from SKILL.md content.
// Returns the parsed frontmatter and the markdown body.
func parseSkillMD(data []byte) (*skillFrontmatter, string) {
	content := string(data)
	if !strings.HasPrefix(content, "---") {
		return &skillFrontmatter{}, content
	}
	end := strings.Index(content[3:], "---")
	if end < 0 {
		return &skillFrontmatter{}, content
	}
	yamlBlock := content[3 : end+3]
	body := content[end+6:]
	var fm skillFrontmatter
	yaml.Unmarshal([]byte(yamlBlock), &fm) //nolint:errcheck
	return &fm, body
}
