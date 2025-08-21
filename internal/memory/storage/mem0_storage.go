package storage

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/ynl/greensoulai/internal/memory"
	"github.com/ynl/greensoulai/pkg/logger"
)

// Mem0Storage Mem0外部记忆存储实现
// 基于crewAI的实现逻辑，支持云端API和本地模式
type Mem0Storage struct {
	storageType string
	crew        interface{}
	config      map[string]interface{}
	logger      logger.Logger

	// Mem0 API相关配置
	apiKey           string
	endpoint         string
	projectID        string
	orgID            string
	userID           string
	agentID          string
	runID            string
	includes         string
	excludes         string
	customCategories []map[string]interface{}
	infer            bool

	// HTTP客户端
	httpClient  *http.Client
	isCloudMode bool // true: 使用MemoryClient（云端API），false: 本地模式
}

// NewMem0Storage 创建Mem0存储实例
// 基于crewAI的实现逻辑
func NewMem0Storage(storageType string, crew interface{}, config map[string]interface{}, logger logger.Logger) *Mem0Storage {
	storage := &Mem0Storage{
		storageType: storageType,
		crew:        crew,
		config:      config,
		logger:      logger,
		infer:       true, // 默认启用推理
		httpClient:  &http.Client{Timeout: 30 * time.Second},
	}

	// 验证存储类型
	storage.validateType(storageType)

	// 提取配置值
	storage.extractConfigValues()

	// 初始化内存客户端
	storage.initializeMemory()

	return storage
}

// validateType 验证存储类型（参照crewAI逻辑）
func (m *Mem0Storage) validateType(storageType string) {
	supportedTypes := map[string]bool{
		"short_term": true,
		"long_term":  true,
		"entities":   true,
		"external":   true,
	}

	if !supportedTypes[storageType] {
		m.logger.Error("Invalid type for Mem0Storage",
			logger.Field{Key: "type", Value: storageType},
			logger.Field{Key: "supported", Value: []string{"short_term", "long_term", "entities", "external"}},
		)
		// 在Go中，我们不抛出异常，而是记录错误并使用默认值
		m.storageType = "short_term"
	}
}

// extractConfigValues 提取配置值（参照crewAI逻辑）
func (m *Mem0Storage) extractConfigValues() {
	if m.config != nil {
		if runID, ok := m.config["run_id"].(string); ok {
			m.runID = runID
		}
		if includes, ok := m.config["includes"].(string); ok {
			m.includes = includes
		}
		if excludes, ok := m.config["excludes"].(string); ok {
			m.excludes = excludes
		}
		if customCategories, ok := m.config["custom_categories"].([]map[string]interface{}); ok {
			m.customCategories = customCategories
		}
		if infer, ok := m.config["infer"].(bool); ok {
			m.infer = infer
		}
		if userID, ok := m.config["user_id"].(string); ok {
			m.userID = userID
		}
		if agentID, ok := m.config["agent_id"].(string); ok {
			m.agentID = agentID
		}
	}
}

// initializeMemory 初始化内存客户端（参照crewAI逻辑）
func (m *Mem0Storage) initializeMemory() {
	// 优先从配置获取，然后从环境变量获取
	apiKey := ""
	if m.config != nil {
		if key, ok := m.config["api_key"].(string); ok {
			apiKey = key
		}
	}
	if apiKey == "" {
		apiKey = os.Getenv("MEM0_API_KEY")
	}

	if apiKey != "" {
		// 云端模式：使用MemoryClient
		m.apiKey = apiKey
		m.isCloudMode = true

		if m.config != nil {
			if orgID, ok := m.config["org_id"].(string); ok {
				m.orgID = orgID
			}
			if projectID, ok := m.config["project_id"].(string); ok {
				m.projectID = projectID
			}
		}

		// 设置默认端点
		if m.endpoint == "" {
			m.endpoint = "https://api.mem0.ai/v1"
		}

		m.logger.Info("initialized Mem0 cloud mode",
			logger.Field{Key: "endpoint", Value: m.endpoint},
			logger.Field{Key: "has_org_id", Value: m.orgID != ""},
			logger.Field{Key: "has_project_id", Value: m.projectID != ""},
		)
	} else {
		// 本地模式：使用Memory
		m.isCloudMode = false
		m.logger.Info("initialized Mem0 local mode")
	}
}

// Save 保存记忆项到Mem0（基于crewAI实现逻辑）
func (m *Mem0Storage) Save(ctx context.Context, item memory.MemoryItem) error {
	if !m.isCloudMode {
		return fmt.Errorf("Mem0 local mode not implemented in Go version")
	}

	// 根据crewAI逻辑，使用assistant消息格式
	assistantMessage := []map[string]interface{}{
		{
			"role":    "assistant",
			"content": fmt.Sprintf("%v", item.Value),
		},
	}

	// 构建基础元数据映射（参照crewAI逻辑）
	baseMetadata := map[string]string{
		"short_term": "short_term",
		"long_term":  "long_term",
		"entities":   "entity",
		"external":   "external",
	}

	// 构建请求参数
	payload := map[string]interface{}{
		"messages": assistantMessage,
		"metadata": map[string]interface{}{
			"type": baseMetadata[m.storageType],
		},
		"infer": m.infer,
	}

	// 添加item的元数据
	if metadataMap, ok := payload["metadata"].(map[string]interface{}); ok && item.Metadata != nil {
		for k, v := range item.Metadata {
			metadataMap[k] = v
		}
	}

	// 云端模式特定参数
	if m.includes != "" {
		payload["includes"] = m.includes
	}
	if m.excludes != "" {
		payload["excludes"] = m.excludes
	}
	payload["output_format"] = "v1.1"
	payload["version"] = "v2"

	// 短期记忆的run_id
	if m.storageType == "short_term" && m.runID != "" {
		payload["run_id"] = m.runID
	}

	// 用户和代理ID
	if m.userID != "" {
		payload["user_id"] = m.userID
	}
	if m.agentID != "" {
		payload["agent_id"] = m.agentID
	} else if agentName := m.getAgentName(); agentName != "" {
		payload["agent_id"] = agentName
	}

	// 发送HTTP请求
	return m.sendMemoryRequest(ctx, "POST", "/memories", payload)
}

// Search 从Mem0搜索记忆项（基于crewAI实现逻辑）
func (m *Mem0Storage) Search(ctx context.Context, query string, limit int, scoreThreshold float64) ([]memory.MemoryItem, error) {
	if !m.isCloudMode {
		return nil, fmt.Errorf("Mem0 local mode not implemented in Go version")
	}

	// 构建搜索参数（参照crewAI逻辑）
	params := map[string]interface{}{
		"query":         query,
		"limit":         limit,
		"version":       "v2",
		"output_format": "v1.1",
		"threshold":     scoreThreshold,
	}

	if m.userID != "" {
		params["user_id"] = m.userID
	}

	// 内存类型映射（参照crewAI逻辑）
	memoryTypeMap := map[string]map[string]interface{}{
		"short_term": {"type": "short_term"},
		"long_term":  {"type": "long_term"},
		"entities":   {"type": "entity"},
		"external":   {"type": "external"},
	}

	if typeData, ok := memoryTypeMap[m.storageType]; ok {
		params["metadata"] = typeData
		if m.storageType == "short_term" && m.runID != "" {
			params["run_id"] = m.runID
		}
	}

	// 创建过滤器
	params["filters"] = m.createFilterForSearch()

	// 发送搜索请求
	respData, err := m.sendSearchRequest(ctx, params)
	if err != nil {
		return nil, err
	}

	// 解析响应并转换为memory.MemoryItem
	return m.parseSearchResponse(respData)
}

// Delete 从Mem0删除记忆项
func (m *Mem0Storage) Delete(ctx context.Context, id string) error {
	if !m.isCloudMode {
		return fmt.Errorf("Mem0 local mode not implemented in Go version")
	}

	// 发送删除请求
	return m.sendMemoryRequest(ctx, "DELETE", "/memories/"+id, nil)
}

// Clear 清除Mem0中的所有记忆项
func (m *Mem0Storage) Clear(ctx context.Context) error {
	if !m.isCloudMode {
		return fmt.Errorf("Mem0 local mode not implemented in Go version")
	}

	m.logger.Info("clearing all memories in Mem0",
		logger.Field{Key: "type", Value: m.storageType},
	)

	// 发送清除请求（如果API支持批量删除）
	return m.sendMemoryRequest(ctx, "DELETE", "/memories", nil)
}

// Close 关闭Mem0存储连接
func (m *Mem0Storage) Close() error {
	m.logger.Info("closing Mem0 storage connection",
		logger.Field{Key: "type", Value: m.storageType},
	)

	// Mem0通常是HTTP API，不需要特殊的关闭操作
	return nil
}

// GetConfig 获取Mem0配置信息
func (m *Mem0Storage) GetConfig() map[string]interface{} {
	return map[string]interface{}{
		"storage_type":   m.storageType,
		"endpoint":       m.endpoint,
		"project_id":     m.projectID,
		"user_id":        m.userID,
		"api_configured": m.apiKey != "",
	}
}

// IsConfigured 检查Mem0是否已正确配置
func (m *Mem0Storage) IsConfigured() bool {
	return m.apiKey != "" && m.userID != ""
}

// TestConnection 测试Mem0连接
func (m *Mem0Storage) TestConnection(ctx context.Context) error {
	if !m.IsConfigured() {
		return fmt.Errorf("Mem0 storage not configured - missing api_key or user_id")
	}

	m.logger.Info("testing Mem0 connection",
		logger.Field{Key: "endpoint", Value: m.endpoint},
		logger.Field{Key: "user_id", Value: m.userID},
	)

	// TODO: 实现连接测试
	// 这里应该发送一个测试请求到Mem0 API
	// 例如获取用户信息或执行健康检查

	// 发送测试连接请求
	return m.sendMemoryRequest(ctx, "GET", "/memories", nil)
}

// 辅助方法：createFilterForSearch 创建搜索过滤器（参照crewAI逻辑）
func (m *Mem0Storage) createFilterForSearch() map[string]interface{} {
	filter := make(map[string]interface{})

	if m.storageType == "short_term" && m.runID != "" {
		andConditions := []map[string]interface{}{
			{"run_id": m.runID},
		}
		filter["AND"] = andConditions
	} else {
		userID := m.userID
		agentID := m.agentID
		if agentID == "" {
			agentID = m.getAgentName()
		}

		if userID != "" && agentID != "" {
			orConditions := []map[string]interface{}{
				{"user_id": userID},
				{"agent_id": agentID},
			}
			filter["OR"] = orConditions
		} else if userID != "" {
			andConditions := []map[string]interface{}{
				{"user_id": userID},
			}
			filter["AND"] = andConditions
		} else if agentID != "" {
			andConditions := []map[string]interface{}{
				{"agent_id": agentID},
			}
			filter["AND"] = andConditions
		}
	}

	return filter
}

// 辅助方法：getAgentName 获取代理名称（参照crewAI逻辑）
func (m *Mem0Storage) getAgentName() string {
	if m.crew == nil {
		return ""
	}

	// 尝试从crew中提取agents信息
	// 这里需要根据实际的crew结构来实现
	// 简化实现，返回空字符串
	return ""
}

// 辅助方法：sendMemoryRequest 发送HTTP请求到Mem0 API
func (m *Mem0Storage) sendMemoryRequest(ctx context.Context, method, path string, payload map[string]interface{}) error {
	url := m.endpoint + path

	var reqBody io.Reader
	if payload != nil {
		jsonData, err := json.Marshal(payload)
		if err != nil {
			return fmt.Errorf("failed to marshal payload: %w", err)
		}
		reqBody = bytes.NewBuffer(jsonData)
	}

	req, err := http.NewRequestWithContext(ctx, method, url, reqBody)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	// 设置请求头
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+m.apiKey)
	if m.orgID != "" {
		req.Header.Set("X-Org-ID", m.orgID)
	}
	if m.projectID != "" {
		req.Header.Set("X-Project-ID", m.projectID)
	}

	// 发送请求
	resp, err := m.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			m.logger.Error("Failed to close response body",
				logger.Field{Key: "error", Value: err})
		}
	}()

	// 检查响应状态码
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("Mem0 API request failed with status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	m.logger.Debug("Mem0 API request successful",
		logger.Field{Key: "method", Value: method},
		logger.Field{Key: "path", Value: path},
		logger.Field{Key: "status", Value: resp.StatusCode},
	)

	return nil
}

// 辅助方法：sendSearchRequest 发送搜索请求到Mem0 API
func (m *Mem0Storage) sendSearchRequest(ctx context.Context, params map[string]interface{}) (map[string]interface{}, error) {
	url := m.endpoint + "/memories/search"

	jsonData, err := json.Marshal(params)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal search params: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create search request: %w", err)
	}

	// 设置请求头
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+m.apiKey)
	if m.orgID != "" {
		req.Header.Set("X-Org-ID", m.orgID)
	}
	if m.projectID != "" {
		req.Header.Set("X-Project-ID", m.projectID)
	}

	// 发送请求
	resp, err := m.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send search request: %w", err)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			m.logger.Error("Failed to close response body",
				logger.Field{Key: "error", Value: err})
		}
	}()

	// 读取响应体
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	// 检查响应状态码
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("Mem0 search request failed with status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	// 解析JSON响应
	var respData map[string]interface{}
	if err := json.Unmarshal(bodyBytes, &respData); err != nil {
		return nil, fmt.Errorf("failed to parse search response: %w", err)
	}

	return respData, nil
}

// 辅助方法：parseSearchResponse 解析搜索响应为memory.MemoryItem
func (m *Mem0Storage) parseSearchResponse(respData map[string]interface{}) ([]memory.MemoryItem, error) {
	results, ok := respData["results"].([]interface{})
	if !ok {
		return []memory.MemoryItem{}, nil
	}

	memoryItems := make([]memory.MemoryItem, 0, len(results))

	for _, result := range results {
		resultMap, ok := result.(map[string]interface{})
		if !ok {
			continue
		}

		// 提取记忆内容（参照crewAI逻辑，添加context字段）
		memoryContent := ""
		if memory, ok := resultMap["memory"].(string); ok {
			memoryContent = memory
			// 为了与crewAI兼容，添加context字段
			resultMap["context"] = memory
		}

		// 提取分数
		score := 0.0
		if scoreVal, ok := resultMap["score"].(float64); ok {
			score = scoreVal
		}

		// 提取ID
		id := ""
		if idVal, ok := resultMap["id"].(string); ok {
			id = idVal
		}

		// 提取元数据
		metadata := make(map[string]interface{})
		if metaVal, ok := resultMap["metadata"].(map[string]interface{}); ok {
			metadata = metaVal
		}

		// 创建MemoryItem
		memoryItem := memory.MemoryItem{
			ID:       id,
			Value:    memoryContent,
			Metadata: metadata,
		}

		// 添加分数到元数据中
		if memoryItem.Metadata == nil {
			memoryItem.Metadata = make(map[string]interface{})
		}
		memoryItem.Metadata["score"] = score
		memoryItem.Metadata["context"] = memoryContent

		memoryItems = append(memoryItems, memoryItem)
	}

	m.logger.Debug("parsed Mem0 search results",
		logger.Field{Key: "count", Value: len(memoryItems)},
	)

	return memoryItems, nil
}
