package postgres

import (
	"crm-project/internal/models/entity"
	"time"

	"gorm.io/gorm"
)

type TaskRepository struct {
	dbgorm *gorm.DB
}

func NewTaskRepository(db *gorm.DB) *TaskRepository {
	return &TaskRepository{dbgorm: db}
}

func (r *TaskRepository) CreateTask(task *entity.Task) (*entity.Task, error) {
	err := r.dbgorm.Create(task).Error
	if err != nil {
		return nil, err
	}
	return task, nil
}

func (r *TaskRepository) GetAll(userID string, role string, priority string, category string) ([]entity.Task, error) {
	var tasks []entity.Task
	query := r.dbgorm.Model(&entity.Task{})
	if role == "sales" {
		query = query.Where("assigned_to = ?", userID)
	}
	if priority != "" {
		query = query.Where("priority = ?", priority)
	}
	if category != "" {
		query = query.Where("category = ?", category)
	}
	err := query.Find(&tasks).Error
	return tasks, err
}

func (r *TaskRepository) GetByID(taskID string) (*entity.Task, error) {
	var task entity.Task
	err := r.dbgorm.Where("id = ?", taskID).First(&task).Error
	if err != nil {
		return nil, err
	}
	return &task, nil
}

func (r *TaskRepository) UpdateTask(taskID string, title string, description *string, dueDate interface{}, priority string, category string) error {
	return r.dbgorm.Model(&entity.Task{}).Where("id = ?", taskID).Updates(map[string]interface{}{
		"title":       title,
		"description": description,
		"due_date":    dueDate,
		"priority":    priority,
		"category":    category,
	}).Error
}

func (r *TaskRepository) GetTasksDueTomorrow() ([]entity.Task, error) {
	tomorrow := time.Now().Truncate(24 * time.Hour).Add(24 * time.Hour)
	dayAfter := tomorrow.Add(24 * time.Hour)
	var task []entity.Task
	err := r.dbgorm.Where("status = ? AND due_date >= ? AND due_date < ? AND reminder_sent = ?",
		"PENDING", tomorrow, dayAfter, false).Find(&task).Error
	if err != nil {
		return nil, err
	}
	return task, nil
}

func (r *TaskRepository) MarkReminderSent(taskID string) error {
	return r.dbgorm.Model(&entity.Task{}).Where("id = ?", taskID).Update("reminder_sent", true).Error
}

func (r *TaskRepository) DeleteTask(taskID string) error {
	return r.dbgorm.Where("id = ?", taskID).Delete(&entity.Task{}).Error
}

func (r *TaskRepository) MarkAsDone(taskID string) error {
	return r.dbgorm.Model(&entity.Task{}).Where("id = ?", taskID).Update("status", "DONE").Error
}
