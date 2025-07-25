// cmd/review-client/main.go
package main

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/joho/godotenv"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	product_client "github.com/datngth03/ecommerce-go-app/pkg/client/product" // Generated Product gRPC client
	review_client "github.com/datngth03/ecommerce-go-app/pkg/client/review"   // Generated Review gRPC client
	"github.com/google/uuid"                                                  // For generating UUIDs
)

func main() {
	// Load .env file
	if err := godotenv.Load(); err != nil {
		log.Println("Không tìm thấy file .env, tiếp tục mà không load biến môi trường.")
	}

	// Get gRPC service addresses from environment variables
	reviewSvcAddr := os.Getenv("REVIEW_GRPC_ADDR")
	if reviewSvcAddr == "" {
		reviewSvcAddr = "localhost:50060"
		log.Printf("REVIEW_GRPC_ADDR không được đặt, sử dụng mặc định: %s", reviewSvcAddr)
	}

	productSvcAddr := os.Getenv("PRODUCT_GRPC_ADDR")
	if productSvcAddr == "" {
		productSvcAddr = "localhost:50052"
		log.Printf("PRODUCT_GRPC_ADDR không được đặt, sử dụng mặc định: %s", productSvcAddr)
	}

	// Set up a connection to the Product Service
	productConn, err := grpc.Dial(productSvcAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Không thể kết nối đến Product Service: %v", err)
	}
	defer productConn.Close()
	productClient := product_client.NewProductServiceClient(productConn)

	// Set up a connection to the Review Service
	reviewConn, err := grpc.Dial(reviewSvcAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Không thể kết nối đến Review Service: %v", err)
	}
	defer reviewConn.Close()
	reviewClient := review_client.NewReviewServiceClient(reviewConn)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	log.Println("-------------------------------------------")
	log.Println("Bắt đầu kiểm tra Review Service...")
	log.Println("-------------------------------------------")

	// Helper to create a dummy product for testing reviews
	// Tạo một sản phẩm giả để kiểm tra đánh giá
	log.Println("Test Case 0: Tạo sản phẩm giả để đánh giá...")
	dummyCategoryID := uuid.New().String() // Tạo Category ID mới mỗi lần
	createCategoryResp, err := productClient.CreateCategory(ctx, &product_client.CreateCategoryRequest{
		Name:        "Test Category " + uuid.New().String(),
		Description: "For review testing",
	})
	if err != nil {
		log.Printf("Lỗi khi tạo danh mục: %v. Có thể danh mục đã tồn tại hoặc Product Service không chạy. Thử bỏ qua nếu bạn biết có danh mục sẵn.", err)
		// Try to list categories to find one if creation failed
		listCategoriesResp, listErr := productClient.ListCategories(ctx, &product_client.ListCategoriesRequest{})
		if listErr == nil && len(listCategoriesResp.GetCategories()) > 0 {
			dummyCategoryID = listCategoriesResp.GetCategories()[0].GetId()
			log.Printf("Đã sử dụng danh mục hiện có: %s", dummyCategoryID)
		} else {
			log.Fatalf("Không thể tạo hoặc tìm thấy danh mục để tiếp tục test: %v", listErr)
		}
	} else {
		dummyCategoryID = createCategoryResp.GetCategory().GetId()
		log.Printf("Đã tạo danh mục mới: %s", dummyCategoryID)
	}

	dummyProductID := uuid.New().String() // Tạo Product ID mới mỗi lần
	createProductResp, err := productClient.CreateProduct(ctx, &product_client.CreateProductRequest{
		Name:        "Test Product " + uuid.New().String(),
		Description: "For review service testing",
		Price:       99.99,
		CategoryId:  dummyCategoryID,
		ImageUrls:   []string{"http://example.com/test_product.jpg"},
	})
	if err != nil {
		log.Printf("Lỗi khi tạo sản phẩm: %v. Có thể sản phẩm đã tồn tại hoặc Product Service không chạy.", err)
		// If product creation fails, try to list products and find one
		listProductsResp, listErr := productClient.ListProducts(ctx, &product_client.ListProductsRequest{
			Limit:  1,
			Offset: 0,
		})
		if listErr == nil && len(listProductsResp.GetProducts()) > 0 {
			dummyProductID = listProductsResp.GetProducts()[0].GetId()
			log.Printf("Đã sử dụng sản phẩm hiện có: %s", dummyProductID)
		} else {
			log.Fatalf("Không thể tạo hoặc tìm thấy sản phẩm để tiếp tục test: %v", listErr)
		}
	} else {
		dummyProductID = createProductResp.GetProduct().GetId()
		log.Printf("Đã tạo sản phẩm mới: %s", dummyProductID)
	}

	dummyUserID := uuid.New().String() // Sử dụng một User ID giả
	log.Printf("Sản phẩm ID giả: %s, User ID giả: %s", dummyProductID, dummyUserID)

	// Test Case 1: Submit a new review
	log.Println("\nTest Case 1: Gửi một đánh giá mới...")
	submitReviewReq := &review_client.SubmitReviewRequest{
		ProductId: dummyProductID,
		UserId:    dummyUserID,
		Rating:    5,
		Comment:   "Sản phẩm tuyệt vời! Rất hài lòng.",
	}
	submitReviewResp, err := reviewClient.SubmitReview(ctx, submitReviewReq)
	if err != nil {
		log.Fatalf("Lỗi khi gửi đánh giá: %v", err)
	}
	log.Printf("Đánh giá đã gửi thành công: %+v", submitReviewResp.GetReview())
	createdReviewID := submitReviewResp.GetReview().GetId()

	// Test Case 2: Try to submit the same review again (should fail due to FindExistingReview logic)
	log.Println("\nTest Case 2: Thử gửi lại đánh giá tương tự (mong đợi lỗi Review Already Exists)...")
	_, err = reviewClient.SubmitReview(ctx, submitReviewReq)
	if err != nil {
		log.Printf("Đã bắt được lỗi như mong đợi: %v", err)
	} else {
		log.Println("Lỗi: Đánh giá trùng lặp đã được gửi thành công. Cần kiểm tra logic FindExistingReview.")
	}

	// Test Case 3: Get review by ID
	log.Println("\nTest Case 3: Lấy đánh giá bằng ID...")
	getReviewResp, err := reviewClient.GetReviewById(ctx, &review_client.GetReviewByIdRequest{Id: createdReviewID})
	if err != nil {
		log.Fatalf("Lỗi khi lấy đánh giá bằng ID: %v", err)
	}
	log.Printf("Đánh giá đã lấy: %+v", getReviewResp.GetReview())

	// Test Case 4: Update an existing review
	log.Println("\nTest Case 4: Cập nhật đánh giá hiện có...")
	updateReviewReq := &review_client.UpdateReviewRequest{
		Id:      createdReviewID,
		Rating:  4,
		Comment: "Sản phẩm tốt, nhưng có thể cải thiện thêm một chút.",
	}
	updateReviewResp, err := reviewClient.UpdateReview(ctx, updateReviewReq)
	if err != nil {
		log.Fatalf("Lỗi khi cập nhật đánh giá: %v", err)
	}
	log.Printf("Đánh giá đã cập nhật: %+v", updateReviewResp.GetReview())

	// Test Case 5: List reviews by product
	log.Println("\nTest Case 5: Liệt kê đánh giá theo sản phẩm...")
	listByProductResp, err := reviewClient.ListReviewsByProduct(ctx, &review_client.ListReviewsByProductRequest{
		ProductId: dummyProductID,
		Limit:     10,
		Offset:    0,
	})
	if err != nil {
		log.Fatalf("Lỗi khi liệt kê đánh giá theo sản phẩm: %v", err)
	}
	log.Printf("Tìm thấy %d đánh giá cho sản phẩm %s:", listByProductResp.GetTotalCount(), dummyProductID)
	for _, review := range listByProductResp.GetReviews() {
		log.Printf("- ID: %s, Rating: %d, Comment: %s", review.GetId(), review.GetRating(), review.GetComment())
	}

	// Test Case 6: List all reviews (with filters)
	log.Println("\nTest Case 6: Liệt kê tất cả đánh giá (có bộ lọc)...")
	listAllReviewsResp, err := reviewClient.ListAllReviews(ctx, &review_client.ListAllReviewsRequest{
		ProductId: dummyProductID, // Lọc theo Product ID
		MinRating: 4,              // Lọc đánh giá từ 4 sao trở lên
		Limit:     10,
		Offset:    0,
	})
	if err != nil {
		log.Fatalf("Lỗi khi liệt kê tất cả đánh giá: %v", err)
	}
	log.Printf("Tìm thấy %d đánh giá với bộ lọc:", listAllReviewsResp.GetTotalCount())
	for _, review := range listAllReviewsResp.GetReviews() {
		log.Printf("- ID: %s, Rating: %d, Comment: %s", review.GetId(), review.GetRating(), review.GetComment())
	}

	// Test Case 7: Delete a review
	log.Println("\nTest Case 7: Xóa một đánh giá...")
	deleteReviewResp, err := reviewClient.DeleteReview(ctx, &review_client.DeleteReviewRequest{Id: createdReviewID})
	if err != nil {
		log.Fatalf("Lỗi khi xóa đánh giá: %v", err)
	}
	if deleteReviewResp.GetSuccess() {
		log.Printf("Đánh giá %s đã xóa thành công: %s", createdReviewID, deleteReviewResp.GetMessage())
	} else {
		log.Printf("Xóa đánh giá %s không thành công: %s", createdReviewID, deleteReviewResp.GetMessage())
	}

	// Test Case 8: Try to get the deleted review (should fail)
	log.Println("\nTest Case 8: Thử lấy đánh giá đã xóa (mong đợi lỗi Review Not Found)...")
	_, err = reviewClient.GetReviewById(ctx, &review_client.GetReviewByIdRequest{Id: createdReviewID})
	if err != nil {
		log.Printf("Đã bắt được lỗi như mong đợi: %v", err)
	} else {
		log.Println("Lỗi: Đánh giá đã xóa vẫn có thể lấy được.")
	}

	log.Println("\n-------------------------------------------")
	log.Println("Hoàn tất kiểm tra Review Service.")
	log.Println("-------------------------------------------")
}
