package handler

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	pb "github.com/datngth03/ecommerce-go-app/proto/payment_service"
	"github.com/datngth03/ecommerce-go-app/services/api-gateway/internal/clients"
	"github.com/datngth03/ecommerce-go-app/services/api-gateway/internal/metrics"
	"github.com/gin-gonic/gin"
)

type PaymentHandler struct {
	paymentClient *clients.PaymentClient
}

func NewPaymentHandler(paymentClient *clients.PaymentClient) *PaymentHandler {
	return &PaymentHandler{
		paymentClient: paymentClient,
	}
}

// ProcessPayment handles POST /api/v1/payments
func (h *PaymentHandler) ProcessPayment(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var req struct {
		OrderID         string  `json:"order_id" binding:"required"`
		Amount          float64 `json:"amount" binding:"required,min=0"`
		Method          string  `json:"method" binding:"required"`
		Currency        string  `json:"currency"`
		PaymentMethodID string  `json:"payment_method_id"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.Currency == "" {
		req.Currency = "USD"
	}

	start := time.Now()
	resp, err := h.paymentClient.ProcessPayment(c.Request.Context(), &pb.ProcessPaymentRequest{
		OrderId:         req.OrderID,
		UserId:          fmt.Sprintf("%d", userID.(int64)),
		Amount:          req.Amount,
		Currency:        req.Currency,
		Method:          req.Method,
		PaymentMethodId: req.PaymentMethodID,
	})

	status := "success"
	if err != nil {
		status = "error"
		metrics.RecordGRPCClientRequest("payment-service", "ProcessPayment", status, time.Since(start))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	metrics.RecordGRPCClientRequest("payment-service", "ProcessPayment", status, time.Since(start))

	c.JSON(http.StatusCreated, gin.H{
		"message": "payment processed successfully",
		"data":    resp.Payment,
		"success": resp.Success,
	})
}

// GetPayment handles GET /api/v1/payments/:id
func (h *PaymentHandler) GetPayment(c *gin.Context) {
	paymentID := c.Param("id")
	if paymentID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "payment_id is required"})
		return
	}

	resp, err := h.paymentClient.GetPayment(c.Request.Context(), &pb.GetPaymentRequest{
		PaymentId: paymentID,
	})

	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "payment retrieved successfully",
		"data":    resp.Payment,
	})
}

// GetPaymentByOrder handles GET /api/v1/payments/order/:order_id
func (h *PaymentHandler) GetPaymentByOrder(c *gin.Context) {
	orderID := c.Param("order_id")
	if orderID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "order_id is required"})
		return
	}

	resp, err := h.paymentClient.GetPaymentByOrder(c.Request.Context(), &pb.GetPaymentByOrderRequest{
		OrderId: orderID,
	})

	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "payment retrieved successfully",
		"data":    resp.Payment,
	})
}

// GetPaymentHistory handles GET /api/v1/payments
func (h *PaymentHandler) GetPaymentHistory(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	page, _ := strconv.ParseInt(c.DefaultQuery("page", "1"), 10, 32)
	pageSize, _ := strconv.ParseInt(c.DefaultQuery("page_size", "10"), 10, 32)
	offset := (page - 1) * pageSize

	resp, err := h.paymentClient.GetPaymentHistory(c.Request.Context(), &pb.GetPaymentHistoryRequest{
		UserId: fmt.Sprintf("%d", userID.(int64)),
		Limit:  int32(pageSize),
		Offset: int32(offset),
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":   "payment history retrieved successfully",
		"data":      resp.Payments,
		"total":     resp.Total,
		"page":      page,
		"page_size": pageSize,
	})
}

// ConfirmPayment handles POST /api/v1/payments/:id/confirm
func (h *PaymentHandler) ConfirmPayment(c *gin.Context) {
	paymentID := c.Param("id")
	if paymentID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "payment_id is required"})
		return
	}

	var req struct {
		PaymentIntentID string `json:"payment_intent_id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, err := h.paymentClient.ConfirmPayment(c.Request.Context(), &pb.ConfirmPaymentRequest{
		PaymentId:       paymentID,
		PaymentIntentId: req.PaymentIntentID,
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "payment confirmed successfully",
		"data":    resp.Payment,
		"success": resp.Success,
	})
}

// RefundPayment handles POST /api/v1/payments/:id/refund
func (h *PaymentHandler) RefundPayment(c *gin.Context) {
	paymentID := c.Param("id")
	if paymentID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "payment_id is required"})
		return
	}

	var req struct {
		Amount float64 `json:"amount" binding:"required,min=0"`
		Reason string  `json:"reason"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, err := h.paymentClient.RefundPayment(c.Request.Context(), &pb.RefundPaymentRequest{
		PaymentId: paymentID,
		Amount:    req.Amount,
		Reason:    req.Reason,
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "payment refunded successfully",
		"data":    resp.Refund,
		"success": resp.Success,
	})
}

// SavePaymentMethod handles POST /api/v1/payment-methods
func (h *PaymentHandler) SavePaymentMethod(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var req struct {
		MethodType      string `json:"method_type" binding:"required"`
		GatewayMethodID string `json:"gateway_method_id" binding:"required"`
		IsDefault       bool   `json:"is_default"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, err := h.paymentClient.SavePaymentMethod(c.Request.Context(), &pb.SavePaymentMethodRequest{
		UserId:          fmt.Sprintf("%d", userID.(int64)),
		MethodType:      req.MethodType,
		GatewayMethodId: req.GatewayMethodID,
		IsDefault:       req.IsDefault,
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "payment method saved successfully",
		"data":    resp.PaymentMethod,
		"success": resp.Success,
	})
}

// GetPaymentMethods handles GET /api/v1/payment-methods
func (h *PaymentHandler) GetPaymentMethods(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	resp, err := h.paymentClient.GetPaymentMethods(c.Request.Context(), &pb.GetPaymentMethodsRequest{
		UserId: fmt.Sprintf("%d", userID.(int64)),
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "payment methods retrieved successfully",
		"data":    resp.PaymentMethods,
	})
}
