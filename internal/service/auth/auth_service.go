package auth

import (
	"errors"
	"time"

	"todo-api/internal/domain/auth"
	"todo-api/pkg/config"
	"todo-api/pkg/utils"

	"github.com/google/uuid"
)

// Service defines the authentication service interface
type Service interface {
	Login(req *auth.LoginRequest) (*auth.TokenResponse, error)
	ValidateToken(token string) (*utils.JWTClaims, error)
	GetUserByEmail(email string) (*auth.User, error)
}

// service implements the authentication service
type service struct {
	config *config.Config
	users  map[string]*auth.User // Mock user storage
}

// NewService creates a new authentication service
func NewService(cfg *config.Config) Service {
	// Initialize mock users
	users := make(map[string]*auth.User)

	// Create some mock users with fixed UUIDs
	user1 := &auth.User{
		ID:        uuid.MustParse("3484ec33-20f9-4993-a25f-f49f6f5dbe54"),
		Email:     "john.doe@example.com",
		Password:  "password123",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	users["john.doe@example.com"] = user1

	user2 := &auth.User{
		ID:        uuid.MustParse("550e8400-e29b-41d4-a716-446655440002"),
		Email:     "jane.smith@example.com",
		Password:  "password123",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	users["jane.smith@example.com"] = user2

	user3 := &auth.User{
		ID:        uuid.MustParse("550e8400-e29b-41d4-a716-446655440003"),
		Email:     "mike.wilson@example.com",
		Password:  "password123",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	users["mike.wilson@example.com"] = user3

	return &service{
		config: cfg,
		users:  users,
	}
}

// Login authenticates a user and returns tokens
func (s *service) Login(req *auth.LoginRequest) (*auth.TokenResponse, error) {
	// Validate request
	if err := req.Validate(); err != nil {
		return nil, err
	}

	// Find user by email
	user, exists := s.users[req.Email]
	if !exists {
		return nil, errors.New("invalid email or password")
	}

	// Check password (in a real app, you'd hash and compare)
	if user.Password != req.Password {
		return nil, errors.New("invalid email or password")
	}

	// Generate access token
	accessToken, err := utils.GenerateToken(
		s.config.JWT.SecretKey,
		user.ID,
		user.Email,
		s.config.JWT.AccessTokenTTL,
	)
	if err != nil {
		return nil, errors.New("failed to generate access token")
	}

	// Generate refresh token
	refreshToken, err := utils.GenerateToken(
		s.config.JWT.SecretKey,
		user.ID,
		user.Email,
		s.config.JWT.RefreshTokenTTL,
	)
	if err != nil {
		return nil, errors.New("failed to generate refresh token")
	}

	return &auth.TokenResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    int64(s.config.JWT.AccessTokenTTL.Seconds()),
	}, nil
}

// ValidateToken validates a JWT token and returns the claims
func (s *service) ValidateToken(token string) (*utils.JWTClaims, error) {
	return utils.ValidateToken(token, s.config.JWT.SecretKey)
}

// GetUserByEmail retrieves a user by email
func (s *service) GetUserByEmail(email string) (*auth.User, error) {
	user, exists := s.users[email]
	if !exists {
		return nil, errors.New("user not found")
	}
	return user, nil
}
