package gui

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"cursor_history/internal/app"
	"cursor_history/internal/storage"
	"cursor_history/internal/upload"

	"github.com/lxn/walk"
	. "github.com/lxn/walk/declarative"
	"github.com/lxn/win"
	"golang.org/x/sys/windows/registry"
)

// 添加日志级别常量
const (
	LogLevelError   = "ERROR"
	LogLevelWarning = "WARN"
	LogLevelSuccess = "SUCCESS"
	LogLevelInfo    = "INFO"
	LogLevelDefault = "DEFAULT"
)

// GUI 结构体
type GUI struct {
	window            *walk.MainWindow
	apiKeyEntry       *walk.LineEdit
	logView           *walk.TextEdit
	status            *walk.StatusBarItem
	apiKeyStatus      *walk.StatusBarItem
	envStatus         *walk.StatusBarItem
	modeStatus        *walk.StatusBarItem
	startBtn          *walk.PushButton
	clearLogBtn       *walk.PushButton
	isMonitoring      bool
	stopChan          chan struct{} // 改用 struct{} 类型
	config            *storage.ConfigManager
	autoStartCheckBox *walk.CheckBox
	countdownTimer    *time.Timer
	countdownSec      int
	isAutoStarting    bool // 新增：标记是否正在自动启动
	countdownStopped  bool // 新增：标记倒计时是否已停止
	statusIndicator   *walk.StatusBarItem
	runningIcon       *walk.Icon
	stoppedIcon       *walk.Icon
}

// 优化内存分配
var (
	logBuffer   strings.Builder
	messagePool = sync.Pool{
		New: func() interface{} {
			return new(strings.Builder)
		},
	}
)

// NewGUI 创建新的 GUI 实例
func NewGUI(configManager *storage.ConfigManager) (*GUI, error) {
	// 生成图标文件
	if err := generateIcons(); err != nil {
		return nil, fmt.Errorf("生成图标失败: %v", err)
	}

	var gui GUI
	gui.stopChan = make(chan struct{})
	gui.config = configManager
	gui.countdownSec = 3 // 设置3秒倒计时

	// 创建字体
	normalFont := Font{
		Family:    "Segoe UI",
		PointSize: 10,
	}

	// 获取主屏幕尺寸
	hDC := win.GetDC(0)
	defer win.ReleaseDC(0, hDC)
	screenWidth := int(win.GetDeviceCaps(hDC, win.HORZRES))
	screenHeight := int(win.GetDeviceCaps(hDC, win.VERTRES))

	// 创建运行和停止状态的图标
	var err error
	gui.runningIcon, err = walk.NewIconFromFile("assets/green-dot.ico")
	if err != nil {
		return nil, fmt.Errorf("加载运行状态图标失败: %v", err)
	}
	gui.stoppedIcon, err = walk.NewIconFromFile("assets/gray-dot.ico")
	if err != nil {
		return nil, fmt.Errorf("加载停止状态图标失败: %v", err)
	}

	// 在创建主窗口之前，先加载图标
	icon, err := walk.NewIconFromFile("logo.ico")
	if err != nil {
		return nil, fmt.Errorf("加载图标失败: %v", err)
	}

	// 创建主窗口时使用图标
	if err := (MainWindow{
		AssignTo:   &gui.window,
		Title:      fmt.Sprintf("Cursor History v%s", app.Version),
		Size:       Size{Width: 800, Height: 560},
		MinSize:    Size{Width: 800, Height: 560},
		Font:       normalFont,
		Background: SolidColorBrush{Color: walk.RGB(240, 240, 240)},
		Icon:       icon,
		Layout:     VBox{Margins: Margins{Left: 20, Top: 20, Right: 20, Bottom: 20}},
		Children: []Widget{
			GroupBox{
				Title:  "设置",
				Layout: HBox{},
				Children: []Widget{
					Label{Text: "API Key:"},
					LineEdit{
						AssignTo:     &gui.apiKeyEntry,
						PasswordMode: false,
						MinSize:      Size{Width: 600},
					},
				},
			},
			Composite{
				Layout: HBox{},
				Children: []Widget{
					PushButton{
						AssignTo: &gui.startBtn,
						Text:     "开始监控",
						MinSize:  Size{Width: 100},
					},
					PushButton{
						AssignTo: &gui.clearLogBtn,
						Text:     "清空日志",
						MinSize:  Size{Width: 100},
					},
				},
			},
			TextEdit{
				AssignTo: &gui.logView,
				ReadOnly: true,
				VScroll:  true,
				MinSize:  Size{Height: 300},
			},
			Composite{
				Layout: HBox{},
				Children: []Widget{
					CheckBox{
						AssignTo: &gui.autoStartCheckBox,
						Text:     "开机自动启动",
						Checked:  true,
					},
					HSpacer{}, // 添加水平空白填充，将复选框推到左边
				},
			},
		},
		StatusBarItems: []StatusBarItem{
			{AssignTo: &gui.status, Text: "就绪", Width: 150},
			{AssignTo: &gui.statusIndicator, Icon: gui.stoppedIcon, Width: 30, Text: " "}, // 增加宽度并添加空格
			{AssignTo: &gui.apiKeyStatus, Text: "Key: 未设置", Width: 300},
			{AssignTo: &gui.envStatus, Text: fmt.Sprintf("环境: %s", app.GetEnv()), Width: 150},
			{AssignTo: &gui.modeStatus, Text: "状态: 未监控", Width: 150},
		},
	}).Create(); err != nil {
		return nil, err
	}

	// 检查当前开机启动状态
	key, err := registry.OpenKey(registry.CURRENT_USER,
		`Software\Microsoft\Windows\CurrentVersion\Run`,
		registry.READ)
	if err == nil {
		_, _, err = key.GetStringValue("CursorHistory")
		gui.autoStartCheckBox.SetChecked(err != registry.ErrNotExist)
		key.Close()
	}

	// 设置窗口位置
	x := (screenWidth - 800) / 2
	y := (screenHeight - 560) / 2
	gui.window.SetBounds(walk.Rectangle{X: x, Y: y, Width: 800, Height: 560})

	// 从配置中加载 API Key
	if savedApiKey, err := configManager.LoadApiKey(); err == nil && savedApiKey != "" {
		gui.apiKeyEntry.SetText(savedApiKey)
		app.Config.ApiKey = savedApiKey
		gui.apiKeyStatus.SetText("Key: " + savedApiKey[:10] + "...")
		gui.Log(LogLevelInfo, "已加载保存的 API Key")
	}

	// 添加复选框事件处理
	gui.autoStartCheckBox.CheckedChanged().Attach(func() {
		if err := gui.setAutoStart(gui.autoStartCheckBox.Checked()); err != nil {
			gui.Log(LogLevelError, "设置开机启动失败: %v", err)
		} else {
			if gui.autoStartCheckBox.Checked() {
				gui.Log(LogLevelSuccess, "已设置开机启动")
			} else {
				gui.Log(LogLevelInfo, "已取消开机启动")
			}
		}
	})

	// 设按钮事件
	gui.clearLogBtn.Clicked().Attach(func() {
		gui.logView.SetText("")
		gui.status.SetText("日志已清空")
	})

	// 启动倒计时
	gui.startCountdown()

	// 添加开始监控按钮事件处理
	gui.startBtn.Clicked().Attach(func() {
		if !gui.isMonitoring {
			gui.Log(LogLevelInfo, "开始按钮被点击")
			gui.countdownStopped = true // 停止倒计时
			gui.countdownSec = 0
			gui.startMonitoring(configManager)
		} else {
			gui.Log(LogLevelInfo, "停止监控")
			upload.CloseWatcher()
			close(gui.stopChan)
			gui.isMonitoring = false
			gui.isAutoStarting = false
			gui.countdownStopped = true // 确保倒计时停止
			gui.updateStatus()
		}
	})

	// 添加窗口关闭事件处理
	gui.window.Closing().Attach(func(canceled *bool, reason walk.CloseReason) {
		*canceled = true  // 取消默认的关闭行为
		gui.window.Hide() // 隐藏窗口而不是关闭
	})

	return &gui, nil
}

// Run 运行窗口
func (gui *GUI) Run() {
	gui.window.Run()
}

// NewTray 创建托盘
func (gui *GUI) NewTray() (*Tray, error) {
	return NewTray(gui.window)
}

// Log 记录日志
func (gui *GUI) Log(level string, format string, args ...interface{}) {
	// 从对象池获取 Builder
	builder := messagePool.Get().(*strings.Builder)
	builder.Reset()
	defer messagePool.Put(builder)

	// 格式化时间
	builder.WriteString(time.Now().Format("2006-01-02 15:04:05"))
	builder.WriteByte(' ')

	// 添加日志级别
	switch level {
	case LogLevelError:
		builder.WriteString("[错误]")
	case LogLevelWarning:
		builder.WriteString("[警告]")
	case LogLevelSuccess:
		builder.WriteString("[成功]")
	case LogLevelInfo:
		builder.WriteString("[信息]")
	default:
		builder.WriteString("[日志]")
	}
	builder.WriteByte(' ')

	// 格式化消息
	fmt.Fprintf(builder, format, args...)
	builder.WriteString("\r\n")

	// 在主线程更新 UI
	gui.window.Synchronize(func() {
		// 获取当前文本长度
		currentLen := gui.logView.TextLength()

		// 追加新文本
		gui.logView.SetTextSelection(currentLen, currentLen)
		gui.logView.ReplaceSelectedText(builder.String(), false)

		// 滚动到底部
		gui.logView.SendMessage(win.WM_VSCROLL, win.SB_BOTTOM, 0)
	})
}

// setAutoStart 设置开机启动
func (gui *GUI) setAutoStart(enable bool) error {
	// 获取当前可执行文件路径
	exePath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("获取可执行文件路径失败: %v", err)
	}

	// 获取注册表键
	key, err := registry.OpenKey(registry.CURRENT_USER,
		`Software\Microsoft\Windows\CurrentVersion\Run`,
		registry.ALL_ACCESS)
	if err != nil {
		return fmt.Errorf("打开注册表失败: %v", err)
	}
	defer key.Close()

	if enable {
		// 添加到开机启动，添加 autostart 参数
		err = key.SetStringValue("CursorHistory", exePath+" -autostart")
	} else {
		// 从开机启动中移除
		err = key.DeleteValue("CursorHistory")
		// 如果键不存在，忽略错误
		if err == registry.ErrNotExist {
			err = nil
		}
	}

	if err != nil {
		return fmt.Errorf("设置开机启动失败: %v", err)
	}
	return nil
}

// 添加开始监的方法
func (gui *GUI) startMonitoring(configManager *storage.ConfigManager) {
	if gui.isMonitoring {
		return
	}

	apiKey := gui.apiKeyEntry.Text()
	if apiKey == "" {
		gui.status.SetText("请输入 API Key")
		gui.Log(LogLevelError, "未输入 API Key")
		gui.isAutoStarting = false
		gui.countdownStopped = true // 如果启动失败，停止倒计时
		return
	}

	// 验证 token
	gui.status.SetText("正在验证 API Key...")
	if err := upload.ValidateApiKey(apiKey, app.Config.ServerURL); err != nil {
		gui.status.SetText(fmt.Sprintf("API Key 无效: %v", err))
		gui.Log(LogLevelError, "API Key 验证失败: %v", err)
		gui.isAutoStarting = false
		gui.countdownStopped = true // 如果验证失败，停止倒计时
		return
	}
	gui.Log(LogLevelSuccess, "API Key 验证成功")

	// 保存 token
	if err := configManager.SaveApiKey(apiKey); err != nil {
		gui.status.SetText(fmt.Sprintf("保存API Key失败: %v", err))
		gui.isAutoStarting = false
		gui.countdownStopped = true // 如果保存失败，停止倒计时
		return
	}

	app.Config.ApiKey = apiKey
	gui.isMonitoring = true
	gui.updateStatus()

	// 创建新的停止通道
	gui.stopChan = make(chan struct{})

	// 在新的 goroutine 中启动监控
	go func() {
		gui.Log(LogLevelInfo, "开始监控目录")
		searchPath := filepath.Join(os.Getenv("APPDATA"), "Cursor", "User", "workspaceStorage")

		done := make(chan error, 1)
		go func() {
			done <- upload.WatchDirectory(searchPath, configManager, gui)
		}()

		select {
		case <-gui.stopChan:
			gui.window.Synchronize(func() {
				gui.isMonitoring = false
				gui.updateStatus()
				gui.Log(LogLevelInfo, "监控已停止")
			})
		case err := <-done:
			if err != nil {
				gui.window.Synchronize(func() {
					gui.isMonitoring = false
					gui.updateStatus()
					gui.Log(LogLevelError, "监控发生错误: %v", err)
				})
			}
		}
	}()
}

// 添加更新状态的方法
func (gui *GUI) updateStatus() {
	gui.window.Synchronize(func() {
		if gui.isMonitoring {
			gui.startBtn.SetText("停止监控")
			gui.status.SetText("正在监控中...")
			gui.modeStatus.SetText("状态: 监控中")
			gui.statusIndicator.SetIcon(gui.runningIcon)
		} else {
			gui.startBtn.SetText("开始监控")
			gui.status.SetText("就绪")
			gui.modeStatus.SetText("状态: 未监控")
			gui.statusIndicator.SetIcon(gui.stoppedIcon)
		}

		// 更新 API Key 状态
		apiKey := gui.apiKeyEntry.Text()
		if apiKey == "" {
			gui.apiKeyStatus.SetText("Key: 未设置")
		} else {
			apiKeyDisplay := apiKey
			if len(apiKey) > 10 {
				apiKeyDisplay = apiKey[:10] + "..."
			}
			gui.apiKeyStatus.SetText("Key: " + apiKeyDisplay)
		}
	})
}

// Close 实现 Logger 接口 Close 方法
func (g *GUI) Close() error {
	if g.isMonitoring {
		close(g.stopChan) // 关闭通道而不是发送信号
		g.isMonitoring = false

		g.window.Synchronize(func() {
			g.updateStatus()
			g.Log(LogLevelInfo, "监控已停止")
		})
	}
	return nil
}

// 在 GUI 结构体中添加 Hide 和 Show 方法
func (gui *GUI) Hide() {
	if gui.window != nil {
		gui.window.Hide()
	}
}

func (gui *GUI) Show() {
	if gui.window != nil {
		gui.window.Show()
	}
}

// 添加倒计时方法
func (gui *GUI) startCountdown() {
	// 每秒更新一次状态栏
	ticker := time.NewTicker(time.Second)
	go func() {
		defer ticker.Stop()
		for range ticker.C {
			gui.window.Synchronize(func() {
				// 如果已经在监控中或倒计时已停止，则退出
				if gui.isMonitoring || gui.countdownStopped {
					return
				}

				if gui.countdownSec > 0 {
					gui.status.SetText(fmt.Sprintf("将在 %d 秒后自动开始监控...", gui.countdownSec))
					gui.countdownSec--
				} else {
					// 检查是否已经在监控中或倒计时已停止
					if !gui.isMonitoring && !gui.isAutoStarting && !gui.countdownStopped {
						gui.isAutoStarting = true
						gui.startMonitoring(gui.config)
						gui.status.SetText("监控已自动启动")
					}
					return
				}
			})
		}
	}()
}
