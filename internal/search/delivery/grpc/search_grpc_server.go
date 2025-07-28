// internal/search/delivery/grpc/search_grpc_server.go
package grpc

import (
	"context"
	"log" // Tạm thời dùng log để in lỗi

	// "time" // Để định dạng thời gian cho Protobuf

	"google.golang.org/grpc/codes"  // Để sử dụng mã lỗi gRPC
	"google.golang.org/grpc/status" // Để tạo lỗi gRPC status

	"github.com/datngth03/ecommerce-go-app/internal/search/application"
	"github.com/datngth03/ecommerce-go-app/internal/search/domain"          // Import domain package để sử dụng các entity và sentinel errors
	search_client "github.com/datngth03/ecommerce-go-app/pkg/client/search" // Generated Search gRPC client
)

// SearchGRPCServer implements the search_client.SearchServiceServer interface.
// SearchGRPCServer triển khai giao diện search_client.SearchServiceServer.
type SearchGRPCServer struct {
	search_client.UnimplementedSearchServiceServer // Nhúng để thỏa mãn tất cả các phương thức, cho phép tương thích ngược
	searchService                                  application.SearchService
}

// NewSearchGRPCServer creates a new instance of SearchGRPCServer.
// NewSearchGRPCServer tạo một thể hiện mới của SearchGRPCServer.
func NewSearchGRPCServer(svc application.SearchService) *SearchGRPCServer {
	return &SearchGRPCServer{
		searchService: svc,
	}
}

// IndexProduct implements the gRPC IndexProduct method.
// IndexProduct triển khai phương thức gRPC IndexProduct.
func (s *SearchGRPCServer) IndexProduct(ctx context.Context, req *search_client.IndexProductRequest) (*search_client.IndexProductResponse, error) {
	log.Printf("Received IndexProduct request for product ID: %s", req.GetId())
	resp, err := s.searchService.IndexProduct(ctx, req)
	if err != nil {
		log.Printf("Error indexing product: %v", err)
		// Chuyển đổi lỗi domain sang mã lỗi gRPC phù hợp
		if err == domain.ErrFailedToIndexProduct {
			return nil, status.Errorf(codes.Internal, "failed to index product: %v", err)
		}
		return nil, status.Errorf(codes.Internal, "internal error: %v", err)
	}
	return resp, nil
}

// SearchProducts implements the gRPC SearchProducts method.
// SearchProducts triển khai phương thức gRPC SearchProducts.
func (s *SearchGRPCServer) SearchProducts(ctx context.Context, req *search_client.SearchProductsRequest) (*search_client.SearchProductsResponse, error) {
	log.Printf("Received SearchProducts request for query: '%s', category: '%s'", req.GetQuery(), req.GetCategoryId())
	resp, err := s.searchService.SearchProducts(ctx, req)
	if err != nil {
		log.Printf("Error searching products: %v", err)
		if err == domain.ErrFailedToSearchProducts {
			return nil, status.Errorf(codes.Internal, "failed to search products: %v", err)
		}
		return nil, status.Errorf(codes.Internal, "internal error: %v", err)
	}

	// Chuyển đổi danh sách domain.SearchProduct từ resp.Products sang search_client.SearchProduct
	protoProducts := make([]*search_client.SearchProduct, len(resp.GetProducts()))
	for i, p := range resp.GetProducts() {
		protoProducts[i] = &search_client.SearchProduct{
			Id:            p.Id,
			Name:          p.Name,
			Description:   p.Description,
			Price:         p.Price,
			CategoryId:    p.CategoryId,
			ImageUrls:     p.ImageUrls,
			StockQuantity: p.StockQuantity,
			// CreatedAt và UpdatedAt không có trong SearchProduct của Protobuf
		}
	}

	return &search_client.SearchProductsResponse{
		Products:  protoProducts,
		TotalHits: resp.GetTotalHits(),
	}, nil
}

// DeleteProductFromIndex implements the gRPC DeleteProductFromIndex method.
// DeleteProductFromIndex triển khai phương thức gRPC DeleteProductFromIndex.
func (s *SearchGRPCServer) DeleteProductFromIndex(ctx context.Context, req *search_client.DeleteProductFromIndexRequest) (*search_client.DeleteProductFromIndexResponse, error) {
	log.Printf("Received DeleteProductFromIndex request for product ID: %s", req.GetProductId())
	resp, err := s.searchService.DeleteProductFromIndex(ctx, req)
	if err != nil {
		log.Printf("Error deleting product from index: %v", err)
		if err == domain.ErrProductNotFoundInSearch {
			return &search_client.DeleteProductFromIndexResponse{
				Success: false,
				Message: "Product not found in search index",
			}, nil
		}
		if err == domain.ErrFailedToDeleteFromIndex {
			return nil, status.Errorf(codes.Internal, "failed to delete product from index: %v", err)
		}
		return nil, status.Errorf(codes.Internal, "internal error: %v", err)
	}
	return resp, nil
}
