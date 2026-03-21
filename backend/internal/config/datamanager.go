package config

// Deprecated: 此模块已废弃，请使用 backend/internal/repository/stats_repo.go
// 保留此文件仅用于向后兼容

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
	"time"

	"WeMediaSpider/backend/internal/models"
	"WeMediaSpider/backend/pkg/logger"
	"WeMediaSpider/backend/pkg/timeutil"

	"go.uber.org/zap"
)

// DataManager 数据管理器
type DataManager struct {
	dataPath string
	mu       sync.RWMutex
}

// NewDataManager 创建数据管理器
func NewDataManager() *DataManager {
	// 获取用户主目录
	homeDir, err := os.UserHomeDir()
	if err != nil {
		logger.Log.Error("无法获取用户主目录", zap.Error(err))
		homeDir = "."
	}

	// 创建配置目录 C:\Users\用户名\.wemediaspider
	configDir := filepath.Join(homeDir, ".wemediaspider")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		logger.Log.Error("无法创建配置目录", zap.Error(err))
	}

	dataPath := filepath.Join(configDir, "appdata.json")
	logger.Log.Info("数据文件路径", zap.String("path", dataPath))

	return &DataManager{
		dataPath: dataPath,
	}
}

// LoadData 加载应用数据
func (m *DataManager) LoadData() (models.AppData, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var data models.AppData

	// 如果文件不存在,返回默认值
	if _, err := os.Stat(m.dataPath); os.IsNotExist(err) {
		return data, nil
	}

	// 读取文件
	fileData, err := os.ReadFile(m.dataPath)
	if err != nil {
		return data, err
	}

	// 解析 JSON
	if err := json.Unmarshal(fileData, &data); err != nil {
		return data, err
	}

	// 检查日期是否改变，如果不是今天则重置今日文章数
	now := timeutil.Now()
	today := now.Format("2006-01-02")
	if data.LastScrapeDate != today {
		data.TodayArticles = 0
	}

	logger.Log.Info("加载应用数据",
		zap.Int("articles", data.TotalArticles),
		zap.Int("accounts", data.TotalAccounts),
		zap.Int("today", data.TodayArticles),
	)
	return data, nil
}

// SaveData 保存应用数据
func (m *DataManager) SaveData(data models.AppData) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// 序列化为 JSON
	jsonData, err := json.Marshal(data)
	if err != nil {
		return err
	}

	// 写入文件
	if err := os.WriteFile(m.dataPath, jsonData, 0644); err != nil {
		return err
	}

	logger.Log.Info("保存应用数据",
		zap.Int("articles", data.TotalArticles),
		zap.Int("accounts", data.TotalAccounts),
	)
	return nil
}

// UpdateStats 更新统计数据
func (m *DataManager) UpdateStats(articles []models.Article) error {
	if len(articles) == 0 {
		return nil
	}

	data, err := m.LoadData()
	if err != nil {
		data = models.AppData{}
	}

	data.TotalArticles = len(articles)

	// 统计公众号
	accountSet := make(map[string]bool)
	for _, article := range articles {
		accountSet[article.AccountName] = true
	}

	data.TotalAccounts = len(accountSet)
	data.AccountNames = make([]string, 0, len(accountSet))
	for name := range accountSet {
		data.AccountNames = append(data.AccountNames, name)
	}

	// 计算今日文章数
	now := timeutil.Now()
	todayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, timeutil.GetChinaLocation())
	todayArticles := 0
	for _, article := range articles {
		publishTime := time.Unix(article.PublishTimestamp, 0)
		if publishTime.After(todayStart) || publishTime.Equal(todayStart) {
			todayArticles++
		}
	}
	data.TodayArticles = todayArticles
	data.LastScrapeDate = now.Format("2006-01-02")

	data.LastScrapeTime = articles[0].PublishTime
	data.LastUpdateTime = articles[0].PublishTime

	return m.SaveData(data)
}

// IncrementExports 增加导出次数
func (m *DataManager) IncrementExports() error {
	data, err := m.LoadData()
	if err != nil {
		data = models.AppData{}
	}

	data.TotalExports++
	return m.SaveData(data)
}

// IncrementImages 增加图片下载数
func (m *DataManager) IncrementImages(count int) error {
	data, err := m.LoadData()
	if err != nil {
		data = models.AppData{}
	}

	data.TotalImages += count
	return m.SaveData(data)
}
