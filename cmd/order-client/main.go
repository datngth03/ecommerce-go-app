// cmd/order-client/main.go
package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid" // For generating dummy IDs
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	order_client "github.com/datngth03/ecommerce-go-app/pkg/client/order"     // Generated Order gRPC client
	product_client "github.com/datngth03/ecommerce-go-app/pkg/client/product" // Generated Product gRPC client
)

func main() {
	// Addresses of services
	orderSvcAddr := "localhost:50053"
	productSvcAddr := "localhost:50052"

	// --- Connect to Order Service ---
	orderConn, err := grpc.Dial(orderSvcAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Failed to connect to Order Service: %v", err)
	}
	defer orderConn.Close()
	orderClient := order_client.NewOrderServiceClient(orderConn)

	// --- Connect to Product Service (needed to get product details for creating orders) ---
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
	log.Printf("Using dummy User ID: %s", dummyUserID)

	// Create a dummy category and product first if they don't exist
	var categoryID string
	var productID string
	var productName string
	var productPrice float64 = 10.00

	fmt.Println("\n--- Preparing test data (Category & Product) ---")
	// Try to create a category
	createCategoryReq := &product_client.CreateCategoryRequest{Name: "Books", Description: "Fiction and Non-Fiction"}
	createCategoryResp, err := productClient.CreateCategory(ctx, createCategoryReq)
	if err != nil {
		log.Printf("Error creating category (might already exist): %v", err)
		// Try to find it if it exists
		listCategoriesResp, listErr := productClient.ListCategories(ctx, &product_client.ListCategoriesRequest{})
		if listErr == nil {
			for _, cat := range listCategoriesResp.GetCategories() {
				if cat.GetName() == "Books" {
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

	// Try to create a product
	createProductReq := &product_client.CreateProductRequest{
		Name:        "Go Programming Basics",
		Description: "A beginner's guide to Go programming.",
		Price:       productPrice,
		CategoryId:  categoryID,
		ImageUrls:   []string{"http://example.com/go-book.jpg"},
	}
	createProductResp, err := productClient.CreateProduct(ctx, createProductReq)
	if err != nil {
		log.Printf("Error creating product (might already exist): %v", err)
		// Try to find it if it exists
		listProductsResp, listErr := productClient.ListProducts(ctx, &product_client.ListProductsRequest{CategoryId: categoryID})
		if listErr == nil {
			for _, prod := range listProductsResp.GetProducts() {
				if prod.GetName() == "Go Programming Basics" {
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
		log.Fatalf("Failed to get or create a product. Cannot proceed with order creation.")
	}

	var createdOrderID string

	// --- Test Case 1: Create a new order ---
	fmt.Println("\n--- Creating a new order ---")
	createOrderReq := &order_client.CreateOrderRequest{
		UserId: dummyUserID,
		Items: []*order_client.OrderItem{
			{
				ProductId:   productID,
				ProductName: productName,
				Price:       productPrice,
				Quantity:    1,
			},
			// {
			// 	ProductId:   uuid.New().String(), // Another dummy product
			// 	ProductName: "Dummy Product 2",
			// 	Price:       5.00,
			// 	Quantity:    2,
			// },
		},
		ShippingAddress: "123 Main St, Anytown, USA",
	}
	createOrderResp, err := orderClient.CreateOrder(ctx, createOrderReq)
	if err != nil {
		log.Printf("Error creating order: %v", err)
	} else {
		createdOrderID = createOrderResp.GetOrder().GetId()
		fmt.Printf("Order created successfully! ID: %s, User ID: %s, Total: %.2f, Status: %s\n",
			createdOrderID, createOrderResp.GetOrder().GetUserId(),
			createOrderResp.GetOrder().GetTotalAmount(), createOrderResp.GetOrder().GetStatus())
	}

	// --- Test Case 2: Get order by ID ---
	fmt.Println("\n--- Getting order by ID ---")
	if createdOrderID != "" {
		getOrderReq := &order_client.GetOrderByIdRequest{Id: createdOrderID}
		getOrderResp, err := orderClient.GetOrderById(ctx, getOrderReq)
		if err != nil {
			log.Printf("Error getting order by ID: %v", err)
		} else {
			fmt.Printf("Order details: ID: %s, Status: %s, Items: %d\n",
				getOrderResp.GetOrder().GetId(), getOrderResp.GetOrder().GetStatus(), len(getOrderResp.GetOrder().GetItems()))
			for _, item := range getOrderResp.GetOrder().GetItems() {
				fmt.Printf("  - Product: %s (Qty: %d, Price: %.2f)\n", item.GetProductName(), item.GetQuantity(), item.GetPrice())
			}
		}
	} else {
		fmt.Println("Cannot get order details as no order was created.")
	}

	// --- Test Case 3: Update order status ---
	fmt.Println("\n--- Updating order status to 'paid' ---")
	if createdOrderID != "" {
		updateStatusReq := &order_client.UpdateOrderStatusRequest{
			OrderId:   createdOrderID,
			NewStatus: "paid",
		}
		updateStatusResp, err := orderClient.UpdateOrderStatus(ctx, updateStatusReq)
		if err != nil {
			log.Printf("Error updating order status: %v", err)
		} else {
			fmt.Printf("Order status updated! ID: %s, New Status: %s\n",
				updateStatusResp.GetOrder().GetId(), updateStatusResp.GetOrder().GetStatus())
		}
	} else {
		fmt.Println("Cannot update order status as no order was created.")
	}

	// --- Test Case 4: List orders ---
	fmt.Println("\n--- Listing orders ---")
	listOrdersReq := &order_client.ListOrdersRequest{
		UserId: dummyUserID, // Filter by user ID
		Limit:  10,
		Offset: 0,
	}
	listOrdersResp, err := orderClient.ListOrders(ctx, listOrdersReq)
	if err != nil {
		log.Printf("Error listing orders: %v", err)
	} else {
		fmt.Printf("Total orders found: %d\n", listOrdersResp.GetTotalCount())
		for _, order := range listOrdersResp.GetOrders() {
			fmt.Printf("  - Order ID: %s, Status: %s, Total: %.2f\n", order.GetId(), order.GetStatus(), order.GetTotalAmount())
		}
	}

	// --- Test Case 5: Cancel order ---
	fmt.Println("\n--- Cancelling order ---")
	if createdOrderID != "" {
		cancelOrderReq := &order_client.CancelOrderRequest{OrderId: createdOrderID}
		cancelOrderResp, err := orderClient.CancelOrder(ctx, cancelOrderReq)
		if err != nil {
			log.Printf("Error cancelling order: %v", err)
		} else {
			fmt.Printf("Order cancelled! ID: %s, New Status: %s\n",
				cancelOrderResp.GetOrder().GetId(), cancelOrderResp.GetOrder().GetStatus())
		}
	} else {
		fmt.Println("Cannot cancel order as no order was created.")
	}

	fmt.Println("\nFinished Order Service test cases.")
	time.Sleep(1 * time.Second) // Give some time for logs
}
