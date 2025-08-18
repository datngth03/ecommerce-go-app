// internal/user/infrastructure/repository/postgres_user_repo.go
package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	_ "github.com/lib/pq"

	"github.com/datngth03/ecommerce-go-app/internal/user/domain"
)

type PostgreSQLUserRepository struct {
	db *sql.DB
}

func NewPostgreSQLUserRepository(db *sql.DB) *PostgreSQLUserRepository {
	return &PostgreSQLUserRepository{db: db}
}

func (r *PostgreSQLUserRepository) Save(ctx context.Context, user *domain.User) error {
	// Kiểm tra user có tồn tại chưa
	existingUser, err := r.FindByID(ctx, user.ID)
	if err != nil && err.Error() != "user not found" {
		return fmt.Errorf("failed to check existing user: %w", err)
	}

	if existingUser != nil {
		// UPDATE
		query := `
			UPDATE users
			SET email = $1, password = $2, full_name = $3, phone_number = $4, updated_at = $5
			WHERE id = $6`
		_, err = r.db.ExecContext(ctx, query,
			user.Email, user.Password, user.FullName, user.PhoneNumber, time.Now(), user.ID)
		if err != nil {
			return fmt.Errorf("failed to update user: %w", err)
		}
		return nil
	}

	// INSERT
	query := `
		INSERT INTO users (id, email, password, full_name, phone_number, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)`
	_, err = r.db.ExecContext(ctx, query,
		user.ID, user.Email, user.Password, user.FullName, user.PhoneNumber, user.CreatedAt, user.UpdatedAt)
	if err != nil {
		if err.Error() == `pq: duplicate key value violates unique constraint "idx_users_email"` ||
			err.Error() == `pq: duplicate key value violates unique constraint "users_email_key"` {
			return errors.New("user with this email already exists")
		}
		return fmt.Errorf("failed to save new user: %w", err)
	}
	return nil
}

func (r *PostgreSQLUserRepository) FindByID(ctx context.Context, id string) (*domain.User, error) {
	user := &domain.User{}
	query := `SELECT id, email, password, full_name, phone_number, created_at, updated_at FROM users WHERE id = $1`
	row := r.db.QueryRowContext(ctx, query, id)

	err := row.Scan(&user.ID, &user.Email, &user.Password, &user.FullName,
		&user.PhoneNumber, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("user not found")
		}
		return nil, fmt.Errorf("failed to find user by ID: %w", err)
	}
	return user, nil
}

func (r *PostgreSQLUserRepository) FindByEmail(ctx context.Context, email string) (*domain.User, error) {
	user := &domain.User{}
	query := `SELECT id, email, password, full_name, phone_number, created_at, updated_at FROM users WHERE email = $1`
	row := r.db.QueryRowContext(ctx, query, email)

	err := row.Scan(&user.ID, &user.Email, &user.Password, &user.FullName,
		&user.PhoneNumber, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("user not found")
		}
		return nil, fmt.Errorf("failed to find user by email: %w", err)
	}
	return user, nil
}

func (r *PostgreSQLUserRepository) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM users WHERE id = $1`
	res, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return errors.New("user not found")
	}
	return nil
}
