package config

import (
	"encoding/json"
	"os"
	"path/filepath"

	"WeMediaSpider/backend/pkg/logger"

	"go.uber.org/zap"
)

// SystemConfig 系统配置
type SystemConfig struct {
	CloseToTray         bool   `json:"closeToTray"`         // 关闭到托盘
	RememberChoice      bool   `json:"rememberChoice"`      // 记住用户选择
	UpdateIgnoredDate   string `json:"updateIgnoredDate"`   // 更新忽略日期 (YYYY-MM-DD)
}

// SystemConfigManager 系统配置管理器
type SystemConfigManager struct {
	configPath string
}

// NewSystemConfigManager 创建系统配置管理器
func NewSystemConfigManager() (*SystemConfigManager, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	configDir := filepath.Join(homeDir, ".wemediaspider")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return nil, err
	}

	return &SystemConfigManager{
		configPath: filepath.Join(configDir, "system_config.json"),
	}, nil
}

// Load 加载系统配置
func (m *SystemConfigManager) Load() (SystemConfig, error) {
	// 默认配置
	defaultConfig := SystemConfig{
		CloseToTray:       false, // 默认禁用关闭到托盘
		RememberChoice:    false, // 默认不记住选择
		UpdateIgnoredDate: "",    // 默认无忽略日期
	}

	// 检查文件是否存在
	if _, err := os.Stat(m.configPath); os.IsNotExist(err) {
		// 文件不存在，保存默认配置
		if err := m.Save(defaultConfig); err != nil {
			logger.Log.Warn("保存默认系统配置失败", zap.Error(err))
		}
		return defaultConfig, nil
	}

	// 读取文件
	data, err := os.ReadFile(m.configPath)
	if err != nil {
		logger.Log.Warn("读取系统配置失败", zap.Error(err))
		return defaultConfig, nil
	}

	// 解析 JSON
	var config SystemConfig
	if err := json.Unmarshal(data, &config); err != nil {
		logger.Log.Warn("解析系统配置失败", zap.Error(err))
		return defaultConfig, nil
	}

	logger.Log.Info("已加载系统配置",
		zap.Bool("close_to_tray", config.CloseToTray),
		zap.Bool("remember_choice", config.RememberChoice),
		zap.String("update_ignored_date", config.UpdateIgnoredDate),
	)
	return config, nil
}

// Save 保存系统配置
func (m *SystemConfigManager) Save(config SystemConfig) error {
	// 确保配置目录存在
	configDir := filepath.Dir(m.configPath)
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return err
	}

	// 序列化为 JSON
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return err
	}

	// 写入文件
	if err := os.WriteFile(m.configPath, data, 0644); err != nil {
		return err
	}

	logger.Log.Info("已保存系统配置",
		zap.Bool("close_to_tray", config.CloseToTray),
		zap.Bool("remember_choice", config.RememberChoice),
		zap.String("update_ignored_date", config.UpdateIgnoredDate),
	)
	return nil
}
