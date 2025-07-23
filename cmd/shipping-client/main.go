// cmd/shipping-client/main.go
package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid" // For generating dummy IDs
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	order_client "github.com/datngth03/ecommerce-go-app/pkg/client/order"       // Generated Order gRPC client
	product_client "github.com/datngth03/ecommerce-go-app/pkg/client/product"   // Generated Product gRPC client
	shipping_client "github.com/datngth03/ecommerce-go-app/pkg/client/shipping" // Generated Shipping gRPC client
)

func main() {
	// Addresses of services
	shippingSvcAddr := "localhost:50056"
	orderSvcAddr := "localhost:50053"
	productSvcAddr := "localhost:50052" // Needed to create product for order

	// --- Connect to Shipping Service ---
	shippingConn, err := grpc.Dial(shippingSvcAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Failed to connect to Shipping Service: %v", err)
	}
	defer shippingConn.Close()
	shippingClient := shipping_client.NewShippingServiceClient(shippingConn)

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
	var productPrice float64 = 75.00

	fmt.Println("\n--- Preparing test data (Category & Product for Order) ---")
	createCategoryReq := &product_client.CreateCategoryRequest{Name: "ShippingTestCategory", Description: "Category for shipping testing"}
	createCategoryResp, err := productClient.CreateCategory(ctx, createCategoryReq)
	if err != nil {
		log.Printf("Error creating category (might already exist): %v", err)
		listCategoriesResp, listErr := productClient.ListCategories(ctx, &product_client.ListCategoriesRequest{})
		if listErr == nil {
			for _, cat := range listCategoriesResp.GetCategories() {
				if cat.GetName() == "ShippingTestCategory" {
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
		Name:        "ShippingTestProduct",
		Description: "Product for shipping testing.",
		Price:       productPrice,
		CategoryId:  categoryID,
		ImageUrls:   []string{"http://example.com/shipping-test-product.jpg"},
	}
	createProductResp, err := productClient.CreateProduct(ctx, createProductReq)
	if err != nil {
		log.Printf("Error creating product (might already exist): %v", err)
		listProductsResp, listErr := productClient.ListProducts(ctx, &product_client.ListProductsRequest{CategoryId: categoryID})
		if listErr == nil {
			for _, prod := range listProductsResp.GetProducts() {
				if prod.GetName() == "ShippingTestProduct" {
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

	// Create a dummy order for shipment
	var orderID string
	fmt.Println("\n--- Creating a dummy order for shipment ---")
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
		ShippingAddress: "789 Shipping Lane, Delivery City, USA",
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
			log.Fatalf("Failed to get or create a dummy order. Cannot proceed with shipping operations.")
		}
	} else {
		orderID = createOrderResp.GetOrder().GetId()
		log.Printf("Created dummy order: %s", orderID)
	}

	// Ensure order status is 'paid' for shipment creation
	_, err = orderClient.UpdateOrderStatus(ctx, &order_client.UpdateOrderStatusRequest{
		OrderId:   orderID,
		NewStatus: "paid",
	})
	if err != nil {
		log.Printf("Error updating order status to paid: %v", err)
		// If it's already paid, that's fine. If not, this might be a problem.
	} else {
		log.Printf("Order %s status updated to 'paid'.", orderID)
	}

	var createdShipmentID string
	var trackingNumber string

	// --- Test Case 1: Calculate Shipping Cost ---
	fmt.Println("\n--- Calculating shipping cost ---")
	calcCostReq := &shipping_client.CalculateShippingCostRequest{
		OrderId:         orderID,
		ShippingAddress: "789 Shipping Lane, Delivery City, USA",
		OriginAddress:   "101 Warehouse Rd, Origin City, USA",
	}
	calcCostResp, err := shippingClient.CalculateShippingCost(ctx, calcCostReq)
	if err != nil {
		log.Printf("Error calculating shipping cost: %v", err)
	} else {
		fmt.Printf("Calculated Shipping Cost: %.2f %s, Estimated Delivery: %s\n",
			calcCostResp.GetCost(), calcCostResp.GetCurrency(), calcCostResp.GetEstimatedDeliveryTime())
	}

	// --- Test Case 2: Create a new shipment ---
	fmt.Println("\n--- Creating a new shipment ---")
	createShipmentReq := &shipping_client.CreateShipmentRequest{
		OrderId:         orderID,
		UserId:          dummyUserID,
		ShippingCost:    calcCostResp.GetCost(), // Use calculated cost
		ShippingAddress: "789 Shipping Lane, Delivery City, USA",
		Carrier:         "FedEx",
	}
	createShipmentResp, err := shippingClient.CreateShipment(ctx, createShipmentReq)
	if err != nil {
		log.Printf("Error creating shipment: %v", err)
	} else {
		createdShipmentID = createShipmentResp.GetShipment().GetId()
		trackingNumber = createShipmentResp.GetShipment().GetTrackingNumber()
		fmt.Printf("Shipment created successfully! ID: %s, Order ID: %s, Tracking: %s, Status: %s\n",
			createdShipmentID, createShipmentResp.GetShipment().GetOrderId(),
			trackingNumber, createShipmentResp.GetShipment().GetStatus())
	}

	// --- Test Case 3: Get shipment by ID ---
	fmt.Println("\n--- Getting shipment by ID ---")
	if createdShipmentID != "" {
		getShipmentReq := &shipping_client.GetShipmentByIdRequest{Id: createdShipmentID}
		getShipmentResp, err := shippingClient.GetShipmentById(ctx, getShipmentReq)
		if err != nil {
			log.Printf("Error getting shipment by ID: %v", err)
		} else {
			fmt.Printf("Shipment details: ID: %s, Status: %s, Carrier: %s, Tracking: %s\n",
				getShipmentResp.GetShipment().GetId(), getShipmentResp.GetShipment().GetStatus(),
				getShipmentResp.GetShipment().GetCarrier(), getShipmentResp.GetShipment().GetTrackingNumber())
		}
	} else {
		fmt.Println("Cannot get shipment details as no shipment was created.")
	}

	// --- Test Case 4: Update shipment status to 'delivered' ---
	fmt.Println("\n--- Updating shipment status to 'delivered' ---")
	if createdShipmentID != "" {
		updateStatusReq := &shipping_client.UpdateShipmentStatusRequest{
			ShipmentId:     createdShipmentID,
			NewStatus:      "delivered",
			TrackingNumber: trackingNumber, // Pass tracking number if it was updated
		}
		updateStatusResp, err := shippingClient.UpdateShipmentStatus(ctx, updateStatusReq)
		if err != nil {
			log.Printf("Error updating shipment status: %v", err)
		} else {
			fmt.Printf("Shipment status updated! ID: %s, New Status: %s\n",
				updateStatusResp.GetShipment().GetId(), updateStatusResp.GetShipment().GetStatus())
		}
	} else {
		fmt.Println("Cannot update shipment status as no shipment was created.")
	}

	// --- Test Case 5: Track shipment ---
	fmt.Println("\n--- Tracking shipment ---")
	if createdShipmentID != "" {
		trackShipmentReq := &shipping_client.TrackShipmentRequest{ShipmentId: createdShipmentID}
		trackShipmentResp, err := shippingClient.TrackShipment(ctx, trackShipmentReq)
		if err != nil {
			log.Printf("Error tracking shipment: %v", err)
		} else {
			fmt.Printf("Shipment tracking details: ID: %s, Status: %s, Carrier: %s\n",
				trackShipmentResp.GetShipment().GetId(), trackShipmentResp.GetShipment().GetStatus(), trackShipmentResp.GetShipment().GetCarrier())
		}
	} else {
		fmt.Println("Cannot track shipment as no shipment was created.")
	}

	// --- Test Case 6: List shipments ---
	fmt.Println("\n--- Listing shipments for user ---")
	listShipmentsReq := &shipping_client.ListShipmentsRequest{
		UserId: dummyUserID, // Filter by user ID
		Limit:  10,
		Offset: 0,
	}
	listShipmentsResp, err := shippingClient.ListShipments(ctx, listShipmentsReq)
	if err != nil {
		log.Printf("Error listing shipments: %v", err)
	} else {
		fmt.Printf("Total shipments found: %d\n", listShipmentsResp.GetTotalCount())
		for _, shipment := range listShipmentsResp.GetShipments() {
			fmt.Printf("  - Shipment ID: %s, Status: %s, Tracking: %s, Cost: %.2f\n",
				shipment.GetId(), shipment.GetStatus(), shipment.GetTrackingNumber(), shipment.GetShippingCost())
		}
	}

	fmt.Println("\nFinished Shipping Service test cases.")
	time.Sleep(1 * time.Second) // Give some time for logs
}
