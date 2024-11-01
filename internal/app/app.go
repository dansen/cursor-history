package app

import (
	"cursor_history/internal/types"
	"fmt"
	"log"
	"os"
)

// App 配置
type App struct {
	ApiKey    string
	ServerURL string
	logger    types.Logger
}

// 服务器 URL 常量
const (
	DevServerURL  = "http://localhost:7600/api/prompt/upload"
	ProdServerURL = "https://cursorai.v8cloud.cn/api/prompt/upload"
)

// 全局配置实例
var Config = &App{}

// InitApp 初始化应用配置
func InitApp(defaultEnv string) {
	// 先尝试从环境变量获取环境设置
	env := os.Getenv("CURSOR_ENV")

	// 如果环境变量为空，则使用默认值
	if env == "" {
		env = defaultEnv
	}

	// 根据环境设置服务器 URL
	if env == "prod" {
		Config.ServerURL = ProdServerURL
	} else {
		Config.ServerURL = DevServerURL
	}

	// 记录当前环境
	log.Printf("当前环境: %s", env)
}

// SetLogger 设置日志记录器
func (a *App) SetLogger(logger types.Logger) {
	a.logger = logger
}

// Stop 停止应用
func (a *App) Stop() error {
	if a.logger != nil {
		if err := a.logger.Close(); err != nil {
			return fmt.Errorf("failed to close logger: %w", err)
		}
	}
	return nil
}

// GetEnv 获取当前环境
func GetEnv() string {
	if Config.ServerURL == ProdServerURL {
		return "生产"
	}
	return "开发"
}
