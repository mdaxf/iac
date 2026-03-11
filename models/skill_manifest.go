package models

// SkillManifest describes the metadata and tool definitions of an installable skill.
// It is read from skill.json inside the skill package/folder.
type SkillManifest struct {
	Name        string          `json:"name"`
	DisplayName string          `json:"display_name"`
	Version     string          `json:"version"`
	Description string          `json:"description"`
	Category    string          `json:"category"`
	Runtime     string          `json:"runtime"`      // "python" | "node" | "bash" | "none"
	EntryPoint  string          `json:"entry_point"`  // relative path within skill dir
	Tools       []SkillToolDef  `json:"tools"`
}

// SkillToolDef defines a single callable tool exposed by an installed skill.
type SkillToolDef struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Parameters  map[string]interface{} `json:"parameters"` // JSON Schema object
}
