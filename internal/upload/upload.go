package upload

import (
	"bytes"
	"crypto/md5"
	"cursor_history/internal/app"
	"cursor_history/internal/storage"
	"cursor_history/internal/types"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/go-git/go-git/v5"
)

// ValidateApiKey 验证 API Token
func ValidateApiKey(apiKey string, serverURL string) error {
	// 从服务器 URL 中提取基础 URL
	baseURL := serverURL[:len(serverURL)-len("/api/prompt/upload")]
	validURL := baseURL + "/api/api-key/valid?key=" + url.QueryEscape(apiKey)

	// 创建请求
	req, err := http.NewRequest("GET", validURL, nil)
	if err != nil {
		return fmt.Errorf("创建请求失败: %v", err)
	}

	// 设置请求头
	req.Header.Set("X-API-Key", apiKey)

	// 发送请求
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("验证请求失败: %v", err)
	}
	defer resp.Body.Close()

	// 检查响应状态
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("token 无效")
	}

	return nil
}

// WatchDirectory 监控目录变化
func WatchDirectory(searchPath string, configManager *storage.ConfigManager, logger types.Logger) error {
	// 使用缓冲通道
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return fmt.Errorf("创建文件监控失败: %v", err)
	}

	// 保存 watcher 到全局变量
	watcherMutex.Lock()
	currentWatcher = watcher
	watcherMutex.Unlock()

	// 确保在函数返回时关闭 watcher
	defer func() {
		watcherMutex.Lock()
		if currentWatcher == watcher {
			currentWatcher = nil
		}
		watcherMutex.Unlock()

		err := watcher.Close()
		if err != nil {
			logger.Log(types.LogLevelError, "关闭 watcher 失败: %v", err)
		}
	}()

	// 使用对象池来减少内存分配
	fileInfoPool := sync.Pool{
		New: func() interface{} {
			return &FileInfo{}
		},
	}

	// 创建一个处理中的文件映射
	processingFiles := make(map[string]bool)
	var mu sync.Mutex

	// 创建一个 done 通道用于清理
	done := make(chan struct{})
	defer close(done)

	// 启动一个 goroutine 来处理文件事件
	go func() {
		for {
			select {
			case <-done:
				return
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}

				// 如果有新目录创建，添加到监控
				if event.Op&fsnotify.Create == fsnotify.Create {
					if fi, err := os.Stat(event.Name); err == nil && fi.IsDir() {
						if err := addWatchDir(watcher, event.Name, logger); err != nil {
							logger.Log(types.LogLevelError, "添加新目录监控失败: %v", err)
						}
					}
				}

				if filepath.Base(event.Name) == "state.vscdb" && (event.Op&fsnotify.Write == fsnotify.Write) {
					// 检查文件是否正在处理中
					mu.Lock()
					if processingFiles[event.Name] {
						mu.Unlock()
						continue
					}
					processingFiles[event.Name] = true
					mu.Unlock()

					// 处理文件
					fileInfo := fileInfoPool.Get().(*FileInfo)
					fileInfo.Path = event.Name
					fileInfo.ModTime = time.Now().Unix()
					processFile(*fileInfo, configManager, logger)

					// 处理完成后移除标记
					mu.Lock()
					delete(processingFiles, event.Name)
					mu.Unlock()

					// 重置并释放文件信息
					fileInfo.Path = ""
					fileInfo.ModTime = 0
					fileInfoPool.Put(fileInfo)
				}
			}
		}
	}()

	// 递归添加所有子目录到监控
	if err := addWatchDir(watcher, searchPath, logger); err != nil {
		return fmt.Errorf("添加目录监控失败: %v", err)
	}

	logger.Log(types.LogLevelInfo, "开始监控目录: %s", searchPath)

	// 主循环监听停止信号
	for {
		select {
		case err, ok := <-watcher.Errors:
			if !ok {
				return nil
			}
			logger.Log(types.LogLevelError, "监控错误: %v", err)
			return fmt.Errorf("监控错误: %v", err)

		case <-configManager.GetContext().Done():
			logger.Log(types.LogLevelInfo, "监控已停止")
			return nil
		}
	}
}

// addWatchDir 递归添加目录到监控
func addWatchDir(watcher *fsnotify.Watcher, dir string, logger types.Logger) error {
	err := watcher.Add(dir)
	if err != nil {
		return fmt.Errorf("添加目录监控失败 %s: %v", dir, err)
	}
	// logger.Log(types.LogLevelInfo, "添加目录监控: %s", dir)

	// 递归遍历子目录
	entries, err := os.ReadDir(dir)
	if err != nil {
		return fmt.Errorf("读取目录失败 %s: %v", dir, err)
	}

	for _, entry := range entries {
		if entry.IsDir() {
			subDir := filepath.Join(dir, entry.Name())
			if err := addWatchDir(watcher, subDir, logger); err != nil {
				logger.Log(types.LogLevelError, "添加子目录监控失败: %v", err)
				// 继续处理其他目录
				continue
			}
		}
	}

	return nil
}

// FileInfo 文件信息结构体
type FileInfo struct {
	Path    string
	ModTime int64
}

// 添加新结构体用于解析 workspace.json
type WorkspaceConfig struct {
	Folder string `json:"folder"`
}

// 修改 processWorkspaceJson 函数来保存 workspace 路径
func processWorkspaceJson(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("读取 workspace.json 失败: %v", err)
	}

	// 解析 JSON
	var config WorkspaceConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return "", fmt.Errorf("解析 workspace.json 失败: %v", err)
	}

	// 解码 URL 编码的路径
	decodedFolder, err := url.QueryUnescape(strings.TrimPrefix(config.Folder, "file:///"))
	if err != nil {
		return "", fmt.Errorf("解码文件夹路径失败: %v", err)
	}

	// fmt.Printf("工作区文件夹: %s\n", decodedFolder)
	return decodedFolder, nil
}

// processFile 处理文件
func processFile(file FileInfo, configManager *storage.ConfigManager, logger types.Logger) {
	// 记录处理开始
	// logger.Log(types.LogLevelInfo, "开始处理文件: %s", file.Path)

	// 这里添加文件处理的具体逻辑
	// 例如：读取文件内容，解析数据，上传到服务器等
	// 在描文件之前，先处理 workspace.json
	workspaceJsonPath := filepath.Join(filepath.Dir(file.Path), "workspace.json")
	workspace, err := processWorkspaceJson(workspaceJsonPath)
	if err != nil {
		logger.Log(types.LogLevelError, "处理 workspace.json 出错: %v", err)
		return
	}

	// 打开数据库
	db, err := sql.Open("sqlite3", file.Path)
	if err != nil {
		logger.Log(types.LogLevelError, "无法打开数据库: %v", err)
		return
	}
	defer db.Close()

	// 查询所有表名
	rows, err := db.Query("SELECT name FROM sqlite_master WHERE type='table';")
	if err != nil {
		logger.Log(types.LogLevelError, "查询表名出错: %v", err)
		return
	}
	defer rows.Close()

	var tableName string
	for rows.Next() {
		err = rows.Scan(&tableName)
		if err != nil {
			logger.Log(types.LogLevelError, "扫描表名出错: %v", err)
			continue
		}

		// 查询表数据
		tableRows, err := db.Query(fmt.Sprintf("SELECT * FROM %s;", tableName))
		if err != nil {
			logger.Log(types.LogLevelError, "查询表数据出错: %v", err)
			continue
		}
		defer tableRows.Close()

		// 获取列名
		columns, err := tableRows.Columns()
		if err != nil {
			logger.Log(types.LogLevelError, "获取列名出错: %v", err)
			continue
		}

		// 创建一个切片来存储每行的值
		values := make([]interface{}, len(columns))
		valuePtrs := make([]interface{}, len(columns))
		for i := range values {
			valuePtrs[i] = &values[i]
		}

		// 输出表数据
		for tableRows.Next() {
			err = tableRows.Scan(valuePtrs...)
			if err != nil {
				logger.Log(types.LogLevelError, "扫描行数据出错: %v", err)
				continue
			}

			// 检查值中是否包含 "cursor"
			containsCursor := false
			for i, value := range values {
				if strValue, ok := value.(string); ok && contains(strValue, "aiService.prompts") && columns[i] == "key" {
					containsCursor = true
					break
				}
			}

			// 修改 uploadPrompt 的调用，传入 workspace 参数
			if containsCursor {
				for i, col := range columns {
					if strValue, ok := values[i].(string); ok {
						uploadPrompt(strValue, col, file.ModTime, configManager, workspace, logger)
					}
				}
			}
		}
	}

	// 记录处理结果
	// logger.Log(types.LogLevelSuccess, "文件处理完成: %s", file.Path)
}

// 检查字符串是否包含子字符串
func contains(s, substr string) bool {
	return strings.Contains(s, substr)
}

// 修改 uploadPrompt 函数签名，添加 workspace 参数
func uploadPrompt(value string, col string, timestamp int64, configManager *storage.ConfigManager, workspace string, logger types.Logger) {
	if col == "key" {
		return
	}

	uploadList, err := convertValueToUploadPrompt(value)
	if err != nil {
		logger.Log(types.LogLevelError, "转换值失败: %v", err)
		return
	}

	// 获取 Git 信息
	gitInfo := getGitInfo(workspace, logger)

	for _, upload := range uploadList {
		uploadSinglePrompt(upload, timestamp, configManager, workspace, logger, gitInfo)
	}
}

type UploadPrompt struct {
	Text        string `json:"text"`
	CommandType int    `json:"commandType"`
}

func convertValueToUploadPrompt(value string) ([]UploadPrompt, error) {
	var uploadList []UploadPrompt
	err := json.Unmarshal([]byte(value), &uploadList)
	return uploadList, err
}

// 修改 uploadSinglePrompt 函数签名，添加 workspace 参数
func uploadSinglePrompt(prompt UploadPrompt, timestamp int64, configManager *storage.ConfigManager, workspace string, logger types.Logger, gitInfo GitInfo) {
	// 计算MD5值
	hash := md5.Sum([]byte(prompt.Text))
	md5Value := hex.EncodeToString(hash[:])

	// 检查MD5是否已上传
	exists, err := configManager.IsMD5Uploaded(md5Value)
	if err != nil {
		logger.Log(types.LogLevelError, "检查MD5失败: %v", err)
		return
	}
	if exists {
		// logger.Log(types.LogLevelInfo, "MD5已存在，跳过上传: %s", md5Value)
		return
	}

	// 准备请求数据
	data := map[string]interface{}{
		"value":       prompt.Text,
		"commandType": strconv.Itoa(prompt.CommandType),
		"md5":         md5Value,
		"timestamp":   timestamp,
		"workspace":   workspace,
		"uploadTime":  time.Now().UnixMilli(),
		"git": map[string]interface{}{
			"isGitRepo":  gitInfo.IsGitRepo,
			"remoteUrl":  gitInfo.RemoteURL,
			"commitHash": gitInfo.CommitHash,
			"branchName": gitInfo.BranchName,
		},
	}
	jsonData, err := json.Marshal(data)
	if err != nil {
		logger.Log(types.LogLevelError, "JSON 编码失败: %v", err)
		return
	}

	// 创建请求，使用 Config.ServerURL 而不是 app.ServerURL
	req, err := http.NewRequest("POST", app.Config.ServerURL, bytes.NewBuffer(jsonData))
	if err != nil {
		logger.Log(types.LogLevelError, "创建请求失败: %v", err)
		return
	}

	// 设置请求头
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-API-Key", app.Config.ApiKey) // 使用 appConfig.Token

	// 发送请求
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		logger.Log(types.LogLevelError, "上传失败: %v", err)
		return
	}
	defer resp.Body.Close()

	// 检查响应状态
	if resp.StatusCode != http.StatusOK {
		logger.Log(types.LogLevelError, "服务器返回错误状态: %d", resp.StatusCode)
		return
	}

	// 上传成功后保存MD5
	if err := configManager.SaveMD5(md5Value); err != nil {
		logger.Log(types.LogLevelError, "保存MD5失败: %v", err)
		return
	}

	logger.Log(types.LogLevelSuccess, "成功上传: %v %v", prompt.Text, prompt.CommandType)
}

// 添加一个全局的 watcher 变量和互斥锁
var (
	currentWatcher *fsnotify.Watcher
	watcherMutex   sync.Mutex
)

// 添加关闭 watcher 的函数
func CloseWatcher() {
	watcherMutex.Lock()
	defer watcherMutex.Unlock()

	if currentWatcher != nil {
		currentWatcher.Close()
		currentWatcher = nil
	}
}

// GitInfo 结构体定义
type GitInfo struct {
	IsGitRepo  bool   `json:"isGitRepo"`
	RemoteURL  string `json:"remoteUrl"`
	CommitHash string `json:"commitHash"`
	BranchName string `json:"branchName"`
}

// 获取 Git 信息的函数
func getGitInfo(workspace string, logger types.Logger) GitInfo {
	info := GitInfo{
		IsGitRepo: false,
	}

	// 递归向上查找 .git 目录
	currentPath := workspace
	for {
		// 尝试打开 Git 仓库
		repo, err := git.PlainOpen(currentPath)
		if err == nil {
			info.IsGitRepo = true

			// 获取远程仓库信息
			remotes, err := repo.Remotes()
			if err == nil && len(remotes) > 0 {
				if urls := remotes[0].Config().URLs; len(urls) > 0 {
					info.RemoteURL = urls[0]
				}
			}

			// 获取当前 HEAD
			head, err := repo.Head()
			if err == nil {
				info.CommitHash = head.Hash().String()

				// 获取分支名
				if head.Name().IsBranch() {
					info.BranchName = head.Name().Short()
				}
			}

			// logger.Log(types.LogLevelInfo, "找到 Git 仓库: %s", currentPath)
			break
		}

		// 获取父目录
		parentPath := filepath.Dir(currentPath)
		if parentPath == currentPath {
			// 已经到达根目录，仍未找到 .git
			// logger.Log(types.LogLevelInfo, "未找到 Git 仓库: %s", workspace)
			break
		}
		currentPath = parentPath
	}

	return info
}
