package services

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/mdaxf/iac/models"
	"gorm.io/gorm"
)

// QueryTemplateService handles query template management
type QueryTemplateService struct {
	db *gorm.DB
}

// NewQueryTemplateService creates a new query template service
func NewQueryTemplateService(db *gorm.DB) *QueryTemplateService {
	return &QueryTemplateService{db: db}
}

// CreateTemplate creates a new query template
func (s *QueryTemplateService) CreateTemplate(ctx context.Context, template *models.QueryTemplate) error {
	if err := s.db.WithContext(ctx).Create(template).Error; err != nil {
		return fmt.Errorf("failed to create query template: %w", err)
	}

	return nil
}

// GetTemplate retrieves a query template by ID
func (s *QueryTemplateService) GetTemplate(ctx context.Context, id string) (*models.QueryTemplate, error) {
	var template models.QueryTemplate

	if err := s.db.WithContext(ctx).First(&template, "id = ?", id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("query template not found")
		}
		return nil, fmt.Errorf("failed to get query template: %w", err)
	}

	return &template, nil
}

// ListTemplates retrieves all query templates for a database
func (s *QueryTemplateService) ListTemplates(ctx context.Context, databaseAlias string) ([]models.QueryTemplate, error) {
	var templates []models.QueryTemplate

	query := s.db.WithContext(ctx).Order("templatename")

	if databaseAlias != "" {
		query = query.Where("databasealias = ?", databaseAlias)
	}

	if err := query.Find(&templates).Error; err != nil {
		return nil, fmt.Errorf("failed to list query templates: %w", err)
	}

	return templates, nil
}

// UpdateTemplate updates a query template
func (s *QueryTemplateService) UpdateTemplate(ctx context.Context, id string, updates map[string]interface{}) error {
	updates["modifiedon"] = gorm.Expr("CURRENT_TIMESTAMP")

	if err := s.db.WithContext(ctx).
		Model(&models.QueryTemplate{}).
		Where("id = ?", id).
		Updates(updates).Error; err != nil {
		return fmt.Errorf("failed to update query template: %w", err)
	}

	return nil
}

// DeleteTemplate deletes a query template
func (s *QueryTemplateService) DeleteTemplate(ctx context.Context, id string) error {
	if err := s.db.WithContext(ctx).Delete(&models.QueryTemplate{}, "id = ?", id).Error; err != nil {
		return fmt.Errorf("failed to delete query template: %w", err)
	}

	return nil
}

// SearchTemplates searches query templates by keyword
func (s *QueryTemplateService) SearchTemplates(ctx context.Context, databaseAlias, keyword string) ([]models.QueryTemplate, error) {
	var templates []models.QueryTemplate

	searchPattern := "%" + keyword + "%"

	query := s.db.WithContext(ctx).
		Where("templatename LIKE ? OR COALESCE(description, '') LIKE ? OR query_pattern LIKE ?",
			searchPattern, searchPattern, searchPattern)

	if databaseAlias != "" {
		query = query.Where("databasealias = ?", databaseAlias)
	}

	if err := query.Order("templatename").Find(&templates).Error; err != nil {
		return nil, fmt.Errorf("failed to search query templates: %w", err)
	}

	return templates, nil
}

// GetTemplatesByIntent searches templates by natural language intent
func (s *QueryTemplateService) GetTemplatesByIntent(ctx context.Context, databaseAlias, intent string) ([]models.QueryTemplate, error) {
	var templates []models.QueryTemplate

	// Extract keywords from intent
	keywords := extractKeywords(intent)

	if len(keywords) == 0 {
		return nil, nil
	}

	// Build search condition
	var conditions []string
	var args []interface{}

	for _, keyword := range keywords {
		searchPattern := "%" + keyword + "%"
		conditions = append(conditions, "(templatename LIKE ? OR COALESCE(description, '') LIKE ? OR COALESCE(examplequestions, '') LIKE ?)")
		args = append(args, searchPattern, searchPattern, searchPattern)
	}

	query := s.db.WithContext(ctx).Where(strings.Join(conditions, " OR "), args...)

	if databaseAlias != "" {
		query = query.Where("databasealias = ?", databaseAlias)
	}

	if err := query.Order("templatename").Find(&templates).Error; err != nil {
		return nil, fmt.Errorf("failed to get templates by intent: %w", err)
	}

	return templates, nil
}

// IncrementUsageCount increments the usage count of a template
func (s *QueryTemplateService) IncrementUsageCount(ctx context.Context, id string) error {
	if err := s.db.WithContext(ctx).
		Model(&models.QueryTemplate{}).
		Where("id = ?", id).
		UpdateColumn("usagecount", gorm.Expr("usagecount + 1")).
		UpdateColumn("lastrunat", gorm.Expr("CURRENT_TIMESTAMP")).
		Error; err != nil {
		return fmt.Errorf("failed to increment usage count: %w", err)
	}

	return nil
}

// GetTemplateContext builds query template context for AI
func (s *QueryTemplateService) GetTemplateContext(ctx context.Context, databaseAlias string, limit int) (string, error) {
	var templates []models.QueryTemplate

	query := s.db.WithContext(ctx).Where("databasealias = ?", databaseAlias).
		Order("usagecount DESC, templatename")

	if limit > 0 {
		query = query.Limit(limit)
	}

	if err := query.Find(&templates).Error; err != nil {
		return "", fmt.Errorf("failed to get template context: %w", err)
	}

	if len(templates) == 0 {
		return "", nil
	}

	context := "Query Templates (Common Patterns):\n\n"

	for _, template := range templates {
		context += fmt.Sprintf("## %s\n", template.TemplateName)

		if template.Description != "" {
			context += fmt.Sprintf("Description: %s\n", template.Description)
		}

		context += fmt.Sprintf("Pattern:\n```sql\n%s\n```\n", template.SQLTemplate)

		if template.ExampleQuestions != nil && len(template.ExampleQuestions) > 0 {
			questionsJSON, _ := json.Marshal(template.ExampleQuestions)
			context += fmt.Sprintf("Example Questions: %s\n", string(questionsJSON))
		}

		context += "\n"
	}

	return context, nil
}

// GetCategories retrieves distinct categories for a database
func (s *QueryTemplateService) GetCategories(ctx context.Context, databaseAlias string) ([]string, error) {
	var categories []string

	query := s.db.WithContext(ctx).
		Model(&models.QueryTemplate{}).
		Distinct().
		Where("category IS NOT NULL AND category != ''")

	if databaseAlias != "" {
		query = query.Where("databasealias = ?", databaseAlias)
	}

	if err := query.Pluck("category", &categories).Error; err != nil {
		return nil, fmt.Errorf("failed to get categories: %w", err)
	}

	return categories, nil
}

// extractKeywords extracts important keywords from a string
func extractKeywords(text string) []string {
	// Simple keyword extraction - remove common words
	stopWords := map[string]bool{
		"the": true, "a": true, "an": true, "and": true, "or": true,
		"but": true, "in": true, "on": true, "at": true, "to": true,
		"for": true, "of": true, "with": true, "by": true, "from": true,
		"is": true, "are": true, "was": true, "were": true, "be": true,
		"been": true, "being": true, "have": true, "has": true, "had": true,
		"do": true, "does": true, "did": true, "will": true, "would": true,
		"could": true, "should": true, "may": true, "might": true, "can": true,
		"what": true, "when": true, "where": true, "which": true, "who": true,
		"how": true, "show": true, "me": true, "get": true, "give": true,
	}

	words := strings.Fields(strings.ToLower(text))
	var keywords []string

	for _, word := range words {
		// Remove punctuation
		word = strings.Trim(word, ".,!?;:\"'")

		// Skip if too short or is a stop word
		if len(word) <= 2 || stopWords[word] {
			continue
		}

		keywords = append(keywords, word)
	}

	return keywords
}
