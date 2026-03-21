package app

import (
	"WeMediaSpider/backend/internal/models"
	"WeMediaSpider/backend/pkg/logger"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// LoadConfig 加载配置
func (a *App) LoadConfig() (models.Config, error) {
	return a.configManager.Load()
}

// SaveConfig 保存配置
func (a *App) SaveConfig(config models.Config) error {
	return a.configManager.Save(config)
}

// GetDefaultConfig 获取默认配置
func (a *App) GetDefaultConfig() models.Config {
	return a.configManager.GetDefault()
}

// SelectDirectory 选择目录
func (a *App) SelectDirectory() (string, error) {
	return runtime.OpenDirectoryDialog(a.ctx, runtime.OpenDialogOptions{
		Title: "选择输出目录",
	})
}

// ClearCache 清除缓存
func (a *App) ClearCache() error {
	if a.cacheManager == nil {
		return nil
	}
	logger.Log.Info("Clearing all cache")
	return a.cacheManager.ClearAll()
}

// ClearExpiredCache 清除过期缓存
func (a *App) ClearExpiredCache() error {
	if a.cacheManager == nil {
		return nil
	}
	logger.Log.Info("Clearing expired cache")
	return a.cacheManager.ClearExpired()
}

// GetCacheStats 获取缓存统计
func (a *App) GetCacheStats() (map[string]int, error) {
	if a.cacheManager == nil {
		return map[string]int{}, nil
	}
	return a.cacheManager.GetStats()
}
