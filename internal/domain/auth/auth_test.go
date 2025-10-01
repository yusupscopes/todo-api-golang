package auth

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewUser(t *testing.T) {
	email := "test@example.com"
	password := "password123"

	user := NewUser(email, password)

	assert.NotNil(t, user)
	assert.Equal(t, email, user.Email)
	assert.Equal(t, password, user.Password)
	assert.NotEqual(t, uuid.Nil, user.ID)
	assert.False(t, user.CreatedAt.IsZero())
	assert.False(t, user.UpdatedAt.IsZero())
}

func TestLoginRequest_Validate(t *testing.T) {
	tests := []struct {
		name    string
		request LoginRequest
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid request",
			request: LoginRequest{
				Email:    "test@example.com",
				Password: "password123",
			},
			wantErr: false,
		},
		{
			name: "empty email",
			request: LoginRequest{
				Email:    "",
				Password: "password123",
			},
			wantErr: true,
			errMsg:  "email is required",
		},
		{
			name: "whitespace email",
			request: LoginRequest{
				Email:    "   ",
				Password: "password123",
			},
			wantErr: true,
			errMsg:  "email is required",
		},
		{
			name: "invalid email format",
			request: LoginRequest{
				Email:    "invalid-email",
				Password: "password123",
			},
			wantErr: true,
			errMsg:  "invalid email format",
		},
		{
			name: "email without @",
			request: LoginRequest{
				Email:    "test.example.com",
				Password: "password123",
			},
			wantErr: true,
			errMsg:  "invalid email format",
		},
		{
			name: "email without domain",
			request: LoginRequest{
				Email:    "test@",
				Password: "password123",
			},
			wantErr: true,
			errMsg:  "invalid email format",
		},
		{
			name: "empty password",
			request: LoginRequest{
				Email:    "test@example.com",
				Password: "",
			},
			wantErr: true,
			errMsg:  "password is required",
		},
		{
			name: "whitespace password",
			request: LoginRequest{
				Email:    "test@example.com",
				Password: "   ",
			},
			wantErr: true,
			errMsg:  "password is required",
		},
		{
			name: "password too short",
			request: LoginRequest{
				Email:    "test@example.com",
				Password: "1234567",
			},
			wantErr: true,
			errMsg:  "password must be at least 8 characters long",
		},
		{
			name: "password exactly 8 characters",
			request: LoginRequest{
				Email:    "test@example.com",
				Password: "12345678",
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

func TestIsValidEmail(t *testing.T) {
	tests := []struct {
		name  string
		email string
		valid bool
	}{
		{"valid email", "test@example.com", true},
		{"valid email with subdomain", "user@mail.example.com", true},
		{"email without @", "test.example.com", false},
		{"email without domain", "test@", false},
		{"email without @ and domain", "test", false},
		{"empty email", "", false},
		{"email with multiple @", "test@@example.com", true}, // Basic validation allows this
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isValidEmail(tt.email)
			assert.Equal(t, tt.valid, result)
		})
	}
}

func TestTokenResponse(t *testing.T) {
	tokenResp := &TokenResponse{
		AccessToken:  "access_token_123",
		RefreshToken: "refresh_token_456",
		TokenType:    "Bearer",
		ExpiresIn:    900,
	}

	assert.Equal(t, "access_token_123", tokenResp.AccessToken)
	assert.Equal(t, "refresh_token_456", tokenResp.RefreshToken)
	assert.Equal(t, "Bearer", tokenResp.TokenType)
	assert.Equal(t, int64(900), tokenResp.ExpiresIn)
}

func TestUser_JSONSerialization(t *testing.T) {
	user := &User{
		ID:        uuid.New(),
		Email:     "test@example.com",
		Password:  "password123",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Test that password is not included in JSON
	jsonData, err := json.Marshal(user)
	require.NoError(t, err)

	jsonStr := string(jsonData)
	assert.Contains(t, jsonStr, user.Email)
	assert.Contains(t, jsonStr, user.ID.String())
	assert.NotContains(t, jsonStr, user.Password)
}
