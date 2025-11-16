// Copyright 2023 IAC. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package documents

import (
	"fmt"
	"time"
)

// Query represents a document database query builder
type Query struct {
	conditions []Condition
	dbType     DocDBType
}

// Condition represents a query condition
type Condition struct {
	Field    string
	Operator string
	Value    interface{}
	Logic    string // "AND" or "OR"
	Children []Condition
}

// NewQuery creates a new query builder
func NewQuery(dbType DocDBType) *Query {
	return &Query{
		conditions: make([]Condition, 0),
		dbType:     dbType,
	}
}

// Equals adds an equality condition
func (q *Query) Equals(field string, value interface{}) *Query {
	q.conditions = append(q.conditions, Condition{
		Field:    field,
		Operator: "$eq",
		Value:    value,
		Logic:    "AND",
	})
	return q
}

// NotEquals adds a not-equals condition
func (q *Query) NotEquals(field string, value interface{}) *Query {
	q.conditions = append(q.conditions, Condition{
		Field:    field,
		Operator: "$ne",
		Value:    value,
		Logic:    "AND",
	})
	return q
}

// GreaterThan adds a greater-than condition
func (q *Query) GreaterThan(field string, value interface{}) *Query {
	q.conditions = append(q.conditions, Condition{
		Field:    field,
		Operator: "$gt",
		Value:    value,
		Logic:    "AND",
	})
	return q
}

// GreaterThanOrEqual adds a greater-than-or-equal condition
func (q *Query) GreaterThanOrEqual(field string, value interface{}) *Query {
	q.conditions = append(q.conditions, Condition{
		Field:    field,
		Operator: "$gte",
		Value:    value,
		Logic:    "AND",
	})
	return q
}

// LessThan adds a less-than condition
func (q *Query) LessThan(field string, value interface{}) *Query {
	q.conditions = append(q.conditions, Condition{
		Field:    field,
		Operator: "$lt",
		Value:    value,
		Logic:    "AND",
	})
	return q
}

// LessThanOrEqual adds a less-than-or-equal condition
func (q *Query) LessThanOrEqual(field string, value interface{}) *Query {
	q.conditions = append(q.conditions, Condition{
		Field:    field,
		Operator: "$lte",
		Value:    value,
		Logic:    "AND",
	})
	return q
}

// In adds an in-array condition
func (q *Query) In(field string, values []interface{}) *Query {
	q.conditions = append(q.conditions, Condition{
		Field:    field,
		Operator: "$in",
		Value:    values,
		Logic:    "AND",
	})
	return q
}

// NotIn adds a not-in-array condition
func (q *Query) NotIn(field string, values []interface{}) *Query {
	q.conditions = append(q.conditions, Condition{
		Field:    field,
		Operator: "$nin",
		Value:    values,
		Logic:    "AND",
	})
	return q
}

// Contains adds a contains/like condition
func (q *Query) Contains(field string, value string) *Query {
	q.conditions = append(q.conditions, Condition{
		Field:    field,
		Operator: "$regex",
		Value:    value,
		Logic:    "AND",
	})
	return q
}

// StartsWith adds a starts-with condition
func (q *Query) StartsWith(field string, value string) *Query {
	q.conditions = append(q.conditions, Condition{
		Field:    field,
		Operator: "$regex",
		Value:    "^" + value,
		Logic:    "AND",
	})
	return q
}

// EndsWith adds an ends-with condition
func (q *Query) EndsWith(field string, value string) *Query {
	q.conditions = append(q.conditions, Condition{
		Field:    field,
		Operator: "$regex",
		Value:    value + "$",
		Logic:    "AND",
	})
	return q
}

// Exists adds a field existence check
func (q *Query) Exists(field string, exists bool) *Query {
	q.conditions = append(q.conditions, Condition{
		Field:    field,
		Operator: "$exists",
		Value:    exists,
		Logic:    "AND",
	})
	return q
}

// IsNull adds a null check
func (q *Query) IsNull(field string) *Query {
	q.conditions = append(q.conditions, Condition{
		Field:    field,
		Operator: "$eq",
		Value:    nil,
		Logic:    "AND",
	})
	return q
}

// IsNotNull adds a not-null check
func (q *Query) IsNotNull(field string) *Query {
	q.conditions = append(q.conditions, Condition{
		Field:    field,
		Operator: "$ne",
		Value:    nil,
		Logic:    "AND",
	})
	return q
}

// Between adds a range condition
func (q *Query) Between(field string, start, end interface{}) *Query {
	q.conditions = append(q.conditions, Condition{
		Field:    field,
		Operator: "$gte",
		Value:    start,
		Logic:    "AND",
	})
	q.conditions = append(q.conditions, Condition{
		Field:    field,
		Operator: "$lte",
		Value:    end,
		Logic:    "AND",
	})
	return q
}

// Or creates an OR condition group
func (q *Query) Or(queries ...*Query) *Query {
	orConditions := make([]Condition, 0)
	for _, query := range queries {
		for _, cond := range query.conditions {
			cond.Logic = "OR"
			orConditions = append(orConditions, cond)
		}
	}

	if len(orConditions) > 0 {
		q.conditions = append(q.conditions, Condition{
			Operator: "$or",
			Children: orConditions,
			Logic:    "AND",
		})
	}

	return q
}

// And creates an AND condition group
func (q *Query) And(queries ...*Query) *Query {
	for _, query := range queries {
		q.conditions = append(q.conditions, query.conditions...)
	}
	return q
}

// Not negates a query
func (q *Query) Not(query *Query) *Query {
	if len(query.conditions) > 0 {
		q.conditions = append(q.conditions, Condition{
			Operator: "$not",
			Children: query.conditions,
			Logic:    "AND",
		})
	}
	return q
}

// Build builds the final filter for the specific database type
func (q *Query) Build() interface{} {
	switch q.dbType {
	case DocDBTypeMongoDB:
		return q.buildMongoDBFilter()
	case DocDBTypePostgres:
		return q.buildPostgresFilter()
	default:
		return q.buildMongoDBFilter() // Default to MongoDB format
	}
}

// buildMongoDBFilter builds a MongoDB-style filter
func (q *Query) buildMongoDBFilter() map[string]interface{} {
	if len(q.conditions) == 0 {
		return map[string]interface{}{}
	}

	filter := make(map[string]interface{})

	for _, cond := range q.conditions {
		if cond.Operator == "$or" || cond.Operator == "$and" || cond.Operator == "$not" {
			// Handle logical operators
			childFilters := make([]map[string]interface{}, len(cond.Children))
			for i, child := range cond.Children {
				childFilters[i] = q.buildCondition(child)
			}
			filter[cond.Operator] = childFilters
		} else {
			// Handle regular conditions
			condFilter := q.buildCondition(cond)
			for k, v := range condFilter {
				filter[k] = v
			}
		}
	}

	return filter
}

// buildCondition builds a single condition
func (q *Query) buildCondition(cond Condition) map[string]interface{} {
	if cond.Operator == "$eq" {
		return map[string]interface{}{cond.Field: cond.Value}
	}

	return map[string]interface{}{
		cond.Field: map[string]interface{}{
			cond.Operator: cond.Value,
		},
	}
}

// buildPostgresFilter builds a PostgreSQL-compatible filter
func (q *Query) buildPostgresFilter() map[string]interface{} {
	// For PostgreSQL JSONB, we'll use a simplified format
	// that the adapter can translate to SQL
	if len(q.conditions) == 0 {
		return map[string]interface{}{}
	}

	filter := make(map[string]interface{})

	// Simplified: only handle simple equality for now
	// The PostgreSQL adapter will need to handle more complex queries
	for _, cond := range q.conditions {
		if cond.Operator == "$eq" {
			filter[cond.Field] = cond.Value
		} else {
			// Store the operator and value for the adapter to process
			filter[cond.Field] = map[string]interface{}{
				"_operator": cond.Operator,
				"_value":    cond.Value,
			}
		}
	}

	return filter
}

// Update represents a document update builder
type Update struct {
	operations map[string]interface{}
	dbType     DocDBType
}

// NewUpdate creates a new update builder
func NewUpdate(dbType DocDBType) *Update {
	return &Update{
		operations: make(map[string]interface{}),
		dbType:     dbType,
	}
}

// Set sets a field value
func (u *Update) Set(field string, value interface{}) *Update {
	if u.operations["$set"] == nil {
		u.operations["$set"] = make(map[string]interface{})
	}
	u.operations["$set"].(map[string]interface{})[field] = value
	return u
}

// Unset removes a field
func (u *Update) Unset(field string) *Update {
	if u.operations["$unset"] == nil {
		u.operations["$unset"] = make(map[string]interface{})
	}
	u.operations["$unset"].(map[string]interface{})[field] = ""
	return u
}

// Increment increments a numeric field
func (u *Update) Increment(field string, value interface{}) *Update {
	if u.operations["$inc"] == nil {
		u.operations["$inc"] = make(map[string]interface{})
	}
	u.operations["$inc"].(map[string]interface{})[field] = value
	return u
}

// Multiply multiplies a numeric field
func (u *Update) Multiply(field string, value interface{}) *Update {
	if u.operations["$mul"] == nil {
		u.operations["$mul"] = make(map[string]interface{})
	}
	u.operations["$mul"].(map[string]interface{})[field] = value
	return u
}

// Min updates to minimum value
func (u *Update) Min(field string, value interface{}) *Update {
	if u.operations["$min"] == nil {
		u.operations["$min"] = make(map[string]interface{})
	}
	u.operations["$min"].(map[string]interface{})[field] = value
	return u
}

// Max updates to maximum value
func (u *Update) Max(field string, value interface{}) *Update {
	if u.operations["$max"] == nil {
		u.operations["$max"] = make(map[string]interface{})
	}
	u.operations["$max"].(map[string]interface{})[field] = value
	return u
}

// Push adds a value to an array
func (u *Update) Push(field string, value interface{}) *Update {
	if u.operations["$push"] == nil {
		u.operations["$push"] = make(map[string]interface{})
	}
	u.operations["$push"].(map[string]interface{})[field] = value
	return u
}

// Pull removes values from an array
func (u *Update) Pull(field string, value interface{}) *Update {
	if u.operations["$pull"] == nil {
		u.operations["$pull"] = make(map[string]interface{})
	}
	u.operations["$pull"].(map[string]interface{})[field] = value
	return u
}

// AddToSet adds to array if not present
func (u *Update) AddToSet(field string, value interface{}) *Update {
	if u.operations["$addToSet"] == nil {
		u.operations["$addToSet"] = make(map[string]interface{})
	}
	u.operations["$addToSet"].(map[string]interface{})[field] = value
	return u
}

// Pop removes first or last element from array
func (u *Update) Pop(field string, position int) *Update {
	if u.operations["$pop"] == nil {
		u.operations["$pop"] = make(map[string]interface{})
	}
	// -1 for first, 1 for last
	u.operations["$pop"].(map[string]interface{})[field] = position
	return u
}

// Rename renames a field
func (u *Update) Rename(oldField, newField string) *Update {
	if u.operations["$rename"] == nil {
		u.operations["$rename"] = make(map[string]interface{})
	}
	u.operations["$rename"].(map[string]interface{})[oldField] = newField
	return u
}

// CurrentDate sets field to current date
func (u *Update) CurrentDate(field string) *Update {
	if u.operations["$currentDate"] == nil {
		u.operations["$currentDate"] = make(map[string]interface{})
	}
	u.operations["$currentDate"].(map[string]interface{})[field] = true
	return u
}

// Build builds the final update operation
func (u *Update) Build() interface{} {
	switch u.dbType {
	case DocDBTypeMongoDB:
		return u.operations
	case DocDBTypePostgres:
		// For PostgreSQL, convert to $set format if needed
		if len(u.operations) == 1 {
			if setOps, ok := u.operations["$set"]; ok {
				return setOps
			}
		}
		return u.operations
	default:
		return u.operations
	}
}

// Sort represents a sort specification builder
type Sort struct {
	fields []SortField
}

// SortField represents a single sort field
type SortField struct {
	Field string
	Order int // 1 for ascending, -1 for descending
}

// NewSort creates a new sort builder
func NewSort() *Sort {
	return &Sort{
		fields: make([]SortField, 0),
	}
}

// Ascending adds an ascending sort
func (s *Sort) Ascending(field string) *Sort {
	s.fields = append(s.fields, SortField{
		Field: field,
		Order: 1,
	})
	return s
}

// Descending adds a descending sort
func (s *Sort) Descending(field string) *Sort {
	s.fields = append(s.fields, SortField{
		Field: field,
		Order: -1,
	})
	return s
}

// Build builds the sort specification
func (s *Sort) Build() map[string]int {
	result := make(map[string]int)
	for _, field := range s.fields {
		result[field.Field] = field.Order
	}
	return result
}

// Projection represents a field projection builder
type Projection struct {
	fields map[string]int
}

// NewProjection creates a new projection builder
func NewProjection() *Projection {
	return &Projection{
		fields: make(map[string]int),
	}
}

// Include includes a field in the result
func (p *Projection) Include(field string) *Projection {
	p.fields[field] = 1
	return p
}

// Exclude excludes a field from the result
func (p *Projection) Exclude(field string) *Projection {
	p.fields[field] = 0
	return p
}

// IncludeFields includes multiple fields
func (p *Projection) IncludeFields(fields ...string) *Projection {
	for _, field := range fields {
		p.fields[field] = 1
	}
	return p
}

// ExcludeFields excludes multiple fields
func (p *Projection) ExcludeFields(fields ...string) *Projection {
	for _, field := range fields {
		p.fields[field] = 0
	}
	return p
}

// Build builds the projection specification
func (p *Projection) Build() map[string]int {
	return p.fields
}

// QueryHelper provides helper methods for common queries
type QueryHelper struct {
	dbType DocDBType
}

// NewQueryHelper creates a new query helper
func NewQueryHelper(dbType DocDBType) *QueryHelper {
	return &QueryHelper{dbType: dbType}
}

// ByID creates a query by document ID
func (qh *QueryHelper) ByID(id string) *Query {
	return NewQuery(qh.dbType).Equals("_id", id)
}

// ByUUID creates a query by UUID
func (qh *QueryHelper) ByUUID(uuid string) *Query {
	return NewQuery(qh.dbType).Equals("_uuid", uuid)
}

// ByField creates a simple equality query
func (qh *QueryHelper) ByField(field string, value interface{}) *Query {
	return NewQuery(qh.dbType).Equals(field, value)
}

// DateRange creates a date range query
func (qh *QueryHelper) DateRange(field string, start, end time.Time) *Query {
	return NewQuery(qh.dbType).Between(field, start, end)
}

// RecentDocuments queries documents created within duration
func (qh *QueryHelper) RecentDocuments(field string, duration time.Duration) *Query {
	since := time.Now().Add(-duration)
	return NewQuery(qh.dbType).GreaterThan(field, since)
}

// TextSearch creates a text search query
func (qh *QueryHelper) TextSearch(field, text string) *Query {
	return NewQuery(qh.dbType).Contains(field, text)
}

// MultiFieldSearch searches across multiple fields
func (qh *QueryHelper) MultiFieldSearch(fields []string, value string) *Query {
	queries := make([]*Query, len(fields))
	for i, field := range fields {
		queries[i] = NewQuery(qh.dbType).Contains(field, value)
	}
	return NewQuery(qh.dbType).Or(queries...)
}

// ActiveRecords queries active (non-deleted) records
func (qh *QueryHelper) ActiveRecords() *Query {
	return NewQuery(qh.dbType).
		Or(
			NewQuery(qh.dbType).Equals("deleted", false),
			NewQuery(qh.dbType).Exists("deleted", false),
		)
}

// DeletedRecords queries deleted records
func (qh *QueryHelper) DeletedRecords() *Query {
	return NewQuery(qh.dbType).Equals("deleted", true)
}

// PaginationOptions creates find options for pagination
func PaginationOptions(page, pageSize int64, sort *Sort) *FindOptions {
	skip := (page - 1) * pageSize

	opts := &FindOptions{
		Limit: pageSize,
		Skip:  skip,
	}

	if sort != nil {
		opts.Sort = sort.Build()
	}

	return opts
}

// Example usage documentation
func ExampleQueryUsage() {
	// Create a query for MongoDB
	query := NewQuery(DocDBTypeMongoDB).
		Equals("status", "active").
		GreaterThan("age", 18).
		In("category", []interface{}{"A", "B", "C"}).
		Contains("name", "John")

	filter := query.Build()
	fmt.Printf("MongoDB Filter: %v\n", filter)

	// Create an update
	update := NewUpdate(DocDBTypeMongoDB).
		Set("status", "updated").
		Increment("views", 1).
		Push("tags", "new-tag")

	updateDoc := update.Build()
	fmt.Printf("Update: %v\n", updateDoc)

	// Create sort and projection
	sort := NewSort().
		Descending("createdAt").
		Ascending("name")

	projection := NewProjection().
		IncludeFields("name", "email", "status").
		Exclude("_id")

	// Create find options with pagination
	opts := &FindOptions{
		Sort:       sort.Build(),
		Projection: projection.Build(),
		Limit:      10,
		Skip:       0,
	}

	fmt.Printf("Find Options: %v\n", opts)
}
