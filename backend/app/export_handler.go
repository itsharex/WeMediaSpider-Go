package app

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"WeMediaSpider/backend/internal/database"
	dbmodels "WeMediaSpider/backend/internal/database/models"
	"WeMediaSpider/backend/internal/export"
	"WeMediaSpider/backend/internal/models"
	"WeMediaSpider/backend/pkg/logger"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

func (a *App) SelectSaveFile(defaultFilename string, filters []runtime.FileFilter) (string, error) {
	return runtime.SaveFileDialog(a.ctx, runtime.SaveDialogOptions{
		Title:           "保存文件",
		DefaultFilename: defaultFilename,
		Filters:         filters,
	})
}

// ============================================================
// 导出相关
// ============================================================

// ExportArticles 导出文章
func (a *App) ExportArticles(articles []models.Article, format string, filename string) error {
	logger.Infof("Exporting %d articles to %s format: %s", len(articles), format, filename)
	exporter := export.GetExporter(format)
	err := exporter.Export(articles, filename)
	if err != nil {
		logger.Errorf("Export failed: %v", err)
	} else {
		logger.Infof("Export completed successfully")
		// 保存导出统计
		if a.statsRepo != nil {
			if err := a.statsRepo.IncrementExports(); err != nil {
				logger.Errorf("Failed to update export stats: %v", err)
			}
		}
	}
	return err
}

// ============================================================
// 缓存相关
// ============================================================

func (a *App) ImportJSONFile(filePath string) error {
	if a.db == nil || a.articleRepo == nil {
		return fmt.Errorf("database not initialized")
	}

	// 读取 JSON 文件
	data, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	// 解析 JSON
	var savedData struct {
		Articles   []models.Article `json:"articles"`
		SaveTime   string           `json:"saveTime"`
		TotalCount int              `json:"totalCount"`
		Accounts   []string         `json:"accounts"`
	}

	if err := json.Unmarshal(data, &savedData); err != nil {
		return fmt.Errorf("failed to parse JSON: %w", err)
	}

	logger.Infof("Importing %d articles from %s", len(savedData.Articles), filePath)

	// 转换并保存到数据库
	dbArticles := make([]*dbmodels.Article, 0, len(savedData.Articles))
	for i := range savedData.Articles {
		article := &savedData.Articles[i]

		// 查找或创建公众号
		account, err := a.accountRepo.FindOrCreate(article.AccountFakeid, article.AccountName)
		if err != nil {
			logger.Warnf("Failed to find or create account: %v", err)
			continue
		}

		dbArticle := database.ConvertToDBArticle(article, account.ID)
		dbArticles = append(dbArticles, dbArticle)
	}

	// 批量保存
	if len(dbArticles) > 0 {
		if err := a.articleRepo.BatchCreate(dbArticles); err != nil {
			return fmt.Errorf("failed to save articles: %w", err)
		}
		logger.Infof("Successfully imported %d articles", len(dbArticles))
	}

	// 更新统计信息
	totalArticles, _ := a.articleRepo.Count()
	accounts, _ := a.accountRepo.List()
	todayArticles := database.CalculateTodayArticles(savedData.Articles)
	lastScrapeTime := time.Now().Format("2006-01-02 15:04:05")

	if err := a.statsRepo.UpdateArticleStats(
		int(totalArticles),
		len(accounts),
		todayArticles,
		lastScrapeTime,
	); err != nil {
		logger.Warnf("Failed to update stats: %v", err)
	}

	return nil
}

// ExportToJSON 导出数据库数据到 JSON 文件
func (a *App) ExportToJSON(dateOrPath string) (string, error) {
	if a.db == nil || a.articleRepo == nil {
		return "", fmt.Errorf("database not initialized")
	}

	// 解析日期
	date, err := time.Parse("2006-01-02", dateOrPath)
	if err != nil {
		return "", fmt.Errorf("invalid date format: %w", err)
	}

	// 查询该日期的文章
	startDate := date
	endDate := date.Add(24 * time.Hour)
	dbArticles, err := a.articleRepo.FindByDateRange(startDate, endDate, 0, 0)
	if err != nil {
		return "", fmt.Errorf("failed to load articles: %w", err)
	}

	articles := database.ConvertToAppArticles(dbArticles)

	// 构建保存数据
	accounts := database.ExtractAccountNames(articles)
	savedData := struct {
		Articles   []models.Article `json:"articles"`
		SaveTime   string           `json:"saveTime"`
		TotalCount int              `json:"totalCount"`
		Accounts   []string         `json:"accounts"`
	}{
		Articles:   articles,
		SaveTime:   time.Now().Format("2006-01-02 15:04:05"),
		TotalCount: len(articles),
		Accounts:   accounts,
	}

	// 序列化为 JSON
	jsonData, err := json.MarshalIndent(savedData, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal JSON: %w", err)
	}

	// 打开保存对话框
	filename := fmt.Sprintf("export_%s.json", dateOrPath)
	savePath, err := runtime.SaveFileDialog(a.ctx, runtime.SaveDialogOptions{
		Title:           "导出为 JSON 文件",
		DefaultFilename: filename,
		Filters: []runtime.FileFilter{
			{
				DisplayName: "JSON 文件 (*.json)",
				Pattern:     "*.json",
			},
		},
	})

	if err != nil || savePath == "" {
		return "", fmt.Errorf("用户取消操作")
	}

	// 写入文件
	if err := os.WriteFile(savePath, jsonData, 0644); err != nil {
		return "", fmt.Errorf("failed to write file: %w", err)
	}

	logger.Infof("Exported %d articles to %s", len(articles), savePath)
	return savePath, nil
}

// ============================================================
// 定时任务管理 API
// ============================================================

