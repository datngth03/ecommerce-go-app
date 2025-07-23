// cmd/payment-client/main.go
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
	payment_client "github.com/datngth03/ecommerce-go-app/pkg/client/payment" // Generated Payment gRPC client
	product_client "github.com/datngth03/ecommerce-go-app/pkg/client/product" // Generated Product gRPC client
)

func main() {
	// Addresses of services
	paymentSvcAddr := "localhost:50055"
	orderSvcAddr := "localhost:50053"
	productSvcAddr := "localhost:50052" // Needed to create product for order

	// --- Connect to Payment Service ---
	paymentConn, err := grpc.Dial(paymentSvcAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Failed to connect to Payment Service: %v", err)
	}
	defer paymentConn.Close()
	paymentClient := payment_client.NewPaymentServiceClient(paymentConn)

	// --- Connect to Order Service ---
	orderConn, err := grpc.Dial(orderSvcAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Failed to connect to Order Service: %v", err)
	}
	defer orderConn.Close()
	orderClient := order_client.NewOrderServiceClient(orderConn)

	// --- Connect to Product Service (to create dummy product for order) ---
	productConn, err := grpc.Dial(productSvcAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Failed to connect to Product Service: %v", err)
	}
	defer productConn.Close()
	productClient := product_client.NewProductServiceClient(productConn)

	// Context with timeout for RPC calls
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second) // Increased timeout for multiple calls
	defer cancel()

	// --- Prepare dummy data: User, Category, Product, Order ---
	dummyUserID := uuid.New().String()
	log.Printf("Using dummy User ID: %s", dummyUserID)

	// Create a dummy category and product first if they don't exist
	var categoryID string
	var productID string
	var productName string
	var productPrice float64 = 50.00

	fmt.Println("\n--- Preparing test data (Category & Product for Order) ---")
	createCategoryReq := &product_client.CreateCategoryRequest{Name: "PaymentTestCategory", Description: "Category for payment testing"}
	createCategoryResp, err := productClient.CreateCategory(ctx, createCategoryReq)
	if err != nil {
		log.Printf("Error creating category (might already exist): %v", err)
		listCategoriesResp, listErr := productClient.ListCategories(ctx, &product_client.ListCategoriesRequest{})
		if listErr == nil {
			for _, cat := range listCategoriesResp.GetCategories() {
				if cat.GetName() == "PaymentTestCategory" {
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
		Name:        "PaymentTestProduct",
		Description: "Product for payment testing.",
		Price:       productPrice,
		CategoryId:  categoryID,
		ImageUrls:   []string{"http://example.com/payment-test-product.jpg"},
	}
	createProductResp, err := productClient.CreateProduct(ctx, createProductReq)
	if err != nil {
		log.Printf("Error creating product (might already exist): %v", err)
		listProductsResp, listErr := productClient.ListProducts(ctx, &product_client.ListProductsRequest{CategoryId: categoryID})
		if listErr == nil {
			for _, prod := range listProductsResp.GetProducts() {
				if prod.GetName() == "PaymentTestProduct" {
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

	// Create a dummy order for payment
	var orderID string
	fmt.Println("\n--- Creating a dummy order for payment ---")
	createOrderReq := &order_client.CreateOrderRequest{
		UserId: dummyUserID,
		Items: []*order_client.OrderItem{
			{
				ProductId:   productID,
				ProductName: productName,
				Price:       productPrice,
				Quantity:    1,
			},
		},
		ShippingAddress: "456 Payment St, Test City, USA",
	}
	createOrderResp, err := orderClient.CreateOrder(ctx, createOrderReq)
	if err != nil {
		log.Printf("Error creating dummy order: %v", err)
		// Try to find if order already exists for this user/product
		listOrdersResp, listErr := orderClient.ListOrders(ctx, &order_client.ListOrdersRequest{UserId: dummyUserID})
		if listErr == nil && len(listOrdersResp.GetOrders()) > 0 {
			orderID = listOrdersResp.GetOrders()[0].GetId()
			log.Printf("Found existing order for user: %s", orderID)
		} else {
			log.Fatalf("Failed to get or create a dummy order. Cannot proceed with payment operations.")
		}
	} else {
		orderID = createOrderResp.GetOrder().GetId()
		log.Printf("Created dummy order: %s", orderID)
	}

	var createdPaymentID string

	// --- Test Case 1: Create a new payment ---
	fmt.Println("\n--- Creating a new payment ---")
	createPaymentReq := &payment_client.CreatePaymentRequest{
		OrderId:       orderID,
		UserId:        dummyUserID,
		Amount:        productPrice, // Amount should match order total
		Currency:      "USD",
		PaymentMethod: "credit_card",
	}
	createPaymentResp, err := paymentClient.CreatePayment(ctx, createPaymentReq)
	if err != nil {
		log.Printf("Error creating payment: %v", err)
	} else {
		createdPaymentID = createPaymentResp.GetPayment().GetId()
		fmt.Printf("Payment created successfully! ID: %s, Order ID: %s, Amount: %.2f, Status: %s\n",
			createdPaymentID, createPaymentResp.GetPayment().GetOrderId(),
			createPaymentResp.GetPayment().GetAmount(), createPaymentResp.GetPayment().GetStatus())
	}

	// --- Test Case 2: Get payment by ID ---
	fmt.Println("\n--- Getting payment by ID ---")
	if createdPaymentID != "" {
		getPaymentReq := &payment_client.GetPaymentByIdRequest{Id: createdPaymentID}
		getPaymentResp, err := paymentClient.GetPaymentById(ctx, getPaymentReq)
		if err != nil {
			log.Printf("Error getting payment by ID: %v", err)
		} else {
			fmt.Printf("Payment details: ID: %s, Status: %s, Transaction ID: %s\n",
				getPaymentResp.GetPayment().GetId(), getPaymentResp.GetPayment().GetStatus(), getPaymentResp.GetPayment().GetTransactionId())
		}
	} else {
		fmt.Println("Cannot get payment details as no payment was created.")
	}

	// --- Test Case 3: Confirm payment (simulate gateway callback) ---
	fmt.Println("\n--- Confirming payment ---")
	if createdPaymentID != "" {
		confirmPaymentReq := &payment_client.ConfirmPaymentRequest{
			PaymentId:     createdPaymentID,
			TransactionId: uuid.New().String(), // Dummy transaction ID from gateway
			Status:        "completed",         // Simulate successful payment
		}
		confirmPaymentResp, err := paymentClient.ConfirmPayment(ctx, confirmPaymentReq)
		if err != nil {
			log.Printf("Error confirming payment: %v", err)
		} else {
			fmt.Printf("Payment confirmed! ID: %s, New Status: %s, Transaction ID: %s\n",
				confirmPaymentResp.GetPayment().GetId(), confirmPaymentResp.GetPayment().GetStatus(), confirmPaymentResp.GetPayment().GetTransactionId())
		}
	} else {
		fmt.Println("Cannot confirm payment as no payment was created.")
	}

	// --- Test Case 4: List payments ---
	fmt.Println("\n--- Listing payments for user ---")
	listPaymentsReq := &payment_client.ListPaymentsRequest{
		UserId: dummyUserID, // Filter by user ID
		Limit:  10,
		Offset: 0,
	}
	listPaymentsResp, err := paymentClient.ListPayments(ctx, listPaymentsReq)
	if err != nil {
		log.Printf("Error listing payments: %v", err)
	} else {
		fmt.Printf("Total payments found: %d\n", listPaymentsResp.GetTotalCount())
		for _, payment := range listPaymentsResp.GetPayments() {
			fmt.Printf("  - Payment ID: %s, Order ID: %s, Status: %s, Amount: %.2f\n",
				payment.GetId(), payment.GetOrderId(), payment.GetStatus(), payment.GetAmount())
		}
	}

	// --- Test Case 5: Refund payment (optional, requires completed status) ---
	fmt.Println("\n--- Refunding payment ---")
	if createdPaymentID != "" && createPaymentResp.GetPayment().GetStatus() == "completed" { // Only if it was completed
		refundPaymentReq := &payment_client.RefundPaymentRequest{
			PaymentId:    createdPaymentID,
			RefundAmount: productPrice, // Full refund
		}
		refundPaymentResp, err := paymentClient.RefundPayment(ctx, refundPaymentReq)
		if err != nil {
			log.Printf("Error refunding payment: %v", err)
		} else {
			fmt.Printf("Payment refunded! ID: %s, New Status: %s\n",
				refundPaymentResp.GetPayment().GetId(), refundPaymentResp.GetPayment().GetStatus())
		}
	} else {
		fmt.Println("Cannot refund payment (either not created or not completed).")
	}

	fmt.Println("\nFinished Payment Service test cases.")
	time.Sleep(1 * time.Second) // Give some time for logs
}
