package handler

import (
	"net/http"
	"strconv"

	pb "github.com/datngth03/ecommerce-go-app/proto/order_service"
	"github.com/datngth03/ecommerce-go-app/services/api-gateway/internal/clients"
	"github.com/gin-gonic/gin"
)

type OrderHandler struct {
	orderClient *clients.OrderClient
}

func NewOrderHandler(orderClient *clients.OrderClient) *OrderHandler {
	return &OrderHandler{
		orderClient: orderClient,
	}
}

// CreateOrder handles POST /api/v1/orders
func (h *OrderHandler) CreateOrder(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var req struct {
		ShippingAddress string `json:"shipping_address" binding:"required"`
		PaymentMethod   string `json:"payment_method" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, err := h.orderClient.CreateOrder(c.Request.Context(), &pb.CreateOrderRequest{
		UserId:          userID.(int64),
		ShippingAddress: req.ShippingAddress,
		PaymentMethod:   req.PaymentMethod,
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "order created successfully",
		"data":    resp.Order,
	})
}

// GetOrder handles GET /api/v1/orders/:id
func (h *OrderHandler) GetOrder(c *gin.Context) {
	orderID := c.Param("id")
	if orderID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "order_id is required"})
		return
	}

	resp, err := h.orderClient.GetOrder(c.Request.Context(), &pb.GetOrderRequest{
		Id: orderID,
	})

	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "order retrieved successfully",
		"data":    resp.Order,
	})
}

// ListOrders handles GET /api/v1/orders
func (h *OrderHandler) ListOrders(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	page, _ := strconv.ParseInt(c.DefaultQuery("page", "1"), 10, 32)
	pageSize, _ := strconv.ParseInt(c.DefaultQuery("page_size", "10"), 10, 32)
	status := c.Query("status")

	resp, err := h.orderClient.ListOrders(c.Request.Context(), &pb.ListOrdersRequest{
		UserId:   userID.(int64),
		Page:     int32(page),
		PageSize: int32(pageSize),
		Status:   status,
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":   "orders retrieved successfully",
		"data":      resp.Orders,
		"total":     resp.TotalCount,
		"page":      page,
		"page_size": pageSize,
	})
}

// CancelOrder handles DELETE /api/v1/orders/:id
func (h *OrderHandler) CancelOrder(c *gin.Context) {
	orderID := c.Param("id")
	if orderID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "order_id is required"})
		return
	}

	userID, _ := c.Get("user_id")

	err := h.orderClient.CancelOrder(c.Request.Context(), &pb.CancelOrderRequest{
		Id:     orderID,
		UserId: userID.(int64),
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "order cancelled successfully",
	})
}

// AddToCart handles POST /api/v1/cart
func (h *OrderHandler) AddToCart(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var req struct {
		ProductID string `json:"product_id" binding:"required"`
		Quantity  int32  `json:"quantity" binding:"required,min=1"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, err := h.orderClient.AddToCart(c.Request.Context(), &pb.AddToCartRequest{
		UserId:    userID.(int64),
		ProductId: req.ProductID,
		Quantity:  req.Quantity,
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "item added to cart successfully",
		"data":    resp.Cart,
	})
}

// GetCart handles GET /api/v1/cart
func (h *OrderHandler) GetCart(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	resp, err := h.orderClient.GetCart(c.Request.Context(), &pb.GetCartRequest{
		UserId: userID.(int64),
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "cart retrieved successfully",
		"data":    resp.Cart,
	})
}

// UpdateCartItem handles PUT /api/v1/cart/:product_id
func (h *OrderHandler) UpdateCartItem(c *gin.Context) {
	productID := c.Param("product_id")
	if productID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "product_id is required"})
		return
	}

	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var req struct {
		Quantity int32 `json:"quantity" binding:"required,min=1"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, err := h.orderClient.UpdateCartItem(c.Request.Context(), &pb.UpdateCartItemRequest{
		UserId:    userID.(int64),
		ProductId: productID,
		Quantity:  req.Quantity,
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "cart item updated successfully",
		"data":    resp.Cart,
	})
}

// RemoveFromCart handles DELETE /api/v1/cart/:product_id
func (h *OrderHandler) RemoveFromCart(c *gin.Context) {
	productID := c.Param("product_id")
	if productID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "product_id is required"})
		return
	}

	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	resp, err := h.orderClient.RemoveFromCart(c.Request.Context(), &pb.RemoveFromCartRequest{
		UserId:    userID.(int64),
		ProductId: productID,
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "item removed from cart successfully",
		"data":    resp.Cart,
	})
}

// ClearCart handles DELETE /api/v1/cart
func (h *OrderHandler) ClearCart(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	err := h.orderClient.ClearCart(c.Request.Context(), &pb.ClearCartRequest{
		UserId: userID.(int64),
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "cart cleared successfully",
	})
}
