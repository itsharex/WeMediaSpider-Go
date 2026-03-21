package main

import (
	"context"
	"embed"
	"flag"
	"fmt"
	"os"

	"WeMediaSpider/backend/app"
	"WeMediaSpider/backend/pkg/windows"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	wailsWindows "github.com/wailsapp/wails/v2/pkg/options/windows"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	// 全局 panic 恢复
	defer func() {
		if r := recover(); r != nil {
			fmt.Printf("程序发生严重错误: %v\n", r)
			fmt.Println("请查看日志文件获取详细信息")
			os.Exit(1)
		}
	}()

	winManager := windows.NewManager()

	// 单实例检测
	windowTitle := "WeMediaSpider - 微信公众号爬虫"
	if !winManager.EnsureSingleInstance(windowTitle) {
		os.Exit(0)
	}

	// 解析命令行参数
	silent := flag.Bool("silent", false, "Start in silent mode (hidden to tray)")
	flag.Parse()

	// Create an instance of the app structure
	application := app.NewApp()

	// 根据屏幕分辨率计算窗口尺寸
	winWidth, winHeight := winManager.CalculateWindowSize()

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
		Windows: &wailsWindows.Options{
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
