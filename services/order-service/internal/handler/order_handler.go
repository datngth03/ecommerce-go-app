package handler

import (
	"net/http"
	"strconv"

	"github.com/datngth03/ecommerce-go-app/services/order-service/internal/service"
	"github.com/gin-gonic/gin"
)

type OrderHandler struct {
	orderService *service.OrderService
}

func NewOrderHandler(orderService *service.OrderService) *OrderHandler {
	return &OrderHandler{
		orderService: orderService,
	}
}

// CreateOrder godoc
// @Summary Create a new order
// @Description Create order from cart items
// @Tags orders
// @Accept json
// @Produce json
// @Param request body CreateOrderRequest true "Create Order Request"
// @Success 201 {object} OrderResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /orders [post]
func (h *OrderHandler) CreateOrder(c *gin.Context) {
	var req CreateOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get user ID from context (set by auth middleware)
	userID := getUserIDFromContext(c)

	order, err := h.orderService.CreateOrder(c.Request.Context(), userID, req.ShippingAddress, req.PaymentMethod)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"data":    order,
	})
}

// GetOrder godoc
// @Summary Get order by ID
// @Description Retrieve order details
// @Tags orders
// @Produce json
// @Param id path string true "Order ID"
// @Success 200 {object} OrderResponse
// @Failure 404 {object} ErrorResponse
// @Router /orders/{id} [get]
func (h *OrderHandler) GetOrder(c *gin.Context) {
	orderID := c.Param("id")
	userID := getUserIDFromContext(c)

	order, err := h.orderService.GetOrder(c.Request.Context(), orderID, userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    order,
	})
}

// ListOrders godoc
// @Summary List user's orders
// @Description Get paginated list of orders
// @Tags orders
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param page_size query int false "Page size" default(10)
// @Param status query string false "Order status filter"
// @Success 200 {object} OrderListResponse
// @Router /orders [get]
func (h *OrderHandler) ListOrders(c *gin.Context) {
	userID := getUserIDFromContext(c)

	page, _ := strconv.ParseInt(c.DefaultQuery("page", "1"), 10, 32)
	pageSize, _ := strconv.ParseInt(c.DefaultQuery("page_size", "10"), 10, 32)
	status := c.Query("status")

	orders, total, err := h.orderService.ListOrders(c.Request.Context(), userID, int32(page), int32(pageSize), status)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"orders":      orders,
			"total_count": total,
			"page":        page,
			"page_size":   pageSize,
		},
	})
}

// UpdateOrderStatus godoc
// @Summary Update order status
// @Description Update the status of an order
// @Tags orders
// @Accept json
// @Produce json
// @Param id path string true "Order ID"
// @Param request body UpdateOrderStatusRequest true "Update Status Request"
// @Success 200 {object} OrderResponse
// @Failure 400 {object} ErrorResponse
// @Router /orders/{id}/status [patch]
func (h *OrderHandler) UpdateOrderStatus(c *gin.Context) {
	orderID := c.Param("id")
	userID := getUserIDFromContext(c)

	var req UpdateOrderStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	order, err := h.orderService.UpdateOrderStatus(c.Request.Context(), orderID, req.Status, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    order,
	})
}

// CancelOrder godoc
// @Summary Cancel an order
// @Description Cancel a pending order
// @Tags orders
// @Produce json
// @Param id path string true "Order ID"
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} ErrorResponse
// @Router /orders/{id}/cancel [post]
func (h *OrderHandler) CancelOrder(c *gin.Context) {
	orderID := c.Param("id")
	userID := getUserIDFromContext(c)

	err := h.orderService.CancelOrder(c.Request.Context(), orderID, userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Order cancelled successfully",
	})
}

// Request/Response types

type CreateOrderRequest struct {
	ShippingAddress string `json:"shipping_address" binding:"required"`
	PaymentMethod   string `json:"payment_method" binding:"required"`
}

type UpdateOrderStatusRequest struct {
	Status string `json:"status" binding:"required"`
}
