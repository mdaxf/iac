# IAC Package and Deployment System

Comprehensive solution for packaging, deploying, and version controlling database data and document objects across different database systems.

## Features

### 1. Database Package System
- **Multi-Database Support**: MySQL, PostgreSQL, MSSQL, Oracle
- **Automatic PK Rebuilding**: Intelligently handles primary keys during transfer
- **Relationship Tracking**: Automatically packages and restores foreign key relationships
- **Parent Data Auto-Include**: Option to automatically include parent records
- **Flexible Filtering**: WHERE clauses, column exclusions, depth limits

### 2. Document Package System
- **MongoDB Support**: Package and deploy MongoDB collections
- **ID Handling**: Skip, preserve, or regenerate ObjectIDs
- **Reference Tracking**: Track and rebuild document references
- **Index Management**: Preserve and rebuild indexes
- **DevOps Object Support**: Special handling for TranCode, Workflow, UI_Page, etc.

### 3. Deployment System
- **PK Mapping**: Automatic mapping of old PKs to new PKs
- **Relationship Restoration**: Rebuild all foreign key relationships
- **Batch Processing**: Configurable batch sizes for large datasets
- **Conflict Resolution**: Skip or update existing records
- **Dry Run Mode**: Validate packages before deployment
- **Rollback Support**: Ability to rollback failed deployments

### 4. DevOps Version Control
- **Git-like Versioning**: Commit, branch, merge, tag operations
- **Change Tracking**: Detailed changelog with field-level changes
- **Branch Management**: Create and manage development branches
- **Diff Support**: Compare versions and see changes
- **Auto-commit**: Automatically version changes to objects
- **Revert Support**: Rollback to previous versions

## Architecture

```
deployment/
├── models/          # Data models and structures
│   └── package.go   # Package, deployment, and filter models
├── package/         # Packaging managers
│   ├── database_packager.go   # Database packaging
│   └── document_packager.go   # Document packaging
├── deploy/          # Deployment managers
│   ├── database_deployer.go   # Database deployment
│   └── document_deployer.go   # Document deployment
└── devops/          # Version control system
    └── version_control.go     # Git-like version control
```

## Usage

### Database Packaging

```go
import (
    dbconn "github.com/mdaxf/iac/databases"
    "github.com/mdaxf/iac/deployment/models"
    packagemgr "github.com/mdaxf/iac/deployment/package"
)

// Create packager
packager := packagemgr.NewDatabasePackager(user, dbTx, "mysql")

// Define filter
filter := models.PackageFilter{
    Tables: []string{"users", "orders", "products"},
    WhereClause: map[string]string{
        "users": "created_at > '2024-01-01'",
    },
    IncludeRelated: true,
    MaxDepth: 2,
    ExcludeColumns: map[string][]string{
        "users": {"password_hash"},
    },
}

// Package tables
pkg, err := packager.PackageTables("MyPackage", "1.0.0", "admin", filter)
if err != nil {
    log.Fatal(err)
}

// Export to JSON
data, err := packager.ExportPackage(pkg)
if err != nil {
    log.Fatal(err)
}

// Save to file
ioutil.WriteFile("package.json", data, 0644)
```

### Database Deployment

```go
import (
    deploymgr "github.com/mdaxf/iac/deployment/deploy"
    "github.com/mdaxf/iac/deployment/models"
)

// Import package
data, _ := ioutil.ReadFile("package.json")
pkg, err := packager.ImportPackage(data)

// Create deployer
deployer := deploymgr.NewDatabaseDeployer(user, dbTx, "postgresql")

// Configure deployment
options := models.DeploymentOptions{
    SkipExisting: false,
    UpdateExisting: true,
    ValidateReferences: true,
    CreateMissing: true,
    RebuildIndexes: true,
    BatchSize: 100,
    ContinueOnError: false,
    DryRun: false,
}

// Deploy
record, err := deployer.Deploy(pkg, options)
if err != nil {
    log.Fatal(err)
}

fmt.Printf("Deployed: %s (Status: %s)\n", record.ID, record.Status)

// Access PK mappings
for table, mappings := range record.PKMappingResult {
    fmt.Printf("Table %s: %d PKs mapped\n", table, len(mappings))
}
```

### Document Packaging

```go
import (
    "github.com/mdaxf/iac/documents"
    packagemgr "github.com/mdaxf/iac/deployment/package"
)

// Create packager
packager := packagemgr.NewDocumentPackager(docDB, user)

// Define filter
filter := models.PackageFilter{
    Collections: []string{"TranCode", "Workflow", "UI_Page"},
    WhereClause: map[string]string{
        "TranCode": `{"status": "active"}`,
    },
    ExcludeFields: map[string][]string{
        "TranCode": {"_internal_cache"},
    },
}

// Package collections
pkg, err := packager.PackageCollections("UIComponents", "2.0.0", "developer", filter)

// Export
data, _ := packager.ExportPackage(pkg)
```

### Document Deployment

```go
import deploymgr "github.com/mdaxf/iac/deployment/deploy"

// Create deployer
deployer := deploymgr.NewDocumentDeployer(docDB, user)

// Deploy with options
options := models.DeploymentOptions{
    UpdateExisting: true,
    RebuildIndexes: true,
    BatchSize: 50,
}

record, err := deployer.Deploy(pkg, options)

// Rollback if needed
if record.Status == "failed" {
    deployer.Rollback(record)
}
```

### Version Control

```go
import "github.com/mdaxf/iac/deployment/devops"

// Create version control
vc := devops.NewVersionControl(docDB, user)

// Commit changes
version, err := vc.Commit(
    devops.ObjectTypeTranCode,
    "trancode123",
    documentData,
    "Updated validation logic",
    "main",
    "developer",
)

// Create branch
branch, err := vc.CreateBranch(
    devops.ObjectTypeTranCode,
    "trancode123",
    "feature/new-validation",
    version.ID,
    "developer",
)

// Make changes on branch
newVersion, err := vc.Commit(
    devops.ObjectTypeTranCode,
    "trancode123",
    updatedData,
    "Added new validation rules",
    "feature/new-validation",
    "developer",
)

// Merge branch
merged, err := vc.MergeBranch(
    devops.ObjectTypeTranCode,
    "trancode123",
    "feature/new-validation",
    "main",
    "developer",
)

// Tag version
vc.TagVersion(merged.ID, []string{"v1.0.0", "production"})

// View changelog
changelog, err := vc.GetChangelog(devops.ObjectTypeTranCode, "trancode123", 10)
for _, entry := range changelog {
    fmt.Printf("%s: %s (%d changes)\n",
        entry.ChangedAt, entry.CommitMessage, len(entry.Changes))
}

// Diff versions
changes, err := vc.Diff(version1.ID, version2.ID)
for _, change := range changes {
    fmt.Printf("%s: %s (%v -> %v)\n",
        change.Action, change.Field, change.OldValue, change.NewValue)
}

// Revert to previous version
reverted, err := vc.Revert(
    devops.ObjectTypeTranCode,
    "trancode123",
    version.ID,
    "developer",
)
```

### DevOps Object Packaging

```go
// Package specific DevOps objects
pkg, err := packager.PackageDevOpsObjects(
    "TranCode",
    []string{"tc_001", "tc_002", "tc_003"},
    "TranCodeRelease",
    "1.5.0",
    "developer",
)
```

## API Endpoints

### Package Database
```http
POST /api/deployment/package/database
Content-Type: application/json

{
  "name": "MyPackage",
  "version": "1.0.0",
  "description": "Package description",
  "filter": {
    "tables": ["users", "orders"],
    "include_related": true,
    "max_depth": 2
  }
}
```

### Package Documents
```http
POST /api/deployment/package/documents
Content-Type: application/json

{
  "name": "UIPackage",
  "version": "1.0.0",
  "filter": {
    "collections": ["TranCode", "Workflow"]
  }
}
```

### Deploy Database Package
```http
POST /api/deployment/deploy/database
Content-Type: application/json

{
  "package_data": <base64_encoded_package>,
  "options": {
    "update_existing": true,
    "batch_size": 100
  }
}
```

### Deploy Document Package
```http
POST /api/deployment/deploy/documents
Content-Type: application/json

{
  "package_data": <base64_encoded_package>,
  "options": {
    "update_existing": true,
    "rebuild_indexes": true
  }
}
```

### Version Control Operations
```http
POST /api/deployment/version-control
Content-Type: application/json

{
  "action": "commit",
  "object_type": "TranCode",
  "object_id": "tc_001",
  "content": { ... },
  "commit_message": "Updated logic",
  "branch": "main"
}
```

## PK Strategies

The system supports multiple PK generation strategies:

1. **auto_increment**: Database auto-generates PK (MySQL AUTO_INCREMENT, PostgreSQL SERIAL)
2. **sequence**: Uses database sequences (PostgreSQL, Oracle)
3. **uuid**: Generates new UUID for each record
4. **preserve**: Keeps original PK values (use with caution)

The packager automatically detects the best strategy based on column types.

## ID Strategies for Documents

1. **regenerate**: Generate new ObjectIDs (recommended)
2. **skip**: Remove _id field, let MongoDB generate
3. **preserve**: Keep original ObjectIDs (may cause conflicts)

## Relationship Handling

### Database
- Foreign key relationships are automatically detected
- PKs are mapped during deployment
- FKs are updated to reference new PKs
- Referential integrity is maintained

### Documents
- Reference patterns are defined per collection type
- Single and array references are supported
- References are updated after all documents are inserted

## MongoDB Collections Created

### Object_Versions
Stores all versions of objects:
```javascript
{
  _id: "version_id",
  object_id: "object_id",
  object_type: "TranCode",
  version: "v5",
  version_number: 5,
  content: { ... },
  content_hash: "sha256_hash",
  created_at: ISODate(...),
  created_by: "user",
  commit_message: "Change description",
  branch: "main",
  parent_version: "parent_version_id"
}
```

### Object_Branches
Stores branch information:
```javascript
{
  _id: "branch_id",
  name: "feature/new-feature",
  object_id: "object_id",
  object_type: "TranCode",
  base_version: "version_id",
  head_version: "version_id",
  is_active: true,
  is_merged: false
}
```

### Object_Changelogs
Stores detailed change history:
```javascript
{
  _id: "changelog_id",
  object_id: "object_id",
  action: "update",
  from_version: "v4",
  to_version: "v5",
  changes: [
    {
      field: "validation_rules",
      old_value: [...],
      new_value: [...],
      action: "modify"
    }
  ],
  changed_by: "user",
  changed_at: ISODate(...)
}
```

## Best Practices

1. **Testing**: Always use dry run mode first
2. **Backups**: Create backups before deployment
3. **Filtering**: Use WHERE clauses to limit package size
4. **Batching**: Use appropriate batch sizes for performance
5. **Version Control**: Commit changes regularly with meaningful messages
6. **Branching**: Use branches for development work
7. **Tagging**: Tag stable versions for easy reference
8. **Validation**: Enable reference validation in production

## Error Handling

All operations include comprehensive error handling:
- Transaction rollback on errors
- Detailed error logs in deployment records
- Continue-on-error option for batch operations
- Rollback support for failed deployments

## Performance

- Batch processing for large datasets
- Configurable transaction sizes
- Parallel processing support (future)
- Efficient PK mapping using hash maps
- Index management for optimal query performance

## Security

- Transaction-based operations
- Validation before deployment
- Audit trail in version control
- User tracking for all operations
- Sensitive data exclusion options

## Future Enhancements

- [ ] Parallel deployment workers
- [ ] Incremental package updates
- [ ] Compression for large packages
- [ ] Encryption for sensitive data
- [ ] Web UI for package management
- [ ] Automated testing framework
- [ ] Migration planning tools
- [ ] Performance analytics
- [ ] Advanced conflict resolution
- [ ] Multi-database deployment

## License

Copyright 2023 IAC. All Rights Reserved.
Licensed under the Apache License, Version 2.0.
