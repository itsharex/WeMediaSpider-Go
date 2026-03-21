package app

import (
	"os"
	"path/filepath"

	"WeMediaSpider/backend/internal/analytics"
	"WeMediaSpider/backend/internal/autostart"
	"WeMediaSpider/backend/internal/cache"
	"WeMediaSpider/backend/internal/config"
	"WeMediaSpider/backend/internal/database"
	"WeMediaSpider/backend/internal/repository"
	"WeMediaSpider/backend/internal/scheduler"
	"WeMediaSpider/backend/internal/spider"
	"WeMediaSpider/backend/pkg/logger"

	"go.uber.org/zap"
)

type configResult struct {
	cacheManager        *cache.Manager
	autostartManager    *autostart.Manager
	systemConfigManager *config.SystemConfigManager
	systemConfig        config.SystemConfig
}

type dbResult struct {
	db            *database.Database
	articleRepo   repository.ArticleRepository
	accountRepo   repository.AccountRepository
	statsRepo     repository.StatsRepository
	taskRepo      repository.TaskRepository
	analyticsRepo repository.AnalyticsRepository
}

type serviceResult struct {
	loginManager  *spider.LoginManager
	taskScheduler *scheduler.TaskScheduler
	cronManager   *scheduler.CronManager
	analyzer      *analytics.Analyzer
}

// initConfig 初始化缓存、自启动和系统配置
func initConfig() configResult {
	cacheManager, err := cache.NewManager(96)
	if err != nil {
		logger.Log.Error("创建缓存管理器失败", zap.Error(err))
	}

	autostartManager, err := autostart.NewManager()
	if err != nil {
		logger.Log.Error("创建自启动管理器失败", zap.Error(err))
	}

	systemConfigManager, err := config.NewSystemConfigManager()
	if err != nil {
		logger.Log.Error("创建系统配置管理器失败", zap.Error(err))
	}

	systemConfig := config.SystemConfig{
		CloseToTray:       true,
		RememberChoice:    false,
		UpdateIgnoredDate: "",
	}
	if systemConfigManager != nil {
		if loaded, err := systemConfigManager.Load(); err == nil {
			systemConfig = loaded
			logger.Log.Info("已加载系统配置", zap.Bool("close_to_tray", systemConfig.CloseToTray), zap.Bool("remember_choice", systemConfig.RememberChoice), zap.String("update_ignored_date", systemConfig.UpdateIgnoredDate))
		} else {
			logger.Log.Warn("加载系统配置失败", zap.Error(err))
		}
	} else {
		logger.Log.Warn("系统配置管理器未初始化")
	}

	return configResult{
		cacheManager:        cacheManager,
		autostartManager:    autostartManager,
		systemConfigManager: systemConfigManager,
		systemConfig:        systemConfig,
	}
}

// initDatabase 初始化数据库连接和仓储层
func initDatabase() dbResult {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		logger.Log.Error("获取用户目录失败", zap.Error(err))
		homeDir = "."
	}
	configDir := filepath.Join(homeDir, ".wemediaspider")

	db, err := database.NewDatabase(configDir)
	if err != nil {
		logger.Log.Error("初始化数据库失败", zap.Error(err))
		return dbResult{}
	}
	if err := db.AutoMigrate(); err != nil {
		logger.Log.Error("数据库迁移失败", zap.Error(err))
	}

	return dbResult{
		db:            db,
		articleRepo:   repository.NewArticleRepository(db.DB),
		accountRepo:   repository.NewAccountRepository(db.DB),
		statsRepo:     repository.NewStatsRepository(db.DB),
		taskRepo:      repository.NewTaskRepository(db.DB),
		analyticsRepo: repository.NewAnalyticsRepository(db.DB),
	}
}

// initServices 初始化登录、调度器和分析器
func initServices(d dbResult) serviceResult {
	loginManager := spider.NewLoginManager()

	var taskScheduler *scheduler.TaskScheduler
	var cronManager *scheduler.CronManager
	if d.taskRepo != nil {
		taskScheduler = scheduler.NewTaskScheduler(d.taskRepo, d.articleRepo, d.accountRepo, loginManager)
		cronManager = scheduler.NewCronManager(taskScheduler)
	}

	var analyzer *analytics.Analyzer
	if d.analyticsRepo != nil {
		analyzer = analytics.NewAnalyzer(d.analyticsRepo, d.articleRepo, d.accountRepo)
	}

	return serviceResult{
		loginManager:  loginManager,
		taskScheduler: taskScheduler,
		cronManager:   cronManager,
		analyzer:      analyzer,
	}
}
