// internal/review/delivery/grpc/review_grpc_server.go
package grpc

import (
	"context"
	"log" // Tạm thời dùng log để in lỗi

	"github.com/datngth03/ecommerce-go-app/internal/review/application"
	review_client "github.com/datngth03/ecommerce-go-app/pkg/client/review" // Import mã gRPC đã tạo
)

// ReviewGRPCServer implements the review_client.ReviewServiceServer interface.
// ReviewGRPCServer triển khai interface review_client.ReviewServiceServer.
type ReviewGRPCServer struct {
	review_client.UnimplementedReviewServiceServer // Embedded to satisfy all methods, allows future additions
	reviewService                                  application.ReviewService
}

// NewReviewGRPCServer creates a new instance of ReviewGRPCServer.
// NewReviewGRPCServer tạo một thể hiện mới của ReviewGRPCServer.
func NewReviewGRPCServer(svc application.ReviewService) *ReviewGRPCServer {
	return &ReviewGRPCServer{
		reviewService: svc,
	}
}

// SubmitReview implements the gRPC SubmitReview method.
// SubmitReview triển khai phương thức gRPC SubmitReview.
func (s *ReviewGRPCServer) SubmitReview(ctx context.Context, req *review_client.SubmitReviewRequest) (*review_client.ReviewResponse, error) {
	log.Printf("Received SubmitReview request for product %s by user %s", req.GetProductId(), req.GetUserId())
	resp, err := s.reviewService.SubmitReview(ctx, req)
	if err != nil {
		log.Printf("Error submitting review: %v", err)
		return nil, err
	}
	return resp, nil
}

// GetReviewById implements the gRPC GetReviewById method.
// GetReviewById triển khai phương thức gRPC GetReviewById.
func (s *ReviewGRPCServer) GetReviewById(ctx context.Context, req *review_client.GetReviewByIdRequest) (*review_client.ReviewResponse, error) {
	log.Printf("Received GetReviewById request for ID: %s", req.GetId())
	resp, err := s.reviewService.GetReviewById(ctx, req)
	if err != nil {
		log.Printf("Error getting review by ID: %v", err)
		return nil, err
	}
	return resp, nil
}

// UpdateReview implements the gRPC UpdateReview method.
// UpdateReview triển khai phương thức gRPC UpdateReview.
func (s *ReviewGRPCServer) UpdateReview(ctx context.Context, req *review_client.UpdateReviewRequest) (*review_client.ReviewResponse, error) {
	log.Printf("Received UpdateReview request for ID: %s", req.GetId())
	resp, err := s.reviewService.UpdateReview(ctx, req)
	if err != nil {
		log.Printf("Error updating review: %v", err)
		return nil, err
	}
	return resp, nil
}

// DeleteReview implements the gRPC DeleteReview method.
// DeleteReview triển khai phương thức gRPC DeleteReview.
func (s *ReviewGRPCServer) DeleteReview(ctx context.Context, req *review_client.DeleteReviewRequest) (*review_client.DeleteReviewResponse, error) {
	log.Printf("Received DeleteReview request for ID: %s", req.GetId())
	resp, err := s.reviewService.DeleteReview(ctx, req)
	if err != nil {
		log.Printf("Error deleting review: %v", err)
		return nil, err
	}
	return resp, nil
}

// ListReviewsByProduct implements the gRPC ListReviewsByProduct method.
// ListReviewsByProduct triển khai phương thức gRPC ListReviewsByProduct.
func (s *ReviewGRPCServer) ListReviewsByProduct(ctx context.Context, req *review_client.ListReviewsByProductRequest) (*review_client.ListReviewsResponse, error) {
	log.Printf("Received ListReviewsByProduct request for Product ID: %s", req.GetProductId())
	resp, err := s.reviewService.ListReviewsByProduct(ctx, req)
	if err != nil {
		log.Printf("Error listing reviews by product: %v", err)
		return nil, err
	}
	return resp, nil
}

// ListAllReviews implements the gRPC ListAllReviews method.
// ListAllReviews triển khai phương thức gRPC ListAllReviews.
func (s *ReviewGRPCServer) ListAllReviews(ctx context.Context, req *review_client.ListAllReviewsRequest) (*review_client.ListReviewsResponse, error) {
	log.Printf("Received ListAllReviews request (Product ID: %s, User ID: %s, Min Rating: %d)",
		req.GetProductId(), req.GetUserId(), req.GetMinRating())
	resp, err := s.reviewService.ListAllReviews(ctx, req)
	if err != nil {
		log.Printf("Error listing all reviews: %v", err)
		return nil, err
	}
	return resp, nil
}
