package gui

import (
	"bytes"
	"encoding/binary"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"os"
)

// ICO 文件头结构
type iconHeader struct {
	Reserved uint16
	Type     uint16
	Count    uint16
}

// ICO 目录项结构
type iconDirEntry struct {
	Width       byte
	Height      byte
	ColorCount  byte
	Reserved    byte
	Planes      uint16
	BitCount    uint16
	BytesInRes  uint32
	ImageOffset uint32
}

// 生成图标文件
func generateIcons() error {
	// 确保 assets 目录存在
	if err := os.MkdirAll("assets", 0755); err != nil {
		return err
	}

	// 生成绿色圆点图标
	if err := generateDotIcon("assets/green-dot.ico", color.RGBA{R: 0, G: 255, B: 0, A: 255}); err != nil {
		return err
	}

	// 生成灰色圆点图标
	if err := generateDotIcon("assets/gray-dot.ico", color.RGBA{R: 128, G: 128, B: 128, A: 255}); err != nil {
		return err
	}

	return nil
}

// 生成单个圆点图标
func generateDotIcon(path string, dotColor color.Color) error {
	// 创建一个 16x16 的 RGBA 图像
	img := image.NewRGBA(image.Rect(0, 0, 16, 16))

	// 填充透明背景
	draw.Draw(img, img.Bounds(), image.Transparent, image.Point{}, draw.Src)

	// 绘制圆点
	center := image.Point{X: 8, Y: 8}
	radius := 6
	for y := -radius; y <= radius; y++ {
		for x := -radius; x <= radius; x++ {
			if x*x+y*y <= radius*radius {
				img.Set(center.X+x, center.Y+y, dotColor)
			}
		}
	}

	// 创建一个缓冲区来存储 PNG 数据
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		return err
	}

	// 创建 ICO 文件
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	// 写入 ICO 文件头
	header := iconHeader{
		Reserved: 0,
		Type:     1,
		Count:    1,
	}
	if err := binary.Write(f, binary.LittleEndian, &header); err != nil {
		return err
	}

	// 写入目录项
	entry := iconDirEntry{
		Width:       16,
		Height:      16,
		ColorCount:  0,
		Reserved:    0,
		Planes:      1,
		BitCount:    32,
		BytesInRes:  uint32(buf.Len()),
		ImageOffset: 22, // 6 (header) + 16 (directory entry)
	}
	if err := binary.Write(f, binary.LittleEndian, &entry); err != nil {
		return err
	}

	// 写入图像数据
	if _, err := f.Write(buf.Bytes()); err != nil {
		return err
	}

	return nil
}
