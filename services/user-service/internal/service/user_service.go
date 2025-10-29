// internal/service/user_service.go
package service

import (
	"context"
	"errors"
	"log"

	"github.com/datngth03/ecommerce-go-app/services/user-service/internal/metrics"
	"github.com/datngth03/ecommerce-go-app/services/user-service/internal/models"
	"github.com/datngth03/ecommerce-go-app/services/user-service/internal/repository"
	"github.com/datngth03/ecommerce-go-app/services/user-service/pkg/utils"
)

// UserServiceInterface defines the user service contract
type UserServiceInterface interface {
	// CRUD operations
	CreateUser(ctx context.Context, user *models.User) (*models.User, error)
	GetUserByID(ctx context.Context, id int64) (*models.User, error)
	GetUserByEmail(ctx context.Context, email string) (*models.User, error)
	UpdateUser(ctx context.Context, updateData *models.UserUpdateData) (*models.User, error)
	DeleteUser(ctx context.Context, id int64) error

	// Auth operations (will be used by auth_server later)
	ValidateUserCredentials(ctx context.Context, email, password string) (*models.User, error)
	ChangePassword(ctx context.Context, userID int64, oldPassword, newPassword string) error
	UpdatePasswordByEmail(ctx context.Context, email, newPassword string) error
}

// UserService implements the UserServiceInterface
type UserService struct {
	userRepo    repository.UserRepositoryInterface
	authService AuthServiceInterface
}

// NewUserService creates a new UserService instance
func NewUserService(userRepo repository.UserRepositoryInterface, authService AuthServiceInterface) UserServiceInterface {
	return &UserService{
		userRepo:    userRepo,
		authService: authService, // Thêm dòng này
	}
}

// =================================
// CRUD Operations Implementation
// =================================

// CreateUser creates a new user
func (s *UserService) CreateUser(ctx context.Context, user *models.User) (*models.User, error) {
	log.Printf("UserService: Creating user with email: %s", user.Email)

	// Check if user already exists
	existingUser, err := s.userRepo.GetByEmail(ctx, user.Email)
	if err == nil && existingUser != nil {
		return nil, errors.New("user already exists")
	}

	// Hash password
	hashedPassword, err := utils.HashPassword(user.Password)
	if err != nil {
		log.Printf("UserService: Failed to hash password: %v", err)
		return nil, errors.New("failed to process password")
	}
	user.Password = hashedPassword

	// Set defaults
	user.IsActive = true

	// Create user
	createdUser, err := s.userRepo.Create(ctx, user)
	if err != nil {
		log.Printf("UserService: Failed to create user: %v", err)
		return nil, errors.New("failed to create user")
	}

	// Record successful user registration
	metrics.RecordUserRegistration()

	log.Printf("UserService: User created successfully with ID: %d", createdUser.ID)
	return createdUser, nil
}

// GetUserByID retrieves a user by ID
func (s *UserService) GetUserByID(ctx context.Context, id int64) (*models.User, error) {
	log.Printf("UserService: Getting user by ID: %d", id)

	user, err := s.userRepo.GetByID(ctx, id)
	if err != nil {
		log.Printf("UserService: User not found with ID: %d", id)
		return nil, errors.New("user not found")
	}

	return user, nil
}

// GetUserByEmail retrieves a user by email
func (s *UserService) GetUserByEmail(ctx context.Context, email string) (*models.User, error) {
	log.Printf("UserService: Getting user by email: %s", email)

	user, err := s.userRepo.GetByEmail(ctx, email)
	if err != nil {
		log.Printf("UserService: User not found with email: %s", email)
		return nil, errors.New("user not found")
	}

	return user, nil
}

// UpdateUser updates user information
func (s *UserService) UpdateUser(ctx context.Context, updateData *models.UserUpdateData) (*models.User, error) {
	log.Printf("UserService: Updating user with ID: %d", updateData.ID)

	// Check if user exists
	// existingUser, err := s.userRepo.GetByID(ctx, updateData.ID)
	// if err != nil {
	// 	log.Printf("UserService: User not found with ID: %d", updateData.ID)
	// 	return nil, errors.New("user not found")
	// }

	// Update user
	updatedUser, err := s.userRepo.Update(ctx, updateData)
	if err != nil {
		log.Printf("UserService: Failed to update user: %v", err)
		return nil, errors.New("failed to update user")
	}

	log.Printf("UserService: User updated successfully with ID: %d", updatedUser.ID)
	return updatedUser, nil
}

// DeleteUser deletes a user
func (s *UserService) DeleteUser(ctx context.Context, id int64) error {
	log.Printf("UserService: Deleting user with ID: %d", id)

	// Check if user exists
	_, err := s.userRepo.GetByID(ctx, id)
	if err != nil {
		log.Printf("UserService: User not found with ID: %d", id)
		return errors.New("user not found")
	}

	// Delete user
	err = s.userRepo.Delete(ctx, id)
	if err != nil {
		log.Printf("UserService: Failed to delete user: %v", err)
		return errors.New("failed to delete user")
	}

	log.Printf("UserService: User deleted successfully with ID: %d", id)
	return nil
}

// =================================
// Auth Operations Implementation
// =================================

// ValidateUserCredentials validates user login credentials
func (s *UserService) ValidateUserCredentials(ctx context.Context, email, password string) (*models.User, error) {
	log.Printf("UserService: Validating credentials for email: %s", email)

	// Get user by email
	user, err := s.userRepo.GetByEmail(ctx, email)
	if err != nil {
		log.Printf("UserService: User not found with email: %s", email)
		return nil, errors.New("invalid credentials")
	}

	// Check if user is active
	if !user.IsActive {
		log.Printf("UserService: User is inactive: %s", email)
		return nil, errors.New("user account is inactive")
	}

	// Verify password
	if !utils.CheckPasswordHash(password, user.Password) {
		log.Printf("UserService: Invalid password for email: %s", email)
		return nil, errors.New("invalid credentials")
	}

	log.Printf("UserService: Credentials validated for user ID: %d", user.ID)
	return user, nil
}

// ChangePassword changes user password
func (s *UserService) ChangePassword(ctx context.Context, userID int64, oldPassword, newPassword string) error {
	log.Printf("UserService: Changing password for user ID: %d", userID)

	// Get user
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return errors.New("user not found")
	}

	// Verify old password
	if !utils.CheckPasswordHash(oldPassword, user.Password) {
		return errors.New("invalid old password")
	}

	// Hash new password
	hashedPassword, err := utils.HashPassword(newPassword)
	if err != nil {
		return errors.New("failed to process new password")
	}

	// Update password
	err = s.userRepo.UpdatePassword(ctx, userID, hashedPassword)
	if err != nil {
		log.Printf("UserService: Failed to update password: %v", err)
		return errors.New("failed to update password")
	}

	log.Printf("UserService: Password changed successfully for user ID: %d", userID)
	return nil
}

// UpdatePasswordByEmail updates password by email (for reset password)
func (s *UserService) UpdatePasswordByEmail(ctx context.Context, email, newPassword string) error {
	log.Printf("UserService: Updating password by email: %s", email)

	// Get user by email
	user, err := s.userRepo.GetByEmail(ctx, email)
	if err != nil {
		return errors.New("user not found")
	}

	// Hash new password
	hashedPassword, err := utils.HashPassword(newPassword)
	if err != nil {
		return errors.New("failed to process new password")
	}

	// Update password
	err = s.userRepo.UpdatePassword(ctx, user.ID, hashedPassword)
	if err != nil {
		log.Printf("UserService: Failed to update password: %v", err)
		return errors.New("failed to update password")
	}

	log.Printf("UserService: Password updated successfully for email: %s", email)
	return nil
}
