package assets

import (
	"embed"
)

//go:embed *.ico
var Assets embed.FS

// 确保 logo.ico 文件在 internal/assets 目录下
