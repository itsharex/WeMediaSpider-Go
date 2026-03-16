package repository

import (
	"fmt"
	"regexp"
	"time"

	"WeMediaSpider/backend/internal/database/models"

	"gorm.io/gorm"
)

// AnalyticsRepository 分析数据访问接口
type AnalyticsRepository interface {
	// 时间分布统计（支持按日/周/月）
	GetTimeDistribution(startDate, endDate time.Time, granularity string) (map[string]int, error)

	// 按公众号分组的时间分布统计
	GetTimeDistributionByAccount(startDate, endDate time.Time, accountNames []string, granularity string) (map[string]map[string]int, error)

	// 获取文章内容用于关键词提取（支持按公众号筛选）
	GetArticleContents(startDate, endDate time.Time, accountNames []string, limit int) ([]string, error)

	// 文章长度统计
	GetLengthDistribution(startDate, endDate time.Time) (map[string]int, error)

	// 公众号活跃度统计
	GetAccountActivity(startDate, endDate time.Time) ([]*models.Account, map[uint]int, error)
}

type analyticsRepository struct {
	db *gorm.DB
}

// NewAnalyticsRepository 创建分析仓储实例
func NewAnalyticsRepository(db *gorm.DB) AnalyticsRepository {
	return &analyticsRepository{db: db}
}

// GetTimeDistribution 按日/周/月统计文章数量
func (r *analyticsRepository) GetTimeDistribution(startDate, endDate time.Time, granularity string) (map[string]int, error) {
	var results []struct {
		Date  string
		Count int
	}

	// SQLite 日期格式化
	var dateFormat string
	switch granularity {
	case "day":
		dateFormat = "%Y-%m-%d"
	case "week":
		dateFormat = "%Y-W%W"
	case "month":
		dateFormat = "%Y-%m"
	default:
		dateFormat = "%Y-%m-%d"
	}

	err := r.db.Model(&models.Article{}).
		Select("strftime(?, datetime(publish_timestamp, 'unixepoch')) as date, COUNT(*) as count", dateFormat).
		Where("publish_timestamp >= ? AND publish_timestamp <= ?", startDate.Unix(), endDate.Unix()).
		Group("date").
		Order("date ASC").
		Scan(&results).Error

	if err != nil {
		return nil, err
	}

	distribution := make(map[string]int)
	for _, r := range results {
		distribution[r.Date] = r.Count
	}

	return distribution, nil
}

// GetTimeDistributionByAccount 按公众号分组统计时间分布
func (r *analyticsRepository) GetTimeDistributionByAccount(startDate, endDate time.Time, accountNames []string, granularity string) (map[string]map[string]int, error) {
	var results []struct {
		AccountName string
		Date        string
		Count       int
	}

	// SQLite 日期格式化
	var dateFormat string
	switch granularity {
	case "day":
		dateFormat = "%Y-%m-%d"
	case "week":
		dateFormat = "%Y-W%W"
	case "month":
		dateFormat = "%Y-%m"
	default:
		dateFormat = "%Y-%m-%d"
	}

	// 打印查询参数
	println(fmt.Sprintf("查询参数: startDate=%d (%s), endDate=%d (%s), accountNames=%v",
		startDate.Unix(), startDate.Format("2006-01-02 15:04:05"),
		endDate.Unix(), endDate.Format("2006-01-02 15:04:05"),
		accountNames))

	query := r.db.Model(&models.Article{}).
		Select("account_name, strftime(?, datetime(publish_timestamp, 'unixepoch')) as date, COUNT(*) as count", dateFormat).
		Where("publish_timestamp >= ? AND publish_timestamp <= ?", startDate.Unix(), endDate.Unix())

	// 如果指定了公众号，则过滤
	if len(accountNames) > 0 {
		query = query.Where("account_name IN ?", accountNames)
	}

	err := query.
		Group("account_name, date").
		Order("account_name ASC, date ASC").
		Scan(&results).Error

	if err != nil {
		return nil, err
	}

	// 打印查询结果
	println(fmt.Sprintf("查询到 %d 条记录", len(results)))
	for _, r := range results {
		println(fmt.Sprintf("  - %s: %s = %d", r.AccountName, r.Date, r.Count))
	}

	// 按公众号分组
	distribution := make(map[string]map[string]int)
	accountSet := make(map[string]bool) // 用于统计不同的公众号
	for _, r := range results {
		if distribution[r.AccountName] == nil {
			distribution[r.AccountName] = make(map[string]int)
			accountSet[r.AccountName] = true
		}
		distribution[r.AccountName][r.Date] = r.Count
	}

	// 打印调试信息
	println(fmt.Sprintf("GetTimeDistributionByAccount: 查询到 %d 个公众号的数据", len(accountSet)))
	for accountName := range accountSet {
		println(fmt.Sprintf("  - %s: %d 个日期", accountName, len(distribution[accountName])))
	}

	return distribution, nil
}

// GetArticleContents 获取文章内容用于关键词提取
func (r *analyticsRepository) GetArticleContents(startDate, endDate time.Time, accountNames []string, limit int) ([]string, error) {
	var articles []*models.Article

	query := r.db.Select("title, digest, content").
		Where("publish_timestamp >= ? AND publish_timestamp <= ?", startDate.Unix(), endDate.Unix())

	// 如果指定了公众号，添加筛选条件
	if len(accountNames) > 0 {
		query = query.Where("account_name IN ?", accountNames)
	}

	query = query.Order("publish_timestamp DESC")

	if limit > 0 {
		query = query.Limit(limit)
	}

	err := query.Find(&articles).Error
	if err != nil {
		return nil, err
	}

	contents := make([]string, 0, len(articles))
	for _, art := range articles {
		// 清理 Markdown 图片链接后再合并标题、摘要、内容
		cleanedContent := cleanMarkdownImages(art.Content)
		text := art.Title + " " + art.Digest + " " + cleanedContent
		contents = append(contents, text)
	}

	return contents, nil
}

// cleanMarkdownImages 清理 Markdown 图片链接
func cleanMarkdownImages(content string) string {
	// 匹配 Markdown 图片语法: ![alt](url) 或 ![](url)
	imagePattern := regexp.MustCompile(`!\[.*?\]\(.*?\)`)
	return imagePattern.ReplaceAllString(content, "")
}

// GetLengthDistribution 统计文章长度分布
func (r *analyticsRepository) GetLengthDistribution(startDate, endDate time.Time) (map[string]int, error) {
	var articles []*models.Article

	err := r.db.Select("content").
		Where("publish_timestamp >= ? AND publish_timestamp <= ?", startDate.Unix(), endDate.Unix()).
		Find(&articles).Error

	if err != nil {
		return nil, err
	}

	// 定义长度区间
	ranges := map[string][2]int{
		"0-500":     {0, 500},
		"500-1000":  {500, 1000},
		"1000-2000": {1000, 2000},
		"2000-5000": {2000, 5000},
		"5000+":     {5000, 999999999},
	}

	distribution := make(map[string]int)
	for rangeName := range ranges {
		distribution[rangeName] = 0
	}

	// 统计每个区间的文章数
	for _, art := range articles {
		// 清理 Markdown 图片链接后再计算长度
		cleanedContent := cleanMarkdownImages(art.Content)
		length := len(cleanedContent)
		for rangeName, bounds := range ranges {
			if length >= bounds[0] && length < bounds[1] {
				distribution[rangeName]++
				break
			}
		}
	}

	return distribution, nil
}

// GetAccountActivity 获取公众号活跃度数据
func (r *analyticsRepository) GetAccountActivity(startDate, endDate time.Time) ([]*models.Account, map[uint]int, error) {
	// 获取所有公众号
	var accounts []*models.Account
	err := r.db.Find(&accounts).Error
	if err != nil {
		return nil, nil, err
	}

	// 统计每个公众号的文章数
	articleCounts := make(map[uint]int)
	for _, acc := range accounts {
		var count int64
		r.db.Model(&models.Article{}).
			Where("account_id = ? AND publish_timestamp >= ? AND publish_timestamp <= ?",
				acc.ID, startDate.Unix(), endDate.Unix()).
			Count(&count)
		articleCounts[acc.ID] = int(count)
	}

	return accounts, articleCounts, nil
}
