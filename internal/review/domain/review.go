// internal/review/domain/review.go
package domain

import "time"

// Review represents a product review entity.
// Review đại diện cho một thực thể đánh giá sản phẩm.
type Review struct {
	ID        string
	ProductID string
	UserID    string
	Rating    int32 // Rating from 1 to 5
	Comment   string
	CreatedAt time.Time
	UpdatedAt time.Time
}

// NewReview creates a new Review instance.
// NewReview tạo một thể hiện Review mới.
func NewReview(id, productID, userID string, rating int32, comment string) *Review {
	now := time.Now().UTC()
	return &Review{
		ID:        id,
		ProductID: productID,
		UserID:    userID,
		Rating:    rating,
		Comment:   comment,
		CreatedAt: now,
		UpdatedAt: now,
	}
}

// UpdateReviewInfo updates the review's rating and comment.
// UpdateReviewInfo cập nhật xếp hạng và bình luận của đánh giá.
func (r *Review) UpdateReviewInfo(rating int32, comment string) {
	r.Rating = rating
	r.Comment = comment
	r.UpdatedAt = time.Now().UTC()
}
