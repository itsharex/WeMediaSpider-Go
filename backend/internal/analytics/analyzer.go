package analytics

import (
	"fmt"
	"sort"
	"strings"
	"time"

	dbmodels "WeMediaSpider/backend/internal/database/models"
	appmodels "WeMediaSpider/backend/internal/models"
	"WeMediaSpider/backend/internal/repository"
	"WeMediaSpider/backend/pkg/logger"
	"WeMediaSpider/backend/pkg/timeutil"
)

// Analyzer 数据分析器
type Analyzer struct {
	analyticsRepo repository.AnalyticsRepository
	articleRepo   repository.ArticleRepository
	accountRepo   repository.AccountRepository
	keywordExt    *KeywordExtractor
	cache         *AnalyticsCache
}

// NewAnalyzer 创建分析器实例
func NewAnalyzer(
	analyticsRepo repository.AnalyticsRepository,
	articleRepo repository.ArticleRepository,
	accountRepo repository.AccountRepository,
) *Analyzer {
	return &Analyzer{
		analyticsRepo: analyticsRepo,
		articleRepo:   articleRepo,
		accountRepo:   accountRepo,
		keywordExt:    NewKeywordExtractor(),
		cache:         NewAnalyticsCache(30 * time.Minute), // 缓存30分钟
	}
}

// Close 关闭分析器
func (a *Analyzer) Close() {
	a.keywordExt.Close()
}

// GetAnalyticsData 获取完整分析数据
func (a *Analyzer) GetAnalyticsData(startDate, endDate time.Time, accountNames []string, forceRefresh bool) (*appmodels.AnalyticsData, error) {
	// 生成更可靠的缓存 key
	accountKey := "all"
	if len(accountNames) > 0 {
		accountKey = strings.Join(accountNames, ",")
	}
	cacheKey := fmt.Sprintf("%s_%s_%s", startDate.Format("2006-01-02"), endDate.Format("2006-01-02"), accountKey)

	logger.Infof("分析数据缓存 key: %s, accountNames: %v", cacheKey, accountNames)

	// 检查缓存
	if !forceRefresh {
		if cached := a.cache.Get(cacheKey); cached != nil {
			logger.Infof("返回缓存的分析数据，公众号数量: %d", len(cached.TimeDistribution))
			return cached, nil
		}
	}

	logger.Info("开始计算分析数据...")

	// 1. 按公众号分组的时间分布（支持按公众号筛选）
	timeDistribution, err := a.GetTimeDistributionByAccount(startDate, endDate, accountNames, "day")
	if err != nil {
		return nil, fmt.Errorf("获取时间分布失败: %w", err)
	}

	logger.Infof("获取到时间分布数据，公众号数量: %d", len(timeDistribution))

	// 2. 热门关键词（支持按公众号筛选）
	topKeywords, err := a.GetTopKeywords(startDate, endDate, accountNames, 50)
	if err != nil {
		return nil, fmt.Errorf("提取关键词失败: %w", err)
	}

	// 3. 文章长度分布
	lengthDistribution, err := a.GetLengthDistribution(startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("获取长度分布失败: %w", err)
	}

	// 4. 公众号活跃度排行
	accountRanking, err := a.GetAccountRanking(startDate, endDate, 20)
	if err != nil {
		return nil, fmt.Errorf("获取活跃度排行失败: %w", err)
	}

	data := &appmodels.AnalyticsData{
		TimeDistribution:   timeDistribution,
		TopKeywords:        topKeywords,
		LengthDistribution: lengthDistribution,
		AccountRanking:     accountRanking,
		CachedAt:           timeutil.Now().Format("2006-01-02 15:04:05"),
	}

	// 保存到缓存
	a.cache.Set(cacheKey, data)

	logger.Info("分析数据计算完成")
	return data, nil
}

// GetTimeDistribution 获取时间分布
func (a *Analyzer) GetTimeDistribution(startDate, endDate time.Time, granularity string) ([]appmodels.TimeDistribution, error) {
	distMap, err := a.analyticsRepo.GetTimeDistribution(startDate, endDate, granularity)
	if err != nil {
		return nil, err
	}

	result := make([]appmodels.TimeDistribution, 0, len(distMap))
	for date, count := range distMap {
		result = append(result, appmodels.TimeDistribution{
			Date:  date,
			Count: count,
		})
	}

	// 按日期排序
	sort.Slice(result, func(i, j int) bool {
		return result[i].Date < result[j].Date
	})

	return result, nil
}

// GetTimeDistributionByAccount 获取按公众号分组的时间分布
func (a *Analyzer) GetTimeDistributionByAccount(startDate, endDate time.Time, accountNames []string, granularity string) ([]appmodels.AccountTimeDistribution, error) {
	distMap, err := a.analyticsRepo.GetTimeDistributionByAccount(startDate, endDate, accountNames, granularity)
	if err != nil {
		return nil, err
	}

	result := make([]appmodels.AccountTimeDistribution, 0, len(distMap))
	for accountName, dateMap := range distMap {
		timeData := make([]appmodels.TimeDistribution, 0, len(dateMap))
		for date, count := range dateMap {
			timeData = append(timeData, appmodels.TimeDistribution{
				Date:  date,
				Count: count,
			})
		}

		// 按日期排序
		sort.Slice(timeData, func(i, j int) bool {
			return timeData[i].Date < timeData[j].Date
		})

		result = append(result, appmodels.AccountTimeDistribution{
			AccountName: accountName,
			Data:        timeData,
		})
	}

	// 按公众号名称排序
	sort.Slice(result, func(i, j int) bool {
		return result[i].AccountName < result[j].AccountName
	})

	return result, nil
}

// GetTopKeywords 获取热门关键词
func (a *Analyzer) GetTopKeywords(startDate, endDate time.Time, accountNames []string, topN int) ([]appmodels.KeywordFrequency, error) {
	// 获取文章内容（限制数量避免内存溢出）
	contents, err := a.analyticsRepo.GetArticleContents(startDate, endDate, accountNames, 1000)
	if err != nil {
		return nil, err
	}

	if len(contents) == 0 {
		return []appmodels.KeywordFrequency{}, nil
	}

	// 提取关键词
	keywords := a.keywordExt.ExtractTopKeywords(contents, topN)

	result := make([]appmodels.KeywordFrequency, len(keywords))
	for i, kw := range keywords {
		result[i] = appmodels.KeywordFrequency{
			Word:  kw.Word,
			Count: kw.Count,
		}
	}

	return result, nil
}

// GetLengthDistribution 获取文章长度分布
func (a *Analyzer) GetLengthDistribution(startDate, endDate time.Time) ([]appmodels.LengthDistribution, error) {
	distMap, err := a.analyticsRepo.GetLengthDistribution(startDate, endDate)
	if err != nil {
		return nil, err
	}

	// 定义区间顺序
	rangeOrder := []string{"0-500", "500-1000", "1000-2000", "2000-5000", "5000+"}

	result := make([]appmodels.LengthDistribution, 0, len(rangeOrder))
	for _, r := range rangeOrder {
		result = append(result, appmodels.LengthDistribution{
			Range: r,
			Count: distMap[r],
		})
	}

	return result, nil
}

// GetAccountRanking 获取公众号活跃度排行
func (a *Analyzer) GetAccountRanking(startDate, endDate time.Time, topN int) ([]appmodels.AccountActivity, error) {
	accounts, articleCounts, err := a.analyticsRepo.GetAccountActivity(startDate, endDate)
	if err != nil {
		return nil, err
	}

	daysDiff := endDate.Sub(startDate).Hours() / 24
	if daysDiff == 0 {
		daysDiff = 1
	}

	activities := make([]appmodels.AccountActivity, 0, len(accounts))

	for _, acc := range accounts {
		count := articleCounts[acc.ID]
		if count == 0 {
			continue
		}

		// 计算平均文章长度
		var avgLength int
		articles, err := a.articleRepo.FindByAccountFakeid(acc.Fakeid, 0, 0)
		if err == nil && len(articles) > 0 {
			totalLength := 0
			for _, art := range articles {
				totalLength += len(art.Content)
			}
			avgLength = totalLength / len(articles)
		}

		// 计算发文频率
		publishFreq := float64(count) / daysDiff

		// 计算活跃度评分（综合考虑发文频率和文章质量）
		activityScore := publishFreq*10 + float64(avgLength)/1000

		activities = append(activities, appmodels.AccountActivity{
			AccountName:   acc.Name,
			ArticleCount:  count,
			AvgLength:     avgLength,
			PublishFreq:   publishFreq,
			LastPublishAt: acc.UpdatedAt.Format("2006-01-02"),
			ActivityScore: activityScore,
		})
	}

	// 按活跃度评分排序
	sort.Slice(activities, func(i, j int) bool {
		return activities[i].ActivityScore > activities[j].ActivityScore
	})

	// 返回 Top N
	if len(activities) > topN {
		activities = activities[:topN]
	}

	return activities, nil
}

// ClearCache 清除缓存
func (a *Analyzer) ClearCache() {
	a.cache.Clear()
}

// GetCache 获取缓存实例（用于外部访问）
func (a *Analyzer) GetCache() *AnalyticsCache {
	return a.cache
}

// 避免未使用的导入警告
var _ = dbmodels.Account{}
