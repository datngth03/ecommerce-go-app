// internal/search/application/service.go
package application

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/datngth03/ecommerce-go-app/internal/search/domain"
	search_client "github.com/datngth03/ecommerce-go-app/pkg/client/search" // Generated gRPC client
	// product_client "github.com/datngth03/ecommerce-go-app/pkg/client/product" // TODO: Nếu cần gọi Product Service để lấy thông tin sản phẩm
)

// SearchService defines the application service interface for search-related operations.
// SearchService định nghĩa interface dịch vụ ứng dụng cho các thao tác liên quan đến tìm kiếm.
type SearchService interface {
	IndexProduct(ctx context.Context, req *search_client.IndexProductRequest) (*search_client.IndexProductResponse, error)
	SearchProducts(ctx context.Context, req *search_client.SearchProductsRequest) (*search_client.SearchProductsResponse, error)
	DeleteProductFromIndex(ctx context.Context, req *search_client.DeleteProductFromIndexRequest) (*search_client.DeleteProductFromIndexResponse, error)
}

// searchService implements the SearchService interface.
// searchService triển khai interface SearchService.
type searchService struct {
	searchRepo domain.SearchProductRepository
	// productClient product_client.ProductServiceClient // TODO: Thêm client Product Service nếu cần
}

// NewSearchService creates a new instance of SearchService.
// NewSearchService tạo một thể hiện mới của SearchService.
func NewSearchService(searchRepo domain.SearchProductRepository /*, productClient product_client.ProductServiceClient */) SearchService {
	return &searchService{
		searchRepo: searchRepo,
		// productClient: productClient,
	}
}

// --- Product Search Use Cases ---

// IndexProduct handles adding or updating a product in the search index.
// IndexProduct xử lý việc thêm hoặc cập nhật một sản phẩm vào chỉ mục tìm kiếm.
func (s *searchService) IndexProduct(ctx context.Context, req *search_client.IndexProductRequest) (*search_client.IndexProductResponse, error) {
	if req.GetId() == "" || req.GetName() == "" || req.GetPrice() <= 0 {
		return nil, errors.New("product ID, name, and price are required for indexing")
	}

	// Chuyển đổi từ Protobuf request sang domain entity
	product := &domain.SearchProduct{
		ID:            req.GetId(),
		Name:          req.GetName(),
		Description:   req.GetDescription(),
		Price:         req.GetPrice(),
		CategoryID:    req.GetCategoryId(),
		ImageURLs:     req.GetImageUrls(),
		StockQuantity: req.GetStockQuantity(),
		CreatedAt:     time.Now(), // Khi Index, có thể lấy thời gian hiện tại
		UpdatedAt:     time.Now(),
	}

	// Nếu cần lấy thêm dữ liệu từ Product Service:
	// _, err := s.productClient.GetProductById(ctx, &product_client.GetProductByIdRequest{Id: req.GetId()})
	// if err != nil {
	//    return nil, fmt.Errorf("product not found in Product Service: %w", err)
	// }

	if err := s.searchRepo.IndexProduct(ctx, product); err != nil {
		return nil, fmt.Errorf("failed to index product: %w", err)
	}

	return &search_client.IndexProductResponse{
		Success: true,
		Message: fmt.Sprintf("Product %s indexed successfully", req.GetId()),
	}, nil
}

// SearchProducts handles searching for products.
// SearchProducts xử lý việc tìm kiếm sản phẩm.
func (s *searchService) SearchProducts(ctx context.Context, req *search_client.SearchProductsRequest) (*search_client.SearchProductsResponse, error) {
	// Basic validation (optional, depending on requirements)
	if req.GetLimit() < 0 || req.GetOffset() < 0 {
		return nil, errors.New("limit and offset cannot be negative")
	}

	products, totalHits, err := s.searchRepo.SearchProducts(
		ctx,
		req.GetQuery(),
		req.GetCategoryId(),
		req.GetMinPrice(),
		req.GetMaxPrice(),
		req.GetLimit(),
		req.GetOffset(),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to search products: %w", err)
	}

	// Chuyển đổi từ domain entities sang Protobuf response
	searchProductsResp := make([]*search_client.SearchProduct, len(products))
	for i, p := range products {
		searchProductsResp[i] = &search_client.SearchProduct{
			Id:            p.ID,
			Name:          p.Name,
			Description:   p.Description,
			Price:         p.Price,
			CategoryId:    p.CategoryID,
			ImageUrls:     p.ImageURLs,
			StockQuantity: p.StockQuantity,
		}
	}

	return &search_client.SearchProductsResponse{
		Products:  searchProductsResp,
		TotalHits: totalHits,
	}, nil
}

// DeleteProductFromIndex handles deleting a product from the search index.
// DeleteProductFromIndex xử lý việc xóa một sản phẩm khỏi chỉ mục tìm kiếm.
func (s *searchService) DeleteProductFromIndex(ctx context.Context, req *search_client.DeleteProductFromIndexRequest) (*search_client.DeleteProductFromIndexResponse, error) {
	if req.GetProductId() == "" {
		return nil, errors.New("product ID is required for deletion from index")
	}

	if err := s.searchRepo.DeleteProductFromIndex(ctx, req.GetProductId()); err != nil {
		// Distinguish between not found and other errors
		if errors.Is(err, domain.ErrProductNotFoundInSearch) {
			return &search_client.DeleteProductFromIndexResponse{
				Success: false,
				Message: "Product not found in search index, nothing to delete",
			}, nil
		}
		return nil, fmt.Errorf("failed to delete product from index: %w", err)
	}

	return &search_client.DeleteProductFromIndexResponse{
		Success: true,
		Message: fmt.Sprintf("Product %s deleted from index successfully", req.GetProductId()),
	}, nil
}
