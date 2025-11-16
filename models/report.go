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
	ChartTypeLine        ChartType = "line"
	ChartTypeBar         ChartType = "bar"
	ChartTypePie         ChartType = "pie"
	ChartTypeArea        ChartType = "area"
	ChartTypeScatter     ChartType = "scatter"
	ChartTypeDonut       ChartType = "donut"
	ChartTypeStackedBar  ChartType = "stacked_bar"
	ChartTypeStackedArea ChartType = "stacked_area"
	ChartTypeBar3D       ChartType = "bar_3d"
	ChartTypePie3D       ChartType = "pie_3d"
	ChartTypeLine3D      ChartType = "line_3d"
)

// BarcodeType represents the type of barcode
type BarcodeType string

const (
	BarcodeTypeCode128    BarcodeType = "code128"
	BarcodeTypeCode39     BarcodeType = "code39"
	BarcodeTypeEAN13      BarcodeType = "ean13"
	BarcodeTypeEAN8       BarcodeType = "ean8"
	BarcodeTypeUPC        BarcodeType = "upc"
	BarcodeTypeQRCode     BarcodeType = "qr_code"
	BarcodeTypeDataMatrix BarcodeType = "data_matrix"
	BarcodeTypePDF417     BarcodeType = "pdf417"
	BarcodeTypeAztec      BarcodeType = "aztec"
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
	ID               string     `json:"id" gorm:"primaryKey;column:id;type:varchar(36);default:(UUID())"`
	Name             string     `json:"name" gorm:"column:name;type:varchar(255);not null"`
	Description      string     `json:"description" gorm:"column:description;type:text"`
	ReportType       ReportType `json:"reporttype" gorm:"column:reporttype;type:enum('manual','ai_generated','template');default:'manual'"`
	IsPublic         bool       `json:"ispublic" gorm:"column:ispublic;default:false"`
	IsTemplate       bool       `json:"istemplate" gorm:"column:istemplate;default:false"`
	LayoutConfig     JSONMap    `json:"layoutconfig" gorm:"column:layoutconfig;type:json"`
	PageSettings     JSONMap    `json:"pagesettings" gorm:"column:pagesettings;type:json"`
	AIPrompt         string     `json:"aiprompt" gorm:"column:aiprompt;type:text"`
	AIAnalysis       JSONMap    `json:"aianalysis" gorm:"column:aianalysis;type:json"`
	TemplateSourceID string     `json:"templatesourceid" gorm:"column:templatesourceid;type:varchar(36)"`
	Tags             JSONMap    `json:"tags" gorm:"column:tags;type:json"`
	Version          int        `json:"version" gorm:"column:version;default:1"`
	LastExecutedOn   *time.Time `json:"lastexecutedon" gorm:"column:lastexecutedon"`

	// Relationships
	Datasources []ReportDatasource `json:"datasources,omitempty" gorm:"foreignKey:ReportID;constraint:OnDelete:CASCADE"`
	Components  []ReportComponent  `json:"components,omitempty" gorm:"foreignKey:ReportID;constraint:OnDelete:CASCADE"`
	Parameters  []ReportParameter  `json:"parameters,omitempty" gorm:"foreignKey:ReportID;constraint:OnDelete:CASCADE"`
	Executions  []ReportExecution  `json:"executions,omitempty" gorm:"foreignKey:ReportID;constraint:OnDelete:CASCADE"`
	Shares      []ReportShare      `json:"shares,omitempty" gorm:"foreignKey:ReportID;constraint:OnDelete:CASCADE"`

	// Standard IAC audit fields (must be at end)
	Active           bool      `json:"active" gorm:"column:active;default:true"`
	ReferenceID      string    `json:"referenceid" gorm:"column:referenceid;type:varchar(36)"`
	CreatedBy        string    `json:"createdby" gorm:"column:createdby;type:varchar(45)"`
	CreatedOn        time.Time `json:"createdon" gorm:"column:createdon;autoCreateTime"`
	ModifiedBy       string    `json:"modifiedby" gorm:"column:modifiedby;type:varchar(45)"`
	ModifiedOn       time.Time `json:"modifiedon" gorm:"column:modifiedon;autoUpdateTime"`
	RowVersionStamp  int       `json:"rowversionstamp" gorm:"column:rowversionstamp;default:1"`
}

// TableName specifies the table name
func (Report) TableName() string {
	return "reports"
}

// ReportDatasource represents a data source for a report
type ReportDatasource struct {
	ID             string  `json:"id" gorm:"primaryKey;type:varchar(36);default:(UUID())"`
	ReportID       string  `json:"reportid" gorm:"type:varchar(36);not null"`
	Alias          string  `json:"alias" gorm:"type:varchar(100);not null"`
	DatabaseAlias  string  `json:"databasealias" gorm:"type:varchar(100)"`
	QueryType      string  `json:"querytype" gorm:"type:varchar(20);default:'visual'"`
	CustomSQL      string  `json:"customsql" gorm:"type:text"`
	SelectedTables JSONMap `json:"selectedtables" gorm:"type:json"`
	SelectedFields JSONMap `json:"selectedfields" gorm:"type:json"`
	Joins          JSONMap `json:"joins" gorm:"type:json"`
	Filters        JSONMap `json:"filters" gorm:"type:json"`
	Sorting        JSONMap `json:"sorting" gorm:"type:json"`
	Grouping       JSONMap `json:"grouping" gorm:"type:json"`
	Parameters     JSONMap `json:"parameters" gorm:"type:json"`

	// Standard IAC audit fields (must be at end)
	Active          bool      `json:"active" gorm:"default:true"`
	ReferenceID     string    `json:"referenceid" gorm:"type:varchar(36)"`
	CreatedBy       string    `json:"createdby" gorm:"type:varchar(45)"`
	CreatedOn       time.Time `json:"createdon" gorm:"autoCreateTime"`
	ModifiedBy      string    `json:"modifiedby" gorm:"type:varchar(45)"`
	ModifiedOn      time.Time `json:"modifiedon" gorm:"autoUpdateTime"`
	RowVersionStamp int       `json:"rowversionstamp" gorm:"default:1"`
}

// TableName specifies the table name
func (ReportDatasource) TableName() string {
	return "reportdatasources"
}

// ReportComponent represents a visual component in a report
type ReportComponent struct {
	ID                    string        `json:"id" gorm:"primaryKey;type:varchar(36);default:(UUID())"`
	ReportID              string        `json:"reportid" gorm:"type:varchar(36);not null"`
	ComponentType         ComponentType `json:"componenttype" gorm:"type:enum('table','chart','barcode','sub_report','text','image','drill_down');not null"`
	Name                  string        `json:"name" gorm:"type:varchar(255);not null"`
	X                     float64       `json:"x" gorm:"type:decimal(10,2);default:0"`
	Y                     float64       `json:"y" gorm:"type:decimal(10,2);default:0"`
	Width                 float64       `json:"width" gorm:"type:decimal(10,2);default:200"`
	Height                float64       `json:"height" gorm:"type:decimal(10,2);default:100"`
	ZIndex                int           `json:"zindex" gorm:"default:0"`
	DatasourceAlias       string        `json:"datasourcealias" gorm:"type:varchar(100)"`
	DataConfig            JSONMap       `json:"dataconfig" gorm:"type:json"`
	ComponentConfig       JSONMap       `json:"componentconfig" gorm:"type:json"`
	StyleConfig           JSONMap       `json:"styleconfig" gorm:"type:json"`
	ChartType             *ChartType    `json:"charttype" gorm:"type:enum('line','bar','pie','area','scatter','donut','stacked_bar','stacked_area','bar_3d','pie_3d','line_3d')"`
	ChartConfig           JSONMap       `json:"chartconfig" gorm:"type:json"`
	BarcodeType           *BarcodeType  `json:"barcodetype" gorm:"type:enum('code128','code39','ean13','ean8','upc','qr_code','data_matrix','pdf417','aztec')"`
	BarcodeConfig         JSONMap       `json:"barcodeconfig" gorm:"type:json"`
	DrillDownConfig       JSONMap       `json:"drilldownconfig" gorm:"type:json"`
	ConditionalFormatting JSONMap       `json:"conditionalformatting" gorm:"type:json"`
	IsVisible             bool          `json:"isvisible" gorm:"default:true"`

	// Standard IAC audit fields (must be at end)
	Active          bool      `json:"active" gorm:"default:true"`
	ReferenceID     string    `json:"referenceid" gorm:"type:varchar(36)"`
	CreatedBy       string    `json:"createdby" gorm:"type:varchar(45)"`
	CreatedOn       time.Time `json:"createdon" gorm:"autoCreateTime"`
	ModifiedBy      string    `json:"modifiedby" gorm:"type:varchar(45)"`
	ModifiedOn      time.Time `json:"modifiedon" gorm:"autoUpdateTime"`
	RowVersionStamp int       `json:"rowversionstamp" gorm:"default:1"`
}

// TableName specifies the table name
func (ReportComponent) TableName() string {
	return "reportcomponents"
}

// ReportParameter represents an input parameter for a report
type ReportParameter struct {
	ID              string        `json:"id" gorm:"primaryKey;type:varchar(36);default:(UUID())"`
	ReportID        string        `json:"reportid" gorm:"type:varchar(36);not null"`
	Name            string        `json:"name" gorm:"type:varchar(100);not null"`
	DisplayName     string        `json:"displayname" gorm:"type:varchar(100)"`
	ParameterType   ParameterType `json:"parametertype" gorm:"type:enum('text','number','date','datetime','select','multi_select','boolean');default:'text'"`
	DefaultValue    string        `json:"defaultvalue" gorm:"type:text"`
	IsRequired      bool          `json:"isrequired" gorm:"default:false"`
	IsEnabled       bool          `json:"isenabled" gorm:"default:true"`
	ValidationRules string        `json:"validationrules" gorm:"type:text"`
	Options         string        `json:"options" gorm:"type:text"`
	Description     string        `json:"description" gorm:"type:text"`
	SortOrder       int           `json:"sortorder" gorm:"default:0"`

	// Standard IAC audit fields (must be at end)
	Active          bool      `json:"active" gorm:"default:true"`
	ReferenceID     string    `json:"referenceid" gorm:"type:varchar(36)"`
	CreatedBy       string    `json:"createdby" gorm:"type:varchar(45)"`
	CreatedOn       time.Time `json:"createdon" gorm:"autoCreateTime"`
	ModifiedBy      string    `json:"modifiedby" gorm:"type:varchar(45)"`
	ModifiedOn      time.Time `json:"modifiedon" gorm:"autoUpdateTime"`
	RowVersionStamp int       `json:"rowversionstamp" gorm:"default:1"`
}

// TableName specifies the table name
func (ReportParameter) TableName() string {
	return "reportparameters"
}

// ReportExecution represents a report execution record
type ReportExecution struct {
	ID              string    `json:"id" gorm:"primaryKey;type:varchar(36);default:(UUID())"`
	ReportID        string    `json:"reportid" gorm:"type:varchar(36);not null"`
	ExecutedBy      string    `json:"executedby" gorm:"type:varchar(36)"`
	ExecutionStatus string    `json:"executionstatus" gorm:"type:varchar(20);default:'pending'"`
	ExecutionTimeMs int       `json:"executiontimems"`
	ErrorMessage    string    `json:"errormessage" gorm:"type:text"`
	Parameters      JSONMap   `json:"parameters" gorm:"type:json"`
	OutputFormat    string    `json:"outputformat" gorm:"type:varchar(20)"`
	OutputSizeBytes int64     `json:"outputsizebytes"`
	OutputPath      string    `json:"outputpath" gorm:"type:varchar(500)"`
	RowCount        int       `json:"rowcount"`

	// Standard IAC audit fields (must be at end)
	Active          bool      `json:"active" gorm:"default:true"`
	ReferenceID     string    `json:"referenceid" gorm:"type:varchar(36)"`
	CreatedBy       string    `json:"createdby" gorm:"type:varchar(45)"`
	CreatedOn       time.Time `json:"createdon" gorm:"autoCreateTime"`
	ModifiedBy      string    `json:"modifiedby" gorm:"type:varchar(45)"`
	ModifiedOn      time.Time `json:"modifiedon" gorm:"autoUpdateTime"`
	RowVersionStamp int       `json:"rowversionstamp" gorm:"default:1"`
}

// TableName specifies the table name
func (ReportExecution) TableName() string {
	return "reportexecutions"
}

// ReportShare represents report sharing permissions
type ReportShare struct {
	ID         string     `json:"id" gorm:"primaryKey;type:varchar(36);default:(UUID())"`
	ReportID   string     `json:"reportid" gorm:"type:varchar(36);not null"`
	SharedBy   string     `json:"sharedby" gorm:"type:varchar(36)"`
	SharedWith string     `json:"sharedwith" gorm:"type:varchar(36)"`
	CanView    bool       `json:"canview" gorm:"default:true"`
	CanEdit    bool       `json:"canedit" gorm:"default:false"`
	CanExecute bool       `json:"canexecute" gorm:"default:true"`
	CanShare   bool       `json:"canshare" gorm:"default:false"`
	ShareToken string     `json:"sharetoken" gorm:"type:varchar(255);uniqueIndex"`
	ExpiresAt  *time.Time `json:"expiresat"`

	// Standard IAC audit fields (must be at end)
	Active          bool      `json:"active" gorm:"default:true"`
	ReferenceID     string    `json:"referenceid" gorm:"type:varchar(36)"`
	CreatedBy       string    `json:"createdby" gorm:"type:varchar(45)"`
	CreatedOn       time.Time `json:"createdon" gorm:"autoCreateTime"`
	ModifiedBy      string    `json:"modifiedby" gorm:"type:varchar(45)"`
	ModifiedOn      time.Time `json:"modifiedon" gorm:"autoUpdateTime"`
	RowVersionStamp int       `json:"rowversionstamp" gorm:"default:1"`
}

// TableName specifies the table name
func (ReportShare) TableName() string {
	return "reportshares"
}

// ReportTemplate represents a pre-built report template
type ReportTemplate struct {
	ID                string  `json:"id" gorm:"primaryKey;type:varchar(36);default:(UUID())"`
	Name              string  `json:"name" gorm:"type:varchar(255);not null"`
	Description       string  `json:"description" gorm:"type:text"`
	Category          string  `json:"category" gorm:"type:varchar(100)"`
	TemplateConfig    JSONMap `json:"templateconfig" gorm:"type:json"`
	PreviewImage      string  `json:"previewimage" gorm:"type:varchar(500)"`
	UsageCount        int     `json:"usagecount" gorm:"default:0"`
	Rating            float64 `json:"rating" gorm:"type:decimal(3,2);default:0.00"`
	AICompatible      bool    `json:"aicompatible" gorm:"default:false"`
	AITags            JSONMap `json:"aitags" gorm:"type:json"`
	SuggestedUseCases JSONMap `json:"suggestedusecases" gorm:"type:json"`
	IsPublic          bool    `json:"ispublic" gorm:"default:true"`
	IsSystem          bool    `json:"issystem" gorm:"default:false"`

	// Standard IAC audit fields (must be at end)
	Active          bool      `json:"active" gorm:"default:true"`
	ReferenceID     string    `json:"referenceid" gorm:"type:varchar(36)"`
	CreatedBy       string    `json:"createdby" gorm:"type:varchar(45)"`
	CreatedOn       time.Time `json:"createdon" gorm:"autoCreateTime"`
	ModifiedBy      string    `json:"modifiedby" gorm:"type:varchar(45)"`
	ModifiedOn      time.Time `json:"modifiedon" gorm:"autoUpdateTime"`
	RowVersionStamp int       `json:"rowversionstamp" gorm:"default:1"`
}

// TableName specifies the table name
func (ReportTemplate) TableName() string {
	return "reporttemplates"
}

// ReportSchedule represents a scheduled report execution
type ReportSchedule struct {
	ID             string     `json:"id" gorm:"primaryKey;type:varchar(36);default:(UUID())"`
	ReportID       string     `json:"reportid" gorm:"type:varchar(36);not null"`
	ScheduleName   string     `json:"schedulename" gorm:"type:varchar(255)"`
	CronExpression string     `json:"cronexpression" gorm:"type:varchar(100);not null"`
	Timezone       string     `json:"timezone" gorm:"type:varchar(50);default:'UTC'"`
	OutputFormat   string     `json:"outputformat" gorm:"type:varchar(20);default:'pdf'"`
	DeliveryMethod string     `json:"deliverymethod" gorm:"type:varchar(20);default:'email'"`
	DeliveryConfig JSONMap    `json:"deliveryconfig" gorm:"type:json"`
	Parameters     JSONMap    `json:"parameters" gorm:"type:json"`
	LastRunAt      *time.Time `json:"lastrunat"`
	NextRunAt      *time.Time `json:"nextrunat"`

	// Standard IAC audit fields (must be at end)
	Active          bool      `json:"active" gorm:"default:true"`
	ReferenceID     string    `json:"referenceid" gorm:"type:varchar(36)"`
	CreatedBy       string    `json:"createdby" gorm:"type:varchar(45)"`
	CreatedOn       time.Time `json:"createdon" gorm:"autoCreateTime"`
	ModifiedBy      string    `json:"modifiedby" gorm:"type:varchar(45)"`
	ModifiedOn      time.Time `json:"modifiedon" gorm:"autoUpdateTime"`
	RowVersionStamp int       `json:"rowversionstamp" gorm:"default:1"`
}

// TableName specifies the table name
func (ReportSchedule) TableName() string {
	return "reportschedules"
}
