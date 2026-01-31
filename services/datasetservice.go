package services

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/mdaxf/iac/documents"
	"github.com/mdaxf/iac/models"
	"go.mongodb.org/mongo-driver/bson"
)

const (
	CollectionDataSets = "DataSets"
)

// DataSetService handles dataset schema storage and management
type DataSetService struct {
	docDB documents.DocumentDB
}

// NewDataSetService creates a new dataset service
func NewDataSetService(docDB documents.DocumentDB) *DataSetService {
	return &DataSetService{
		docDB: docDB,
	}
}

// CreateDataSet creates a new dataset schema
func (ds *DataSetService) CreateDataSet(ctx context.Context, req *models.DataSetCreateRequest, user string) (*models.DataSet, error) {
	// Generate unique ID
	dataSetID := uuid.New().String()

	// Create dataset record
	now := time.Now()
	dataset := &models.DataSet{
		ID:              dataSetID,
		Schema:          req.Schema,
		Ref:             req.Ref,
		Name:            req.Name,
		Version:         req.Version,
		IsDefault:       req.IsDefault,
		DataSourceType:  req.DataSourceType,
		DataSource:      req.DataSource,
		ListFields:      req.ListFields,
		HiddenFields:    req.HiddenFields,
		KeyField:        req.KeyField,
		Query:           req.Query,
		DetailPage:      req.DetailPage,
		Definitions:     req.Definitions,
		Description:     req.Description,
		Tags:            req.Tags,
		ExtraPages:      req.ExtraPages,
		CreatedBy:       user,
		CreatedOn:       now,
		ModifiedBy:      user,
		ModifiedOn:      now,
		RowVersionStamp: 1,
	}

	// Save to database
	data, err := bson.Marshal(dataset)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal dataset: %w", err)
	}

	var datasetMap map[string]interface{}
	if err := bson.Unmarshal(data, &datasetMap); err != nil {
		return nil, fmt.Errorf("failed to unmarshal to map: %w", err)
	}

	if _, err := ds.docDB.InsertOne(ctx, CollectionDataSets, datasetMap); err != nil {
		return nil, fmt.Errorf("failed to insert dataset record: %w", err)
	}

	return dataset, nil
}

// ListDataSets retrieves all datasets
func (ds *DataSetService) ListDataSets(ctx context.Context) ([]models.DataSet, error) {
	filter := bson.M{}
	results, err := ds.docDB.FindMany(ctx, CollectionDataSets, filter, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to list datasets: %w", err)
	}

	var datasets []models.DataSet
	for _, result := range results {
		doc := bson.M(result)
		var d models.DataSet
		data, err := bson.Marshal(doc)
		if err != nil {
			continue
		}
		if err := bson.Unmarshal(data, &d); err != nil {
			continue
		}
		datasets = append(datasets, d)
	}

	return datasets, nil
}

// GetDataSet retrieves a dataset by ID
func (ds *DataSetService) GetDataSet(ctx context.Context, id string) (*models.DataSet, error) {
	filter := bson.M{"_id": id}
	result, err := ds.docDB.FindOne(ctx, CollectionDataSets, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to get dataset: %w", err)
	}

	doc := bson.M(result)
	var d models.DataSet
	data, err := bson.Marshal(doc)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal dataset: %w", err)
	}
	if err := bson.Unmarshal(data, &d); err != nil {
		return nil, fmt.Errorf("failed to unmarshal dataset: %w", err)
	}

	return &d, nil
}

// GetDataSetByName retrieves a dataset by name
func (ds *DataSetService) GetDataSetByName(ctx context.Context, name string) (*models.DataSet, error) {
	filter := bson.M{"name": name}
	result, err := ds.docDB.FindOne(ctx, CollectionDataSets, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to get dataset by name: %w", err)
	}

	doc := bson.M(result)
	var d models.DataSet
	data, err := bson.Marshal(doc)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal dataset: %w", err)
	}
	if err := bson.Unmarshal(data, &d); err != nil {
		return nil, fmt.Errorf("failed to unmarshal dataset: %w", err)
	}

	return &d, nil
}

// UpdateDataSet updates dataset metadata and schema
func (ds *DataSetService) UpdateDataSet(ctx context.Context, id string, req *models.DataSetUpdateRequest, user string) (*models.DataSet, error) {
	// Get existing dataset
	existing, err := ds.GetDataSet(ctx, id)
	if err != nil {
		return nil, err
	}

	now := time.Now()

	// Update fields if provided
	if req.Schema != "" {
		existing.Schema = req.Schema
	}
	if req.Ref != "" {
		existing.Ref = req.Ref
	}
	if req.Name != "" {
		existing.Name = req.Name
	}
	if req.Version != "" {
		existing.Version = req.Version
	}
	existing.IsDefault = req.IsDefault
	if req.DataSourceType != "" {
		existing.DataSourceType = req.DataSourceType
	}
	if req.DataSource != "" {
		existing.DataSource = req.DataSource
	}
	if req.ListFields != nil {
		existing.ListFields = req.ListFields
	}
	if req.HiddenFields != nil {
		existing.HiddenFields = req.HiddenFields
	}
	if req.KeyField != "" {
		existing.KeyField = req.KeyField
	}
	if req.Query != "" {
		existing.Query = req.Query
	}
	if req.DetailPage != nil {
		existing.DetailPage = req.DetailPage
	}
	if req.Definitions != nil {
		existing.Definitions = req.Definitions
	}
	if req.Description != "" {
		existing.Description = req.Description
	}
	if req.Tags != nil {
		existing.Tags = req.Tags
	}
	if req.ExtraPages != nil {
		existing.ExtraPages = req.ExtraPages
	}

	existing.ModifiedBy = user
	existing.ModifiedOn = now
	existing.RowVersionStamp++

	// Convert to map for update
	data, err := bson.Marshal(existing)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal dataset: %w", err)
	}

	var datasetMap map[string]interface{}
	if err := bson.Unmarshal(data, &datasetMap); err != nil {
		return nil, fmt.Errorf("failed to unmarshal to map: %w", err)
	}

	filter := bson.M{"_id": id}
	if err := ds.docDB.UpdateOne(ctx, CollectionDataSets, filter, bson.M{"$set": datasetMap}); err != nil {
		return nil, fmt.Errorf("failed to update dataset: %w", err)
	}

	return existing, nil
}

// DeleteDataSet deletes a dataset
func (ds *DataSetService) DeleteDataSet(ctx context.Context, id string) error {
	filter := bson.M{"_id": id}
	if err := ds.docDB.DeleteOne(ctx, CollectionDataSets, filter); err != nil {
		return fmt.Errorf("failed to delete dataset record: %w", err)
	}

	return nil
}

// SearchDataSets searches datasets by name or tags
func (ds *DataSetService) SearchDataSets(ctx context.Context, query string) ([]models.DataSet, error) {
	filter := bson.M{
		"$or": []bson.M{
			{"name": bson.M{"$regex": query, "$options": "i"}},
			{"tags": bson.M{"$regex": query, "$options": "i"}},
			{"description": bson.M{"$regex": query, "$options": "i"}},
			{"datasource": bson.M{"$regex": query, "$options": "i"}},
		},
	}

	results, err := ds.docDB.FindMany(ctx, CollectionDataSets, filter, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to search datasets: %w", err)
	}

	var datasets []models.DataSet
	for _, result := range results {
		doc := bson.M(result)
		var d models.DataSet
		data, err := bson.Marshal(doc)
		if err != nil {
			continue
		}
		if err := bson.Unmarshal(data, &d); err != nil {
			continue
		}
		datasets = append(datasets, d)
	}

	return datasets, nil
}
