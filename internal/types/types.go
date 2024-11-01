package types

// Logger 定义日志接口
type Logger interface {
	Log(level string, format string, args ...interface{})
	// Close 用于清理资源并停止日志记录
	Close() error
}

// 日志级别常量
const (
	LogLevelError   = "ERROR"
	LogLevelWarning = "WARN"
	LogLevelSuccess = "SUCCESS"
	LogLevelInfo    = "INFO"
	LogLevelDefault = "DEFAULT"
)
