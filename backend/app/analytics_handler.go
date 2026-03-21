package app

import (
	"fmt"
	"time"

	"WeMediaSpider/backend/internal/models"
	"WeMediaSpider/backend/pkg/logger"

	"go.uber.org/zap"
)

func (a *App) GetAnalyticsData(startDate, endDate string, accountNames []string, forceRefresh bool) (*models.AnalyticsData, error) {
	if a.analyzer == nil {
		return nil, fmt.Errorf("analyzer not initialized")
	}

	// 解析日期
	start, err := time.Parse("2006-01-02", startDate)
	if err != nil {
		return nil, fmt.Errorf("invalid start date: %w", err)
	}

	end, err := time.Parse("2006-01-02", endDate)
	if err != nil {
		return nil, fmt.Errorf("invalid end date: %w", err)
	}

	// 将结束日期调整到当天的23:59:59，以包含整天的数据
	end = time.Date(end.Year(), end.Month(), end.Day(), 23, 59, 59, 0, end.Location())

	logger.Log.Info("分析数据日期范围",
		zap.String("start", startDate+" 00:00:00"),
		zap.String("end", endDate+" 23:59:59"),
		zap.Any("accounts", accountNames),
	)

	// 获取分析数据
	data, err := a.analyzer.GetAnalyticsData(start, end, accountNames, forceRefresh)
	if err != nil {
		return nil, fmt.Errorf("failed to get analytics data: %w", err)
	}

	return data, nil
}

// GetAllAccountNames 获取所有公众号名称列表
func (a *App) GetAllAccountNames() ([]string, error) {
	accounts, err := a.accountRepo.List()
	if err != nil {
		return nil, fmt.Errorf("failed to get accounts: %w", err)
	}

	names := make([]string, len(accounts))
	for i, acc := range accounts {
		names[i] = acc.Name
	}

	return names, nil
}

// ClearAnalyticsCache 清除分析缓存
func (a *App) ClearAnalyticsCache() error {
	if a.analyzer == nil {
		return fmt.Errorf("analyzer not initialized")
	}

	a.analyzer.ClearCache()
	logger.Log.Info("分析缓存已清除")
	return nil
}
