package app

import (
	"context"
	_ "embed"
	"sync"

	"WeMediaSpider/backend/internal/analytics"
	"WeMediaSpider/backend/internal/autostart"
	"WeMediaSpider/backend/internal/cache"
	"WeMediaSpider/backend/internal/config"
	"WeMediaSpider/backend/internal/database"
	"WeMediaSpider/backend/internal/repository"
	"WeMediaSpider/backend/internal/scheduler"
	"WeMediaSpider/backend/internal/spider"
	"WeMediaSpider/backend/internal/tray"
	"WeMediaSpider/backend/pkg/logger"

	"go.uber.org/zap"
)

//go:embed icon.ico
var embeddedIcon []byte

// App 应用结构
type App struct {
	ctx                 context.Context
	loginManager        *spider.LoginManager
	scraper             *spider.AsyncScraper
	scrapeMu            sync.Mutex // 保护 scraper 和 imageDownloader 的并发访问
	configManager       *config.Manager
	systemConfigManager *config.SystemConfigManager
	cacheManager        *cache.Manager
	imageDownloader     *spider.ImageDownloader
	db                  *database.Database
	articleRepo         repository.ArticleRepository
	accountRepo         repository.AccountRepository
	statsRepo           repository.StatsRepository
	taskRepo            repository.TaskRepository
	analyticsRepo       repository.AnalyticsRepository
	analyzer            *analytics.Analyzer
	cronManager         *scheduler.CronManager
	taskScheduler       *scheduler.TaskScheduler
	trayManager         *tray.Manager
	autostartManager    *autostart.Manager
	closeToTray         bool   // 关闭到托盘
	rememberChoice      bool   // 记住用户选择
	updateIgnoredDate   string // 更新忽略日期
	forceQuit           bool   // 强制退出标志
}

// NewApp 创建应用实例
func NewApp() *App {
	_ = logger.Init()

	cfg := initConfig()
	db := initDatabase()
	svc := initServices(db)

	app := &App{
		loginManager:        svc.loginManager,
		configManager:       config.NewManager(),
		systemConfigManager: cfg.systemConfigManager,
		cacheManager:        cfg.cacheManager,
		db:                  db.db,
		articleRepo:         db.articleRepo,
		accountRepo:         db.accountRepo,
		statsRepo:           db.statsRepo,
		taskRepo:            db.taskRepo,
		analyticsRepo:       db.analyticsRepo,
		analyzer:            svc.analyzer,
		cronManager:         svc.cronManager,
		taskScheduler:       svc.taskScheduler,
		trayManager:         tray.NewManager(),
		autostartManager:    cfg.autostartManager,
		closeToTray:         cfg.systemConfig.CloseToTray,
		rememberChoice:      cfg.systemConfig.RememberChoice,
		updateIgnoredDate:   cfg.systemConfig.UpdateIgnoredDate,
	}

	logger.Log.Info("应用实例已创建", zap.String("update_ignored_date", app.updateIgnoredDate))
	return app
}

