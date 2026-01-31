package services

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/mdaxf/iac/documents"
	"github.com/mdaxf/iac/logger"
	"github.com/mdaxf/iac/models"
	"go.mongodb.org/mongo-driver/bson"
)

// Collection names for hierarchical hub
const (
	CollectionHubInstances      = "Integration_Hub_Instances"
	CollectionHubProtocolGroups = "Integration_Hub_Protocol_Groups"
	CollectionHubEndpoints      = "Integration_Hub_Endpoints"
	CollectionHubRoutes         = "Integration_Hub_Routes"
)

// HubService provides hierarchical hub management using MongoDB
type HubService struct {
	docDB documents.DocumentDB
	iLog  logger.Log
}

// NewHubService creates new hub service with document database
func NewHubService(docDB documents.DocumentDB) *HubService {
	return &HubService{
		docDB: docDB,
		iLog:  logger.Log{ModuleName: logger.Framework, User: "System", ControllerName: "HubService"},
	}
}

// =====================================================
// Hub Instance Methods
// =====================================================

// CreateInstance creates a new hub instance in MongoDB
func (hs *HubService) CreateInstance(ctx context.Context, instance *models.HubInstance) error {
	startTime := time.Now()
	defer func() {
		hs.iLog.Debug(fmt.Sprintf("CreateInstance completed in %v", time.Since(startTime)))
	}()

	if instance.ID == "" {
		instance.ID = uuid.New().String()
	}

	if instance.CreatedOn.IsZero() {
		instance.CreatedOn = time.Now()
	}
	instance.ModifiedOn = time.Now()
	instance.RowVersionStamp = 1
	instance.Active = true

	// Convert to BSON document
	doc := bson.M{
		"_id":              instance.ID,
		"instance_id":      instance.InstanceID,
		"name":             instance.Name,
		"description":      instance.Description,
		"enabled":          instance.Enabled,
		"metadata":         instance.Metadata,
		"active":           instance.Active,
		"referenceid":      instance.ReferenceID,
		"createdby":        instance.CreatedBy,
		"createdon":        instance.CreatedOn,
		"modifiedby":       instance.ModifiedBy,
		"modifiedon":       instance.ModifiedOn,
		"rowversionstamp":  instance.RowVersionStamp,
	}

	_, err := hs.docDB.InsertOne(ctx, CollectionHubInstances, doc)
	if err != nil {
		hs.iLog.Error(fmt.Sprintf("Failed to create hub instance: %v", err))
		return fmt.Errorf("failed to create hub instance: %w", err)
	}

	hs.iLog.Info(fmt.Sprintf("Created hub instance: %s", instance.ID))
	return nil
}

// GetInstanceByID retrieves a hub instance by ID from MongoDB
func (hs *HubService) GetInstanceByID(ctx context.Context, id string) (*models.HubInstance, error) {
	filter := bson.M{"_id": id}

	result, err := hs.docDB.FindOne(ctx, CollectionHubInstances, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to get hub instance: %w", err)
	}

	var instance models.HubInstance
	bsonBytes, err := bson.Marshal(result)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal result: %w", err)
	}

	err = bson.Unmarshal(bsonBytes, &instance)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal instance: %w", err)
	}

	// Map _id to ID
	if objID, ok := result["_id"].(string); ok {
		instance.ID = objID
	}

	return &instance, nil
}

// ListInstances retrieves all hub instances from MongoDB
func (hs *HubService) ListInstances(ctx context.Context) ([]models.HubInstance, error) {
	filter := bson.M{"active": true}

	results, err := hs.docDB.FindMany(ctx, CollectionHubInstances, filter, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to list hub instances: %w", err)
	}

	instances := make([]models.HubInstance, 0, len(results))
	for _, result := range results {
		var instance models.HubInstance
		bsonBytes, err := bson.Marshal(result)
		if err != nil {
			hs.iLog.Error(fmt.Sprintf("Failed to marshal result: %v", err))
			continue
		}

		err = bson.Unmarshal(bsonBytes, &instance)
		if err != nil {
			hs.iLog.Error(fmt.Sprintf("Failed to unmarshal instance: %v", err))
			continue
		}

		// Map _id to ID
		if objID, ok := result["_id"].(string); ok {
			instance.ID = objID
		}

		instances = append(instances, instance)
	}

	return instances, nil
}

// UpdateInstance updates an existing hub instance in MongoDB
func (hs *HubService) UpdateInstance(ctx context.Context, instance *models.HubInstance) error {
	instance.ModifiedOn = time.Now()
	instance.RowVersionStamp++

	// Build update document
	update := bson.M{
		"$set": bson.M{
			"instance_id":      instance.InstanceID,
			"name":             instance.Name,
			"description":      instance.Description,
			"enabled":          instance.Enabled,
			"metadata":         instance.Metadata,
			"active":           instance.Active,
			"referenceid":      instance.ReferenceID,
			"modifiedby":       instance.ModifiedBy,
			"modifiedon":       instance.ModifiedOn,
			"rowversionstamp":  instance.RowVersionStamp,
		},
	}

	filter := bson.M{"_id": instance.ID}

	err := hs.docDB.UpdateOne(ctx, CollectionHubInstances, filter, update)
	if err != nil {
		hs.iLog.Error(fmt.Sprintf("Failed to update hub instance: %v", err))
		return fmt.Errorf("failed to update hub instance: %w", err)
	}

	hs.iLog.Info(fmt.Sprintf("Updated hub instance: %s", instance.ID))
	return nil
}

// DeleteInstance soft deletes a hub instance from MongoDB
func (hs *HubService) DeleteInstance(ctx context.Context, id string) error {
	update := bson.M{
		"$set": bson.M{
			"active":     false,
			"modifiedon": time.Now(),
		},
	}

	filter := bson.M{"_id": id}

	err := hs.docDB.UpdateOne(ctx, CollectionHubInstances, filter, update)
	if err != nil {
		hs.iLog.Error(fmt.Sprintf("Failed to delete hub instance: %v", err))
		return fmt.Errorf("failed to delete hub instance: %w", err)
	}

	hs.iLog.Info(fmt.Sprintf("Deleted hub instance: %s", id))
	return nil
}

// =====================================================
// Protocol Group Methods
// =====================================================

// CreateProtocolGroup creates a new protocol group in MongoDB
func (hs *HubService) CreateProtocolGroup(ctx context.Context, group *models.HubProtocolGroup) error {
	if group.ID == "" {
		group.ID = uuid.New().String()
	}

	if group.CreatedOn.IsZero() {
		group.CreatedOn = time.Now()
	}
	group.ModifiedOn = time.Now()
	group.RowVersionStamp = 1
	group.Active = true

	doc := bson.M{
		"_id":              group.ID,
		"instance_id":      group.InstanceID,
		"direction":        group.Direction,
		"name":             group.Name,
		"description":      group.Description,
		"protocol":         group.Protocol,
		"base_config":      group.BaseConfig,
		"message_type":     group.MessageType,
		"timeout":          group.Timeout,
		"retry_attempts":   group.RetryAttempts,
		"retry_interval":   group.RetryInterval,
		"auth_type":        group.AuthType,
		"auth_config":      group.AuthConfig,
		"enabled":          group.Enabled,
		"metadata":         group.Metadata,
		"active":           group.Active,
		"referenceid":      group.ReferenceID,
		"createdby":        group.CreatedBy,
		"createdon":        group.CreatedOn,
		"modifiedby":       group.ModifiedBy,
		"modifiedon":       group.ModifiedOn,
		"rowversionstamp":  group.RowVersionStamp,
	}

	_, err := hs.docDB.InsertOne(ctx, CollectionHubProtocolGroups, doc)
	if err != nil {
		hs.iLog.Error(fmt.Sprintf("Failed to create protocol group: %v", err))
		return fmt.Errorf("failed to create protocol group: %w", err)
	}

	hs.iLog.Info(fmt.Sprintf("Created protocol group: %s", group.ID))
	return nil
}

// ListProtocolGroups retrieves protocol groups for an instance from MongoDB
func (hs *HubService) ListProtocolGroups(ctx context.Context, instanceID string, direction string) ([]models.HubProtocolGroup, error) {
	filter := bson.M{
		"instance_id": instanceID,
		"active":      true,
	}

	if direction != "" {
		filter["direction"] = direction
	}

	results, err := hs.docDB.FindMany(ctx, CollectionHubProtocolGroups, filter, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to list protocol groups: %w", err)
	}

	groups := make([]models.HubProtocolGroup, 0, len(results))
	for _, result := range results {
		var group models.HubProtocolGroup
		bsonBytes, err := bson.Marshal(result)
		if err != nil {
			hs.iLog.Error(fmt.Sprintf("Failed to marshal result: %v", err))
			continue
		}

		err = bson.Unmarshal(bsonBytes, &group)
		if err != nil {
			hs.iLog.Error(fmt.Sprintf("Failed to unmarshal group: %v", err))
			continue
		}

		if objID, ok := result["_id"].(string); ok {
			group.ID = objID
		}

		groups = append(groups, group)
	}

	return groups, nil
}

// UpdateProtocolGroup updates an existing protocol group in MongoDB
func (hs *HubService) UpdateProtocolGroup(ctx context.Context, group *models.HubProtocolGroup) error {
	group.ModifiedOn = time.Now()
	group.RowVersionStamp++

	update := bson.M{
		"$set": bson.M{
			"instance_id":      group.InstanceID,
			"direction":        group.Direction,
			"name":             group.Name,
			"description":      group.Description,
			"protocol":         group.Protocol,
			"base_config":      group.BaseConfig,
			"message_type":     group.MessageType,
			"timeout":          group.Timeout,
			"retry_attempts":   group.RetryAttempts,
			"retry_interval":   group.RetryInterval,
			"auth_type":        group.AuthType,
			"auth_config":      group.AuthConfig,
			"enabled":          group.Enabled,
			"metadata":         group.Metadata,
			"active":           group.Active,
			"referenceid":      group.ReferenceID,
			"modifiedby":       group.ModifiedBy,
			"modifiedon":       group.ModifiedOn,
			"rowversionstamp":  group.RowVersionStamp,
		},
	}

	filter := bson.M{"_id": group.ID}

	err := hs.docDB.UpdateOne(ctx, CollectionHubProtocolGroups, filter, update)
	if err != nil {
		hs.iLog.Error(fmt.Sprintf("Failed to update protocol group: %v", err))
		return fmt.Errorf("failed to update protocol group: %w", err)
	}

	hs.iLog.Info(fmt.Sprintf("Updated protocol group: %s", group.ID))
	return nil
}

// DeleteProtocolGroup soft deletes a protocol group from MongoDB
func (hs *HubService) DeleteProtocolGroup(ctx context.Context, id string) error {
	update := bson.M{
		"$set": bson.M{
			"active":     false,
			"modifiedon": time.Now(),
		},
	}

	filter := bson.M{"_id": id}

	err := hs.docDB.UpdateOne(ctx, CollectionHubProtocolGroups, filter, update)
	if err != nil {
		hs.iLog.Error(fmt.Sprintf("Failed to delete protocol group: %v", err))
		return fmt.Errorf("failed to delete protocol group: %w", err)
	}

	hs.iLog.Info(fmt.Sprintf("Deleted protocol group: %s", id))
	return nil
}

// =====================================================
// Endpoint Methods
// =====================================================

// CreateEndpoint creates a new endpoint in MongoDB
func (hs *HubService) CreateEndpoint(ctx context.Context, endpoint *models.HubEndpoint) error {
	if endpoint.ID == "" {
		endpoint.ID = uuid.New().String()
	}

	if endpoint.CreatedOn.IsZero() {
		endpoint.CreatedOn = time.Now()
	}
	endpoint.ModifiedOn = time.Now()
	endpoint.RowVersionStamp = 1
	endpoint.Active = true

	doc := bson.M{
		"_id":                endpoint.ID,
		"protocol_group_id":  endpoint.ProtocolGroupID,
		"name":               endpoint.Name,
		"description":        endpoint.Description,
		"endpoint_url":       endpoint.EndpointURL,
		"port":               endpoint.Port,
		"path":               endpoint.Path,
		"method":             endpoint.Method,
		"override_config":    endpoint.OverrideConfig,
		"message_type":       endpoint.MessageType,
		"timeout":            endpoint.Timeout,
		"retry_attempts":     endpoint.RetryAttempts,
		"retry_interval":     endpoint.RetryInterval,
		"auth_type":          endpoint.AuthType,
		"auth_config":        endpoint.AuthConfig,
		"queue_config":       endpoint.QueueConfig,
		"file_config":        endpoint.FileConfig,
		"validate_schema":    endpoint.ValidateSchema,
		"schema_definition":  endpoint.SchemaDefinition,
		"enabled":            endpoint.Enabled,
		"metadata":           endpoint.Metadata,
		"active":             endpoint.Active,
		"referenceid":        endpoint.ReferenceID,
		"createdby":          endpoint.CreatedBy,
		"createdon":          endpoint.CreatedOn,
		"modifiedby":         endpoint.ModifiedBy,
		"modifiedon":         endpoint.ModifiedOn,
		"rowversionstamp":    endpoint.RowVersionStamp,
	}

	_, err := hs.docDB.InsertOne(ctx, CollectionHubEndpoints, doc)
	if err != nil {
		hs.iLog.Error(fmt.Sprintf("Failed to create endpoint: %v", err))
		return fmt.Errorf("failed to create endpoint: %w", err)
	}

	hs.iLog.Info(fmt.Sprintf("Created endpoint: %s", endpoint.ID))
	return nil
}

// GetEndpointByID retrieves an endpoint by ID from MongoDB
func (hs *HubService) GetEndpointByID(ctx context.Context, id string) (*models.HubEndpoint, error) {
	filter := bson.M{"_id": id}

	result, err := hs.docDB.FindOne(ctx, CollectionHubEndpoints, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to get endpoint: %w", err)
	}

	var endpoint models.HubEndpoint
	bsonBytes, err := bson.Marshal(result)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal result: %w", err)
	}

	err = bson.Unmarshal(bsonBytes, &endpoint)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal endpoint: %w", err)
	}

	if objID, ok := result["_id"].(string); ok {
		endpoint.ID = objID
	}

	return &endpoint, nil
}

// ListEndpoints retrieves endpoints for a protocol group from MongoDB
func (hs *HubService) ListEndpoints(ctx context.Context, groupID string) ([]models.HubEndpoint, error) {
	filter := bson.M{
		"protocol_group_id": groupID,
		"active":            true,
	}

	results, err := hs.docDB.FindMany(ctx, CollectionHubEndpoints, filter, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to list endpoints: %w", err)
	}

	endpoints := make([]models.HubEndpoint, 0, len(results))
	for _, result := range results {
		var endpoint models.HubEndpoint
		bsonBytes, err := bson.Marshal(result)
		if err != nil {
			hs.iLog.Error(fmt.Sprintf("Failed to marshal result: %v", err))
			continue
		}

		err = bson.Unmarshal(bsonBytes, &endpoint)
		if err != nil {
			hs.iLog.Error(fmt.Sprintf("Failed to unmarshal endpoint: %v", err))
			continue
		}

		if objID, ok := result["_id"].(string); ok {
			endpoint.ID = objID
		}

		endpoints = append(endpoints, endpoint)
	}

	return endpoints, nil
}

// UpdateEndpoint updates an existing endpoint in MongoDB
func (hs *HubService) UpdateEndpoint(ctx context.Context, endpoint *models.HubEndpoint) error {
	endpoint.ModifiedOn = time.Now()
	endpoint.RowVersionStamp++

	update := bson.M{
		"$set": bson.M{
			"protocol_group_id":  endpoint.ProtocolGroupID,
			"name":               endpoint.Name,
			"description":        endpoint.Description,
			"endpoint_url":       endpoint.EndpointURL,
			"port":               endpoint.Port,
			"path":               endpoint.Path,
			"method":             endpoint.Method,
			"override_config":    endpoint.OverrideConfig,
			"message_type":       endpoint.MessageType,
			"timeout":            endpoint.Timeout,
			"retry_attempts":     endpoint.RetryAttempts,
			"retry_interval":     endpoint.RetryInterval,
			"auth_type":          endpoint.AuthType,
			"auth_config":        endpoint.AuthConfig,
			"queue_config":       endpoint.QueueConfig,
			"file_config":        endpoint.FileConfig,
			"validate_schema":    endpoint.ValidateSchema,
			"schema_definition":  endpoint.SchemaDefinition,
			"enabled":            endpoint.Enabled,
			"metadata":           endpoint.Metadata,
			"active":             endpoint.Active,
			"referenceid":        endpoint.ReferenceID,
			"modifiedby":         endpoint.ModifiedBy,
			"modifiedon":         endpoint.ModifiedOn,
			"rowversionstamp":    endpoint.RowVersionStamp,
		},
	}

	filter := bson.M{"_id": endpoint.ID}

	err := hs.docDB.UpdateOne(ctx, CollectionHubEndpoints, filter, update)
	if err != nil {
		hs.iLog.Error(fmt.Sprintf("Failed to update endpoint: %v", err))
		return fmt.Errorf("failed to update endpoint: %w", err)
	}

	hs.iLog.Info(fmt.Sprintf("Updated endpoint: %s", endpoint.ID))
	return nil
}

// DeleteEndpoint soft deletes an endpoint from MongoDB
func (hs *HubService) DeleteEndpoint(ctx context.Context, id string) error {
	update := bson.M{
		"$set": bson.M{
			"active":     false,
			"modifiedon": time.Now(),
		},
	}

	filter := bson.M{"_id": id}

	err := hs.docDB.UpdateOne(ctx, CollectionHubEndpoints, filter, update)
	if err != nil {
		hs.iLog.Error(fmt.Sprintf("Failed to delete endpoint: %v", err))
		return fmt.Errorf("failed to delete endpoint: %w", err)
	}

	hs.iLog.Info(fmt.Sprintf("Deleted endpoint: %s", id))
	return nil
}

// =====================================================
// Configuration Resolution
// =====================================================

// ResolveEndpointConfig resolves final configuration with inheritance
func (hs *HubService) ResolveEndpointConfig(ctx context.Context, endpointID string) (*models.ResolvedEndpointConfig, error) {
	// Get endpoint
	endpoint, err := hs.GetEndpointByID(ctx, endpointID)
	if err != nil {
		return nil, err
	}

	// Get protocol group
	groupFilter := bson.M{"_id": endpoint.ProtocolGroupID}
	groupResult, err := hs.docDB.FindOne(ctx, CollectionHubProtocolGroups, groupFilter)
	if err != nil {
		return nil, fmt.Errorf("protocol group not found")
	}

	var group models.HubProtocolGroup
	bsonBytes, _ := bson.Marshal(groupResult)
	bson.Unmarshal(bsonBytes, &group)
	if objID, ok := groupResult["_id"].(string); ok {
		group.ID = objID
	}

	// Get instance
	instanceFilter := bson.M{"_id": group.InstanceID}
	instanceResult, err := hs.docDB.FindOne(ctx, CollectionHubInstances, instanceFilter)
	if err != nil {
		return nil, fmt.Errorf("instance not found")
	}

	var instance models.HubInstance
	bsonBytes, _ = bson.Marshal(instanceResult)
	bson.Unmarshal(bsonBytes, &instance)
	if objID, ok := instanceResult["_id"].(string); ok {
		instance.ID = objID
	}

	// Resolve configuration with inheritance
	resolved := &models.ResolvedEndpointConfig{
		Endpoint:      endpoint,
		ProtocolGroup: &group,
		Instance:      &instance,
	}

	// Resolve timeout
	if endpoint.Timeout > 0 {
		resolved.FinalTimeout = endpoint.Timeout
	} else {
		resolved.FinalTimeout = group.Timeout
	}

	// Resolve retry attempts
	if endpoint.RetryAttempts > 0 {
		resolved.FinalRetryAttempts = endpoint.RetryAttempts
	} else {
		resolved.FinalRetryAttempts = group.RetryAttempts
	}

	// Resolve retry interval
	if endpoint.RetryInterval > 0 {
		resolved.FinalRetryInterval = endpoint.RetryInterval
	} else {
		resolved.FinalRetryInterval = group.RetryInterval
	}

	// Resolve message type
	if endpoint.MessageType != "" {
		resolved.FinalMessageType = endpoint.MessageType
	} else {
		resolved.FinalMessageType = group.MessageType
	}

	// Resolve auth type
	if endpoint.AuthType != "" {
		resolved.FinalAuthType = endpoint.AuthType
	} else {
		resolved.FinalAuthType = group.AuthType
	}

	// Resolve auth config
	if len(endpoint.AuthConfig) > 0 {
		resolved.FinalAuthConfig = endpoint.AuthConfig
	} else {
		resolved.FinalAuthConfig = group.AuthConfig
	}

	// Merge configurations: base_config <- override_config
	resolved.FinalConfig = make(map[string]interface{})
	for k, v := range group.BaseConfig {
		resolved.FinalConfig[k] = v
	}
	for k, v := range endpoint.OverrideConfig {
		resolved.FinalConfig[k] = v
	}

	return resolved, nil
}

// =====================================================
// Hierarchical Tree
// =====================================================

// GetHierarchicalTree builds hierarchical tree structure for an instance
func (hs *HubService) GetHierarchicalTree(ctx context.Context, instanceID string) (*models.HierarchicalTreeNode, error) {
	// Get instance
	instance, err := hs.GetInstanceByID(ctx, instanceID)
	if err != nil {
		return nil, err
	}

	// Build tree root
	root := &models.HierarchicalTreeNode{
		ID:    instance.ID,
		Type:  "instance",
		Label: instance.Name,
		Icon:  "🏢",
		Data:  instance,
		Children: []*models.HierarchicalTreeNode{
			{
				ID:    fmt.Sprintf("%s-inbound", instance.ID),
				Type:  "direction",
				Label: "Inbound",
				Icon:  "📥",
				Data: models.DirectionNode{
					Direction:  "inbound",
					InstanceID: instance.ID,
				},
				Children: make([]*models.HierarchicalTreeNode, 0),
			},
			{
				ID:    fmt.Sprintf("%s-outbound", instance.ID),
				Type:  "direction",
				Label: "Outbound",
				Icon:  "📤",
				Data: models.DirectionNode{
					Direction:  "outbound",
					InstanceID: instance.ID,
				},
				Children: make([]*models.HierarchicalTreeNode, 0),
			},
		},
	}

	// Load protocol groups
	inboundGroups, _ := hs.ListProtocolGroups(ctx, instanceID, "inbound")
	outboundGroups, _ := hs.ListProtocolGroups(ctx, instanceID, "outbound")

	// Add inbound groups to tree
	for _, group := range inboundGroups {
		groupNode := &models.HierarchicalTreeNode{
			ID:       group.ID,
			Type:     "protocol-group",
			Label:    fmt.Sprintf("%s - %s", group.Protocol, group.Name),
			Icon:     "📋",
			Data:     group,
			Children: make([]*models.HierarchicalTreeNode, 0),
		}

		// Load endpoints for this group
		endpoints, _ := hs.ListEndpoints(ctx, group.ID)
		for _, endpoint := range endpoints {
			endpointNode := &models.HierarchicalTreeNode{
				ID:    endpoint.ID,
				Type:  "endpoint",
				Label: endpoint.Name,
				Icon:  "🔗",
				Data:  endpoint,
			}
			groupNode.Children = append(groupNode.Children, endpointNode)
		}

		root.Children[0].Children = append(root.Children[0].Children, groupNode)
	}

	// Add outbound groups to tree
	for _, group := range outboundGroups {
		groupNode := &models.HierarchicalTreeNode{
			ID:       group.ID,
			Type:     "protocol-group",
			Label:    fmt.Sprintf("%s - %s", group.Protocol, group.Name),
			Icon:     "📋",
			Data:     group,
			Children: make([]*models.HierarchicalTreeNode, 0),
		}

		// Load endpoints for this group
		endpoints, _ := hs.ListEndpoints(ctx, group.ID)
		for _, endpoint := range endpoints {
			endpointNode := &models.HierarchicalTreeNode{
				ID:    endpoint.ID,
				Type:  "endpoint",
				Label: endpoint.Name,
				Icon:  "🔗",
				Data:  endpoint,
			}
			groupNode.Children = append(groupNode.Children, endpointNode)
		}

		root.Children[1].Children = append(root.Children[1].Children, groupNode)
	}

	return root, nil
}
