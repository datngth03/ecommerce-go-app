// internal/product/delivery/grpc/product_grpc_server.go
package grpc

import (
	"context"
	"log" // Temporarily using log for errors

	"errors"

	"github.com/datngth03/ecommerce-go-app/internal/product/application"
	product_client "github.com/datngth03/ecommerce-go-app/pkg/client/product" // Generated gRPC client
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// ProductGRPCServer implements the product_client.ProductServiceServer interface.
type ProductGRPCServer struct {
	product_client.UnimplementedProductServiceServer // Embedded to satisfy all methods
	productService                                   application.ProductService
}

// NewProductGRPCServer creates a new instance of ProductGRPCServer.
func NewProductGRPCServer(svc application.ProductService) *ProductGRPCServer {
	return &ProductGRPCServer{
		productService: svc,
	}
}

// --- Product RPC Methods ---

// CreateProduct implements the gRPC CreateProduct method.
func (s *ProductGRPCServer) CreateProduct(ctx context.Context, req *product_client.CreateProductRequest) (*product_client.Product, error) {
	log.Printf("Received CreateProduct request for product: %s", req.GetName())

	// Extract user ID from context (assuming it's set by auth middleware)
	userID, err := s.getUserFromContext(ctx)
	if err != nil {
		log.Printf("Error extracting user from context: %v", err)
		return nil, status.Errorf(codes.Unauthenticated, "user not authenticated")
	}

	// Convert protobuf request to command
	cmd := application.CreateProductCommand{
		BrandID:        req.BrandId,
		Name:           req.Name,
		Slug:           req.Slug,
		Description:    req.Description,
		CategoryIDs:    req.CategoryIds,
		TagIDs:         req.TagIds,
		Images:         convertProtoImagesToCommand(req.Images),
		Specifications: convertProtoSpecsToCommand(req.Specifications),
		CreatedBy:      userID, // From context
	}

	// Call application service with command
	domainProduct, err := s.productService.CreateProduct(ctx, cmd)
	if err != nil {
		log.Printf("Error creating product: %v", err)
		return nil, err
	}

	// Convert domain model back to protobuf
	protoProduct := s.convertDomainToProto(domainProduct)

	log.Printf("Successfully created product with ID: %s", protoProduct.Id)
	return protoProduct, nil
}

// Helper functions for conversion
func convertProtoImagesToCommand(protoImages []*product_client.CreateProductImageRequest) []application.CreateProductImageCommand {
	if protoImages == nil {
		return nil
	}

	images := make([]application.CreateProductImageCommand, len(protoImages))
	for i, img := range protoImages {
		images[i] = application.CreateProductImageCommand{
			URL:       img.Url,
			IsPrimary: img.IsPrimary,
		}
	}
	return images
}

func convertProtoSpecsToCommand(protoSpecs []*product_client.CreateProductSpecificationRequest) []application.CreateProductSpecificationCommand {
	if protoSpecs == nil {
		return nil
	}

	specs := make([]application.CreateProductSpecificationCommand, len(protoSpecs))
	for i, spec := range protoSpecs {
		specs[i] = application.CreateProductSpecificationCommand{
			Value:       spec.Value,
			AttributeID: spec.AttributeId,
		}
	}
	return specs
}

func (s *ProductGRPCServer) convertDomainToProto(domain *application.CreateProductResult) *product_client.Product {
	proto := &product_client.Product{
		Id:          domain.ProductID,
		Name:        domain.Name,
		Slug:        domain.Slug,
		Description: domain.Description,
		Rating:      domain.Rating,
		ReviewCount: domain.ReviewCount,
		CreatedAt:   timestamppb.New(domain.CreatedAt),
		UpdatedAt:   timestamppb.New(domain.UpdatedAt),
	}

	// Convert brand if present
	// if domain.Brand != nil {
	// 	proto.Brand = &product_client.Brand{
	// 		Id:   domain.Brand.ID,
	// 		Name: domain.Brand.Name,
	// 		// ... other brand fields
	// 	}
	// }

	// Convert categories, images, etc.
	// ... similar conversion logic

	return proto
}

func (s *ProductGRPCServer) getUserFromContext(ctx context.Context) (string, error) {
	// Extract user ID from JWT token or metadata
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return "", errors.New("no metadata found")
	}

	userIDs := md.Get("user-id")
	if len(userIDs) == 0 {
		return "", errors.New("user-id not found in metadata")
	}

	return userIDs[0], nil
}

// func (s *ProductGRPCServer) handleServiceError(err error) error {
// 	// Convert domain errors to gRPC status codes
// 	switch {
// 	case errors.Is(err, ErrProductNotFound):
// 		return status.Errorf(codes.NotFound, err.Error())
// 	case errors.Is(err, ErrInvalidInput):
// 		return status.Errorf(codes.InvalidArgument, err.Error())
// 	case errors.Is(err, ErrUnauthorized):
// 		return status.Errorf(codes.PermissionDenied, err.Error())
// 	default:
// 		return status.Errorf(codes.Internal, "internal server error")
// 	}
// }
