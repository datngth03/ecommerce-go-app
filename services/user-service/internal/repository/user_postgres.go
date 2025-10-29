// internal/repository/user_repository.go
package repository

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/datngth03/ecommerce-go-app/services/user-service/internal/metrics"
	"github.com/datngth03/ecommerce-go-app/services/user-service/internal/models"
)

var ErrNotFound = errors.New("record not found")

type sqlUserRepository struct {
	db *sql.DB
}

func NewSQLUserRepository(db *sql.DB) UserRepositoryInterface {
	return &sqlUserRepository{
		db: db,
	}
}

func (r *sqlUserRepository) Create(ctx context.Context, user *models.User) (*models.User, error) {
	start := time.Now()
	defer func() {
		metrics.RecordDatabaseQuery("INSERT", "users", time.Since(start))
	}()

	query := `
		INSERT INTO users (email, password_hash, name, phone, is_active)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, created_at, updated_at`

	err := r.db.QueryRowContext(
		ctx, query,
		user.Email, user.Password, user.Name, user.Phone, user.IsActive, // Dùng user.IsActive
	).Scan(&user.ID, &user.CreatedAt, &user.UpdatedAt)

	if err != nil {
		return nil, err
	}
	return user, nil
}

func (r *sqlUserRepository) GetByID(ctx context.Context, id int64) (*models.User, error) {
	start := time.Now()
	defer func() {
		metrics.RecordDatabaseQuery("SELECT", "users", time.Since(start))
	}()

	var user models.User
	query := `
		SELECT id, email, password_hash, name, phone, is_active, created_at, updated_at
		FROM users
		WHERE id = $1`

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&user.ID, &user.Email, &user.Password, &user.Name,
		&user.Phone, &user.IsActive, &user.CreatedAt, &user.UpdatedAt, // Dùng user.IsActive
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &user, nil
}

func (r *sqlUserRepository) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	start := time.Now()
	defer func() {
		metrics.RecordDatabaseQuery("SELECT", "users", time.Since(start))
	}()

	var user models.User
	query := `
		SELECT id, email, password_hash, name, phone, is_active, created_at, updated_at
		FROM users
		WHERE email = $1`

	err := r.db.QueryRowContext(ctx, query, email).Scan(
		&user.ID, &user.Email, &user.Password, &user.Name,
		&user.Phone, &user.IsActive, &user.CreatedAt, &user.UpdatedAt, // Dùng user.IsActive
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &user, nil
}

func (r *sqlUserRepository) Update(ctx context.Context, updateData *models.UserUpdateData) (*models.User, error) {
	start := time.Now()
	defer func() {
		metrics.RecordDatabaseQuery("UPDATE", "users", time.Since(start))
	}()

	query := `
		UPDATE users
		SET name = $1, phone = $2, updated_at = NOW()
		WHERE id = $3
		RETURNING id, email, password_hash, name, phone, is_active, created_at, updated_at`

	var user models.User
	err := r.db.QueryRowContext(ctx, query, updateData.Name, updateData.Phone, updateData.ID).Scan(
		&user.ID, &user.Email, &user.Password, &user.Name,
		&user.Phone, &user.IsActive, &user.CreatedAt, &user.UpdatedAt, // Dùng user.IsActive
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &user, nil
}

// Các hàm Delete, UpdatePassword, ExistsByEmail không cần thay đổi
// vì chúng không truy vấn hay chỉnh sửa cột is_active.
// ... (giữ nguyên các hàm còn lại)

func (r *sqlUserRepository) Delete(ctx context.Context, id int64) error {
	start := time.Now()
	defer func() {
		metrics.RecordDatabaseQuery("UPDATE", "users", time.Since(start))
	}()

	query := "UPDATE users SET is_active = FALSE, updated_at = NOW() WHERE id = $1"
	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *sqlUserRepository) UpdatePassword(ctx context.Context, userID int64, hashedPassword string) error {
	start := time.Now()
	defer func() {
		metrics.RecordDatabaseQuery("UPDATE", "users", time.Since(start))
	}()

	query := "UPDATE users SET password_hash = $1, updated_at = NOW() WHERE id = $2"
	result, err := r.db.ExecContext(ctx, query, hashedPassword, userID)
	if err != nil {
		return err
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *sqlUserRepository) ExistsByEmail(ctx context.Context, email string) (bool, error) {
	start := time.Now()
	defer func() {
		metrics.RecordDatabaseQuery("SELECT", "users", time.Since(start))
	}()

	var exists bool
	query := "SELECT EXISTS(SELECT 1 FROM users WHERE email = $1 AND is_active = TRUE)"
	err := r.db.QueryRowContext(ctx, query, email).Scan(&exists)
	if err != nil {
		return false, err
	}
	return exists, nil
}
