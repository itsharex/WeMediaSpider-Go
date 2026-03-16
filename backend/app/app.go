package app

import (
	"context"
	_ "embed"
	"os"
	"path/filepath"
	"time"

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
)

//go:embed icon.ico
var embeddedIcon []byte

// App 应用结构
type App struct {
	ctx                 context.Context
	loginManager        *spider.LoginManager
	scraper             *spider.AsyncScraper
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
	// 初始化日志
	logger.Init()

	// 创建缓存管理器
	cacheManager, err := cache.NewManager(96) // 96小时过期
	if err != nil {
		logger.Errorf("Failed to create cache manager: %v", err)
	}

	// 创建自启动管理器
	autostartManager, err := autostart.NewManager()
	if err != nil {
		logger.Errorf("Failed to create autostart manager: %v", err)
	}

	// 创建系统配置管理器
	systemConfigManager, err := config.NewSystemConfigManager()
	if err != nil {
		logger.Errorf("Failed to create system config manager: %v", err)
	}

	// 加载系统配置
	systemConfig := config.SystemConfig{
		CloseToTray:       true,
		RememberChoice:    false,
		UpdateIgnoredDate: "",
	}
	if systemConfigManager != nil {
		loadedConfig, err := systemConfigManager.Load()
		if err == nil {
			systemConfig = loadedConfig
			logger.Infof("[NewApp] Loaded system config: closeToTray=%v, rememberChoice=%v, updateIgnoredDate=%s",
				systemConfig.CloseToTray, systemConfig.RememberChoice, systemConfig.UpdateIgnoredDate)
		} else {
			logger.Warnf("[NewApp] Failed to load system config: %v", err)
		}
	} else {
		logger.Warnf("[NewApp] systemConfigManager is nil")
	}

	// 初始化数据库
	homeDir, err := os.UserHomeDir()
	if err != nil {
		logger.Errorf("Failed to get home directory: %v", err)
		homeDir = "."
	}
	configDir := filepath.Join(homeDir, ".wemediaspider")

	db, err := database.NewDatabase(configDir)
	if err != nil {
		logger.Errorf("Failed to initialize database: %v", err)
	} else {
		// 自动迁移表结构
		if err := db.AutoMigrate(); err != nil {
			logger.Errorf("Failed to migrate database: %v", err)
		}
	}

	// 初始化仓储
	var articleRepo repository.ArticleRepository
	var accountRepo repository.AccountRepository
	var statsRepo repository.StatsRepository
	var taskRepo repository.TaskRepository
	var analyticsRepo repository.AnalyticsRepository
	if db != nil {
		articleRepo = repository.NewArticleRepository(db.DB)
		accountRepo = repository.NewAccountRepository(db.DB)
		statsRepo = repository.NewStatsRepository(db.DB)
		taskRepo = repository.NewTaskRepository(db.DB)
		analyticsRepo = repository.NewAnalyticsRepository(db.DB)
	}

	// 创建登录管理器
	loginManager := spider.NewLoginManager()

	// 初始化调度器组件
	var taskScheduler *scheduler.TaskScheduler
	var cronManager *scheduler.CronManager
	if taskRepo != nil {
		taskScheduler = scheduler.NewTaskScheduler(taskRepo, articleRepo, accountRepo, loginManager)
		cronManager = scheduler.NewCronManager(taskScheduler)
	}

	// 初始化分析器
	var analyzer *analytics.Analyzer
	if analyticsRepo != nil {
		analyzer = analytics.NewAnalyzer(analyticsRepo, articleRepo, accountRepo)
	}

	app := &App{
		loginManager:        loginManager,
		configManager:       config.NewManager(),
		systemConfigManager: systemConfigManager,
		cacheManager:        cacheManager,
		db:                  db,
		articleRepo:         articleRepo,
		accountRepo:         accountRepo,
		statsRepo:           statsRepo,
		taskRepo:            taskRepo,
		analyticsRepo:       analyticsRepo,
		analyzer:            analyzer,
		cronManager:         cronManager,
		taskScheduler:       taskScheduler,
		trayManager:         tray.NewManager(),
		autostartManager:    autostartManager,
		closeToTray:         systemConfig.CloseToTray,
		rememberChoice:      systemConfig.RememberChoice,
		updateIgnoredDate:   systemConfig.UpdateIgnoredDate,
	}

	logger.Infof("[NewApp] App created with updateIgnoredDate=%s", app.updateIgnoredDate)
	return app
}

type VersionInfo struct {
	CurrentVersion string `json:"currentVersion"`
	LatestVersion  string `json:"latestVersion"`
	HasUpdate      bool   `json:"hasUpdate"`
	UpdateURL      string `json:"updateUrl"`
	ReleaseNotes   string `json:"releaseNotes"`
}

// updateCache 更新检查缓存
type updateCache struct {
	Version      string    `json:"version"`
	UpdateURL    string    `json:"updateUrl"`
	ReleaseNotes string    `json:"releaseNotes"`
	CheckedAt    time.Time `json:"checkedAt"`
}

