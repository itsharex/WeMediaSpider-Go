package windows

import (
	"syscall"
	"unsafe"

	"WeMediaSpider/backend/pkg/logger"
)

// Manager Windows平台特定功能管理器
type Manager struct {
	kernel32 *syscall.LazyDLL
	user32   *syscall.LazyDLL
}

// NewManager 创建Windows管理器
func NewManager() *Manager {
	return &Manager{
		kernel32: syscall.NewLazyDLL("kernel32.dll"),
		user32:   syscall.NewLazyDLL("user32.dll"),
	}
}

// EnsureSingleInstance 确保单实例运行
// 返回 true 表示这是第一个实例，false 表示已存在实例（已尝试激活）
func (m *Manager) EnsureSingleInstance(windowTitle string) bool {
	procCreateMutex := m.kernel32.NewProc("CreateMutexW")
	mutexName := "Global\\WeMediaSpider_SingleInstance_Mutex"
	mutexNamePtr, _ := syscall.UTF16PtrFromString(mutexName)
	_, _, err := procCreateMutex.Call(
		0,
		0,
		uintptr(unsafe.Pointer(mutexNamePtr)),
	)

	if err != nil {
		if errno, ok := err.(syscall.Errno); ok && errno == syscall.ERROR_ALREADY_EXISTS {
			// 程序已在运行，尝试激活现有窗口
			m.activateExistingWindow(windowTitle)
			return false
		}
	}

	return true
}

// CalculateWindowSize 根据屏幕分辨率计算合适的窗口尺寸
func (m *Manager) CalculateWindowSize() (width, height int) {
	screenW, screenH := m.getScreenSize()

	// 基准: 1100x700 @ 1920x1080
	width = screenW * 1100 / 1920
	height = screenH * 700 / 1080

	// 最小尺寸限制
	if width < 900 {
		width = 900
	}
	if height < 580 {
		height = 580
	}

	return width, height
}

// getScreenSize 获取屏幕分辨率
func (m *Manager) getScreenSize() (int, int) {
	procGetSystemMetrics := m.user32.NewProc("GetSystemMetrics")
	const (
		SM_CXSCREEN = 0
		SM_CYSCREEN = 1
	)

	w, _, _ := procGetSystemMetrics.Call(SM_CXSCREEN)
	h, _, _ := procGetSystemMetrics.Call(SM_CYSCREEN)

	if w == 0 || h == 0 {
		return 1920, 1080 // fallback
	}

	return int(w), int(h)
}

// activateExistingWindow 激活已存在的窗口
func (m *Manager) activateExistingWindow(windowTitle string) {
	procFindWindow := m.user32.NewProc("FindWindowW")
	procShowWindow := m.user32.NewProc("ShowWindow")
	procSetForegroundWindow := m.user32.NewProc("SetForegroundWindow")

	const SW_RESTORE = 9

	titlePtr, _ := syscall.UTF16PtrFromString(windowTitle)
	hwnd, _, _ := procFindWindow.Call(0, uintptr(unsafe.Pointer(titlePtr)))

	if hwnd != 0 {
		showWindow(hwnd, SW_RESTORE, procShowWindow)
		setForegroundWindow(hwnd, procSetForegroundWindow)
		logger.Log.Info("已激活现有程序窗口")
	} else {
		logger.Log.Warn("未找到现有程序窗口，可能在托盘运行")
	}
}

// 辅助函数
func showWindow(hwnd uintptr, nCmdShow int, proc *syscall.LazyProc) bool {
	ret, _, _ := proc.Call(hwnd, uintptr(nCmdShow))
	return ret != 0
}

func setForegroundWindow(hwnd uintptr, proc *syscall.LazyProc) bool {
	ret, _, _ := proc.Call(hwnd)
	return ret != 0
}
