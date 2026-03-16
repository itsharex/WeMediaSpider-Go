package repository

import (
	"WeMediaSpider/backend/internal/database/models"

	"gorm.io/gorm"
)

// TaskRepository 任务数据访问接口
type TaskRepository interface {
	// 任务管理
	Create(task *models.ScheduledTask) error
	Update(task *models.ScheduledTask) error
	Delete(id uint) error
	FindByID(id uint) (*models.ScheduledTask, error)
	List(enabled *bool) ([]*models.ScheduledTask, error)

	// 执行日志管理
	CreateExecutionLog(log *models.TaskExecutionLog) error
	UpdateExecutionLog(log *models.TaskExecutionLog) error
	GetExecutionLogs(taskID uint, limit int) ([]*models.TaskExecutionLog, error)
	GetRecentExecutionLogs(limit int) ([]*models.TaskExecutionLog, error)
	CleanupRunningLogs() error
}

type taskRepository struct {
	db *gorm.DB
}

// NewTaskRepository 创建任务仓储实例
func NewTaskRepository(db *gorm.DB) TaskRepository {
	return &taskRepository{db: db}
}

// Create 创建任务
func (r *taskRepository) Create(task *models.ScheduledTask) error {
	return r.db.Create(task).Error
}

// Update 更新任务
func (r *taskRepository) Update(task *models.ScheduledTask) error {
	return r.db.Save(task).Error
}

// Delete 删除任务（级联删除执行日志）
func (r *taskRepository) Delete(id uint) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		// 删除执行日志
		if err := tx.Where("task_id = ?", id).Delete(&models.TaskExecutionLog{}).Error; err != nil {
			return err
		}
		// 删除任务
		return tx.Delete(&models.ScheduledTask{}, id).Error
	})
}

// FindByID 根据 ID 查找任务
func (r *taskRepository) FindByID(id uint) (*models.ScheduledTask, error) {
	var task models.ScheduledTask
	err := r.db.First(&task, id).Error
	if err != nil {
		return nil, err
	}
	return &task, nil
}

// List 列出任务
func (r *taskRepository) List(enabled *bool) ([]*models.ScheduledTask, error) {
	var tasks []*models.ScheduledTask
	query := r.db.Order("created_at DESC")

	if enabled != nil {
		query = query.Where("enabled = ?", *enabled)
	}

	err := query.Find(&tasks).Error
	return tasks, err
}

// CreateExecutionLog 创建执行日志
func (r *taskRepository) CreateExecutionLog(log *models.TaskExecutionLog) error {
	return r.db.Create(log).Error
}

// UpdateExecutionLog 更新执行日志
func (r *taskRepository) UpdateExecutionLog(log *models.TaskExecutionLog) error {
	return r.db.Save(log).Error
}

// GetExecutionLogs 获取指定任务的执行日志
func (r *taskRepository) GetExecutionLogs(taskID uint, limit int) ([]*models.TaskExecutionLog, error) {
	var logs []*models.TaskExecutionLog
	query := r.db.Where("task_id = ?", taskID).
		Order("start_time DESC")

	if limit > 0 {
		query = query.Limit(limit)
	}

	err := query.Find(&logs).Error
	return logs, err
}

// GetRecentLogs 获取最近的执行日志
func (r *taskRepository) GetRecentLogs(limit int) ([]*models.TaskExecutionLog, error) {
	var logs []*models.TaskExecutionLog
	query := r.db.Order("start_time DESC")

	if limit > 0 {
		query = query.Limit(limit)
	}

	err := query.Find(&logs).Error
	return logs, err
}

// GetRecentExecutionLogs 获取最近的执行日志（别名方法）
func (r *taskRepository) GetRecentExecutionLogs(limit int) ([]*models.TaskExecutionLog, error) {
	return r.GetRecentLogs(limit)
}

// CleanupRunningLogs 清理应用重启时正在运行的任务状态
func (r *taskRepository) CleanupRunningLogs() error {
	return r.db.Model(&models.TaskExecutionLog{}).
		Where("status = ?", "running").
		Update("status", "canceled").
		Error
}
