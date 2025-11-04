package validator

import (
	"errors"
	"fmt"
)

// --- Order Validation ---

// ValidateCreateOrderRequest validates order creation request
func ValidateCreateOrderRequest(userID string, items []OrderItem) error {
	// Validate user ID
	if err := ValidateRequired(userID, "user_id"); err != nil {
		return err
	}

	// Validate items
	if len(items) == 0 {
		return errors.New("order must contain at least one item")
	}

	if len(items) > 50 {
		return errors.New("order cannot contain more than 50 items")
	}

	// Validate each item
	for i, item := range items {
		if err := ValidateOrderItem(item); err != nil {
			return fmt.Errorf("item[%d]: %w", i, err)
		}
	}

	return nil
}

// OrderItem represents an order item (define in your model)
type OrderItem struct {
	ProductID string
	Quantity  int32
	Price     float64
}

// ValidateOrderItem validates a single order item
func ValidateOrderItem(item OrderItem) error {
	if err := ValidateRequired(item.ProductID, "product_id"); err != nil {
		return err
	}

	if item.Quantity <= 0 {
		return errors.New("quantity must be positive")
	}

	if item.Quantity > 1000 {
		return errors.New("quantity cannot exceed 1000 per item")
	}

	if item.Price < 0 {
		return errors.New("price cannot be negative")
	}

	if item.Price > 1000000 {
		return errors.New("price cannot exceed 1,000,000")
	}

	return nil
}

// ValidateUpdateOrderStatusRequest validates order status update
func ValidateUpdateOrderStatusRequest(orderID, status string) error {
	if err := ValidateRequired(orderID, "order_id"); err != nil {
		return err
	}

	allowedStatuses := []string{"pending", "confirmed", "processing", "shipped", "delivered", "cancelled"}
	return ValidateEnum(status, allowedStatuses, "status")
}

// --- Product Validation ---

// ValidateCreateProductRequest validates product creation request
func ValidateCreateProductRequest(name, description, categoryID string, price float64, stock int32) error {
	// Validate name
	if err := ValidateRequired(name, "name"); err != nil {
		return err
	}
	if err := ValidateLength(name, 3, 200, "name"); err != nil {
		return err
	}

	// Validate description
	if description != "" {
		if err := ValidateLength(description, 10, 5000, "description"); err != nil {
			return err
		}
	}

	// Validate category ID
	if err := ValidateRequired(categoryID, "category_id"); err != nil {
		return err
	}

	// Validate price
	if price <= 0 {
		return errors.New("price must be positive")
	}
	if price > 10000000 {
		return errors.New("price cannot exceed 10,000,000")
	}

	// Validate stock
	if stock < 0 {
		return errors.New("stock cannot be negative")
	}
	if stock > 1000000 {
		return errors.New("stock cannot exceed 1,000,000")
	}

	return nil
}

// ValidateUpdateProductRequest validates product update request
func ValidateUpdateProductRequest(productID, name, description string, price float64, stock int32) error {
	if err := ValidateRequired(productID, "product_id"); err != nil {
		return err
	}

	// If updating name, validate it
	if name != "" {
		if err := ValidateLength(name, 3, 200, "name"); err != nil {
			return err
		}
	}

	// If updating description, validate it
	if description != "" {
		if err := ValidateLength(description, 10, 5000, "description"); err != nil {
			return err
		}
	}

	// If updating price, validate it
	if price > 0 {
		if price > 10000000 {
			return errors.New("price cannot exceed 10,000,000")
		}
	}

	// If updating stock, validate it
	if stock < 0 {
		return errors.New("stock cannot be negative")
	}

	return nil
}

// ValidateProductSearch validates search parameters
func ValidateProductSearch(query string, minPrice, maxPrice float64, page, pageSize int) error {
	// Validate search query
	if query != "" {
		if err := ValidateLength(query, 2, 100, "search_query"); err != nil {
			return err
		}
	}

	// Validate price range
	if minPrice < 0 {
		return errors.New("min_price cannot be negative")
	}
	if maxPrice < 0 {
		return errors.New("max_price cannot be negative")
	}
	if maxPrice > 0 && minPrice > maxPrice {
		return errors.New("min_price cannot be greater than max_price")
	}

	// Validate pagination
	return ValidatePaginationParams(page, pageSize)
}

// --- User Validation ---

// ValidateCreateUserRequest validates user registration request
func ValidateCreateUserRequest(email, password, fullName, phone string) error {
	// Validate email
	if err := ValidateEmail(email); err != nil {
		return err
	}

	// Validate password
	if err := ValidatePassword(password); err != nil {
		return err
	}

	// Validate full name
	if err := ValidateRequired(fullName, "full_name"); err != nil {
		return err
	}
	if err := ValidateLength(fullName, 2, 100, "full_name"); err != nil {
		return err
	}

	// Validate phone (optional)
	if phone != "" {
		if err := ValidatePhone(phone); err != nil {
			return err
		}
	}

	return nil
}

// ValidateLoginRequest validates login credentials
func ValidateLoginRequest(email, password string) error {
	if err := ValidateEmail(email); err != nil {
		return err
	}

	if err := ValidateRequired(password, "password"); err != nil {
		return err
	}

	return nil
}

// ValidateUpdateUserRequest validates user profile update
func ValidateUpdateUserRequest(userID, fullName, phone string) error {
	if err := ValidateRequired(userID, "user_id"); err != nil {
		return err
	}

	if fullName != "" {
		if err := ValidateLength(fullName, 2, 100, "full_name"); err != nil {
			return err
		}
	}

	if phone != "" {
		if err := ValidatePhone(phone); err != nil {
			return err
		}
	}

	return nil
}

// ValidateChangePasswordRequest validates password change request
func ValidateChangePasswordRequest(oldPassword, newPassword string) error {
	if err := ValidateRequired(oldPassword, "old_password"); err != nil {
		return err
	}

	if err := ValidatePassword(newPassword); err != nil {
		return fmt.Errorf("new_password: %w", err)
	}

	if oldPassword == newPassword {
		return errors.New("new password must be different from old password")
	}

	return nil
}

// --- Payment Validation ---

// ValidateCreatePaymentRequest validates payment creation
func ValidateCreatePaymentRequest(orderID, paymentMethod string, amount float64) error {
	if err := ValidateRequired(orderID, "order_id"); err != nil {
		return err
	}

	allowedMethods := []string{"credit_card", "debit_card", "paypal", "stripe", "bank_transfer"}
	if err := ValidateEnum(paymentMethod, allowedMethods, "payment_method"); err != nil {
		return err
	}

	if amount <= 0 {
		return errors.New("amount must be positive")
	}

	if amount > 100000000 {
		return errors.New("amount cannot exceed 100,000,000")
	}

	return nil
}

// ValidateStripePaymentRequest validates Stripe payment details
func ValidateStripePaymentRequest(token, email string, amount float64) error {
	if err := ValidateRequired(token, "stripe_token"); err != nil {
		return err
	}

	if err := ValidateEmail(email); err != nil {
		return err
	}

	if amount <= 0 {
		return errors.New("amount must be positive")
	}

	return nil
}

// --- Inventory Validation ---

// ValidateUpdateStockRequest validates inventory stock update
func ValidateUpdateStockRequest(productID string, quantity int32) error {
	if err := ValidateRequired(productID, "product_id"); err != nil {
		return err
	}

	if quantity < -1000000 || quantity > 1000000 {
		return errors.New("quantity change must be between -1,000,000 and 1,000,000")
	}

	return nil
}

// ValidateReserveStockRequest validates stock reservation
func ValidateReserveStockRequest(productID, orderID string, quantity int32) error {
	if err := ValidateRequired(productID, "product_id"); err != nil {
		return err
	}

	if err := ValidateRequired(orderID, "order_id"); err != nil {
		return err
	}

	if quantity <= 0 {
		return errors.New("quantity must be positive")
	}

	if quantity > 1000 {
		return errors.New("cannot reserve more than 1000 units at once")
	}

	return nil
}

// --- Notification Validation ---

// ValidateEmailNotificationRequest validates email notification request
func ValidateEmailNotificationRequest(toEmail, subject, body string) error {
	if err := ValidateEmail(toEmail); err != nil {
		return fmt.Errorf("to_email: %w", err)
	}

	if err := ValidateRequired(subject, "subject"); err != nil {
		return err
	}

	if err := ValidateLength(subject, 3, 200, "subject"); err != nil {
		return err
	}

	if err := ValidateRequired(body, "body"); err != nil {
		return err
	}

	if err := ValidateLength(body, 10, 10000, "body"); err != nil {
		return err
	}

	return nil
}

// ValidateSMSNotificationRequest validates SMS notification request
func ValidateSMSNotificationRequest(phone, message string) error {
	if err := ValidatePhone(phone); err != nil {
		return err
	}

	if err := ValidateRequired(message, "message"); err != nil {
		return err
	}

	if err := ValidateLength(message, 10, 160, "message"); err != nil {
		return err
	}

	return nil
}
