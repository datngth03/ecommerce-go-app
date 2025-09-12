package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"

	"github.com/ecommerce/services/user-service/internal/models"
	"github.com/ecommerce/services/user-service/internal/repository"
)

type userRepository struct {
	db *sqlx.DB
}

func NewUserRepository(db *sqlx.DB) repository.UserRepository {
	return &userRepository{db: db}
}

func (r *userRepository) Create(ctx context.Context, user *models.User) (*models.User, error) {
	query := `
		INSERT INTO users (email, password, name, phone, role, created_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, email, name, phone, role, created_at`

	var createdUser models.User
	err := r.db.GetContext(ctx, &createdUser, query,
		user.Email,
		user.Password,
		user.Name,
		user.Phone,
		user.Role,
		time.Now(),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	return &createdUser, nil
}

func (r *userRepository) GetByID(ctx context.Context, id int64) (*models.User, error) {
	query := `SELECT id, email, name, phone, role, created_at FROM users WHERE id = $1`

	var user models.User
	err := r.db.GetContext(ctx, &user, query, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("user with id %d not found", id)
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return &user, nil
}

func (r *userRepository) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	query := `SELECT id, email, password, name, phone, role, created_at FROM users WHERE email = $1`

	var user models.User
	err := r.db.GetContext(ctx, &user, query, email)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("user with email %s not found", email)
		}
		return nil, fmt.Errorf("failed to get user by email: %w", err)
	}

	return &user, nil
}

func (r *userRepository) Update(ctx context.Context, id int64, req *models.UpdateUserRequest) (*models.User, error) {
	query := `
		UPDATE users 
		SET name = $1, phone = $2, role = $3
		WHERE id = $4
		RETURNING id, email, name, phone, role, created_at`

	var user models.User
	err := r.db.GetContext(ctx, &user, query, req.Name, req.Phone, req.Role, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("user with id %d not found", id)
		}
		return nil, fmt.Errorf("failed to update user: %w", err)
	}

	return &user, nil
}

func (r *userRepository) Delete(ctx context.Context, id int64) error {
	query := `DELETE FROM users WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("user with id %d not found", id)
	}

	return nil
}

func (r *userRepository) List(ctx context.Context, page, limit int, role string) ([]models.User, int, error) {
	offset := (page - 1) * limit

	// Build query with optional role filter
	whereClause := ""
	countWhereClause := ""
	args := []interface{}{limit, offset}
	countArgs := []interface{}{}

	if role != "" {
		whereClause = " WHERE role = $3"
		countWhereClause = " WHERE role = $1"
		args = append(args, role)
		countArgs = append(countArgs, role)
	}

	// Get users
	query := fmt.Sprintf(`
		SELECT id, email, name, phone, role, created_at 
		FROM users%s 
		ORDER BY created_at DESC 
		LIMIT $1 OFFSET $2`, whereClause)

	var users []models.User
	err := r.db.SelectContext(ctx, &users, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list users: %w", err)
	}

	// Get total count
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM users%s", countWhereClause)
	var total int
	err = r.db.GetContext(ctx, &total, countQuery, countArgs...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count users: %w", err)
	}

	return users, total, nil
}

func (r *userRepository) EmailExists(ctx context.Context, email string, excludeID int64) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM users WHERE email = $1 AND id != $2)`

	var exists bool
	err := r.db.GetContext(ctx, &exists, query, email, excludeID)
	if err != nil {
		return false, fmt.Errorf("failed to check email existence: %w", err)
	}

	return exists, nil
}
