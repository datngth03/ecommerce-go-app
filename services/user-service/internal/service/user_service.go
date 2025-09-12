package service

import (
	"context"
	"fmt"

	"github.com/ecommerce/services/user-service/internal/models"
	"github.com/ecommerce/services/user-service/internal/repository"
	"github.com/ecommerce/services/user-service/pkg/utils"
	"github.com/ecommerce/services/user-service/pkg/validator"
)

type UserService interface {
	CreateUser(ctx context.Context, req *models.CreateUserRequest) (*models.User, error)
	GetUserByID(ctx context.Context, id int64) (*models.User, error)
	UpdateUser(ctx context.Context, id int64, req *models.UpdateUserRequest) (*models.User, error)
	DeleteUser(ctx context.Context, id int64) error
	ListUsers(ctx context.Context, page, limit int, role string) ([]models.User, int, error)
	Login(ctx context.Context, req *models.LoginRequest) (*models.LoginResponse, error)
	ValidateToken(ctx context.Context, token string) (*models.User, error)
}

type userService struct {
	userRepo  repository.UserRepository
	validator *validator.UserValidator
}

func NewUserService(userRepo repository.UserRepository, validator *validator.UserValidator) UserService {
	return &userService{
		userRepo:  userRepo,
		validator: validator,
	}
}

func (s *userService) CreateUser(ctx context.Context, req *models.CreateUserRequest) (*models.User, error) {
	// Validate request
	if err := s.validator.ValidateCreateUser(req); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	// Check if email already exists
	exists, err := s.userRepo.EmailExists(ctx, req.Email, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to check email existence: %w", err)
	}
	if exists {
		return nil, fmt.Errorf("email already exists")
	}

	// Hash password
	hashedPassword, err := utils.HashPassword(req.Password)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	// Create user model
	user := &models.User{
		Email:    req.Email,
		Password: hashedPassword,
		Name:     req.Name,
		Phone:    req.Phone,
		Role:     req.Role,
	}

	// Save to database
	createdUser, err := s.userRepo.Create(ctx, user)
	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	return createdUser, nil
}

func (s *userService) GetUserByID(ctx context.Context, id int64) (*models.User, error) {
	if id <= 0 {
		return nil, fmt.Errorf("invalid user id")
	}

	user, err := s.userRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return user, nil
}

func (s *userService) UpdateUser(ctx context.Context, id int64, req *models.UpdateUserRequest) (*models.User, error) {
	if id <= 0 {
		return nil, fmt.Errorf("invalid user id")
	}

	// Validate request
	if err := s.validator.ValidateUpdateUser(req); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	// Check if user exists
	_, err := s.userRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}

	// Update user
	updatedUser, err := s.userRepo.Update(ctx, id, req)
	if err != nil {
		return nil, fmt.Errorf("failed to update user: %w", err)
	}

	return updatedUser, nil
}

func (s *userService) DeleteUser(ctx context.Context, id int64) error {
	if id <= 0 {
		return fmt.Errorf("invalid user id")
	}

	// Check if user exists
	_, err := s.userRepo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("user not found: %w", err)
	}

	// Delete user
	err = s.userRepo.Delete(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}

	return nil
}

func (s *userService) ListUsers(ctx context.Context, page, limit int, role string) ([]models.User, int, error) {
	if page <= 0 {
		page = 1
	}
	if limit <= 0 || limit > 100 {
		limit = 10
	}

	users, total, err := s.userRepo.List(ctx, page, limit, role)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list users: %w", err)
	}

	return users, total, nil
}

func (s *userService) Login(ctx context.Context, req *models.LoginRequest) (*models.LoginResponse, error) {
	// Validate request
	if err := s.validator.ValidateLogin(req); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	// Get user by email
	user, err := s.userRepo.GetByEmail(ctx, req.Email)
	if err != nil {
		return nil, fmt.Errorf("invalid credentials")
	}

	// Verify password
	if !utils.CheckPassword(req.Password, user.Password) {
		return nil, fmt.Errorf("invalid credentials")
	}

	// Generate JWT token
	token, err := utils.GenerateToken(user.ID, user.Email, user.Role)
	if err != nil {
		return nil, fmt.Errorf("failed to generate token: %w", err)
	}

	// Clear password from response
	user.Password = ""

	return &models.LoginResponse{
		Token: token,
		User:  user,
	}, nil
}

func (s *userService) ValidateToken(ctx context.Context, token string) (*models.User, error) {
	claims, err := utils.ValidateToken(token)
	if err != nil {
		return nil, fmt.Errorf("invalid token: %w", err)
	}

	userID := int64(claims["user_id"].(float64))
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}

	return user, nil
}
