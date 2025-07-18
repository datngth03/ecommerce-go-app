// cmd/product-client/main.go
package main

import (
	"context"
	"fmt"
	"log"
	"time"

	product_client "github.com/datngth03/ecommerce-go-app/pkg/client/product" // Import mã gRPC đã tạo
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure" // Dùng cho môi trường dev, không có TLS
)

func main() {
	// Địa chỉ của Product Service (cùng cổng mà Product Service đang lắng nghe)
	productSvcAddr := "localhost:50052"

	// Thiết lập kết nối gRPC đến Product Service
	conn, err := grpc.Dial(productSvcAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Không thể kết nối đến Product Service: %v", err)
	}
	defer conn.Close() // Đảm bảo đóng kết nối khi hàm main kết thúc

	// Tạo một gRPC client cho ProductService
	client := product_client.NewProductServiceClient(conn)

	// Context với timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var createdCategoryID string
	var createdProductID string

	// --- Test Case 1: Tạo một danh mục mới ---
	fmt.Println("\n--- Tạo một danh mục mới ---")
	createCategoryReq := &product_client.CreateCategoryRequest{
		Name:        "Electronics",
		Description: "Electronic gadgets and devices",
	}
	createCategoryResp, err := client.CreateCategory(ctx, createCategoryReq)
	if err != nil {
		log.Printf("Lỗi khi tạo danh mục: %v", err)
	} else {
		createdCategoryID = createCategoryResp.GetCategory().GetId()
		fmt.Printf("Tạo danh mục thành công! ID: %s, Tên: %s\n", createdCategoryID, createCategoryResp.GetCategory().GetName())
	}

	// --- Test Case 2: Lấy danh mục theo ID ---
	fmt.Println("\n--- Lấy danh mục theo ID ---")
	if createdCategoryID != "" {
		getCategoryReq := &product_client.GetCategoryByIdRequest{Id: createdCategoryID}
		getCategoryResp, err := client.GetCategoryById(ctx, getCategoryReq)
		if err != nil {
			log.Printf("Lỗi khi lấy danh mục theo ID: %v", err)
		} else {
			fmt.Printf("Lấy danh mục thành công! ID: %s, Tên: %s\n", getCategoryResp.GetCategory().GetId(), getCategoryResp.GetCategory().GetName())
		}
	} else {
		fmt.Println("Không thể lấy danh mục vì chưa tạo được.")
	}

	// --- Test Case 3: Liệt kê tất cả danh mục ---
	fmt.Println("\n--- Liệt kê tất cả danh mục ---")
	listCategoriesResp, err := client.ListCategories(ctx, &product_client.ListCategoriesRequest{})
	if err != nil {
		log.Printf("Lỗi khi liệt kê danh mục: %v", err)
	} else {
		fmt.Printf("Tổng số danh mục: %d\n", len(listCategoriesResp.GetCategories()))
		for _, cat := range listCategoriesResp.GetCategories() {
			fmt.Printf("  - ID: %s, Tên: %s\n", cat.GetId(), cat.GetName())
		}
	}

	// --- Test Case 4: Tạo một sản phẩm mới ---
	fmt.Println("\n--- Tạo một sản phẩm mới ---")
	if createdCategoryID != "" {
		createProductReq := &product_client.CreateProductRequest{
			Name:        "Smartphone X",
			Description: "Latest model smartphone with advanced features.",
			Price:       999.99,
			CategoryId:  createdCategoryID,
			ImageUrls:   []string{"http://example.com/phone-x-1.jpg", "http://example.com/phone-x-2.jpg"},
		}
		createProductResp, err := client.CreateProduct(ctx, createProductReq)
		if err != nil {
			log.Printf("Lỗi khi tạo sản phẩm: %v", err)
		} else {
			createdProductID = createProductResp.GetProduct().GetId()
			fmt.Printf("Tạo sản phẩm thành công! ID: %s, Tên: %s, Giá: %.2f\n",
				createdProductID, createProductResp.GetProduct().GetName(), createProductResp.GetProduct().GetPrice())
		}
	} else {
		fmt.Println("Không thể tạo sản phẩm vì chưa có danh mục.")
	}

	// --- Test Case 5: Lấy sản phẩm theo ID ---
	fmt.Println("\n--- Lấy sản phẩm theo ID ---")
	if createdProductID != "" {
		getProductReq := &product_client.GetProductByIdRequest{Id: createdProductID}
		getProductResp, err := client.GetProductById(ctx, getProductReq)
		if err != nil {
			log.Printf("Lỗi khi lấy sản phẩm theo ID: %v", err)
		} else {
			fmt.Printf("Lấy sản phẩm thành công! ID: %s, Tên: %s, Giá: %.2f, Tồn kho: %d\n",
				getProductResp.GetProduct().GetId(), getProductResp.GetProduct().GetName(),
				getProductResp.GetProduct().GetPrice(), getProductResp.GetProduct().GetStockQuantity())
		}
	} else {
		fmt.Println("Không thể lấy sản phẩm vì chưa tạo được.")
	}

	// --- Test Case 6: Cập nhật thông tin sản phẩm ---
	fmt.Println("\n--- Cập nhật thông tin sản phẩm ---")
	if createdProductID != "" {
		updateProductReq := &product_client.UpdateProductRequest{
			Id:          createdProductID,
			Name:        "Smartphone X Pro",
			Description: "Updated: Latest model smartphone with advanced features and better camera.",
			Price:       1099.99,
			CategoryId:  createdCategoryID, // Giữ nguyên hoặc thay đổi nếu có danh mục khác
			ImageUrls:   []string{"http://example.com/phone-x-pro-1.jpg"},
		}
		updateProductResp, err := client.UpdateProduct(ctx, updateProductReq)
		if err != nil {
			log.Printf("Lỗi khi cập nhật sản phẩm: %v", err)
		} else {
			fmt.Printf("Cập nhật sản phẩm thành công! ID: %s, Tên mới: %s, Giá mới: %.2f\n",
				updateProductResp.GetProduct().GetId(), updateProductResp.GetProduct().GetName(), updateProductResp.GetProduct().GetPrice())
		}
	} else {
		fmt.Println("Không thể cập nhật sản phẩm vì chưa tạo được.")
	}

	// --- Test Case 7: Liệt kê tất cả sản phẩm ---
	fmt.Println("\n--- Liệt kê tất cả sản phẩm ---")
	listProductsResp, err := client.ListProducts(ctx, &product_client.ListProductsRequest{
		Limit:  10,
		Offset: 0,
	})
	if err != nil {
		log.Printf("Lỗi khi liệt kê sản phẩm: %v", err)
	} else {
		fmt.Printf("Tổng số sản phẩm: %d\n", listProductsResp.GetTotalCount())
		for _, p := range listProductsResp.GetProducts() {
			fmt.Printf("  - ID: %s, Tên: %s, Giá: %.2f\n", p.GetId(), p.GetName(), p.GetPrice())
		}
	}

	// --- Test Case 8: Xóa sản phẩm ---
	fmt.Println("\n--- Xóa sản phẩm ---")
	if createdProductID != "" {
		deleteProductReq := &product_client.DeleteProductRequest{Id: createdProductID}
		deleteProductResp, err := client.DeleteProduct(ctx, deleteProductReq)
		if err != nil {
			log.Printf("Lỗi khi xóa sản phẩm: %v", err)
		} else {
			fmt.Printf("Xóa sản phẩm thành công! Trạng thái: %t, Thông báo: %s\n",
				deleteProductResp.GetSuccess(), deleteProductResp.GetMessage())
		}
	} else {
		fmt.Println("Không thể xóa sản phẩm vì chưa tạo được.")
	}

	// --- Test Case 9: Xóa danh mục (chỉ khi không còn sản phẩm liên quan) ---
	fmt.Println("\n--- Xóa danh mục ---")
	// Lưu ý: Nếu có sản phẩm liên quan đến danh mục này, việc xóa sẽ thất bại do ràng buộc FOREIGN KEY ON DELETE RESTRICT
	// Để xóa thành công, bạn cần đảm bảo không còn sản phẩm nào thuộc danh mục này.
	// if createdCategoryID != "" {
	// 	deleteCategoryReq := &product_client.DeleteCategoryRequest{Id: createdCategoryID}
	// 	// product_client.DeleteCategoryRequest không tồn tại trong product.proto
	// 	// Cần thêm DeleteCategory RPC vào product.proto và tạo lại mã
	// 	// Tạm thời bỏ qua test case này hoặc thêm vào sau khi cập nhật product.proto
	// 	fmt.Println("Xóa danh mục tạm thời bị bỏ qua do thiếu RPC trong product.proto.")
	// } else {
	// 	fmt.Println("Không thể xóa danh mục vì chưa tạo được.")
	// }

	fmt.Println("\nHoàn thành các test case cho Product Service.")
	time.Sleep(1 * time.Second) // Đợi một chút để log kịp in ra
}
