// internal/cart/domain/cart.go
package domain

import (
	"time"
)

// CartItem represents a product within a shopping cart.
type CartItem struct {
	ProductID   string  `json:"product_id"`
	ProductName string  `json:"product_name"`
	Price       float64 `json:"price"`
	Quantity    int32   `json:"quantity"`
}

// Cart represents the core shopping cart entity in the domain.
type Cart struct {
	UserID        string     `json:"user_id"`
	Items         []CartItem `json:"items"`
	TotalAmount   float64    `json:"total_amount"`
	LastUpdatedAt time.Time  `json:"last_updated_at"`
}

// NewCart creates a new Cart instance for a given user.
func NewCart(userID string) *Cart {
	now := time.Now()
	return &Cart{
		UserID:        userID,
		Items:         []CartItem{},
		TotalAmount:   0.0,
		LastUpdatedAt: now,
	}
}

// AddItem adds a product to the cart or updates its quantity if it already exists.
func (c *Cart) AddItem(productID, productName string, price float64, quantity int32) {
	for i, item := range c.Items {
		if item.ProductID == productID {
			// Update quantity if item already exists
			c.Items[i].Quantity += quantity
			c.recalculateTotal()
			c.LastUpdatedAt = time.Now()
			return
		}
	}
	// Add new item
	c.Items = append(c.Items, CartItem{
		ProductID:   productID,
		ProductName: productName,
		Price:       price,
		Quantity:    quantity,
	})
	c.recalculateTotal()
	c.LastUpdatedAt = time.Now()
}

// UpdateItemQuantity updates the quantity of a specific item in the cart.
func (c *Cart) UpdateItemQuantity(productID string, newQuantity int32) bool {
	if newQuantity <= 0 {
		return c.RemoveItem(productID) // Remove if quantity is 0 or less
	}
	for i, item := range c.Items {
		if item.ProductID == productID {
			c.Items[i].Quantity = newQuantity
			c.recalculateTotal()
			c.LastUpdatedAt = time.Now()
			return true
		}
	}
	return false // Item not found
}

// RemoveItem removes a product from the cart.
func (c *Cart) RemoveItem(productID string) bool {
	for i, item := range c.Items {
		if item.ProductID == productID {
			c.Items = append(c.Items[:i], c.Items[i+1:]...) // Remove item
			c.recalculateTotal()
			c.LastUpdatedAt = time.Now()
			return true
		}
	}
	return false // Item not found
}

// Clear clears all items from the cart.
func (c *Cart) Clear() {
	c.Items = []CartItem{}
	c.TotalAmount = 0.0
	c.LastUpdatedAt = time.Now()
}

// recalculateTotal calculates the total amount of the cart.
func (c *Cart) recalculateTotal() {
	total := 0.0
	for _, item := range c.Items {
		total += item.Price * float64(item.Quantity)
	}
	c.TotalAmount = total
}
