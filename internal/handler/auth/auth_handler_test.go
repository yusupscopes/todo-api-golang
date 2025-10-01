package auth

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"todo-api/internal/domain/auth"
	"todo-api/pkg/config"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewHandler(t *testing.T) {
	cfg := &config.Config{
		JWT: config.JWTConfig{
			SecretKey:       "test-secret",
			AccessTokenTTL:  15 * time.Minute,
			RefreshTokenTTL: 7 * 24 * time.Hour,
		},
	}

	handler := NewHandler(cfg)

	assert.NotNil(t, handler)
	assert.IsType(t, &Handler{}, handler)
}

func TestHandler_Login_ValidCredentials(t *testing.T) {
	cfg := &config.Config{
		JWT: config.JWTConfig{
			SecretKey:       "test-secret",
			AccessTokenTTL:  15 * time.Minute,
			RefreshTokenTTL: 7 * 24 * time.Hour,
		},
	}

	handler := NewHandler(cfg)
	app := fiber.New()

	app.Post("/login", handler.Login)

	req := auth.LoginRequest{
		Email:    "john.doe@example.com",
		Password: "password123",
	}

	reqBody, _ := json.Marshal(req)
	httpReq := httptest.NewRequest(http.MethodPost, "/login", bytes.NewBuffer(reqBody))
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(httpReq)

	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var response map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&response)
	require.NoError(t, err)

	assert.Equal(t, false, response["error"])
	assert.Equal(t, "Login successful", response["message"])
	assert.NotNil(t, response["data"])

	data := response["data"].(map[string]interface{})
	assert.NotEmpty(t, data["access_token"])
	assert.NotEmpty(t, data["refresh_token"])
	assert.Equal(t, "Bearer", data["token_type"])
}

func TestHandler_Login_InvalidCredentials(t *testing.T) {
	cfg := &config.Config{
		JWT: config.JWTConfig{
			SecretKey:       "test-secret",
			AccessTokenTTL:  15 * time.Minute,
			RefreshTokenTTL: 7 * 24 * time.Hour,
		},
	}

	handler := NewHandler(cfg)
	app := fiber.New()

	app.Post("/login", handler.Login)

	req := auth.LoginRequest{
		Email:    "nonexistent@example.com",
		Password: "wrongpassword",
	}

	reqBody, _ := json.Marshal(req)
	httpReq := httptest.NewRequest(http.MethodPost, "/login", bytes.NewBuffer(reqBody))
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(httpReq)

	require.NoError(t, err)
	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)

	var response map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&response)
	require.NoError(t, err)

	assert.Equal(t, true, response["error"])
	assert.Equal(t, "invalid email or password", response["message"])
}

func TestHandler_Login_InvalidRequest(t *testing.T) {
	cfg := &config.Config{
		JWT: config.JWTConfig{
			SecretKey:       "test-secret",
			AccessTokenTTL:  15 * time.Minute,
			RefreshTokenTTL: 7 * 24 * time.Hour,
		},
	}

	handler := NewHandler(cfg)
	app := fiber.New()

	app.Post("/login", handler.Login)

	// Send invalid JSON
	httpReq := httptest.NewRequest(http.MethodPost, "/login", bytes.NewBufferString("invalid json"))
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(httpReq)

	require.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)

	var response map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&response)
	require.NoError(t, err)

	assert.Equal(t, true, response["error"])
	assert.Equal(t, "Invalid request body", response["message"])
}

func TestHandler_Login_EmptyBody(t *testing.T) {
	cfg := &config.Config{
		JWT: config.JWTConfig{
			SecretKey:       "test-secret",
			AccessTokenTTL:  15 * time.Minute,
			RefreshTokenTTL: 7 * 24 * time.Hour,
		},
	}

	handler := NewHandler(cfg)
	app := fiber.New()

	app.Post("/login", handler.Login)

	httpReq := httptest.NewRequest(http.MethodPost, "/login", bytes.NewBufferString(""))
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(httpReq)

	require.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)

	var response map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&response)
	require.NoError(t, err)

	assert.Equal(t, true, response["error"])
	assert.Equal(t, "Invalid request body", response["message"])
}

func TestHandler_Login_ValidationErrors(t *testing.T) {
	cfg := &config.Config{
		JWT: config.JWTConfig{
			SecretKey:       "test-secret",
			AccessTokenTTL:  15 * time.Minute,
			RefreshTokenTTL: 7 * 24 * time.Hour,
		},
	}

	handler := NewHandler(cfg)
	app := fiber.New()

	app.Post("/login", handler.Login)

	tests := []struct {
		name     string
		request  auth.LoginRequest
		expected string
	}{
		{
			name: "empty email",
			request: auth.LoginRequest{
				Email:    "",
				Password: "password123",
			},
			expected: "email is required",
		},
		{
			name: "invalid email",
			request: auth.LoginRequest{
				Email:    "invalid-email",
				Password: "password123",
			},
			expected: "invalid email format",
		},
		{
			name: "empty password",
			request: auth.LoginRequest{
				Email:    "test@example.com",
				Password: "",
			},
			expected: "password is required",
		},
		{
			name: "short password",
			request: auth.LoginRequest{
				Email:    "test@example.com",
				Password: "1234567",
			},
			expected: "password must be at least 8 characters long",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reqBody, _ := json.Marshal(tt.request)
			httpReq := httptest.NewRequest(http.MethodPost, "/login", bytes.NewBuffer(reqBody))
			httpReq.Header.Set("Content-Type", "application/json")

			resp, err := app.Test(httpReq)

			require.NoError(t, err)
			assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)

			var response map[string]interface{}
			err = json.NewDecoder(resp.Body).Decode(&response)
			require.NoError(t, err)

			assert.Equal(t, true, response["error"])
			assert.Equal(t, tt.expected, response["message"])
		})
	}
}

func TestHandler_Login_AllMockUsers(t *testing.T) {
	cfg := &config.Config{
		JWT: config.JWTConfig{
			SecretKey:       "test-secret",
			AccessTokenTTL:  15 * time.Minute,
			RefreshTokenTTL: 7 * 24 * time.Hour,
		},
	}

	handler := NewHandler(cfg)
	app := fiber.New()

	app.Post("/login", handler.Login)

	users := []string{
		"john.doe@example.com",
		"jane.smith@example.com",
		"mike.wilson@example.com",
	}

	for _, email := range users {
		t.Run(email, func(t *testing.T) {
			req := auth.LoginRequest{
				Email:    email,
				Password: "password123",
			}

			reqBody, _ := json.Marshal(req)
			httpReq := httptest.NewRequest(http.MethodPost, "/login", bytes.NewBuffer(reqBody))
			httpReq.Header.Set("Content-Type", "application/json")

			resp, err := app.Test(httpReq)

			require.NoError(t, err)
			assert.Equal(t, http.StatusOK, resp.StatusCode)

			var response map[string]interface{}
			err = json.NewDecoder(resp.Body).Decode(&response)
			require.NoError(t, err)

			assert.Equal(t, false, response["error"])
			assert.Equal(t, "Login successful", response["message"])
		})
	}
}
