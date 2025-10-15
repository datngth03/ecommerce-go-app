package handler

import (
	"net/http"

	"github.com/datngth03/ecommerce-go-app/services/order-service/internal/service"
	"github.com/gin-gonic/gin"
)

type CartHandler struct {
	cartService *service.CartService
}

func NewCartHandler(cartService *service.CartService) *CartHandler {
	return &CartHandler{
		cartService: cartService,
	}
}

// GetCart godoc
// @Summary Get user's cart
// @Description Retrieve current shopping cart
// @Tags cart
// @Produce json
// @Success 200 {object} CartResponse
// @Router /cart [get]
func (h *CartHandler) GetCart(c *gin.Context) {
	userID := getUserIDFromContext(c)

	cart, err := h.cartService.GetCart(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    cart,
	})
}

// AddToCart godoc
// @Summary Add item to cart
// @Description Add product to shopping cart
// @Tags cart
// @Accept json
// @Produce json
// @Param request body AddToCartRequest true "Add to Cart Request"
// @Success 200 {object} CartResponse
// @Failure 400 {object} ErrorResponse
// @Router /cart/items [post]
func (h *CartHandler) AddToCart(c *gin.Context) {
	userID := getUserIDFromContext(c)

	var req AddToCartRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	cart, err := h.cartService.AddToCart(c.Request.Context(), userID, req.ProductID, req.Quantity)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Item added to cart",
		"data":    cart,
	})
}

// UpdateCartItem godoc
// @Summary Update cart item quantity
// @Description Update quantity of item in cart
// @Tags cart
// @Accept json
// @Produce json
// @Param product_id path string true "Product ID"
// @Param request body UpdateCartItemRequest true "Update Cart Item Request"
// @Success 200 {object} CartResponse
// @Failure 400 {object} ErrorResponse
// @Router /cart/items/{product_id} [patch]
func (h *CartHandler) UpdateCartItem(c *gin.Context) {
	userID := getUserIDFromContext(c)
	productID := c.Param("product_id")

	var req UpdateCartItemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	cart, err := h.cartService.UpdateCartItem(c.Request.Context(), userID, productID, req.Quantity)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Cart item updated",
		"data":    cart,
	})
}

// RemoveFromCart godoc
// @Summary Remove item from cart
// @Description Remove product from shopping cart
// @Tags cart
// @Produce json
// @Param product_id path string true "Product ID"
// @Success 200 {object} CartResponse
// @Router /cart/items/{product_id} [delete]
func (h *CartHandler) RemoveFromCart(c *gin.Context) {
	userID := getUserIDFromContext(c)
	productID := c.Param("product_id")

	cart, err := h.cartService.RemoveFromCart(c.Request.Context(), userID, productID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Item removed from cart",
		"data":    cart,
	})
}

// ClearCart godoc
// @Summary Clear cart
// @Description Remove all items from cart
// @Tags cart
// @Produce json
// @Success 200 {object} SuccessResponse
// @Router /cart [delete]
func (h *CartHandler) ClearCart(c *gin.Context) {
	userID := getUserIDFromContext(c)

	err := h.cartService.ClearCart(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Cart cleared successfully",
	})
}

// Request types

type AddToCartRequest struct {
	ProductID string `json:"product_id" binding:"required"`
	Quantity  int32  `json:"quantity" binding:"required,gt=0"`
}

type UpdateCartItemRequest struct {
	Quantity int32 `json:"quantity" binding:"required,gt=0"`
}
