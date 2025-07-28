// internal/recommendation/domain/repository.go
package domain

import (
	"context"
	"errors"
)

// Define sentinel errors for the repository.
// Định nghĩa các lỗi sentinel cho repository.
var (
	ErrFailedToSaveInteraction      = errors.New("failed to save user interaction")
	ErrFailedToRetrieveInteractions = errors.New("failed to retrieve user interactions")
)

// UserInteractionRepository defines the interface for interacting with user interaction data.
// UserInteractionRepository định nghĩa interface để tương tác với dữ liệu tương tác người dùng.
type UserInteractionRepository interface {
	// Save records a user interaction.
	// Save lưu một tương tác người dùng.
	Save(ctx context.Context, interaction *UserInteraction) error
	// FindByUserID retrieves interactions for a specific user.
	// FindByUserID truy xuất các tương tác cho một người dùng cụ thể.
	FindByUserID(ctx context.Context, userID string) ([]*UserInteraction, error)
	// FindPopularProducts retrieves a list of top N popular products based on interactions.
	// FindPopularProducts truy xuất danh sách N sản phẩm phổ biến nhất dựa trên các tương tác.
	FindPopularProducts(ctx context.Context, limit int32) ([]*UserInteraction, error)
	// TODO: Add methods for more complex recommendation logic
}
