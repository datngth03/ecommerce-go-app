// internal/recommendation/delivery/grpc/recommendation_grpc_server.go
package grpc

import (
	"context"
	// "fmt"
	"log" // Tạm thời dùng log để in lỗi
	// "time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/datngth03/ecommerce-go-app/internal/recommendation/application"
	// "github.com/datngth03/ecommerce-go-app/internal/recommendation/domain" // Import domain package
	recommendation_client "github.com/datngth03/ecommerce-go-app/pkg/client/recommendation" // Generated gRPC client
)

// RecommendationGRPCServer implements the recommendation_client.RecommendationServiceServer interface.
// RecommendationGRPCServer triển khai interface recommendation_client.RecommendationServiceServer.
type RecommendationGRPCServer struct {
	recommendation_client.UnimplementedRecommendationServiceServer // Embedded to satisfy all methods
	recommendationService                                          application.RecommendationService
}

// NewRecommendationGRPCServer creates a new instance of RecommendationGRPCServer.
// NewRecommendationGRPCServer tạo một thể hiện mới của RecommendationGRPCServer.
func NewRecommendationGRPCServer(svc application.RecommendationService) *RecommendationGRPCServer {
	return &RecommendationGRPCServer{
		recommendationService: svc,
	}
}

// RecordInteraction implements the gRPC RecordInteraction method.
// RecordInteraction triển khai phương thức gRPC RecordInteraction.
func (s *RecommendationGRPCServer) RecordInteraction(ctx context.Context, req *recommendation_client.RecordInteractionRequest) (*recommendation_client.RecordInteractionResponse, error) {
	log.Printf("Nhận yêu cầu RecordInteraction cho User ID: %s, Product ID: %s, Event: %s", req.GetUserId(), req.GetProductId(), req.GetEventType())
	resp, err := s.recommendationService.RecordInteraction(ctx, req)
	if err != nil {
		log.Printf("Lỗi khi ghi lại tương tác người dùng: %v", err)
		return nil, status.Errorf(codes.Internal, "failed to record interaction: %v", err)
	}
	return resp, nil
}

// GetRecommendations implements the gRPC GetRecommendations method.
// GetRecommendations triển khai phương thức gRPC GetRecommendations.
func (s *RecommendationGRPCServer) GetRecommendations(ctx context.Context, req *recommendation_client.GetRecommendationsRequest) (*recommendation_client.GetRecommendationsResponse, error) {
	log.Printf("Nhận yêu cầu GetRecommendations cho User ID: %s, Limit: %d", req.GetUserId(), req.GetLimit())
	resp, err := s.recommendationService.GetRecommendations(ctx, req)
	if err != nil {
		log.Printf("Lỗi khi lấy gợi ý sản phẩm: %v", err)
		return nil, status.Errorf(codes.Internal, "failed to get recommendations: %v", err)
	}
	return resp, nil
}

// GetPopularProducts implements the gRPC GetPopularProducts method.
// GetPopularProducts triển khai phương thức gRPC GetPopularProducts.
func (s *RecommendationGRPCServer) GetPopularProducts(ctx context.Context, req *recommendation_client.GetPopularProductsRequest) (*recommendation_client.GetPopularProductsResponse, error) {
	log.Printf("Nhận yêu cầu GetPopularProducts, Limit: %d", req.GetLimit())
	resp, err := s.recommendationService.GetPopularProducts(ctx, req)
	if err != nil {
		log.Printf("Lỗi khi lấy sản phẩm phổ biến: %v", err)
		return nil, status.Errorf(codes.Internal, "failed to get popular products: %v", err)
	}
	return resp, nil
}
