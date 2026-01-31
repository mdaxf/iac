package models

import (
	"time"
)

// DataSet represents a dataset schema stored in the system
type DataSet struct {
	ID              string                 `json:"_id" bson:"_id"`
	Schema          string                 `json:"$schema" bson:"$schema"`
	Ref             string                 `json:"$ref" bson:"$ref"`
	Name            string                 `json:"name" bson:"name"`
	Version         string                 `json:"version" bson:"version"`
	IsDefault       bool                   `json:"isdefault" bson:"isdefault"`
	DataSourceType  string                 `json:"datasourcetype" bson:"datasourcetype"`
	DataSource      string                 `json:"datasource" bson:"datasource"`
	ListFields      []string               `json:"listfields" bson:"listfields"`
	HiddenFields    []string               `json:"hiddenfields" bson:"hiddenfields"`
	KeyField        string                 `json:"keyfield" bson:"keyfield"`
	Query           string                 `json:"query,omitempty" bson:"query,omitempty"`
	DetailPage      map[string]interface{} `json:"detailpage" bson:"detailpage"`
	Definitions     map[string]interface{} `json:"definitions" bson:"definitions"`
	Description     string                 `json:"description,omitempty" bson:"description,omitempty"`
	Tags            []string               `json:"tags,omitempty" bson:"tags,omitempty"`

	// Dynamic root-level page nodes (simplepage, listpage, cardpage, etc.)
	// These are stored as a map to allow any custom page node names
	ExtraPages map[string]interface{} `json:"extrapages,omitempty" bson:"extrapages,omitempty"`

	// Audit Fields
	CreatedBy       string    `json:"createdby" bson:"createdby"`
	CreatedOn       time.Time `json:"createdon" bson:"createdon"`
	ModifiedBy      string    `json:"modifiedby" bson:"modifiedby"`
	ModifiedOn      time.Time `json:"modifiedon" bson:"modifiedon"`
	RowVersionStamp int       `json:"rowversionstamp" bson:"rowversionstamp"`
}

// DataSetCreateRequest represents the request to create a new dataset
type DataSetCreateRequest struct {
	Schema         string                 `json:"$schema" binding:"required"`
	Ref            string                 `json:"$ref" binding:"required"`
	Name           string                 `json:"name" binding:"required"`
	Version        string                 `json:"version"`
	IsDefault      bool                   `json:"isdefault"`
	DataSourceType string                 `json:"datasourcetype" binding:"required"`
	DataSource     string                 `json:"datasource" binding:"required"`
	ListFields     []string               `json:"listfields"`
	HiddenFields   []string               `json:"hiddenfields"`
	KeyField       string                 `json:"keyfield" binding:"required"`
	Query          string                 `json:"query,omitempty"`
	DetailPage     map[string]interface{} `json:"detailpage"`
	Definitions    map[string]interface{} `json:"definitions"`
	Description    string                 `json:"description,omitempty"`
	Tags           []string               `json:"tags,omitempty"`
	ExtraPages     map[string]interface{} `json:"extrapages,omitempty"`
}

// DataSetUpdateRequest represents the request to update an existing dataset
type DataSetUpdateRequest struct {
	Schema         string                 `json:"$schema,omitempty"`
	Ref            string                 `json:"$ref,omitempty"`
	Name           string                 `json:"name,omitempty"`
	Version        string                 `json:"version,omitempty"`
	IsDefault      bool                   `json:"isdefault,omitempty"`
	DataSourceType string                 `json:"datasourcetype,omitempty"`
	DataSource     string                 `json:"datasource,omitempty"`
	ListFields     []string               `json:"listfields,omitempty"`
	HiddenFields   []string               `json:"hiddenfields,omitempty"`
	KeyField       string                 `json:"keyfield,omitempty"`
	Query          string                 `json:"query,omitempty"`
	DetailPage     map[string]interface{} `json:"detailpage,omitempty"`
	Definitions    map[string]interface{} `json:"definitions,omitempty"`
	Description    string                 `json:"description,omitempty"`
	Tags           []string               `json:"tags,omitempty"`
	ExtraPages     map[string]interface{} `json:"extrapages,omitempty"`
}

// DataSetListResponse represents the response for listing datasets
type DataSetListResponse struct {
	Data  []DataSet `json:"data"`
	Count int       `json:"count"`
}
