package storage

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"github.com/ynl/greensoulai/internal/memory"
	"github.com/ynl/greensoulai/pkg/logger"
)

// LTMSQLiteStorage 长期记忆SQLite存储实现
type LTMSQLiteStorage struct {
	dbPath string
	db     *sql.DB
	logger logger.Logger
	mu     sync.RWMutex // 读写锁保护并发访问
}

// NewLTMSQLiteStorage 创建SQLite存储实例
func NewLTMSQLiteStorage(dbPath string, log logger.Logger) *LTMSQLiteStorage {
	if dbPath == "" {
		// 使用默认路径
		dbPath = filepath.Join(".", "data", "long_term_memory.db")
	}

	storage := &LTMSQLiteStorage{
		dbPath: dbPath,
		logger: log,
	}

	// 初始化数据库
	if err := storage.initialize(); err != nil {
		log.Error("failed to initialize SQLite storage",
			logger.Field{Key: "error", Value: err})
	}

	return storage
}

// initialize 初始化数据库
func (s *LTMSQLiteStorage) initialize() error {
	// 创建数据库目录
	dir := filepath.Dir(s.dbPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create database directory: %w", err)
	}

	// 打开数据库连接
	db, err := sql.Open("sqlite3", s.dbPath)
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}

	// 测试连接
	if err := db.Ping(); err != nil {
		if err := db.Close(); err != nil {
			s.logger.Error("Failed to close database connection",
				logger.Field{Key: "error", Value: err})
		}
		return fmt.Errorf("failed to ping database: %w", err)
	}

	s.db = db
	s.logger.Info("SQLite storage initialized successfully",
		logger.Field{Key: "db_path", Value: s.dbPath},
	)

	// 创建表结构
	return s.createTables()
}

// createTables 创建数据库表
func (s *LTMSQLiteStorage) createTables() error {
	// 创建主表
	createTableSQL := `
	CREATE TABLE IF NOT EXISTS long_term_memories (
		id TEXT PRIMARY KEY,
		value TEXT NOT NULL,
		metadata TEXT,
		agent TEXT,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		score REAL DEFAULT 0.0,
		access_count INTEGER DEFAULT 0,
		last_access DATETIME
	);
	
	CREATE INDEX IF NOT EXISTS idx_agent ON long_term_memories(agent);
	CREATE INDEX IF NOT EXISTS idx_created_at ON long_term_memories(created_at);
	CREATE INDEX IF NOT EXISTS idx_score ON long_term_memories(score);
	CREATE INDEX IF NOT EXISTS idx_last_access ON long_term_memories(last_access);
	`

	_, err := s.db.Exec(createTableSQL)
	if err != nil {
		return fmt.Errorf("failed to create main tables: %w", err)
	}

	// 尝试创建FTS5表（可选，如果不支持则跳过）
	s.tryCreateFTSTables()

	return nil
}

// tryCreateFTSTables 尝试创建FTS5全文搜索表（如果支持）
func (s *LTMSQLiteStorage) tryCreateFTSTables() {
	ftsSQL := `
	CREATE VIRTUAL TABLE IF NOT EXISTS memories_fts USING fts5(
		id, value, metadata, agent, content='long_term_memories', content_rowid='rowid'
	);
	
	-- 创建触发器以保持FTS表同步
	CREATE TRIGGER IF NOT EXISTS memories_ai AFTER INSERT ON long_term_memories BEGIN
		INSERT INTO memories_fts(id, value, metadata, agent) VALUES (new.id, new.value, new.metadata, new.agent);
	END;
	
	CREATE TRIGGER IF NOT EXISTS memories_ad AFTER DELETE ON long_term_memories BEGIN
		DELETE FROM memories_fts WHERE id = old.id;
	END;
	
	CREATE TRIGGER IF NOT EXISTS memories_au AFTER UPDATE ON long_term_memories BEGIN
		DELETE FROM memories_fts WHERE id = old.id;
		INSERT INTO memories_fts(id, value, metadata, agent) VALUES (new.id, new.value, new.metadata, new.agent);
	END;
	`

	if _, err := s.db.Exec(ftsSQL); err != nil {
		s.logger.Info("FTS5 not available, using basic search instead",
			logger.Field{Key: "error", Value: err.Error()},
		)
	} else {
		s.logger.Info("FTS5 full-text search enabled")
	}
}

// Save 保存记忆项到SQLite
func (s *LTMSQLiteStorage) Save(ctx context.Context, item memory.MemoryItem) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// 序列化元数据
	metadataJSON, err := json.Marshal(item.Metadata)
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}

	// 将值转换为字符串
	valueStr := fmt.Sprintf("%v", item.Value)

	insertSQL := `
	INSERT OR REPLACE INTO long_term_memories 
	(id, value, metadata, agent, created_at, score, access_count, last_access)
	VALUES (?, ?, ?, ?, ?, ?, 0, CURRENT_TIMESTAMP)
	`

	_, err = s.db.ExecContext(ctx, insertSQL,
		item.ID,
		valueStr,
		string(metadataJSON),
		item.Agent,
		item.CreatedAt.Format(time.RFC3339),
		item.Score,
	)

	if err != nil {
		return fmt.Errorf("failed to insert memory item: %w", err)
	}

	s.logger.Debug("memory item saved to SQLite",
		logger.Field{Key: "id", Value: item.ID},
		logger.Field{Key: "agent", Value: item.Agent},
	)

	return nil
}

// Search 搜索记忆项
func (s *LTMSQLiteStorage) Search(ctx context.Context, query string, limit int, scoreThreshold float64) ([]memory.MemoryItem, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// 确保总是返回非nil的slice
	results := make([]memory.MemoryItem, 0)

	// 使用FTS进行全文搜索
	searchSQL := `
	SELECT m.id, m.value, m.metadata, m.agent, m.created_at, m.score, m.access_count, m.last_access
	FROM memories_fts fts
	JOIN long_term_memories m ON fts.id = m.id
	WHERE memories_fts MATCH ? AND m.score >= ?
	ORDER BY bm25(memories_fts) ASC, m.score DESC, m.last_access DESC
	LIMIT ?
	`

	rows, err := s.db.QueryContext(ctx, searchSQL, query, scoreThreshold, limit)
	if err != nil {
		// 如果FTS搜索失败，尝试简单的LIKE搜索
		return s.simpleLikeSearch(ctx, query, limit, scoreThreshold)
	}
	defer rows.Close()

	for rows.Next() {
		var item memory.MemoryItem
		var metadataJSON, createdAtStr, lastAccessStr string
		var accessCount int

		err := rows.Scan(
			&item.ID,
			&item.Value,
			&metadataJSON,
			&item.Agent,
			&createdAtStr,
			&item.Score,
			&accessCount,
			&lastAccessStr,
		)
		if err != nil {
			s.logger.Error("failed to scan row", logger.Field{Key: "error", Value: err})
			continue
		}

		// 反序列化元数据
		if metadataJSON != "" {
			if err := json.Unmarshal([]byte(metadataJSON), &item.Metadata); err != nil {
				s.logger.Error("failed to unmarshal metadata", logger.Field{Key: "error", Value: err})
				item.Metadata = make(map[string]interface{})
			}
		} else {
			item.Metadata = make(map[string]interface{})
		}

		// 解析时间
		if createdAt, err := time.Parse(time.RFC3339, createdAtStr); err == nil {
			item.CreatedAt = createdAt
		}

		// 更新访问统计（在并发环境下暂时禁用以避免读锁升级为写锁的死锁）
		// TODO: 重构为异步更新或使用更细粒度的锁
		// s.updateAccessStats(ctx, item.ID)

		results = append(results, item)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration error: %w", err)
	}

	s.logger.Debug("SQLite search completed",
		logger.Field{Key: "query", Value: query},
		logger.Field{Key: "results_count", Value: len(results)},
	)

	return results, nil
}

// simpleLikeSearch 简单的LIKE搜索作为后备方案
func (s *LTMSQLiteStorage) simpleLikeSearch(ctx context.Context, query string, limit int, scoreThreshold float64) ([]memory.MemoryItem, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	results := make([]memory.MemoryItem, 0)

	searchSQL := `
	SELECT id, value, metadata, agent, created_at, score, access_count, last_access
	FROM long_term_memories
	WHERE (value LIKE ? OR agent LIKE ? OR metadata LIKE ?) AND score >= ?
	ORDER BY score DESC, created_at DESC
	LIMIT ?
	`

	likeQuery := "%" + query + "%"
	rows, err := s.db.QueryContext(ctx, searchSQL, likeQuery, likeQuery, likeQuery, scoreThreshold, limit)
	if err != nil {
		return nil, fmt.Errorf("simple search failed: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var item memory.MemoryItem
		var metadataJSON, createdAtStr, lastAccessStr string
		var accessCount int

		err := rows.Scan(
			&item.ID,
			&item.Value,
			&metadataJSON,
			&item.Agent,
			&createdAtStr,
			&item.Score,
			&accessCount,
			&lastAccessStr,
		)
		if err != nil {
			continue
		}

		// 反序列化元数据
		if metadataJSON != "" {
			json.Unmarshal([]byte(metadataJSON), &item.Metadata)
		} else {
			item.Metadata = make(map[string]interface{})
		}

		// 解析时间
		if createdAt, err := time.Parse(time.RFC3339, createdAtStr); err == nil {
			item.CreatedAt = createdAt
		}

		// 更新访问统计（在并发环境下暂时禁用以避免读锁升级为写锁的死锁）
		// TODO: 重构为异步更新或使用更细粒度的锁
		// s.updateAccessStats(ctx, item.ID)

		results = append(results, item)
	}

	return results, nil
}

// Delete 删除记忆项
func (s *LTMSQLiteStorage) Delete(ctx context.Context, id string) error {
	deleteSQL := `DELETE FROM long_term_memories WHERE id = ?`

	result, err := s.db.ExecContext(ctx, deleteSQL, id)
	if err != nil {
		return fmt.Errorf("failed to delete memory item: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("memory item not found: %s", id)
	}

	s.logger.Debug("memory item deleted from SQLite",
		logger.Field{Key: "id", Value: id},
	)

	return nil
}

// Clear 清除所有记忆项
func (s *LTMSQLiteStorage) Clear(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// 获取删除前的数量
	var count int
	countSQL := `SELECT COUNT(*) FROM long_term_memories`
	s.db.QueryRowContext(ctx, countSQL).Scan(&count)

	// 删除所有记录
	clearSQL := `DELETE FROM long_term_memories`
	_, err := s.db.ExecContext(ctx, clearSQL)
	if err != nil {
		return fmt.Errorf("failed to clear memories: %w", err)
	}

	s.logger.Info("SQLite storage cleared",
		logger.Field{Key: "deleted_count", Value: count},
	)

	return nil
}

// Close 关闭存储
func (s *LTMSQLiteStorage) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.db != nil {
		s.logger.Info("closing SQLite storage")
		return s.db.Close()
	}
	return nil
}

// GetStats 获取存储统计信息
func (s *LTMSQLiteStorage) GetStats(ctx context.Context) (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	// 总记录数
	var totalCount int
	s.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM long_term_memories`).Scan(&totalCount)
	stats["total_memories"] = totalCount

	// 按agent统计
	agentStats := make(map[string]int)
	rows, err := s.db.QueryContext(ctx, `SELECT agent, COUNT(*) FROM long_term_memories GROUP BY agent`)
	if err == nil {
		for rows.Next() {
			var agent string
			var count int
			if rows.Scan(&agent, &count) == nil {
				agentStats[agent] = count
			}
		}
		rows.Close()
	}
	stats["agent_stats"] = agentStats

	// 最近访问的记忆数量（24小时内）
	var recentCount int
	s.db.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM long_term_memories WHERE last_access > datetime('now', '-24 hours')`).Scan(&recentCount)
	stats["recent_accessed"] = recentCount

	// 平均分数
	var avgScore float64
	s.db.QueryRowContext(ctx, `SELECT AVG(score) FROM long_term_memories WHERE score > 0`).Scan(&avgScore)
	stats["average_score"] = avgScore

	// 数据库信息
	stats["db_path"] = s.dbPath

	return stats, nil
}

// Vacuum 压缩数据库
func (s *LTMSQLiteStorage) Vacuum(ctx context.Context) error {
	_, err := s.db.ExecContext(ctx, `VACUUM`)
	if err != nil {
		return fmt.Errorf("failed to vacuum database: %w", err)
	}

	s.logger.Info("database vacuumed successfully")
	return nil
}
