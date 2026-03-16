package models

// CronValidationResult Cron 表达式验证结果
type CronValidationResult struct {
	Valid    bool   `json:"valid"`    // 是否有效
	NextTime string `json:"nextTime"` // 下次运行时间
	Error    string `json:"error"`    // 错误信息
}

// AccountTimeDistribution 公众号时间分布数据
type AccountTimeDistribution struct {
	AccountName string         `json:"accountName"` // 公众号名称
	Data        []TimeDistribution `json:"data"`    // 时间分布数据
}

// TimeDistribution 时间分布数据
type TimeDistribution struct {
	Date  string `json:"date"`  // 日期 YYYY-MM-DD
	Count int    `json:"count"` // 文章数量
}

// KeywordFrequency 关键词频率
type KeywordFrequency struct {
	Word  string `json:"word"`  // 关键词
	Count int    `json:"count"` // 出现次数
}

// LengthDistribution 文章长度分布
type LengthDistribution struct {
	Range string `json:"range"` // 长度区间 "0-500"
	Count int    `json:"count"` // 文章数量
}

// AccountActivity 公众号活跃度
type AccountActivity struct {
	AccountName   string  `json:"accountName"`   // 公众号名称
	ArticleCount  int     `json:"articleCount"`  // 文章总数
	AvgLength     int     `json:"avgLength"`     // 平均文章长度
	PublishFreq   float64 `json:"publishFreq"`   // 发文频率（篇/天）
	LastPublishAt string  `json:"lastPublishAt"` // 最后发文时间
	ActivityScore float64 `json:"activityScore"` // 活跃度评分
}

// AnalyticsData 分析数据汇总
type AnalyticsData struct {
	TimeDistribution   []AccountTimeDistribution `json:"timeDistribution"`   // 按公众号分组的时间分布
	TopKeywords        []KeywordFrequency        `json:"topKeywords"`        // 热门关键词
	LengthDistribution []LengthDistribution      `json:"lengthDistribution"` // 长度分布
	AccountRanking     []AccountActivity         `json:"accountRanking"`     // 公众号排行
	CachedAt           string                    `json:"cachedAt"`           // 缓存时间
}
