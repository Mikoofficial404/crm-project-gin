package service

import (
	"crm-project/internal/delivery/websocket"
	"crm-project/internal/models/entity"
	"crm-project/internal/repository/postgres"
	"errors"
	"fmt"
	"log"
	"time"
)

var validPriorities = map[string]bool{
	"LOW": true, "MEDIUM": true, "HIGH": true, "URGENT": true,
}

var validCategories = map[string]bool{
	"FOLLOW_UP": true, "MEETING": true, "CALL": true, "EMAIL": true, "OTHER": true,
}

func validatePriority(p string) error {
	if !validPriorities[p] {
		return errors.New("priority tidak valid: gunakan LOW, MEDIUM, HIGH, atau URGENT")
	}
	return nil
}

func validateCategory(c string) error {
	if !validCategories[c] {
		return errors.New("category tidak valid: gunakan FOLLOW_UP, MEETING, CALL, EMAIL, atau OTHER")
	}
	return nil
}

type TaskService struct {
	task  *postgres.TaskRepository
	user  *postgres.UserRepository
	notif *NotificationService
}

func NewTaskService(taskRepo *postgres.TaskRepository, userRepo *postgres.UserRepository, notifService *NotificationService) *TaskService {
	return &TaskService{task: taskRepo, user: userRepo, notif: notifService}
}

func (s *TaskService) CreateTask(title string, description *string, dueDate *time.Time, assignedTo string, leadID *string, priority string, category string) (*entity.Task, error) {
	if title == "" || assignedTo == "" {
		return nil, errors.New("title dan assigned_to wajib diisi")
	}
	if dueDate != nil && dueDate.Before(time.Now()) {
		return nil, errors.New("due_date tidak boleh di masa lalu")
	}
	if err := validatePriority(priority); err != nil {
		return nil, err
	}
	if err := validateCategory(category); err != nil {
		return nil, err
	}

	task := entity.Task{
		Title:       title,
		Description: description,
		DueDate:     dueDate,
		Status:      "PENDING",
		AssignedTo:  assignedTo,
		LeadID:      leadID,
		Priority:    priority,
		Category:    category,
	}
	return s.task.CreateTask(&task)
}

func (s *TaskService) GetAllTasks(userID string, role string, priority string, category string) ([]entity.Task, error) {
	return s.task.GetAll(userID, role, priority, category)
}

func (s *TaskService) GetTaskByID(taskID string) (*entity.Task, error) {
	return s.task.GetByID(taskID)
}

func (s *TaskService) UpdateTask(taskID string, title string, description *string, dueDate *time.Time, priority string, category string, userID string, role string) error {
	task, err := s.task.GetByID(taskID)
	if err != nil {
		return errors.New("task tidak ditemukan")
	}
	if role == "sales" && task.AssignedTo != userID {
		return errors.New("unauthorized: ini bukan task Anda")
	}
	if dueDate != nil && dueDate.Before(time.Now()) {
		return errors.New("due_date tidak boleh di masa lalu")
	}
	if err := validatePriority(priority); err != nil {
		return err
	}
	if err := validateCategory(category); err != nil {
		return err
	}

	return s.task.UpdateTask(taskID, title, description, dueDate, priority, category)
}
func (s *TaskService) DeleteTask(taskID string, userID string, role string) error {
	task, err := s.task.GetByID(taskID)
	if err != nil {
		return errors.New("task tidak ditemukan")
	}
	if role == "sales" && task.AssignedTo != userID {
		return errors.New("unauthorized: ini bukan task Anda")
	}
	return s.task.DeleteTask(taskID)
}

func (s *TaskService) MarkAsDone(taskID string, userID string, role string) error {
	task, err := s.task.GetByID(taskID)
	if err != nil {
		return errors.New("task tidak ditemukan")
	}
	if role == "sales" && task.AssignedTo != userID {
		return errors.New("unauthorized: ini bukan task Anda")
	}
	return s.task.MarkAsDone(taskID)
}

func (s *TaskService) GetTrashedTasks() ([]entity.Task, error) {
	return s.task.GetTrashedTasks()
}

func (s *TaskService) RestoreTask(taskID string) error {
	if taskID == "" {
		return errors.New("ID task wajib diisi")
	}
	return s.task.RestoreTask(taskID)
}

func (s *TaskService) SendDueDateReminders() error {
	task, err := s.task.GetTasksDueTomorrow()
	if err != nil {
		return err
	}
	for _, task := range task {
		if task.AssignedTo == "" {
			continue
		}
		user, err := s.user.FindByID(task.AssignedTo)
		if err != nil {
			log.Printf("SendDueDateReminders: user not found for task %s: %v", task.ID, err)
			continue
		}
		title := "Pengingat Jatuh Tempo"
		message := fmt.Sprintf("Task %s kamu jatuh tempo besok!", task.Title)
		err = s.notif.CreateNotification(user.ID, title, message)
		if err != nil {
			log.Printf("SendDueDateReminders: failed to create notification for user %s: %v", user.ID, err)
			continue
		}
		websocket.SendMessageToUser(user.ID, message)
		if err := s.task.MarkReminderSent(task.ID); err != nil {
			log.Printf("SendDueDateReminders: failed to mark reminder sent for task %s: %v", task.ID, err)
		}
	}
	return  nil
}
