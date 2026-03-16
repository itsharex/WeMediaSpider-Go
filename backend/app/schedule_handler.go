package app

import (
	"context"
	"fmt"

	dbmodels "WeMediaSpider/backend/internal/database/models"
	"WeMediaSpider/backend/internal/models"
	"WeMediaSpider/backend/pkg/logger"
)

func (a *App) loadScheduledTasks() {
	if a.taskRepo == nil || a.cronManager == nil {
		return
	}

	// 清理应用重启时正在运行的任务状态
	if err := a.taskRepo.CleanupRunningLogs(); err != nil {
		logger.Errorf("Failed to cleanup running logs: %v", err)
	}

	// 加载所有已启用的任务
	enabled := true
	tasks, err := a.taskRepo.List(&enabled)
	if err != nil {
		logger.Errorf("Failed to load scheduled tasks: %v", err)
		return
	}

	// 添加到 CronManager
	for _, task := range tasks {
		if err := a.cronManager.AddTask(task); err != nil {
			logger.Errorf("Failed to add task %d to cron: %v", task.ID, err)
		}
	}

	logger.Infof("Loaded %d enabled scheduled tasks", len(tasks))
}

// CreateScheduledTask 创建定时任务
func (a *App) CreateScheduledTask(task dbmodels.ScheduledTask) error {
	if a.taskRepo == nil {
		return fmt.Errorf("task repository not initialized")
	}

	// 验证 Cron 表达式
	// 验证 Cron 表达式
	validationResult, err := a.ValidateCronExpression(task.CronExpression)
	if err != nil || !validationResult.Valid {
		if err != nil {
			return fmt.Errorf("failed to validate cron expression: %w", err)
		}
		return fmt.Errorf("invalid cron expression: %s", validationResult.Error)
	}

	// 创建任务
	if err := a.taskRepo.Create(&task); err != nil {
		return fmt.Errorf("failed to create task: %w", err)
	}

	// 如果任务已启用，添加到 CronManager
	if task.Enabled && a.cronManager != nil {
		if err := a.cronManager.AddTask(&task); err != nil {
			logger.Errorf("Failed to add task to cron: %v", err)
		}
	}

	logger.Infof("Created scheduled task: %s (ID: %d)", task.Name, task.ID)
	return nil
}

// UpdateScheduledTask 更新定时任务
func (a *App) UpdateScheduledTask(task dbmodels.ScheduledTask) error {
	if a.taskRepo == nil {
		return fmt.Errorf("task repository not initialized")
	}

	// 验证 Cron 表达式
	// 验证 Cron 表达式
	validationResult, err := a.ValidateCronExpression(task.CronExpression)
	if err != nil || !validationResult.Valid {
		if err != nil {
			return fmt.Errorf("failed to validate cron expression: %w", err)
		}
		return fmt.Errorf("invalid cron expression: %s", validationResult.Error)
	}

	// 更新任务
	if err := a.taskRepo.Update(&task); err != nil {
		return fmt.Errorf("failed to update task: %w", err)
	}

	// 更新 CronManager
	if a.cronManager != nil {
		if err := a.cronManager.UpdateTask(&task); err != nil {
			logger.Errorf("Failed to update task in cron: %v", err)
		}
	}

	logger.Infof("Updated scheduled task: %s (ID: %d)", task.Name, task.ID)
	return nil
}

// DeleteScheduledTask 删除定时任务
func (a *App) DeleteScheduledTask(id uint) error {
	if a.taskRepo == nil {
		return fmt.Errorf("task repository not initialized")
	}

	// 从 CronManager 移除
	if a.cronManager != nil {
		a.cronManager.RemoveTask(id)
	}

	// 删除任务（会级联删除执行日志）
	if err := a.taskRepo.Delete(id); err != nil {
		return fmt.Errorf("failed to delete task: %w", err)
	}

	logger.Infof("Deleted scheduled task: %d", id)
	return nil
}

// GetScheduledTask 获取单个定时任务
func (a *App) GetScheduledTask(id uint) (*dbmodels.ScheduledTask, error) {
	if a.taskRepo == nil {
		return nil, fmt.Errorf("task repository not initialized")
	}

	task, err := a.taskRepo.FindByID(id)
	if err != nil {
		return nil, fmt.Errorf("failed to find task: %w", err)
	}

	// 获取下次运行时间
	if a.cronManager != nil {
		nextRunTime := a.cronManager.GetNextRunTime(id)
		task.NextRunTime = nextRunTime
	}

	return task, nil
}

// ListScheduledTasks 获取定时任务列表
func (a *App) ListScheduledTasks(enabledOnly bool) ([]*dbmodels.ScheduledTask, error) {
	if a.taskRepo == nil {
		return nil, fmt.Errorf("task repository not initialized")
	}

	var enabled *bool
	if enabledOnly {
		t := true
		enabled = &t
	}

	tasks, err := a.taskRepo.List(enabled)
	if err != nil {
		return nil, fmt.Errorf("failed to list tasks: %w", err)
	}

	// 填充下次运行时间
	if a.cronManager != nil {
		for _, task := range tasks {
			nextRunTime := a.cronManager.GetNextRunTime(task.ID)
			task.NextRunTime = nextRunTime
		}
	}

	return tasks, nil
}

// RunScheduledTaskNow 立即运行定时任务
func (a *App) RunScheduledTaskNow(id uint) error {
	if a.taskScheduler == nil {
		return fmt.Errorf("task scheduler not initialized")
	}

	// 在后台执行任务
	go a.taskScheduler.ExecuteTask(context.Background(), id, "manual")

	logger.Infof("Manually triggered task: %d", id)
	return nil
}

// CancelScheduledTask 取消正在运行的任务
func (a *App) CancelScheduledTask(id uint) error {
	if a.taskScheduler == nil {
		return fmt.Errorf("task scheduler not initialized")
	}

	if err := a.taskScheduler.CancelTask(id); err != nil {
		return fmt.Errorf("failed to cancel task: %w", err)
	}

	logger.Infof("Canceled task: %d", id)
	return nil
}

// GetTaskExecutionLogs 获取任务执行日志
func (a *App) GetTaskExecutionLogs(taskID uint, limit int) ([]*dbmodels.TaskExecutionLog, error) {
	if a.taskRepo == nil {
		return nil, fmt.Errorf("task repository not initialized")
	}

	if limit <= 0 {
		limit = 50
	}

	logs, err := a.taskRepo.GetExecutionLogs(taskID, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get execution logs: %w", err)
	}

	return logs, nil
}

// GetRecentExecutionLogs 获取最近的执行日志
func (a *App) GetRecentExecutionLogs(limit int) ([]*dbmodels.TaskExecutionLog, error) {
	if a.taskRepo == nil {
		return nil, fmt.Errorf("task repository not initialized")
	}

	if limit <= 0 {
		limit = 20
	}

	logs, err := a.taskRepo.GetRecentExecutionLogs(limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get recent logs: %w", err)
	}

	return logs, nil
}

// ValidateCronExpression 验证 Cron 表达式
func (a *App) ValidateCronExpression(expression string) (*models.CronValidationResult, error) {
	if a.cronManager == nil {
		return &models.CronValidationResult{
			Valid: false,
			Error: "cron manager not initialized",
		}, nil
	}

	// 尝试解析 Cron 表达式
	nextTime, err := a.cronManager.ParseCronExpression(expression)
	if err != nil {
		return &models.CronValidationResult{
			Valid: false,
			Error: err.Error(),
		}, nil
	}

	// 返回下次运行时间
	nextTimeStr := nextTime.Format("2006-01-02 15:04:05")
	return &models.CronValidationResult{
		Valid:    true,
		NextTime: nextTimeStr,
		Error:    "",
	}, nil
}

// ============================================================
// 数据分析 API
// ============================================================

