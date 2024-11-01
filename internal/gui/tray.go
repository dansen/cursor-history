package gui

import (
	"cursor_history/internal/assets"
	"fmt"
	"os"
	"path/filepath"

	"github.com/lxn/walk"
)

// Tray 系统托盘
type Tray struct {
	ni         *walk.NotifyIcon
	mainWindow *walk.MainWindow
}

// NewTray 创建系统托盘
func NewTray(mw *walk.MainWindow) (*Tray, error) {
	ni, err := walk.NewNotifyIcon(mw)
	if err != nil {
		return nil, err
	}

	tray := &Tray{
		ni:         ni,
		mainWindow: mw,
	}

	// 从嵌入的资源中读取图标
	iconData, err := assets.Assets.ReadFile("logo.ico")
	if err != nil {
		return nil, fmt.Errorf("读取图标资源失败: %v", err)
	}

	// 创建临时文件
	tempDir := os.TempDir()
	tempIconPath := filepath.Join(tempDir, "cursor_history_logo.ico")

	// 写入临时文件
	if err := os.WriteFile(tempIconPath, iconData, 0644); err != nil {
		return nil, fmt.Errorf("写入临时图标文件失败: %v", err)
	}

	// 设置托盘图标
	icon, err := walk.NewIconFromFile(tempIconPath)
	if err != nil {
		return nil, fmt.Errorf("加载图标失败: %v", err)
	}

	// 删除临时文件
	defer os.Remove(tempIconPath)

	ni.SetIcon(icon)
	ni.SetToolTip("Cursor History")

	// 创建托盘菜单
	menu := ni.ContextMenu()

	showAction := walk.NewAction()
	showAction.SetText("显示")
	showAction.Triggered().Attach(tray.show)
	menu.Actions().Add(showAction)

	hideAction := walk.NewAction()
	hideAction.SetText("隐藏")
	hideAction.Triggered().Attach(tray.hide)
	menu.Actions().Add(hideAction)

	menu.Actions().Add(walk.NewSeparatorAction())

	exitAction := walk.NewAction()
	exitAction.SetText("退出")
	exitAction.Triggered().Attach(tray.exit)
	menu.Actions().Add(exitAction)

	// 双击托盘图标显示窗口
	ni.MouseDown().Attach(func(x, y int, button walk.MouseButton) {
		if button == walk.LeftButton {
			tray.show()
		}
	})

	// 显示托盘图标
	ni.SetVisible(true)

	return tray, nil
}

// Dispose 释放托盘资源
func (t *Tray) Dispose() {
	if t.ni != nil {
		t.ni.SetVisible(false)
		t.ni.Dispose()
	}
}

func (t *Tray) show() {
	t.mainWindow.Show()
}

func (t *Tray) hide() {
	t.mainWindow.Hide()
}

func (t *Tray) exit() {
	// 先清理托盘图标
	t.Dispose()

	// 使用 walk.App().Exit(0) 来退出程序
	walk.App().Exit(0)
}
