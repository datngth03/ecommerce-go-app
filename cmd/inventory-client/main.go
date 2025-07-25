// cmd/inventory-client/main.go
package main

import (
	"context"
	"fmt"
	"log"
	"time"

	// For generating dummy IDs
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	inventory_client "github.com/datngth03/ecommerce-go-app/pkg/client/inventory" // Generated Inventory gRPC client
	product_client "github.com/datngth03/ecommerce-go-app/pkg/client/product"     // Generated Product gRPC client
)

func main() {
	// Addresses of services
	inventorySvcAddr := "localhost:50059"
	productSvcAddr := "localhost:50052" // Needed to create product for inventory

	// --- Connect to Inventory Service ---
	inventoryConn, err := grpc.Dial(inventorySvcAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Failed to connect to Inventory Service: %v", err)
	}
	defer inventoryConn.Close()
	inventoryClient := inventory_client.NewInventoryServiceClient(inventoryConn)

	// --- Connect to Product Service (to create dummy product for inventory) ---
	productConn, err := grpc.Dial(productSvcAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Failed to connect to Product Service: %v", err)
	}
	defer productConn.Close()
	productClient := product_client.NewProductServiceClient(productConn)

	// Context with timeout for RPC calls
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second) // Increased timeout for multiple calls
	defer cancel()

	// --- Prepare dummy data: Category, Product ---
	var categoryID string
	var productID string
	var initialStock int32 = 100

	fmt.Println("\n--- Preparing test data (Category & Product for Inventory) ---")
	// Try to create a category
	createCategoryReq := &product_client.CreateCategoryRequest{Name: "InventoryTestCategory", Description: "Category for inventory testing"}
	createCategoryResp, err := productClient.CreateCategory(ctx, createCategoryReq)
	if err != nil {
		log.Printf("Error creating category (might already exist): %v", err)
		listCategoriesResp, listErr := productClient.ListCategories(ctx, &product_client.ListCategoriesRequest{})
		if listErr == nil {
			for _, cat := range listCategoriesResp.GetCategories() {
				if cat.GetName() == "InventoryTestCategory" {
					categoryID = cat.GetId()
					log.Printf("Found existing category: %s", categoryID)
					break
				}
			}
		}
	} else {
		categoryID = createCategoryResp.GetCategory().GetId()
		log.Printf("Created category: %s", categoryID)
	}
	if categoryID == "" {
		log.Fatalf("Failed to get or create a category. Cannot proceed with product creation.")
	}

	createProductReq := &product_client.CreateProductRequest{
		Name:        "InventoryTestProduct",
		Description: "Product for inventory testing.",
		Price:       150.00,
		CategoryId:  categoryID,
		ImageUrls:   []string{"http://example.com/inventory-test-product.jpg"},
	}
	createProductResp, err := productClient.CreateProduct(ctx, createProductReq)
	if err != nil {
		log.Printf("Error creating product (might already exist): %v", err)
		listProductsResp, listErr := productClient.ListProducts(ctx, &product_client.ListProductsRequest{CategoryId: categoryID})
		if listErr == nil {
			for _, prod := range listProductsResp.GetProducts() {
				if prod.GetName() == "InventoryTestProduct" {
					productID = prod.GetId()
					log.Printf("Found existing product: %s", productID)
					break
				}
			}
		}
	} else {
		productID = createProductResp.GetProduct().GetId()
		log.Printf("Created product: %s", productID)
	}
	if productID == "" {
		log.Fatalf("Failed to get or create a product. Cannot proceed with inventory operations.")
	}

	// --- Test Case 1: Set initial stock for the product ---
	fmt.Println("\n--- Setting initial stock ---")
	setStockReq := &inventory_client.SetStockRequest{
		ProductId: productID,
		Quantity:  initialStock,
	}
	setStockResp, err := inventoryClient.SetStock(ctx, setStockReq)
	if err != nil {
		log.Printf("Error setting initial stock: %v", err)
	} else {
		fmt.Printf("Initial stock set! Product ID: %s, Quantity: %d, Reserved: %d\n",
			setStockResp.GetItem().GetProductId(), setStockResp.GetItem().GetQuantity(), setStockResp.GetItem().GetReservedQuantity())
	}

	// --- Test Case 2: Get stock quantity ---
	fmt.Println("\n--- Getting stock quantity ---")
	getStockReq := &inventory_client.GetStockQuantityRequest{ProductId: productID}
	getStockResp, err := inventoryClient.GetStockQuantity(ctx, getStockReq)
	if err != nil {
		log.Printf("Error getting stock quantity: %v", err)
	} else {
		fmt.Printf("Stock: Product ID: %s, Quantity: %d, Reserved: %d\n",
			getStockResp.GetItem().GetProductId(), getStockResp.GetItem().GetQuantity(), getStockResp.GetItem().GetReservedQuantity())
	}

	// --- Test Case 3: Reserve stock ---
	fmt.Println("\n--- Reserving 5 units of stock ---")
	reserveStockReq := &inventory_client.ReserveStockRequest{
		ProductId: productID,
		Quantity:  5,
	}
	reserveStockResp, err := inventoryClient.ReserveStock(ctx, reserveStockReq)
	if err != nil {
		log.Printf("Error reserving stock: %v", err)
	} else {
		fmt.Printf("Stock after reserve: Product ID: %s, Quantity: %d, Reserved: %d\n",
			reserveStockResp.GetItem().GetProductId(), reserveStockResp.GetItem().GetQuantity(), reserveStockResp.GetItem().GetReservedQuantity())
	}

	// --- Test Case 4: Decrease stock (fulfill order) ---
	fmt.Println("\n--- Decreasing 3 units of stock (from reserved) ---")
	decreaseStockReq := &inventory_client.DecreaseStockRequest{
		ProductId: productID,
		Quantity:  3,
	}
	decreaseStockResp, err := inventoryClient.DecreaseStock(ctx, decreaseStockReq)
	if err != nil {
		log.Printf("Error decreasing stock: %v", err)
	} else {
		fmt.Printf("Stock after decrease: Product ID: %s, Quantity: %d, Reserved: %d\n",
			decreaseStockResp.GetItem().GetProductId(), decreaseStockResp.GetItem().GetQuantity(), decreaseStockResp.GetItem().GetReservedQuantity())
	}

	// --- Test Case 5: Release reserved stock (e.g., order cancelled) ---
	fmt.Println("\n--- Releasing 2 units of reserved stock ---")
	releaseStockReq := &inventory_client.ReleaseStockRequest{
		ProductId: productID,
		Quantity:  2,
	}
	releaseStockResp, err := inventoryClient.ReleaseStock(ctx, releaseStockReq)
	if err != nil {
		log.Printf("Error releasing stock: %v", err)
	} else {
		fmt.Printf("Stock after release: Product ID: %s, Quantity: %d, Reserved: %d\n",
			releaseStockResp.GetItem().GetProductId(), releaseStockResp.GetItem().GetQuantity(), releaseStockResp.GetItem().GetReservedQuantity())
	}

	// --- Test Case 6: Increase stock (e.g., new delivery) ---
	fmt.Println("\n--- Increasing 10 units of stock ---")
	increaseStockReq := &inventory_client.IncreaseStockRequest{
		ProductId: productID,
		Quantity:  10,
	}
	increaseStockResp, err := inventoryClient.IncreaseStock(ctx, increaseStockReq)
	if err != nil {
		log.Printf("Error increasing stock: %v", err)
	} else {
		fmt.Printf("Stock after increase: Product ID: %s, Quantity: %d, Reserved: %d\n",
			increaseStockResp.GetItem().GetProductId(), increaseStockResp.GetItem().GetQuantity(), increaseStockResp.GetItem().GetReservedQuantity())
	}

	fmt.Println("\nFinished Inventory Service test cases.")
	time.Sleep(1 * time.Second) // Give some time for logs
}
