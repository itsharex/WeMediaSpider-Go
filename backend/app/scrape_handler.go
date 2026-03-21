package app

import (
	"context"
	"errors"
	"fmt"
	"os"
	"time"

	"WeMediaSpider/backend/internal/database"
	dbmodels "WeMediaSpider/backend/internal/database/models"
	"WeMediaSpider/backend/internal/models"
	"WeMediaSpider/backend/internal/spider"
	"WeMediaSpider/backend/pkg/logger"

	"github.com/wailsapp/wails/v2/pkg/runtime"
	"go.uber.org/zap"
)

func (a *App) Login() error {
	return a.loginManager.Login(a.ctx)
}

// Logout 退出登录
func (a *App) Logout() error {
	return a.loginManager.Logout()
}

// GetLoginStatus 获取登录状态
func (a *App) GetLoginStatus() models.LoginStatus {
	return a.loginManager.GetStatus()
}

// ClearLoginCache 清除登录缓存
func (a *App) ClearLoginCache() error {
	return a.loginManager.ClearCache()
}

// ExportCredentials 导出加密的登录凭证到文件
func (a *App) ExportCredentials() (string, error) {
	// 导出凭证数据
	data, err := a.loginManager.ExportCredentials()
	if err != nil {
		return "", err
	}

	// 打开保存文件对话框
	filepath, err := runtime.SaveFileDialog(a.ctx, runtime.SaveDialogOptions{
		Title:           "导出登录凭证",
		DefaultFilename: fmt.Sprintf("wemedia_credentials_%d.zgswx", time.Now().Unix()),
		Filters: []runtime.FileFilter{
			{
				DisplayName: "WeMediaSpider 凭证文件 (*.zgswx)",
				Pattern:     "*.zgswx",
			},
		},
	})

	if err != nil || filepath == "" {
		return "", fmt.Errorf("用户取消操作")
	}

	// 写入文件
	if err := os.WriteFile(filepath, data, 0600); err != nil {
		return "", fmt.Errorf("保存文件失败: %w", err)
	}

	logger.Log.Info("凭证已导出", zap.String("path", filepath))
	return filepath, nil
}

// ImportCredentials 从文件导入加密的登录凭证
func (a *App) ImportCredentials() error {
	// 打开文件选择对话框
	filepath, err := runtime.OpenFileDialog(a.ctx, runtime.OpenDialogOptions{
		Title: "导入登录凭证",
		Filters: []runtime.FileFilter{
			{
				DisplayName: "WeMediaSpider 凭证文件 (*.zgswx)",
				Pattern:     "*.zgswx",
			},
		},
	})

	if err != nil || filepath == "" {
		return fmt.Errorf("用户取消操作")
	}

	// 读取文件
	data, err := os.ReadFile(filepath)
	if err != nil {
		return fmt.Errorf("读取文件失败: %w", err)
	}

	// 导入凭证
	if err := a.loginManager.ImportCredentials(data); err != nil {
		return err
	}

	logger.Log.Info("凭证已从文件导入", zap.String("path", filepath))
	return nil
}

// ============================================================
// 爬取相关
// ============================================================

// SearchAccount 搜索公众号
func (a *App) SearchAccount(query string) ([]models.Account, error) {
	// 确保已登录
	status := a.loginManager.GetStatus()
	if !status.IsLoggedIn {
		return nil, nil
	}

	// 创建临时爬虫
	scraper := spider.NewScraper(
		a.loginManager.GetToken(),
		a.loginManager.GetHeaders(),
	)

	return scraper.SearchAccount(query)
}

// StartScrape 开始爬取
func (a *App) StartScrape(config models.ScrapeConfig) ([]models.Article, error) {
	// 创建异步爬虫（加锁保护赋值操作）
	a.scrapeMu.Lock()
	a.scraper = spider.NewAsyncScraper(
		a.loginManager.GetToken(),
		a.loginManager.GetHeaders(),
		config.MaxWorkers,
	)
	a.scrapeMu.Unlock()

	// 创建进度通道
	progressChan := make(chan models.Progress, 100)
	statusChan := make(chan models.AccountStatus, 100)

	// 启动进度发送协程
	go func() {
		for {
			select {
			case progress, ok := <-progressChan:
				if !ok {
					return
				}
				runtime.EventsEmit(a.ctx, "scrape:progress", progress)
			case status, ok := <-statusChan:
				if !ok {
					return
				}
				runtime.EventsEmit(a.ctx, "scrape:status", status)
			case <-a.ctx.Done():
				return
			}
		}
	}()

	// 执行爬取
	articles, err := a.scraper.BatchScrapeAsync(a.ctx, config, progressChan, statusChan)

	// 关闭通道
	close(progressChan)
	close(statusChan)

	// 发送完成事件
	if err == nil && len(articles) > 0 {
		// 保存到数据库
		if a.db != nil && a.articleRepo != nil {
			logger.Log.Info("保存文章到数据库")
			dbArticles := make([]*dbmodels.Article, 0, len(articles))

			for i := range articles {
				article := &articles[i]
				// 查找或创建公众号
				account, accErr := a.accountRepo.FindOrCreate(article.AccountFakeid, article.AccountName)
				if accErr != nil {
					logger.Log.Error("查找或创建公众号失败", zap.Error(accErr))
					continue
				}

				dbArticle := database.ConvertToDBArticle(article, account.ID)
				dbArticles = append(dbArticles, dbArticle)
			}

			// 批量保存到数据库
			if len(dbArticles) > 0 {
				if saveErr := a.articleRepo.BatchCreate(dbArticles); saveErr != nil {
				logger.Log.Error("批量保存文章到数据库失败", zap.Error(saveErr))
				} else {
					logger.Log.Info("文章已保存到数据库", zap.Int("count", len(dbArticles)))
				}
			}

			// 更新统计信息
			totalArticles, _ := a.articleRepo.Count()
			accounts, _ := a.accountRepo.List()
			todayArticles := database.CalculateTodayArticles(articles)
			lastScrapeTime := articles[0].PublishTime

			if statsErr := a.statsRepo.UpdateArticleStats(
				int(totalArticles),
				len(accounts),
				todayArticles,
				lastScrapeTime,
			); statsErr != nil {
				logger.Log.Error("更新统计信息失败", zap.Error(statsErr))
			}
		}

		runtime.EventsEmit(a.ctx, "scrape:completed", map[string]interface{}{
			"total": len(articles),
		})
	} else if err != nil {
		// 检查是否是取消操作导致的错误
		if !errors.Is(err, context.Canceled) && !errors.Is(err, context.DeadlineExceeded) {
			// 只有非取消的错误才发送错误事件
			runtime.EventsEmit(a.ctx, "scrape:error", map[string]string{
				"error": err.Error(),
			})
		}
	}

	return articles, err
}

// CancelScrape 取消爬取
func (a *App) CancelScrape() {
	a.scrapeMu.Lock()
	scraper := a.scraper
	a.scrapeMu.Unlock()
	if scraper != nil {
		scraper.Cancel()
	}
}

// ============================================================
// 配置相关
// ============================================================

// LoadConfig 加载配置
func (a *App) ExtractArticleImages(content string) []spider.ImageInfo {
	downloader := spider.NewImageDownloader(a.loginManager.GetHeaders())
	return downloader.ExtractImages(content)
}

// BatchDownloadImages 批量下载图片
func (a *App) BatchDownloadImages(images []spider.ImageInfo, baseDir string, maxWorkers int) error {
	// 创建新的下载器（加锁保护赋值操作）
	a.scrapeMu.Lock()
	a.imageDownloader = spider.NewImageDownloader(a.loginManager.GetHeaders())
	a.scrapeMu.Unlock()

	// 创建进度通道
	progressChan := make(chan spider.ImageDownloadProgress, 100)

	// 启动进度发送协程
	go func() {
		for progress := range progressChan {
			runtime.EventsEmit(a.ctx, "image:progress", progress)
		}
	}()

	// 执行下载
	err := a.imageDownloader.DownloadImagesWithProgress(images, baseDir, maxWorkers, progressChan)

	// 发送完成事件
	if err == nil {
		runtime.EventsEmit(a.ctx, "image:completed", map[string]interface{}{
			"total": len(images),
		})
		// 保存图片下载统计
		if a.statsRepo != nil {
			if err := a.statsRepo.IncrementImages(len(images)); err != nil {
				logger.Log.Error("更新图片下载统计失败", zap.Error(err))
			}
		}
	} else {
		// 检查是否是取消操作导致的错误
		if !errors.Is(err, context.Canceled) && !errors.Is(err, context.DeadlineExceeded) {
			runtime.EventsEmit(a.ctx, "image:error", map[string]string{
				"error": err.Error(),
			})
		}
	}

	return err
}

// CancelImageDownload 取消图片下载
func (a *App) CancelImageDownload() {
	a.scrapeMu.Lock()
	downloader := a.imageDownloader
	a.scrapeMu.Unlock()
	if downloader != nil {
		downloader.Cancel()
	}
}

// ============================================================
// 数据管理相关
// ============================================================

