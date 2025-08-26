// internal/user/domain/repository.go
package domain

import (
	"context"
	"errors"
)

var (
	ErrProductNotFound     = errors.New("product not found")
	ErrInvalidProductData  = errors.New("invalid product data")
	ErrInvalidCategoryData = errors.New("invalid category data")
)

// UserRepository defines the interface for user data operations.
// UserRepository định nghĩa interface cho các thao tác dữ liệu người dùng.
type UserRepository interface {
	// Save creates a new user or updates an existing one.
	// Save tạo người dùng mới hoặc cập nhật người dùng hiện có.
	Save(ctx context.Context, user *User) error

	// FindByID retrieves a user by their ID.
	// FindByID lấy người dùng theo ID của họ.
	FindByID(ctx context.Context, id string) (*User, error)

	// FindByEmail retrieves a user by their email address.
	// FindByEmail lấy người dùng theo địa chỉ email của họ.
	FindByEmail(ctx context.Context, email string) (*User, error)

	// Delete removes a user from the repository.
	// Delete xóa người dùng khỏi repository.
	Delete(ctx context.Context, id string) error
}
