# DataHub Array Mapping Guide

## Overview

The DataHub Transform Engine supports comprehensive array mapping capabilities for complex nested data structures. This guide explains how to map arrays with multiple levels of nesting, handle missing optional nodes, and transform data between different schemas.

## Table of Contents

- [Basic Concepts](#basic-concepts)
- [Array Mapping Modes](#array-mapping-modes)
- [Nested Array Mapping](#nested-array-mapping)
- [Handling Optional Nodes](#handling-optional-nodes)
- [Complex Example](#complex-example)
- [Advanced Features](#advanced-features)
- [Best Practices](#best-practices)

## Basic Concepts

### Field Mapping with Arrays

When mapping arrays, you define an `ArrayMapping` configuration that specifies:

1. **Mode**: How to process the array (iterate, flatten, filter, etc.)
2. **Item Mappings**: Field mappings to apply to each array item
3. **Filter Conditions**: Optional conditions to filter items
4. **Sorting/Limiting**: Optional sorting and result limiting
5. **Grouping/Aggregation**: Optional data aggregation

### Path Syntax

- **Absolute paths**: `$.orders[0].operations` - From root of document
- **Relative paths**: `.part_id` - Relative to current array item
- **Wildcard paths**: `$.orders[*].operations` - All items at that level

## Array Mapping Modes

### 1. Iterate Mode

Processes each array item individually and applies item mappings.

**Use Case**: Transform each order, operation, part, etc.

```json
{
  "source_path": "$.orders",
  "target_path": "$.ProcessedOrders",
  "data_type": "array",
  "array_mapping": {
    "mode": "iterate",
    "item_mappings": [
      {
        "source_path": ".order_id",
        "target_path": "$.OrderNumber",
        "data_type": "string",
        "required": true
      },
      {
        "source_path": ".total",
        "target_path": "$.Amount",
        "data_type": "float",
        "required": true
      }
    ]
  }
}
```

### 2. Flatten Mode

Flattens nested arrays into a single-level array.

**Use Case**: Combine all parts from all operations into one list

```json
{
  "source_path": "$.orders[*].operations[*].parts",
  "target_path": "$.AllParts",
  "data_type": "array",
  "array_mapping": {
    "mode": "flatten"
  }
}
```

### 3. Filter Mode

Filters array items based on conditions.

**Use Case**: Only include assembly operations, exclude test operations

```json
{
  "source_path": "$.operations",
  "target_path": "$.AssemblyOperations",
  "data_type": "array",
  "array_mapping": {
    "mode": "filter",
    "filter_condition": {
      "field": "type",
      "operator": "eq",
      "value": "assembly"
    }
  }
}
```

### 4. Merge Mode

Merges array of objects into a single object.

**Use Case**: Combine multiple configuration objects

```json
{
  "source_path": "$.configs",
  "target_path": "$.MergedConfig",
  "data_type": "array",
  "array_mapping": {
    "mode": "merge"
  }
}
```

### 5. Expand Mode

Expands objects into array items (useful for splitting).

```json
{
  "source_path": "$.dataSet",
  "target_path": "$.ExpandedData",
  "data_type": "array",
  "array_mapping": {
    "mode": "expand"
  }
}
```

## Nested Array Mapping

For deeply nested structures like Orders → Operations → Parts → WIS, you nest `array_mapping` configurations.

### Example: Three Levels Deep

```json
{
  "source_path": "$.orders",
  "target_path": "$.ProcessOrders.Orders",
  "data_type": "array",
  "array_mapping": {
    "mode": "iterate",
    "item_mappings": [
      {
        "source_path": ".order_id",
        "target_path": "$.OrderNumber",
        "data_type": "string",
        "required": true
      },
      {
        "source_path": ".operations",
        "target_path": "$.Operations",
        "data_type": "array",
        "array_mapping": {
          "mode": "iterate",
          "item_mappings": [
            {
              "source_path": ".operation_id",
              "target_path": "$.OpCode",
              "data_type": "string",
              "required": true
            },
            {
              "source_path": ".parts",
              "target_path": "$.PartsList",
              "data_type": "array",
              "optional": true,
              "array_mapping": {
                "mode": "iterate",
                "item_mappings": [
                  {
                    "source_path": ".part_id",
                    "target_path": "$.PartNumber",
                    "data_type": "string",
                    "required": true
                  },
                  {
                    "source_path": ".wis",
                    "target_path": "$.WorkInstructions",
                    "data_type": "array",
                    "optional": true,
                    "array_mapping": {
                      "mode": "iterate",
                      "item_mappings": [
                        {
                          "source_path": ".wis_id",
                          "target_path": "$.InstructionID",
                          "data_type": "string",
                          "required": true
                        }
                      ]
                    }
                  }
                ]
              }
            }
          ]
        }
      }
    ]
  }
}
```

## Handling Optional Nodes

### Problem

In real-world data, some nodes may be missing:
- Not all parts have WIS (work instructions)
- Not all operations have tools
- Some fields may be null or undefined

### Solution

Use `optional: true` and `default_value`:

```json
{
  "source_path": ".wis",
  "target_path": "$.WorkInstructions",
  "data_type": "array",
  "optional": true,
  "default_value": []
}
```

### Examples

**Optional Scalar Field:**
```json
{
  "source_path": ".description",
  "target_path": "$.Description",
  "data_type": "string",
  "optional": true,
  "default_value": "No description available"
}
```

**Optional Array Field:**
```json
{
  "source_path": ".tools",
  "target_path": "$.ToolsRequired",
  "data_type": "array",
  "optional": true,
  "default_value": []
}
```

**Optional Nested Object:**
```json
{
  "source_path": ".metadata",
  "target_path": "$.Metadata",
  "data_type": "object",
  "optional": true,
  "default_value": {}
}
```

## Complex Example

### Scenario

You have a manufacturing order system with:
- **Orders**: Multiple orders per message
- **Operations**: Each order has multiple operations
- **Parts**: Each operation uses multiple parts
- **WIS**: Each part may have work instructions (optional)
- **Tools**: Each operation may require tools (optional)

### Source Data

```json
{
  "orders": [
    {
      "order_id": "ORD-001",
      "customer": "ACME Corp",
      "operations": [
        {
          "operation_id": "OP-001",
          "type": "assembly",
          "parts": [
            {
              "part_id": "PT-100",
              "quantity": 5,
              "wis": [
                {"wis_id": "WIS-A", "step": 1, "description": "Prepare surface"},
                {"wis_id": "WIS-B", "step": 2, "description": "Apply adhesive"}
              ]
            },
            {
              "part_id": "PT-101",
              "quantity": 3
              // Note: no WIS for this part
            }
          ],
          "tools": [
            {"tool_id": "T-001", "name": "Screwdriver"}
          ]
        },
        {
          "operation_id": "OP-002",
          "type": "testing",
          "parts": [
            {"part_id": "PT-200", "quantity": 1}
          ]
          // Note: no tools for testing operation
        }
      ]
    }
  ]
}
```

### Target Schema

```json
{
  "ProcessOrders": {
    "Orders": [
      {
        "OrderNumber": "ORD-001",
        "CustomerName": "ACME Corp",
        "Operations": [
          {
            "OperationCode": "OP-001",
            "OperationType": "assembly",
            "PartsList": [
              {
                "PartNumber": "PT-100",
                "Qty": 5,
                "WorkInstructions": [
                  {"InstructionID": "WIS-A", "StepNumber": 1, "StepDescription": "Prepare surface"},
                  {"InstructionID": "WIS-B", "StepNumber": 2, "StepDescription": "Apply adhesive"}
                ]
              },
              {
                "PartNumber": "PT-101",
                "Qty": 3,
                "WorkInstructions": []
              }
            ],
            "ToolsRequired": [
              {"ToolCode": "T-001", "ToolName": "Screwdriver"}
            ]
          },
          {
            "OperationCode": "OP-002",
            "OperationType": "testing",
            "PartsList": [
              {"PartNumber": "PT-200", "Qty": 1, "WorkInstructions": []}
            ],
            "ToolsRequired": []
          }
        ]
      }
    ]
  }
}
```

### Complete Mapping

See `complex_array_mapping_example.json` for the complete mapping definition.

## Advanced Features

### Sorting

Sort array items by a field:

```json
{
  "array_mapping": {
    "mode": "iterate",
    "sort_by": "step",
    "sort_order": "asc",
    "item_mappings": [...]
  }
}
```

**Sort Orders:**
- `asc`: Ascending order
- `desc`: Descending order

### Limiting

Limit the number of results:

```json
{
  "array_mapping": {
    "mode": "iterate",
    "limit": 10,
    "item_mappings": [...]
  }
}
```

**Use Cases:**
- Top 10 parts by quantity
- First 5 operations
- Latest 20 orders

### Filtering

Filter items based on conditions:

```json
{
  "array_mapping": {
    "mode": "filter",
    "filter_condition": {
      "field": "quantity",
      "operator": "gt",
      "value": 0
    }
  }
}
```

**Operators:**
- `eq`: Equal
- `ne`: Not equal
- `gt`: Greater than
- `lt`: Less than
- `contains`: String contains
- `exists`: Field exists

### Grouping and Aggregation

Group items and aggregate values:

```json
{
  "array_mapping": {
    "mode": "iterate",
    "group_by": "part_id",
    "aggregate_func": "sum"
  }
}
```

**Aggregate Functions:**
- `sum`: Sum values
- `avg`: Average values
- `count`: Count items
- `min`: Minimum value
- `max`: Maximum value

## Best Practices

### 1. Use Optional for Variable Structures

```json
{
  "source_path": ".optional_field",
  "target_path": "$.OptionalField",
  "optional": true,
  "default_value": null
}
```

### 2. Provide Meaningful Default Values

```json
{
  "source_path": ".tools",
  "target_path": "$.Tools",
  "data_type": "array",
  "optional": true,
  "default_value": []  // Empty array instead of null
}
```

### 3. Use Relative Paths in Item Mappings

```json
{
  "item_mappings": [
    {
      "source_path": ".part_id",  // Relative to current item
      "target_path": "$.PartNumber",
      "data_type": "string"
    }
  ]
}
```

### 4. Nest Array Mappings for Nested Arrays

Don't try to access deeply nested arrays in one go - nest the mappings:

**Bad:**
```json
{
  "source_path": "$.orders[*].operations[*].parts[*].wis",
  ...
}
```

**Good:**
```json
{
  "source_path": "$.orders",
  "array_mapping": {
    "mode": "iterate",
    "item_mappings": [
      {
        "source_path": ".operations",
        "array_mapping": {
          "mode": "iterate",
          "item_mappings": [
            {
              "source_path": ".parts",
              "array_mapping": {
                "mode": "iterate",
                "item_mappings": [
                  {
                    "source_path": ".wis",
                    ...
                  }
                ]
              }
            }
          ]
        }
      }
    ]
  }
}
```

### 5. Filter Early

Apply filters at the earliest level to reduce processing:

```json
{
  "source_path": "$.operations",
  "array_mapping": {
    "mode": "filter",
    "filter_condition": {
      "field": "type",
      "operator": "eq",
      "value": "assembly"
    },
    "item_mappings": [...]
  }
}
```

### 6. Sort and Limit for Performance

When dealing with large arrays:

```json
{
  "array_mapping": {
    "mode": "iterate",
    "sort_by": "priority",
    "sort_order": "desc",
    "limit": 100,
    "item_mappings": [...]
  }
}
```

### 7. Use Transform Functions

Apply built-in functions for data transformation:

```json
{
  "source_path": ".created_at",
  "target_path": "$.CreatedDate",
  "data_type": "date",
  "transform_func": "iso8601_to_soap_datetime"
}
```

## Common Patterns

### Pattern 1: One-to-Many Array Transformation

Source has one array, target needs multiple derived arrays:

```json
{
  "mappings": [
    {
      "source_path": "$.operations",
      "target_path": "$.AssemblyOps",
      "data_type": "array",
      "array_mapping": {
        "mode": "filter",
        "filter_condition": {"field": "type", "operator": "eq", "value": "assembly"}
      }
    },
    {
      "source_path": "$.operations",
      "target_path": "$.TestingOps",
      "data_type": "array",
      "array_mapping": {
        "mode": "filter",
        "filter_condition": {"field": "type", "operator": "eq", "value": "testing"}
      }
    }
  ]
}
```

### Pattern 2: Flatten and Group

Flatten nested structure then group by key:

```json
{
  "source_path": "$.orders[*].operations[*].parts",
  "target_path": "$.PartsSummary",
  "data_type": "array",
  "array_mapping": {
    "mode": "iterate",
    "group_by": "part_id",
    "aggregate_func": "count"
  }
}
```

### Pattern 3: Conditional Nested Mapping

Only process nested arrays when parent meets condition:

```json
{
  "source_path": "$.orders",
  "target_path": "$.ActiveOrders",
  "data_type": "array",
  "array_mapping": {
    "mode": "filter",
    "filter_condition": {"field": "status", "operator": "eq", "value": "active"},
    "item_mappings": [
      {
        "source_path": ".operations",
        "target_path": "$.Operations",
        "data_type": "array",
        "array_mapping": {
          "mode": "iterate",
          "item_mappings": [...]
        }
      }
    ]
  }
}
```

## Troubleshooting

### Issue: Array is empty in target

**Causes:**
1. Source path doesn't exist → Use `optional: true`
2. Filter condition excludes all items → Check filter logic
3. Nested path is incorrect → Verify path with sample data

**Solution:**
```json
{
  "source_path": "$.orders",
  "target_path": "$.Orders",
  "data_type": "array",
  "optional": true,
  "default_value": []
}
```

### Issue: Missing nested fields

**Cause:** Parent object is missing

**Solution:**
```json
{
  "source_path": ".parts",
  "target_path": "$.Parts",
  "data_type": "array",
  "optional": true,
  "default_value": [],
  "array_mapping": {...}
}
```

### Issue: Transformation fails on some items

**Cause:** Field doesn't exist in all items

**Solution:**
```json
{
  "item_mappings": [
    {
      "source_path": ".optional_field",
      "target_path": "$.Field",
      "optional": true,
      "default_value": "N/A"
    }
  ]
}
```

## Performance Considerations

1. **Filter Early**: Apply filters before expensive transformations
2. **Limit Results**: Use `limit` for large arrays when you don't need all items
3. **Avoid Deep Nesting**: Consider flattening if nesting exceeds 4-5 levels
4. **Use Indexes**: When accessing specific items, use indexes instead of wildcards
5. **Batch Processing**: For very large datasets, consider splitting into batches

## Summary

The DataHub array mapping system provides:

- ✅ **Nested Array Support**: Map arrays within arrays (unlimited depth)
- ✅ **Optional Node Handling**: Handle missing fields gracefully
- ✅ **Multiple Modes**: Iterate, flatten, filter, merge, expand
- ✅ **Advanced Features**: Sorting, limiting, filtering, grouping
- ✅ **Flexible Paths**: Absolute, relative, and wildcard paths
- ✅ **Transform Functions**: Built-in and custom transformations

For more examples, see `complex_array_mapping_example.json`.
