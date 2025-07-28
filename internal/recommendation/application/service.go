// internal/recommendation/application/service.go
package application

import (
	"context"
	"fmt"
	"log" // Tạm thời dùng log để in lỗi
	"time"

	"github.com/google/uuid"

	"github.com/datngth03/ecommerce-go-app/internal/recommendation/domain"
	product_client "github.com/datngth03/ecommerce-go-app/pkg/client/product"               // Client for Product Service
	recommendation_client "github.com/datngth03/ecommerce-go-app/pkg/client/recommendation" // Generated gRPC client
)

// RecommendationService defines the application service interface for recommendation-related operations.
// RecommendationService định nghĩa interface dịch vụ ứng dụng cho các thao tác liên quan đến gợi ý.
type RecommendationService interface {
	RecordInteraction(ctx context.Context, req *recommendation_client.RecordInteractionRequest) (*recommendation_client.RecordInteractionResponse, error)
	GetRecommendations(ctx context.Context, req *recommendation_client.GetRecommendationsRequest) (*recommendation_client.GetRecommendationsResponse, error)
	GetPopularProducts(ctx context.Context, req *recommendation_client.GetPopularProductsRequest) (*recommendation_client.GetPopularProductsResponse, error)
}

// recommendationService implements the RecommendationService interface.
// recommendationService triển khai interface RecommendationService.
type recommendationService struct {
	interactionRepo domain.UserInteractionRepository
	productClient   product_client.ProductServiceClient // Client để lấy thông tin sản phẩm
	// TODO: Add other dependencies like ML model integration
	// Thêm các dependency khác như tích hợp mô hình ML
}

// NewRecommendationService creates a new instance of RecommendationService.
// NewRecommendationService tạo một thể hiện mới của RecommendationService.
func NewRecommendationService(
	interactionRepo domain.UserInteractionRepository,
	productClient product_client.ProductServiceClient,
) RecommendationService {
	return &recommendationService{
		interactionRepo: interactionRepo,
		productClient:   productClient,
	}
}

// RecordInteraction handles recording a user's interaction with a product.
// RecordInteraction xử lý việc ghi lại tương tác của người dùng với sản phẩm.
func (s *recommendationService) RecordInteraction(ctx context.Context, req *recommendation_client.RecordInteractionRequest) (*recommendation_client.RecordInteractionResponse, error) {
	if req.GetUserId() == "" || req.GetProductId() == "" || req.GetEventType() == "" {
		return nil, fmt.Errorf("user ID, product ID, and event type are required")
	}

	// Optionally, validate product existence via Product Service
	// Có thể, xác thực sự tồn tại của sản phẩm thông qua Product Service
	_, err := s.productClient.GetProductById(ctx, &product_client.GetProductByIdRequest{Id: req.GetProductId()})
	if err != nil {
		log.Printf("Warning: Product %s not found for interaction. Error: %v", req.GetProductId(), err)
		// Decide whether to error out or continue. For now, we'll continue.
		// Quyết định có nên báo lỗi hay tiếp tục. Hiện tại, chúng ta sẽ tiếp tục.
	}

	interactionID := uuid.New().String()
	interaction := domain.NewUserInteraction(
		interactionID,
		req.GetUserId(),
		req.GetProductId(),
		req.GetEventType(),
		time.Unix(req.GetTimestamp(), 0), // Convert Unix timestamp to time.Time
	)

	if err := s.interactionRepo.Save(ctx, interaction); err != nil {
		return nil, fmt.Errorf("%w: %v", domain.ErrFailedToSaveInteraction, err)
	}

	return &recommendation_client.RecordInteractionResponse{
		Success: true,
		Message: "Interaction recorded successfully",
	}, nil
}

// GetRecommendations provides personalized product recommendations for a user.
// (Currently a basic implementation, can be enhanced with ML models).
// GetRecommendations cung cấp gợi ý sản phẩm cá nhân hóa cho người dùng.
// (Hiện tại là triển khai cơ bản, có thể nâng cao với các mô hình ML).
func (s *recommendationService) GetRecommendations(ctx context.Context, req *recommendation_client.GetRecommendationsRequest) (*recommendation_client.GetRecommendationsResponse, error) {
	if req.GetUserId() == "" {
		return nil, fmt.Errorf("user ID is required for recommendations")
	}

	// For a basic implementation, we can just get recent interactions
	// For more complex logic, use collaborative filtering, content-based, etc.
	// Với triển khai cơ bản, chúng ta có thể lấy các tương tác gần đây
	// Đối với logic phức tạp hơn, hãy sử dụng lọc cộng tác, dựa trên nội dung, v.v.
	interactions, err := s.interactionRepo.FindByUserID(ctx, req.GetUserId())
	if err != nil {
		return nil, fmt.Errorf("%w: %v", domain.ErrFailedToRetrieveInteractions, err)
	}

	// Extract unique product IDs from interactions
	// Trích xuất các ID sản phẩm duy nhất từ các tương tác
	productIDs := make(map[string]bool)
	var recommendedProducts []*recommendation_client.RecommendedProduct
	for _, interaction := range interactions {
		if !productIDs[interaction.ProductID] {
			productIDs[interaction.ProductID] = true
			// Fetch product details from Product Service
			// Lấy chi tiết sản phẩm từ Product Service
			productResp, err := s.productClient.GetProductById(ctx, &product_client.GetProductByIdRequest{Id: interaction.ProductID})
			if err != nil {
				log.Printf("Warning: Could not fetch details for product %s for recommendation. Error: %v", interaction.ProductID, err)
				continue // Skip if product details can't be fetched
			}

			recommendedProducts = append(recommendedProducts, &recommendation_client.RecommendedProduct{
				ProductId: productResp.GetProduct().GetId(),
				Name:      productResp.GetProduct().GetName(),
				// Assuming first image URL is main_image_url
				ImageUrl: func() string {
					if len(productResp.GetProduct().GetImageUrls()) > 0 {
						return productResp.GetProduct().GetImageUrls()[0]
					}
					return ""
				}(),
				Score: 1.0, // Basic score, can be more sophisticated
			})
		}
	}

	// Apply limit and offset (simple in-memory pagination for now)
	// Áp dụng giới hạn và offset (phân trang trong bộ nhớ đơn giản)
	if req.GetLimit() > 0 && len(recommendedProducts) > int(req.GetOffset()) {
		start := int(req.GetOffset())
		end := start + int(req.GetLimit())
		if end > len(recommendedProducts) {
			end = len(recommendedProducts)
		}
		recommendedProducts = recommendedProducts[start:end]
	} else if req.GetLimit() == 0 {
		recommendedProducts = recommendedProducts[req.GetOffset():]
	} else {
		recommendedProducts = []*recommendation_client.RecommendedProduct{} // No products if offset is too high
	}

	return &recommendation_client.GetRecommendationsResponse{
		Products: recommendedProducts,
		Message:  "Recommendations generated successfully",
	}, nil
}

// GetPopularProducts provides a list of top N popular products based on total interactions.
// GetPopularProducts cung cấp danh sách N sản phẩm phổ biến nhất dựa trên tổng số tương tác.
func (s *recommendationService) GetPopularProducts(ctx context.Context, req *recommendation_client.GetPopularProductsRequest) (*recommendation_client.GetPopularProductsResponse, error) {
	if req.GetLimit() <= 0 {
		return nil, fmt.Errorf("limit must be positive for popular products")
	}

	interactions, err := s.interactionRepo.FindPopularProducts(ctx, req.GetLimit())
	if err != nil {
		return nil, fmt.Errorf("%w: %v", domain.ErrFailedToRetrieveInteractions, err)
	}

	var popularProducts []*recommendation_client.RecommendedProduct
	for _, interaction := range interactions {
		// Fetch product details from Product Service for rich information
		// Lấy chi tiết sản phẩm từ Product Service để có thông tin đầy đủ
		productResp, err := s.productClient.GetProductById(ctx, &product_client.GetProductByIdRequest{Id: interaction.ProductID})
		if err != nil {
			log.Printf("Warning: Could not fetch details for popular product %s. Error: %v", interaction.ProductID, err)
			continue
		}

		popularProducts = append(popularProducts, &recommendation_client.RecommendedProduct{
			ProductId: productResp.GetProduct().GetId(),
			Name:      productResp.GetProduct().GetName(),
			ImageUrl: func() string {
				if len(productResp.GetProduct().GetImageUrls()) > 0 {
					return productResp.GetProduct().GetImageUrls()[0]
				}
				return ""
			}(),
			Score: 1.0, // Score can be based on interaction count from repo
		})
	}

	return &recommendation_client.GetPopularProductsResponse{
		Products: popularProducts,
		Message:  "Popular products retrieved successfully",
	}, nil
}
