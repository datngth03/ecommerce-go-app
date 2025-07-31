// internal/review/infrastructure/repository/postgres_review_repo.go
package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/datngth03/ecommerce-go-app/internal/review/domain" // Import domain package
	_ "github.com/lib/pq"                                          // PostgreSQL driver
)

// PostgreSQLReviewRepository implements the domain.ReviewRepository interface.
// PostgreSQLReviewRepository triển khai giao diện domain.ReviewRepository.
type PostgreSQLReviewRepository struct {
	db *sql.DB
}

// NewPostgreSQLReviewRepository creates a new instance of PostgreSQLReviewRepository.
// NewPostgreSQLReviewRepository tạo một thể hiện mới của PostgreSQLReviewRepository.
func NewPostgreSQLReviewRepository(db *sql.DB) *PostgreSQLReviewRepository {
	return &PostgreSQLReviewRepository{db: db}
}

// Save inserts a new review or updates an existing one.
// Save chèn một đánh giá mới hoặc cập nhật một đánh giá hiện có.
func (r *PostgreSQLReviewRepository) Save(ctx context.Context, review *domain.Review) error {
	query := `
        INSERT INTO reviews (id, product_id, user_id, rating, comment, created_at, updated_at)
        VALUES ($1, $2, $3, $4, $5, $6, $7)
        ON CONFLICT (id) DO UPDATE SET
            rating = EXCLUDED.rating,
            comment = EXCLUDED.comment,
            updated_at = EXCLUDED.updated_at
        RETURNING id;
    `
	err := r.db.QueryRowContext(
		ctx,
		query,
		review.ID,
		review.ProductID,
		review.UserID,
		review.Rating,
		review.Comment,
		review.CreatedAt,
		review.UpdatedAt,
	).Scan(&review.ID) // Scan the returned ID back into the struct
	if err != nil {
		// Check for unique constraint violation (product_id, user_id)
		if strings.Contains(err.Error(), "unq_product_user_review") {
			return fmt.Errorf("%w: %v", domain.ErrReviewAlreadyExists, err)
		}
		return fmt.Errorf("%w: %v", domain.ErrFailedToSaveReview, err)
	}
	return nil
}

// FindByID retrieves a review by its ID.
// FindByID truy xuất một đánh giá bằng ID của nó.
func (r *PostgreSQLReviewRepository) FindByID(ctx context.Context, id string) (*domain.Review, error) {
	review := &domain.Review{}
	query := `
        SELECT id, product_id, user_id, rating, comment, created_at, updated_at
        FROM reviews
        WHERE id = $1;
    `
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&review.ID,
		&review.ProductID,
		&review.UserID,
		&review.Rating,
		&review.Comment,
		&review.CreatedAt,
		&review.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("%w: %v", domain.ErrReviewNotFound, err)
		}
		return nil, fmt.Errorf("%w: %v", domain.ErrFailedToRetrieveReview, err)
	}
	return review, nil
}

// FindByProductID retrieves a list of reviews for a specific product with pagination.
// FindByProductID truy xuất danh sách đánh giá cho một sản phẩm cụ thể với phân trang.
func (r *PostgreSQLReviewRepository) FindByProductID(ctx context.Context, productID string, limit, offset int32) ([]*domain.Review, int64, error) {
	var reviews []*domain.Review
	query := `
        SELECT id, product_id, user_id, rating, comment, created_at, updated_at
        FROM reviews
        WHERE product_id = $1
        ORDER BY created_at DESC
        LIMIT $2 OFFSET $3;
    `
	rows, err := r.db.QueryContext(ctx, query, productID, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("%w: %v", domain.ErrFailedToListReviews, err)
	}
	defer rows.Close()

	for rows.Next() {
		review := &domain.Review{}
		err := rows.Scan(
			&review.ID,
			&review.ProductID,
			&review.UserID,
			&review.Rating,
			&review.Comment,
			&review.CreatedAt,
			&review.UpdatedAt,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("%w: %v", domain.ErrFailedToListReviews, err)
		}
		reviews = append(reviews, review)
	}
	if err = rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("%w: %v", domain.ErrFailedToListReviews, err)
	}

	// Get total count for pagination
	var totalCount int64
	countQuery := `SELECT COUNT(id) FROM reviews WHERE product_id = $1;`
	err = r.db.QueryRowContext(ctx, countQuery, productID).Scan(&totalCount)
	if err != nil {
		return nil, 0, fmt.Errorf("%w: %v", domain.ErrFailedToListReviews, err)
	}

	return reviews, totalCount, nil
}

// FindExistingReview checks if a review exists for a given product by a specific user.
// FindExistingReview kiểm tra xem có đánh giá nào tồn tại cho một sản phẩm bởi một người dùng cụ thể hay không.
func (r *PostgreSQLReviewRepository) FindExistingReview(ctx context.Context, productID, userID string) (*domain.Review, error) {
	review := &domain.Review{}
	query := `
        SELECT id, product_id, user_id, rating, comment, created_at, updated_at
        FROM reviews
        WHERE product_id = $1 AND user_id = $2;
    `
	err := r.db.QueryRowContext(ctx, query, productID, userID).Scan(
		&review.ID,
		&review.ProductID,
		&review.UserID,
		&review.Rating,
		&review.Comment,
		&review.CreatedAt,
		&review.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("%w: %v", domain.ErrReviewNotFound, err) // Return specific error if not found
		}
		return nil, fmt.Errorf("%w: %v", domain.ErrFailedToRetrieveReview, err)
	}
	return review, nil
}

// Delete removes a review by its ID.
// Delete xóa một đánh giá bằng ID của nó.
func (r *PostgreSQLReviewRepository) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM reviews WHERE id = $1 RETURNING id;`
	var deletedID string
	err := r.db.QueryRowContext(ctx, query, id).Scan(&deletedID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("%w: %v", domain.ErrReviewNotFound, err)
		}
		return fmt.Errorf("%w: %v", domain.ErrFailedToDeleteReview, err)
	}
	return nil
}

// FindAll retrieves all reviews with optional filters and pagination.
// FindAll truy xuất tất cả đánh giá với bộ lọc tùy chọn và phân trang.
func (r *PostgreSQLReviewRepository) FindAll(ctx context.Context, productID, userID string, minRating int32, limit, offset int32) ([]*domain.Review, int64, error) {
	var reviews []*domain.Review
	var args []interface{}
	argCount := 1
	whereClauses := []string{}

	if productID != "" {
		whereClauses = append(whereClauses, fmt.Sprintf("product_id = $%d", argCount))
		args = append(args, productID)
		argCount++
	}
	if userID != "" {
		whereClauses = append(whereClauses, fmt.Sprintf("user_id = $%d", argCount))
		args = append(args, userID)
		argCount++
	}
	if minRating > 0 && minRating <= 5 {
		whereClauses = append(whereClauses, fmt.Sprintf("rating >= $%d", argCount))
		args = append(args, minRating)
		argCount++
	}

	whereClauseStr := ""
	if len(whereClauses) > 0 {
		whereClauseStr = " WHERE " + strings.Join(whereClauses, " AND ")
	}

	query := fmt.Sprintf(`
        SELECT id, product_id, user_id, rating, comment, created_at, updated_at
        FROM reviews
        %s
        ORDER BY created_at DESC
        LIMIT $%d OFFSET $%d;
    `, whereClauseStr, argCount, argCount+1)
	args = append(args, limit, offset)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("%w: %v", domain.ErrFailedToListReviews, err)
	}
	defer rows.Close()

	for rows.Next() {
		review := &domain.Review{}
		err := rows.Scan(
			&review.ID,
			&review.ProductID,
			&review.UserID,
			&review.Rating,
			&review.Comment,
			&review.CreatedAt,
			&review.UpdatedAt,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("%w: %v", domain.ErrFailedToListReviews, err)
		}
		reviews = append(reviews, review)
	}
	if err = rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("%w: %v", domain.ErrFailedToListReviews, err)
	}

	// Get total count for pagination
	var totalCount int64
	countQuery := fmt.Sprintf(`SELECT COUNT(id) FROM reviews %s;`, whereClauseStr)
	// Remove limit and offset args for count query as they are only for main query
	countArgs := args[:len(args)-2] // Exclude limit and offset
	err = r.db.QueryRowContext(ctx, countQuery, countArgs...).Scan(&totalCount)
	if err != nil {
		return nil, 0, fmt.Errorf("%w: %v", domain.ErrFailedToListReviews, err)
	}

	return reviews, totalCount, nil
}
