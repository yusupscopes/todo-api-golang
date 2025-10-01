package auth

import (
	"testing"
	"time"

	"todo-api/internal/domain/auth"
	"todo-api/pkg/config"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewService(t *testing.T) {
	cfg := &config.Config{
		JWT: config.JWTConfig{
			SecretKey:       "test-secret",
			AccessTokenTTL:  15 * time.Minute,
			RefreshTokenTTL: 7 * 24 * time.Hour,
		},
	}

	service := NewService(cfg)

	assert.NotNil(t, service)
}

func TestService_Login_ValidCredentials(t *testing.T) {
	cfg := &config.Config{
		JWT: config.JWTConfig{
			SecretKey:       "test-secret",
			AccessTokenTTL:  15 * time.Minute,
			RefreshTokenTTL: 7 * 24 * time.Hour,
		},
	}

	service := NewService(cfg)

	req := &auth.LoginRequest{
		Email:    "john.doe@example.com",
		Password: "password123",
	}

	tokenResp, err := service.Login(req)

	require.NoError(t, err)
	assert.NotNil(t, tokenResp)
	assert.NotEmpty(t, tokenResp.AccessToken)
	assert.NotEmpty(t, tokenResp.RefreshToken)
	assert.Equal(t, "Bearer", tokenResp.TokenType)
	assert.Equal(t, int64(900), tokenResp.ExpiresIn) // 15 minutes in seconds
}

func TestService_Login_InvalidEmail(t *testing.T) {
	cfg := &config.Config{
		JWT: config.JWTConfig{
			SecretKey:       "test-secret",
			AccessTokenTTL:  15 * time.Minute,
			RefreshTokenTTL: 7 * 24 * time.Hour,
		},
	}

	service := NewService(cfg)

	req := &auth.LoginRequest{
		Email:    "nonexistent@example.com",
		Password: "password123",
	}

	tokenResp, err := service.Login(req)

	require.Error(t, err)
	assert.Nil(t, tokenResp)
	assert.Equal(t, "invalid email or password", err.Error())
}

func TestService_Login_InvalidPassword(t *testing.T) {
	cfg := &config.Config{
		JWT: config.JWTConfig{
			SecretKey:       "test-secret",
			AccessTokenTTL:  15 * time.Minute,
			RefreshTokenTTL: 7 * 24 * time.Hour,
		},
	}

	service := NewService(cfg)

	req := &auth.LoginRequest{
		Email:    "john.doe@example.com",
		Password: "wrongpassword",
	}

	tokenResp, err := service.Login(req)

	require.Error(t, err)
	assert.Nil(t, tokenResp)
	assert.Equal(t, "invalid email or password", err.Error())
}

func TestService_Login_InvalidRequest(t *testing.T) {
	cfg := &config.Config{
		JWT: config.JWTConfig{
			SecretKey:       "test-secret",
			AccessTokenTTL:  15 * time.Minute,
			RefreshTokenTTL: 7 * 24 * time.Hour,
		},
	}

	service := NewService(cfg)

	req := &auth.LoginRequest{
		Email:    "", // Invalid email
		Password: "password123",
	}

	tokenResp, err := service.Login(req)

	require.Error(t, err)
	assert.Nil(t, tokenResp)
	assert.Equal(t, "email is required", err.Error())
}

func TestService_ValidateToken_ValidToken(t *testing.T) {
	cfg := &config.Config{
		JWT: config.JWTConfig{
			SecretKey:       "test-secret",
			AccessTokenTTL:  15 * time.Minute,
			RefreshTokenTTL: 7 * 24 * time.Hour,
		},
	}

	service := NewService(cfg)

	// First login to get a valid token
	req := &auth.LoginRequest{
		Email:    "john.doe@example.com",
		Password: "password123",
	}

	tokenResp, err := service.Login(req)
	require.NoError(t, err)

	// Validate the token
	claims, err := service.ValidateToken(tokenResp.AccessToken)

	require.NoError(t, err)
	assert.NotNil(t, claims)
	assert.Equal(t, uuid.MustParse("3484ec33-20f9-4993-a25f-f49f6f5dbe54"), claims.UserID)
	assert.Equal(t, "john.doe@example.com", claims.Email)
}

func TestService_ValidateToken_InvalidToken(t *testing.T) {
	cfg := &config.Config{
		JWT: config.JWTConfig{
			SecretKey:       "test-secret",
			AccessTokenTTL:  15 * time.Minute,
			RefreshTokenTTL: 7 * 24 * time.Hour,
		},
	}

	service := NewService(cfg)

	claims, err := service.ValidateToken("invalid-token")

	require.Error(t, err)
	assert.Nil(t, claims)
}

func TestService_GetUserByEmail_ExistingUser(t *testing.T) {
	cfg := &config.Config{
		JWT: config.JWTConfig{
			SecretKey:       "test-secret",
			AccessTokenTTL:  15 * time.Minute,
			RefreshTokenTTL: 7 * 24 * time.Hour,
		},
	}

	service := NewService(cfg)

	user, err := service.GetUserByEmail("john.doe@example.com")

	require.NoError(t, err)
	assert.NotNil(t, user)
	assert.Equal(t, "john.doe@example.com", user.Email)
	assert.Equal(t, "password123", user.Password)
	assert.Equal(t, uuid.MustParse("3484ec33-20f9-4993-a25f-f49f6f5dbe54"), user.ID)
}

func TestService_GetUserByEmail_NonExistingUser(t *testing.T) {
	cfg := &config.Config{
		JWT: config.JWTConfig{
			SecretKey:       "test-secret",
			AccessTokenTTL:  15 * time.Minute,
			RefreshTokenTTL: 7 * 24 * time.Hour,
		},
	}

	service := NewService(cfg)

	user, err := service.GetUserByEmail("nonexistent@example.com")

	require.Error(t, err)
	assert.Nil(t, user)
	assert.Equal(t, "user not found", err.Error())
}

func TestService_AllMockUsers(t *testing.T) {
	cfg := &config.Config{
		JWT: config.JWTConfig{
			SecretKey:       "test-secret",
			AccessTokenTTL:  15 * time.Minute,
			RefreshTokenTTL: 7 * 24 * time.Hour,
		},
	}

	service := NewService(cfg)

	// Test all mock users
	mockUsers := []struct {
		email    string
		password string
		userID   string
	}{
		{"john.doe@example.com", "password123", "3484ec33-20f9-4993-a25f-f49f6f5dbe54"},
		{"jane.smith@example.com", "password123", "550e8400-e29b-41d4-a716-446655440002"},
		{"mike.wilson@example.com", "password123", "550e8400-e29b-41d4-a716-446655440003"},
	}

	for _, mockUser := range mockUsers {
		t.Run(mockUser.email, func(t *testing.T) {
			// Test login
			req := &auth.LoginRequest{
				Email:    mockUser.email,
				Password: mockUser.password,
			}

			tokenResp, err := service.Login(req)
			require.NoError(t, err)
			assert.NotEmpty(t, tokenResp.AccessToken)

			// Test get user by email
			user, err := service.GetUserByEmail(mockUser.email)
			require.NoError(t, err)
			assert.Equal(t, mockUser.email, user.Email)
			assert.Equal(t, uuid.MustParse(mockUser.userID), user.ID)
		})
	}
}

func TestService_Login_AllUsers(t *testing.T) {
	cfg := &config.Config{
		JWT: config.JWTConfig{
			SecretKey:       "test-secret",
			AccessTokenTTL:  15 * time.Minute,
			RefreshTokenTTL: 7 * 24 * time.Hour,
		},
	}

	service := NewService(cfg)

	// Test login for all users
	users := []string{
		"john.doe@example.com",
		"jane.smith@example.com",
		"mike.wilson@example.com",
	}

	for _, email := range users {
		t.Run(email, func(t *testing.T) {
			req := &auth.LoginRequest{
				Email:    email,
				Password: "password123",
			}

			tokenResp, err := service.Login(req)
			require.NoError(t, err)
			assert.NotNil(t, tokenResp)
			assert.NotEmpty(t, tokenResp.AccessToken)
			assert.NotEmpty(t, tokenResp.RefreshToken)
		})
	}
}
