package services

// Deployment package management tools — allow agents to create, manage, deploy,
// export, import, and push IAC data packages between environments.
//
// Tools registered here:
//   package_list        — list packages with optional filters
//   package_get         — get a single package by ID
//   package_create      — create/pack a new package (database or document)
//   package_export      — export package JSON to a local file
//   package_import_file — import package from a local JSON file
//   package_deploy      — deploy a package to a target environment
//   package_push        — push a package to a remote IAC instance
//   definition_list     — list all package definitions
//   definition_pack     — trigger packing from a saved definition

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	dbconn "github.com/mdaxf/iac/databases"
	depmodels "github.com/mdaxf/iac/deployment/models"
	deploymgr "github.com/mdaxf/iac/deployment/deploy"
	packagemgr "github.com/mdaxf/iac/deployment/package"
	deprepo "github.com/mdaxf/iac/deployment/repository"
	"github.com/mdaxf/iac/documents"
	"github.com/mdaxf/iac/signalrserver"
	"go.mongodb.org/mongo-driver/bson"
)

func (s *AgentToolRegistryService) registerDeploymentTools() {
	s.registerPackageListTool()
	s.registerPackageGetTool()
	s.registerPackageCreateTool()
	s.registerPackageExportTool()
	s.registerPackageImportFileTool()
	s.registerPackageDeployTool()
	s.registerPackagePushTool()
	s.registerDefinitionListTool()
	s.registerDefinitionPackTool()
}

// ── package_list ──────────────────────────────────────────────────────────────

func (s *AgentToolRegistryService) registerPackageListTool() {
	s.register(ToolDefinition{
		Type: "function",
		Function: ToolFunctionDef{
			Name:        "package_list",
			Description: "List IAC data packages. Returns package metadata: ID, name, version, type, status, environment, and record counts. Use package_get to retrieve full package details.",
			Parameters: ToolParameterSchema{
				Type: "object",
				Properties: map[string]ToolPropertySchema{
					"package_type": {Type: "string", Description: "Filter by type: 'database' or 'document'. Omit for all types."},
					"environment":  {Type: "string", Description: "Filter by environment: 'dev', 'staging', 'production', etc."},
					"status":       {Type: "string", Description: "Filter by status: 'active', 'archived'. Omit for all active."},
					"limit":        {Type: "string", Description: "Max results to return (default 20, max 100)."},
					"offset":       {Type: "string", Description: "Pagination offset (default 0)."},
				},
			},
		},
	}, func(ctx context.Context, args map[string]interface{}) (string, error) {
		if s.sqlDB == nil {
			return "", fmt.Errorf("database not available")
		}

		packageType := stringArg(args, "package_type")
		environment := stringArg(args, "environment")
		status := stringArg(args, "status")
		limit := intArgDefault(args, "limit", 20)
		offset := intArgDefault(args, "offset", 0)
		if limit > 100 {
			limit = 100
		}

		dbTx, err := s.sqlDB.BeginTx(ctx, nil)
		if err != nil {
			return "", fmt.Errorf("failed to begin transaction: %w", err)
		}
		defer dbTx.Rollback() //nolint:errcheck

		repo := deprepo.NewPackageRepository("AgentTool", dbTx)
		packages, err := repo.ListPackages(packageType, environment, status, limit, offset)
		if err != nil {
			return "", fmt.Errorf("failed to list packages: %w", err)
		}
		dbTx.Commit() //nolint:errcheck

		out := map[string]interface{}{
			"count":    len(packages),
			"packages": packages,
		}
		b, _ := json.Marshal(out)
		return string(b), nil
	})
}

// ── package_get ───────────────────────────────────────────────────────────────

func (s *AgentToolRegistryService) registerPackageGetTool() {
	s.register(ToolDefinition{
		Type: "function",
		Function: ToolFunctionDef{
			Name:        "package_get",
			Description: "Get detailed information about a specific package by ID, including its content summary and recent deployment actions.",
			Parameters: ToolParameterSchema{
				Type:       "object",
				Properties: map[string]ToolPropertySchema{"id": {Type: "string", Description: "Package ID (UUID)"}},
				Required:   []string{"id"},
			},
		},
	}, func(ctx context.Context, args map[string]interface{}) (string, error) {
		id := stringArg(args, "id")
		if id == "" {
			return "", fmt.Errorf("id is required")
		}
		if s.sqlDB == nil {
			return "", fmt.Errorf("database not available")
		}

		dbTx, err := s.sqlDB.BeginTx(ctx, nil)
		if err != nil {
			return "", fmt.Errorf("failed to begin transaction: %w", err)
		}
		defer dbTx.Rollback() //nolint:errcheck

		repo := deprepo.NewPackageRepository("AgentTool", dbTx)
		pkg, err := repo.GetPackage(id)
		if err != nil {
			return "", fmt.Errorf("package not found: %w", err)
		}

		// Build summary
		summary := map[string]interface{}{
			"id":           pkg.ID,
			"name":         pkg.Name,
			"version":      pkg.Version,
			"package_type": pkg.PackageType,
			"created_by":   pkg.CreatedBy,
			"created_at":   pkg.CreatedAt,
		}

		if pkg.DatabaseData != nil {
			tables := make([]map[string]interface{}, 0, len(pkg.DatabaseData.Tables))
			for _, t := range pkg.DatabaseData.Tables {
				tables = append(tables, map[string]interface{}{
					"table":     t.TableName,
					"row_count": t.RowCount,
				})
			}
			summary["database_type"] = pkg.DatabaseData.DatabaseType
			summary["tables"] = tables
		}
		if pkg.DocumentData != nil {
			colls := make([]map[string]interface{}, 0, len(pkg.DocumentData.Collections))
			for _, c := range pkg.DocumentData.Collections {
				colls = append(colls, map[string]interface{}{
					"collection":     c.CollectionName,
					"document_count": c.DocumentCount,
				})
			}
			summary["collections"] = colls
		}

		actions, _ := repo.GetActionsByPackage(id, 5)
		summary["recent_actions"] = actions
		dbTx.Commit() //nolint:errcheck

		b, _ := json.Marshal(summary)
		return string(b), nil
	})
}

// ── package_create ─────────────────────────────────────────────────────────────

func (s *AgentToolRegistryService) registerPackageCreateTool() {
	s.register(ToolDefinition{
		Type: "function",
		Function: ToolFunctionDef{
			Name: "package_create",
			Description: "Create a new data package by packing selected tables or MongoDB collections. " +
				"For 'database' type, specify tables to include. For 'document' type, specify collections. " +
				"Returns the package ID and summary. Use package_deploy afterward to deploy the package.",
			Parameters: ToolParameterSchema{
				Type: "object",
				Properties: map[string]ToolPropertySchema{
					"name":              {Type: "string", Description: "Package name (e.g. 'user-config-export')"},
					"version":           {Type: "string", Description: "Version string (e.g. '1.0.0')"},
					"package_type":      {Type: "string", Description: "'database' for SQL tables, 'document' for MongoDB collections"},
					"description":       {Type: "string", Description: "Optional description of what this package contains"},
					"tables_json":       {Type: "string", Description: "JSON array of SQL table names to include. Required for 'database' type. Example: [\"users\",\"roles\"]"},
					"collections_json":  {Type: "string", Description: "JSON array of MongoDB collection names to include. Required for 'document' type."},
					"where_clause_json": {Type: "string", Description: "Optional JSON object mapping table/collection names to WHERE conditions. Example: {\"users\":\"active=true\"}"},
					"environment":       {Type: "string", Description: "Target environment: 'dev', 'staging', or 'production'"},
				},
				Required: []string{"name", "version", "package_type"},
			},
		},
	}, func(ctx context.Context, args map[string]interface{}) (string, error) {
		name := stringArg(args, "name")
		version := stringArg(args, "version")
		pkgType := stringArg(args, "package_type")
		description := stringArg(args, "description")
		environment := stringArg(args, "environment")

		if name == "" || version == "" || pkgType == "" {
			return "", fmt.Errorf("name, version, and package_type are required")
		}
		if pkgType != "database" && pkgType != "document" {
			return "", fmt.Errorf("package_type must be 'database' or 'document'")
		}
		if s.sqlDB == nil {
			return "", fmt.Errorf("database not available")
		}

		filter := depmodels.PackageFilter{}

		if tablesJSON := stringArg(args, "tables_json"); tablesJSON != "" {
			if err := json.Unmarshal([]byte(tablesJSON), &filter.Tables); err != nil {
				return "", fmt.Errorf("invalid tables_json: %w", err)
			}
		}
		if colsJSON := stringArg(args, "collections_json"); colsJSON != "" {
			if err := json.Unmarshal([]byte(colsJSON), &filter.Collections); err != nil {
				return "", fmt.Errorf("invalid collections_json: %w", err)
			}
		}
		if whereJSON := stringArg(args, "where_clause_json"); whereJSON != "" {
			if err := json.Unmarshal([]byte(whereJSON), &filter.WhereClause); err != nil {
				return "", fmt.Errorf("invalid where_clause_json: %w", err)
			}
		}

		dbTx, err := s.sqlDB.BeginTx(ctx, nil)
		if err != nil {
			return "", fmt.Errorf("failed to begin transaction: %w", err)
		}
		defer dbTx.Rollback() //nolint:errcheck

		var pkg *depmodels.Package

		if pkgType == "database" {
			if len(filter.Tables) == 0 {
				return "", fmt.Errorf("tables_json is required for database packages")
			}
			packager := packagemgr.NewDatabasePackager("AgentTool", dbTx, dbconn.DatabaseType)
			p, err := packager.PackageTables(name, version, "AgentTool", filter)
			if err != nil {
				return "", fmt.Errorf("packaging failed: %w", err)
			}
			pkg = p
		} else {
			if len(filter.Collections) == 0 {
				return "", fmt.Errorf("collections_json is required for document packages")
			}
			if documents.DocDBCon == nil {
				return "", fmt.Errorf("document database not available")
			}
			packager := packagemgr.NewDocumentPackager(documents.DocDBCon, "AgentTool")
			p, err := packager.PackageCollections(name, version, "AgentTool", filter)
			if err != nil {
				return "", fmt.Errorf("packaging failed: %w", err)
			}
			pkg = p
		}

		if description != "" {
			if pkg.Metadata == nil {
				pkg.Metadata = make(map[string]interface{})
			}
			pkg.Metadata["description"] = description
		}

		repo := deprepo.NewPackageRepository("AgentTool", dbTx)
		record, err := repo.SavePackage(pkg, environment, &filter)
		if err != nil {
			return "", fmt.Errorf("failed to save package: %w", err)
		}

		dbTx.Commit() //nolint:errcheck

		out := map[string]interface{}{
			"package_id":   record.ID,
			"name":         record.Name,
			"version":      record.Version,
			"package_type": record.PackageType,
			"file_size":    record.FileSize,
			"checksum":     record.Checksum,
			"environment":  record.Environment,
			"message":      fmt.Sprintf("Package '%s v%s' created successfully (ID: %s)", name, version, record.ID),
		}
		b, _ := json.Marshal(out)
		return string(b), nil
	})
}

// ── package_export ─────────────────────────────────────────────────────────────

func (s *AgentToolRegistryService) registerPackageExportTool() {
	s.register(ToolDefinition{
		Type: "function",
		Function: ToolFunctionDef{
			Name:        "package_export",
			Description: "Export a package to a local JSON file. The package data is saved to the specified path. Returns the file path and size.",
			Parameters: ToolParameterSchema{
				Type: "object",
				Properties: map[string]ToolPropertySchema{
					"id":          {Type: "string", Description: "Package ID to export"},
					"output_path": {Type: "string", Description: "Local file path to save the package JSON (e.g. /exports/my-package.json)"},
				},
				Required: []string{"id", "output_path"},
			},
		},
	}, func(ctx context.Context, args map[string]interface{}) (string, error) {
		id := stringArg(args, "id")
		outputPath := stringArg(args, "output_path")

		if id == "" || outputPath == "" {
			return "", fmt.Errorf("id and output_path are required")
		}
		if err := validateFSPath(outputPath); err != nil {
			return "", err
		}
		if s.sqlDB == nil {
			return "", fmt.Errorf("database not available")
		}

		dbTx, err := s.sqlDB.BeginTx(ctx, nil)
		if err != nil {
			return "", fmt.Errorf("failed to begin transaction: %w", err)
		}
		defer dbTx.Rollback() //nolint:errcheck

		repo := deprepo.NewPackageRepository("AgentTool", dbTx)
		pkg, err := repo.GetPackage(id)
		if err != nil {
			return "", fmt.Errorf("package not found: %w", err)
		}

		var data []byte
		if pkg.PackageType == "database" {
			packager := packagemgr.NewDatabasePackager("AgentTool", dbTx, dbconn.DatabaseType)
			data, err = packager.ExportPackage(pkg)
		} else {
			packager := packagemgr.NewDocumentPackager(documents.DocDBCon, "AgentTool")
			data, err = packager.ExportPackage(pkg)
		}
		if err != nil {
			return "", fmt.Errorf("export failed: %w", err)
		}
		dbTx.Commit() //nolint:errcheck

		// Ensure output directory exists
		if dir := filepath.Dir(outputPath); dir != "" {
			if mkErr := os.MkdirAll(dir, 0o750); mkErr != nil {
				return "", fmt.Errorf("failed to create output directory: %w", mkErr)
			}
		}

		if err := os.WriteFile(outputPath, data, 0o640); err != nil {
			return "", fmt.Errorf("failed to write file: %w", err)
		}

		out := map[string]interface{}{
			"output_path":  outputPath,
			"file_size":    len(data),
			"package_name": pkg.Name,
			"version":      pkg.Version,
			"message":      fmt.Sprintf("Package exported to %s (%d bytes)", outputPath, len(data)),
		}
		b, _ := json.Marshal(out)
		return string(b), nil
	})
}

// ── package_import_file ────────────────────────────────────────────────────────

func (s *AgentToolRegistryService) registerPackageImportFileTool() {
	s.register(ToolDefinition{
		Type: "function",
		Function: ToolFunctionDef{
			Name:        "package_import_file",
			Description: "Import a package from a local JSON file (previously exported with package_export). Saves the package to the IAC database. Returns the imported package ID.",
			Parameters: ToolParameterSchema{
				Type: "object",
				Properties: map[string]ToolPropertySchema{
					"file_path":   {Type: "string", Description: "Path to the package JSON file to import"},
					"environment": {Type: "string", Description: "Override the environment for the imported package (optional)"},
				},
				Required: []string{"file_path"},
			},
		},
	}, func(ctx context.Context, args map[string]interface{}) (string, error) {
		filePath := stringArg(args, "file_path")
		environment := stringArg(args, "environment")

		if filePath == "" {
			return "", fmt.Errorf("file_path is required")
		}
		if err := validateFSPath(filePath); err != nil {
			return "", err
		}
		if s.sqlDB == nil {
			return "", fmt.Errorf("database not available")
		}

		data, err := os.ReadFile(filePath)
		if err != nil {
			return "", fmt.Errorf("failed to read file: %w", err)
		}

		var pkg depmodels.Package
		if err := json.Unmarshal(data, &pkg); err != nil {
			return "", fmt.Errorf("invalid package file format: %w", err)
		}

		dbTx, err := s.sqlDB.BeginTx(ctx, nil)
		if err != nil {
			return "", fmt.Errorf("failed to begin transaction: %w", err)
		}
		defer dbTx.Rollback() //nolint:errcheck

		repo := deprepo.NewPackageRepository("AgentTool", dbTx)
		record, err := repo.SavePackage(&pkg, environment, nil)
		if err != nil {
			return "", fmt.Errorf("failed to import package: %w", err)
		}
		dbTx.Commit() //nolint:errcheck

		out := map[string]interface{}{
			"package_id":   record.ID,
			"name":         record.Name,
			"version":      record.Version,
			"package_type": record.PackageType,
			"environment":  record.Environment,
			"message":      fmt.Sprintf("Package '%s v%s' imported successfully (ID: %s)", record.Name, record.Version, record.ID),
		}
		b, _ := json.Marshal(out)
		return string(b), nil
	})
}

// ── package_deploy ─────────────────────────────────────────────────────────────

func (s *AgentToolRegistryService) registerPackageDeployTool() {
	s.register(ToolDefinition{
		Type: "function",
		Function: ToolFunctionDef{
			Name: "package_deploy",
			Description: "Deploy a package to the connected database/MongoDB environment. " +
				"For database packages: inserts/updates SQL rows. For document packages: upserts MongoDB documents. " +
				"Use dry_run=true to preview what would happen without making changes.",
			Parameters: ToolParameterSchema{
				Type: "object",
				Properties: map[string]ToolPropertySchema{
					"id":              {Type: "string", Description: "Package ID to deploy"},
					"environment":     {Type: "string", Description: "Target environment label (for audit logging)"},
					"update_existing": {Type: "string", Description: "true/false — update rows that already exist (default: true)"},
					"skip_existing":   {Type: "string", Description: "true/false — skip rows that already exist instead of updating (default: false)"},
					"dry_run":         {Type: "string", Description: "true/false — preview deployment without writing any data (default: false)"},
					"continue_on_error": {Type: "string", Description: "true/false — continue deployment if individual row errors occur (default: false)"},
					"batch_size":      {Type: "string", Description: "Number of rows per batch (default: 100)"},
				},
				Required: []string{"id", "environment"},
			},
		},
	}, func(ctx context.Context, args map[string]interface{}) (string, error) {
		id := stringArg(args, "id")
		environment := stringArg(args, "environment")

		if id == "" || environment == "" {
			return "", fmt.Errorf("id and environment are required")
		}
		if s.sqlDB == nil {
			return "", fmt.Errorf("database not available")
		}

		options := depmodels.DeploymentOptions{
			UpdateExisting:  boolArgDefault(args, "update_existing", true),
			SkipExisting:    boolArgDefault(args, "skip_existing", false),
			DryRun:          boolArgDefault(args, "dry_run", false),
			ContinueOnError: boolArgDefault(args, "continue_on_error", false),
			BatchSize:       intArgDefault(args, "batch_size", 100),
		}

		// Load package from SQL
		loadTx, err := s.sqlDB.BeginTx(ctx, nil)
		if err != nil {
			return "", fmt.Errorf("failed to begin load transaction: %w", err)
		}
		defer loadTx.Rollback() //nolint:errcheck

		repo := deprepo.NewPackageRepository("AgentTool", loadTx)
		pkg, err := repo.GetPackage(id)
		if err != nil {
			return "", fmt.Errorf("package not found: %w", err)
		}
		loadTx.Commit() //nolint:errcheck

		// Deploy
		deployTx, err := s.sqlDB.BeginTx(ctx, nil)
		if err != nil {
			return "", fmt.Errorf("failed to begin deploy transaction: %w", err)
		}
		defer deployTx.Rollback() //nolint:errcheck

		var deployResult *depmodels.DeploymentRecord

		if pkg.PackageType == "database" {
			deployer := deploymgr.NewDatabaseDeployer("AgentTool", deployTx, dbconn.DatabaseType)
			deployResult, err = deployer.Deploy(pkg, options)
		} else if pkg.PackageType == "document" {
			if documents.DocDBCon == nil {
				return "", fmt.Errorf("document database not available")
			}
			deployer := deploymgr.NewDocumentDeployer(documents.DocDBCon, "AgentTool")
			deployResult, err = deployer.Deploy(pkg, options)
		} else {
			return "", fmt.Errorf("unknown package type: %s", pkg.PackageType)
		}

		if err != nil {
			return "", fmt.Errorf("deployment failed: %w", err)
		}

		if !options.DryRun {
			deployTx.Commit() //nolint:errcheck
		}

		// Save action log (non-fatal)
		if !options.DryRun {
			if actionTx, txErr := s.sqlDB.BeginTx(ctx, nil); txErr == nil {
				now := time.Now()
				actionRepo := deprepo.NewPackageRepository("AgentTool", actionTx)
				_ = actionRepo.SaveAction(&deprepo.PackageActionRecord{
					PackageID:         id,
					ActionType:        deprepo.ActionTypeDeploy,
					ActionStatus:      deprepo.ActionStatusCompleted,
					TargetEnvironment: environment,
					PerformedBy:       "AgentTool",
					PerformedAt:       now,
					StartedAt:         &now,
					CompletedAt:       &now,
				})
				actionTx.Commit() //nolint:errcheck
			}
		}

		out := map[string]interface{}{
			"package_id":    id,
			"package_name":  pkg.Name,
			"version":       pkg.Version,
			"environment":   environment,
			"dry_run":       options.DryRun,
			"deploy_result": deployResult,
			"message": fmt.Sprintf("Package '%s v%s' %s to environment '%s'",
				pkg.Name, pkg.Version, ternaryStr(options.DryRun, "dry-run completed", "deployed"), environment),
		}
		b, _ := json.Marshal(out)
		return string(b), nil
	})
}

// ── package_push ───────────────────────────────────────────────────────────────

func (s *AgentToolRegistryService) registerPackagePushTool() {
	s.register(ToolDefinition{
		Type: "function",
		Function: ToolFunctionDef{
			Name: "package_push",
			Description: "Push a package to another IAC instance by sending it to the remote /api/packages/import endpoint. " +
				"The remote IAC instance must be reachable. Returns success or error from the remote instance. " +
				"Real-time status is broadcast via SignalR topic 'package.push.status.{id}'.",
			Parameters: ToolParameterSchema{
				Type: "object",
				Properties: map[string]ToolPropertySchema{
					"id":          {Type: "string", Description: "Package ID to push"},
					"target_url":  {Type: "string", Description: "Base URL of the target IAC instance (e.g. https://iac-prod.example.com)"},
					"api_key":     {Type: "string", Description: "Bearer token for target IAC authentication (optional)"},
					"environment": {Type: "string", Description: "Override environment on import (optional, inherits from package record if blank)"},
				},
				Required: []string{"id", "target_url"},
			},
		},
	}, func(ctx context.Context, args map[string]interface{}) (string, error) {
		id := stringArg(args, "id")
		targetURL := strings.TrimRight(stringArg(args, "target_url"), "/")
		apiKey := stringArg(args, "api_key")
		environment := stringArg(args, "environment")

		if id == "" || targetURL == "" {
			return "", fmt.Errorf("id and target_url are required")
		}
		if s.sqlDB == nil {
			return "", fmt.Errorf("database not available")
		}

		deployToolBroadcastPush(id, "starting", targetURL, "")

		// Load package record (for environment) and package data from SQL
		dbTx, err := s.sqlDB.BeginTx(ctx, nil)
		if err != nil {
			return "", fmt.Errorf("failed to begin transaction: %w", err)
		}
		defer dbTx.Rollback() //nolint:errcheck

		repo := deprepo.NewPackageRepository("AgentTool", dbTx)

		// Get record for environment field
		records, err := repo.ListPackages("", "", "", 1, 0)
		_ = records // list used only to confirm table exists

		pkg, err := repo.GetPackage(id)
		if err != nil {
			deployToolBroadcastPush(id, "error", targetURL, "Package not found: "+err.Error())
			return "", fmt.Errorf("package not found: %w", err)
		}
		dbTx.Commit() //nolint:errcheck

		deployToolBroadcastPush(id, "pushing", targetURL, "")

		// Serialize package for sending
		pkgBytes, err := json.Marshal(pkg)
		if err != nil {
			deployToolBroadcastPush(id, "error", targetURL, "Serialization failed: "+err.Error())
			return "", fmt.Errorf("failed to serialize package: %w", err)
		}

		payload := map[string]interface{}{
			"package_data": json.RawMessage(pkgBytes),
			"environment":  environment,
		}
		payloadBytes, _ := json.Marshal(payload)

		// POST to target
		importURL := targetURL + "/api/packages/import"
		httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, importURL, bytes.NewReader(payloadBytes))
		if err != nil {
			deployToolBroadcastPush(id, "error", targetURL, "Failed to build request: "+err.Error())
			return "", fmt.Errorf("failed to build request: %w", err)
		}
		httpReq.Header.Set("Content-Type", "application/json")
		if apiKey != "" {
			httpReq.Header.Set("Authorization", "Bearer "+apiKey)
		}

		client := &http.Client{Timeout: 60 * time.Second}
		resp, err := client.Do(httpReq)
		if err != nil {
			deployToolBroadcastPush(id, "error", targetURL, "Failed to reach target: "+err.Error())
			return "", fmt.Errorf("failed to reach target %s: %w", targetURL, err)
		}
		defer resp.Body.Close()

		var result map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&result) //nolint:errcheck

		if resp.StatusCode >= 300 {
			errMsg := "Target IAC returned error"
			if m, ok := result["error"].(string); ok {
				errMsg = m
			}
			deployToolBroadcastPush(id, "error", targetURL, errMsg)
			return "", fmt.Errorf("push failed (HTTP %d): %s", resp.StatusCode, errMsg)
		}

		deployToolBroadcastPush(id, "success", targetURL, "")

		out := map[string]interface{}{
			"package_id":    id,
			"package_name":  pkg.Name,
			"version":       pkg.Version,
			"target_url":    targetURL,
			"target_result": result,
			"message":       fmt.Sprintf("Package '%s v%s' pushed to %s successfully", pkg.Name, pkg.Version, targetURL),
		}
		b, _ := json.Marshal(out)
		return string(b), nil
	})
}

// deployToolBroadcastPush broadcasts a package push status event via SignalR.
func deployToolBroadcastPush(packageID, status, targetURL, errMsg string) {
	srv := signalrserver.GetGlobalSignalRServer()
	if srv == nil || !srv.IsRunning() {
		return
	}
	payload := map[string]string{"package_id": packageID, "status": status}
	if targetURL != "" {
		payload["target_url"] = targetURL
	}
	if errMsg != "" {
		payload["error"] = errMsg
	}
	data, _ := json.Marshal(payload)
	srv.BroadcastToHub("package.push.status."+packageID, string(data))
}

// ── definition_list ────────────────────────────────────────────────────────────

func (s *AgentToolRegistryService) registerDefinitionListTool() {
	s.register(ToolDefinition{
		Type: "function",
		Function: ToolFunctionDef{
			Name:        "definition_list",
			Description: "List all active package definitions (reusable templates for creating packages). Returns name, type, environment, version, and entity counts.",
			Parameters: ToolParameterSchema{
				Type:       "object",
				Properties: map[string]ToolPropertySchema{},
			},
		},
	}, func(ctx context.Context, args map[string]interface{}) (string, error) {
		if documents.DocDBCon == nil || documents.DocDBCon.MongoDBDatabase == nil {
			return "", fmt.Errorf("document database not available")
		}

		dbCtx, cancel := context.WithTimeout(ctx, 15*time.Second)
		defer cancel()

		coll := documents.DocDBCon.MongoDBDatabase.Collection("package_definitions")
		cursor, err := coll.Find(dbCtx, bson.M{"active": true})
		if err != nil {
			return "", fmt.Errorf("failed to list definitions: %w", err)
		}
		defer cursor.Close(dbCtx)

		type defSummary struct {
			ID          string `json:"id" bson:"_id"`
			Name        string `json:"name" bson:"name"`
			PackageType string `json:"package_type" bson:"package_type"`
			Environment string `json:"environment" bson:"environment"`
			Version     string `json:"version" bson:"version"`
			Description string `json:"description" bson:"description"`
		}

		defs := make([]defSummary, 0)
		for cursor.Next(dbCtx) {
			var d defSummary
			if err := cursor.Decode(&d); err == nil {
				defs = append(defs, d)
			}
		}

		out := map[string]interface{}{"count": len(defs), "definitions": defs}
		b, _ := json.Marshal(out)
		return string(b), nil
	})
}

// ── definition_pack ────────────────────────────────────────────────────────────

func (s *AgentToolRegistryService) registerDefinitionPackTool() {
	s.register(ToolDefinition{
		Type: "function",
		Function: ToolFunctionDef{
			Name: "definition_pack",
			Description: "Trigger a packaging run from a saved package definition. " +
				"The definition specifies which tables/collections to include, filters, and global key fields. " +
				"Returns the new package ID and build number. Use definition_list to find definition IDs.",
			Parameters: ToolParameterSchema{
				Type: "object",
				Properties: map[string]ToolPropertySchema{
					"definition_id": {Type: "string", Description: "Package definition ID (UUID)"},
					"version":       {Type: "string", Description: "Version string override. Leave blank to use the definition version (with auto-suffix if auto_version is enabled)."},
				},
				Required: []string{"definition_id"},
			},
		},
	}, func(ctx context.Context, args map[string]interface{}) (string, error) {
		definitionID := stringArg(args, "definition_id")
		versionOverride := stringArg(args, "version")

		if definitionID == "" {
			return "", fmt.Errorf("definition_id is required")
		}
		if s.sqlDB == nil {
			return "", fmt.Errorf("database not available")
		}
		if documents.DocDBCon == nil || documents.DocDBCon.MongoDBDatabase == nil {
			return "", fmt.Errorf("document database not available")
		}

		// Load definition from MongoDB
		dbCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
		defer cancel()

		coll := documents.DocDBCon.MongoDBDatabase.Collection("package_definitions")
		var def depmodels.PackageDefinitionDoc
		if err := coll.FindOne(dbCtx, bson.M{"_id": definitionID}).Decode(&def); err != nil {
			return "", fmt.Errorf("definition not found: %w", err)
		}

		// Determine version
		version := def.Version
		if versionOverride != "" {
			version = versionOverride
		} else if def.AutoVersion {
			version = def.Version + "-" + time.Now().Format("20060102150405")
		}

		// Get next build number
		buildNumber := 1
		if buildTx, txErr := s.sqlDB.BeginTx(ctx, nil); txErr == nil {
			sqlRepo := deprepo.NewPackageDefinitionRepository("AgentTool", buildTx)
			if bn, err := sqlRepo.GetNextBuildNumber(definitionID); err == nil {
				buildNumber = bn
			}
			buildTx.Rollback() //nolint:errcheck
		}

		// Pack
		packTx, err := s.sqlDB.BeginTx(ctx, nil)
		if err != nil {
			return "", fmt.Errorf("failed to begin transaction: %w", err)
		}
		defer packTx.Rollback() //nolint:errcheck

		startTime := time.Now()
		var packageID, packageName, packageVersion string

		if def.PackageType == "database" {
			packager := packagemgr.NewDatabasePackager("AgentTool", packTx, dbconn.DatabaseType)
			p, err := packager.PackageTables(def.Name, version, "AgentTool", def.Filter)
			if err != nil {
				return "", fmt.Errorf("packaging failed: %w", err)
			}
			repo := deprepo.NewPackageRepository("AgentTool", packTx)
			record, err := repo.SavePackage(p, def.Environment, &def.Filter)
			if err != nil {
				return "", fmt.Errorf("failed to save package: %w", err)
			}
			packageID = record.ID
			packageName = record.Name
			packageVersion = record.Version
		} else {
			packager := packagemgr.NewDocumentPackager(documents.DocDBCon, "AgentTool")
			p, err := packager.PackageCollections(def.Name, version, "AgentTool", def.Filter)
			if err != nil {
				return "", fmt.Errorf("packaging failed: %w", err)
			}
			repo := deprepo.NewPackageRepository("AgentTool", packTx)
			record, err := repo.SavePackage(p, def.Environment, &def.Filter)
			if err != nil {
				return "", fmt.Errorf("failed to save package: %w", err)
			}
			packageID = record.ID
			packageName = record.Name
			packageVersion = record.Version
		}

		packTx.Commit() //nolint:errcheck

		// Save build record (non-fatal)
		if buildTx2, txErr2 := s.sqlDB.BeginTx(ctx, nil); txErr2 == nil {
			buildRepo := deprepo.NewPackageDefinitionRepository("AgentTool", buildTx2)
			now := time.Now()
			elapsed := int(now.Sub(startTime).Seconds())
			buildRec := &deprepo.PackageBuildRecord{
				DefinitionID:    definitionID,
				DefinitionName:  def.Name,
				DefVersion:      def.VersionNumber,
				BuildNumber:     buildNumber,
				PackageID:       packageID,
				PackageName:     packageName,
				PackageVersion:  packageVersion,
				Status:          "completed",
				StartedAt:       &startTime,
				CompletedAt:     &now,
				DurationSeconds: elapsed,
				TriggeredBy:     "AgentTool",
				Active:          true,
				CreatedBy:       "AgentTool",
			}
			if err := buildRepo.SaveBuild(buildRec); err != nil {
				s.iLog.Warn(fmt.Sprintf("Failed to save build record: %v", err))
				buildTx2.Rollback() //nolint:errcheck
			} else {
				buildTx2.Commit() //nolint:errcheck
			}
		}

		out := map[string]interface{}{
			"package_id":      packageID,
			"package_name":    packageName,
			"package_version": packageVersion,
			"build_number":    buildNumber,
			"definition_id":   definitionID,
			"definition_name": def.Name,
			"message": fmt.Sprintf("Build #%d created for definition '%s' → package '%s v%s' (ID: %s)",
				buildNumber, def.Name, packageName, packageVersion, packageID),
		}
		b, _ := json.Marshal(out)
		return string(b), nil
	})
}

// ── helpers ───────────────────────────────────────────────────────────────────

func intArgDefault(args map[string]interface{}, key string, def int) int {
	v := args[key]
	if v == nil {
		return def
	}
	switch t := v.(type) {
	case float64:
		return int(t)
	case int:
		return t
	case string:
		var n int
		if _, err := fmt.Sscanf(t, "%d", &n); err == nil {
			return n
		}
	}
	return def
}

func boolArgDefault(args map[string]interface{}, key string, def bool) bool {
	v := args[key]
	if v == nil {
		return def
	}
	switch t := v.(type) {
	case bool:
		return t
	case string:
		return t == "true" || t == "1" || t == "yes"
	case float64:
		return t != 0
	}
	return def
}

func ternaryStr(cond bool, a, b string) string {
	if cond {
		return a
	}
	return b
}
