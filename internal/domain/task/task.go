package task

import (
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
)

// TaskStatus represents the status of a task
type TaskStatus string

const (
	StatusPending    TaskStatus = "pending"
	StatusInProgress TaskStatus = "in_progress"
	StatusCompleted  TaskStatus = "completed"
	StatusCancelled  TaskStatus = "cancelled"
)

// Task represents a task in the system
type Task struct {
	ID        uuid.UUID  `json:"id"`
	Title     string     `json:"title"`
	Status    TaskStatus `json:"status"`
	UserID    uuid.UUID  `json:"user_id"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
}

// CreateTaskRequest represents a request to create a task
type CreateTaskRequest struct {
	Title string `json:"title" validate:"required,min=1,max=200"`
}

// UpdateTaskRequest represents a request to update a task
type UpdateTaskRequest struct {
	Title  *string     `json:"title,omitempty" validate:"omitempty,min=1,max=200"`
	Status *TaskStatus `json:"status,omitempty" validate:"omitempty,oneof=pending in_progress completed cancelled"`
}

// TaskFilter represents filters for task queries
type TaskFilter struct {
	Status *TaskStatus `json:"status,omitempty"`
	Search string      `json:"search,omitempty"`
}

// TaskSort represents sorting options for task queries
type TaskSort struct {
	Field string `json:"field"` // created_at, updated_at, title, status
	Order string `json:"order"` // asc, desc
}

// NewTask creates a new task instance
func NewTask(title string, userID uuid.UUID) *Task {
	return &Task{
		ID:        uuid.New(),
		Title:     title,
		Status:    StatusPending,
		UserID:    userID,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}

// ValidateCreateRequest validates create task request
func (req *CreateTaskRequest) Validate() error {
	if strings.TrimSpace(req.Title) == "" {
		return errors.New("title is required")
	}

	if len(req.Title) > 200 {
		return errors.New("title must be at most 200 characters")
	}

	return nil
}

// ValidateUpdateRequest validates update task request
func (req *UpdateTaskRequest) Validate() error {
	if req.Title != nil {
		if strings.TrimSpace(*req.Title) == "" {
			return errors.New("title cannot be empty")
		}
		if len(*req.Title) > 200 {
			return errors.New("title must be at most 200 characters")
		}
	}

	if req.Status != nil && !isValidStatus(*req.Status) {
		return errors.New("invalid status")
	}

	return nil
}

// Update updates the task with the provided request
func (t *Task) Update(req *UpdateTaskRequest) {
	if req.Title != nil {
		t.Title = *req.Title
	}
	if req.Status != nil {
		t.Status = *req.Status
	}
	t.UpdatedAt = time.Now()
}

// Helper functions
func isValidStatus(status TaskStatus) bool {
	switch status {
	case StatusPending, StatusInProgress, StatusCompleted, StatusCancelled:
		return true
	default:
		return false
	}
}
