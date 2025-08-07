package entity

import (
	"context"
	"fmt"
	"strings"

	"github.com/ynl/greensoulai/internal/memory"
	"github.com/ynl/greensoulai/internal/memory/storage"
	"github.com/ynl/greensoulai/pkg/events"
	"github.com/ynl/greensoulai/pkg/logger"
)

// EntityMemory 实体记忆实现
// 用于管理实体及其关系的结构化信息
type EntityMemory struct {
	*memory.BaseMemory
	memoryProvider string
}

// EntityMemoryItem 实体记忆项
type EntityMemoryItem struct {
	memory.MemoryItem
	EntityName    string                 `json:"entity_name"`
	EntityType    string                 `json:"entity_type"`
	Relationships []EntityRelation       `json:"relationships"`
	Attributes    map[string]interface{} `json:"attributes"`
}

// EntityRelation 实体关系
type EntityRelation struct {
	RelationType string  `json:"relation_type"`
	TargetEntity string  `json:"target_entity"`
	Strength     float64 `json:"strength"`  // 关系强度 0-1
	Direction    string  `json:"direction"` // bidirectional, forward, backward
	Description  string  `json:"description,omitempty"`
}

// EntityType 实体类型
type EntityType string

const (
	EntityTypePerson       EntityType = "person"
	EntityTypeOrganization EntityType = "organization"
	EntityTypeLocation     EntityType = "location"
	EntityTypeEvent        EntityType = "event"
	EntityTypeConcept      EntityType = "concept"
	EntityTypeResource     EntityType = "resource"
	EntityTypeTask         EntityType = "task"
	EntityTypeAgent        EntityType = "agent"
)

// RelationType 关系类型
type RelationType string

const (
	RelationWorksFor     RelationType = "works_for"
	RelationLocatedIn    RelationType = "located_in"
	RelationPartOf       RelationType = "part_of"
	RelationRelatedTo    RelationType = "related_to"
	RelationDependsOn    RelationType = "depends_on"
	RelationCreatedBy    RelationType = "created_by"
	RelationAssignedTo   RelationType = "assigned_to"
	RelationCollaborates RelationType = "collaborates_with"
)

// NewEntityMemory 创建实体记忆实例
func NewEntityMemory(crew interface{}, embedderConfig *memory.EmbedderConfig, memStorage memory.MemoryStorage, path string, eventBus events.EventBus, logger logger.Logger) *EntityMemory {
	var memoryProvider string
	var storageInstance memory.MemoryStorage

	// 根据配置选择存储提供者
	if embedderConfig != nil {
		memoryProvider = embedderConfig.Provider
	}

	if memoryProvider == "mem0" {
		// 如果使用mem0存储
		logger.Info("using mem0 storage for entity memory")
		storageInstance = storage.NewMem0Storage("entity", crew, embedderConfig.Config, logger)
	} else {
		// 默认使用RAG存储
		if memStorage != nil {
			storageInstance = memStorage
		} else {
			logger.Info("using default RAG storage for entity memory")
			storageInstance = storage.NewRAGStorage("entities", embedderConfig, crew, path, logger)
		}
	}

	baseMemory := memory.NewBaseMemory(storageInstance, eventBus, logger)

	return &EntityMemory{
		BaseMemory:     baseMemory,
		memoryProvider: memoryProvider,
	}
}

// Save 保存实体记忆项（覆盖基类方法以处理实体特定逻辑）
func (em *EntityMemory) Save(ctx context.Context, value interface{}, metadata map[string]interface{}, agent string) error {
	// 为实体记忆添加特定的元数据
	if metadata == nil {
		metadata = make(map[string]interface{})
	}

	metadata["memory_type"] = "entity"
	metadata["provider"] = em.memoryProvider

	return em.BaseMemory.Save(ctx, value, metadata, agent)
}

// SaveEntity 保存单个实体
func (em *EntityMemory) SaveEntity(ctx context.Context, entityName, entityType string, attributes map[string]interface{}, relationships []EntityRelation, agent string) error {
	item := EntityMemoryItem{
		MemoryItem: memory.MemoryItem{
			Value: fmt.Sprintf("Entity: %s (Type: %s)", entityName, entityType),
		},
		EntityName:    entityName,
		EntityType:    entityType,
		Relationships: relationships,
		Attributes:    attributes,
	}

	metadata := map[string]interface{}{
		"entity_name":   entityName,
		"entity_type":   entityType,
		"attributes":    attributes,
		"relationships": relationships,
	}

	return em.Save(ctx, item, metadata, agent)
}

// SaveRelationship 保存实体关系
func (em *EntityMemory) SaveRelationship(ctx context.Context, sourceEntity, targetEntity string, relationType RelationType, strength float64, description string, agent string) error {
	relationshipData := fmt.Sprintf("Relationship: %s -[%s]-> %s (Strength: %.2f)", sourceEntity, relationType, targetEntity, strength)

	metadata := map[string]interface{}{
		"source_entity": sourceEntity,
		"target_entity": targetEntity,
		"relation_type": string(relationType),
		"strength":      strength,
		"description":   description,
		"relationship":  true,
	}

	return em.Save(ctx, relationshipData, metadata, agent)
}

// SearchEntities 搜索实体
func (em *EntityMemory) SearchEntities(ctx context.Context, query string, entityType string, limit int, scoreThreshold float64) ([]memory.MemoryItem, error) {
	// 如果指定了实体类型，修改查询
	searchQuery := query
	if entityType != "" {
		searchQuery = fmt.Sprintf("%s entity_type:%s", query, entityType)
	}

	// 执行基础搜索
	results, err := em.BaseMemory.Search(ctx, searchQuery, limit*2, scoreThreshold)
	if err != nil {
		return nil, err
	}

	// 过滤出实体相关的记忆
	var filteredResults []memory.MemoryItem
	for _, item := range results {
		if item.Metadata != nil {
			// 检查是否是实体记忆
			if memType, ok := item.Metadata["memory_type"].(string); ok && memType == "entity" {
				// 如果指定了实体类型，进一步过滤
				if entityType == "" {
					filteredResults = append(filteredResults, item)
				} else if itemType, ok := item.Metadata["entity_type"].(string); ok && itemType == entityType {
					filteredResults = append(filteredResults, item)
				}

				if len(filteredResults) >= limit {
					break
				}
			}
		}
	}

	return filteredResults, nil
}

// SearchRelationships 搜索关系
func (em *EntityMemory) SearchRelationships(ctx context.Context, sourceEntity, targetEntity string, relationType string, limit int) ([]memory.MemoryItem, error) {
	// 构建关系查询
	var queryParts []string

	if sourceEntity != "" {
		queryParts = append(queryParts, fmt.Sprintf("source_entity:%s", sourceEntity))
	}

	if targetEntity != "" {
		queryParts = append(queryParts, fmt.Sprintf("target_entity:%s", targetEntity))
	}

	if relationType != "" {
		queryParts = append(queryParts, fmt.Sprintf("relation_type:%s", relationType))
	}

	query := strings.Join(queryParts, " ")
	if query == "" {
		query = "relationship:true" // 搜索所有关系
	}

	// 执行搜索
	results, err := em.BaseMemory.Search(ctx, query, limit*2, 0.1)
	if err != nil {
		return nil, err
	}

	// 过滤出关系记忆
	var filteredResults []memory.MemoryItem
	for _, item := range results {
		if item.Metadata != nil {
			if isRel, ok := item.Metadata["relationship"].(bool); ok && isRel {
				filteredResults = append(filteredResults, item)
				if len(filteredResults) >= limit {
					break
				}
			}
		}
	}

	return filteredResults, nil
}

// GetEntityAttributes 获取实体属性
func (em *EntityMemory) GetEntityAttributes(ctx context.Context, entityName string) (map[string]interface{}, error) {
	// 搜索指定实体
	results, err := em.SearchEntities(ctx, entityName, "", 1, 0.8)
	if err != nil {
		return nil, err
	}

	if len(results) == 0 {
		return nil, fmt.Errorf("entity not found: %s", entityName)
	}

	// 从元数据中提取属性
	if attributes, ok := results[0].Metadata["attributes"].(map[string]interface{}); ok {
		return attributes, nil
	}

	return make(map[string]interface{}), nil
}

// GetEntityRelationships 获取实体关系
func (em *EntityMemory) GetEntityRelationships(ctx context.Context, entityName string, relationType string) ([]EntityRelation, error) {
	// 搜索指定实体的关系
	results, err := em.SearchRelationships(ctx, entityName, "", relationType, 50)
	if err != nil {
		return nil, err
	}

	var relationships []EntityRelation
	for _, item := range results {
		if item.Metadata != nil {
			if targetEntity, ok := item.Metadata["target_entity"].(string); ok {
				if relType, ok := item.Metadata["relation_type"].(string); ok {
					strength, _ := item.Metadata["strength"].(float64)
					description, _ := item.Metadata["description"].(string)

					relationship := EntityRelation{
						RelationType: relType,
						TargetEntity: targetEntity,
						Strength:     strength,
						Direction:    "forward",
						Description:  description,
					}
					relationships = append(relationships, relationship)
				}
			}
		}
	}

	return relationships, nil
}

// UpdateEntityAttributes 更新实体属性
func (em *EntityMemory) UpdateEntityAttributes(ctx context.Context, entityName string, newAttributes map[string]interface{}, agent string) error {
	// 获取当前属性
	currentAttributes, err := em.GetEntityAttributes(ctx, entityName)
	if err != nil {
		// 如果实体不存在，创建新实体
		return em.SaveEntity(ctx, entityName, "unknown", newAttributes, nil, agent)
	}

	// 合并属性
	for key, value := range newAttributes {
		currentAttributes[key] = value
	}

	// 保存更新后的实体
	return em.SaveEntity(ctx, entityName, "unknown", currentAttributes, nil, agent)
}

// AnalyzeEntityNetwork 分析实体网络（基础实现）
func (em *EntityMemory) AnalyzeEntityNetwork(ctx context.Context, entityName string, depth int) (interface{}, error) {
	// 获取实体的直接关系
	relationships, err := em.GetEntityRelationships(ctx, entityName, "")
	if err != nil {
		return nil, err
	}

	network := map[string]interface{}{
		"center_entity":      entityName,
		"direct_connections": len(relationships),
		"relationships":      relationships,
		"analysis_depth":     depth,
	}

	// 如果需要更深层次的分析，递归获取二级关系
	if depth > 1 {
		secondaryConnections := make(map[string][]EntityRelation)
		for _, rel := range relationships {
			secondaryRels, err := em.GetEntityRelationships(ctx, rel.TargetEntity, "")
			if err == nil {
				secondaryConnections[rel.TargetEntity] = secondaryRels
			}
		}
		network["secondary_connections"] = secondaryConnections
	}

	return network, nil
}
