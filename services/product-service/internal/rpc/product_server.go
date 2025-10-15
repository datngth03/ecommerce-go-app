package rpc

import (
	"context"
	"strings"

	pb "github.com/datngth03/ecommerce-go-app/proto/product_service"
	"github.com/datngth03/ecommerce-go-app/services/product-service/internal/models"
	"github.com/datngth03/ecommerce-go-app/services/product-service/internal/service"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// ProductGRPCServer implements pb.ProductServiceServer
type ProductGRPCServer struct {
	pb.UnimplementedProductServiceServer
	productService  *service.ProductService
	categoryService *service.CategoryService
}

// CategoryGRPCServer implements pb.CategoryServiceServer
type CategoryGRPCServer struct {
	pb.UnimplementedCategoryServiceServer
	categoryService *service.CategoryService
}

// NewProductGRPCServer creates a new gRPC server for products
func NewProductGRPCServer(productService *service.ProductService, categoryService *service.CategoryService) *ProductGRPCServer {
	return &ProductGRPCServer{
		productService:  productService,
		categoryService: categoryService,
	}
}

// NewCategoryGRPCServer creates a new gRPC server for categories
func NewCategoryGRPCServer(categoryService *service.CategoryService) *CategoryGRPCServer {
	return &CategoryGRPCServer{
		categoryService: categoryService,
	}
}

// ==================== PRODUCT SERVICE METHODS ====================

func (s *ProductGRPCServer) CreateProduct(ctx context.Context, req *pb.CreateProductRequest) (*pb.CreateProductResponse, error) {
	createReq := &models.CreateProductRequest{
		Name:        req.Name,
		Description: req.Description,
		Price:       req.Price,
		CategoryID:  req.CategoryId,
		ImageURL:    req.ImageUrl,
	}

	product, err := s.productService.CreateProduct(ctx, createReq)
	if err != nil {
		if strings.Contains(err.Error(), "already exists") {
			return nil, status.Errorf(codes.AlreadyExists, err.Error())
		}
		if strings.Contains(err.Error(), "not found") {
			return nil, status.Errorf(codes.NotFound, err.Error())
		}
		return nil, status.Errorf(codes.Internal, "failed to create product: %v", err)
	}

	return &pb.CreateProductResponse{
		Product: productResponseToProto(product),
	}, nil
}

func (s *ProductGRPCServer) GetProduct(ctx context.Context, req *pb.GetProductRequest) (*pb.GetProductResponse, error) {
	product, err := s.productService.GetProduct(ctx, req.Id)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			return nil, status.Errorf(codes.NotFound, "product with id %s not found", req.Id)
		}
		return nil, status.Error(codes.Internal, "failed to get product")
	}
	return &pb.GetProductResponse{Product: productResponseToProto(product)}, nil
}

func (s *ProductGRPCServer) UpdateProduct(ctx context.Context, req *pb.UpdateProductRequest) (*pb.UpdateProductResponse, error) {
	updateReq := &models.UpdateProductRequest{
		Name:        req.Name,
		Description: req.Description,
		Price:       req.Price,
		CategoryID:  req.CategoryId,
		ImageURL:    req.ImageUrl,
		IsActive:    req.IsActive,
	}

	product, err := s.productService.UpdateProduct(ctx, req.Id, updateReq)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			return nil, status.Errorf(codes.NotFound, err.Error())
		}
		if strings.Contains(err.Error(), "already exists") {
			return nil, status.Errorf(codes.AlreadyExists, err.Error())
		}
		return nil, status.Errorf(codes.Internal, "failed to update product: %v", err)
	}

	return &pb.UpdateProductResponse{
		Product: productResponseToProto(product),
	}, nil
}

func (s *ProductGRPCServer) DeleteProduct(ctx context.Context, req *pb.DeleteProductRequest) (*emptypb.Empty, error) {
	if err := s.productService.DeleteProduct(ctx, req.Id); err != nil {
		if strings.Contains(err.Error(), "not found") {
			return nil, status.Errorf(codes.NotFound, err.Error())
		}
		return nil, status.Errorf(codes.Internal, "failed to delete product: %v", err)
	}
	return &emptypb.Empty{}, nil
}

// ListProducts được triển khai đầy đủ vì các service khác (ví dụ: Search) có thể cần nó.
func (s *ProductGRPCServer) ListProducts(ctx context.Context, req *pb.ListProductsRequest) (*pb.ListProductsResponse, error) {
	serviceReq := &models.ListProductsRequest{
		Page:       int(req.Page),
		PageSize:   int(req.PageSize),
		CategoryID: req.CategoryId,
	}

	listResponse, err := s.productService.ListProducts(ctx, serviceReq)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to list products")
	}

	return listProductsResponseToProto(listResponse), nil
}

// ==================== CATEGORY SERVICE METHODS ====================

func (s *CategoryGRPCServer) CreateCategory(ctx context.Context, req *pb.CreateCategoryRequest) (*pb.CreateCategoryResponse, error) {
	createReq := &models.CreateCategoryRequest{
		Name: req.Name,
	}

	category, err := s.categoryService.CreateCategory(ctx, createReq)
	if err != nil {
		if strings.Contains(err.Error(), "already exists") {
			return nil, status.Errorf(codes.AlreadyExists, err.Error())
		}
		return nil, status.Errorf(codes.Internal, "failed to create category: %v", err)
	}

	return &pb.CreateCategoryResponse{
		Category: categoryResponseToProto(category),
	}, nil
}

func (s *CategoryGRPCServer) GetCategory(ctx context.Context, req *pb.GetCategoryRequest) (*pb.GetCategoryResponse, error) {
	category, err := s.categoryService.GetCategory(ctx, req.Id)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			return nil, status.Errorf(codes.NotFound, "category with id %s not found", req.Id)
		}
		return nil, status.Error(codes.Internal, "failed to get category")
	}
	return &pb.GetCategoryResponse{Category: categoryResponseToProto(category)}, nil
}

func (s *CategoryGRPCServer) UpdateCategory(ctx context.Context, req *pb.UpdateCategoryRequest) (*pb.UpdateCategoryResponse, error) {
	updateReq := &models.UpdateCategoryRequest{
		Name: req.Name,
	}

	category, err := s.categoryService.UpdateCategory(ctx, req.Id, updateReq)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			return nil, status.Errorf(codes.NotFound, err.Error())
		}
		if strings.Contains(err.Error(), "already exists") {
			return nil, status.Errorf(codes.AlreadyExists, err.Error())
		}
		return nil, status.Errorf(codes.Internal, "failed to update category: %v", err)
	}

	return &pb.UpdateCategoryResponse{
		Category: categoryResponseToProto(category),
	}, nil
}

func (s *CategoryGRPCServer) DeleteCategory(ctx context.Context, req *pb.DeleteCategoryRequest) (*emptypb.Empty, error) {
	if err := s.categoryService.DeleteCategory(ctx, req.Id); err != nil {
		if strings.Contains(err.Error(), "not found") {
			return nil, status.Errorf(codes.NotFound, err.Error())
		}
		if strings.Contains(err.Error(), "contains") {
			return nil, status.Errorf(codes.FailedPrecondition, err.Error())
		}
		return nil, status.Errorf(codes.Internal, "failed to delete category: %v", err)
	}
	return &emptypb.Empty{}, nil
}

func (s *CategoryGRPCServer) ListCategories(ctx context.Context, req *pb.ListCategoriesRequest) (*pb.ListCategoriesResponse, error) {
	listResponse, err := s.categoryService.ListCategories(ctx)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to list categories")
	}

	return listCategoriesResponseToProto(listResponse), nil
}

// ==================== HELPER CONVERTERS ====================

// Helper: convert models.ProductResponse -> pb.Product
func productResponseToProto(p *models.ProductResponse) *pb.Product {
	if p == nil {
		return nil
	}
	return &pb.Product{
		Id:          p.ID,
		Name:        p.Name,
		Slug:        p.Slug,
		Description: p.Description,
		Price:       p.Price,
		CategoryId:  p.Category.ID,
		ImageUrl:    p.ImageURL,
		IsActive:    p.IsActive,
		CreatedAt:   timestamppb.New(p.CreatedAt),
		UpdatedAt:   timestamppb.New(p.UpdatedAt),
	}
}

// listProductsResponseToProto chuyển đổi từ models.ListProductsResponse sang pb.ListProductsResponse
func listProductsResponseToProto(resp *models.ListProductsResponse) *pb.ListProductsResponse {
	if resp == nil {
		return nil
	}

	protoProducts := make([]*pb.Product, len(resp.Products))
	for i, p := range resp.Products {
		tempProduct := p
		protoProducts[i] = productResponseToProto(&tempProduct)
	}

	return &pb.ListProductsResponse{
		Products:   protoProducts,
		TotalCount: resp.Total,
	}
}

// Helper: convert models.CategoryResponse -> pb.Category
func categoryResponseToProto(c *models.CategoryResponse) *pb.Category {
	if c == nil {
		return nil
	}
	return &pb.Category{
		Id:        c.ID,
		Name:      c.Name,
		Slug:      c.Slug,
		CreatedAt: timestamppb.New(c.CreatedAt),
		UpdatedAt: timestamppb.New(c.UpdatedAt),
	}
}

// listCategoriesResponseToProto chuyển đổi từ models.ListCategoriesResponse sang pb.ListCategoriesResponse
func listCategoriesResponseToProto(resp *models.ListCategoriesResponse) *pb.ListCategoriesResponse {
	if resp == nil {
		return nil
	}

	protoCategories := make([]*pb.Category, len(resp.Categories))
	for i, c := range resp.Categories {
		tempCategory := c
		protoCategories[i] = categoryResponseToProto(&tempCategory)
	}

	return &pb.ListCategoriesResponse{
		Categories: protoCategories,
	}
}
