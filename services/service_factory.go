package services

import (
	"fmt"

	dbconn "github.com/mdaxf/iac/databases"
	"gorm.io/gorm"
)

// ServiceFactory creates service instances with multi-database support
type ServiceFactory struct {
	dbHelper             *DatabaseHelper
	appDB                *gorm.DB
	businessEntitySvc    *BusinessEntityService
	queryTemplateSvc     *QueryTemplateService
	reportSvc            *ReportService
	chatSvc              *ChatService
	schemaMetadataSvc    *SchemaMetadataService
	schemaMetadataMultiDBSvc *SchemaMetadataServiceMultiDB
	aiReportSvc          *AIReportService
}

// NewServiceFactory creates a new service factory
func NewServiceFactory(poolManager *dbconn.PoolManager, appDB *gorm.DB) (*ServiceFactory, error) {
	if poolManager == nil || appDB == nil {
		return nil, fmt.Errorf("poolManager and appDB cannot be nil")
	}

	// Create database selector
	selector := dbconn.NewDatabaseSelector(poolManager)

	// Set round-robin strategy
	selector.SetStrategy(dbconn.StrategyRoundRobin)

	// Create database helper
	dbHelper := NewDatabaseHelper(selector, appDB)

	return &ServiceFactory{
		dbHelper: dbHelper,
		appDB:    appDB,
	}, nil
}

// GetBusinessEntityService returns the business entity service
func (f *ServiceFactory) GetBusinessEntityService() *BusinessEntityService {
	if f.businessEntitySvc == nil {
		f.businessEntitySvc = NewBusinessEntityService(f.appDB)
	}
	return f.businessEntitySvc
}

// GetQueryTemplateService returns the query template service
func (f *ServiceFactory) GetQueryTemplateService() *QueryTemplateService {
	if f.queryTemplateSvc == nil {
		f.queryTemplateSvc = NewQueryTemplateService(f.appDB)
	}
	return f.queryTemplateSvc
}

// GetSchemaMetadataService returns the schema metadata service (legacy)
func (f *ServiceFactory) GetSchemaMetadataService() *SchemaMetadataService {
	if f.schemaMetadataSvc == nil {
		f.schemaMetadataSvc = NewSchemaMetadataService(f.appDB)
	}
	return f.schemaMetadataSvc
}

// GetSchemaMetadataServiceMultiDB returns the multi-database schema metadata service
func (f *ServiceFactory) GetSchemaMetadataServiceMultiDB() *SchemaMetadataServiceMultiDB {
	if f.schemaMetadataMultiDBSvc == nil {
		f.schemaMetadataMultiDBSvc = NewSchemaMetadataServiceMultiDB(f.dbHelper, f.appDB)
	}
	return f.schemaMetadataMultiDBSvc
}

// GetDatabaseHelper returns the database helper for custom operations
func (f *ServiceFactory) GetDatabaseHelper() *DatabaseHelper {
	return f.dbHelper
}

// GetAppDB returns the application's GORM database instance
func (f *ServiceFactory) GetAppDB() *gorm.DB {
	return f.appDB
}

// Close closes all database connections (call on shutdown)
func (f *ServiceFactory) Close() error {
	// Note: The pool manager should be closed separately by the application
	// This method is here for future cleanup tasks
	return nil
}
