package main

import (
	"log"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"unsafe"

	"cursor_history/internal/app"
	"cursor_history/internal/gui"
	"cursor_history/internal/storage"

	"github.com/lxn/win"
)

var (
	kernel32                = syscall.NewLazyDLL("kernel32.dll")
	user32                  = syscall.NewLazyDLL("user32.dll")
	procCreateMutex         = kernel32.NewProc("CreateMutexW")
	procFindWindow          = user32.NewProc("FindWindowW")
	procShowWindow          = user32.NewProc("ShowWindow")
	procSetForegroundWindow = user32.NewProc("SetForegroundWindow")
)

func createMutex(name string) (syscall.Handle, error) {
	ret, _, err := procCreateMutex.Call(
		0,
		0,
		uintptr(unsafe.Pointer(syscall.StringToUTF16Ptr(name))),
	)

	if ret == 0 {
		return 0, err
	}

	return syscall.Handle(ret), nil
}

func findAndActivateExistingWindow() bool {
	// 查找已存在的窗口
	hwnd, _, _ := procFindWindow.Call(
		0,
		uintptr(unsafe.Pointer(syscall.StringToUTF16Ptr("Cursor History"))),
	)

	if hwnd != 0 {
		// 显示窗口
		procShowWindow.Call(hwnd, win.SW_RESTORE)
		// 将窗口置于前台
		procSetForegroundWindow.Call(hwnd)
		return true
	}
	return false
}

func main() {
	// 获取可执行文件路径
	exePath, err := os.Executable()
	if err != nil {
		log.Fatal("获取可执行文件路径失败:", err)
	}
	exeDir := filepath.Dir(exePath)

	// 切换工作目录到可执行文件所在目录
	if err := os.Chdir(exeDir); err != nil {
		log.Fatal("切换工作目录失败:", err)
	}

	// 设置日志输出
	logFile, err := os.OpenFile("cursor.log", os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatal("无法创建日志文件:", err)
	}
	defer logFile.Close()
	log.SetOutput(logFile)

	// 记录启动信息和路径
	log.Printf("应用程序启动")
	log.Printf("可执行文件路径: %s", exePath)
	log.Printf("工作目录: %s", exeDir)

	// 创建命名互斥锁
	mutex, err := createMutex("Global\\CursorHistory")
	if err != nil {
		log.Fatal("创建互斥锁失败:", err)
	}
	defer syscall.CloseHandle(mutex)

	// 检查是否已经有实例在运行
	if syscall.GetLastError() == syscall.ERROR_ALREADY_EXISTS {
		log.Println("程序已经在运行")
		return
	}

	if findAndActivateExistingWindow() {
		log.Println("程序已经在运行，已激活现有窗口")
		return
	}

	// 初始化应用配置
	app.InitApp("prod") // 默认使用 prod，但如果有环境变量则使用环境变量的值
	log.Println("应用配置初始化完成")

	// 获取应用数据目录
	appDataDir, err := os.UserConfigDir()
	if err != nil {
		log.Fatal("获取配置目录失败:", err)
	}

	// 创建应用配置目录
	configDir := filepath.Join(appDataDir, "CursorHistory")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		log.Fatal("创建配置目录失败:", err)
	}
	log.Printf("配置目录: %s", configDir)

	// 初始化数据库
	dbPath := filepath.Join(configDir, "config.db")
	log.Printf("数据库路径: %s", dbPath)

	configManager, err := storage.NewConfigManager(dbPath)
	if err != nil {
		log.Fatal("初始化配置管理器失败:", err)
	}
	defer configManager.Close()
	log.Println("配置管理器初始化完成")

	// 创建 GUI
	mainWindow, err := gui.NewGUI(configManager)
	if err != nil {
		log.Fatal("创建窗口失败:", err)
	}
	log.Println("GUI 创建完成")

	// 创建托盘图标
	tray, err := mainWindow.NewTray()
	if err != nil {
		log.Fatal("创建托盘图标失败:", err)
	}
	defer tray.Dispose()
	log.Println("托盘图标创建完成")

	// 检查是否是开机启动
	isAutoStart := false
	for _, arg := range os.Args {
		if strings.Contains(strings.ToLower(arg), "autostart") {
			isAutoStart = true
			break
		}
	}

	// 如果是开机启动，则隐藏主窗口
	if isAutoStart {
		mainWindow.Hide()
		log.Println("开机启动模式，窗口已隐藏")
	}

	// 运行主窗口
	log.Println("开始运行主窗口")
	mainWindow.Run()
	log.Println("应用程序退出")
}
