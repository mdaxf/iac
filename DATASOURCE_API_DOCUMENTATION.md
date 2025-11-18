# Report Datasource API Documentation

## Overview
Complete API reference for managing report datasources with request/response payload examples.

---

## 1. Create Datasource

**Endpoint:** `POST /api/reports/:id/datasources`
**Handler:** `AddDatasource`

### Request Payload

```json
{
  "alias": "default",
  "databasealias": "default",
  "querytype": "visual",
  "customsql": "",
  "selectedtables": [
    {"name": "users"},
    {"name": "orders"}
  ],
  "selectedfields": [
    {
      "table": "users",
      "field": "name",
      "alias": "user_name",
      "aggregation": "",
      "data_type": "string"
    },
    {
      "table": "orders",
      "field": "total",
      "alias": "order_total",
      "aggregation": "SUM",
      "data_type": "decimal"
    }
  ],
  "joins": [
    {
      "left_table": "orders",
      "right_table": "users",
      "left_field": "user_id",
      "right_field": "id",
      "join_type": "INNER"
    }
  ],
  "filters": [
    {
      "field": "users.status",
      "operator": "=",
      "value": "active"
    }
  ],
  "sorting": [
    {
      "field": "users.name",
      "direction": "ASC"
    }
  ],
  "grouping": [
    {
      "field": "users.id"
    }
  ],
  "parameters": {
    "start_date": {
      "type": "date",
      "default": "2025-01-01"
    }
  }
}
```

### Field Descriptions

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `alias` | string | Yes | Unique identifier for this datasource within the report |
| `databasealias` | string | Yes | Database connection alias (from database configuration) |
| `querytype` | string | Yes | Query type: `"visual"` or `"custom"` |
| `customsql` | string | No | Custom SQL query (when querytype is "custom") |
| `selectedtables` | array | No | List of tables to query (for visual queries) |
| `selectedfields` | array | No | Fields to select with optional aliases and aggregations |
| `joins` | array | No | Table join definitions |
| `filters` | array | No | WHERE clause filters |
| `sorting` | array | No | ORDER BY clause |
| `grouping` | array | No | GROUP BY clause |
| `parameters` | object | No | Query parameters with types and defaults |

### Response (201 Created)

```json
{
  "id": "aaf5611e-c071-4082-abe5-e43ef94099c2",
  "reportid": "4da18c13-9190-4a5b-90b2-840a8cf248dc",
  "alias": "default",
  "databasealias": "default",
  "querytype": "visual",
  "customsql": "",
  "selectedtables": [...],
  "selectedfields": [...],
  "joins": [...],
  "filters": [...],
  "sorting": [...],
  "grouping": [...],
  "parameters": {...},
  "active": true,
  "referenceid": "",
  "createdby": "john.doe",
  "createdon": "2025-11-18T22:30:00Z",
  "modifiedby": "john.doe",
  "modifiedon": "2025-11-18T22:30:00Z",
  "rowversionstamp": 1
}
```

---

## 2. Get Datasources

**Endpoint:** `GET /api/reports/:id/datasources`
**Handler:** `GetDatasources`

### Request
No request body required.

### URL Parameters
- `:id` - Report ID (UUID)

### Response (200 OK)

```json
[
  {
    "id": "aaf5611e-c071-4082-abe5-e43ef94099c2",
    "reportid": "4da18c13-9190-4a5b-90b2-840a8cf248dc",
    "alias": "default",
    "databasealias": "default",
    "querytype": "visual",
    "selectedtables": [...],
    "selectedfields": [...],
    "joins": [...],
    "filters": [],
    "sorting": [],
    "grouping": [],
    "parameters": null,
    "active": true,
    "createdby": "john.doe",
    "createdon": "2025-11-18T22:30:00Z",
    "modifiedby": "john.doe",
    "modifiedon": "2025-11-18T22:30:00Z",
    "rowversionstamp": 1
  }
]
```

---

## 3. Update Datasource (NEW)

**Endpoint:** `PUT /api/reports/:id/datasources/:datasourceId`
**Handler:** `UpdateDatasourceEndpoint`

### URL Parameters
- `:id` - Report ID (UUID)
- `:datasourceId` - Datasource ID (UUID)

### Request Payload (Partial Update Supported)

You can send only the fields you want to update:

**Example 1: Update query fields**
```json
{
  "selectedfields": [
    {
      "table": "users",
      "field": "email",
      "alias": "user_email",
      "aggregation": "",
      "data_type": "string"
    }
  ],
  "filters": [
    {
      "field": "users.created_at",
      "operator": ">",
      "value": "2025-01-01"
    }
  ]
}
```

**Example 2: Update tables and joins**
```json
{
  "selectedtables": [
    {"name": "lngcodes"},
    {"name": "lngcode_contents"}
  ],
  "joins": [
    {
      "left_table": "lngcode_contents",
      "right_table": "lngcodes",
      "left_field": "lngcodeid",
      "right_field": "id",
      "join_type": "INNER"
    }
  ]
}
```

**Example 3: Switch to custom SQL**
```json
{
  "querytype": "custom",
  "customsql": "SELECT u.name, COUNT(o.id) as order_count FROM users u LEFT JOIN orders o ON u.id = o.user_id GROUP BY u.id"
}
```

**Example 4: Update all fields**
```json
{
  "alias": "main_query",
  "databasealias": "analytics_db",
  "querytype": "visual",
  "customsql": "",
  "selectedtables": [
    {"name": "products"},
    {"name": "categories"}
  ],
  "selectedfields": [
    {
      "table": "products",
      "field": "name",
      "alias": "product_name",
      "aggregation": "",
      "data_type": "string"
    },
    {
      "table": "categories",
      "field": "name",
      "alias": "category_name",
      "aggregation": "",
      "data_type": "string"
    }
  ],
  "joins": [
    {
      "left_table": "products",
      "right_table": "categories",
      "left_field": "category_id",
      "right_field": "id",
      "join_type": "LEFT"
    }
  ],
  "filters": [
    {
      "field": "products.active",
      "operator": "=",
      "value": true
    }
  ],
  "sorting": [
    {
      "field": "products.name",
      "direction": "ASC"
    }
  ],
  "grouping": [],
  "parameters": null
}
```

### Response (200 OK)

```json
{
  "message": "Datasource updated successfully",
  "datasources": [
    {
      "id": "aaf5611e-c071-4082-abe5-e43ef94099c2",
      "reportid": "4da18c13-9190-4a5b-90b2-840a8cf248dc",
      "alias": "main_query",
      "databasealias": "analytics_db",
      "querytype": "visual",
      "selectedtables": [...],
      "selectedfields": [...],
      "joins": [...],
      "filters": [...],
      "sorting": [...],
      "grouping": [],
      "parameters": null,
      "modifiedby": "john.doe",
      "modifiedon": "2025-11-18T23:00:00Z",
      "rowversionstamp": 2
    }
  ]
}
```

### Notes
- **Partial updates supported**: Only include fields you want to change
- **JSON arrays can be empty**: Use `[]` for empty filters, sorting, grouping
- **Auto-serialization**: Complex JSON fields are automatically serialized
- **Ownership validation**: Only report owner can update datasources

---

## 4. Delete Datasource (NEW)

**Endpoint:** `DELETE /api/reports/:id/datasources/:datasourceId`
**Handler:** `DeleteDatasourceEndpoint`

### URL Parameters
- `:id` - Report ID (UUID)
- `:datasourceId` - Datasource ID (UUID)

### Request
No request body required.

### Response (200 OK)

```json
{
  "message": "Datasource deleted successfully"
}
```

### Notes
- **Hard delete**: Permanently removes the datasource from the database
- **Ownership validation**: Only report owner can delete datasources
- **Cascade**: If this is the only datasource, report may become invalid

---

## 5. Batch Update via Report Update (Legacy)

**Endpoint:** `PUT /api/reports/:id`
**Handler:** `UpdateReport`

You can still include datasources in report updates for batch operations:

### Request Payload

```json
{
  "name": "Updated Report Name",
  "description": "Updated description",
  "datasources": [
    {
      "id": "aaf5611e-c071-4082-abe5-e43ef94099c2",
      "alias": "default",
      "databasealias": "default",
      "querytype": "visual",
      "selectedtables": [...],
      "selectedfields": [...],
      "joins": [...],
      "filters": [],
      "sorting": [],
      "grouping": []
    }
  ]
}
```

### Notes
- **For existing datasources**: Include the `id` field to update
- **For new datasources**: Omit the `id` field to create
- **Auto-handling**: Controller automatically routes to create/update logic

---

## Common Data Structures

### SelectedTables
```json
[
  {"name": "users"},
  {"name": "orders"},
  {"name": "products"}
]
```

### SelectedFields
```json
[
  {
    "table": "users",
    "field": "id",
    "alias": "user_id",
    "aggregation": "",
    "data_type": "integer"
  },
  {
    "table": "orders",
    "field": "total",
    "alias": "total_amount",
    "aggregation": "SUM",
    "data_type": "decimal"
  }
]
```

**Aggregations:** `""`, `"COUNT"`, `"SUM"`, `"AVG"`, `"MIN"`, `"MAX"`

### Joins
```json
[
  {
    "left_table": "orders",
    "right_table": "users",
    "left_field": "user_id",
    "right_field": "id",
    "join_type": "INNER"
  },
  {
    "left_table": "orders",
    "right_table": "products",
    "left_field": "product_id",
    "right_field": "id",
    "join_type": "LEFT"
  }
]
```

**Join Types:** `"INNER"`, `"LEFT"`, `"RIGHT"`, `"FULL"`, `"CROSS"`

### Filters
```json
[
  {
    "field": "users.status",
    "operator": "=",
    "value": "active"
  },
  {
    "field": "orders.total",
    "operator": ">",
    "value": 100
  },
  {
    "field": "orders.created_at",
    "operator": "BETWEEN",
    "value": ["2025-01-01", "2025-12-31"]
  }
]
```

**Operators:** `"="`, `"!="`, `">"`, `">="`, `"<"`, `"<="`, `"LIKE"`, `"IN"`, `"NOT IN"`, `"BETWEEN"`, `"IS NULL"`, `"IS NOT NULL"`

### Sorting
```json
[
  {
    "field": "users.name",
    "direction": "ASC"
  },
  {
    "field": "orders.created_at",
    "direction": "DESC"
  }
]
```

**Directions:** `"ASC"`, `"DESC"`

### Grouping
```json
[
  {"field": "users.id"},
  {"field": "users.name"}
]
```

### Parameters
```json
{
  "start_date": {
    "type": "date",
    "default": "2025-01-01",
    "required": true
  },
  "status": {
    "type": "select",
    "default": "active",
    "options": ["active", "inactive", "pending"]
  },
  "min_amount": {
    "type": "number",
    "default": 0
  }
}
```

---

## Error Responses

### 400 Bad Request
```json
{
  "error": "Invalid request body"
}
```

### 403 Forbidden
```json
{
  "error": "You don't have permission to update this report's datasources"
}
```

### 404 Not Found
```json
{
  "error": "Report not found"
}
```

### 500 Internal Server Error
```json
{
  "error": "Failed to update datasource"
}
```

---

## cURL Examples

### Create Datasource
```bash
curl -X POST http://localhost:8080/api/reports/4da18c13-9190-4a5b-90b2-840a8cf248dc/datasources \
  -H "Content-Type: application/json" \
  -d '{
    "alias": "default",
    "databasealias": "default",
    "querytype": "visual",
    "selectedtables": [{"name": "users"}],
    "selectedfields": [{"table": "users", "field": "name", "alias": "", "aggregation": "", "data_type": "string"}],
    "joins": [],
    "filters": [],
    "sorting": [],
    "grouping": []
  }'
```

### Get Datasources
```bash
curl -X GET http://localhost:8080/api/reports/4da18c13-9190-4a5b-90b2-840a8cf248dc/datasources
```

### Update Datasource
```bash
curl -X PUT http://localhost:8080/api/reports/4da18c13-9190-4a5b-90b2-840a8cf248dc/datasources/aaf5611e-c071-4082-abe5-e43ef94099c2 \
  -H "Content-Type: application/json" \
  -d '{
    "selectedtables": [{"name": "lngcodes"}, {"name": "lngcode_contents"}],
    "joins": [{"left_table": "lngcode_contents", "right_table": "lngcodes", "left_field": "lngcodeid", "right_field": "id", "join_type": "INNER"}]
  }'
```

### Delete Datasource
```bash
curl -X DELETE http://localhost:8080/api/reports/4da18c13-9190-4a5b-90b2-840a8cf248dc/datasources/aaf5611e-c071-4082-abe5-e43ef94099c2
```

---

## Best Practices

1. **Use dedicated endpoints for single datasource updates** - More efficient and clearer intent
2. **Use batch update for multiple datasources** - When updating multiple datasources at once
3. **Send only changed fields** - Partial updates are supported and recommended
4. **Empty arrays vs null** - Use `[]` for empty arrays, `null` for unused fields
5. **Validate JSON structure** - Ensure arrays/objects are properly formatted
6. **Handle errors gracefully** - Check ownership and validation errors

---

## Migration Guide

### From Nested Update to Dedicated Endpoints

**Before (Nested in Report Update):**
```javascript
PUT /api/reports/:id
{
  "datasources": [{
    "id": "xxx",
    "alias": "updated"
  }]
}
```

**After (Dedicated Endpoint):**
```javascript
PUT /api/reports/:id/datasources/:datasourceId
{
  "alias": "updated"
}
```

**Benefits:**
- ✅ Simpler payloads
- ✅ Clearer API semantics
- ✅ Better error messages
- ✅ More efficient
