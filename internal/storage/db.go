package storage

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"time"

	_ "github.com/mattn/go-sqlite3" // 导入 sqlite3 驱动
)

// ConfigManager 配置管理器
type ConfigManager struct {
	db     *sql.DB
	ctx    context.Context
	cancel context.CancelFunc
}

// NewConfigManager 创建新的配置管理器
func NewConfigManager(dbPath string) (*ConfigManager, error) {
	// 确保数据库目录存在
	dbDir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dbDir, 0755); err != nil {
		return nil, fmt.Errorf("创建数据库目录失败: %v", err)
	}

	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("打开数据库失败: %v", err)
	}

	// 测试数据库连接
	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("连接数据库失败: %v", err)
	}

	// 创建配置表
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS config (
			key TEXT PRIMARY KEY,
			value TEXT
		)
	`)
	if err != nil {
		db.Close()
		return nil, fmt.Errorf("创建配置表失败: %v", err)
	}

	// 创建 MD5 表
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS uploaded_md5 (
			md5 TEXT PRIMARY KEY,
			upload_time INTEGER
		)
	`)
	if err != nil {
		db.Close()
		return nil, fmt.Errorf("创建 MD5 表失败: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())

	return &ConfigManager{
		db:     db,
		ctx:    ctx,
		cancel: cancel,
	}, nil
}

// SaveApiKey 保存 API Key
func (c *ConfigManager) SaveApiKey(apiKey string) error {
	_, err := c.db.Exec(`
		INSERT OR REPLACE INTO config (key, value)
		VALUES ('api_key', ?)
	`, apiKey)
	if err != nil {
		return fmt.Errorf("保存 API Key 失败: %v", err)
	}
	return nil
}

// LoadApiKey 加载 API Key
func (c *ConfigManager) LoadApiKey() (string, error) {
	var apiKey string
	err := c.db.QueryRow(`
		SELECT value FROM config
		WHERE key = 'api_key'
	`).Scan(&apiKey)

	if err == sql.ErrNoRows {
		return "", nil // 返回空字符串，表示没有保存的 API Key
	}
	if err != nil {
		return "", fmt.Errorf("加载 API Key 失败: %v", err)
	}
	return apiKey, nil
}

// GetContext 获取 context
func (c *ConfigManager) GetContext() context.Context {
	return c.ctx
}

// Close 关闭数据库连接
func (c *ConfigManager) Close() error {
	c.cancel() // 取消 context
	if c.db != nil {
		return c.db.Close()
	}
	return nil
}

// 移动其他数据库相关方法...

func (cm *ConfigManager) IsMD5Uploaded(md5 string) (bool, error) {
	var exists bool
	err := cm.db.QueryRow(`
		SELECT EXISTS(
			SELECT 1 FROM uploaded_md5 WHERE md5 = ?
		)
	`, md5).Scan(&exists)

	if err != nil {
		return false, fmt.Errorf("检查 MD5 失败: %v", err)
	}
	return exists, nil
}

func (cm *ConfigManager) SaveMD5(md5 string) error {
	_, err := cm.db.Exec(`
		INSERT OR REPLACE INTO uploaded_md5 (md5, upload_time)
		VALUES (?, ?)
	`, md5, time.Now().Unix())

	if err != nil {
		return fmt.Errorf("保存 MD5 失败: %v", err)
	}
	return nil
}
