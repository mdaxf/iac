package services

// File system tools — allow agents to read, create, and modify files with safety guarantees:
//
// - read_file        : read file content (up to 1 MB; truncates with warning)
// - list_directory   : list files and subdirectories
// - write_file       : create or overwrite a file — always backs up existing files first
// - append_to_file   : append lines to a file — backs up existing file first
// - delete_file      : soft-delete by renaming to .deleted.{timestamp}
// - backup_file      : explicit backup of an existing file
// - find_files       : search for files by name pattern or content substring

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

const (
	maxReadBytes   = 1 * 1024 * 1024 // 1 MB read limit
	backupTimeFmt  = "20060102-150405"
)

// blockedPathPrefixes prevents agents from writing to system-critical locations.
var blockedPathPrefixes = []string{
	"/proc", "/sys", "/dev",
	"/etc/passwd", "/etc/shadow", "/etc/sudoers",
	`C:\Windows\System32`, `C:\Windows\SysWOW64`,
}

func validateFSPath(path string) error {
	if path == "" {
		return fmt.Errorf("path is required")
	}
	// Block path traversal components
	if strings.Contains(path, "..") {
		return fmt.Errorf("path traversal (..) is not allowed")
	}
	clean := filepath.Clean(path)
	for _, blocked := range blockedPathPrefixes {
		if strings.HasPrefix(clean, blocked) {
			return fmt.Errorf("access to system path %q is restricted", clean)
		}
	}
	return nil
}

// createBackup copies path → path.bak.{timestamp} and returns the backup path.
// Returns ("", nil) if the file does not exist (nothing to back up).
func createBackup(path string) (string, error) {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return "", nil
	}
	src, err := os.Open(path)
	if err != nil {
		return "", fmt.Errorf("failed to open file for backup: %w", err)
	}
	defer src.Close()

	backupPath := path + ".bak." + time.Now().UTC().Format(backupTimeFmt)
	dst, err := os.Create(backupPath)
	if err != nil {
		return "", fmt.Errorf("failed to create backup file: %w", err)
	}
	defer dst.Close()

	if _, err := io.Copy(dst, src); err != nil {
		return "", fmt.Errorf("failed to write backup: %w", err)
	}
	return backupPath, nil
}

func (s *AgentToolRegistryService) registerFileSystemTools() {
	s.registerReadFileTool()
	s.registerListDirectoryTool()
	s.registerWriteFileTool()
	s.registerAppendToFileTool()
	s.registerDeleteFileTool()
	s.registerBackupFileTool()
	s.registerFindFilesTool()
}

// ─── read_file ────────────────────────────────────────────────────────────────

func (s *AgentToolRegistryService) registerReadFileTool() {
	s.register(ToolDefinition{
		Type: "function",
		Function: ToolFunctionDef{
			Name: "read_file",
			Description: "Reads the content of a file and returns it as a string. " +
				"Files larger than 1 MB are truncated and a warning is included. " +
				"Use max_lines to limit output for very large text files.",
			Parameters: ToolParameterSchema{
				Type: "object",
				Properties: map[string]ToolPropertySchema{
					"path":      {Type: "string", Description: "Absolute path to the file (required). Path traversal (..) is not allowed."},
					"max_lines": {Type: "integer", Description: "Maximum number of lines to return (default: unlimited)"},
					"encoding":  {Type: "string", Description: "File encoding: 'utf8' (default) or 'base64' for binary files"},
				},
				Required: []string{"path"},
			},
		},
	}, func(ctx context.Context, args map[string]interface{}) (string, error) {
		path := stringArg(args, "path")
		if err := validateFSPath(path); err != nil {
			return "", err
		}

		info, err := os.Stat(path)
		if err != nil {
			return "", fmt.Errorf("file not found: %w", err)
		}
		if info.IsDir() {
			return "", fmt.Errorf("%q is a directory — use list_directory instead", path)
		}

		f, err := os.Open(path)
		if err != nil {
			return "", fmt.Errorf("failed to open file: %w", err)
		}
		defer f.Close()

		// Read up to maxReadBytes
		limited := io.LimitReader(f, maxReadBytes+1)
		data, err := io.ReadAll(limited)
		if err != nil {
			return "", fmt.Errorf("failed to read file: %w", err)
		}
		truncated := len(data) > maxReadBytes
		if truncated {
			data = data[:maxReadBytes]
		}

		content := string(data)

		// Apply max_lines if requested
		maxLines := 0
		if v, ok := args["max_lines"].(float64); ok && v > 0 {
			maxLines = int(v)
		}
		linesTruncated := false
		if maxLines > 0 {
			lines := strings.Split(content, "\n")
			if len(lines) > maxLines {
				lines = lines[:maxLines]
				content = strings.Join(lines, "\n")
				linesTruncated = true
			}
		}

		result := map[string]interface{}{
			"path":     path,
			"size":     info.Size(),
			"modified": info.ModTime().UTC().Format(time.RFC3339),
			"content":  content,
		}
		if truncated {
			result["warning"] = fmt.Sprintf("file truncated at 1 MB (%d bytes total)", info.Size())
		}
		if linesTruncated {
			result["lines_truncated"] = true
		}
		out, _ := json.Marshal(result)
		return string(out), nil
	})
}

// ─── list_directory ────────────────────────────────────────────────────────────

func (s *AgentToolRegistryService) registerListDirectoryTool() {
	s.register(ToolDefinition{
		Type: "function",
		Function: ToolFunctionDef{
			Name: "list_directory",
			Description: "Lists files and subdirectories at a given path. " +
				"Returns name, type (file/dir), size, and last modified time for each entry. " +
				"Set recursive to true to list subdirectories recursively (max depth: 4).",
			Parameters: ToolParameterSchema{
				Type: "object",
				Properties: map[string]ToolPropertySchema{
					"path":           {Type: "string", Description: "Directory path to list (required)"},
					"recursive":      {Type: "boolean", Description: "List subdirectories recursively (default: false, max depth: 4)"},
					"include_hidden": {Type: "boolean", Description: "Include hidden files/dirs (starting with '.') (default: false)"},
					"filter":         {Type: "string", Description: "Optional name filter (glob pattern, e.g. '*.go', '*.sql')"},
				},
				Required: []string{"path"},
			},
		},
	}, func(ctx context.Context, args map[string]interface{}) (string, error) {
		path := stringArg(args, "path")
		if err := validateFSPath(path); err != nil {
			return "", err
		}
		recursive := boolArg(args, "recursive")
		includeHidden := boolArg(args, "include_hidden")
		filter := stringArg(args, "filter")

		var entries []map[string]interface{}
		err := listDirEntries(path, filter, includeHidden, recursive, 0, 4, &entries)
		if err != nil {
			return "", fmt.Errorf("failed to list directory: %w", err)
		}

		out, _ := json.Marshal(map[string]interface{}{
			"path":    path,
			"count":   len(entries),
			"entries": entries,
		})
		return string(out), nil
	})
}

func listDirEntries(dir, filter string, includeHidden, recursive bool, depth, maxDepth int, results *[]map[string]interface{}) error {
	if depth > maxDepth {
		return nil
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		return err
	}
	for _, e := range entries {
		name := e.Name()
		if !includeHidden && strings.HasPrefix(name, ".") {
			continue
		}
		if filter != "" {
			matched, _ := filepath.Match(filter, name)
			if !matched {
				if !e.IsDir() {
					continue
				}
			}
		}
		info, _ := e.Info()
		entry := map[string]interface{}{
			"name":  name,
			"path":  filepath.Join(dir, name),
			"type":  "file",
		}
		if e.IsDir() {
			entry["type"] = "dir"
		}
		if info != nil {
			entry["size"] = info.Size()
			entry["modified"] = info.ModTime().UTC().Format(time.RFC3339)
		}
		*results = append(*results, entry)

		if recursive && e.IsDir() {
			listDirEntries(filepath.Join(dir, name), filter, includeHidden, recursive, depth+1, maxDepth, results) //nolint:errcheck
		}
	}
	return nil
}

// ─── write_file ────────────────────────────────────────────────────────────────

func (s *AgentToolRegistryService) registerWriteFileTool() {
	s.register(ToolDefinition{
		Type: "function",
		Function: ToolFunctionDef{
			Name: "write_file",
			Description: "Writes content to a file. " +
				"If the file already exists, it is automatically backed up to {path}.bak.{timestamp} before writing. " +
				"Creates parent directories automatically if they don't exist. " +
				"Use this to create new files or update existing ones.",
			Parameters: ToolParameterSchema{
				Type: "object",
				Properties: map[string]ToolPropertySchema{
					"path":    {Type: "string", Description: "Absolute path to write to (required). Path traversal (..) is not allowed."},
					"content": {Type: "string", Description: "Text content to write (required)"},
					"mode":    {Type: "string", Description: "File permission mode (default: '0644')"},
				},
				Required: []string{"path", "content"},
			},
		},
	}, func(ctx context.Context, args map[string]interface{}) (string, error) {
		path := stringArg(args, "path")
		content := stringArg(args, "content")
		if err := validateFSPath(path); err != nil {
			return "", err
		}

		// Back up existing file
		backupPath, err := createBackup(path)
		if err != nil {
			return "", fmt.Errorf("backup failed before write: %w", err)
		}

		// Ensure parent directory exists
		if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
			return "", fmt.Errorf("failed to create parent directory: %w", err)
		}

		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			return "", fmt.Errorf("failed to write file: %w", err)
		}

		result := map[string]interface{}{
			"path":         path,
			"bytes":        len(content),
			"action":       "created",
		}
		if backupPath != "" {
			result["action"] = "overwritten"
			result["backup"] = backupPath
		}
		out, _ := json.Marshal(result)
		return string(out), nil
	})
}

// ─── append_to_file ───────────────────────────────────────────────────────────

func (s *AgentToolRegistryService) registerAppendToFileTool() {
	s.register(ToolDefinition{
		Type: "function",
		Function: ToolFunctionDef{
			Name: "append_to_file",
			Description: "Appends text to the end of an existing file. " +
				"Automatically backs up the file to {path}.bak.{timestamp} before appending. " +
				"Creates the file if it does not exist (no backup needed).",
			Parameters: ToolParameterSchema{
				Type: "object",
				Properties: map[string]ToolPropertySchema{
					"path":    {Type: "string", Description: "Absolute path to the file (required)"},
					"content": {Type: "string", Description: "Text to append (required). A newline is added automatically before the content if the file is not empty."},
				},
				Required: []string{"path", "content"},
			},
		},
	}, func(ctx context.Context, args map[string]interface{}) (string, error) {
		path := stringArg(args, "path")
		content := stringArg(args, "content")
		if err := validateFSPath(path); err != nil {
			return "", err
		}

		// Backup if exists
		backupPath, err := createBackup(path)
		if err != nil {
			return "", fmt.Errorf("backup failed before append: %w", err)
		}

		if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
			return "", fmt.Errorf("failed to create parent directory: %w", err)
		}

		f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			return "", fmt.Errorf("failed to open file for append: %w", err)
		}
		defer f.Close()

		// Add leading newline if file already has content
		info, _ := f.Stat()
		if info != nil && info.Size() > 0 {
			f.WriteString("\n") //nolint:errcheck
		}
		if _, err := f.WriteString(content); err != nil {
			return "", fmt.Errorf("failed to append to file: %w", err)
		}

		result := map[string]interface{}{
			"path":          path,
			"bytes_appended": len(content),
		}
		if backupPath != "" {
			result["backup"] = backupPath
		}
		out, _ := json.Marshal(result)
		return string(out), nil
	})
}

// ─── delete_file ──────────────────────────────────────────────────────────────

func (s *AgentToolRegistryService) registerDeleteFileTool() {
	s.register(ToolDefinition{
		Type: "function",
		Function: ToolFunctionDef{
			Name: "delete_file",
			Description: "Deletes a file by renaming it to {path}.deleted.{timestamp} (recoverable). " +
				"The original file is NOT permanently removed — it can be restored by renaming it back. " +
				"Use backup_file first if you want an explicit copy before permanent removal.",
			Parameters: ToolParameterSchema{
				Type: "object",
				Properties: map[string]ToolPropertySchema{
					"path": {Type: "string", Description: "Absolute path of the file to delete (required)"},
				},
				Required: []string{"path"},
			},
		},
	}, func(ctx context.Context, args map[string]interface{}) (string, error) {
		path := stringArg(args, "path")
		if err := validateFSPath(path); err != nil {
			return "", err
		}
		if _, err := os.Stat(path); os.IsNotExist(err) {
			return "", fmt.Errorf("file not found: %q", path)
		}

		deletedPath := path + ".deleted." + time.Now().UTC().Format(backupTimeFmt)
		if err := os.Rename(path, deletedPath); err != nil {
			return "", fmt.Errorf("failed to delete file: %w", err)
		}

		out, _ := json.Marshal(map[string]interface{}{
			"deleted_from": path,
			"moved_to":    deletedPath,
			"recoverable": true,
		})
		return string(out), nil
	})
}

// ─── backup_file ──────────────────────────────────────────────────────────────

func (s *AgentToolRegistryService) registerBackupFileTool() {
	s.register(ToolDefinition{
		Type: "function",
		Function: ToolFunctionDef{
			Name: "backup_file",
			Description: "Creates an explicit backup copy of a file at {path}.bak.{timestamp}. " +
				"Use this before risky operations to ensure the original is preserved. " +
				"Returns the backup file path.",
			Parameters: ToolParameterSchema{
				Type: "object",
				Properties: map[string]ToolPropertySchema{
					"path": {Type: "string", Description: "Absolute path of the file to back up (required)"},
				},
				Required: []string{"path"},
			},
		},
	}, func(ctx context.Context, args map[string]interface{}) (string, error) {
		path := stringArg(args, "path")
		if err := validateFSPath(path); err != nil {
			return "", err
		}
		if _, err := os.Stat(path); os.IsNotExist(err) {
			return "", fmt.Errorf("file not found: %q", path)
		}
		backupPath, err := createBackup(path)
		if err != nil {
			return "", err
		}
		out, _ := json.Marshal(map[string]interface{}{
			"source": path,
			"backup": backupPath,
		})
		return string(out), nil
	})
}

// ─── find_files ───────────────────────────────────────────────────────────────

var safeFindPattern = regexp.MustCompile(`^[a-zA-Z0-9_\-\.\*\?\[\]]+$`)

func (s *AgentToolRegistryService) registerFindFilesTool() {
	s.register(ToolDefinition{
		Type: "function",
		Function: ToolFunctionDef{
			Name: "find_files",
			Description: "Searches for files in a directory by name pattern and/or content substring. " +
				"name_pattern supports glob wildcards (e.g. '*.go', '*.sql', 'agent*'). " +
				"content_search does a case-insensitive substring search inside file content. " +
				"Skips hidden directories and common build artifacts (node_modules, .git, dist).",
			Parameters: ToolParameterSchema{
				Type: "object",
				Properties: map[string]ToolPropertySchema{
					"directory":      {Type: "string", Description: "Root directory to search (required)"},
					"name_pattern":   {Type: "string", Description: "Glob file name pattern (e.g. '*.go'). Matches all files if omitted."},
					"content_search": {Type: "string", Description: "Case-insensitive text to search inside file content"},
					"max_depth":      {Type: "integer", Description: "Max directory depth (default: 5)"},
					"max_results":    {Type: "integer", Description: "Max number of matching files to return (default: 30)"},
				},
				Required: []string{"directory"},
			},
		},
	}, func(ctx context.Context, args map[string]interface{}) (string, error) {
		dir := stringArg(args, "directory")
		if err := validateFSPath(dir); err != nil {
			return "", err
		}
		namePattern := stringArg(args, "name_pattern")
		contentSearch := strings.ToLower(stringArg(args, "content_search"))
		maxDepth := 5
		if v, ok := args["max_depth"].(float64); ok && v > 0 {
			maxDepth = int(v)
		}
		maxResults := 30
		if v, ok := args["max_results"].(float64); ok && v > 0 {
			maxResults = int(v)
		}

		var matches []map[string]interface{}
		skipDirs := map[string]bool{
			"node_modules": true, ".git": true, "dist": true, "build": true,
			"vendor": true, ".cache": true, "__pycache__": true, ".idea": true,
		}

		var walk func(path string, depth int) error
		walk = func(path string, depth int) error {
			if depth > maxDepth || len(matches) >= maxResults {
				return nil
			}
			entries, err := os.ReadDir(path)
			if err != nil {
				return nil // skip inaccessible dirs
			}
			for _, e := range entries {
				if len(matches) >= maxResults {
					break
				}
				name := e.Name()
				if strings.HasPrefix(name, ".") {
					continue
				}
				fullPath := filepath.Join(path, name)

				if e.IsDir() {
					if skipDirs[name] {
						continue
					}
					walk(fullPath, depth+1) //nolint:errcheck
					continue
				}

				// Name pattern match
				if namePattern != "" {
					matched, _ := filepath.Match(namePattern, name)
					if !matched {
						continue
					}
				}

				// Content search
				if contentSearch != "" {
					data, err := os.ReadFile(fullPath)
					if err != nil {
						continue
					}
					// Skip binary files (check for null bytes in first 512 bytes)
					preview := data
					if len(preview) > 512 {
						preview = preview[:512]
					}
					if strings.ContainsRune(string(preview), 0) {
						continue
					}
					if !strings.Contains(strings.ToLower(string(data)), contentSearch) {
						continue
					}
				}

				info, _ := e.Info()
				entry := map[string]interface{}{"path": fullPath, "name": name}
				if info != nil {
					entry["size"] = info.Size()
					entry["modified"] = info.ModTime().UTC().Format(time.RFC3339)
				}
				matches = append(matches, entry)
			}
			return nil
		}
		walk(dir, 0) //nolint:errcheck

		out, _ := json.Marshal(map[string]interface{}{
			"directory": dir,
			"count":     len(matches),
			"matches":   matches,
		})
		return string(out), nil
	})
}
