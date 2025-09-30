package task

import (
	"strconv"
	"strings"

	"todo-api/internal/domain/task"
	authService "todo-api/internal/service/auth"
	taskService "todo-api/internal/service/task"
	"todo-api/pkg/types"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

// Handler handles task HTTP requests
type Handler struct {
	taskService taskService.Service
}

// NewHandler creates a new task handler instance
func NewHandler(authSvc authService.Service) *Handler {
	// Initialize service
	taskSvc := taskService.NewService(authSvc)

	return &Handler{
		taskService: taskSvc,
	}
}

// CreateTask handles task creation
func (h *Handler) CreateTask(c *fiber.Ctx) error {
	var req task.CreateTaskRequest

	// Parse request body
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   true,
			"message": "Invalid request body",
		})
	}

	// Get user ID from context (set by auth middleware)
	userID := c.Locals("user_id").(uuid.UUID)

	// Create task
	newTask, err := h.taskService.CreateTask(&req, userID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   true,
			"message": err.Error(),
		})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"error":   false,
		"message": "Task created successfully",
		"data":    newTask,
	})
}

// GetTask handles getting a single task
func (h *Handler) GetTask(c *fiber.Ctx) error {
	// Parse task ID from URL parameter
	taskIDStr := c.Params("id")
	taskID, err := uuid.Parse(taskIDStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   true,
			"message": "Invalid task ID",
		})
	}

	// Get user ID from context
	userID := c.Locals("user_id").(uuid.UUID)

	// Get task
	task, err := h.taskService.GetTaskByID(taskID, userID)
	if err != nil {
		if err.Error() == "task not found" {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error":   true,
				"message": "Task not found",
			})
		}
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error":   true,
			"message": err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"error":   false,
		"message": "Task retrieved successfully",
		"data":    task,
	})
}

// UpdateTask handles task updates
func (h *Handler) UpdateTask(c *fiber.Ctx) error {
	// Parse task ID from URL parameter
	taskIDStr := c.Params("id")
	taskID, err := uuid.Parse(taskIDStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   true,
			"message": "Invalid task ID",
		})
	}

	var req task.UpdateTaskRequest

	// Parse request body
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   true,
			"message": "Invalid request body",
		})
	}

	// Get user ID from context
	userID := c.Locals("user_id").(uuid.UUID)

	// Update task
	updatedTask, err := h.taskService.UpdateTask(taskID, &req, userID)
	if err != nil {
		if err.Error() == "task not found" {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error":   true,
				"message": "Task not found",
			})
		}
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   true,
			"message": err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"error":   false,
		"message": "Task updated successfully",
		"data":    updatedTask,
	})
}

// DeleteTask handles task deletion
func (h *Handler) DeleteTask(c *fiber.Ctx) error {
	// Parse task ID from URL parameter
	taskIDStr := c.Params("id")
	taskID, err := uuid.Parse(taskIDStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   true,
			"message": "Invalid task ID",
		})
	}

	// Get user ID from context
	userID := c.Locals("user_id").(uuid.UUID)

	// Delete task
	err = h.taskService.DeleteTask(taskID, userID)
	if err != nil {
		if err.Error() == "task not found" {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error":   true,
				"message": "Task not found",
			})
		}
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error":   true,
			"message": err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"error":   false,
		"message": "Task deleted successfully",
	})
}

// ListTasks handles task listing with filtering, sorting, and pagination
func (h *Handler) ListTasks(c *fiber.Ctx) error {
	// Get user ID from context
	userID := c.Locals("user_id").(uuid.UUID)

	// Parse query parameters
	filter := h.parseFilter(c)
	sort := h.parseSort(c)
	page, limit := h.parsePagination(c)

	// Get tasks
	tasks, paginationInfo, err := h.taskService.ListTasks(filter, sort, page, limit, userID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   true,
			"message": "Failed to retrieve tasks",
		})
	}

	// Prepare meta information
	meta := &types.MetaInfo{
		Pagination: *paginationInfo,
	}

	if sort != nil {
		meta.Sort = sort.Field + ":" + sort.Order
	}

	if filter != nil {
		var filterParts []string
		if filter.Status != nil {
			filterParts = append(filterParts, "status:"+string(*filter.Status))
		}
		if filter.Search != "" {
			filterParts = append(filterParts, "search:"+filter.Search)
		}
		meta.Filter = strings.Join(filterParts, ",")
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"error":   false,
		"message": "Tasks retrieved successfully",
		"data":    tasks,
		"meta":    meta,
	})
}

// parseFilter parses filter parameters from query string
func (h *Handler) parseFilter(c *fiber.Ctx) *task.TaskFilter {
	filter := &task.TaskFilter{}

	// Status filter
	if statusStr := c.Query("status"); statusStr != "" {
		status := task.TaskStatus(statusStr)
		filter.Status = &status
	}

	// Search filter
	if search := c.Query("search"); search != "" {
		filter.Search = search
	}

	// Return nil if no filters are applied
	if filter.Status == nil && filter.Search == "" {
		return nil
	}

	return filter
}

// parseSort parses sort parameters from query string
func (h *Handler) parseSort(c *fiber.Ctx) *task.TaskSort {
	sortField := c.Query("sort_field", "created_at")
	sortOrder := c.Query("sort_order", "desc")

	// Validate sort field
	validFields := map[string]bool{
		"created_at": true,
		"updated_at": true,
		"title":      true,
		"status":     true,
	}

	if !validFields[sortField] {
		sortField = "created_at"
	}

	// Validate sort order
	if sortOrder != "asc" && sortOrder != "desc" {
		sortOrder = "desc"
	}

	return &task.TaskSort{
		Field: sortField,
		Order: sortOrder,
	}
}

// parsePagination parses pagination parameters from query string
func (h *Handler) parsePagination(c *fiber.Ctx) (int, int) {
	page := 1
	limit := 10

	if pageStr := c.Query("page"); pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}

	if limitStr := c.Query("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 100 {
			limit = l
		}
	}

	return page, limit
}
