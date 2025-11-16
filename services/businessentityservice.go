package services

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/mdaxf/iac/models"
	"gorm.io/gorm"
)

// BusinessEntityService handles business entity management
type BusinessEntityService struct {
	db *gorm.DB
}

// NewBusinessEntityService creates a new business entity service
func NewBusinessEntityService(db *gorm.DB) *BusinessEntityService {
	return &BusinessEntityService{db: db}
}

// CreateEntity creates a new business entity
func (s *BusinessEntityService) CreateEntity(ctx context.Context, entity *models.BusinessEntity) error {
	if err := s.db.WithContext(ctx).Create(entity).Error; err != nil {
		return fmt.Errorf("failed to create business entity: %w", err)
	}

	return nil
}

// GetEntity retrieves a business entity by ID
func (s *BusinessEntityService) GetEntity(ctx context.Context, id string) (*models.BusinessEntity, error) {
	var entity models.BusinessEntity

	if err := s.db.WithContext(ctx).First(&entity, "id = ?", id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("business entity not found")
		}
		return nil, fmt.Errorf("failed to get business entity: %w", err)
	}

	return &entity, nil
}

// ListEntities retrieves all business entities for a database
func (s *BusinessEntityService) ListEntities(ctx context.Context, databaseAlias string) ([]models.BusinessEntity, error) {
	var entities []models.BusinessEntity

	query := s.db.WithContext(ctx).Order("entityname")

	if databaseAlias != "" {
		query = query.Where("databasealias = ?", databaseAlias)
	}

	if err := query.Find(&entities).Error; err != nil {
		return nil, fmt.Errorf("failed to list business entities: %w", err)
	}

	return entities, nil
}

// UpdateEntity updates a business entity
func (s *BusinessEntityService) UpdateEntity(ctx context.Context, id string, updates map[string]interface{}) error {
	updates["modifiedon"] = gorm.Expr("CURRENT_TIMESTAMP")

	if err := s.db.WithContext(ctx).
		Model(&models.BusinessEntity{}).
		Where("id = ?", id).
		Updates(updates).Error; err != nil {
		return fmt.Errorf("failed to update business entity: %w", err)
	}

	return nil
}

// DeleteEntity deletes a business entity
func (s *BusinessEntityService) DeleteEntity(ctx context.Context, id string) error {
	if err := s.db.WithContext(ctx).Delete(&models.BusinessEntity{}, "id = ?", id).Error; err != nil {
		return fmt.Errorf("failed to delete business entity: %w", err)
	}

	return nil
}

// SearchEntities searches business entities by keyword
func (s *BusinessEntityService) SearchEntities(ctx context.Context, databaseAlias, keyword string) ([]models.BusinessEntity, error) {
	var entities []models.BusinessEntity

	searchPattern := "%" + keyword + "%"

	query := s.db.WithContext(ctx).
		Where("entityname LIKE ? OR COALESCE(description, '') LIKE ? OR COALESCE(synonyms, '') LIKE ?",
			searchPattern, searchPattern, searchPattern)

	if databaseAlias != "" {
		query = query.Where("databasealias = ?", databaseAlias)
	}

	if err := query.Order("entityname").Find(&entities).Error; err != nil {
		return nil, fmt.Errorf("failed to search business entities: %w", err)
	}

	return entities, nil
}

// GetEntitiesByTable retrieves business entities mapped to a specific table
func (s *BusinessEntityService) GetEntitiesByTable(ctx context.Context, databaseAlias, tableName string) ([]models.BusinessEntity, error) {
	var entities []models.BusinessEntity

	mappingPattern := fmt.Sprintf("%%\"table_name\":\"%s\"%%", tableName)

	if err := s.db.WithContext(ctx).
		Where("databasealias = ? AND tablemappings LIKE ?", databaseAlias, mappingPattern).
		Order("entityname").
		Find(&entities).Error; err != nil {
		return nil, fmt.Errorf("failed to get entities by table: %w", err)
	}

	return entities, nil
}

// GetEntityContext builds business entity context for AI
func (s *BusinessEntityService) GetEntityContext(ctx context.Context, databaseAlias string, entityNames []string) (string, error) {
	var entities []models.BusinessEntity

	query := s.db.WithContext(ctx).Where("databasealias = ?", databaseAlias)

	if len(entityNames) > 0 {
		query = query.Where("entityname IN ?", entityNames)
	}

	if err := query.Order("entityname").Find(&entities).Error; err != nil {
		return "", fmt.Errorf("failed to get entity context: %w", err)
	}

	if len(entities) == 0 {
		return "", nil
	}

	context := "Business Entity Definitions:\n\n"

	for _, entity := range entities {
		context += fmt.Sprintf("## %s\n", entity.EntityName)

		if entity.Description != "" {
			context += fmt.Sprintf("Description: %s\n", entity.Description)
		}

		if entity.Synonyms != nil && len(entity.Synonyms) > 0 {
			synonymsJSON, _ := json.Marshal(entity.Synonyms)
			context += fmt.Sprintf("Synonyms: %s\n", string(synonymsJSON))
		}

		if entity.EntityType != "" {
			context += fmt.Sprintf("Type: %s\n", string(entity.EntityType))
		}

		context += "\n"
	}

	return context, nil
}
