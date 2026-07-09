package v1

import (
	"crm-project/internal/service"
	"crm-project/pkg/response"
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

// @Summary      Buat task baru
// @Tags         Tasks
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        body body CreateTaskRequest true "Data task"
// @Success      201 {object} response.Response
// @Failure      400 {object} response.Response
// @Router       /tasks [post]
func (h *TaskHandler) CreateTask(c *gin.Context) {
	var req CreateTaskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.Error(err.Error()))
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
		c.JSON(http.StatusBadRequest, response.Error(err.Error()))
		return
	}
	c.JSON(http.StatusCreated, response.Success("Task berhasil dibuat", data))
}

// @Summary      List semua task
// @Tags         Tasks
// @Produce      json
// @Security     BearerAuth
// @Param        priority query string false "Filter priority (LOW/MEDIUM/HIGH/URGENT)"
// @Param        category query string false "Filter category (FOLLOW_UP/MEETING/CALL/EMAIL/OTHER)"
// @Success      200 {object} response.Response
// @Router       /tasks [get]
func (h *TaskHandler) GetAllTasks(c *gin.Context) {
	userID := c.MustGet("user_id").(string)
	role := c.MustGet("role").(string)
	priority := c.Query("priority")
	category := c.Query("category")
	data, err := h.taskService.GetAllTasks(userID, role, priority, category)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Error(err.Error()))
		return
	}
	c.JSON(http.StatusOK, response.Success("Berhasil mengambil tasks", data))
}

// @Summary      Get task by ID
// @Tags         Tasks
// @Produce      json
// @Security     BearerAuth
// @Param        id path string true "Task ID"
// @Success      200 {object} response.Response
// @Failure      404 {object} response.Response
// @Router       /tasks/{id} [get]
func (h *TaskHandler) GetTaskByID(c *gin.Context) {
	taskID := c.Param("id")
	data, err := h.taskService.GetTaskByID(taskID)
	if err != nil {
		c.JSON(http.StatusNotFound, response.Error("Task tidak ditemukan"))
		return
	}
	c.JSON(http.StatusOK, response.Success("Berhasil mengambil task", data))
}

// @Summary      Update task
// @Tags         Tasks
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id path string true "Task ID"
// @Param        body body UpdateTaskRequest true "Data task"
// @Success      200 {object} response.Response
// @Failure      400 {object} response.Response
// @Router       /tasks/{id} [put]
func (h *TaskHandler) UpdateTask(c *gin.Context) {
	taskID := c.Param("id")
	var req UpdateTaskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.Error(err.Error()))
		return
	}
	userID := c.MustGet("user_id").(string)
	role := c.MustGet("role").(string)

	existingTask, errTask := h.taskService.GetTaskByID(taskID)
	if errTask != nil {
		c.JSON(http.StatusNotFound, response.Error("Task tidak ditemukan"))
		return
	}

	priority := existingTask.Priority
	if req.Priority != nil {
		priority = *req.Priority
	}
	
	category := existingTask.Category
	if req.Category != nil {
		category = *req.Category
	}

	err := h.taskService.UpdateTask(taskID, req.Title, req.Description, req.DueDate, priority, category, userID, role)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.Error(err.Error()))
		return
	}
	c.JSON(http.StatusOK, response.Success("Task berhasil diupdate", nil))
}

// @Summary      List task yang dihapus (admin)
// @Tags         Tasks
// @Produce      json
// @Security     BearerAuth
// @Success      200 {object} response.Response
// @Router       /tasks/trash [get]
func (h *TaskHandler) GetTrashedTasks(c *gin.Context) {
	tasks, err := h.taskService.GetTrashedTasks()
	if err != nil {
		c.JSON(http.StatusBadRequest, response.Error(err.Error()))
		return
	}
	c.JSON(http.StatusOK, response.Success("Berhasil mengambil tasks yang dihapus", tasks))
}

// @Summary      Restore task yang dihapus (admin)
// @Tags         Tasks
// @Produce      json
// @Security     BearerAuth
// @Param        id path string true "Task ID"
// @Success      200 {object} response.Response
// @Failure      400 {object} response.Response
// @Router       /tasks/{id}/restore [post]
func (h *TaskHandler) RestoreTask(c *gin.Context) {
	taskID := c.Param("id")
	if err := h.taskService.RestoreTask(taskID); err != nil {
		c.JSON(http.StatusBadRequest, response.Error(err.Error()))
		return
	}
	c.JSON(http.StatusOK, response.Success("Berhasil mengembalikan task", nil))
}

// @Summary      Hapus task
// @Tags         Tasks
// @Produce      json
// @Security     BearerAuth
// @Param        id path string true "Task ID"
// @Success      200 {object} response.Response
// @Failure      400 {object} response.Response
// @Router       /tasks/{id} [delete]
func (h *TaskHandler) DeleteTask(c *gin.Context) {
	taskID := c.Param("id")
	userID := c.MustGet("user_id").(string)
	role := c.MustGet("role").(string)
	err := h.taskService.DeleteTask(taskID, userID, role)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.Error(err.Error()))
		return
	}
	c.JSON(http.StatusOK, response.Success("Task berhasil dihapus", nil))
}

// @Summary      Tandai task selesai
// @Tags         Tasks
// @Produce      json
// @Security     BearerAuth
// @Param        id path string true "Task ID"
// @Success      200 {object} response.Response
// @Failure      400 {object} response.Response
// @Router       /tasks/{id}/done [patch]
func (h *TaskHandler) MarkAsDone(c *gin.Context) {
	taskID := c.Param("id")
	userID := c.MustGet("user_id").(string)
	role := c.MustGet("role").(string)
	err := h.taskService.MarkAsDone(taskID, userID, role)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.Error(err.Error()))
		return
	}
	c.JSON(http.StatusOK, response.Success("Task berhasil ditandai selesai", nil))
}
