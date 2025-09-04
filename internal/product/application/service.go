// internal/product/application/service.go
package application

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/datngth03/ecommerce-go-app/internal/product/domain"
	"github.com/datngth03/ecommerce-go-app/internal/product/infrastructure/messaging"
	"github.com/google/uuid"
	"github.com/gosimple/slug"
)

// ProductService defines the application service interface for product-related operations.
type ProductService interface {
	// Core Product operations
	CreateProduct(ctx context.Context, cmd CreateProductCommand) (*CreateProductResult, error)
	GetProduct(ctx context.Context, query GetProductQuery) (*GetProductResult, error)
	UpdateProduct(ctx context.Context, cmd UpdateProductCommand) (*UpdateProductResult, error)
	DeleteProduct(ctx context.Context, cmd DeleteProductCommand) error
	ListProducts(ctx context.Context, query ListProductsQuery) (*ListProductsResult, error)

	// Category relations
	AddProductCategories(ctx context.Context, cmd AddProductCategoriesCommand) error
	RemoveProductCategories(ctx context.Context, cmd RemoveProductCategoriesCommand) error
	GetProductCategories(ctx context.Context, query GetProductCategoriesQuery) ([]*CategorySummary, error)

	// Tag relations
	AddProductTags(ctx context.Context, cmd AddProductTagsCommand) error
	RemoveProductTags(ctx context.Context, cmd RemoveProductTagsCommand) error
	GetProductTags(ctx context.Context, query GetProductTagsQuery) ([]*TagSummary, error)
}

// productService implements ProductService interface
type productService struct {
	productRepo  domain.ProductRepository
	categoryRepo domain.CategoryRepository
	brandRepo    domain.BrandRepository
	tagRepo      domain.TagRepository
	// specificationAttributeRepo    domain.SpecificationAttributeRepository
	// variantAttributeRepo          domain.VariantAttributeRepository
	// variantAttributeValueRepo     domain.VariantAttributeValueRepository
	eventPublisher messaging.ProductEventPublisher
}

// NewProductService creates a new instance of ProductService.
func NewProductService(
	productRepo domain.ProductRepository,
	categoryRepo domain.CategoryRepository,
	brandRepo domain.BrandRepository,
	tagRepo domain.TagRepository,
	// specificationAttributeRepo    domain.SpecificationAttributeRepository,
	// variantAttributeRepo          domain.VariantAttributeRepository,
	// variantAttributeValueRepo     domain.VariantAttributeValueRepository,
	eventPublisher messaging.ProductEventPublisher,
) ProductService {
	return &productService{
		productRepo:  productRepo,
		categoryRepo: categoryRepo,
		brandRepo:    brandRepo,
		tagRepo:      tagRepo,
		// specificationAttributeRepo:    specificationAttributeRepo,
		// variantAttributeRepo:          variantAttributeRepo,
		// variantAttributeValueRepo:     variantAttributeValueRepo,
		eventPublisher: eventPublisher,
	}
}

// ================================================================ Product Use Cases ===============================================================

// CreateProduct creates a new product from command
func (s *productService) CreateProduct(ctx context.Context, cmd CreateProductCommand) (*CreateProductResult, error) {
	// Validate command
	if err := cmd.Validate(); err != nil {
		return nil, fmt.Errorf("command validation failed: %w", err)
	}

	// Convert command to domain entity
	product, err := s.convertCreateCommandToDomain(cmd)
	if err != nil {
		return nil, fmt.Errorf("failed to convert command to domain: %w", err)
	}

	// Business validation
	if err := s.validateProductForCreation(ctx, product); err != nil {
		return nil, fmt.Errorf("business validation failed: %w", err)
	}

	// Generate slug if not provided
	if product.Slug == "" {
		product.Slug, err = s.generateSlug(ctx, product.Name)
		if err != nil {
			return nil, fmt.Errorf("generate slug fail: %w", err)
		}
	}

	// Validate brand exists if provided
	if product.BrandID != nil && *product.BrandID != "" {
		brand, err := s.brandRepo.GetByID(*product.BrandID)
		if err != nil {
			return nil, fmt.Errorf("brand not found: %w", err)
		}
		product.Brand = brand
	}

	// Validate slug uniqueness
	if err := s.validateSlugUniqueness(ctx, product.Slug, product.ID); err != nil {
		return nil, err
	}

	// Save product to repository
	if err := s.productRepo.Create(product); err != nil {
		return nil, fmt.Errorf("failed to create product: %w", err)
	}

	// Add categories if provided
	categoriesAdded := 0
	if len(cmd.CategoryIDs) > 0 {
		if err := s.productRepo.AddCategories(product.ID, cmd.CategoryIDs); err != nil {
			log.Printf("Failed to add categories to product %s: %v", product.ID, err)
		} else {
			categoriesAdded = len(cmd.CategoryIDs)
		}
	}

	// Add tags if provided
	tagsAdded := 0
	if len(cmd.TagIDs) > 0 {
		if err := s.productRepo.AddTags(product.ID, cmd.TagIDs); err != nil {
			log.Printf("Failed to add tags to product %s: %v", product.ID, err)
		} else {
			tagsAdded = len(cmd.TagIDs)
		}
	}

	// Publish product created event
	if err := s.eventPublisher.PublishProductCreated(ctx, product); err != nil {
		log.Printf("Failed to publish product created event: %v", err)
	}

	// Return result
	result := &CreateProductResult{
		ProductID:       product.ID,
		Name:            product.Name,
		Slug:            product.Slug,
		CreatedAt:       product.CreatedAt,
		CategoriesAdded: categoriesAdded,
		TagsAdded:       tagsAdded,
		ImagesAdded:     len(product.Images),
		SpecsAdded:      len(product.Specifications),
	}

	return result, nil
}

// GetProduct retrieves a product based on query
func (s *productService) GetProduct(ctx context.Context, query GetProductQuery) (*GetProductResult, error) {
	// Validate query
	if query.ID == nil && query.Slug == nil {
		return nil, errors.New("either ID or slug must be provided")
	}

	if query.RequestedBy == "" {
		return nil, errors.New("requested_by is required")
	}

	// Set defaults
	query.SetDefaults()

	// Lấy sản phẩm chính
	product, err := s.productRepo.GetByID(*query.ID)
	if err != nil {
		// Xử lý lỗi, ví dụ: domain.ErrProductNotFound
		return nil, err
	}

	// Lazy loading: Lấy thêm dữ liệu liên quan nếu được yêu cầu
	if query.IncludeBrand {
		brand, err := s.brandRepo.GetByID(*product.BrandID)
		if err != nil {
			return nil, err
		}
		product.Brand = brand
	}

	if query.IncludeCategories {
		categories, err := s.productRepo.GetCategories(product.ID)
		if err != nil {
			return nil, err
		}
		product.Categories = categories
	}

	if query.IncludeTags {
		tags, err := s.productRepo.GetTags(product.ID)
		if err != nil {
			return nil, err
		}
		product.Tags = tags
	}

	// if query.IncludeImages {
	// 	images, err := s.productRepo.GetImagesByProductID(ctx, product.ID)
	// 	if err != nil {
	// 		return nil, err
	// 	}
	// 	product.Images = images
	// }

	// if query.IncludeSpecifications {
	// 	specs, err := s.productRepo.GetSpecificationsByProductID(ctx, product.ID)
	// 	if err != nil {
	// 		return nil, err
	// 	}
	// 	product.Specifications = specs
	// }

	// if query.IncludeVariants {
	// 	variants, err := s.productRepo.GetVariantsByProductID(ctx, product.ID)
	// 	if err != nil {
	// 		return nil, err
	// 	}
	// 	product.Variants = variants
	// }

	// Convert to detailed result
	productDetail, err := s.convertDomainToProductDetail(ctx, product)
	if err != nil {
		return nil, fmt.Errorf("failed to convert product to detail: %w", err)
	}

	return &GetProductResult{
		Product: productDetail,
	}, nil
}

// UpdateProduct updates an existing product from command
func (s *productService) UpdateProduct(ctx context.Context, cmd UpdateProductCommand) (*UpdateProductResult, error) {
	// Validate command
	if err := cmd.Validate(); err != nil {
		return nil, fmt.Errorf("command validation failed: %w", err)
	}

	// Get existing product
	existingProduct, err := s.productRepo.GetByID(cmd.ID)
	if err != nil {
		return nil, fmt.Errorf("product not found: %w", err)
	}

	// Convert command to domain entity
	product, err := s.convertUpdateCommandToDomain(ctx, cmd)
	if err != nil {
		return nil, fmt.Errorf("failed to convert command to domain: %w", err)
	}

	// Business validation
	if err := s.validateProductForUpdate(ctx, cmd); err != nil {
		return nil, fmt.Errorf("business validation failed: %w", err)
	}

	// Validate brand exists if changed
	if product.BrandID != nil && *product.BrandID != "" {
		if existingProduct.BrandID == nil || *product.BrandID != *existingProduct.BrandID {
			brand, err := s.brandRepo.GetByID(*product.BrandID)
			if err != nil {
				return nil, fmt.Errorf("brand not found: %w", err)
			}
			product.Brand = brand
		}
	}

	// Validate slug uniqueness if changed
	if product.Slug != existingProduct.Slug {
		if err := s.validateSlugUniqueness(ctx, product.Slug, product.ID); err != nil {
			return nil, err
		}
	}

	// Save updated product
	if err := s.productRepo.Update(product); err != nil {
		return nil, fmt.Errorf("failed to update product: %w", err)
	}

	// Publish product updated event
	if err := s.eventPublisher.PublishProductUpdated(ctx, product); err != nil {
		log.Printf("Failed to publish product updated event: %v", err)
	}

	// Return result
	result := &UpdateProductResult{
		ProductID:     product.ID,
		Name:          product.Name,
		Slug:          product.Slug,
		UpdatedAt:     product.UpdatedAt,
		ImagesUpdated: len(product.Images),
		SpecsUpdated:  len(product.Specifications),
	}

	return result, nil
}

// DeleteProduct deletes a product from command
func (s *productService) DeleteProduct(ctx context.Context, cmd DeleteProductCommand) error {
	// Validate command
	if cmd.ID == "" {
		return errors.New("product ID is required")
	}
	if cmd.DeletedBy == "" {
		return errors.New("deleted_by is required")
	}

	// Check if product exists
	product, err := s.productRepo.GetByID(cmd.ID)
	if err != nil {
		return fmt.Errorf("product not found: %w", err)
	}

	// Business validation for deletion
	// if err := s.validateProductForDeletion(product); err != nil {
	// 	return fmt.Errorf("deletion validation failed: %w", err)
	// }

	// Delete the product
	if err := s.productRepo.Delete(cmd.ID); err != nil {
		return fmt.Errorf("failed to delete product: %w", err)
	}

	// Publish product deleted event
	if err := s.eventPublisher.PublishProductDeleted(ctx, product.ID); err != nil {
		log.Printf("Failed to publish product deleted event: %v", err)
	}

	return nil
}

// ListProducts retrieves a list of products based on query
func (s *productService) ListProducts(ctx context.Context, query ListProductsQuery) (*ListProductsResult, error) {
	// Validate query
	if query.RequestedBy == "" {
		return nil, errors.New("requested_by is required")
	}

	// Set defaults
	query.SetDefaults()

	// Convert query to domain filter
	filter, err := s.convertQueryToDomainFilter(query)
	if err != nil {
		return nil, fmt.Errorf("failed to convert query to filter: %w", err)
	}

	// s.applyBusinessRulesToFilter(&filter)

	// Get products from repository
	products, total, err := s.productRepo.List(*filter)
	if err != nil {
		return nil, fmt.Errorf("failed to list products: %w", err)
	}

	// Convert to summaries
	productSummaries := make([]*ProductSummary, len(products))
	for i, product := range products {
		productSummaries[i] = s.convertDomainToProductSummary(product)
	}

	// Calculate pagination info
	hasNext := query.Offset+query.Limit < total
	hasPrev := query.Offset > 0

	result := &ListProductsResult{
		Products: productSummaries,
		Total:    total,
		Limit:    query.Limit,
		Offset:   query.Offset,
		HasNext:  hasNext,
		HasPrev:  hasPrev,
	}

	return result, nil
}

// validateSlugUniqueness
func (s *productService) validateSlugUniqueness(_ context.Context, slug string, productID string) error {
	// 1. Kiểm tra slug có rỗng không
	if slug == "" {
		return errors.New("product slug cannot be empty")
	}

	// 2. Chuẩn hóa slug để đảm bảo tính nhất quán (ví dụ: chuyển về chữ thường)
	slug = strings.ToLower(slug)

	// 3. Tìm kiếm sản phẩm bằng slug trong cơ sở dữ liệu
	existingProduct, err := s.productRepo.GetBySlug(slug)

	// 4. Kiểm tra kết quả
	if err == nil && existingProduct != nil {
		// Nếu tìm thấy một sản phẩm có cùng slug
		// Kiểm tra xem đó có phải là sản phẩm đang được cập nhật không
		if existingProduct.ID != productID {
			return fmt.Errorf("slug '%s' is already in use by another product", slug)
		}
	} else if !errors.Is(err, domain.ErrNotFound) {
		// Xử lý các lỗi khác ngoài lỗi "không tìm thấy"
		return fmt.Errorf("failed to check slug uniqueness: %w", err)
	}

	return nil
}

// validateProductForCreation performs business validation on a product before creation.
func (s *productService) validateProductForCreation(_ context.Context, product *domain.Product) error {
	// Kiểm tra xem tên sản phẩm đã tồn tại trong cơ sở dữ liệu hay chưa
	existingProduct, err := s.productRepo.GetByID(product.ID)
	if err != nil {
		// Xử lý lỗi từ repository, ví dụ: lỗi kết nối database
		return fmt.Errorf("failed to check for existing product: %w", err)
	}

	if existingProduct != nil {
		// Nếu tên sản phẩm đã tồn tại, trả về lỗi validation
		return domain.ErrProductAlreadyExists
	}

	// Nếu tất cả các quy tắc nghiệp vụ đều hợp lệ
	return nil
}

// Validate product for update
func (s *productService) validateProductForUpdate(ctx context.Context, cmd UpdateProductCommand) error {
	// 1. Lấy sản phẩm hiện tại từ kho lưu trữ
	currentProduct, err := s.productRepo.GetByID(cmd.ID)
	if err != nil {
		return fmt.Errorf("product with ID %s not found for update: %w", cmd.ID, err)
	}

	// 2. Xác thực tính duy nhất của slug nếu slug được cập nhật
	if cmd.Slug != "" && cmd.Slug != currentProduct.Slug {
		if err := s.validateSlugUniqueness(ctx, cmd.Slug, cmd.ID); err != nil {
			return err
		}
	}

	// 3. Xác thực các quy tắc nghiệp vụ khác

	// 4. (Tùy chọn) Xác thực brand
	// if cmd.BrandID != "" && cmd.BrandID != currentProduct.BrandID {
	// 	_, err := s.brandRepo.GetByID(cmd.BrandID)
	// 	if err != nil {
	// 		return fmt.Errorf("brand with ID %s not found: %w", cmd.BrandID, err)
	// 	}
	// }

	return nil
}

// Validate product for deletion
// func (s *productService) validateProductForDeletion(_ context.Context, _ string) error {
// 	// 1. Kiểm tra sản phẩm có tồn tại không
// 	// product, err := s.productRepo.GetByID(productID)
// 	// if err != nil {
// 	// return fmt.Errorf("product with ID %s not found: %w", productID, err)
// 	// }

// 	// 2. Kiểm tra các quy tắc nghiệp vụ
// 	// Ví dụ: Không cho phép xóa sản phẩm đang có đơn hàng đang chờ xử lý
// 	// Đây chỉ là một ví dụ, logic thực tế sẽ phức tạp hơn.
// 	// if product.HasPendingOrders { // Giả sử domain.Product có trường này
// 	// return errors.New("cannot delete product; it has pending orders")
// 	// }

// 	// Ví dụ: Không cho phép xóa sản phẩm đang nằm trong chiến dịch quảng cáo
// 	// if product.IsPartOfActiveCampaign {
// 	// return errors.New("cannot delete product; it is part of an active marketing campaign")
// 	// }

// 	// 3. (Tùy chọn) Kiểm tra các mối quan hệ phức tạp
// 	// Nếu sản phẩm có các mối quan hệ phức tạp với các thực thể khác (ví dụ: biến thể, đánh giá,...)
// 	// bạn có thể thêm các kiểm tra để đảm bảo việc xóa sản phẩm không gây ra lỗi logic.
// 	// Ví dụ:
// 	// variants, err := s.productRepo.GetProductVariants(productID)
// 	// if err != nil || len(variants) > 0 { ... }

// 	return nil
// }

// Convert CreateProductCommand to domain.Product
func (s *productService) convertCreateCommandToDomain(cmd CreateProductCommand) (*domain.Product, error) {
	productID := uuid.New().String()

	// Create a new Product instance with mapped data
	product := &domain.Product{
		ID:          productID,
		BrandID:     cmd.BrandID,
		Name:        cmd.Name,
		Slug:        cmd.Slug,
		Description: cmd.Description,
		Rating:      0, // New products start with a rating of 0
		ReviewCount: 0, // And a review count of 0
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	// Relations are handled in separate logic and not part of this conversion.
	// We only map the basic fields. The associated entities (Images, Variants, etc.)
	// will be created or linked in subsequent steps of the service.

	return product, nil
}

// Convert UpdateProductCommand to domain.Product
func (s *productService) convertUpdateCommandToDomain(_ context.Context, cmd UpdateProductCommand) (*domain.Product, error) {
	product := &domain.Product{
		ID:        cmd.ID,
		Name:      cmd.Name,
		Slug:      cmd.Slug,
		UpdatedAt: time.Now(),
	}

	// Cập nhật các trường nullable
	if cmd.Description != nil {
		product.Description = cmd.Description
	}

	// Cập nhật BrandID, nếu có
	if cmd.BrandID != "" {
		product.BrandID = &cmd.BrandID
	} else {
		// Set to nil if the brand ID is an empty string, meaning the brand relationship is to be removed.
		// Note: This logic assumes that an empty string in the command means removing the brand.
		product.BrandID = nil
	}

	// Xử lý các mối quan hệ (Images, Specifications)
	// Lưu ý: Các mối quan hệ này thường được xử lý riêng biệt trong lớp repository.
	// Hàm này chỉ tập trung vào việc tạo đối tượng Product chính.
	// Logic cập nhật sẽ phức tạp hơn, có thể cần thêm các hàm để
	// xóa, thêm, hoặc cập nhật các đối tượng liên quan.

	// Ví dụ, bạn có thể tạo các đối tượng con từ command
	// product.Images = make([]*domain.ProductImage, len(cmd.Images))
	// for i, imgCmd := range cmd.Images {
	//		product.Images[i] = &domain.ProductImage{
	//			ID:  imgCmd.ID,
	//			URL: imgCmd.URL,
	//			IsDefault: imgCmd.IsDefault,
	//		}
	// }

	// Tương tự với Specifications
	// product.Specifications = make([]*domain.ProductSpecification, len(cmd.Specifications))
	// ...

	return product, nil
}

// convertDomainToProductDetail chuyển đổi một đối tượng domain.Product thành
func (s *productService) convertDomainToProductDetail(ctx context.Context, product *domain.Product) (*ProductDetail, error) {
	if product == nil {
		return nil, nil
	}

	result := &ProductDetail{
		ID:          product.ID,
		Name:        product.Name,
		Slug:        product.Slug,
		Description: product.Description,
		Rating:      product.Rating,
		ReviewCount: product.ReviewCount,
		CreatedAt:   product.CreatedAt,
		UpdatedAt:   product.UpdatedAt,
		BrandID:     product.BrandID,
	}

	// Chuyển đổi các mối quan hệ đã được tải (nếu có)
	if product.Brand != nil {
		brandSummary, err := s.convertDomainToBrandSummary(ctx, product.Brand)
		if err != nil {
			return nil, err
		}
		result.Brand = brandSummary
	}

	if product.Categories != nil {
		categorySummaries, err := s.convertCategoriesToCategorySummaries(ctx, product.Categories)
		if err != nil {
			return nil, err
		}
		result.Categories = categorySummaries
	}

	if product.Tags != nil {
		tagSummaries, err := s.convertTagsToTagSummaries(ctx, product.Tags)
		if err != nil {
			return nil, err
		}
		result.Tags = tagSummaries
	}

	if product.Images != nil {
		imageSummaries, err := s.convertImagesToImageSummaries(ctx, product.Images)
		if err != nil {
			return nil, err
		}
		result.Images = imageSummaries
	}

	if product.Specifications != nil {
		specSummaries, err := s.convertSpecificationsToSpecSummaries(ctx, product.Specifications)
		if err != nil {
			return nil, err
		}
		result.Specifications = specSummaries
	}

	if product.Variants != nil {
		variantSummaries, err := s.convertVariantsToVariantSummaries(ctx, product.Variants)
		if err != nil {
			return nil, err
		}
		result.Variants = variantSummaries
	}

	return result, nil
}

// convertDomainToBrandSummary maps a domain.Brand to a BrandSummary DTO.
func (s *productService) convertDomainToBrandSummary(_ context.Context, brand *domain.Brand) (*BrandSummary, error) {
	if brand == nil {
		return nil, nil
	}

	return &BrandSummary{
		ID:   brand.ID,
		Name: brand.Name,
		Slug: brand.Slug,
	}, nil
}

// convertCategoriesToCategorySummaries maps a slice of domain.Category
func (s *productService) convertCategoriesToCategorySummaries(_ context.Context, categories []*domain.Category) ([]*CategorySummary, error) {
	if categories == nil {
		return nil, nil
	}

	var summaries []*CategorySummary
	for _, cat := range categories {
		summaries = append(summaries, &CategorySummary{
			ID:   cat.ID,
			Name: cat.Name,
			Slug: cat.Slug,
		})
	}
	return summaries, nil
}

// convertTagsToTagSummaries maps a slice of domain.Tag
func (s *productService) convertTagsToTagSummaries(_ context.Context, tags []*domain.Tag) ([]*TagSummary, error) {
	if tags == nil {
		return nil, nil
	}

	var summaries []*TagSummary
	for _, tag := range tags {
		summaries = append(summaries, &TagSummary{
			ID:   tag.ID,
			Name: tag.Name,
			Slug: tag.Slug,
		})
	}
	return summaries, nil
}

// convertImagesToImageSummaries maps a slice of domain.Image
func (s *productService) convertImagesToImageSummaries(_ context.Context, images []*domain.ProductImage) ([]*ImageSummary, error) {
	if images == nil {
		return nil, nil
	}

	var summaries []*ImageSummary
	for _, img := range images {
		summaries = append(summaries, &ImageSummary{
			ID:  img.ID,
			URL: img.URL,
		})
	}
	return summaries, nil
}

// convertSpecificationsToSpecSummaries maps a slice of domain.Specification
func (s *productService) convertSpecificationsToSpecSummaries(_ context.Context, specs []*domain.ProductSpecification) ([]*SpecificationSummary, error) {
	if specs == nil {
		return nil, nil
	}

	var summaries []*SpecificationSummary
	for _, spec := range specs {
		// Kiểm tra xem trường Attribute có nil không trước khi truy cập
		attributeName := ""
		if spec.Attribute != nil {
			attributeName = spec.Attribute.Name
		}

		summaries = append(summaries, &SpecificationSummary{
			ID:            spec.ID,
			AttributeID:   spec.AttributeID,
			AttributeName: attributeName, // Lấy AttributeName từ trường quan hệ Attribute
			Value:         spec.Value,
		})
	}
	return summaries, nil
}

// convertVariantsToVariantSummaries maps a slice of domain.Variant
func (s *productService) convertVariantsToVariantSummaries(_ context.Context, variants []*domain.ProductVariant) ([]*VariantSummary, error) {
	if variants == nil {
		return nil, nil
	}

	var summaries []*VariantSummary
	for _, variant := range variants {
		summaries = append(summaries, &VariantSummary{
			ID:            variant.ID,
			SKU:           variant.SKU,
			Price:         variant.Price,
			OriginalPrice: variant.OriginalPrice, // Ánh xạ từ OriginalPrice
			Discount:      variant.Discount,      // Ánh xạ trường Discount
		})
	}
	return summaries, nil
}

func (s *productService) generateSlug(_ context.Context, name string) (string, error) {
	// Bước 1: Tạo slug cơ bản từ tên sản phẩm bằng thư viện gosimple/slug.
	// Ví dụ: "Áo thun nam" -> "ao-thun-nam"
	baseSlug := slug.Make(name)
	currentSlug := baseSlug
	suffix := 1

	// Bước 2: Lặp lại để kiểm tra tính duy nhất của slug.
	for {
		// Gọi repository để tìm kiếm sản phẩm bằng slug hiện tại.
		existingProduct, err := s.productRepo.GetBySlug(currentSlug)

		// Xử lý lỗi từ database. Nếu lỗi không phải là "không tìm thấy", chúng ta trả về lỗi.
		if err != nil && err != domain.ErrNotFound {
			return "", fmt.Errorf("failed to check slug uniqueness: %w", err)
		}

		// Bước 3: Nếu không tìm thấy sản phẩm nào với slug này,
		// nghĩa là slug là duy nhất, chúng ta có thể sử dụng nó.
		if existingProduct == nil {
			return currentSlug, nil
		}

		// Bước 4: Nếu đã tìm thấy một sản phẩm với slug này,
		// chúng ta thêm một số thứ tự (ví dụ: -1, -2) vào cuối và thử lại.
		currentSlug = fmt.Sprintf("%s-%d", baseSlug, suffix)
		suffix++
	}
}

// ============================================================== Category Relations ======================================================

// AddProductCategories adds categories to a product
func (s *productService) AddProductCategories(ctx context.Context, cmd AddProductCategoriesCommand) error {
	// Validate command
	if cmd.ProductID == "" {
		return errors.New("product ID is required")
	}
	if len(cmd.CategoryIDs) == 0 {
		return errors.New("at least one category ID is required")
	}
	if cmd.AddedBy == "" {
		return errors.New("added_by is required")
	}

	// Validate product exists
	_, err := s.productRepo.GetByID(cmd.ProductID)
	if err != nil {
		return fmt.Errorf("product not found: %w", err)
	}

	// Validate all categories exist
	for _, categoryID := range cmd.CategoryIDs {
		_, err := s.categoryRepo.GetByID(categoryID)
		if err != nil {
			return fmt.Errorf("category %s not found: %w", categoryID, err)
		}
	}

	// Business validation
	if err := s.validateCategoryAssignment(cmd.ProductID, cmd.CategoryIDs); err != nil {
		return fmt.Errorf("category assignment validation failed: %w", err)
	}

	// Add categories
	if err := s.productRepo.AddCategories(cmd.ProductID, cmd.CategoryIDs); err != nil {
		return fmt.Errorf("failed to add categories to product: %w", err)
	}

	return nil
}

// RemoveProductCategories removes categories from a product
func (s *productService) RemoveProductCategories(ctx context.Context, cmd RemoveProductCategoriesCommand) error {
	// Validate command
	if cmd.ProductID == "" {
		return errors.New("product ID is required")
	}
	if len(cmd.CategoryIDs) == 0 {
		return errors.New("at least one category ID is required")
	}
	if cmd.RemovedBy == "" {
		return errors.New("removed_by is required")
	}

	// Validate product exists
	_, err := s.productRepo.GetByID(cmd.ProductID)
	if err != nil {
		return fmt.Errorf("product not found: %w", err)
	}

	// Remove categories
	if err := s.productRepo.RemoveCategories(cmd.ProductID, cmd.CategoryIDs); err != nil {
		return fmt.Errorf("failed to remove categories from product: %w", err)
	}

	return nil
}

// GetProductCategories retrieves categories for a product
func (s *productService) GetProductCategories(ctx context.Context, query GetProductCategoriesQuery) ([]*CategorySummary, error) {
	// Validate query
	if query.ProductID == "" {
		return nil, errors.New("product ID is required")
	}
	if query.RequestedBy == "" {
		return nil, errors.New("requested_by is required")
	}

	// Validate product exists
	_, err := s.productRepo.GetByID(query.ProductID)
	if err != nil {
		return nil, fmt.Errorf("product not found: %w", err)
	}

	// Get categories
	categories, err := s.productRepo.GetCategories(query.ProductID)
	if err != nil {
		return nil, fmt.Errorf("failed to get product categories: %w", err)
	}

	// Convert to summaries
	categorySummaries := make([]*CategorySummary, len(categories))
	for i, category := range categories {
		categorySummaries[i] = &CategorySummary{
			ID:          category.ID,
			Name:        category.Name,
			Slug:        category.Slug,
			Description: category.Description,
			Image:       category.Image,
			ParentID:    category.ParentID,
		}
	}

	return categorySummaries, nil
}

// validateCategoryAssignment performs business validation for category assignment
func (s *productService) validateCategoryAssignment(productID string, categoryIDs []string) error {
	// Get current product categories
	currentCategories, err := s.productRepo.GetCategories(productID)
	if err != nil {
		return fmt.Errorf("failed to get current product categories: %w", err)
	}

	// Get category details for new categories
	newCategories := make([]*domain.Category, 0, len(categoryIDs))
	for _, categoryID := range categoryIDs {
		category, err := s.categoryRepo.GetByID(categoryID)
		if err != nil {
			return fmt.Errorf("failed to get category %s: %w", categoryID, err)
		}
		newCategories = append(newCategories, category)
	}

	// 2. Check for duplicate categories
	if err := s.validateNoDuplicateCategories(currentCategories, newCategories); err != nil {
		return err
	}

	// 3. Check maximum categories limit
	if err := s.validateMaxCategoriesLimit(currentCategories, newCategories); err != nil {
		return err
	}

	// 4. Check category hierarchy rules
	if err := s.validateCategoryHierarchy(currentCategories, newCategories); err != nil {
		return err
	}

	return nil
}

// validateNoDuplicateCategories checks for duplicate category assignments
func (s *productService) validateNoDuplicateCategories(currentCategories []*domain.Category, newCategories []*domain.Category) error {
	currentCategoryMap := make(map[string]bool)
	for _, category := range currentCategories {
		currentCategoryMap[category.ID] = true
	}

	var duplicates []string
	for _, newCategory := range newCategories {
		if currentCategoryMap[newCategory.ID] {
			duplicates = append(duplicates, newCategory.Name)
		}
	}

	if len(duplicates) > 0 {
		return fmt.Errorf("categories already assigned to product: %s", strings.Join(duplicates, ", "))
	}

	return nil
}

// validateMaxCategoriesLimit checks maximum categories per product
func (s *productService) validateMaxCategoriesLimit(currentCategories []*domain.Category, newCategories []*domain.Category) error {
	const MaxCategoriesPerProduct = 5

	totalCategories := len(currentCategories) + len(newCategories)
	if totalCategories > MaxCategoriesPerProduct {
		return fmt.Errorf("cannot assign categories: would exceed maximum limit of %d categories per product (current: %d, adding: %d)",
			MaxCategoriesPerProduct, len(currentCategories), len(newCategories))
	}
	return nil
}

// validateCategoryHierarchy checks category hierarchy rules
func (s *productService) validateCategoryHierarchy(currentCategories []*domain.Category, newCategories []*domain.Category) error {
	// Combine all categories for validation
	allCategories := make([]*domain.Category, 0, len(currentCategories)+len(newCategories))
	allCategories = append(allCategories, currentCategories...)
	allCategories = append(allCategories, newCategories...)

	// Check: Cannot assign parent and child categories to same product
	categoryMap := make(map[string]*domain.Category)
	for _, category := range allCategories {
		categoryMap[category.ID] = category
	}

	for _, category := range allCategories {
		if *category.ParentID != "" {
			if _, hasParent := categoryMap[*category.ParentID]; hasParent {
				parentCategory := categoryMap[*category.ParentID]
				return fmt.Errorf("cannot assign both parent category '%s' and child category '%s' to same product",
					parentCategory.Name, category.Name)
			}
		}
	}

	// Check: Cannot assign categories from different root branches (business rule)
	rootCategories := make(map[string]string) // category_id -> root_category_id
	for _, category := range allCategories {
		rootID, err := s.getRootCategoryID(category)
		if err != nil {
			return fmt.Errorf("failed to get root category for %s: %w", category.Name, err)
		}
		if existingRootID, exists := rootCategories[category.ID]; exists && existingRootID != rootID {
			return fmt.Errorf("cannot assign categories from different root branches")
		}
		rootCategories[category.ID] = rootID
	}

	return nil
}

// getRootCategoryID finds the root category ID for a given category
func (s *productService) getRootCategoryID(category *domain.Category) (string, error) {
	current := category
	for *current.ParentID != "" {
		parent, err := s.categoryRepo.GetByID(*current.ParentID)
		if err != nil {
			return "", fmt.Errorf("failed to get parent category %s: %w", *current.ParentID, err)
		}
		current = parent
	}
	return current.ID, nil
}

// convertDomainToProductSummary converts domain.Product to ProductSummary
func (s *productService) convertDomainToProductSummary(domainProduct *domain.Product) *ProductSummary {
	// Get tag count for the product
	tagCount := 0
	if tags, err := s.productRepo.GetTags(domainProduct.ID); err == nil {
		tagCount = len(tags)
	}

	// 	type ProductSummary struct {
	// 	ID            string    `json:"id"`
	// 	Name          string    `json:"name"`
	// 	Slug          string    `json:"slug"`
	// 	BrandID       *string   `json:"brand_id,omitempty"`
	// 	BrandName     *string   `json:"brand_name,omitempty"`
	// 	PrimaryImage  *string   `json:"primary_image,omitempty"`
	// 	CategoryCount int       `json:"category_count"`
	// 	TagCount      int       `json:"tag_count"`
	// 	ImageCount    int       `json:"image_count"`
	// 	SpecCount     int       `json:"specification_count"`
	// 	VariantCount  int       `json:"variant_count"`
	// 	Rating        float64   `json:"rating"`
	// 	ReviewCount   int32     `json:"review_count"`
	// 	CreatedAt     time.Time `json:"created_at"`
	// 	UpdatedAt     time.Time `json:"updated_at"`
	// }

	// 	type Product struct {
	// 	ID          string    `json:"id" db:"id"`
	// 	Name        string    `json:"name" db:"name"`
	// 	Slug        string    `json:"slug" db:"slug"`
	// 	BrandID     *string   `json:"brand_id,omitempty" db:"brand_id"`
	// 	Description *string   `json:"description,omitempty" db:"description"`
	// 	Rating      float64   `json:"rating" db:"rating"`
	// 	ReviewCount int32     `json:"review_count" db:"review_count"`
	// 	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	// 	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`

	// 	// Relations (not stored in database)
	// 	Brand          *Brand                  `json:"brand,omitempty" db:"-"`
	// 	Categories     []*Category             `json:"categories,omitempty" db:"-"`
	// 	Images         []*ProductImage         `json:"images,omitempty" db:"-"`
	// 	Variants       []*ProductVariant       `json:"variants,omitempty" db:"-"`
	// 	Tags           []*Tag                  `json:"tags,omitempty" db:"-"`
	// 	Specifications []*ProductSpecification `json:"specifications,omitempty" db:"-"`
	// }

	return &ProductSummary{
		ID:          domainProduct.ID,
		Name:        domainProduct.Name,
		Slug:        domainProduct.Slug,
		BrandID:     domainProduct.BrandID,
		Rating:      domainProduct.Rating,
		ReviewCount: domainProduct.ReviewCount,
		CreatedAt:   domainProduct.CreatedAt,
		UpdatedAt:   domainProduct.UpdatedAt,
		TagCount:    tagCount,
	}
}

// applyBusinessRulesToFilter applies business rules to filter based on user context
// func (s *productService) applyBusinessRulesToFilter(ctx context.Context, filter *DomainFilter, userID string) error {
// 	// Get user permissions/role
// 	user, err := s.userService.GetByID(ctx, userID)
// 	if err != nil {
// 		return fmt.Errorf("failed to get user context: %w", err)
// 	}

// 	// Apply role-based filtering
// 	switch user.Role {
// 	case "admin":
// 		// Admins can see all products, no additional filtering needed

// 	case "manager":
// 		// Managers can see products in their assigned categories
// 		managedCategories, err := s.userService.GetManagedCategories(ctx, userID)
// 		if err != nil {
// 			return fmt.Errorf("failed to get managed categories: %w", err)
// 		}

// 		// If no category filter specified, apply managed categories filter
// 		if len(filter.CategoryIDs) == 0 {
// 			categoryIDs := make([]string, len(managedCategories))
// 			for i, cat := range managedCategories {
// 				categoryIDs[i] = cat.ID
// 			}
// 			filter.CategoryIDs = categoryIDs
// 		} else {
// 			// Intersect requested categories with managed categories
// 			filter.CategoryIDs = s.intersectCategories(filter.CategoryIDs, managedCategories)
// 		}

// 	case "user":
// 		// Regular users can only see published products
// 		if len(filter.Status) == 0 {
// 			filter.Status = []string{"published"}
// 		} else {
// 			// Ensure only published products are included
// 			publishedOnly := []string{}
// 			for _, status := range filter.Status {
// 				if status == "published" {
// 					publishedOnly = append(publishedOnly, status)
// 				}
// 			}
// 			filter.Status = publishedOnly
// 		}

// 		// Users cannot see certain categories (e.g., internal, test)
// 		restrictedCategories := []string{"internal", "test", "draft"}
// 		filter.CategoryIDs = s.excludeCategories(filter.CategoryIDs, restrictedCategories)

// 	default:
// 		return fmt.Errorf("unknown user role: %s", user.Role)
// 	}

// 	// Apply business-specific filters

// 	// Filter by user's organization if multi-tenant
// 	if user.OrganizationID != "" {
// 		// This would require adding organization filtering to domain filter
// 		// filter.OrganizationID = user.OrganizationID
// 	}

// 	// Apply time-based restrictions
// 	if user.Role == "user" {
// 		// Users can only see products created in last year
// 		oneYearAgo := time.Now().AddDate(-1, 0, 0)
// 		if filter.CreatedAfter == nil || filter.CreatedAfter.Before(oneYearAgo) {
// 			filter.CreatedAfter = &oneYearAgo
// 		}
// 	}

// 	return nil
// }

// convertQueryToDomainFilter converts ListProductsQuery to DomainFilter
func (s *productService) convertQueryToDomainFilter(query ListProductsQuery) (*domain.ProductFilter, error) {
	// Validate and set defaults
	limit := query.Limit
	if limit <= 0 {
		limit = 20 // Default limit
	}
	if limit > 100 {
		limit = 100 // Max limit
	}

	offset := query.Offset
	if offset < 0 {
		offset = 0
	}

	// Set default sorting
	sortBy := query.SortBy
	if sortBy == "" {
		sortBy = "created_at"
	}

	sortOrder := strings.ToLower(query.SortOrder)
	if sortOrder != "asc" && sortOrder != "desc" {
		sortOrder = "desc"
	}

	// Validate sort field
	validSortFields := map[string]bool{
		"id":         true,
		"name":       true,
		"price":      true,
		"created_at": true,
		"updated_at": true,
		"status":     true,
	}
	if !validSortFields[sortBy] {
		return nil, fmt.Errorf("invalid sort field: %s", sortBy)
	}

	// Validate price range
	if query.MinPrice != nil && query.MaxPrice != nil {
		if *query.MinPrice > *query.MaxPrice {
			return nil, fmt.Errorf("min_price cannot be greater than max_price")
		}
		if *query.MinPrice < 0 {
			return nil, fmt.Errorf("min_price cannot be negative")
		}
	}

	// Validate date range
	if query.CreatedAfter != nil && query.CreatedBefore != nil {
		if query.CreatedAfter.After(*query.CreatedBefore) {
			return nil, fmt.Errorf("created_after cannot be after created_before")
		}
	}

	// 	type ProductFilter struct {
	// 	Search      string
	// 	CategoryIDs []string
	// 	BrandIDs    []string
	// 	TagIDs      []string
	// 	MinPrice    *float64
	// 	MaxPrice    *float64
	// 	MinRating   *float64
	// 	SortBy      string
	// 	SortOrder   SortOrder
	// 	Limit       int
	// 	Offset      int

	// 	// Include options
	// 	IncludeBrand          bool
	// 	IncludeCategories     bool
	// 	IncludeImages         bool
	// 	IncludeVariants       bool
	// 	IncludeTags           bool
	// 	IncludeSpecifications bool
	// }

	// Convert to domain filter
	domainFilter := &domain.ProductFilter{
		Search:      query.SearchQuery,
		CategoryIDs: query.CategoryIDs,
		BrandID:     query.BrandID,
		TagIDs:      query.TagIDs,
		MinPrice:    query.MinPrice,
		MaxPrice:    query.MaxPrice,
		SortBy:      sortBy,
		SortOrder:   ParseSortOrder(query.SortOrder),
		Limit:       limit,
		Offset:      offset,
	}

	return domainFilter, nil
}

// Helper functions

// Add a helper method to your SortOrder type to parse a string.
func ParseSortOrder(s string) domain.SortOrder {
	switch strings.ToUpper(s) {
	case "ASC":
		return domain.SortOrderAsc
	case "DESC":
		return domain.SortOrderDesc
	default:
		return domain.SortOrderUnspecified
	}
}

// func (s *productService) intersectCategories(requested []string, managed []*domain.Category) []string {
// 	managedMap := make(map[string]bool)
// 	for _, cat := range managed {
// 		managedMap[cat.ID] = true
// 	}

// 	var result []string
// 	for _, catID := range requested {
// 		if managedMap[catID] {
// 			result = append(result, catID)
// 		}
// 	}
// 	return result
// }

// func (s *productService) excludeCategories(categoryIDs []string, restrictedCategories []string) []string {
// 	if len(categoryIDs) == 0 {
// 		return categoryIDs
// 	}

// 	restrictedMap := make(map[string]bool)
// 	for _, restricted := range restrictedCategories {
// 		restrictedMap[restricted] = true
// 	}

// 	var result []string
// 	for _, catID := range categoryIDs {
// 		// Get category to check its type/name
// 		category, err := s.categoryRepo.GetByID(catID)
// 		if err != nil {
// 			continue // Skip invalid categories
// 		}

// 		// Check if category is restricted
// 		if !restrictedMap[category.CategoryType] && !restrictedMap[strings.ToLower(category.Name)] {
// 			result = append(result, catID)
// 		}
// 	}
// 	return result
// }

// ========================================================= Tag Relations =========================================================

// AddProductTags adds tags to a product
func (s *productService) AddProductTags(ctx context.Context, cmd AddProductTagsCommand) error {
	// Validate command
	if cmd.ProductID == "" {
		return errors.New("product ID is required")
	}
	if len(cmd.TagIDs) == 0 {
		return errors.New("at least one tag ID is required")
	}
	if cmd.AddedBy == "" {
		return errors.New("added_by is required")
	}

	// Validate product exists
	_, err := s.productRepo.GetByID(cmd.ProductID)
	if err != nil {
		return fmt.Errorf("product not found: %w", err)
	}

	// Validate all tags exist
	for _, tagID := range cmd.TagIDs {
		_, err := s.tagRepo.GetByID(tagID)
		if err != nil {
			return fmt.Errorf("tag %s not found: %w", tagID, err)
		}
	}
	// Business validation
	if err := s.validateTagAssignment(cmd.ProductID, cmd.TagIDs); err != nil {
		return fmt.Errorf("tag assignment validation failed: %w", err)
	}
	// Add tags
	if err := s.productRepo.AddTags(cmd.ProductID, cmd.TagIDs); err != nil {
		return fmt.Errorf("failed to add tags to product: %w", err)
	}
	return nil
}

// RemoveProductTags removes tags from a product
func (s *productService) RemoveProductTags(ctx context.Context, cmd RemoveProductTagsCommand) error {
	// Validate command
	if cmd.ProductID == "" {
		return errors.New("product ID is required")
	}
	if len(cmd.TagIDs) == 0 {
		return errors.New("at least one tag ID is required")
	}
	if cmd.RemovedBy == "" {
		return errors.New("removed_by is required")
	}

	// Validate product exists
	_, err := s.productRepo.GetByID(cmd.ProductID)
	if err != nil {
		return fmt.Errorf("product not found: %w", err)
	}

	// Remove tags
	if err := s.productRepo.RemoveTags(cmd.ProductID, cmd.TagIDs); err != nil {
		return fmt.Errorf("failed to remove tags from product: %w", err)
	}

	return nil
}

// GetProductTags retrieves tags for a product
func (s *productService) GetProductTags(ctx context.Context, query GetProductTagsQuery) ([]*TagSummary, error) {
	// Validate query
	if query.ProductID == "" {
		return nil, errors.New("product ID is required")
	}
	if query.RequestedBy == "" {
		return nil, errors.New("requested_by is required")
	}

	// Validate product exists
	_, err := s.productRepo.GetByID(query.ProductID)
	if err != nil {
		return nil, fmt.Errorf("product not found: %w", err)
	}

	// Get tags
	tags, err := s.productRepo.GetTags(query.ProductID)
	if err != nil {
		return nil, fmt.Errorf("failed to get product tags: %w", err)
	}

	// Convert to summaries
	tagSummaries := make([]*TagSummary, len(tags))
	for i, tag := range tags {
		tagSummaries[i] = &TagSummary{
			ID:   tag.ID,
			Name: tag.Name,
			Slug: tag.Slug,
		}
	}

	return tagSummaries, nil
}

// Business validation constants
const (
	MaxTagsPerProduct       = 20 // Maximum number of tags per product
	MaxDuplicateTagAttempts = 5  // Maximum duplicate tag attempts before blocking
)

// validateTagAssignment performs business validation for tag assignment
func (s *productService) validateTagAssignment(productID string, tagIDs []string) error {
	// Get current product tags
	currentTags, err := s.productRepo.GetTags(productID)
	if err != nil {
		return fmt.Errorf("failed to get current product tags: %w", err)
	}

	// Get tag details for new tags
	newTags := make([]*TagSummary, 0, len(tagIDs))
	for _, tagID := range tagIDs {
		tag, err := s.tagRepo.GetByID(tagID)
		if err != nil {
			return fmt.Errorf("failed to get tag %s: %w", tagID, err)
		}

		// Convert to Tag struct (assuming your repository returns compatible type)
		newTag := &TagSummary{
			ID:   tag.ID,
			Name: tag.Name,
			Slug: tag.Slug,
		}
		newTags = append(newTags, newTag)
	}

	// 2. Check for duplicates
	if err := s.validateNoDuplicateTags(currentTags, newTags); err != nil {
		return err
	}

	// 3. Check maximum tags limit
	if err := s.validateMaxTagsLimit(currentTags, newTags); err != nil {
		return err
	}

	return nil
}

// validateNoDuplicateTags checks for duplicate tag assignments
func (s *productService) validateNoDuplicateTags(currentTags []*domain.Tag, newTags []*TagSummary) error {
	currentTagMap := make(map[string]bool)
	for _, tag := range currentTags {
		currentTagMap[tag.ID] = true
	}

	var duplicates []string
	for _, newTag := range newTags {
		if currentTagMap[newTag.ID] {
			duplicates = append(duplicates, newTag.Name)
		}
	}

	if len(duplicates) > 0 {
		return fmt.Errorf("tags already assigned to product: %s", strings.Join(duplicates, ", "))
	}

	return nil
}

// validateMaxTagsLimit checks if adding new tags exceeds the maximum limit
func (s *productService) validateMaxTagsLimit(currentTags []*domain.Tag, newTags []*TagSummary) error {
	totalTags := len(currentTags) + len(newTags)
	if totalTags > MaxTagsPerProduct {
		return fmt.Errorf("cannot assign tags: would exceed maximum limit of %d tags per product (current: %d, adding: %d)",
			MaxTagsPerProduct, len(currentTags), len(newTags))
	}
	return nil
}
