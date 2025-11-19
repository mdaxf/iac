# Package Deployment API Documentation

## Overview

The Package Deployment API provides comprehensive endpoints for managing, generating, deploying, and monitoring IAC packages. This includes both database and document packages with full support for background job processing.

## Base URL

```
http://<server>:<port>/api
```

## Authentication

All endpoints require authentication. Include authentication headers as per your IAC configuration.

---

## Endpoints

### 1. Package Management

#### 1.1 List All Packages

**Endpoint:** `GET /api/packages`

**Description:** Retrieve a list of all packages with optional filters.

**Query Parameters:**
- `package_type` (optional): Filter by type ("database" or "document")
- `environment` (optional): Filter by environment ("dev", "staging", "production")
- `status` (optional): Filter by status ("active", "archived", "deleted")
- `limit` (optional, default: 50): Maximum number of results
- `offset` (optional, default: 0): Pagination offset

**Example Request:**
```bash
curl -X GET "http://localhost:8080/api/packages?package_type=database&environment=production&limit=20"
```

**Example Response:**
```json
{
  "packages": [
    {
      "id": "pkg-123",
      "name": "customer-data",
      "version": "1.0.0",
      "packagetype": "database",
      "description": "Customer data package",
      "databasetype": "mysql",
      "checksum": "abc123...",
      "filesize": 1024000,
      "status": "active",
      "environment": "production",
      "active": true,
      "createdby": "admin",
      "createdon": "2025-11-18T10:00:00Z"
    }
  ],
  "count": 1,
  "limit": 20,
  "offset": 0
}
```

---

#### 1.2 Get Package Details

**Endpoint:** `GET /api/packages/{id}`

**Description:** Get detailed information about a specific package including recent actions.

**Path Parameters:**
- `id` (required): Package ID

**Example Request:**
```bash
curl -X GET "http://localhost:8080/api/packages/pkg-123"
```

**Example Response:**
```json
{
  "package": {
    "id": "pkg-123",
    "name": "customer-data",
    "version": "1.0.0",
    "package_type": "database",
    "description": "Customer data package",
    "created_at": "2025-11-18T10:00:00Z",
    "created_by": "admin",
    "metadata": {
      "total_records": 1500,
      "tables": ["customers", "orders"]
    },
    "database_data": {
      "tables": [...],
      "pk_mappings": {...}
    }
  },
  "recent_actions": [
    {
      "id": "action-456",
      "packageid": "pkg-123",
      "actiontype": "pack",
      "actionstatus": "completed",
      "performedat": "2025-11-18T10:00:00Z",
      "performedby": "admin",
      "recordsprocessed": 1500
    }
  ]
}
```

---

#### 1.3 Create Package

**Endpoint:** `POST /api/packages`

**Description:** Create a new package definition and package selected tables/collections.

**Request Body:**
```json
{
  "name": "customer-data",
  "version": "1.0.0",
  "description": "Customer data package",
  "package_type": "database",
  "environment": "production",
  "include_parent": false,
  "filter": {
    "tables": ["customers", "orders"],
    "where_clause": {
      "customers": "created_date >= '2025-01-01'",
      "orders": "order_date >= '2025-01-01'"
    },
    "exclude_columns": {
      "customers": ["password_hash", "internal_notes"]
    }
  },
  "metadata": {
    "department": "sales",
    "compliance": "gdpr"
  }
}
```

**Example Request:**
```bash
curl -X POST "http://localhost:8080/api/packages" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "customer-data",
    "version": "1.0.0",
    "package_type": "database",
    "filter": {
      "tables": ["customers", "orders"]
    }
  }'
```

**Example Response:**
```json
{
  "package": {
    "id": "pkg-123",
    "name": "customer-data",
    "version": "1.0.0",
    "packagetype": "database",
    "checksum": "abc123...",
    "filesize": 1024000,
    "status": "active"
  },
  "message": "Package created successfully"
}
```

---

#### 1.4 Update Package

**Endpoint:** `PUT /api/packages/{id}`

**Description:** Update package metadata (description, status, tags, etc.).

**Path Parameters:**
- `id` (required): Package ID

**Request Body:**
```json
{
  "description": "Updated customer data package",
  "status": "archived",
  "tags": ["production", "critical", "gdpr"],
  "metadata": {
    "last_reviewed": "2025-11-18",
    "reviewer": "admin"
  }
}
```

**Example Response:**
```json
{
  "message": "Package updated successfully",
  "package_id": "pkg-123"
}
```

---

#### 1.5 Delete Package (Soft Delete)

**Endpoint:** `DELETE /api/packages/{id}`

**Description:** Soft delete a package by setting active = false.

**Path Parameters:**
- `id` (required): Package ID

**Example Request:**
```bash
curl -X DELETE "http://localhost:8080/api/packages/pkg-123"
```

**Example Response:**
```json
{
  "message": "Package deleted successfully",
  "package_id": "pkg-123"
}
```

---

### 2. Package Generation & Export

#### 2.1 Generate Package File

**Endpoint:** `POST /api/packages/{id}/generate`

**Description:** Generate a downloadable package file in JSON or ZIP format.

**Path Parameters:**
- `id` (required): Package ID

**Request Body:**
```json
{
  "tables": ["customers", "orders"],
  "collections": [],
  "where_clause": {
    "customers": "active = TRUE"
  },
  "format": "zip"
}
```

**Example Request:**
```bash
curl -X POST "http://localhost:8080/api/packages/pkg-123/generate" \
  -H "Content-Type: application/json" \
  -d '{"format": "zip"}' \
  --output package.zip
```

**Response:** Binary file (ZIP or JSON)

**ZIP File Contents:**
- `customer-data-v1.0.0.json` - Package data
- `metadata.json` - Package metadata

---

### 3. Package Import & Deployment

#### 3.1 Import Package

**Endpoint:** `POST /api/packages/import`

**Description:** Import a package from JSON data and save it to the database.

**Request Body:**
```json
{
  "package_data": {
    "id": "pkg-789",
    "name": "imported-package",
    "version": "1.0.0",
    "package_type": "database",
    "database_data": {...}
  },
  "environment": "staging"
}
```

**Example Response:**
```json
{
  "message": "Package imported successfully",
  "package": {
    "id": "pkg-789",
    "name": "imported-package",
    "version": "1.0.0"
  }
}
```

---

#### 3.2 Deploy Package

**Endpoint:** `POST /api/packages/{id}/deploy`

**Description:** Deploy a package to the target environment. Supports immediate execution or background job scheduling.

**Path Parameters:**
- `id` (required): Package ID

**Request Body:**
```json
{
  "environment": "production",
  "run_as_background": false,
  "schedule_at": null,
  "options": {
    "skip_existing": true,
    "update_existing": false,
    "validate_references": true,
    "create_missing": false,
    "batch_size": 100,
    "transaction_size": 1000,
    "continue_on_error": false,
    "dry_run": false
  }
}
```

**Deployment Options:**
- `skip_existing`: Skip records that already exist
- `update_existing`: Update existing records
- `validate_references`: Validate foreign key references before deploy
- `create_missing`: Create missing parent records
- `batch_size`: Number of records per batch
- `transaction_size`: Records per transaction
- `continue_on_error`: Continue deployment on errors
- `dry_run`: Validate but don't deploy

**Example Request (Immediate Deployment):**
```bash
curl -X POST "http://localhost:8080/api/packages/pkg-123/deploy" \
  -H "Content-Type: application/json" \
  -d '{
    "environment": "production",
    "options": {
      "skip_existing": true,
      "batch_size": 100
    }
  }'
```

**Example Response (Immediate):**
```json
{
  "message": "Package deployed successfully",
  "deployment": {
    "id": "deploy-123",
    "package_id": "pkg-123",
    "status": "completed",
    "deployed_at": "2025-11-18T11:00:00Z",
    "pk_mapping_result": {
      "customers": {
        "1": "1001",
        "2": "1002"
      }
    }
  }
}
```

**Example Request (Background Job):**
```bash
curl -X POST "http://localhost:8080/api/packages/pkg-123/deploy" \
  -H "Content-Type: application/json" \
  -d '{
    "environment": "production",
    "run_as_background": true,
    "schedule_at": "2025-11-18T23:00:00Z",
    "options": {
      "skip_existing": true
    }
  }'
```

**Example Response (Background Job):**
```json
{
  "message": "Deployment job created",
  "job_id": "deploy-pkg123-1731945600",
  "status": "scheduled"
}
```

---

### 4. Package Monitoring

#### 4.1 Get Package Actions (History)

**Endpoint:** `GET /api/packages/{id}/actions`

**Description:** Get action history for a package (pack, deploy, rollback, etc.).

**Path Parameters:**
- `id` (required): Package ID

**Query Parameters:**
- `limit` (optional, default: 50): Maximum number of results

**Example Request:**
```bash
curl -X GET "http://localhost:8080/api/packages/pkg-123/actions?limit=10"
```

**Example Response:**
```json
{
  "actions": [
    {
      "id": "action-456",
      "packageid": "pkg-123",
      "actiontype": "deploy",
      "actionstatus": "completed",
      "targetenvironment": "production",
      "performedat": "2025-11-18T11:00:00Z",
      "performedby": "admin",
      "startedat": "2025-11-18T11:00:00Z",
      "completedat": "2025-11-18T11:05:00Z",
      "durationseconds": 300,
      "recordsprocessed": 1500,
      "recordssucceeded": 1500,
      "recordsfailed": 0,
      "tablesprocessed": 2
    },
    {
      "id": "action-455",
      "packageid": "pkg-123",
      "actiontype": "pack",
      "actionstatus": "completed",
      "sourceenvironment": "production",
      "performedat": "2025-11-18T10:00:00Z",
      "performedby": "admin",
      "recordsprocessed": 1500,
      "tablesprocessed": 2
    }
  ],
  "count": 2
}
```

---

### 5. Deployment Monitoring

#### 5.1 List All Deployments

**Endpoint:** `GET /api/deployments`

**Description:** Get a list of all package deployments with optional filters.

**Query Parameters:**
- `environment` (optional): Filter by environment
- `is_active` (optional): Filter by active status (true/false)
- `limit` (optional, default: 50): Maximum number of results
- `offset` (optional, default: 0): Pagination offset

**Example Request:**
```bash
curl -X GET "http://localhost:8080/api/deployments?environment=production&is_active=true"
```

**Example Response:**
```json
{
  "deployments": [
    {
      "id": "deploy-123",
      "packageid": "pkg-123",
      "actionid": "action-456",
      "environment": "production",
      "databasename": "iacdb",
      "deployedat": "2025-11-18T11:00:00Z",
      "deployedby": "admin",
      "isactive": true,
      "rolledbackat": null,
      "rolledbackby": null,
      "active": true,
      "createdby": "admin",
      "createdon": "2025-11-18T11:00:00Z"
    }
  ],
  "count": 1,
  "limit": 50,
  "offset": 0
}
```

---

#### 5.2 Get Deployment Details

**Endpoint:** `GET /api/deployments/{id}`

**Description:** Get detailed information about a specific deployment.

**Path Parameters:**
- `id` (required): Deployment ID

**Example Request:**
```bash
curl -X GET "http://localhost:8080/api/deployments/deploy-123"
```

**Example Response:**
```json
{
  "deployment": {
    "id": "deploy-123",
    "packageid": "pkg-123",
    "actionid": "action-456",
    "environment": "production",
    "databasename": "iacdb",
    "deployedat": "2025-11-18T11:00:00Z",
    "deployedby": "admin",
    "isactive": true,
    "active": true
  }
}
```

---

### 6. Background Job Management

#### 6.1 Get Job Status

**Endpoint:** `GET /api/jobs/{id}`

**Description:** Get status and details of a background deployment job.

**Path Parameters:**
- `id` (required): Job ID

**Example Request:**
```bash
curl -X GET "http://localhost:8080/api/jobs/deploy-pkg123-1731945600"
```

**Example Response:**
```json
{
  "job_id": "deploy-pkg123-1731945600",
  "job_type": "package_deployment",
  "status": "completed",
  "scheduled_at": "2025-11-18T23:00:00Z",
  "started_at": "2025-11-18T23:00:01Z",
  "completed_at": "2025-11-18T23:05:30Z",
  "job_data": {
    "package_id": "pkg-123",
    "environment": "production",
    "options": {
      "skip_existing": true
    },
    "user": "admin"
  }
}
```

**Job Statuses:**
- `pending`: Waiting to be queued
- `queued`: Queued for execution
- `processing`: Currently executing
- `completed`: Successfully completed
- `failed`: Failed with errors
- `cancelled`: Cancelled by user
- `timeout`: Execution timed out
- `retry`: Waiting to retry after failure

---

## Complete Workflow Examples

### Example 1: Create and Deploy a Database Package

```bash
# Step 1: Create a package
curl -X POST "http://localhost:8080/api/packages" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "customer-data",
    "version": "1.0.0",
    "package_type": "database",
    "environment": "dev",
    "filter": {
      "tables": ["customers", "orders"],
      "where_clause": {
        "customers": "created_date >= '\''2025-01-01'\''"
      }
    }
  }' | jq

# Response: {"package": {"id": "pkg-123", ...}}

# Step 2: Generate ZIP file
curl -X POST "http://localhost:8080/api/packages/pkg-123/generate" \
  -H "Content-Type: application/json" \
  -d '{"format": "zip"}' \
  --output customer-data-v1.0.0.zip

# Step 3: Deploy to staging
curl -X POST "http://localhost:8080/api/packages/pkg-123/deploy" \
  -H "Content-Type: application/json" \
  -d '{
    "environment": "staging",
    "options": {
      "skip_existing": true,
      "validate_references": true,
      "batch_size": 100
    }
  }' | jq

# Step 4: Check deployment status
curl -X GET "http://localhost:8080/api/packages/pkg-123/actions?limit=1" | jq
```

### Example 2: Schedule a Background Deployment

```bash
# Step 1: Schedule deployment for midnight
curl -X POST "http://localhost:8080/api/packages/pkg-123/deploy" \
  -H "Content-Type: application/json" \
  -d '{
    "environment": "production",
    "run_as_background": true,
    "schedule_at": "2025-11-19T00:00:00Z",
    "options": {
      "skip_existing": true,
      "batch_size": 100
    }
  }' | jq

# Response: {"job_id": "deploy-pkg123-1732003200", "status": "scheduled"}

# Step 2: Monitor job status
curl -X GET "http://localhost:8080/api/jobs/deploy-pkg123-1732003200" | jq

# Step 3: Check deployment after completion
curl -X GET "http://localhost:8080/api/deployments?environment=production&is_active=true" | jq
```

### Example 3: Import and Deploy Package from Another System

```bash
# Step 1: Export package from source system
curl -X POST "http://source-server/api/packages/pkg-456/generate" \
  -H "Content-Type: application/json" \
  -d '{"format": "json"}' \
  --output package.json

# Step 2: Import package to target system
curl -X POST "http://target-server/api/packages/import" \
  -H "Content-Type: application/json" \
  -d "{\"package_data\": $(cat package.json), \"environment\": \"staging\"}" | jq

# Response: {"package": {"id": "pkg-789", ...}}

# Step 3: Deploy imported package
curl -X POST "http://target-server/api/packages/pkg-789/deploy" \
  -H "Content-Type: application/json" \
  -d '{
    "environment": "staging",
    "options": {
      "skip_existing": false,
      "update_existing": true,
      "validate_references": true
    }
  }' | jq
```

---

## Error Responses

All endpoints return standard error responses:

```json
{
  "error": "Error message describing what went wrong"
}
```

**HTTP Status Codes:**
- `200 OK`: Successful request
- `201 Created`: Resource created successfully
- `202 Accepted`: Request accepted for background processing
- `400 Bad Request`: Invalid request data
- `404 Not Found`: Resource not found
- `500 Internal Server Error`: Server error

---

## Notes

1. **Background Jobs:** All deployment jobs are processed by the IAC framework's JobWorker. Jobs are stored in the `queue_jobs` table and executed by registered handlers:
   - `PACKAGE_DEPLOYMENT`: Handles package deployment jobs
   - `PACKAGE_GENERATION`: Handles package generation jobs

   The framework also provides standalone functions that can be called directly:
   - `handlers.DeployPackageJob()`: Deploy a package directly
   - `handlers.GeneratePackageJob()`: Generate a package file directly

2. **Transaction Safety:** Deployments use database transactions. Failed deployments will rollback automatically.

3. **Monitoring:** Use the `/actions` and `/deployments` endpoints to monitor package history and deployment status.

4. **ZIP Format:** ZIP files contain both package data and metadata for easy distribution.

5. **Filters:** Use WHERE clauses and column exclusions to create targeted packages with only the data you need.

6. **Dependencies:** Package dependencies are preserved during import/export and can be used to enforce deployment order.

---

## Database Schema

To use these endpoints, ensure the following tables are created:

1. `iacpackages` - Package definitions
2. `packageactions` - Action history
3. `packagerelationships` - Package dependencies
4. `packagedeployments` - Deployment records
5. `packagetags` - Package tags
6. `queue_jobs` - IAC framework job queue (already exists in framework)

Schema files are located in:
- `deployment/schema/packages_schema.sql` (MySQL)
- `deployment/schema/packages_schema_postgresql.sql` (PostgreSQL)

Note: The job queue system uses IAC's existing `queue_jobs` table from the framework. No additional job tables are needed.

---

## Job Handlers

The package deployment system provides the following job handlers that can be used by the IAC framework:

### Handler Functions (for trancode.ExecutebyExternal)

These handlers are designed to be called by the framework's job worker via `trancode.ExecutebyExternal()`:

1. **PACKAGE_DEPLOYMENT** - Handles package deployment jobs
   - Handler: `handlers.PackageDeploymentHandler()`
   - Input: `package_id`, `environment`, `options`, `user`
   - Returns: Deployment result with status, deployment_id, records_deployed

2. **PACKAGE_GENERATION** - Handles package generation jobs
   - Handler: `handlers.PackageGenerationHandler()`
   - Input: `package_id`, `format`, `user`
   - Returns: Generation result with package_data, size_bytes, action_id

### Standalone Functions (direct call)

These functions can be called directly without needing trancode.ExecutebyExternal:

1. **DeployPackageJob** - Deploy a package directly
   ```go
   func DeployPackageJob(packageID, environment string, options models.DeploymentOptions, userName string) (*models.DeploymentRecord, error)
   ```
   - Manages its own transaction
   - Returns deployment record with PK/ID mappings
   - Logs all actions to packageactions table

2. **GeneratePackageJob** - Generate package file directly
   ```go
   func GeneratePackageJob(packageID, format, userName string) ([]byte, map[string]interface{}, error)
   ```
   - Returns package data as bytes and metadata
   - Supports "json" format
   - Logs generation action to packageactions table

### Handler Registration

All handlers are registered in `deployment/handlers/register.go`:

```go
var HandlerRegistry = map[string]HandlerFunc{
    "PACKAGE_DEPLOYMENT": PackageDeploymentHandler,
    "PACKAGE_GENERATION": PackageGenerationHandler,
}
```

To use a handler:
```go
handler, exists := handlers.GetHandler("PACKAGE_DEPLOYMENT")
if exists {
    result, err := handler(inputs, tx, docDB)
}
```
