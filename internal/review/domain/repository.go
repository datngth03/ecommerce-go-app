// internal/review/domain/repository.go
package domain

import (
	"context"
	"errors" // Import errors package
)

// Define common errors for the domain
// Định nghĩa các lỗi chung cho domain
var (
	ErrReviewNotFound         = errors.New("review not found")
	ErrFailedToSaveReview     = errors.New("failed to save review")
	ErrFailedToDeleteReview   = errors.New("failed to delete review")
	ErrFailedToUpdateReview   = errors.New("failed to update review")
	ErrFailedToRetrieveReview = errors.New("failed to retrieve review")
	ErrReviewAlreadyExists    = errors.New("review already exists for this product by this user")
	ErrFailedToListReviews    = errors.New("failed to list reviews for product")
)

// ReviewRepository defines the interface for interacting with review data.
// ReviewRepository định nghĩa interface để tương tác với dữ liệu đánh giá.
type ReviewRepository interface {
	Save(ctx context.Context, review *Review) error
	FindByID(ctx context.Context, id string) (*Review, error)
	FindByProductID(ctx context.Context, productID string, limit, offset int32) ([]*Review, int64, error)
	FindExistingReview(ctx context.Context, productID, userID string) (*Review, error) // <-- ĐÃ THÊM PHƯƠNG THỨC NÀY

	FindAll(ctx context.Context, productID, userID string, minRating int32, limit, offset int32) ([]*Review, int64, error)
	Delete(ctx context.Context, id string) error
}
