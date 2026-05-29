package auth

import (
	"errors"
	"fmt"
	"opsight-backend/internal/database"
	"opsight-backend/internal/model"
	"opsight-backend/pkg/logger"

	"golang.org/x/crypto/bcrypt"
)

// AuthService wraps authentication and user management logic.
type AuthService struct{}

// NewAuthService creates a new AuthService.
func NewAuthService() *AuthService {
	return &AuthService{}
}

// LoginResponse is returned on successful login.
type LoginResponse struct {
	Token string       `json:"token"`
	User  UserResponse `json:"user"`
}

// UserResponse contains the public user fields (password excluded by model tag).
type UserResponse struct {
	ID    uint   `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
	Role  string `json:"role"`
	Team  string `json:"team"`
}

func userToResponse(u *model.User) UserResponse {
	return UserResponse{
		ID:    u.ID,
		Name:  u.Name,
		Email: u.Email,
		Role:  u.Role,
		Team:  u.Team,
	}
}

// Login validates credentials and returns a JWT token with user info.
func (s *AuthService) Login(email, password string) (*LoginResponse, error) {
	var user model.User
	if err := database.DB.Where("LOWER(email) = LOWER(?)", email).First(&user).Error; err != nil {
		logger.Warn().Str("email", email).Msg("Login failed: user not found")
		return nil, errors.New("invalid email or password")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		logger.Warn().Str("email", email).Msg("Login failed: password mismatch")
		return nil, errors.New("invalid email or password")
	}

	token, err := GenerateToken(user.ID, user.Email, user.Role)
	if err != nil {
		logger.Error().Err(err).Uint("user_id", user.ID).Msg("Failed to generate JWT token")
		return nil, fmt.Errorf("failed to generate token")
	}

	logger.Info().Uint("user_id", user.ID).Str("email", user.Email).Msg("User logged in")
	return &LoginResponse{
		Token: token,
		User:  userToResponse(&user),
	}, nil
}

// Register creates a new user with a bcrypt-hashed password.
func (s *AuthService) Register(name, email, password, role string) (*UserResponse, error) {
	// Check for existing user
	var existing model.User
	if err := database.DB.Where("LOWER(email) = LOWER(?)", email).First(&existing).Error; err == nil {
		return nil, errors.New("user with this email already exists")
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		logger.Error().Err(err).Msg("Failed to hash password during registration")
		return nil, fmt.Errorf("internal error")
	}

	user := model.User{
		Name:     name,
		Email:    email,
		Password: string(hash),
		Role:     role,
		Team:     "",
	}
	if err := database.DB.Create(&user).Error; err != nil {
		logger.Error().Err(err).Str("email", email).Msg("Failed to create user")
		return nil, fmt.Errorf("failed to create user")
	}

	logger.Info().Uint("user_id", user.ID).Str("email", user.Email).Msg("User registered")
	resp := userToResponse(&user)
	return &resp, nil
}

// GetUserByID returns a user by their primary key.
func (s *AuthService) GetUserByID(id uint) (*UserResponse, error) {
	var user model.User
	if err := database.DB.First(&user, id).Error; err != nil {
		return nil, errors.New("user not found")
	}
	resp := userToResponse(&user)
	return &resp, nil
}

// ListUsers returns all users in the system.
func (s *AuthService) ListUsers() ([]UserResponse, error) {
	var users []model.User
	if err := database.DB.Find(&users).Error; err != nil {
		return nil, fmt.Errorf("failed to list users: %w", err)
	}
	result := make([]UserResponse, len(users))
	for i, u := range users {
		result[i] = userToResponse(&u)
	}
	return result, nil
}
