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

package schema

import (
	"time"
)

// ReportDocument represents the complete report configuration stored in MongoDB
// This structure consolidates all report-related data into a single document
type ReportDocument struct {
	// MongoDB ObjectID (auto-generated)
	ID string `json:"_id,omitempty" bson:"_id,omitempty"`

	// Root Information
	Name        string     `json:"name" bson:"name"`
	Description string     `json:"description,omitempty" bson:"description,omitempty"`
	Revision    int        `json:"revision" bson:"revision"` // Version/revision number
	IsDefault   bool       `json:"isdefault" bson:"isdefault"`
	Category    string     `json:"category,omitempty" bson:"category,omitempty"`

	// Report Configuration
	ReportType       string                 `json:"reporttype" bson:"reporttype"` // manual, ai_generated, template
	IsPublic         bool                   `json:"ispublic" bson:"ispublic"`
	IsTemplate       bool                   `json:"istemplate" bson:"istemplate"`
	LayoutConfig     map[string]interface{} `json:"layoutconfig,omitempty" bson:"layoutconfig,omitempty"`
	PageSettings     map[string]interface{} `json:"pagesettings,omitempty" bson:"pagesettings,omitempty"`
	AIPrompt         string                 `json:"aiprompt,omitempty" bson:"aiprompt,omitempty"`
	AIAnalysis       map[string]interface{} `json:"aianalysis,omitempty" bson:"aianalysis,omitempty"`
	TemplateSourceID string                 `json:"templatesourceid,omitempty" bson:"templatesourceid,omitempty"`
	Tags             []string               `json:"tags,omitempty" bson:"tags,omitempty"`

	// Embedded Data Sources
	Datasources []ReportDatasourceDoc `json:"datasources,omitempty" bson:"datasources,omitempty"`

	// Embedded Components
	Components []ReportComponentDoc `json:"components,omitempty" bson:"components,omitempty"`

	// Embedded Parameters
	Parameters []ReportParameterDoc `json:"parameters,omitempty" bson:"parameters,omitempty"`

	// Execution History (last N executions)
	RecentExecutions []ReportExecutionDoc `json:"recentexecutions,omitempty" bson:"recentexecutions,omitempty"`
	LastExecutedOn   *time.Time           `json:"lastexecutedon,omitempty" bson:"lastexecutedon,omitempty"`

	// Sharing Configuration
	Shares []ReportShareDoc `json:"shares,omitempty" bson:"shares,omitempty"`

	// Standard IAC Audit Fields
	Active          bool      `json:"active" bson:"active"`
	ReferenceID     string    `json:"referenceid,omitempty" bson:"referenceid,omitempty"`
	CreatedBy       string    `json:"createdby" bson:"createdby"`
	CreatedOn       time.Time `json:"createdon" bson:"createdon"`
	ModifiedBy      string    `json:"modifiedby" bson:"modifiedby"`
	ModifiedOn      time.Time `json:"modifiedon" bson:"modifiedon"`
	RowVersionStamp int       `json:"rowversionstamp" bson:"rowversionstamp"`
}

// ReportDatasourceDoc represents a data source embedded in the report document
type ReportDatasourceDoc struct {
	ID             string                 `json:"id" bson:"id"`
	Alias          string                 `json:"alias" bson:"alias"`
	DatabaseAlias  string                 `json:"databasealias,omitempty" bson:"databasealias,omitempty"`
	QueryType      string                 `json:"querytype" bson:"querytype"` // visual, custom
	CustomSQL      string                 `json:"customsql,omitempty" bson:"customsql,omitempty"`
	SelectedTables interface{}            `json:"selectedtables,omitempty" bson:"selectedtables,omitempty"`
	SelectedFields interface{}            `json:"selectedfields,omitempty" bson:"selectedfields,omitempty"`
	Joins          interface{}            `json:"joins,omitempty" bson:"joins,omitempty"`
	Filters        interface{}            `json:"filters,omitempty" bson:"filters,omitempty"`
	Sorting        interface{}            `json:"sorting,omitempty" bson:"sorting,omitempty"`
	Grouping       interface{}            `json:"grouping,omitempty" bson:"grouping,omitempty"`
	Parameters     interface{}            `json:"parameters,omitempty" bson:"parameters,omitempty"`

	// Audit fields
	Active          bool      `json:"active" bson:"active"`
	CreatedBy       string    `json:"createdby" bson:"createdby"`
	CreatedOn       time.Time `json:"createdon" bson:"createdon"`
	ModifiedBy      string    `json:"modifiedby" bson:"modifiedby"`
	ModifiedOn      time.Time `json:"modifiedon" bson:"modifiedon"`
	RowVersionStamp int       `json:"rowversionstamp" bson:"rowversionstamp"`
}

// ReportComponentDoc represents a visual component embedded in the report document
type ReportComponentDoc struct {
	ID                    string                 `json:"id" bson:"id"`
	ComponentType         string                 `json:"componenttype" bson:"componenttype"`
	Name                  string                 `json:"name" bson:"name"`
	X                     float64                `json:"x" bson:"x"`
	Y                     float64                `json:"y" bson:"y"`
	Width                 float64                `json:"width" bson:"width"`
	Height                float64                `json:"height" bson:"height"`
	ZIndex                int                    `json:"zindex" bson:"zindex"`
	DatasourceAlias       string                 `json:"datasourcealias,omitempty" bson:"datasourcealias,omitempty"`
	DataConfig            interface{}            `json:"dataconfig,omitempty" bson:"dataconfig,omitempty"`
	ComponentConfig       interface{}            `json:"componentconfig,omitempty" bson:"componentconfig,omitempty"`
	StyleConfig           interface{}            `json:"styleconfig,omitempty" bson:"styleconfig,omitempty"`
	ChartType             string                 `json:"charttype,omitempty" bson:"charttype,omitempty"`
	ChartConfig           interface{}            `json:"chartconfig,omitempty" bson:"chartconfig,omitempty"`
	BarcodeType           string                 `json:"barcodetype,omitempty" bson:"barcodetype,omitempty"`
	BarcodeConfig         interface{}            `json:"barcodeconfig,omitempty" bson:"barcodeconfig,omitempty"`
	DrillDownConfig       interface{}            `json:"drilldownconfig,omitempty" bson:"drilldownconfig,omitempty"`
	PageBreakConfig       interface{}            `json:"pagebreakconfig,omitempty" bson:"pagebreakconfig,omitempty"`
	PageHeaderConfig      interface{}            `json:"pageheaderconfig,omitempty" bson:"pageheaderconfig,omitempty"`
	PageFooterConfig      interface{}            `json:"pagefooterconfig,omitempty" bson:"pagefooterconfig,omitempty"`
	ConditionalFormatting interface{}            `json:"conditionalformatting,omitempty" bson:"conditionalformatting,omitempty"`
	IsVisible             bool                   `json:"isvisible" bson:"isvisible"`

	// Audit fields
	Active          bool      `json:"active" bson:"active"`
	CreatedBy       string    `json:"createdby" bson:"createdby"`
	CreatedOn       time.Time `json:"createdon" bson:"createdon"`
	ModifiedBy      string    `json:"modifiedby" bson:"modifiedby"`
	ModifiedOn      time.Time `json:"modifiedon" bson:"modifiedon"`
	RowVersionStamp int       `json:"rowversionstamp" bson:"rowversionstamp"`
}

// ReportParameterDoc represents an input parameter embedded in the report document
type ReportParameterDoc struct {
	ID              string `json:"id" bson:"id"`
	Name            string `json:"name" bson:"name"`
	DisplayName     string `json:"displayname,omitempty" bson:"displayname,omitempty"`
	ParameterType   string `json:"parametertype" bson:"parametertype"` // text, number, date, datetime, select, multi_select, boolean
	DefaultValue    string `json:"defaultvalue,omitempty" bson:"defaultvalue,omitempty"`
	IsRequired      bool   `json:"isrequired" bson:"isrequired"`
	IsEnabled       bool   `json:"isenabled" bson:"isenabled"`
	ValidationRules string `json:"validationrules,omitempty" bson:"validationrules,omitempty"`
	Options         string `json:"options,omitempty" bson:"options,omitempty"`
	Description     string `json:"description,omitempty" bson:"description,omitempty"`
	SortOrder       int    `json:"sortorder" bson:"sortorder"`

	// Audit fields
	Active          bool      `json:"active" bson:"active"`
	CreatedBy       string    `json:"createdby" bson:"createdby"`
	CreatedOn       time.Time `json:"createdon" bson:"createdon"`
	ModifiedBy      string    `json:"modifiedby" bson:"modifiedby"`
	ModifiedOn      time.Time `json:"modifiedon" bson:"modifiedon"`
	RowVersionStamp int       `json:"rowversionstamp" bson:"rowversionstamp"`
}

// ReportExecutionDoc represents a report execution record embedded in the report document
type ReportExecutionDoc struct {
	ID              string                 `json:"id" bson:"id"`
	ExecutedBy      string                 `json:"executedby" bson:"executedby"`
	ExecutedOn      time.Time              `json:"executedon" bson:"executedon"`
	ExecutionStatus string                 `json:"executionstatus" bson:"executionstatus"` // pending, running, success, failed
	ExecutionTimeMs int                    `json:"executiontimems" bson:"executiontimems"`
	ErrorMessage    string                 `json:"errormessage,omitempty" bson:"errormessage,omitempty"`
	Parameters      map[string]interface{} `json:"parameters,omitempty" bson:"parameters,omitempty"`
	OutputFormat    string                 `json:"outputformat" bson:"outputformat"`
	OutputSizeBytes int64                  `json:"outputsizebytes" bson:"outputsizebytes"`
	OutputPath      string                 `json:"outputpath,omitempty" bson:"outputpath,omitempty"`
	RowCount        int                    `json:"rowcount" bson:"rowcount"`
}

// ReportShareDoc represents report sharing permissions embedded in the report document
type ReportShareDoc struct {
	ID         string     `json:"id" bson:"id"`
	SharedBy   string     `json:"sharedby" bson:"sharedby"`
	SharedWith string     `json:"sharedwith" bson:"sharedwith"`
	CanView    bool       `json:"canview" bson:"canview"`
	CanEdit    bool       `json:"canedit" bson:"canedit"`
	CanExecute bool       `json:"canexecute" bson:"canexecute"`
	CanShare   bool       `json:"canshare" bson:"canshare"`
	ShareToken string     `json:"sharetoken,omitempty" bson:"sharetoken,omitempty"`
	ExpiresAt  *time.Time `json:"expiresat,omitempty" bson:"expiresat,omitempty"`

	// Audit fields
	Active          bool      `json:"active" bson:"active"`
	CreatedBy       string    `json:"createdby" bson:"createdby"`
	CreatedOn       time.Time `json:"createdon" bson:"createdon"`
	ModifiedBy      string    `json:"modifiedby" bson:"modifiedby"`
	ModifiedOn      time.Time `json:"modifiedon" bson:"modifiedon"`
	RowVersionStamp int       `json:"rowversionstamp" bson:"rowversionstamp"`
}

// Collection name constant
const ReportsCollection = "Reports"

// IndexDefinitions returns the index definitions for the Reports collection
func GetReportIndexDefinitions() []map[string]interface{} {
	return []map[string]interface{}{
		// Index on name for text search
		{
			"keys": map[string]int{
				"name": 1,
			},
			"options": map[string]interface{}{
				"name": "idx_name",
			},
		},
		// Index on category for filtering
		{
			"keys": map[string]int{
				"category": 1,
			},
			"options": map[string]interface{}{
				"name": "idx_category",
			},
		},
		// Index on createdby for user-based queries
		{
			"keys": map[string]int{
				"createdby": 1,
			},
			"options": map[string]interface{}{
				"name": "idx_createdby",
			},
		},
		// Index on isdefault and active for quick default lookup
		{
			"keys": map[string]int{
				"isdefault": 1,
				"active":    1,
			},
			"options": map[string]interface{}{
				"name": "idx_isdefault_active",
			},
		},
		// Compound index for common queries
		{
			"keys": map[string]int{
				"ispublic":   1,
				"active":     1,
				"modifiedon": -1,
			},
			"options": map[string]interface{}{
				"name": "idx_ispublic_active_modifiedon",
			},
		},
		// Index on reporttype for filtering
		{
			"keys": map[string]int{
				"reporttype": 1,
			},
			"options": map[string]interface{}{
				"name": "idx_reporttype",
			},
		},
	}
}
