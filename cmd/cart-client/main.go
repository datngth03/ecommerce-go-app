// cmd/cart-client/main.go
package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid" // For generating dummy IDs
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	cart_client "github.com/datngth03/ecommerce-go-app/pkg/client/cart"       // Generated Cart gRPC client
	product_client "github.com/datngth03/ecommerce-go-app/pkg/client/product" // Generated Product gRPC client
)

func main() {
	// Addresses of services
	cartSvcAddr := "localhost:50054"
	productSvcAddr := "localhost:50052"

	// --- Connect to Cart Service ---
	cartConn, err := grpc.Dial(cartSvcAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Failed to connect to Cart Service: %v", err)
	}
	defer cartConn.Close()
	cartClient := cart_client.NewCartServiceClient(cartConn)

	// --- Connect to Product Service (needed to get product details for adding to cart) ---
	productConn, err := grpc.Dial(productSvcAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Failed to connect to Product Service: %v", err)
	}
	defer productConn.Close()
	productClient := product_client.NewProductServiceClient(productConn)

	// Context with timeout for RPC calls
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// --- Prepare dummy data ---
	// You should ideally get a real user ID from User Service
	dummyUserID := uuid.New().String()
	log.Printf("Using dummy User ID for cart: %s", dummyUserID)

	// Create a dummy category and product first if they don't exist
	var categoryID string
	var productID string
	var productName string
	var productPrice float64 = 10.00

	fmt.Println("\n--- Preparing test data (Category & Product for Cart) ---")
	// Try to create a category
	createCategoryReq := &product_client.CreateCategoryRequest{Name: "CartTestCategory", Description: "Category for cart testing"}
	createCategoryResp, err := productClient.CreateCategory(ctx, createCategoryReq)
	if err != nil {
		log.Printf("Error creating category (might already exist): %v", err)
		// Try to find it if it exists
		listCategoriesResp, listErr := productClient.ListCategories(ctx, &product_client.ListCategoriesRequest{})
		if listErr == nil {
			for _, cat := range listCategoriesResp.GetCategories() {
				if cat.GetName() == "CartTestCategory" {
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
		log.Fatalf("Failed to get or create a category. Cannot proceed with product creation for cart.")
	}

	// Try to create a product
	createProductReq := &product_client.CreateProductRequest{
		Name:        "CartTestProduct",
		Description: "Product for cart testing.",
		Price:       productPrice,
		CategoryId:  categoryID,
		ImageUrls:   []string{"http://example.com/cart-test-product.jpg"},
	}
	createProductResp, err := productClient.CreateProduct(ctx, createProductReq)
	if err != nil {
		log.Printf("Error creating product (might already exist): %v", err)
		// Try to find it if it exists
		listProductsResp, listErr := productClient.ListProducts(ctx, &product_client.ListProductsRequest{CategoryId: categoryID})
		if listErr == nil {
			for _, prod := range listProductsResp.GetProducts() {
				if prod.GetName() == "CartTestProduct" {
					productID = prod.GetId()
					productName = prod.GetName()
					productPrice = prod.GetPrice()
					log.Printf("Found existing product: %s", productID)
					break
				}
			}
		}
	} else {
		productID = createProductResp.GetProduct().GetId()
		productName = createProductResp.GetProduct().GetName()
		productPrice = createProductResp.GetProduct().GetPrice()
		log.Printf("Created product: %s", productID)
	}

	if productID == "" {
		log.Fatalf("Failed to get or create a product. Cannot proceed with cart operations.")
	}

	// --- Test Case 1: Get an empty cart (or new cart) ---
	fmt.Println("\n--- Getting initial cart ---")
	getCartResp, err := cartClient.GetCart(ctx, &cart_client.GetCartRequest{UserId: dummyUserID})
	if err != nil {
		log.Printf("Error getting initial cart: %v", err)
	} else {
		fmt.Printf("Initial Cart for User %s: Total Items: %d, Total Amount: %.2f\n",
			getCartResp.GetCart().GetUserId(), len(getCartResp.GetCart().GetItems()), getCartResp.GetCart().GetTotalAmount())
	}

	// --- Test Case 2: Add item to cart ---
	fmt.Println("\n--- Adding item to cart ---")
	addItemReq := &cart_client.AddItemToCartRequest{
		UserId:      dummyUserID,
		ProductId:   productID,
		ProductName: productName,
		Price:       productPrice,
		Quantity:    2,
	}
	addItemResp, err := cartClient.AddItemToCart(ctx, addItemReq)
	if err != nil {
		log.Printf("Error adding item to cart: %v", err)
	} else {
		fmt.Printf("Item added to cart! User: %s, Total Items in Cart: %d, Total Amount: %.2f\n",
			addItemResp.GetCart().GetUserId(), len(addItemResp.GetCart().GetItems()), addItemResp.GetCart().GetTotalAmount())
		for _, item := range addItemResp.GetCart().GetItems() {
			fmt.Printf("  - Product: %s (Qty: %d, Price: %.2f)\n", item.GetProductName(), item.GetQuantity(), item.GetPrice())
		}
	}

	// --- Test Case 3: Add the same item again (should update quantity) ---
	fmt.Println("\n--- Adding same item again (should update quantity) ---")
	addItemReq2 := &cart_client.AddItemToCartRequest{
		UserId:      dummyUserID,
		ProductId:   productID,
		ProductName: productName, // Name/Price might be ignored by service, but good to pass
		Price:       productPrice,
		Quantity:    1,
	}
	addItemResp2, err := cartClient.AddItemToCart(ctx, addItemReq2)
	if err != nil {
		log.Printf("Error adding same item to cart: %v", err)
	} else {
		fmt.Printf("Item quantity updated! User: %s, Total Items in Cart: %d, Total Amount: %.2f\n",
			addItemResp2.GetCart().GetUserId(), len(addItemResp2.GetCart().GetItems()), addItemResp2.GetCart().GetTotalAmount())
		for _, item := range addItemResp2.GetCart().GetItems() {
			fmt.Printf("  - Product: %s (Qty: %d, Price: %.2f)\n", item.GetProductName(), item.GetQuantity(), item.GetPrice())
		}
	}

	// --- Test Case 4: Update item quantity ---
	fmt.Println("\n--- Updating item quantity to 5 ---")
	updateItemReq := &cart_client.UpdateCartItemQuantityRequest{
		UserId:    dummyUserID,
		ProductId: productID,
		Quantity:  5,
	}
	updateItemResp, err := cartClient.UpdateCartItemQuantity(ctx, updateItemReq)
	if err != nil {
		log.Printf("Error updating item quantity: %v", err)
	} else {
		fmt.Printf("Item quantity updated! User: %s, Total Items in Cart: %d, Total Amount: %.2f\n",
			updateItemResp.GetCart().GetUserId(), len(updateItemResp.GetCart().GetItems()), updateItemResp.GetCart().GetTotalAmount())
		for _, item := range updateItemResp.GetCart().GetItems() {
			fmt.Printf("  - Product: %s (Qty: %d, Price: %.2f)\n", item.GetProductName(), item.GetQuantity(), item.GetPrice())
		}
	}

	// --- Test Case 5: Remove item from cart ---
	fmt.Println("\n--- Removing item from cart ---")
	removeItemReq := &cart_client.RemoveItemFromCartRequest{
		UserId:    dummyUserID,
		ProductId: productID,
	}
	removeItemResp, err := cartClient.RemoveItemFromCart(ctx, removeItemReq)
	if err != nil {
		log.Printf("Error removing item from cart: %v", err)
	} else {
		fmt.Printf("Item removed from cart! User: %s, Total Items in Cart: %d, Total Amount: %.2f\n",
			removeItemResp.GetCart().GetUserId(), len(removeItemResp.GetCart().GetItems()), removeItemResp.GetCart().GetTotalAmount())
	}

	// --- Test Case 6: Clear cart ---
	fmt.Println("\n--- Clearing cart ---")
	clearCartReq := &cart_client.ClearCartRequest{UserId: dummyUserID}
	clearCartResp, err := cartClient.ClearCart(ctx, clearCartReq)
	if err != nil {
		log.Printf("Error clearing cart: %v", err)
	} else {
		fmt.Printf("Cart cleared! User: %s, Total Items: %d, Total Amount: %.2f\n",
			clearCartResp.GetCart().GetUserId(), len(clearCartResp.GetCart().GetItems()), clearCartResp.GetCart().GetTotalAmount())
	}

	fmt.Println("\nFinished Cart Service test cases.")
	time.Sleep(1 * time.Second) // Give some time for logs
}
