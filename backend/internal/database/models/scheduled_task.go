package models

import "time"

// ScheduledTask 定时任务
type ScheduledTask struct {
	ID             uint       `gorm:"primaryKey"`
	Name           string     `gorm:"size:255;not null;index" json:"name"`
	Description    string     `gorm:"type:text" json:"description"`
	CronExpression string     `gorm:"size:128;not null" json:"cronExpression"`
	Enabled        bool       `gorm:"default:true;index" json:"enabled"`
	ScrapeConfig   string     `gorm:"type:text;not null" json:"scrapeConfig"` // JSON 格式的爬取配置
	LastRunTime    *time.Time `gorm:"index" json:"lastRunTime"`
	NextRunTime    *time.Time `gorm:"index" json:"nextRunTime"`
	LastRunStatus  string     `gorm:"size:32" json:"lastRunStatus"` // success/failed/running
	LastRunError   string     `gorm:"type:text" json:"lastRunError"`
	TotalRuns      int        `gorm:"default:0" json:"totalRuns"`
	SuccessRuns    int        `gorm:"default:0" json:"successRuns"`
	FailedRuns     int        `gorm:"default:0" json:"failedRuns"`
	CreatedAt      time.Time  `json:"createdAt"`
	UpdatedAt      time.Time  `json:"updatedAt"`
}

// TableName 指定表名
func (ScheduledTask) TableName() string {
	return "scheduled_tasks"
}
