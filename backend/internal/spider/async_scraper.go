package spider

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"WeMediaSpider/backend/internal/models"
	"WeMediaSpider/backend/pkg/logger"

	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
)

// AsyncScraper 异步爬虫
type AsyncScraper struct {
	*Scraper
	maxWorkers int
	mu         sync.Mutex
	progress   models.Progress
}

// NewAsyncScraper 创建异步爬虫
func NewAsyncScraper(token string, headers map[string]string, maxWorkers int) *AsyncScraper {
	return &AsyncScraper{
		Scraper:    NewScraper(token, headers),
		maxWorkers: maxWorkers,
	}
}

// BatchScrapeAsync 异步批量爬取
func (as *AsyncScraper) BatchScrapeAsync(
	ctx context.Context,
	config models.ScrapeConfig,
	progressChan chan<- models.Progress,
	statusChan chan<- models.AccountStatus,
) ([]models.Article, error) {
	// 创建可取消的上下文
	ctx, cancel := context.WithCancel(ctx)
	as.cancelFunc = cancel
	defer cancel()

	// 创建 errgroup
	g, ctx := errgroup.WithContext(ctx)
	g.SetLimit(as.maxWorkers)

	// 结果通道
	resultsChan := make(chan []models.Article, len(config.Accounts))

	// 为每个公众号启动 goroutine
	for _, accountName := range config.Accounts {
		accountName := accountName // 捕获循环变量

		g.Go(func() (err error) {
			// 添加 panic 恢复
			defer func() {
				if r := recover(); r != nil {
					logger.Log.Error("爬取过程发生 panic", zap.String("account", accountName), zap.Any("panic", r))
					err = fmt.Errorf("panic: %v", r)
					if statusChan != nil {
						statusChan <- models.AccountStatus{
							AccountName: accountName,
							Status:      "error",
							Message:     fmt.Sprintf("爬取异常: %v", r),
						}
					}
				}
			}()

			// 检查上下文是否已取消
			select {
			case <-ctx.Done():
				logger.Log.Warn("爬取被取消", zap.String("account", accountName))
				return ctx.Err()
			default:
			}

			logger.Log.Info("开始处理公众号", zap.String("account", accountName))

			// 发送状态
			if statusChan != nil {
				statusChan <- models.AccountStatus{
					AccountName: accountName,
					Status:      "searching",
					Message:     "正在搜索公众号...",
				}
			}

			// 搜索公众号
			logger.Log.Info("正在搜索公众号", zap.String("account", accountName))
			accounts, err := as.SearchAccount(accountName)
			if err != nil || len(accounts) == 0 {
				logger.Log.Error("未找到公众号", zap.String("account", accountName), zap.Error(err))
				if statusChan != nil {
					statusChan <- models.AccountStatus{
						AccountName: accountName,
						Status:      "error",
						Message:     "未找到公众号",
					}
				}
				return nil // 不中断其他任务
			}

			account := accounts[0]
			logger.Log.Info("找到公众号", zap.String("account", accountName), zap.String("fakeid", account.Fakeid), zap.String("alias", account.Alias))

			// 发送状态
			if statusChan != nil {
				statusChan <- models.AccountStatus{
					AccountName: accountName,
					Status:      "fetching",
					Message:     "正在获取文章列表...",
				}
			}

			// 获取文章列表（串行获取，避免频率限制）
			logger.Log.Info("开始获取文章列表", zap.String("account", accountName), zap.Int("maxPages", config.MaxPages))

			var allArticles []models.Article
			for page := 0; page < config.MaxPages; page++ {
				// 检查上下文是否已取消
				select {
				case <-ctx.Done():
					logger.Log.Warn("爬取被取消", zap.String("account", accountName))
					return ctx.Err()
				default:
				}

				// 获取文章列表
				articles, err := as.GetArticlesList(ctx, account.Fakeid, page)
				if err != nil {
					// 如果是频率限制错误，清空已获取的数据并停止
					if strings.Contains(err.Error(), "freq control") || strings.Contains(err.Error(), "频率") {
						logger.Log.Error("遇到频率限制，清空数据并停止爬取", zap.String("account", accountName), zap.Int("page", page+1))
						if statusChan != nil {
							statusChan <- models.AccountStatus{
								AccountName: accountName,
								Status:      "error",
								Message:     "遇到频率限制，请稍后重试",
							}
						}
						// 清空已获取的文章列表
						allArticles = nil
						return nil
					}

					logger.Log.Warn("获取文章列表失败", zap.String("account", accountName), zap.Int("page", page+1), zap.Error(err))
					break
				}

				if len(articles) == 0 {
					logger.Log.Info("第N页为空，停止获取", zap.Int("page", page+1), zap.String("account", accountName))
					break
				}

				logger.Log.Info("获取到文章", zap.String("account", accountName), zap.Int("page", page+1), zap.Int("maxPages", config.MaxPages), zap.Int("count", len(articles)))

				// 设置公众号名称
				for i := range articles {
					articles[i].AccountName = accountName
					articles[i].AccountFakeid = account.Fakeid
				}

				allArticles = append(allArticles, articles...)
			}

			logger.Log.Info("总共获取文章", zap.String("account", accountName), zap.Int("count", len(allArticles)))

			// 打印所有文章链接
			logger.Log.Info("文章链接列表", zap.String("account", accountName))
			for i, article := range allArticles {
				logger.Log.Info("article", zap.Int("index", i+1), zap.String("link", article.Link))
				logger.Log.Info("article title", zap.String("title", article.Title))
			}

			// 日期过滤
			if config.StartDate != "" && config.EndDate != "" {
				beforeFilter := len(allArticles)
				allArticles = as.FilterArticlesByDate(allArticles, config.StartDate, config.EndDate)
				logger.Log.Info("日期过滤", zap.String("account", accountName), zap.String("start", config.StartDate), zap.String("end", config.EndDate), zap.Int("before", beforeFilter), zap.Int("after", len(allArticles)))

				// 如果有文章被过滤掉，显示过滤后的文章列表
				if beforeFilter != len(allArticles) {
					logger.Log.Info("过滤后的文章列表", zap.String("account", accountName))
					for i, article := range allArticles {
						logger.Log.Info("filtered article", zap.Int("index", i+1), zap.String("title", article.Title), zap.String("publishTime", article.PublishTime))
					}
				}
			}

			// 发送获取到的文章数
			if statusChan != nil {
				statusChan <- models.AccountStatus{
					AccountName:  accountName,
					Status:       "fetching",
					Message:      fmt.Sprintf("已获取 %d 篇文章", len(allArticles)),
					ArticleCount: len(allArticles),
				}
			}

			// 获取文章内容
			if config.IncludeContent {
				if statusChan != nil {
					statusChan <- models.AccountStatus{
						AccountName:  accountName,
						Status:       "content",
						Message:      fmt.Sprintf("正在获取文章内容 (并发数: %d)...", as.maxWorkers),
						ArticleCount: len(allArticles),
					}
				}

				logger.Log.Info("开始获取文章内容", zap.String("account", accountName), zap.Int("total", len(allArticles)), zap.Int("workers", as.maxWorkers))

				// 使用 errgroup 并发获取文章内容
				contentGroup, contentCtx := errgroup.WithContext(ctx)
				contentGroup.SetLimit(as.maxWorkers)

				// 进度计数器
				var contentMu sync.Mutex
				contentProgress := 0
				successCount := 0

				for i := range allArticles {
					i := i // 捕获循环变量

					contentGroup.Go(func() error {
						select {
						case <-contentCtx.Done():
							return contentCtx.Err()
						default:
						}

						content, err := as.GetArticleContent(contentCtx, allArticles[i].Link)

						// 更新进度（无论成功失败）
						contentMu.Lock()
						contentProgress++
						currentProgress := contentProgress
						if err == nil {
							allArticles[i].Content = content
							successCount++
							logger.Log.Info("获取文章内容成功", zap.String("account", accountName), zap.Int("progress", currentProgress), zap.Int("total", len(allArticles)), zap.String("title", allArticles[i].Title), zap.Int("length", len(content)))
						} else {
							logger.Log.Warn("获取文章内容失败", zap.String("account", accountName), zap.Int("progress", currentProgress), zap.Int("total", len(allArticles)), zap.String("title", allArticles[i].Title), zap.Error(err))
						}
						contentMu.Unlock()

						// 发送进度和状态更新
						if progressChan != nil {
							progressChan <- models.Progress{
								Type:    models.ProgressTypeContent,
								Current: currentProgress,
								Total:   len(allArticles),
								Message: fmt.Sprintf("正在获取文章内容 [%s] (%d/%d)", accountName, currentProgress, len(allArticles)),
							}
						}

						// 同时更新账号状态，包含进度信息
						if statusChan != nil {
							statusChan <- models.AccountStatus{
								AccountName:  accountName,
								Status:       "content",
								Message:      fmt.Sprintf("正在获取文章内容 (%d/%d)", currentProgress, len(allArticles)),
								ArticleCount: len(allArticles),
								Progress: &models.ProgressInfo{
									Current: currentProgress,
									Total:   len(allArticles),
								},
							}
						}

						return nil
					})
				}

				// 等待所有内容获取完成
				if err := contentGroup.Wait(); err != nil {
					// 如果是取消错误，直接返回
					if err == context.Canceled {
						logger.Log.Warn("获取文章内容被取消", zap.String("account", accountName))
						return err
					}
					logger.Log.Error("获取文章内容过程中出错", zap.String("account", accountName), zap.Error(err))
				}

				logger.Log.Info("文章内容获取完成", zap.String("account", accountName), zap.Int("success", successCount), zap.Int("failed", len(allArticles)-successCount), zap.Int("total", len(allArticles)))
			}

			// 关键词过滤（在获取正文内容之后进行，以便搜索全文）
			if config.KeywordFilter != "" {
				beforeFilter := len(allArticles)
				allArticles = as.FilterArticlesByKeyword(allArticles, config.KeywordFilter)
				logger.Log.Info("关键词过滤", zap.String("account", accountName), zap.String("keyword", config.KeywordFilter), zap.Int("before", beforeFilter), zap.Int("after", len(allArticles)))

				// 如果有文章被过滤掉，显示过滤后的文章列表
				if beforeFilter != len(allArticles) {
					logger.Log.Info("关键词过滤后的文章列表", zap.String("account", accountName))
					for i, article := range allArticles {
						logger.Log.Info("keyword filtered article", zap.Int("index", i+1), zap.String("title", article.Title))
					}
				}
			}

			// 发送完成状态
			if statusChan != nil {
				statusChan <- models.AccountStatus{
					AccountName:  accountName,
					Status:       "completed",
					Message:      "爬取完成",
					ArticleCount: len(allArticles),
				}
			}

			logger.Log.Info("公众号爬取完成", zap.String("account", accountName), zap.Int("count", len(allArticles)))

			resultsChan <- allArticles
			return nil
		})
	}

	// 等待所有任务完成
	go func() {
		g.Wait()
		close(resultsChan)
		logger.Log.Info("所有公众号爬取任务已完成")
	}()

	// 收集结果
	logger.Log.Info("开始收集所有公众号的文章结果...")
	var allResults []models.Article
	accountCount := 0
	for articles := range resultsChan {
		accountCount++
		allResults = append(allResults, articles...)
		logger.Log.Info("收集公众号文章", zap.Int("accountIndex", accountCount), zap.Int("thisCount", len(articles)), zap.Int("total", len(allResults)))
	}

	logger.Log.Info("所有文章收集完成", zap.Int("accounts", accountCount), zap.Int("articles", len(allResults)))

	// 检查错误
	if err := g.Wait(); err != nil {
		// 如果是取消错误，不记录为错误
		if err == context.Canceled {
			logger.Log.Warn("爬取被用户取消")
			return allResults, err
		}
		logger.Log.Error("爬取过程中出现错误", zap.Error(err))
		return allResults, err
	}

	return allResults, nil
}

// GetProgress 获取进度
func (as *AsyncScraper) GetProgress() models.Progress {
	as.mu.Lock()
	defer as.mu.Unlock()
	return as.progress
}
