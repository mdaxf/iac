package models

import (
	"database/sql/driver"
	"encoding/json"
	"time"
)

// ReportType represents the type of report
type ReportType string

const (
	ReportTypeManual      ReportType = "manual"
	ReportTypeAIGenerated ReportType = "ai_generated"
	ReportTypeTemplate    ReportType = "template"
)

// ComponentType represents the type of report component
type ComponentType string

const (
	ComponentTypeTable     ComponentType = "table"
	ComponentTypeChart     ComponentType = "chart"
	ComponentTypeBarcode   ComponentType = "barcode"
	ComponentTypeSubReport ComponentType = "sub_report"
	ComponentTypeText      ComponentType = "text"
	ComponentTypeImage     ComponentType = "image"
	ComponentTypeDrillDown ComponentType = "drill_down"
)

// ChartType represents the type of chart
type ChartType string

const (
	ChartTypeLine       ChartType = "line"
	ChartTypeBar        ChartType = "bar"
	ChartTypePie        ChartType = "pie"
	ChartTypeArea       ChartType = "area"
	ChartTypeScatter    ChartType = "scatter"
	ChartTypeDonut      ChartType = "donut"
	ChartTypeStackedBar ChartType = "stacked_bar"
	ChartTypeStackedArea ChartType = "stacked_area"
	ChartTypeBar3D      ChartType = "bar_3d"
	ChartTypePie3D      ChartType = "pie_3d"
	ChartTypeLine3D     ChartType = "line_3d"
)

// BarcodeType represents the type of barcode
type BarcodeType string

const (
	BarcodeTypeCode128   BarcodeType = "code128"
	BarcodeTypeCode39    BarcodeType = "code39"
	BarcodeTypeEAN13     BarcodeType = "ean13"
	BarcodeTypeEAN8      BarcodeType = "ean8"
	BarcodeTypeUPC       BarcodeType = "upc"
	BarcodeTypeQRCode    BarcodeType = "qr_code"
	BarcodeTypeDataMatrix BarcodeType = "data_matrix"
	BarcodeTypePDF417    BarcodeType = "pdf417"
	BarcodeTypeAztec     BarcodeType = "aztec"
)

// ParameterType represents the type of report parameter
type ParameterType string

const (
	ParameterTypeText        ParameterType = "text"
	ParameterTypeNumber      ParameterType = "number"
	ParameterTypeDate        ParameterType = "date"
	ParameterTypeDateTime    ParameterType = "datetime"
	ParameterTypeSelect      ParameterType = "select"
	ParameterTypeMultiSelect ParameterType = "multi_select"
	ParameterTypeBoolean     ParameterType = "boolean"
)

// JSONMap is a helper type for JSON fields
type JSONMap map[string]interface{}

// Scan implements the sql.Scanner interface
func (j *JSONMap) Scan(value interface{}) error {
	if value == nil {
		*j = nil
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return nil
	}
	return json.Unmarshal(bytes, j)
}

// Value implements the driver.Valuer interface
func (j JSONMap) Value() (driver.Value, error) {
	if j == nil {
		return nil, nil
	}
	return json.Marshal(j)
}

// Report represents a report definition
type Report struct {
	ID               string     `json:"id" gorm:"primaryKey;type:varchar(36);default:(UUID())"`
	Name             string     `json:"name" gorm:"type:varchar(255);not null"`
	Description      string     `json:"description" gorm:"type:text"`
	ReportType       ReportType `json:"report_type" gorm:"type:enum('manual','ai_generated','template');default:'manual'"`
	CreatedBy        string     `json:"created_by" gorm:"type:varchar(36)"`
	IsPublic         bool       `json:"is_public" gorm:"default:false"`
	IsTemplate       bool       `json:"is_template" gorm:"default:false"`
	LayoutConfig     JSONMap    `json:"layout_config" gorm:"type:json"`
	PageSettings     JSONMap    `json:"page_settings" gorm:"type:json"`
	AIPrompt         string     `json:"ai_prompt" gorm:"type:text"`
	AIAnalysis       JSONMap    `json:"ai_analysis" gorm:"type:json"`
	TemplateSourceID string     `json:"template_source_id" gorm:"type:varchar(36)"`
	Tags             JSONMap    `json:"tags" gorm:"type:json"`
	Version          int        `json:"version" gorm:"default:1"`
	IsActive         bool       `json:"is_active" gorm:"default:true"`
	CreatedAt        time.Time  `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt        time.Time  `json:"updated_at" gorm:"autoUpdateTime"`
	LastExecutedAt   *time.Time `json:"last_executed_at"`

	// Relationships
	Datasources []ReportDatasource `json:"datasources,omitempty" gorm:"foreignKey:ReportID;constraint:OnDelete:CASCADE"`
	Components  []ReportComponent  `json:"components,omitempty" gorm:"foreignKey:ReportID;constraint:OnDelete:CASCADE"`
	Parameters  []ReportParameter  `json:"parameters,omitempty" gorm:"foreignKey:ReportID;constraint:OnDelete:CASCADE"`
	Executions  []ReportExecution  `json:"executions,omitempty" gorm:"foreignKey:ReportID;constraint:OnDelete:CASCADE"`
	Shares      []ReportShare      `json:"shares,omitempty" gorm:"foreignKey:ReportID;constraint:OnDelete:CASCADE"`
}

// TableName specifies the table name
func (Report) TableName() string {
	return "reports"
}

// ReportDatasource represents a data source for a report
type ReportDatasource struct {
	ID             string    `json:"id" gorm:"primaryKey;type:varchar(36);default:(UUID())"`
	ReportID       string    `json:"report_id" gorm:"type:varchar(36);not null"`
	Alias          string    `json:"alias" gorm:"type:varchar(100);not null"`
	DatabaseAlias  string    `json:"database_alias" gorm:"type:varchar(100)"`
	QueryType      string    `json:"query_type" gorm:"type:varchar(20);default:'visual'"`
	CustomSQL      string    `json:"custom_sql" gorm:"type:text"`
	SelectedTables JSONMap   `json:"selected_tables" gorm:"type:json"`
	SelectedFields JSONMap   `json:"selected_fields" gorm:"type:json"`
	Joins          JSONMap   `json:"joins" gorm:"type:json"`
	Filters        JSONMap   `json:"filters" gorm:"type:json"`
	Sorting        JSONMap   `json:"sorting" gorm:"type:json"`
	Grouping       JSONMap   `json:"grouping" gorm:"type:json"`
	Parameters     JSONMap   `json:"parameters" gorm:"type:json"`
	CreatedAt      time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt      time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

// TableName specifies the table name
func (ReportDatasource) TableName() string {
	return "report_datasources"
}

// ReportComponent represents a visual component in a report
type ReportComponent struct {
	ID                     string         `json:"id" gorm:"primaryKey;type:varchar(36);default:(UUID())"`
	ReportID               string         `json:"report_id" gorm:"type:varchar(36);not null"`
	ComponentType          ComponentType  `json:"component_type" gorm:"type:enum('table','chart','barcode','sub_report','text','image','drill_down');not null"`
	Name                   string         `json:"name" gorm:"type:varchar(255);not null"`
	X                      float64        `json:"x" gorm:"type:decimal(10,2);default:0"`
	Y                      float64        `json:"y" gorm:"type:decimal(10,2);default:0"`
	Width                  float64        `json:"width" gorm:"type:decimal(10,2);default:200"`
	Height                 float64        `json:"height" gorm:"type:decimal(10,2);default:100"`
	ZIndex                 int            `json:"z_index" gorm:"default:0"`
	DatasourceAlias        string         `json:"datasource_alias" gorm:"type:varchar(100)"`
	DataConfig             JSONMap        `json:"data_config" gorm:"type:json"`
	ComponentConfig        JSONMap        `json:"component_config" gorm:"type:json"`
	StyleConfig            JSONMap        `json:"style_config" gorm:"type:json"`
	ChartType              *ChartType     `json:"chart_type" gorm:"type:enum('line','bar','pie','area','scatter','donut','stacked_bar','stacked_area','bar_3d','pie_3d','line_3d')"`
	ChartConfig            JSONMap        `json:"chart_config" gorm:"type:json"`
	BarcodeType            *BarcodeType   `json:"barcode_type" gorm:"type:enum('code128','code39','ean13','ean8','upc','qr_code','data_matrix','pdf417','aztec')"`
	BarcodeConfig          JSONMap        `json:"barcode_config" gorm:"type:json"`
	DrillDownConfig        JSONMap        `json:"drill_down_config" gorm:"type:json"`
	ConditionalFormatting  JSONMap        `json:"conditional_formatting" gorm:"type:json"`
	IsVisible              bool           `json:"is_visible" gorm:"default:true"`
	CreatedAt              time.Time      `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt              time.Time      `json:"updated_at" gorm:"autoUpdateTime"`
}

// TableName specifies the table name
func (ReportComponent) TableName() string {
	return "report_components"
}

// ReportParameter represents an input parameter for a report
type ReportParameter struct {
	ID              string        `json:"id" gorm:"primaryKey;type:varchar(36);default:(UUID())"`
	ReportID        string        `json:"report_id" gorm:"type:varchar(36);not null"`
	Name            string        `json:"name" gorm:"type:varchar(100);not null"`
	DisplayName     string        `json:"display_name" gorm:"type:varchar(100)"`
	ParameterType   ParameterType `json:"parameter_type" gorm:"type:enum('text','number','date','datetime','select','multi_select','boolean');default:'text'"`
	DefaultValue    string        `json:"default_value" gorm:"type:text"`
	IsRequired      bool          `json:"is_required" gorm:"default:false"`
	IsEnabled       bool          `json:"is_enabled" gorm:"default:true"`
	ValidationRules string        `json:"validation_rules" gorm:"type:text"`
	Options         string        `json:"options" gorm:"type:text"`
	Description     string        `json:"description" gorm:"type:text"`
	SortOrder       int           `json:"sort_order" gorm:"default:0"`
	CreatedAt       time.Time     `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt       time.Time     `json:"updated_at" gorm:"autoUpdateTime"`
}

// TableName specifies the table name
func (ReportParameter) TableName() string {
	return "report_parameters"
}

// ReportExecution represents a report execution record
type ReportExecution struct {
	ID                string    `json:"id" gorm:"primaryKey;type:varchar(36);default:(UUID())"`
	ReportID          string    `json:"report_id" gorm:"type:varchar(36);not null"`
	ExecutedBy        string    `json:"executed_by" gorm:"type:varchar(36)"`
	ExecutionStatus   string    `json:"execution_status" gorm:"type:varchar(20);default:'pending'"`
	ExecutionTimeMs   int       `json:"execution_time_ms"`
	ErrorMessage      string    `json:"error_message" gorm:"type:text"`
	Parameters        JSONMap   `json:"parameters" gorm:"type:json"`
	OutputFormat      string    `json:"output_format" gorm:"type:varchar(20)"`
	OutputSizeBytes   int64     `json:"output_size_bytes"`
	OutputPath        string    `json:"output_path" gorm:"type:varchar(500)"`
	RowCount          int       `json:"row_count"`
	CreatedAt         time.Time `json:"created_at" gorm:"autoCreateTime"`
}

// TableName specifies the table name
func (ReportExecution) TableName() string {
	return "report_executions"
}

// ReportShare represents report sharing permissions
type ReportShare struct {
	ID          string     `json:"id" gorm:"primaryKey;type:varchar(36);default:(UUID())"`
	ReportID    string     `json:"report_id" gorm:"type:varchar(36);not null"`
	SharedBy    string     `json:"shared_by" gorm:"type:varchar(36)"`
	SharedWith  string     `json:"shared_with" gorm:"type:varchar(36)"`
	CanView     bool       `json:"can_view" gorm:"default:true"`
	CanEdit     bool       `json:"can_edit" gorm:"default:false"`
	CanExecute  bool       `json:"can_execute" gorm:"default:true"`
	CanShare    bool       `json:"can_share" gorm:"default:false"`
	ShareToken  string     `json:"share_token" gorm:"type:varchar(255);uniqueIndex"`
	ExpiresAt   *time.Time `json:"expires_at"`
	IsActive    bool       `json:"is_active" gorm:"default:true"`
	CreatedAt   time.Time  `json:"created_at" gorm:"autoCreateTime"`
}

// TableName specifies the table name
func (ReportShare) TableName() string {
	return "report_shares"
}

// ReportTemplate represents a pre-built report template
type ReportTemplate struct {
	ID                 string    `json:"id" gorm:"primaryKey;type:varchar(36);default:(UUID())"`
	Name               string    `json:"name" gorm:"type:varchar(255);not null"`
	Description        string    `json:"description" gorm:"type:text"`
	Category           string    `json:"category" gorm:"type:varchar(100)"`
	TemplateConfig     JSONMap   `json:"template_config" gorm:"type:json"`
	PreviewImage       string    `json:"preview_image" gorm:"type:varchar(500)"`
	UsageCount         int       `json:"usage_count" gorm:"default:0"`
	Rating             float64   `json:"rating" gorm:"type:decimal(3,2);default:0.00"`
	AICompatible       bool      `json:"ai_compatible" gorm:"default:false"`
	AITags             JSONMap   `json:"ai_tags" gorm:"type:json"`
	SuggestedUseCases  JSONMap   `json:"suggested_use_cases" gorm:"type:json"`
	CreatedBy          string    `json:"created_by" gorm:"type:varchar(36)"`
	IsPublic           bool      `json:"is_public" gorm:"default:true"`
	IsSystem           bool      `json:"is_system" gorm:"default:false"`
	CreatedAt          time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt          time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

// TableName specifies the table name
func (ReportTemplate) TableName() string {
	return "report_templates"
}

// ReportSchedule represents a scheduled report execution
type ReportSchedule struct {
	ID             string     `json:"id" gorm:"primaryKey;type:varchar(36);default:(UUID())"`
	ReportID       string     `json:"report_id" gorm:"type:varchar(36);not null"`
	ScheduleName   string     `json:"schedule_name" gorm:"type:varchar(255)"`
	CronExpression string     `json:"cron_expression" gorm:"type:varchar(100);not null"`
	Timezone       string     `json:"timezone" gorm:"type:varchar(50);default:'UTC'"`
	IsActive       bool       `json:"is_active" gorm:"default:true"`
	OutputFormat   string     `json:"output_format" gorm:"type:varchar(20);default:'pdf'"`
	DeliveryMethod string     `json:"delivery_method" gorm:"type:varchar(20);default:'email'"`
	DeliveryConfig JSONMap    `json:"delivery_config" gorm:"type:json"`
	Parameters     JSONMap    `json:"parameters" gorm:"type:json"`
	LastRunAt      *time.Time `json:"last_run_at"`
	NextRunAt      *time.Time `json:"next_run_at"`
	CreatedBy      string     `json:"created_by" gorm:"type:varchar(36)"`
	CreatedAt      time.Time  `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt      time.Time  `json:"updated_at" gorm:"autoUpdateTime"`
}

// TableName specifies the table name
func (ReportSchedule) TableName() string {
	return "report_schedules"
}
