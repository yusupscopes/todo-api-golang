package task

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"todo-api/internal/domain/task"
	"todo-api/internal/service/auth"
	"todo-api/pkg/config"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTestService(t *testing.T) Service {
	cfg := &config.Config{
		JWT: config.JWTConfig{
			SecretKey:       "test-secret",
			AccessTokenTTL:  15 * time.Minute,
			RefreshTokenTTL: 7 * 24 * time.Hour,
		},
	}

	authSvc := auth.NewService(cfg)
	return NewService(authSvc)
}

func TestNewService(t *testing.T) {
	cfg := &config.Config{
		JWT: config.JWTConfig{
			SecretKey:       "test-secret",
			AccessTokenTTL:  15 * time.Minute,
			RefreshTokenTTL: 7 * 24 * time.Hour,
		},
	}

	authSvc := auth.NewService(cfg)
	service := NewService(authSvc)

	assert.NotNil(t, service)
}

func TestService_CreateTask_ValidRequest(t *testing.T) {
	service := setupTestService(t)
	userID := uuid.MustParse("3484ec33-20f9-4993-a25f-f49f6f5dbe54") // john.doe@example.com

	req := &task.CreateTaskRequest{
		Title: "Test Task",
	}

	createdTask, err := service.CreateTask(req, userID)

	require.NoError(t, err)
	assert.NotNil(t, createdTask)
	assert.Equal(t, "Test Task", createdTask.Title)
	assert.Equal(t, task.StatusPending, createdTask.Status)
	assert.Equal(t, userID, createdTask.UserID)
	assert.NotEqual(t, uuid.Nil, createdTask.ID)
}

func TestService_CreateTask_InvalidRequest(t *testing.T) {
	service := setupTestService(t)
	userID := uuid.MustParse("3484ec33-20f9-4993-a25f-f49f6f5dbe54")

	req := &task.CreateTaskRequest{
		Title: "", // Invalid title
	}

	createdTask, err := service.CreateTask(req, userID)

	require.Error(t, err)
	assert.Nil(t, createdTask)
	assert.Equal(t, "title is required", err.Error())
}

func TestService_GetTaskByID_ExistingTask(t *testing.T) {
	service := setupTestService(t)
	userID := uuid.MustParse("3484ec33-20f9-4993-a25f-f49f6f5dbe54")

	// First create a task
	req := &task.CreateTaskRequest{
		Title: "Test Task",
	}

	createdTask, err := service.CreateTask(req, userID)
	require.NoError(t, err)

	// Then retrieve it
	retrievedTask, err := service.GetTaskByID(createdTask.ID, userID)

	require.NoError(t, err)
	assert.NotNil(t, retrievedTask)
	assert.Equal(t, createdTask.ID, retrievedTask.ID)
	assert.Equal(t, "Test Task", retrievedTask.Title)
	assert.Equal(t, userID, retrievedTask.UserID)
}

func TestService_GetTaskByID_NonExistingTask(t *testing.T) {
	service := setupTestService(t)
	userID := uuid.MustParse("3484ec33-20f9-4993-a25f-f49f6f5dbe54")
	nonExistingID := uuid.New()

	retrievedTask, err := service.GetTaskByID(nonExistingID, userID)

	require.Error(t, err)
	assert.Nil(t, retrievedTask)
	assert.Equal(t, "task not found", err.Error())
}

func TestService_GetTaskByID_WrongUser(t *testing.T) {
	service := setupTestService(t)
	user1ID := uuid.MustParse("3484ec33-20f9-4993-a25f-f49f6f5dbe54") // john.doe@example.com
	user2ID := uuid.MustParse("550e8400-e29b-41d4-a716-446655440002") // jane.smith@example.com

	// Create task for user1
	req := &task.CreateTaskRequest{
		Title: "User1 Task",
	}

	createdTask, err := service.CreateTask(req, user1ID)
	require.NoError(t, err)

	// Try to get task with user2
	retrievedTask, err := service.GetTaskByID(createdTask.ID, user2ID)

	require.Error(t, err)
	assert.Nil(t, retrievedTask)
	assert.Equal(t, "access denied", err.Error())
}

func TestService_UpdateTask_ValidRequest(t *testing.T) {
	service := setupTestService(t)
	userID := uuid.MustParse("3484ec33-20f9-4993-a25f-f49f6f5dbe54")

	// Create a task first
	createReq := &task.CreateTaskRequest{
		Title: "Original Title",
	}

	createdTask, err := service.CreateTask(createReq, userID)
	require.NoError(t, err)

	// Update the task
	updateReq := &task.UpdateTaskRequest{
		Title:  stringPtr("Updated Title"),
		Status: statusPtr(task.StatusInProgress),
	}

	updatedTask, err := service.UpdateTask(createdTask.ID, updateReq, userID)

	require.NoError(t, err)
	assert.NotNil(t, updatedTask)
	assert.Equal(t, "Updated Title", updatedTask.Title)
	assert.Equal(t, task.StatusInProgress, updatedTask.Status)
	assert.Equal(t, createdTask.ID, updatedTask.ID)
	assert.True(t, updatedTask.UpdatedAt.After(createdTask.UpdatedAt) || updatedTask.UpdatedAt.Equal(createdTask.UpdatedAt))
}

func TestService_UpdateTask_NonExistingTask(t *testing.T) {
	service := setupTestService(t)
	userID := uuid.MustParse("3484ec33-20f9-4993-a25f-f49f6f5dbe54")
	nonExistingID := uuid.New()

	updateReq := &task.UpdateTaskRequest{
		Title: stringPtr("Updated Title"),
	}

	updatedTask, err := service.UpdateTask(nonExistingID, updateReq, userID)

	require.Error(t, err)
	assert.Nil(t, updatedTask)
	assert.Equal(t, "task not found", err.Error())
}

func TestService_UpdateTask_InvalidRequest(t *testing.T) {
	service := setupTestService(t)
	userID := uuid.MustParse("3484ec33-20f9-4993-a25f-f49f6f5dbe54")

	// Create a task first
	createReq := &task.CreateTaskRequest{
		Title: "Original Title",
	}

	createdTask, err := service.CreateTask(createReq, userID)
	require.NoError(t, err)

	// Try to update with invalid request
	updateReq := &task.UpdateTaskRequest{
		Title: stringPtr(""), // Invalid title
	}

	updatedTask, err := service.UpdateTask(createdTask.ID, updateReq, userID)

	require.Error(t, err)
	assert.Nil(t, updatedTask)
	assert.Equal(t, "title cannot be empty", err.Error())
}

func TestService_DeleteTask_ExistingTask(t *testing.T) {
	service := setupTestService(t)
	userID := uuid.MustParse("3484ec33-20f9-4993-a25f-f49f6f5dbe54")

	// Create a task first
	req := &task.CreateTaskRequest{
		Title: "Task to Delete",
	}

	createdTask, err := service.CreateTask(req, userID)
	require.NoError(t, err)

	// Delete the task
	err = service.DeleteTask(createdTask.ID, userID)

	require.NoError(t, err)

	// Verify task is deleted
	_, err = service.GetTaskByID(createdTask.ID, userID)
	require.Error(t, err)
	assert.Equal(t, "task not found", err.Error())
}

func TestService_DeleteTask_NonExistingTask(t *testing.T) {
	service := setupTestService(t)
	userID := uuid.MustParse("3484ec33-20f9-4993-a25f-f49f6f5dbe54")
	nonExistingID := uuid.New()

	err := service.DeleteTask(nonExistingID, userID)

	require.Error(t, err)
	assert.Equal(t, "task not found", err.Error())
}

func TestService_ListTasks_NoFilters(t *testing.T) {
	service := setupTestService(t)
	userID := uuid.MustParse("3484ec33-20f9-4993-a25f-f49f6f5dbe54")

	// Create some tasks
	req1 := &task.CreateTaskRequest{Title: "Task 1"}
	req2 := &task.CreateTaskRequest{Title: "Task 2"}

	_, err := service.CreateTask(req1, userID)
	require.NoError(t, err)

	_, err = service.CreateTask(req2, userID)
	require.NoError(t, err)

	// List tasks
	tasks, pagination, err := service.ListTasks(nil, nil, 1, 10, userID)

	require.NoError(t, err)
	assert.NotNil(t, tasks)
	assert.NotNil(t, pagination)
	assert.GreaterOrEqual(t, len(tasks), 2) // At least 2 tasks we just created
	assert.Equal(t, 1, pagination.Page)
	assert.Equal(t, 10, pagination.Limit)
}

func TestService_ListTasks_WithStatusFilter(t *testing.T) {
	service := setupTestService(t)
	userID := uuid.MustParse("3484ec33-20f9-4993-a25f-f49f6f5dbe54")

	// Create tasks with different statuses
	req1 := &task.CreateTaskRequest{Title: "Pending Task"}
	req2 := &task.CreateTaskRequest{Title: "In Progress Task"}

	_, err := service.CreateTask(req1, userID)
	require.NoError(t, err)

	task2, err := service.CreateTask(req2, userID)
	require.NoError(t, err)

	// Update task2 to in_progress
	updateReq := &task.UpdateTaskRequest{Status: statusPtr(task.StatusInProgress)}
	_, err = service.UpdateTask(task2.ID, updateReq, userID)
	require.NoError(t, err)

	// Filter by pending status
	filter := &task.TaskFilter{
		Status: statusPtr(task.StatusPending),
	}

	tasks, pagination, err := service.ListTasks(filter, nil, 1, 10, userID)

	require.NoError(t, err)
	assert.NotNil(t, tasks)
	assert.NotNil(t, pagination)

	// All returned tasks should be pending
	for _, taskItem := range tasks {
		assert.Equal(t, task.StatusPending, taskItem.Status)
	}
}

func TestService_ListTasks_WithSearchFilter(t *testing.T) {
	service := setupTestService(t)
	userID := uuid.MustParse("3484ec33-20f9-4993-a25f-f49f6f5dbe54")

	// Create tasks with different titles
	req1 := &task.CreateTaskRequest{Title: "Documentation Task"}
	req2 := &task.CreateTaskRequest{Title: "Code Review Task"}

	_, err := service.CreateTask(req1, userID)
	require.NoError(t, err)

	_, err = service.CreateTask(req2, userID)
	require.NoError(t, err)

	// Search for "documentation"
	filter := &task.TaskFilter{
		Search: "documentation",
	}

	tasks, pagination, err := service.ListTasks(filter, nil, 1, 10, userID)

	require.NoError(t, err)
	assert.NotNil(t, tasks)
	assert.NotNil(t, pagination)

	// All returned tasks should contain "documentation" in title (case-insensitive)
	for _, task := range tasks {
		assert.Contains(t, strings.ToLower(task.Title), "documentation")
	}
}

func TestService_ListTasks_WithSorting(t *testing.T) {
	service := setupTestService(t)
	userID := uuid.MustParse("3484ec33-20f9-4993-a25f-f49f6f5dbe54")

	// Create tasks
	req1 := &task.CreateTaskRequest{Title: "A Task"}
	req2 := &task.CreateTaskRequest{Title: "B Task"}

	_, err := service.CreateTask(req1, userID)
	require.NoError(t, err)

	_, err = service.CreateTask(req2, userID)
	require.NoError(t, err)

	// Sort by title ascending
	sort := &task.TaskSort{
		Field: "title",
		Order: "asc",
	}

	tasks, pagination, err := service.ListTasks(nil, sort, 1, 10, userID)

	require.NoError(t, err)
	assert.NotNil(t, tasks)
	assert.NotNil(t, pagination)

	// Tasks should be sorted by title
	if len(tasks) >= 2 {
		assert.True(t, tasks[0].Title <= tasks[1].Title)
	}
}

func TestService_ListTasks_Pagination(t *testing.T) {
	service := setupTestService(t)
	userID := uuid.MustParse("3484ec33-20f9-4993-a25f-f49f6f5dbe54")

	// Create multiple tasks
	for i := 0; i < 5; i++ {
		req := &task.CreateTaskRequest{Title: fmt.Sprintf("Task %d", i)}
		_, err := service.CreateTask(req, userID)
		require.NoError(t, err)
	}

	// Test pagination
	tasks, pagination, err := service.ListTasks(nil, nil, 1, 2, userID)

	require.NoError(t, err)
	assert.NotNil(t, tasks)
	assert.NotNil(t, pagination)
	assert.Equal(t, 1, pagination.Page)
	assert.Equal(t, 2, pagination.Limit)
	assert.LessOrEqual(t, len(tasks), 2)
}

// Helper functions for tests
func stringPtr(s string) *string {
	return &s
}

func statusPtr(s task.TaskStatus) *task.TaskStatus {
	return &s
}
