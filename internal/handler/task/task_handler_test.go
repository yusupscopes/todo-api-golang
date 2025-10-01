package task

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"todo-api/internal/domain/task"
	"todo-api/internal/service/auth"
	"todo-api/pkg/config"
	"todo-api/pkg/utils"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTestHandler(t *testing.T) (*Handler, string) {
	cfg := &config.Config{
		JWT: config.JWTConfig{
			SecretKey:       "test-secret",
			AccessTokenTTL:  15 * time.Minute,
			RefreshTokenTTL: 7 * 24 * time.Hour,
		},
	}

	authSvc := auth.NewService(cfg)
	handler := NewHandler(authSvc)

	// Generate a valid token for testing
	userID := uuid.MustParse("3484ec33-20f9-4993-a25f-f49f6f5dbe54")
	token, err := utils.GenerateToken(cfg.JWT.SecretKey, userID, "john.doe@example.com", cfg.JWT.AccessTokenTTL)
	require.NoError(t, err)

	return handler, token
}

func TestNewHandler(t *testing.T) {
	cfg := &config.Config{
		JWT: config.JWTConfig{
			SecretKey:       "test-secret",
			AccessTokenTTL:  15 * time.Minute,
			RefreshTokenTTL: 7 * 24 * time.Hour,
		},
	}

	authSvc := auth.NewService(cfg)
	handler := NewHandler(authSvc)

	assert.NotNil(t, handler)
	assert.IsType(t, &Handler{}, handler)
}

func TestHandler_CreateTask_ValidRequest(t *testing.T) {
	handler, token := setupTestHandler(t)
	app := fiber.New()

	// Add auth middleware
	app.Use(func(c *fiber.Ctx) error {
		// Mock auth middleware for testing
		c.Locals("user_id", uuid.MustParse("3484ec33-20f9-4993-a25f-f49f6f5dbe54"))
		c.Locals("user_email", "john.doe@example.com")
		return c.Next()
	})

	app.Post("/tasks", handler.CreateTask)

	req := task.CreateTaskRequest{
		Title: "Test Task",
	}

	reqBody, _ := json.Marshal(req)
	httpReq := httptest.NewRequest(http.MethodPost, "/tasks", bytes.NewBuffer(reqBody))
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+token)

	resp, err := app.Test(httpReq)

	require.NoError(t, err)
	assert.Equal(t, http.StatusCreated, resp.StatusCode)

	var response map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&response)
	require.NoError(t, err)

	assert.Equal(t, false, response["error"])
	assert.Equal(t, "Task created successfully", response["message"])
	assert.NotNil(t, response["data"])

	data := response["data"].(map[string]interface{})
	assert.Equal(t, "Test Task", data["title"])
	assert.Equal(t, "pending", data["status"])
}

func TestHandler_CreateTask_InvalidRequest(t *testing.T) {
	handler, token := setupTestHandler(t)
	app := fiber.New()

	// Add auth middleware
	app.Use(func(c *fiber.Ctx) error {
		c.Locals("user_id", uuid.MustParse("3484ec33-20f9-4993-a25f-f49f6f5dbe54"))
		c.Locals("user_email", "john.doe@example.com")
		return c.Next()
	})

	app.Post("/tasks", handler.CreateTask)

	req := task.CreateTaskRequest{
		Title: "", // Invalid title
	}

	reqBody, _ := json.Marshal(req)
	httpReq := httptest.NewRequest(http.MethodPost, "/tasks", bytes.NewBuffer(reqBody))
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+token)

	resp, err := app.Test(httpReq)

	require.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)

	var response map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&response)
	require.NoError(t, err)

	assert.Equal(t, true, response["error"])
	assert.Equal(t, "title is required", response["message"])
}

func TestHandler_GetTaskByID_ExistingTask(t *testing.T) {
	handler, token := setupTestHandler(t)
	app := fiber.New()

	// Add auth middleware
	app.Use(func(c *fiber.Ctx) error {
		c.Locals("user_id", uuid.MustParse("3484ec33-20f9-4993-a25f-f49f6f5dbe54"))
		c.Locals("user_email", "john.doe@example.com")
		return c.Next()
	})

	// First create a task
	app.Post("/tasks", handler.CreateTask)
	createReq := task.CreateTaskRequest{Title: "Test Task"}
	createReqBody, _ := json.Marshal(createReq)
	createHttpReq := httptest.NewRequest(http.MethodPost, "/tasks", bytes.NewBuffer(createReqBody))
	createHttpReq.Header.Set("Content-Type", "application/json")
	createHttpReq.Header.Set("Authorization", "Bearer "+token)

	createResp, err := app.Test(createHttpReq)
	require.NoError(t, err)
	require.Equal(t, http.StatusCreated, createResp.StatusCode)

	var createResponse map[string]interface{}
	err = json.NewDecoder(createResp.Body).Decode(&createResponse)
	require.NoError(t, err)

	taskID := createResponse["data"].(map[string]interface{})["id"].(string)

	// Now get the task
	app.Get("/tasks/:id", handler.GetTask)
	httpReq := httptest.NewRequest(http.MethodGet, "/tasks/"+taskID, nil)
	httpReq.Header.Set("Authorization", "Bearer "+token)

	resp, err := app.Test(httpReq)

	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var response map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&response)
	require.NoError(t, err)

	assert.Equal(t, false, response["error"])
	assert.Equal(t, "Task retrieved successfully", response["message"])
	assert.NotNil(t, response["data"])

	data := response["data"].(map[string]interface{})
	assert.Equal(t, "Test Task", data["title"])
	assert.Equal(t, taskID, data["id"])
}

func TestHandler_GetTaskByID_NonExistingTask(t *testing.T) {
	handler, token := setupTestHandler(t)
	app := fiber.New()

	// Add auth middleware
	app.Use(func(c *fiber.Ctx) error {
		c.Locals("user_id", uuid.MustParse("3484ec33-20f9-4993-a25f-f49f6f5dbe54"))
		c.Locals("user_email", "john.doe@example.com")
		return c.Next()
	})

	app.Get("/tasks/:id", handler.GetTask)
	nonExistingID := uuid.New().String()
	httpReq := httptest.NewRequest(http.MethodGet, "/tasks/"+nonExistingID, nil)
	httpReq.Header.Set("Authorization", "Bearer "+token)

	resp, err := app.Test(httpReq)

	require.NoError(t, err)
	assert.Equal(t, http.StatusNotFound, resp.StatusCode)

	var response map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&response)
	require.NoError(t, err)

	assert.Equal(t, true, response["error"])
	assert.Equal(t, "Task not found", response["message"])
}

func TestHandler_UpdateTask_ValidRequest(t *testing.T) {
	handler, token := setupTestHandler(t)
	app := fiber.New()

	// Add auth middleware
	app.Use(func(c *fiber.Ctx) error {
		c.Locals("user_id", uuid.MustParse("3484ec33-20f9-4993-a25f-f49f6f5dbe54"))
		c.Locals("user_email", "john.doe@example.com")
		return c.Next()
	})

	// First create a task
	app.Post("/tasks", handler.CreateTask)
	createReq := task.CreateTaskRequest{Title: "Original Title"}
	createReqBody, _ := json.Marshal(createReq)
	createHttpReq := httptest.NewRequest(http.MethodPost, "/tasks", bytes.NewBuffer(createReqBody))
	createHttpReq.Header.Set("Content-Type", "application/json")
	createHttpReq.Header.Set("Authorization", "Bearer "+token)

	createResp, err := app.Test(createHttpReq)
	require.NoError(t, err)
	require.Equal(t, http.StatusCreated, createResp.StatusCode)

	var createResponse map[string]interface{}
	err = json.NewDecoder(createResp.Body).Decode(&createResponse)
	require.NoError(t, err)

	taskID := createResponse["data"].(map[string]interface{})["id"].(string)

	// Now update the task
	app.Put("/tasks/:id", handler.UpdateTask)
	updateReq := task.UpdateTaskRequest{
		Title:  stringPtr("Updated Title"),
		Status: statusPtr(task.StatusInProgress),
	}

	updateReqBody, _ := json.Marshal(updateReq)
	httpReq := httptest.NewRequest(http.MethodPut, "/tasks/"+taskID, bytes.NewBuffer(updateReqBody))
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+token)

	resp, err := app.Test(httpReq)

	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var response map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&response)
	require.NoError(t, err)

	assert.Equal(t, false, response["error"])
	assert.Equal(t, "Task updated successfully", response["message"])
	assert.NotNil(t, response["data"])

	data := response["data"].(map[string]interface{})
	assert.Equal(t, "Updated Title", data["title"])
	assert.Equal(t, "in_progress", data["status"])
}

func TestHandler_DeleteTask_ExistingTask(t *testing.T) {
	handler, token := setupTestHandler(t)
	app := fiber.New()

	// Add auth middleware
	app.Use(func(c *fiber.Ctx) error {
		c.Locals("user_id", uuid.MustParse("3484ec33-20f9-4993-a25f-f49f6f5dbe54"))
		c.Locals("user_email", "john.doe@example.com")
		return c.Next()
	})

	// First create a task
	app.Post("/tasks", handler.CreateTask)
	createReq := task.CreateTaskRequest{Title: "Task to Delete"}
	createReqBody, _ := json.Marshal(createReq)
	createHttpReq := httptest.NewRequest(http.MethodPost, "/tasks", bytes.NewBuffer(createReqBody))
	createHttpReq.Header.Set("Content-Type", "application/json")
	createHttpReq.Header.Set("Authorization", "Bearer "+token)

	createResp, err := app.Test(createHttpReq)
	require.NoError(t, err)
	require.Equal(t, http.StatusCreated, createResp.StatusCode)

	var createResponse map[string]interface{}
	err = json.NewDecoder(createResp.Body).Decode(&createResponse)
	require.NoError(t, err)

	taskID := createResponse["data"].(map[string]interface{})["id"].(string)

	// Now delete the task
	app.Delete("/tasks/:id", handler.DeleteTask)
	httpReq := httptest.NewRequest(http.MethodDelete, "/tasks/"+taskID, nil)
	httpReq.Header.Set("Authorization", "Bearer "+token)

	resp, err := app.Test(httpReq)

	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var response map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&response)
	require.NoError(t, err)

	assert.Equal(t, false, response["error"])
	assert.Equal(t, "Task deleted successfully", response["message"])
}

func TestHandler_ListTasks_NoFilters(t *testing.T) {
	handler, token := setupTestHandler(t)
	app := fiber.New()

	// Add auth middleware
	app.Use(func(c *fiber.Ctx) error {
		c.Locals("user_id", uuid.MustParse("3484ec33-20f9-4993-a25f-f49f6f5dbe54"))
		c.Locals("user_email", "john.doe@example.com")
		return c.Next()
	})

	app.Get("/tasks", handler.ListTasks)
	httpReq := httptest.NewRequest(http.MethodGet, "/tasks", nil)
	httpReq.Header.Set("Authorization", "Bearer "+token)

	resp, err := app.Test(httpReq)

	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var response map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&response)
	require.NoError(t, err)

	assert.Equal(t, false, response["error"])
	assert.Equal(t, "Tasks retrieved successfully", response["message"])
	assert.NotNil(t, response["data"])
	assert.NotNil(t, response["meta"])
}

func TestHandler_ListTasks_WithFilters(t *testing.T) {
	handler, token := setupTestHandler(t)
	app := fiber.New()

	// Add auth middleware
	app.Use(func(c *fiber.Ctx) error {
		c.Locals("user_id", uuid.MustParse("3484ec33-20f9-4993-a25f-f49f6f5dbe54"))
		c.Locals("user_email", "john.doe@example.com")
		return c.Next()
	})

	app.Get("/tasks", handler.ListTasks)
	httpReq := httptest.NewRequest(http.MethodGet, "/tasks?status=pending&search=test&page=1&limit=10", nil)
	httpReq.Header.Set("Authorization", "Bearer "+token)

	resp, err := app.Test(httpReq)

	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var response map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&response)
	require.NoError(t, err)

	assert.Equal(t, false, response["error"])
	assert.Equal(t, "Tasks retrieved successfully", response["message"])
	assert.NotNil(t, response["data"])
	assert.NotNil(t, response["meta"])
}

// Helper functions for tests
func stringPtr(s string) *string {
	return &s
}

func statusPtr(s task.TaskStatus) *task.TaskStatus {
	return &s
}
