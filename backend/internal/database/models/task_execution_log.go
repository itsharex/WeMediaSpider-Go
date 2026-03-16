package models

import "time"

// TaskExecutionLog 任务执行日志
type TaskExecutionLog struct {
	ID            uint       `gorm:"primaryKey"`
	TaskID        uint       `gorm:"not null;index" json:"taskId"`
	TaskName      string     `gorm:"size:255;not null;index" json:"taskName"`
	StartTime     time.Time  `gorm:"not null;index" json:"startTime"`
	EndTime       *time.Time `json:"endTime"`
	Duration      int64      `gorm:"default:0" json:"duration"` // 执行时长（毫秒）
	Status        string     `gorm:"size:32;not null;index" json:"status"` // success/failed/running/canceled
	ArticlesCount int        `gorm:"default:0" json:"articlesCount"`
	ErrorMessage  string     `gorm:"type:text" json:"errorMessage"`
	TriggerType   string     `gorm:"size:32;not null" json:"triggerType"` // cron/manual
	CreatedAt     time.Time  `json:"createdAt"`
	Task          ScheduledTask `gorm:"foreignKey:TaskID" json:"-"`
}

// TableName 指定表名
func (TaskExecutionLog) TableName() string {
	return "task_execution_logs"
}
