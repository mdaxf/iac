package datahub

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
)

// applyArrayMapping handles array transformations with iteration support
func (te *TransformEngine) applyArrayMapping(sourceJSON string, targetBody map[string]interface{}, mapping *FieldMapping) error {
	dhLogger.Infof("Applying array mapping for %s -> %s (mode: %s)", mapping.SourcePath, mapping.TargetPath, mapping.ArrayMapping.Mode)

	// Get source array
	result := gjson.Get(sourceJSON, mapping.SourcePath)
	if !result.Exists() || !result.IsArray() {
		if mapping.Optional {
			return nil
		}
		if mapping.Required {
			return fmt.Errorf("source array path %s not found or is not an array", mapping.SourcePath)
		}
		return nil
	}

	sourceArray := result.Array()
	dhLogger.Debugf("Found %d items in source array %s", len(sourceArray), mapping.SourcePath)

	// Process based on mode
	switch mapping.ArrayMapping.Mode {
	case "iterate":
		return te.arrayModeIterate(sourceJSON, targetBody, mapping, sourceArray)
	case "flatten":
		return te.arrayModeFlatten(sourceJSON, targetBody, mapping, sourceArray)
	case "filter":
		return te.arrayModeFilter(sourceJSON, targetBody, mapping, sourceArray)
	case "merge":
		return te.arrayModeMerge(sourceJSON, targetBody, mapping, sourceArray)
	case "expand":
		return te.arrayModeExpand(sourceJSON, targetBody, mapping, sourceArray)
	default:
		return fmt.Errorf("unknown array mapping mode: %s", mapping.ArrayMapping.Mode)
	}
}

// arrayModeIterate iterates over each array item and applies item mappings
func (te *TransformEngine) arrayModeIterate(sourceJSON string, targetBody map[string]interface{}, mapping *FieldMapping, sourceArray []gjson.Result) error {
	targetArray := make([]interface{}, 0)

	for i, item := range sourceArray {
		// Create JSON for this item
		itemJSON := item.String()
		if !item.IsObject() && !item.IsArray() {
			// Wrap primitive values
			itemJSON = fmt.Sprintf(`{"value": %s}`, item.Raw)
		}

		// Apply item mappings
		itemTarget := make(map[string]interface{})

		for _, itemMapping := range mapping.ArrayMapping.ItemMappings {
			// Adjust paths for current item context
			itemMappingCopy := itemMapping

			// If the source path starts with ".", it's relative to current item
			// Otherwise, it's relative to the root with array index
			if strings.HasPrefix(itemMapping.SourcePath, ".") {
				// Relative to current item
				itemMappingCopy.SourcePath = strings.TrimPrefix(itemMapping.SourcePath, ".")
			} else if strings.HasPrefix(itemMapping.SourcePath, "$") {
				// Keep as is (root reference)
			} else {
				// Relative to current item, add $ prefix
				itemMappingCopy.SourcePath = "$." + itemMapping.SourcePath
			}

			// Apply the mapping to this item
			if err := te.applyFieldMapping(itemJSON, itemTarget, &itemMappingCopy); err != nil {
				if !itemMapping.Optional {
					dhLogger.Warnf("Failed to apply item mapping at index %d: %v", i, err)
				}
			}
		}

		// Apply filter condition if specified
		if mapping.ArrayMapping.FilterCondition != nil {
			if !te.evaluateConditionOnData(itemTarget, mapping.ArrayMapping.FilterCondition) {
				continue // Skip this item
			}
		}

		targetArray = append(targetArray, itemTarget)
	}

	// Apply sorting if specified
	if mapping.ArrayMapping.SortBy != "" {
		targetArray = te.sortArray(targetArray, mapping.ArrayMapping.SortBy, mapping.ArrayMapping.SortOrder)
	}

	// Apply limit if specified
	if mapping.ArrayMapping.Limit > 0 && len(targetArray) > mapping.ArrayMapping.Limit {
		targetArray = targetArray[:mapping.ArrayMapping.Limit]
	}

	// Apply grouping/aggregation if specified
	if mapping.ArrayMapping.GroupBy != "" {
		targetArray = te.groupArray(targetArray, mapping.ArrayMapping.GroupBy, mapping.ArrayMapping.AggregateFunc)
	}

	// Set the target array
	targetJSON, err := json.Marshal(targetBody)
	if err != nil {
		return fmt.Errorf("failed to marshal target body: %w", err)
	}

	targetPath := mapping.TargetPath
	if strings.HasPrefix(targetPath, "//") {
		targetPath = te.xpathToJSONPath(targetPath)
	}

	updatedJSON, err := sjson.Set(string(targetJSON), targetPath, targetArray)
	if err != nil {
		return fmt.Errorf("failed to set target array: %w", err)
	}

	if err := json.Unmarshal([]byte(updatedJSON), &targetBody); err != nil {
		return fmt.Errorf("failed to unmarshal updated target body: %w", err)
	}

	dhLogger.Infof("Array mapping completed: %d items processed, %d items in result", len(sourceArray), len(targetArray))
	return nil
}

// arrayModeFlatten flattens nested arrays into a single array
func (te *TransformEngine) arrayModeFlatten(sourceJSON string, targetBody map[string]interface{}, mapping *FieldMapping, sourceArray []gjson.Result) error {
	flatArray := make([]interface{}, 0)

	for _, item := range sourceArray {
		if item.IsArray() {
			// Recursively flatten nested arrays
			for _, nestedItem := range item.Array() {
				flatArray = append(flatArray, nestedItem.Value())
			}
		} else {
			flatArray = append(flatArray, item.Value())
		}
	}

	// Set the flattened array
	targetJSON, err := json.Marshal(targetBody)
	if err != nil {
		return fmt.Errorf("failed to marshal target body: %w", err)
	}

	targetPath := mapping.TargetPath
	if strings.HasPrefix(targetPath, "//") {
		targetPath = te.xpathToJSONPath(targetPath)
	}

	updatedJSON, err := sjson.Set(string(targetJSON), targetPath, flatArray)
	if err != nil {
		return fmt.Errorf("failed to set flattened array: %w", err)
	}

	if err := json.Unmarshal([]byte(updatedJSON), &targetBody); err != nil {
		return fmt.Errorf("failed to unmarshal updated target body: %w", err)
	}

	dhLogger.Infof("Flattened array: %d items -> %d items", len(sourceArray), len(flatArray))
	return nil
}

// arrayModeFilter filters array items based on condition
func (te *TransformEngine) arrayModeFilter(sourceJSON string, targetBody map[string]interface{}, mapping *FieldMapping, sourceArray []gjson.Result) error {
	if mapping.ArrayMapping.FilterCondition == nil {
		return fmt.Errorf("filter mode requires a filter_condition")
	}

	filteredArray := make([]interface{}, 0)

	for _, item := range sourceArray {
		itemData := item.Value()

		// Convert to map for condition evaluation
		var itemMap map[string]interface{}
		switch v := itemData.(type) {
		case map[string]interface{}:
			itemMap = v
		default:
			itemMap = map[string]interface{}{"value": v}
		}

		if te.evaluateConditionOnData(itemMap, mapping.ArrayMapping.FilterCondition) {
			filteredArray = append(filteredArray, itemData)
		}
	}

	// Set the filtered array
	targetJSON, err := json.Marshal(targetBody)
	if err != nil {
		return fmt.Errorf("failed to marshal target body: %w", err)
	}

	targetPath := mapping.TargetPath
	if strings.HasPrefix(targetPath, "//") {
		targetPath = te.xpathToJSONPath(targetPath)
	}

	updatedJSON, err := sjson.Set(string(targetJSON), targetPath, filteredArray)
	if err != nil {
		return fmt.Errorf("failed to set filtered array: %w", err)
	}

	if err := json.Unmarshal([]byte(updatedJSON), &targetBody); err != nil {
		return fmt.Errorf("failed to unmarshal updated target body: %w", err)
	}

	dhLogger.Infof("Filtered array: %d items -> %d items", len(sourceArray), len(filteredArray))
	return nil
}

// arrayModeMerge merges multiple arrays or objects
func (te *TransformEngine) arrayModeMerge(sourceJSON string, targetBody map[string]interface{}, mapping *FieldMapping, sourceArray []gjson.Result) error {
	// Merge all objects in the array into a single object
	mergedObject := make(map[string]interface{})

	for _, item := range sourceArray {
		if item.IsObject() {
			itemMap := item.Map()
			for key, value := range itemMap {
				mergedObject[key] = value.Value()
			}
		}
	}

	// Set the merged object
	targetJSON, err := json.Marshal(targetBody)
	if err != nil {
		return fmt.Errorf("failed to marshal target body: %w", err)
	}

	targetPath := mapping.TargetPath
	if strings.HasPrefix(targetPath, "//") {
		targetPath = te.xpathToJSONPath(targetPath)
	}

	updatedJSON, err := sjson.Set(string(targetJSON), targetPath, mergedObject)
	if err != nil {
		return fmt.Errorf("failed to set merged object: %w", err)
	}

	if err := json.Unmarshal([]byte(updatedJSON), &targetBody); err != nil {
		return fmt.Errorf("failed to unmarshal updated target body: %w", err)
	}

	dhLogger.Infof("Merged %d objects into one", len(sourceArray))
	return nil
}

// arrayModeExpand expands a single object into an array of objects
func (te *TransformEngine) arrayModeExpand(sourceJSON string, targetBody map[string]interface{}, mapping *FieldMapping, sourceArray []gjson.Result) error {
	// This mode can be used to split objects by a certain criterion
	// For now, just convert the array as-is
	expandedArray := make([]interface{}, 0)
	for _, item := range sourceArray {
		expandedArray = append(expandedArray, item.Value())
	}

	// Set the expanded array
	targetJSON, err := json.Marshal(targetBody)
	if err != nil {
		return fmt.Errorf("failed to marshal target body: %w", err)
	}

	targetPath := mapping.TargetPath
	if strings.HasPrefix(targetPath, "//") {
		targetPath = te.xpathToJSONPath(targetPath)
	}

	updatedJSON, err := sjson.Set(string(targetJSON), targetPath, expandedArray)
	if err != nil {
		return fmt.Errorf("failed to set expanded array: %w", err)
	}

	if err := json.Unmarshal([]byte(updatedJSON), &targetBody); err != nil {
		return fmt.Errorf("failed to unmarshal updated target body: %w", err)
	}

	return nil
}

// applyNestedMapping handles nested object mappings
func (te *TransformEngine) applyNestedMapping(sourceJSON string, targetBody map[string]interface{}, mapping *FieldMapping) error {
	dhLogger.Infof("Applying nested mapping for %s -> %s", mapping.SourcePath, mapping.TargetPath)

	// Get source object
	result := gjson.Get(sourceJSON, mapping.SourcePath)
	if !result.Exists() {
		if mapping.Optional {
			return nil
		}
		if mapping.Required {
			return fmt.Errorf("source object path %s not found", mapping.SourcePath)
		}
		return nil
	}

	// Get the source object as JSON string
	sourceObjectJSON := result.String()
	if !result.IsObject() {
		sourceObjectJSON = fmt.Sprintf(`{"value": %s}`, result.Raw)
	}

	// Apply nested mappings
	nestedTarget := make(map[string]interface{})
	for _, nestedMapping := range mapping.NestedMappings {
		if err := te.applyFieldMapping(sourceObjectJSON, nestedTarget, &nestedMapping); err != nil {
			if !nestedMapping.Optional {
				dhLogger.Warnf("Failed to apply nested mapping: %v", err)
			}
		}
	}

	// Set the nested object
	targetJSON, err := json.Marshal(targetBody)
	if err != nil {
		return fmt.Errorf("failed to marshal target body: %w", err)
	}

	targetPath := mapping.TargetPath
	if strings.HasPrefix(targetPath, "//") {
		targetPath = te.xpathToJSONPath(targetPath)
	}

	updatedJSON, err := sjson.Set(string(targetJSON), targetPath, nestedTarget)
	if err != nil {
		return fmt.Errorf("failed to set nested object: %w", err)
	}

	if err := json.Unmarshal([]byte(updatedJSON), &targetBody); err != nil {
		return fmt.Errorf("failed to unmarshal updated target body: %w", err)
	}

	return nil
}

// sortArray sorts an array of objects by a field
func (te *TransformEngine) sortArray(array []interface{}, sortBy string, sortOrder string) []interface{} {
	if sortOrder == "" {
		sortOrder = "asc"
	}

	sort.Slice(array, func(i, j int) bool {
		iMap, iOk := array[i].(map[string]interface{})
		jMap, jOk := array[j].(map[string]interface{})

		if !iOk || !jOk {
			return false
		}

		iVal, iExists := iMap[sortBy]
		jVal, jExists := jMap[sortBy]

		if !iExists || !jExists {
			return false
		}

		// Simple comparison - could be enhanced for different types
		iStr := fmt.Sprintf("%v", iVal)
		jStr := fmt.Sprintf("%v", jVal)

		if sortOrder == "desc" {
			return iStr > jStr
		}
		return iStr < jStr
	})

	return array
}

// groupArray groups array items by a field and applies aggregation
func (te *TransformEngine) groupArray(array []interface{}, groupBy string, aggregateFunc string) []interface{} {
	groups := make(map[string][]interface{})

	// Group items
	for _, item := range array {
		itemMap, ok := item.(map[string]interface{})
		if !ok {
			continue
		}

		groupKey := fmt.Sprintf("%v", itemMap[groupBy])
		groups[groupKey] = append(groups[groupKey], item)
	}

	// Apply aggregation
	result := make([]interface{}, 0)
	for key, items := range groups {
		groupedItem := map[string]interface{}{
			groupBy: key,
			"count": len(items),
		}

		// Apply aggregate function if specified
		if aggregateFunc != "" {
			// Simple implementation - could be enhanced
			groupedItem["items"] = items
		}

		result = append(result, groupedItem)
	}

	return result
}

// evaluateConditionOnData evaluates a condition against data
func (te *TransformEngine) evaluateConditionOnData(data map[string]interface{}, condition *MappingCondition) bool {
	// Get field value
	fieldValue, exists := data[condition.Field]
	if !exists {
		if condition.Operator == "exists" {
			return condition.Value == false || condition.Value == "false"
		}
		return false
	}

	// Evaluate based on operator
	switch condition.Operator {
	case "eq":
		return fmt.Sprintf("%v", fieldValue) == fmt.Sprintf("%v", condition.Value)
	case "ne":
		return fmt.Sprintf("%v", fieldValue) != fmt.Sprintf("%v", condition.Value)
	case "gt":
		// Simple string comparison - could be enhanced for numbers
		return fmt.Sprintf("%v", fieldValue) > fmt.Sprintf("%v", condition.Value)
	case "lt":
		return fmt.Sprintf("%v", fieldValue) < fmt.Sprintf("%v", condition.Value)
	case "contains":
		fieldStr := fmt.Sprintf("%v", fieldValue)
		valueStr := fmt.Sprintf("%v", condition.Value)
		return strings.Contains(fieldStr, valueStr)
	case "exists":
		return condition.Value == true || condition.Value == "true"
	default:
		return false
	}
}
