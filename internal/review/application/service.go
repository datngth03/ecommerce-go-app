// internal/review/application/service.go
package application

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/datngth03/ecommerce-go-app/internal/review/domain"
	product_client "github.com/datngth03/ecommerce-go-app/pkg/client/product" // Generated Product gRPC client
	review_client "github.com/datngth03/ecommerce-go-app/pkg/client/review"   // Generated Review gRPC client
	"github.com/google/uuid"                                                  // For generating UUIDs
)

// ReviewService defines the application service interface for review-related operations.
// ReviewService định nghĩa interface dịch vụ ứng dụng cho các thao tác liên quan đến đánh giá sản phẩm.
type ReviewService interface {
	SubmitReview(ctx context.Context, req *review_client.SubmitReviewRequest) (*review_client.ReviewResponse, error)
	GetReviewById(ctx context.Context, req *review_client.GetReviewByIdRequest) (*review_client.ReviewResponse, error)
	UpdateReview(ctx context.Context, req *review_client.UpdateReviewRequest) (*review_client.ReviewResponse, error)
	DeleteReview(ctx context.Context, req *review_client.DeleteReviewRequest) (*review_client.DeleteReviewResponse, error)
	ListReviewsByProduct(ctx context.Context, req *review_client.ListReviewsByProductRequest) (*review_client.ListReviewsResponse, error)
	ListAllReviews(ctx context.Context, req *review_client.ListAllReviewsRequest) (*review_client.ListReviewsResponse, error)
}

// reviewService implements the ReviewService interface.
// reviewService triển khai interface ReviewService.
type reviewService struct {
	reviewRepo    domain.ReviewRepository
	productClient product_client.ProductServiceClient // To verify product existence
}

// NewReviewService creates a new instance of ReviewService.
// NewReviewService tạo một thể hiện mới của ReviewService.
func NewReviewService(reviewRepo domain.ReviewRepository, productClient product_client.ProductServiceClient) ReviewService {
	return &reviewService{
		reviewRepo:    reviewRepo,
		productClient: productClient,
	}
}

// SubmitReview handles the submission of a new product review.
// SubmitReview xử lý việc gửi một đánh giá sản phẩm mới.
func (s *reviewService) SubmitReview(ctx context.Context, req *review_client.SubmitReviewRequest) (*review_client.ReviewResponse, error) {
	// Basic validation
	if req.GetProductId() == "" || req.GetUserId() == "" || req.GetRating() < 1 || req.GetRating() > 5 {
		return nil, errors.New("product ID, user ID, and valid rating (1-5) are required")
	}

	// Verify product existence using Product Service
	_, err := s.productClient.GetProductById(ctx, &product_client.GetProductByIdRequest{Id: req.GetProductId()})
	if err != nil {
		// If product not found, return an error. Differentiate between actual gRPC error and not found.
		return nil, fmt.Errorf("product not found or unable to verify product: %w", err)
	}

	// Check if user has already reviewed this product
	existingReview, err := s.reviewRepo.FindExistingReview(ctx, req.GetProductId(), req.GetUserId())
	if err != nil && !errors.Is(err, domain.ErrReviewNotFound) {
		return nil, fmt.Errorf("failed to check for existing review: %w", err)
	}
	if existingReview != nil {
		return nil, fmt.Errorf("%w : %v", domain.ErrReviewAlreadyExists, err) // User already reviewed this product
	}

	// Create new review domain entity
	reviewID := uuid.New().String()
	review := domain.NewReview(
		reviewID,
		req.GetProductId(),
		req.GetUserId(),
		req.GetRating(),
		req.GetComment(),
	)

	// Save review to repository
	if err := s.reviewRepo.Save(ctx, review); err != nil {
		return nil, fmt.Errorf("%w: %v", domain.ErrFailedToSaveReview, err)
	}

	return &review_client.ReviewResponse{
		Review: &review_client.Review{
			Id:        review.ID,
			ProductId: review.ProductID,
			UserId:    review.UserID,
			Rating:    review.Rating,
			Comment:   review.Comment,
			CreatedAt: review.CreatedAt.Format(time.RFC3339),
			UpdatedAt: review.UpdatedAt.Format(time.RFC3339),
		},
	}, nil
}

// GetReviewById handles retrieving review details.
// GetReviewById xử lý việc lấy chi tiết đánh giá.
func (s *reviewService) GetReviewById(ctx context.Context, req *review_client.GetReviewByIdRequest) (*review_client.ReviewResponse, error) {
	if req.GetId() == "" {
		return nil, errors.New("review ID is required")
	}

	review, err := s.reviewRepo.FindByID(ctx, req.GetId())
	if err != nil {
		if errors.Is(err, domain.ErrReviewNotFound) {
			return nil, fmt.Errorf("%w: %v", domain.ErrReviewNotFound, err)
		}
		return nil, fmt.Errorf("%w: %v", domain.ErrFailedToRetrieveReview, err)
	}

	return &review_client.ReviewResponse{
		Review: &review_client.Review{
			Id:        review.ID,
			ProductId: review.ProductID,
			UserId:    review.UserID,
			Rating:    review.Rating,
			Comment:   review.Comment,
			CreatedAt: review.CreatedAt.Format(time.RFC3339),
			UpdatedAt: review.UpdatedAt.Format(time.RFC3339),
		},
	}, nil
}

// UpdateReview handles updating an existing review.
// UpdateReview xử lý việc cập nhật một đánh giá hiện có.
func (s *reviewService) UpdateReview(ctx context.Context, req *review_client.UpdateReviewRequest) (*review_client.ReviewResponse, error) {
	if req.GetId() == "" || req.GetRating() < 1 || req.GetRating() > 5 {
		return nil, errors.New("review ID and valid rating (1-5) are required for update")
	}

	existingReview, err := s.reviewRepo.FindByID(ctx, req.GetId())
	if err != nil {
		if errors.Is(err, domain.ErrReviewNotFound) {
			return nil, fmt.Errorf("%w: %v", domain.ErrReviewNotFound, err)
		}
		return nil, fmt.Errorf("%w: %v", domain.ErrFailedToRetrieveReview, err)
	}

	// You might add logic here to ensure only the original user can update their review
	// if existingReview.UserID != req.GetUserId() {
	//    return nil, errors.New("unauthorized to update this review")
	// }

	existingReview.UpdateReviewInfo(req.GetRating(), req.GetComment())

	if err := s.reviewRepo.Save(ctx, existingReview); err != nil {
		return nil, fmt.Errorf("%w: %v", domain.ErrFailedToSaveReview, err)
	}

	return &review_client.ReviewResponse{
		Review: &review_client.Review{
			Id:        existingReview.ID,
			ProductId: existingReview.ProductID,
			UserId:    existingReview.UserID,
			Rating:    existingReview.Rating,
			Comment:   existingReview.Comment,
			CreatedAt: existingReview.CreatedAt.Format(time.RFC3339),
			UpdatedAt: existingReview.UpdatedAt.Format(time.RFC3339),
		},
	}, nil
}

// DeleteReview handles deleting a review by its ID.
// DeleteReview xử lý việc xóa một đánh giá bằng ID của nó.
func (s *reviewService) DeleteReview(ctx context.Context, req *review_client.DeleteReviewRequest) (*review_client.DeleteReviewResponse, error) {
	if req.GetId() == "" {
		return nil, errors.New("review ID is required for deletion")
	}

	err := s.reviewRepo.Delete(ctx, req.GetId())
	if err != nil {
		if errors.Is(err, domain.ErrReviewNotFound) {
			return &review_client.DeleteReviewResponse{Success: false, Message: "Review not found"}, nil
		}
		return nil, fmt.Errorf("%w: %v", domain.ErrFailedToDeleteReview, err)
	}

	return &review_client.DeleteReviewResponse{Success: true, Message: "Review deleted successfully"}, nil
}

// ListReviewsByProduct handles listing reviews for a specific product with pagination.
// ListReviewsByProduct xử lý việc liệt kê đánh giá cho một sản phẩm cụ thể với phân trang.
func (s *reviewService) ListReviewsByProduct(ctx context.Context, req *review_client.ListReviewsByProductRequest) (*review_client.ListReviewsResponse, error) {
	if req.GetProductId() == "" {
		return nil, errors.New("product ID is required to list reviews by product")
	}

	reviews, totalCount, err := s.reviewRepo.FindByProductID(ctx, req.GetProductId(), req.GetLimit(), req.GetOffset())
	if err != nil {
		return nil, fmt.Errorf("%w: %v", domain.ErrFailedToListReviews, err)
	}

	reviewResponses := make([]*review_client.Review, len(reviews))
	for i, r := range reviews {
		reviewResponses[i] = &review_client.Review{
			Id:        r.ID,
			ProductId: r.ProductID,
			UserId:    r.UserID,
			Rating:    r.Rating,
			Comment:   r.Comment,
			CreatedAt: r.CreatedAt.Format(time.RFC3339),
			UpdatedAt: r.UpdatedAt.Format(time.RFC3339),
		}
	}

	return &review_client.ListReviewsResponse{
		Reviews:    reviewResponses,
		TotalCount: totalCount,
	}, nil
}

// ListAllReviews handles listing all reviews with optional filters and pagination.
// ListAllReviews xử lý việc liệt kê tất cả đánh giá với bộ lọc tùy chọn và phân trang.
func (s *reviewService) ListAllReviews(ctx context.Context, req *review_client.ListAllReviewsRequest) (*review_client.ListReviewsResponse, error) {
	reviews, totalCount, err := s.reviewRepo.FindAll(
		ctx,
		req.GetProductId(),
		req.GetUserId(),
		req.GetMinRating(),
		req.GetLimit(),
		req.GetOffset(),
	)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", domain.ErrFailedToListReviews, err)
	}

	reviewResponses := make([]*review_client.Review, len(reviews))
	for i, r := range reviews {
		reviewResponses[i] = &review_client.Review{
			Id:        r.ID,
			ProductId: r.ProductID,
			UserId:    r.UserID,
			Rating:    r.Rating,
			Comment:   r.Comment,
			CreatedAt: r.CreatedAt.Format(time.RFC3339),
			UpdatedAt: r.UpdatedAt.Format(time.RFC3339),
		}
	}

	return &review_client.ListReviewsResponse{
		Reviews:    reviewResponses,
		TotalCount: totalCount,
	}, nil
}
