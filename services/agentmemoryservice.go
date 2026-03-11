package services

import (
	"context"
	"database/sql"
	"fmt"
	"sort"
	"strings"
	"time"
	"unicode"

	"github.com/google/uuid"
	"github.com/mdaxf/iac/llm"
	"github.com/mdaxf/iac/logger"
	"github.com/mdaxf/iac/models"
	"gorm.io/gorm"
)

// ─── Request / option types ───────────────────────────────────────────────────

// CreateMemoryRequest is the payload for creating a new memory item.
type CreateMemoryRequest struct {
	Title         string   `json:"title"`
	Summary       string   `json:"summary"`
	Content       string   `json:"content"`
	ContentType   string   `json:"contenttype"`
	Tags          []string `json:"tags"`
	Layer         string   `json:"layer"`         // L0 | L1 | L2
	Priority      string   `json:"priority"`      // P0 | P1 | P2
	RetentionType string   `json:"retentiontype"` // permanent | interval | temporary
	RetentionDays *int     `json:"retentiondays"`
	ExpiresAt     *string  `json:"expiresat"` // RFC3339 string, optional
	CreatedBy     string   `json:"createdby"`
}

// UpdateMemoryRequest is the payload for updating an existing memory item.
type UpdateMemoryRequest struct {
	Title         *string  `json:"title"`
	Summary       *string  `json:"summary"`
	Content       *string  `json:"content"`
	ContentType   *string  `json:"contenttype"`
	Tags          []string `json:"tags"`
	Layer         *string  `json:"layer"`
	Priority      *string  `json:"priority"`
	RetentionType *string  `json:"retentiontype"`
	RetentionDays *int     `json:"retentiondays"`
	ExpiresAt     *string  `json:"expiresat"`
}

// ListMemoryOptions filters for ListMemory.
type ListMemoryOptions struct {
	Layer        string // L0 | L1 | L2 | "" (all)
	Priority     string // P0 | P1 | P2 | "" (all)
	ShowArchived bool
	SortBy       string // priority | createdon | access_count
}

// CleanOptions controls bulk-clean behaviour.
type CleanOptions struct {
	Priority     string // optional filter
	Layer        string // optional filter
	ArchivedOnly bool   // when true, hard-delete archived items; otherwise soft-delete matches
}

// ─── Service ─────────────────────────────────────────────────────────────────

// AgentMemoryService manages the 3-layer memory store for agents.
type AgentMemoryService struct {
	db    *gorm.DB
	sqlDB *sql.DB
	iLog  logger.Log
}

var (
	globalAgentMemoryService *AgentMemoryService
)

func GetGlobalAgentMemoryService() *AgentMemoryService { return globalAgentMemoryService }
func SetGlobalAgentMemoryService(s *AgentMemoryService) { globalAgentMemoryService = s }

func NewAgentMemoryService(db *gorm.DB, sqlDB *sql.DB) *AgentMemoryService {
	return &AgentMemoryService{
		db:    db,
		sqlDB: sqlDB,
		iLog: logger.Log{
			ModuleName:     logger.Framework,
			User:           "System",
			ControllerName: "AgentMemoryService",
		},
	}
}

// ─── CRUD (API-facing) ────────────────────────────────────────────────────────

// CreateMemory creates a new memory item for an agent.
func (s *AgentMemoryService) CreateMemory(agentID string, req CreateMemoryRequest) (*models.AgentMemory, error) {
	layer := req.Layer
	if layer == "" {
		layer = "L1"
	}
	priority := req.Priority
	if priority == "" {
		priority = "P1"
	}
	retType := req.RetentionType
	if retType == "" {
		retType = "permanent"
	}
	contentType := req.ContentType
	if contentType == "" {
		contentType = "text"
	}

	mem := &models.AgentMemory{
		ID:                uuid.New().String(),
		AgentDefinitionID: agentID,
		Title:             req.Title,
		Summary:           req.Summary,
		Content:           req.Content,
		ContentType:       contentType,
		Tags:              models.StringSlice(req.Tags),
		Layer:             layer,
		Priority:          priority,
		RetentionType:     retType,
		RetentionDays:     req.RetentionDays,
		Active:            true,
		Archived:          false,
		CreatedBy:         req.CreatedBy,
		ModifiedBy:        req.CreatedBy,
	}

	// Parse optional expires_at
	if req.ExpiresAt != nil && *req.ExpiresAt != "" {
		t, err := time.Parse(time.RFC3339, *req.ExpiresAt)
		if err == nil {
			mem.ExpiresAt = &sql.NullTime{Time: t, Valid: true}
		}
	}

	if err := s.db.Create(mem).Error; err != nil {
		return nil, fmt.Errorf("create memory: %w", err)
	}
	return mem, nil
}

// ListMemory returns memory items for an agent with optional filters.
func (s *AgentMemoryService) ListMemory(agentID string, opts ListMemoryOptions) ([]models.AgentMemory, error) {
	query := s.db.Where("agent_definition_id = ? AND active = ?", agentID, true)

	if !opts.ShowArchived {
		query = query.Where("archived = ?", false)
	}
	if opts.Layer != "" {
		query = query.Where("layer = ?", opts.Layer)
	}
	if opts.Priority != "" {
		query = query.Where("priority = ?", opts.Priority)
	}

	switch opts.SortBy {
	case "access_count":
		query = query.Order("access_count DESC, createdon DESC")
	case "createdon":
		query = query.Order("createdon DESC")
	default: // priority
		query = query.Order("CASE priority WHEN 'P0' THEN 0 WHEN 'P1' THEN 1 ELSE 2 END, createdon DESC")
	}

	var items []models.AgentMemory
	if err := query.Find(&items).Error; err != nil {
		return nil, fmt.Errorf("list memory: %w", err)
	}
	return items, nil
}

// GetMemory returns a single memory item by ID.
func (s *AgentMemoryService) GetMemory(id string) (*models.AgentMemory, error) {
	var mem models.AgentMemory
	if err := s.db.Where("id = ? AND active = ?", id, true).First(&mem).Error; err != nil {
		return nil, err
	}
	return &mem, nil
}

// UpdateMemory updates mutable fields of a memory item.
func (s *AgentMemoryService) UpdateMemory(id string, req UpdateMemoryRequest, by string) (*models.AgentMemory, error) {
	updates := map[string]interface{}{
		"modifiedby":      by,
		"modifiedon":      time.Now(),
		"rowversionstamp": gorm.Expr("rowversionstamp + 1"),
	}
	if req.Title != nil {
		updates["title"] = *req.Title
	}
	if req.Summary != nil {
		updates["summary"] = *req.Summary
	}
	if req.Content != nil {
		updates["content"] = *req.Content
	}
	if req.ContentType != nil {
		updates["content_type"] = *req.ContentType
	}
	if req.Tags != nil {
		updates["tags"] = models.StringSlice(req.Tags)
	}
	if req.Layer != nil {
		updates["layer"] = *req.Layer
	}
	if req.Priority != nil {
		updates["priority"] = *req.Priority
	}
	if req.RetentionType != nil {
		updates["retention_type"] = *req.RetentionType
	}
	if req.RetentionDays != nil {
		updates["retention_days"] = *req.RetentionDays
	}
	if req.ExpiresAt != nil && *req.ExpiresAt != "" {
		t, err := time.Parse(time.RFC3339, *req.ExpiresAt)
		if err == nil {
			updates["expires_at"] = t
		}
	}

	if err := s.db.Model(&models.AgentMemory{}).Where("id = ?", id).Updates(updates).Error; err != nil {
		return nil, fmt.Errorf("update memory: %w", err)
	}
	return s.GetMemory(id)
}

// DeleteMemory soft-deletes a memory item (sets active=false).
func (s *AgentMemoryService) DeleteMemory(id, by string) error {
	return s.db.Model(&models.AgentMemory{}).Where("id = ?", id).Updates(map[string]interface{}{
		"active":     false,
		"modifiedby": by,
		"modifiedon": time.Now(),
	}).Error
}

// ArchiveMemory marks a memory item as archived.
func (s *AgentMemoryService) ArchiveMemory(id, by string) error {
	return s.db.Model(&models.AgentMemory{}).Where("id = ?", id).Updates(map[string]interface{}{
		"archived":   true,
		"modifiedby": by,
		"modifiedon": time.Now(),
	}).Error
}

// CleanMemory bulk-removes or archives memory items for an agent.
// When opts.ArchivedOnly=true, it hard-deletes archived items.
// Otherwise it soft-deletes active items matching the priority/layer filters.
func (s *AgentMemoryService) CleanMemory(agentID string, opts CleanOptions) (int64, error) {
	query := s.db.Model(&models.AgentMemory{}).Where("agent_definition_id = ?", agentID)

	if opts.ArchivedOnly {
		query = query.Where("archived = ?", true)
		if opts.Priority != "" {
			query = query.Where("priority = ?", opts.Priority)
		}
		if opts.Layer != "" {
			query = query.Where("layer = ?", opts.Layer)
		}
		res := query.Delete(&models.AgentMemory{})
		return res.RowsAffected, res.Error
	}

	// Soft-delete matching active items
	conditions := []string{"active = ?"}
	args := []interface{}{true}
	if opts.Priority != "" {
		conditions = append(conditions, "priority = ?")
		args = append(args, opts.Priority)
	}
	if opts.Layer != "" {
		conditions = append(conditions, "layer = ?")
		args = append(args, opts.Layer)
	}
	fullWhere := strings.Join(conditions, " AND ")
	allArgs := append([]interface{}{agentID}, args...)
	res := s.db.Model(&models.AgentMemory{}).
		Where("agent_definition_id = ? AND "+fullWhere, allArgs...).
		Updates(map[string]interface{}{"active": false, "modifiedon": time.Now()})
	return res.RowsAffected, res.Error
}

// ─── Agent-facing (called during run) ────────────────────────────────────────

// SearchMemory performs a title/tag text search against active, non-archived memory.
// layer="" searches all layers; otherwise restricts to the given layer.
func (s *AgentMemoryService) SearchMemory(agentID, query, layer string) ([]models.AgentMemory, error) {
	q := s.db.Where("agent_definition_id = ? AND active = ? AND archived = ?", agentID, true, false)
	if layer != "" {
		q = q.Where("layer = ?", layer)
	}
	if query != "" {
		like := "%" + query + "%"
		q = q.Where("title ILIKE ? OR summary ILIKE ? OR tags::text ILIKE ?", like, like, like)
	}
	q = q.Order("CASE priority WHEN 'P0' THEN 0 WHEN 'P1' THEN 1 ELSE 2 END, access_count DESC")

	var items []models.AgentMemory
	if err := q.Find(&items).Error; err != nil {
		return nil, err
	}
	return items, nil
}

// ReadMemoryItem returns a memory item and increments its access_count + last_accessed_on.
func (s *AgentMemoryService) ReadMemoryItem(id string) (*models.AgentMemory, error) {
	mem, err := s.GetMemory(id)
	if err != nil {
		return nil, err
	}
	now := time.Now()
	_ = s.db.Model(&models.AgentMemory{}).Where("id = ?", id).Updates(map[string]interface{}{
		"access_count":    gorm.Expr("access_count + 1"),
		"last_accessed_on": now,
	})
	mem.AccessCount++
	mem.LastAccessedOn = &sql.NullTime{Time: now, Valid: true}
	return mem, nil
}

// SaveMemoryItem creates or upserts a memory item during an agent run.
func (s *AgentMemoryService) SaveMemoryItem(
	agentID, title, summary, content, layer, priority, retention string,
	retentionDays *int,
) (*models.AgentMemory, error) {
	return s.CreateMemory(agentID, CreateMemoryRequest{
		Title:         title,
		Summary:       summary,
		Content:       content,
		ContentType:   "text",
		Layer:         layer,
		Priority:      priority,
		RetentionType: retention,
		RetentionDays: retentionDays,
		CreatedBy:     "agent",
	})
}

// ─── Maintenance ─────────────────────────────────────────────────────────────

// ArchiveExpiredMemory archives items whose expires_at is in the past.
func (s *AgentMemoryService) ArchiveExpiredMemory() (int64, error) {
	res := s.db.Model(&models.AgentMemory{}).
		Where("expires_at IS NOT NULL AND expires_at < ? AND archived = ? AND active = ?", time.Now(), false, true).
		Updates(map[string]interface{}{
			"archived":   true,
			"modifiedon": time.Now(),
		})
	if res.RowsAffected > 0 {
		s.iLog.Info(fmt.Sprintf("ArchiveExpiredMemory: archived %d items", res.RowsAffected))
	}
	return res.RowsAffected, res.Error
}

// ArchiveIntervalMemory archives interval-retention items whose age exceeds retention_days.
func (s *AgentMemoryService) ArchiveIntervalMemory() (int64, error) {
	res := s.db.Model(&models.AgentMemory{}).
		Where("retention_type = 'interval' AND retention_days IS NOT NULL AND archived = ? AND active = ?", false, true).
		Where("createdon + (retention_days || ' days')::interval < ?", time.Now()).
		Updates(map[string]interface{}{
			"archived":   true,
			"modifiedon": time.Now(),
		})
	if res.RowsAffected > 0 {
		s.iLog.Info(fmt.Sprintf("ArchiveIntervalMemory: archived %d items", res.RowsAffected))
	}
	return res.RowsAffected, res.Error
}

// ─── Auto-memory: post-run save & pre-run retrieval ───────────────────────────

// SimilarMemoryResult holds a matched L1 memory and its similarity score.
type SimilarMemoryResult struct {
	L0    models.AgentMemory
	L1    models.AgentMemory
	Score float64
}

// AutoSaveRunMemory persists a completed agent run into the 3-layer memory store.
// L2 receives the full conversation history, L1 an LLM-generated summary, and
// L0 an index entry (keywords + reference to the L1 ID).
// This is intended to be called in a background goroutine — never blocks the run.
func (s *AgentMemoryService) AutoSaveRunMemory(
	ctx context.Context,
	agentID, runID, convID, taskPrompt, finalText, historyJSON string,
) error {
	startTime := time.Now()
	s.iLog.Info(fmt.Sprintf("AutoSaveRunMemory: start agent=%s run=%s", agentID, runID[:8]))

	if finalText == "" {
		s.iLog.Info("AutoSaveRunMemory: skipped — empty response")
		return nil
	}

	// Deduplication: skip when a near-identical run is already indexed (≥ 0.9).
	existing, _ := s.FindSimilarMemory(agentID, taskPrompt, 0.9)
	if len(existing) > 0 {
		s.iLog.Info(fmt.Sprintf("AutoSaveRunMemory: skipped — duplicate (score=%.2f)", existing[0].Score))
		return nil
	}

	topicTitle := memTruncate(taskPrompt, 80)
	runShort := runID
	if len(runShort) > 8 {
		runShort = runShort[:8]
	}

	// ── L2: full conversation content ──────────────────────────────────────────
	fullContent := fmt.Sprintf("Task: %s\n\nFinal Answer: %s\n\nFull History:\n%s",
		taskPrompt, finalText, historyJSON)

	l2, err := s.CreateMemory(agentID, CreateMemoryRequest{
		Title:         fmt.Sprintf("Run %s: %s", runShort, topicTitle),
		Summary:       memTruncate(finalText, 200),
		Content:       fullContent,
		ContentType:   "conversation",
		Tags:          []string{"auto", "run:" + runID, "conv:" + convID},
		Layer:         "L2",
		Priority:      "P2",
		RetentionType: "interval",
		RetentionDays: memIntPtr(90),
		CreatedBy:     "agent_auto",
	})
	if err != nil {
		return fmt.Errorf("AutoSaveRunMemory L2: %w", err)
	}

	// ── L1: LLM-generated summary ──────────────────────────────────────────────
	l1Content, err := s.summarizeWithLLM(ctx, taskPrompt, finalText)
	if err != nil {
		s.iLog.Warn(fmt.Sprintf("AutoSaveRunMemory: LLM summary failed, using truncated response: %v", err))
		l1Content = memTruncate(finalText, 500)
	}

	l1, err := s.CreateMemory(agentID, CreateMemoryRequest{
		Title:         topicTitle,
		Summary:       memTruncate(l1Content, 200),
		Content:       l1Content,
		ContentType:   "summary",
		Tags:          []string{"auto", "run:" + runID, "l2:" + l2.ID},
		Layer:         "L1",
		Priority:      "P1",
		RetentionType: "interval",
		RetentionDays: memIntPtr(180),
		CreatedBy:     "agent_auto",
	})
	if err != nil {
		return fmt.Errorf("AutoSaveRunMemory L1: %w", err)
	}

	// ── L0: keyword index pointing to L1 ───────────────────────────────────────
	keywords, err := s.extractKeywordsWithLLM(ctx, taskPrompt, l1Content)
	if err != nil {
		s.iLog.Warn(fmt.Sprintf("AutoSaveRunMemory: keyword extraction failed, using title: %v", err))
		keywords = topicTitle
	}

	_, err = s.CreateMemory(agentID, CreateMemoryRequest{
		Title:         keywords,
		Summary:       memTruncate(l1Content, 150),
		Content:       l1.ID, // L0.content stores the L1 ID for retrieval
		ContentType:   "index",
		Tags:          []string{"auto", "run:" + runID, "l1:" + l1.ID},
		Layer:         "L0",
		Priority:      "P1",
		RetentionType: "permanent",
		CreatedBy:     "agent_auto",
	})
	if err != nil {
		return fmt.Errorf("AutoSaveRunMemory L0: %w", err)
	}

	s.iLog.Info(fmt.Sprintf("AutoSaveRunMemory: saved L2=%s L1=%s in %v", l2.ID, l1.ID, time.Since(startTime)))
	return nil
}

// FindSimilarMemory searches the L0 index for entries similar to taskPrompt.
// For each L0 match above threshold it also validates the L1 summary similarity.
// Returns results sorted by score descending; only entries with score >= threshold in both layers are returned.
func (s *AgentMemoryService) FindSimilarMemory(agentID, taskPrompt string, threshold float64) ([]SimilarMemoryResult, error) {
	if s.db == nil || taskPrompt == "" {
		return nil, nil
	}

	// Use first ~10 significant words as the search term for the ILIKE query.
	searchTerms := memSearchTerms(taskPrompt)
	l0Items, err := s.SearchMemory(agentID, searchTerms, "L0")
	if err != nil {
		return nil, err
	}
	if len(l0Items) == 0 {
		return nil, nil
	}

	var results []SimilarMemoryResult
	for _, l0 := range l0Items {
		// Skip auto-saved L0 entries that don't store a valid L1 ID in content.
		l1ID := strings.TrimSpace(l0.Content)
		if l1ID == "" || strings.Contains(l1ID, " ") {
			continue
		}

		// Score against L0 title+summary (keyword index).
		l0Score := memJaccard(taskPrompt, l0.Title+" "+l0.Summary)
		if l0Score < threshold {
			continue
		}

		// Load L1 and validate similarity of the full summary.
		l1, err := s.GetMemory(l1ID)
		if err != nil {
			continue
		}
		l1Score := memJaccard(taskPrompt, l1.Content+" "+l1.Summary)
		if l1Score < threshold {
			continue
		}

		results = append(results, SimilarMemoryResult{
			L0:    l0,
			L1:    *l1,
			Score: (l0Score + l1Score) / 2,
		})
	}

	sort.Slice(results, func(i, j int) bool { return results[i].Score > results[j].Score })
	return results, nil
}

// ─── LLM helpers ─────────────────────────────────────────────────────────────

func (s *AgentMemoryService) summarizeWithLLM(ctx context.Context, taskPrompt, response string) (string, error) {
	input := "Task: " + taskPrompt + "\n\nResponse: " + memTruncate(response, 3000)
	msgs := []map[string]interface{}{
		{"role": "system", "content": "Summarize the following agent task and its response in 150-200 words. Focus on: what was asked, what approach was taken, what the result was, and any key facts. Be concise and factual."},
		{"role": "user", "content": input},
	}
	return llm.CallLLM(ctx, "agent_memory", "", "gpt-4o-mini", msgs, 0.3)
}

func (s *AgentMemoryService) extractKeywordsWithLLM(ctx context.Context, taskPrompt, summary string) (string, error) {
	input := taskPrompt + "\n\n" + memTruncate(summary, 500)
	msgs := []map[string]interface{}{
		{"role": "system", "content": "Extract 5-10 key topic keywords and phrases from the following task and summary. Return only the keywords separated by spaces, all lowercase, no punctuation. Example: database migration user authentication api endpoint error handling"},
		{"role": "user", "content": input},
	}
	return llm.CallLLM(ctx, "agent_memory", "", "gpt-4o-mini", msgs, 0.1)
}

// ─── Similarity utilities ─────────────────────────────────────────────────────

// memJaccard computes Jaccard similarity (word-set intersection/union) between two texts.
func memJaccard(a, b string) float64 {
	wa := memTokenize(a)
	wb := memTokenize(b)
	if len(wa) == 0 || len(wb) == 0 {
		return 0
	}
	setA := make(map[string]struct{}, len(wa))
	for _, w := range wa {
		setA[w] = struct{}{}
	}
	intersection := 0
	setB := make(map[string]struct{}, len(wb))
	for _, w := range wb {
		setB[w] = struct{}{}
		if _, ok := setA[w]; ok {
			intersection++
		}
	}
	union := len(setA) + len(setB) - intersection
	if union == 0 {
		return 0
	}
	return float64(intersection) / float64(union)
}

// memTokenize lowercases and splits text into significant words (len > 2).
func memTokenize(text string) []string {
	text = strings.ToLower(text)
	words := strings.FieldsFunc(text, func(r rune) bool {
		return !unicode.IsLetter(r) && !unicode.IsDigit(r)
	})
	result := make([]string, 0, len(words))
	for _, w := range words {
		if len(w) > 2 {
			result = append(result, w)
		}
	}
	return result
}

// memSearchTerms extracts the first 10 significant words as a search query.
func memSearchTerms(text string) string {
	words := memTokenize(text)
	if len(words) > 10 {
		words = words[:10]
	}
	return strings.Join(words, " ")
}

// memTruncate truncates s to at most n bytes.
func memTruncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n]
}

// memIntPtr returns a pointer to an int literal.
func memIntPtr(v int) *int { return &v }
