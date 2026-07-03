package v1

import (
	"crm-project/internal/service"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

type TaskHandler struct {
	taskService *service.TaskService
}

func NewTaskHandler(taskService *service.TaskService) *TaskHandler {
	return &TaskHandler{taskService: taskService}
}

type CreateTaskRequest struct {
	Title       string     `json:"title" binding:"required"`
	Description *string    `json:"description"`
	DueDate     *time.Time `json:"due_date"`
	LeadID      *string    `json:"lead_id"`
	Priority    *string    `json:"priority"`
	Category    *string    `json:"category"`
}

type UpdateTaskRequest struct {
	Title       string     `json:"title" binding:"required"`
	Description *string    `json:"description"`
	DueDate     *time.Time `json:"due_date"`
	Priority    *string    `json:"priority"`
	Category    *string    `json:"category"`
}

func (h *TaskHandler) CreateTask(c *gin.Context) {
	var req CreateTaskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	userID := c.MustGet("user_id").(string)

	priority := "MEDIUM"
	if req.Priority != nil {
		priority = *req.Priority
	}
	category := "OTHER"
	if req.Category != nil {
		category = *req.Category
	}

	data, err := h.taskService.CreateTask(req.Title, req.Description, req.DueDate, userID, req.LeadID, priority, category)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"status": "Task berhasil dibuat", "data": data})
}

func (h *TaskHandler) GetAllTasks(c *gin.Context) {
	userID := c.MustGet("user_id").(string)
	role := c.MustGet("role").(string)
	priority := c.Query("priority")
	category := c.Query("category")
	data, err := h.taskService.GetAllTasks(userID, role, priority, category)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": data})
}

func (h *TaskHandler) GetTaskByID(c *gin.Context) {
	taskID := c.Param("id")
	data, err := h.taskService.GetTaskByID(taskID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Task tidak ditemukan"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": data})
}

func (h *TaskHandler) UpdateTask(c *gin.Context) {
	taskID := c.Param("id")
	var req UpdateTaskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	userID := c.MustGet("user_id").(string)
	role := c.MustGet("role").(string)

	priority := "MEDIUM"
	if req.Priority != nil {
		priority = *req.Priority
	}
	category := "OTHER"
	if req.Category != nil {
		category = *req.Category
	}

	err := h.taskService.UpdateTask(taskID, req.Title, req.Description, req.DueDate, priority, category, userID, role)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "Task berhasil diupdate"})
}

func (h *TaskHandler) DeleteTask(c *gin.Context) {
	taskID := c.Param("id")
	userID := c.MustGet("user_id").(string)
	role := c.MustGet("role").(string)
	err := h.taskService.DeleteTask(taskID, userID, role)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "Task berhasil dihapus"})
}

func (h *TaskHandler) MarkAsDone(c *gin.Context) {
	taskID := c.Param("id")
	userID := c.MustGet("user_id").(string)
	role := c.MustGet("role").(string)
	err := h.taskService.MarkAsDone(taskID, userID, role)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "Task berhasil ditandai selesai"})
}
