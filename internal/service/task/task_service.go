package task

import (
	"errors"
	"sort"
	"strings"

	"todo-api/internal/domain/task"
	authService "todo-api/internal/service/auth"
	"todo-api/pkg/types"

	"github.com/google/uuid"
)

// Service defines the task service interface
type Service interface {
	CreateTask(req *task.CreateTaskRequest, userID uuid.UUID) (*task.Task, error)
	GetTaskByID(id uuid.UUID, userID uuid.UUID) (*task.Task, error)
	UpdateTask(id uuid.UUID, req *task.UpdateTaskRequest, userID uuid.UUID) (*task.Task, error)
	DeleteTask(id uuid.UUID, userID uuid.UUID) error
	ListTasks(filter *task.TaskFilter, sort *task.TaskSort, page, limit int, userID uuid.UUID) ([]*task.Task, *types.PaginationInfo, error)
}

// service implements the task service
type service struct {
	tasks       map[uuid.UUID]*task.Task // Mock task storage
	authService authService.Service
}

// NewService creates a new task service
func NewService(authSvc authService.Service) Service {
	// Initialize mock tasks
	tasks := make(map[uuid.UUID]*task.Task)

	// Get actual user IDs from auth service
	user1, _ := authSvc.GetUserByEmail("john.doe@example.com")
	user2, _ := authSvc.GetUserByEmail("jane.smith@example.com")

	if user1 != nil {
		// Tasks for user 1
		task1 := task.NewTask(
			"Complete project documentation",
			user1.ID,
		)
		task1.Status = task.StatusInProgress
		tasks[task1.ID] = task1

		task2 := task.NewTask(
			"Review code changes",
			user1.ID,
		)
		tasks[task2.ID] = task2
	}

	if user2 != nil {
		// Tasks for user 2
		task3 := task.NewTask(
			"Plan team meeting",
			user2.ID,
		)
		task3.Status = task.StatusCompleted
		tasks[task3.ID] = task3

		task4 := task.NewTask(
			"Update system configuration",
			user2.ID,
		)
		tasks[task4.ID] = task4
	}

	return &service{
		tasks:       tasks,
		authService: authSvc,
	}
}

// CreateTask creates a new task
func (s *service) CreateTask(req *task.CreateTaskRequest, userID uuid.UUID) (*task.Task, error) {
	// Validate request
	if err := req.Validate(); err != nil {
		return nil, err
	}

	// Create new task
	newTask := task.NewTask(req.Title, userID)

	// Store task
	s.tasks[newTask.ID] = newTask

	return newTask, nil
}

// GetTaskByID retrieves a task by ID
func (s *service) GetTaskByID(id uuid.UUID, userID uuid.UUID) (*task.Task, error) {
	task, exists := s.tasks[id]
	if !exists {
		return nil, errors.New("task not found")
	}

	// Check if user owns the task (or is admin)
	if task.UserID != userID {
		return nil, errors.New("access denied")
	}

	return task, nil
}

// UpdateTask updates an existing task
func (s *service) UpdateTask(id uuid.UUID, req *task.UpdateTaskRequest, userID uuid.UUID) (*task.Task, error) {
	// Validate request
	if err := req.Validate(); err != nil {
		return nil, err
	}

	// Find task
	task, exists := s.tasks[id]
	if !exists {
		return nil, errors.New("task not found")
	}

	// Check if user owns the task (or is admin)
	if task.UserID != userID {
		return nil, errors.New("access denied")
	}

	// Update task
	task.Update(req)

	return task, nil
}

// DeleteTask deletes a task
func (s *service) DeleteTask(id uuid.UUID, userID uuid.UUID) error {
	// Find task
	task, exists := s.tasks[id]
	if !exists {
		return errors.New("task not found")
	}

	// Check if user owns the task (or is admin)
	if task.UserID != userID {
		return errors.New("access denied")
	}

	// Delete task
	delete(s.tasks, id)

	return nil
}

// ListTasks retrieves tasks with filtering, sorting, and pagination
func (s *service) ListTasks(filter *task.TaskFilter, sort *task.TaskSort, page, limit int, userID uuid.UUID) ([]*task.Task, *types.PaginationInfo, error) {
	// Get all tasks for the user
	var userTasks []*task.Task
	for _, task := range s.tasks {
		if task.UserID == userID {
			userTasks = append(userTasks, task)
		}
	}

	// Apply filters
	filteredTasks := s.applyFilters(userTasks, filter)

	// Apply sorting
	sortedTasks := s.applySorting(filteredTasks, sort)

	// Calculate pagination
	total := int64(len(sortedTasks))
	totalPages := int((total + int64(limit) - 1) / int64(limit))

	// Apply pagination
	start := (page - 1) * limit
	end := start + limit

	if start >= len(sortedTasks) {
		return []*task.Task{}, &types.PaginationInfo{
			Page:       page,
			Limit:      limit,
			Total:      total,
			TotalPages: totalPages,
		}, nil
	}

	if end > len(sortedTasks) {
		end = len(sortedTasks)
	}

	paginatedTasks := sortedTasks[start:end]

	paginationInfo := &types.PaginationInfo{
		Page:       page,
		Limit:      limit,
		Total:      total,
		TotalPages: totalPages,
	}

	return paginatedTasks, paginationInfo, nil
}

// applyFilters applies filters to the task list
func (s *service) applyFilters(tasks []*task.Task, filter *task.TaskFilter) []*task.Task {
	if filter == nil {
		return tasks
	}

	var filtered []*task.Task
	for _, task := range tasks {
		// Status filter
		if filter.Status != nil && task.Status != *filter.Status {
			continue
		}

		// Search filter
		if filter.Search != "" {
			searchLower := strings.ToLower(filter.Search)
			titleMatch := strings.Contains(strings.ToLower(task.Title), searchLower)
			if !titleMatch {
				continue
			}
		}

		filtered = append(filtered, task)
	}

	return filtered
}

// applySorting applies sorting to the task list
func (s *service) applySorting(tasks []*task.Task, sortOptions *task.TaskSort) []*task.Task {
	if sortOptions == nil {
		// Default sort by created_at desc
		sortOptions = &task.TaskSort{Field: "created_at", Order: "desc"}
	}

	sort.Slice(tasks, func(i, j int) bool {
		switch sortOptions.Field {
		case "title":
			if sortOptions.Order == "asc" {
				return tasks[i].Title < tasks[j].Title
			}
			return tasks[i].Title > tasks[j].Title
		case "status":
			statusOrder := map[task.TaskStatus]int{
				task.StatusPending:    1,
				task.StatusInProgress: 2,
				task.StatusCompleted:  3,
				task.StatusCancelled:  4,
			}
			if sortOptions.Order == "asc" {
				return statusOrder[tasks[i].Status] < statusOrder[tasks[j].Status]
			}
			return statusOrder[tasks[i].Status] > statusOrder[tasks[j].Status]
		case "updated_at":
			if sortOptions.Order == "asc" {
				return tasks[i].UpdatedAt.Before(tasks[j].UpdatedAt)
			}
			return tasks[i].UpdatedAt.After(tasks[j].UpdatedAt)
		case "created_at":
			fallthrough
		default:
			if sortOptions.Order == "asc" {
				return tasks[i].CreatedAt.Before(tasks[j].CreatedAt)
			}
			return tasks[i].CreatedAt.After(tasks[j].CreatedAt)
		}
	})

	return tasks
}
