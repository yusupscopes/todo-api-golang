package task

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewTask(t *testing.T) {
	title := "Test Task"
	userID := uuid.New()

	task := NewTask(title, userID)

	assert.NotNil(t, task)
	assert.Equal(t, title, task.Title)
	assert.Equal(t, StatusPending, task.Status)
	assert.Equal(t, userID, task.UserID)
	assert.NotEqual(t, uuid.Nil, task.ID)
	assert.False(t, task.CreatedAt.IsZero())
	assert.False(t, task.UpdatedAt.IsZero())
}

func TestCreateTaskRequest_Validate(t *testing.T) {
	tests := []struct {
		name    string
		request CreateTaskRequest
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid request",
			request: CreateTaskRequest{
				Title: "Valid Task Title",
			},
			wantErr: false,
		},
		{
			name: "empty title",
			request: CreateTaskRequest{
				Title: "",
			},
			wantErr: true,
			errMsg:  "title is required",
		},
		{
			name: "whitespace title",
			request: CreateTaskRequest{
				Title: "   ",
			},
			wantErr: true,
			errMsg:  "title is required",
		},
		{
			name: "title too long",
			request: CreateTaskRequest{
				Title: string(make([]byte, 201)), // 201 characters
			},
			wantErr: true,
			errMsg:  "title must be at most 200 characters",
		},
		{
			name: "title exactly 200 characters",
			request: CreateTaskRequest{
				Title: string(make([]byte, 200)), // 200 characters
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.request.Validate()
			if tt.wantErr {
				require.Error(t, err)
				assert.Equal(t, tt.errMsg, err.Error())
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestUpdateTaskRequest_Validate(t *testing.T) {
	tests := []struct {
		name    string
		request UpdateTaskRequest
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid request with title",
			request: UpdateTaskRequest{
				Title: stringPtr("Updated Title"),
			},
			wantErr: false,
		},
		{
			name: "valid request with status",
			request: UpdateTaskRequest{
				Status: statusPtr(StatusInProgress),
			},
			wantErr: false,
		},
		{
			name: "valid request with both",
			request: UpdateTaskRequest{
				Title:  stringPtr("Updated Title"),
				Status: statusPtr(StatusCompleted),
			},
			wantErr: false,
		},
		{
			name:    "empty request",
			request: UpdateTaskRequest{},
			wantErr: false,
		},
		{
			name: "empty title",
			request: UpdateTaskRequest{
				Title: stringPtr(""),
			},
			wantErr: true,
			errMsg:  "title cannot be empty",
		},
		{
			name: "whitespace title",
			request: UpdateTaskRequest{
				Title: stringPtr("   "),
			},
			wantErr: true,
			errMsg:  "title cannot be empty",
		},
		{
			name: "title too long",
			request: UpdateTaskRequest{
				Title: stringPtr(string(make([]byte, 201))),
			},
			wantErr: true,
			errMsg:  "title must be at most 200 characters",
		},
		{
			name: "invalid status",
			request: UpdateTaskRequest{
				Status: statusPtr(TaskStatus("invalid")),
			},
			wantErr: true,
			errMsg:  "invalid status",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.request.Validate()
			if tt.wantErr {
				require.Error(t, err)
				assert.Equal(t, tt.errMsg, err.Error())
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestTask_Update(t *testing.T) {
	originalTime := time.Now().Add(-1 * time.Hour)
	task := &Task{
		ID:        uuid.New(),
		Title:     "Original Title",
		Status:    StatusPending,
		UserID:    uuid.New(),
		CreatedAt: originalTime,
		UpdatedAt: originalTime,
	}

	// Test updating title only
	updateReq := &UpdateTaskRequest{
		Title: stringPtr("Updated Title"),
	}

	task.Update(updateReq)

	assert.Equal(t, "Updated Title", task.Title)
	assert.Equal(t, StatusPending, task.Status) // Should remain unchanged
	assert.True(t, task.UpdatedAt.After(originalTime))

	// Test updating status only
	originalUpdatedAt := task.UpdatedAt
	updateReq = &UpdateTaskRequest{
		Status: statusPtr(StatusInProgress),
	}

	task.Update(updateReq)

	assert.Equal(t, "Updated Title", task.Title) // Should remain unchanged
	assert.Equal(t, StatusInProgress, task.Status)
	assert.True(t, task.UpdatedAt.After(originalUpdatedAt))

	// Test updating both
	originalUpdatedAt = task.UpdatedAt
	updateReq = &UpdateTaskRequest{
		Title:  stringPtr("Final Title"),
		Status: statusPtr(StatusCompleted),
	}

	task.Update(updateReq)

	assert.Equal(t, "Final Title", task.Title)
	assert.Equal(t, StatusCompleted, task.Status)
	assert.True(t, task.UpdatedAt.After(originalUpdatedAt))
}

func TestIsValidStatus(t *testing.T) {
	tests := []struct {
		name   string
		status TaskStatus
		valid  bool
	}{
		{"pending", StatusPending, true},
		{"in_progress", StatusInProgress, true},
		{"completed", StatusCompleted, true},
		{"cancelled", StatusCancelled, true},
		{"invalid", TaskStatus("invalid"), false},
		{"empty", TaskStatus(""), false},
		{"PENDING", TaskStatus("PENDING"), false}, // Case sensitive
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isValidStatus(tt.status)
			assert.Equal(t, tt.valid, result)
		})
	}
}

func TestTaskStatus_Constants(t *testing.T) {
	assert.Equal(t, TaskStatus("pending"), StatusPending)
	assert.Equal(t, TaskStatus("in_progress"), StatusInProgress)
	assert.Equal(t, TaskStatus("completed"), StatusCompleted)
	assert.Equal(t, TaskStatus("cancelled"), StatusCancelled)
}

func TestTaskFilter(t *testing.T) {
	filter := &TaskFilter{
		Status: statusPtr(StatusPending),
		Search: "test",
	}

	assert.Equal(t, StatusPending, *filter.Status)
	assert.Equal(t, "test", filter.Search)

	// Test empty filter
	emptyFilter := &TaskFilter{}
	assert.Nil(t, emptyFilter.Status)
	assert.Empty(t, emptyFilter.Search)
}

func TestTaskSort(t *testing.T) {
	sort := &TaskSort{
		Field: "created_at",
		Order: "desc",
	}

	assert.Equal(t, "created_at", sort.Field)
	assert.Equal(t, "desc", sort.Order)
}

// Helper functions for tests
func stringPtr(s string) *string {
	return &s
}

func statusPtr(s TaskStatus) *TaskStatus {
	return &s
}
