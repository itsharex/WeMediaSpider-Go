package main

import (
	"context"
	"embed"
	"flag"
	"fmt"
	"os"
	"syscall"
	"unsafe"

	"WeMediaSpider/backend/app"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"github.com/wailsapp/wails/v2/pkg/options/windows"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

//go:embed all:frontend/dist
var assets embed.FS

var (
	kernel32         = syscall.NewLazyDLL("kernel32.dll")
	user32           = syscall.NewLazyDLL("user32.dll")
	procCreateMutex  = kernel32.NewProc("CreateMutexW")
	procFindWindow   = user32.NewProc("FindWindowW")
	procShowWindow   = user32.NewProc("ShowWindow")
	procSetForegroundWindow = user32.NewProc("SetForegroundWindow")
	procGetSystemMetrics    = user32.NewProc("GetSystemMetrics")
)

const (
	SW_RESTORE       = 9
	SM_CXSCREEN      = 0
	SM_CYSCREEN      = 1
)

// getScreenSize 获取屏幕分辨率
func getScreenSize() (int, int) {
	w, _, _ := procGetSystemMetrics.Call(SM_CXSCREEN)
	h, _, _ := procGetSystemMetrics.Call(SM_CYSCREEN)
	if w == 0 || h == 0 {
		return 1920, 1080 // fallback
	}
	return int(w), int(h)
}

// calcWindowSize 根据屏幕分辨率计算窗口尺寸
// 基准: 1100x700 @ 1920x1080，占屏幕约 57%x65%
func calcWindowSize() (width, height int) {
	screenW, screenH := getScreenSize()

	// 按屏幕比例缩放，保持与基准相同的占比
	width = screenW * 1100 / 1920
	height = screenH * 700 / 1080

	// 限制最小尺寸
	if width < 900 {
		width = 900
	}
	if height < 580 {
		height = 580
	}

	return width, height
}

// createMutex 创建互斥锁，用于防止多实例运行
func createMutex(name string) (uintptr, error) {
	ret, _, err := procCreateMutex.Call(
		0,
		0,
		uintptr(unsafe.Pointer(syscall.StringToUTF16Ptr(name))),
	)
	switch int(err.(syscall.Errno)) {
	case 0:
		return ret, nil
	default:
		return ret, err
	}
}

// findWindow 查找窗口
func findWindow(className, windowName string) uintptr {
	var classNamePtr, windowNamePtr uintptr
	if className != "" {
		classNamePtr = uintptr(unsafe.Pointer(syscall.StringToUTF16Ptr(className)))
	}
	if windowName != "" {
		windowNamePtr = uintptr(unsafe.Pointer(syscall.StringToUTF16Ptr(windowName)))
	}
	ret, _, _ := procFindWindow.Call(classNamePtr, windowNamePtr)
	return ret
}

// showWindow 显示窗口
func showWindow(hwnd uintptr, nCmdShow int) bool {
	ret, _, _ := procShowWindow.Call(hwnd, uintptr(nCmdShow))
	return ret != 0
}

// setForegroundWindow 将窗口置于前台
func setForegroundWindow(hwnd uintptr) bool {
	ret, _, _ := procSetForegroundWindow.Call(hwnd)
	return ret != 0
}

func main() {
	// 全局 panic 恢复
	defer func() {
		if r := recover(); r != nil {
			fmt.Printf("程序发生严重错误: %v\n", r)
			fmt.Println("请查看日志文件获取详细信息")
			os.Exit(1)
		}
	}()

	// 单实例检测 - 使用互斥锁
	mutexName := "Global\\WeMediaSpider_SingleInstance_Mutex"
	_, mutexErr := createMutex(mutexName)
	if mutexErr != nil {
		// 如果互斥锁已存在，说明程序已经在运行
		if mutexErr.(syscall.Errno) == syscall.ERROR_ALREADY_EXISTS {
			fmt.Println("程序已经在运行中，尝试显示主窗口...")

			// 查找已运行的窗口并显示
			windowTitle := "WeMediaSpider - 微信公众号爬虫"
			hwnd := findWindow("", windowTitle)
			if hwnd != 0 {
				// 恢复窗口（如果最小化）
				showWindow(hwnd, SW_RESTORE)
				// 将窗口置于前台
				setForegroundWindow(hwnd)
				fmt.Println("已显示主窗口")
			} else {
				fmt.Println("未找到主窗口，可能程序在托盘中运行")
			}

			os.Exit(0)
		}
	}

	// 解析命令行参数
	silent := flag.Bool("silent", false, "Start in silent mode (hidden to tray)")
	flag.Parse()

	// Create an instance of the app structure
	application := app.NewApp()

	// 根据屏幕分辨率计算窗口尺寸
	winWidth, winHeight := calcWindowSize()

	// Create application with options
	err := wails.Run(&options.App{
		Title:     "WeMediaSpider - 微信公众号爬虫",
		Width:     winWidth,
		Height:    winHeight,
		MinWidth:  900,
		MinHeight: 580,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		BackgroundColour: &options.RGBA{R: 26, G: 26, B: 26, A: 1},
		OnStartup: func(ctx context.Context) {
			application.Startup(ctx)

			// 如果是静默启动，隐藏窗口
			if *silent {
				runtime.WindowHide(ctx)
			}
		},
		OnShutdown: application.Shutdown,
		OnBeforeClose: func(ctx context.Context) bool {
			// 检查是否应该阻止关闭
			return application.ShouldBlockClose()
		},
		Bind: []interface{}{
			application,
		},
		Frameless: true, // 隐藏系统标题栏
		Windows: &windows.Options{
			WebviewIsTransparent: false,
			WindowIsTranslucent:  false,
			DisableWindowIcon:    false,
		},
	})

	if err != nil {
		println("Error:", err.Error())
		os.Exit(1)
	}
}
