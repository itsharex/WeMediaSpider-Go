package cache

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"

	"WeMediaSpider/backend/internal/models"
	"WeMediaSpider/backend/pkg/timeutil"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

// cacheRow SQLite 缓存行
type cacheRow struct {
	Bucket    string `gorm:"primaryKey;size:32"`
	Key       string `gorm:"primaryKey;size:128"`
	Data      string `gorm:"type:text"`
	Timestamp int64
}

func (cacheRow) TableName() string { return "request_cache" }

// Manager 缓存管理器（SQLite 实现）
type Manager struct {
	db         *gorm.DB
	expireTime time.Duration
}

// NewManager 创建缓存管理器
func NewManager(expireHours int) (*Manager, error) {
	homeDir, _ := os.UserHomeDir()
	cacheDir := filepath.Join(homeDir, ".wemediaspider")
	os.MkdirAll(cacheDir, 0755)

	dbPath := filepath.Join(cacheDir, "cache.db")
	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{
		Logger: gormlogger.Default.LogMode(gormlogger.Silent),
	})
	if err != nil {
		return nil, err
	}

	db.Exec("PRAGMA journal_mode=WAL")
	db.Exec("PRAGMA synchronous=NORMAL")

	if err := db.AutoMigrate(&cacheRow{}); err != nil {
		return nil, err
	}

	return &Manager{
		db:         db,
		expireTime: time.Duration(expireHours) * time.Hour,
	}, nil
}

// Close 关闭数据库
func (m *Manager) Close() error {
	sqlDB, err := m.db.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}

func (m *Manager) set(bucket, key string, v any) error {
	data, err := json.Marshal(v)
	if err != nil {
		return err
	}
	row := cacheRow{
		Bucket:    bucket,
		Key:       key,
		Data:      string(data),
		Timestamp: timeutil.Now().Unix(),
	}
	return m.db.Save(&row).Error
}

func (m *Manager) get(bucket, key string, out any) bool {
	var row cacheRow
	if err := m.db.Where("bucket = ? AND key = ?", bucket, key).First(&row).Error; err != nil {
		return false
	}
	if time.Since(time.Unix(row.Timestamp, 0)) > m.expireTime {
		return false
	}
	return json.Unmarshal([]byte(row.Data), out) == nil
}

// SaveArticles 保存文章缓存
func (m *Manager) SaveArticles(accountFakeid string, articles []models.Article) error {
	return m.set("articles", accountFakeid, articles)
}

// GetArticles 获取文章缓存
func (m *Manager) GetArticles(accountFakeid string) ([]models.Article, bool) {
	var articles []models.Article
	if !m.get("articles", accountFakeid, &articles) || len(articles) == 0 {
		return nil, false
	}
	return articles, true
}

// SaveAccount 保存公众号信息
func (m *Manager) SaveAccount(account models.Account) error {
	return m.set("accounts", account.Fakeid, account)
}

// GetAccount 获取公众号信息
func (m *Manager) GetAccount(fakeid string) (models.Account, bool) {
	var account models.Account
	if !m.get("accounts", fakeid, &account) || account.Fakeid == "" {
		return models.Account{}, false
	}
	return account, true
}

// ClearExpired 清除过期缓存
func (m *Manager) ClearExpired() error {
	cutoff := timeutil.Now().Unix() - int64(m.expireTime.Seconds())
	return m.db.Where("timestamp < ?", cutoff).Delete(&cacheRow{}).Error
}

// ClearAll 清除所有缓存
func (m *Manager) ClearAll() error {
	return m.db.Where("1 = 1").Delete(&cacheRow{}).Error
}

// GetStats 获取缓存统计
func (m *Manager) GetStats() (map[string]int, error) {
	type result struct {
		Bucket string
		Count  int
	}
	var rows []result
	if err := m.db.Model(&cacheRow{}).Select("bucket, count(*) as count").Group("bucket").Scan(&rows).Error; err != nil {
		return nil, err
	}
	stats := map[string]int{"articles": 0, "accounts": 0}
	for _, r := range rows {
		stats[r.Bucket] = r.Count
	}
	return stats, nil
}
