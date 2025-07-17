// internal/user/infrastructure/repository/postgres_user_repo.go
package repository

import (
	"context"
	"database/sql" // Thư viện chuẩn để tương tác với DB
	"errors"
	"fmt"
	"time"

	_ "github.com/lib/pq" // PostgreSQL driver

	"github.com/datngth03/ecommerce-go-app/internal/user/domain"
)

// PostgreSQLUserRepository implements the domain.UserRepository interface
// for PostgreSQL database operations.
// PostgreSQLUserRepository triển khai interface domain.UserRepository
// cho các thao tác cơ sở dữ liệu PostgreSQL.
type PostgreSQLUserRepository struct {
	db *sql.DB // Con trỏ đến đối tượng kết nối cơ sở dữ liệu
}

// NewPostgreSQLUserRepository creates a new instance of PostgreSQLUserRepository.
// NewPostgreSQLUserRepository tạo một thể hiện mới của PostgreSQLUserRepository.
func NewPostgreSQLUserRepository(db *sql.DB) *PostgreSQLUserRepository {
	return &PostgreSQLUserRepository{db: db}
}

// Save creates a new user or updates an existing one in PostgreSQL.
// Save tạo người dùng mới hoặc cập nhật người dùng hiện có trong PostgreSQL.
func (r *PostgreSQLUserRepository) Save(ctx context.Context, user *domain.User) error {
	// Kiểm tra xem người dùng đã tồn tại chưa để quyết định INSERT hay UPDATE
	existingUser, err := r.FindByID(ctx, user.ID)
	if err != nil && err.Error() != "user not found" {
		return fmt.Errorf("failed to check existing user: %w", err)
	}

	if existingUser != nil {
		// User exists, perform UPDATE
		query := `
                UPDATE users
                SET email = $1, password = $2, full_name = $3, phone_number = $4, address = $5, updated_at = $6
                WHERE id = $7`
		_, err = r.db.ExecContext(ctx, query,
			user.Email, user.Password, user.FullName, user.PhoneNumber, user.Address, time.Now(), user.ID)
		if err != nil {
			return fmt.Errorf("failed to update user: %w", err)
		}
		return nil
	}

	// User does not exist, perform INSERT
	query := `
            INSERT INTO users (id, email, password, full_name, phone_number, address, created_at, updated_at)
            VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`
	_, err = r.db.ExecContext(ctx, query,
		user.ID, user.Email, user.Password, user.FullName, user.PhoneNumber, user.Address, user.CreatedAt, user.UpdatedAt)
	if err != nil {
		// Handle unique constraint violation for email
		if err.Error() == `pq: duplicate key value violates unique constraint "idx_users_email"` ||
			err.Error() == `pq: duplicate key value violates unique constraint "users_email_key"` {
			return errors.New("user with this email already exists")
		}
		return fmt.Errorf("failed to save new user: %w", err)
	}
	return nil
}

// FindByID retrieves a user by their ID from PostgreSQL.
// FindByID lấy người dùng theo ID của họ từ PostgreSQL.
func (r *PostgreSQLUserRepository) FindByID(ctx context.Context, id string) (*domain.User, error) {
	user := &domain.User{}
	query := `SELECT id, email, password, full_name, phone_number, address, created_at, updated_at FROM users WHERE id = $1`
	row := r.db.QueryRowContext(ctx, query, id)

	err := row.Scan(&user.ID, &user.Email, &user.Password, &user.FullName,
		&user.PhoneNumber, &user.Address, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("user not found")
		}
		return nil, fmt.Errorf("failed to find user by ID: %w", err)
	}
	return user, nil
}

// FindByEmail retrieves a user by their email address from PostgreSQL.
// FindByEmail lấy người dùng theo địa chỉ email của họ từ PostgreSQL.
func (r *PostgreSQLUserRepository) FindByEmail(ctx context.Context, email string) (*domain.User, error) {
	user := &domain.User{}
	query := `SELECT id, email, password, full_name, phone_number, address, created_at, updated_at FROM users WHERE email = $1`
	row := r.db.QueryRowContext(ctx, query, email)

	err := row.Scan(&user.ID, &user.Email, &user.Password, &user.FullName,
		&user.PhoneNumber, &user.Address, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("user not found")
		}
		return nil, fmt.Errorf("failed to find user by email: %w", err)
	}
	return user, nil
}

// Delete removes a user from PostgreSQL by their ID.
// Delete xóa người dùng khỏi PostgreSQL theo ID của họ.
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
