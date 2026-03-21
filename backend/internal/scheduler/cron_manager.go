package scheduler

import (
	"context"
	"fmt"
	"sync"
	"time"

	"WeMediaSpider/backend/internal/database/models"
	"WeMediaSpider/backend/pkg/logger"
	"WeMediaSpider/backend/pkg/timeutil"

	"go.uber.org/zap"
	"github.com/robfig/cron/v3"
)

// CronManager 定时任务管理器
type CronManager struct {
	cron      *cron.Cron
	scheduler *TaskScheduler
	tasks     map[uint]cron.EntryID // taskID -> cronEntryID
	mu        sync.RWMutex
	ctx       context.Context
	cancel    context.CancelFunc
}

// NewCronManager 创建定时任务管理器
func NewCronManager(scheduler *TaskScheduler) *CronManager {
	ctx, cancel := context.WithCancel(context.Background())

	return &CronManager{
		cron:      cron.New(cron.WithSeconds()), // 支持秒级精度
		scheduler: scheduler,
		tasks:     make(map[uint]cron.EntryID),
		ctx:       ctx,
		cancel:    cancel,
	}
}

// Start 启动定时任务管理器
func (cm *CronManager) Start() {
	cm.cron.Start()
	logger.Log.Info("CronManager started")
}

// Stop 停止定时任务管理器
func (cm *CronManager) Stop() {
	cm.cancel()
	ctx := cm.cron.Stop()
	<-ctx.Done()
	logger.Log.Info("CronManager stopped")
}

// AddTask 添加定时任务
func (cm *CronManager) AddTask(task *models.ScheduledTask) error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	// 解析 Cron 表达式
	entryID, err := cm.cron.AddFunc(task.CronExpression, func() {
		cm.scheduler.ExecuteTask(cm.ctx, task.ID, "cron")
	})

	if err != nil {
		return fmt.Errorf("invalid cron expression: %w", err)
	}

	cm.tasks[task.ID] = entryID
	logger.Log.Info("Added task", zap.Uint("taskID", task.ID), zap.String("cron", task.CronExpression))

	return nil
}

// RemoveTask 移除定时任务
func (cm *CronManager) RemoveTask(taskID uint) {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	if entryID, exists := cm.tasks[taskID]; exists {
		cm.cron.Remove(entryID)
		delete(cm.tasks, taskID)
		logger.Log.Info("Removed task", zap.Uint("taskID", taskID))
	}
}

// UpdateTask 更新定时任务
func (cm *CronManager) UpdateTask(task *models.ScheduledTask) error {
	cm.RemoveTask(task.ID)
	if task.Enabled {
		return cm.AddTask(task)
	}
	return nil
}

// GetNextRunTime 获取下次运行时间
func (cm *CronManager) GetNextRunTime(taskID uint) *time.Time {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	if entryID, exists := cm.tasks[taskID]; exists {
		entry := cm.cron.Entry(entryID)
		nextTime := entry.Next
		return &nextTime
	}
	return nil
}

// ParseCronExpression 解析 Cron 表达式并返回下次运行时间
func (cm *CronManager) ParseCronExpression(expression string) (time.Time, error) {
	parser := cron.NewParser(cron.Second | cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow | cron.Descriptor)
	schedule, err := parser.Parse(expression)
	if err != nil {
		return time.Time{}, fmt.Errorf("invalid cron expression: %w", err)
	}

	nextTime := schedule.Next(timeutil.Now())
	return nextTime, nil
}
