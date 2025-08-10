package storage

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/ynl/greensoulai/pkg/logger"
)

// ChromaDBClient ChromaDB HTTP客户端
type ChromaDBClient struct {
	baseURL    string
	httpClient *http.Client
	logger     logger.Logger
	apiVersion string
}

// ChromaDBConfig ChromaDB配置
type ChromaDBConfig struct {
	Host       string        `json:"host" yaml:"host"`
	Port       int           `json:"port" yaml:"port"`
	Timeout    time.Duration `json:"timeout" yaml:"timeout"`
	APIVersion string        `json:"api_version" yaml:"api_version"`
}

// ChromaCollection ChromaDB集合信息
type ChromaCollection struct {
	Name     string                 `json:"name"`
	ID       string                 `json:"id,omitempty"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
	Tenant   string                 `json:"tenant,omitempty"`
	Database string                 `json:"database,omitempty"`
}

// ChromaDocument ChromaDB文档
type ChromaDocument struct {
	ID        string                 `json:"id"`
	Document  string                 `json:"document,omitempty"`
	Embedding []float64              `json:"embedding,omitempty"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
}

// ChromaQueryRequest ChromaDB查询请求
type ChromaQueryRequest struct {
	QueryEmbeddings [][]float64            `json:"query_embeddings,omitempty"`
	QueryTexts      []string               `json:"query_texts,omitempty"`
	NResults        int                    `json:"n_results,omitempty"`
	Where           map[string]interface{} `json:"where,omitempty"`
	WhereDocument   map[string]interface{} `json:"where_document,omitempty"`
	Include         []string               `json:"include,omitempty"`
}

// ChromaQueryResponse ChromaDB查询响应
type ChromaQueryResponse struct {
	IDs        [][]string                 `json:"ids"`
	Documents  [][]string                 `json:"documents"`
	Metadatas  [][]map[string]interface{} `json:"metadatas"`
	Distances  [][]float64                `json:"distances"`
	Embeddings [][]interface{}            `json:"embeddings,omitempty"`
}

// ChromaAddRequest ChromaDB添加请求
type ChromaAddRequest struct {
	Documents  []string                 `json:"documents,omitempty"`
	Embeddings [][]float64              `json:"embeddings,omitempty"`
	IDs        []string                 `json:"ids"`
	Metadatas  []map[string]interface{} `json:"metadatas,omitempty"`
}

// ChromaDeleteRequest ChromaDB删除请求
type ChromaDeleteRequest struct {
	IDs           []string               `json:"ids,omitempty"`
	Where         map[string]interface{} `json:"where,omitempty"`
	WhereDocument map[string]interface{} `json:"where_document,omitempty"`
}

// ChromaUpdateRequest ChromaDB更新请求
type ChromaUpdateRequest struct {
	IDs        []string                 `json:"ids"`
	Documents  []string                 `json:"documents,omitempty"`
	Embeddings [][]float64              `json:"embeddings,omitempty"`
	Metadatas  []map[string]interface{} `json:"metadatas,omitempty"`
}

// ChromaUpsertRequest ChromaDB插入或更新请求
type ChromaUpsertRequest struct {
	Documents  []string                 `json:"documents,omitempty"`
	Embeddings [][]float64              `json:"embeddings,omitempty"`
	IDs        []string                 `json:"ids"`
	Metadatas  []map[string]interface{} `json:"metadatas,omitempty"`
}

// NewChromaDBClient 创建ChromaDB客户端
func NewChromaDBClient(config *ChromaDBConfig, log logger.Logger) *ChromaDBClient {
	if config == nil {
		config = &ChromaDBConfig{
			Host:       "localhost",
			Port:       8000,
			Timeout:    30 * time.Second,
			APIVersion: "v1",
		}
	}

	// 设置默认值
	if config.Host == "" {
		config.Host = "localhost"
	}
	if config.Port == 0 {
		config.Port = 8000
	}
	if config.Timeout == 0 {
		config.Timeout = 30 * time.Second
	}
	if config.APIVersion == "" {
		config.APIVersion = "v1"
	}

	baseURL := fmt.Sprintf("http://%s:%d/api/%s", config.Host, config.Port, config.APIVersion)

	client := &ChromaDBClient{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: config.Timeout,
		},
		logger:     log,
		apiVersion: config.APIVersion,
	}

	log.Info("ChromaDB client initialized",
		logger.Field{Key: "base_url", Value: baseURL},
		logger.Field{Key: "timeout", Value: config.Timeout},
	)

	return client
}

// doRequest 执行HTTP请求
func (c *ChromaDBClient) doRequest(ctx context.Context, method, path string, body interface{}) (*http.Response, error) {
	url := fmt.Sprintf("%s%s", c.baseURL, path)

	var reqBody io.Reader
	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request body: %w", err)
		}
		reqBody = bytes.NewReader(jsonBody)
	}

	req, err := http.NewRequestWithContext(ctx, method, url, reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	c.logger.Debug("making ChromaDB request",
		logger.Field{Key: "method", Value: method},
		logger.Field{Key: "url", Value: url},
	)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}

	return resp, nil
}

// parseResponse 解析响应
func (c *ChromaDBClient) parseResponse(resp *http.Response, result interface{}) error {
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode >= 400 {
		c.logger.Error("ChromaDB request failed",
			logger.Field{Key: "status", Value: resp.StatusCode},
			logger.Field{Key: "body", Value: string(body)},
		)
		return fmt.Errorf("ChromaDB request failed with status %d: %s", resp.StatusCode, string(body))
	}

	if result != nil {
		if err := json.Unmarshal(body, result); err != nil {
			return fmt.Errorf("failed to unmarshal response: %w", err)
		}
	}

	return nil
}

// Heartbeat 健康检查
func (c *ChromaDBClient) Heartbeat(ctx context.Context) error {
	resp, err := c.doRequest(ctx, "GET", "/heartbeat", nil)
	if err != nil {
		return err
	}

	var result map[string]interface{}
	return c.parseResponse(resp, &result)
}

// ListCollections 列出所有集合
func (c *ChromaDBClient) ListCollections(ctx context.Context) ([]ChromaCollection, error) {
	resp, err := c.doRequest(ctx, "GET", "/collections", nil)
	if err != nil {
		return nil, err
	}

	var collections []ChromaCollection
	if err := c.parseResponse(resp, &collections); err != nil {
		return nil, err
	}

	c.logger.Debug("listed ChromaDB collections",
		logger.Field{Key: "count", Value: len(collections)},
	)

	return collections, nil
}

// CreateCollection 创建集合
func (c *ChromaDBClient) CreateCollection(ctx context.Context, name string, metadata map[string]interface{}) (*ChromaCollection, error) {
	req := ChromaCollection{
		Name:     name,
		Metadata: metadata,
	}

	resp, err := c.doRequest(ctx, "POST", "/collections", req)
	if err != nil {
		return nil, err
	}

	var collection ChromaCollection
	if err := c.parseResponse(resp, &collection); err != nil {
		return nil, err
	}

	c.logger.Info("created ChromaDB collection",
		logger.Field{Key: "name", Value: name},
		logger.Field{Key: "id", Value: collection.ID},
	)

	return &collection, nil
}

// GetCollection 获取集合信息
func (c *ChromaDBClient) GetCollection(ctx context.Context, name string) (*ChromaCollection, error) {
	path := fmt.Sprintf("/collections/%s", name)

	resp, err := c.doRequest(ctx, "GET", path, nil)
	if err != nil {
		return nil, err
	}

	var collection ChromaCollection
	if err := c.parseResponse(resp, &collection); err != nil {
		return nil, err
	}

	return &collection, nil
}

// DeleteCollection 删除集合
func (c *ChromaDBClient) DeleteCollection(ctx context.Context, name string) error {
	path := fmt.Sprintf("/collections/%s", name)

	resp, err := c.doRequest(ctx, "DELETE", path, nil)
	if err != nil {
		return err
	}

	if err := c.parseResponse(resp, nil); err != nil {
		return err
	}

	c.logger.Info("deleted ChromaDB collection",
		logger.Field{Key: "name", Value: name},
	)

	return nil
}

// Add 添加文档到集合
func (c *ChromaDBClient) Add(ctx context.Context, collectionName string, req *ChromaAddRequest) error {
	path := fmt.Sprintf("/collections/%s/add", collectionName)

	resp, err := c.doRequest(ctx, "POST", path, req)
	if err != nil {
		return err
	}

	if err := c.parseResponse(resp, nil); err != nil {
		return err
	}

	c.logger.Debug("added documents to ChromaDB collection",
		logger.Field{Key: "collection", Value: collectionName},
		logger.Field{Key: "count", Value: len(req.IDs)},
	)

	return nil
}

// Query 查询集合
func (c *ChromaDBClient) Query(ctx context.Context, collectionName string, req *ChromaQueryRequest) (*ChromaQueryResponse, error) {
	path := fmt.Sprintf("/collections/%s/query", collectionName)

	resp, err := c.doRequest(ctx, "POST", path, req)
	if err != nil {
		return nil, err
	}

	var result ChromaQueryResponse
	if err := c.parseResponse(resp, &result); err != nil {
		return nil, err
	}

	c.logger.Debug("queried ChromaDB collection",
		logger.Field{Key: "collection", Value: collectionName},
		logger.Field{Key: "results", Value: len(result.IDs)},
	)

	return &result, nil
}

// Delete 删除文档
func (c *ChromaDBClient) Delete(ctx context.Context, collectionName string, req *ChromaDeleteRequest) error {
	path := fmt.Sprintf("/collections/%s/delete", collectionName)

	resp, err := c.doRequest(ctx, "POST", path, req)
	if err != nil {
		return err
	}

	if err := c.parseResponse(resp, nil); err != nil {
		return err
	}

	c.logger.Debug("deleted documents from ChromaDB collection",
		logger.Field{Key: "collection", Value: collectionName},
	)

	return nil
}

// Update 更新文档
func (c *ChromaDBClient) Update(ctx context.Context, collectionName string, req *ChromaUpdateRequest) error {
	path := fmt.Sprintf("/collections/%s/update", collectionName)

	resp, err := c.doRequest(ctx, "POST", path, req)
	if err != nil {
		return err
	}

	if err := c.parseResponse(resp, nil); err != nil {
		return err
	}

	c.logger.Debug("updated documents in ChromaDB collection",
		logger.Field{Key: "collection", Value: collectionName},
		logger.Field{Key: "count", Value: len(req.IDs)},
	)

	return nil
}

// Upsert 插入或更新文档
func (c *ChromaDBClient) Upsert(ctx context.Context, collectionName string, req *ChromaUpsertRequest) error {
	path := fmt.Sprintf("/collections/%s/upsert", collectionName)

	resp, err := c.doRequest(ctx, "POST", path, req)
	if err != nil {
		return err
	}

	if err := c.parseResponse(resp, nil); err != nil {
		return err
	}

	c.logger.Debug("upserted documents in ChromaDB collection",
		logger.Field{Key: "collection", Value: collectionName},
		logger.Field{Key: "count", Value: len(req.IDs)},
	)

	return nil
}

// GetCollectionCount 获取集合中的文档数量
func (c *ChromaDBClient) GetCollectionCount(ctx context.Context, collectionName string) (int, error) {
	path := fmt.Sprintf("/collections/%s/count", collectionName)

	resp, err := c.doRequest(ctx, "GET", path, nil)
	if err != nil {
		return 0, err
	}

	var count int
	if err := c.parseResponse(resp, &count); err != nil {
		return 0, err
	}

	return count, nil
}

// Reset 重置ChromaDB（删除所有集合）
func (c *ChromaDBClient) Reset(ctx context.Context) error {
	resp, err := c.doRequest(ctx, "POST", "/reset", nil)
	if err != nil {
		return err
	}

	if err := c.parseResponse(resp, nil); err != nil {
		return err
	}

	c.logger.Warn("reset ChromaDB - all collections deleted")
	return nil
}

// Close 关闭客户端
func (c *ChromaDBClient) Close() error {
	c.logger.Debug("closing ChromaDB client")
	// HTTP客户端不需要显式关闭
	return nil
}

// sanitizeCollectionName 清理集合名称，确保符合ChromaDB规范
func sanitizeCollectionName(name string) string {
	// ChromaDB集合名称规则：
	// - 长度在3-63字符之间
	// - 只能包含字母、数字、下划线、连字符
	// - 不能以连字符开头或结尾
	// - 必须以字母或数字开头和结尾

	// 替换非法字符为下划线
	cleaned := ""
	for _, char := range strings.ToLower(name) {
		if (char >= 'a' && char <= 'z') ||
			(char >= '0' && char <= '9') ||
			char == '_' || char == '-' {
			cleaned += string(char)
		} else {
			cleaned += "_"
		}
	}

	// 确保开头是字母或数字
	if len(cleaned) > 0 && (cleaned[0] == '-' || cleaned[0] == '_') {
		cleaned = "c" + cleaned
	}
	
	// 处理结尾的连字符和下划线
	originalLength := len(cleaned)
	cleaned = strings.TrimRight(cleaned, "-_")
	
	// 如果移除了结尾的字符，加上'c'
	if len(cleaned) < originalLength {
		cleaned = cleaned + "c"
	}

	// 确保长度在合理范围内
	if len(cleaned) < 3 {
		cleaned = cleaned + "collection"
	}
	if len(cleaned) > 63 {
		// 截断到60个字符，然后加上"col"后缀
		cleaned = cleaned[:60] + "col"
	}

	return cleaned
}
