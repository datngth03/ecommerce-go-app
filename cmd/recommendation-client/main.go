// cmd/recommendation-client/main.go
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/joho/godotenv"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	product_client "github.com/datngth03/ecommerce-go-app/pkg/client/product"
	recommendation_client "github.com/datngth03/ecommerce-go-app/pkg/client/recommendation"
)

// init loads environment variables from .env file.
func init() {
	if err := godotenv.Load(); err != nil {
		log.Printf("Warning: .env file not loaded, falling back to environment variables: %v", err)
	}
}

func main() {
	// Get service addresses from environment variables
	recommendationSvcAddr := os.Getenv("RECOMMENDATION_GRPC_ADDR")
	if recommendationSvcAddr == "" {
		recommendationSvcAddr = "localhost:50062"
	}

	productSvcAddr := os.Getenv("PRODUCT_GRPC_ADDR")
	if productSvcAddr == "" {
		productSvcAddr = "localhost:50052"
	}

	// Set up a connection to the Recommendation Service.
	conn, err := grpc.Dial(recommendationSvcAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("did not connect to Recommendation Service: %v", err)
	}
	defer conn.Close()
	recommendationClient := recommendation_client.NewRecommendationServiceClient(conn)

	// Set up a connection to the Product Service (needed for creating sample products).
	productConn, err := grpc.Dial(productSvcAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("did not connect to Product Service: %v", err)
	}
	defer productConn.Close()
	productClient := product_client.NewProductServiceClient(productConn)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	log.Println("--- Bắt đầu kiểm tra Recommendation Service ---")

	dummyUserID1 := uuid.New().String()
	dummyUserID2 := uuid.New().String()
	log.Printf("Sử dụng User ID 1: %s", dummyUserID1)
	log.Printf("Sử dụng User ID 2: %s", dummyUserID2)

	// Test Case 1: Create a dummy category and products for interactions
	log.Println("\n--- Test Case 1: Tạo danh mục và sản phẩm mẫu ---")
	categoryID := uuid.New().String()
	createCategoryResp, err := productClient.CreateCategory(ctx, &product_client.CreateCategoryRequest{
		Name:        fmt.Sprintf("ClientCategory-%s", categoryID[:8]),
		Description: "Category created by recommendation client",
	})
	if err != nil {
		log.Printf("Failed to create category: %v", err)
		// Try to list categories to find an existing one if creation failed (e.g. name conflict)
		listCatResp, listCatErr := productClient.ListCategories(ctx, &product_client.ListCategoriesRequest{})
		if listCatErr == nil && len(listCatResp.GetCategories()) > 0 {
			categoryID = listCatResp.GetCategories()[0].GetId()
			log.Printf("Using existing category ID: %s", categoryID)
		} else {
			log.Fatalf("Cannot get or create category: %v", listCatErr)
		}
	} else {
		categoryID = createCategoryResp.GetCategory().GetId()
		log.Printf("Created category ID: %s", categoryID)
	}

	productIDs := make([]string, 5)
	for i := 0; i < 5; i++ {
		productID := uuid.New().String()
		_, err := productClient.CreateProduct(ctx, &product_client.CreateProductRequest{
			Name:        fmt.Sprintf("Product%d-%s", i+1, productID[:8]),
			Description: fmt.Sprintf("Description for Product %d", i+1),
			Price:       float64(100 + i*10),
			CategoryId:  categoryID,
			ImageUrls:   []string{fmt.Sprintf("http://example.com/product%d.jpg", i+1)},
		})
		if err != nil {
			log.Printf("Failed to create product %d: %v", i+1, err)
		} else {
			productIDs[i] = productID
			log.Printf("Created product ID: %s", productIDs[i])
		}
	}

	// Test Case 2: Record various user interactions
	log.Println("\n--- Test Case 2: Ghi lại các tương tác người dùng ---")
	interactions := []struct {
		UserID    string
		ProductID string
		EventType string
	}{
		{dummyUserID1, productIDs[0], "view"},
		{dummyUserID1, productIDs[1], "view"},
		{dummyUserID1, productIDs[0], "add_to_cart"},
		{dummyUserID1, productIDs[2], "view"},
		{dummyUserID1, productIDs[1], "purchase"}, // Giả định mua hàng
		{dummyUserID2, productIDs[0], "view"},
		{dummyUserID2, productIDs[3], "view"},
		{dummyUserID2, productIDs[0], "add_to_cart"},
		{dummyUserID2, productIDs[4], "purchase"}, // Giả định mua hàng
	}

	for _, interaction := range interactions {
		_, err := recommendationClient.RecordInteraction(ctx, &recommendation_client.RecordInteractionRequest{
			UserId:    interaction.UserID,
			ProductId: interaction.ProductID,
			EventType: interaction.EventType,
			Timestamp: time.Now().Unix(),
		})
		if err != nil {
			log.Printf("Failed to record interaction (%s-%s-%s): %v", interaction.UserID, interaction.ProductID, interaction.EventType, err)
		} else {
			log.Printf("Recorded interaction: User %s, Product %s, Event %s", interaction.UserID, interaction.ProductID, interaction.EventType)
		}
	}

	// Test Case 3: Get recommendations for dummyUserID1
	log.Println("\n--- Test Case 3: Lấy gợi ý cho User ID 1 ---")
	getRecReq := &recommendation_client.GetRecommendationsRequest{
		UserId: dummyUserID1,
		Limit:  2,
		Offset: 0,
	}
	getRecResp, err := recommendationClient.GetRecommendations(ctx, getRecReq)
	if err != nil {
		log.Printf("Failed to get recommendations for User ID 1: %v", err)
	} else {
		log.Printf("Gợi ý cho User ID 1 (%d sản phẩm):", len(getRecResp.GetProducts()))
		for _, p := range getRecResp.GetProducts() {
			log.Printf("  - ID: %s, Tên: %s, Hình ảnh: %s, Điểm: %.2f", p.GetProductId(), p.GetName(), p.GetImageUrl(), p.GetScore())
		}
	}

	// Test Case 4: Get popular products
	log.Println("\n--- Test Case 4: Lấy sản phẩm phổ biến ---")
	getPopReq := &recommendation_client.GetPopularProductsRequest{
		Limit: 3,
	}
	getPopResp, err := recommendationClient.GetPopularProducts(ctx, getPopReq)
	if err != nil {
		log.Printf("Failed to get popular products: %v", err)
	} else {
		log.Printf("Sản phẩm phổ biến (%d sản phẩm):", len(getPopResp.GetProducts()))
		for _, p := range getPopResp.GetProducts() {
			log.Printf("  - ID: %s, Tên: %s, Hình ảnh: %s, Điểm: %.2f", p.GetProductId(), p.GetName(), p.GetImageUrl(), p.GetScore())
		}
	}

	log.Println("\n--- Hoàn tất kiểm tra Recommendation Service ---")
}
