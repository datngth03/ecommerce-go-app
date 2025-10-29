package handler

import (
	"net/http"
	"strconv"
	"time"

	pb "github.com/datngth03/ecommerce-go-app/proto/inventory_service"
	"github.com/datngth03/ecommerce-go-app/services/api-gateway/internal/clients"
	"github.com/datngth03/ecommerce-go-app/services/api-gateway/internal/metrics"
	"github.com/gin-gonic/gin"
)

type InventoryHandler struct {
	inventoryClient *clients.InventoryClient
}

func NewInventoryHandler(inventoryClient *clients.InventoryClient) *InventoryHandler {
	return &InventoryHandler{
		inventoryClient: inventoryClient,
	}
}

// GetStock handles GET /api/v1/inventory/:product_id
func (h *InventoryHandler) GetStock(c *gin.Context) {
	productID := c.Param("product_id")
	if productID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "product_id is required"})
		return
	}

	start := time.Now()
	resp, err := h.inventoryClient.GetStock(c.Request.Context(), &pb.GetStockRequest{
		ProductId: productID,
	})

	status := "success"
	if err != nil {
		status = "error"
		metrics.RecordGRPCClientRequest("inventory-service", "GetStock", status, time.Since(start))
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	metrics.RecordGRPCClientRequest("inventory-service", "GetStock", status, time.Since(start))

	c.JSON(http.StatusOK, gin.H{
		"message": "stock retrieved successfully",
		"data":    resp.Stock,
	})
}

// UpdateStock handles PUT /api/v1/inventory/:product_id (Admin only)
func (h *InventoryHandler) UpdateStock(c *gin.Context) {
	productID := c.Param("product_id")
	if productID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "product_id is required"})
		return
	}

	var req struct {
		Quantity int32  `json:"quantity" binding:"required"`
		Reason   string `json:"reason"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, err := h.inventoryClient.UpdateStock(c.Request.Context(), &pb.UpdateStockRequest{
		ProductId: productID,
		Quantity:  req.Quantity,
		Reason:    req.Reason,
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "stock updated successfully",
		"data":    resp.Stock,
	})
}

// CheckAvailability handles POST /api/v1/inventory/check-availability
func (h *InventoryHandler) CheckAvailability(c *gin.Context) {
	var req struct {
		Items []struct {
			ProductID string `json:"product_id" binding:"required"`
			Quantity  int32  `json:"quantity" binding:"required,min=1"`
		} `json:"items" binding:"required,min=1"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Convert to proto format
	items := make([]*pb.StockItem, len(req.Items))
	for i, item := range req.Items {
		items[i] = &pb.StockItem{
			ProductId: item.ProductID,
			Quantity:  item.Quantity,
		}
	}

	resp, err := h.inventoryClient.CheckAvailability(c.Request.Context(), &pb.CheckAvailabilityRequest{
		Items: items,
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":           "availability checked successfully",
		"available":         resp.Available,
		"unavailable_items": resp.UnavailableItems,
	})
}

// GetStockHistory handles GET /api/v1/inventory/:product_id/history (Admin only)
func (h *InventoryHandler) GetStockHistory(c *gin.Context) {
	productID := c.Param("product_id")
	if productID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "product_id is required"})
		return
	}

	page, _ := strconv.ParseInt(c.DefaultQuery("page", "1"), 10, 32)
	pageSize, _ := strconv.ParseInt(c.DefaultQuery("page_size", "20"), 10, 32)

	resp, err := h.inventoryClient.GetStockHistory(c.Request.Context(), &pb.GetStockHistoryRequest{
		ProductId: productID,
		Limit:     int32(pageSize),
		Offset:    int32((page - 1) * pageSize),
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":   "stock history retrieved successfully",
		"data":      resp.Movements,
		"total":     resp.Total,
		"page":      page,
		"page_size": pageSize,
	})
}
