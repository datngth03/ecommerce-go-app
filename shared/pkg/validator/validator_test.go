package validator

import (
	"testing"
)

func TestValidateEmail(t *testing.T) {
	tests := []struct {
		name    string
		email   string
		wantErr bool
	}{
		{"Valid email", "test@example.com", false},
		{"Valid email with subdomain", "user@mail.example.com", false},
		{"Valid email with plus", "user+tag@example.com", false},
		{"Empty email", "", true},
		{"Missing @", "testexample.com", true},
		{"Missing domain", "test@", true},
		{"Missing local part", "@example.com", true},
		{"Invalid characters", "test@exam ple.com", true},
		{"Too long", string(make([]byte, 255)) + "@example.com", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateEmail(tt.email)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateEmail() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidatePassword(t *testing.T) {
	tests := []struct {
		name     string
		password string
		wantErr  bool
	}{
		{"Valid password", "Password123", false},
		{"Valid with special chars", "Pass@word123!", false},
		{"Too short", "Pass1", true},
		{"No uppercase", "password123", true},
		{"No lowercase", "PASSWORD123", true},
		{"No digit", "PasswordABC", true},
		{"Empty", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidatePassword(tt.password)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidatePassword() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidatePhone(t *testing.T) {
	tests := []struct {
		name    string
		phone   string
		wantErr bool
	}{
		{"Valid international", "+1234567890", false},
		{"Valid with country code", "+84912345678", false},
		{"Valid long number", "+12345678901234", false},
		{"Empty", "", true},
		{"Invalid characters", "+123-456-7890", false}, // Hyphens removed during validation
		{"Too short", "+1", true},
		{"No digits", "abcdefghij", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidatePhone(tt.phone)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidatePhone() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateRequired(t *testing.T) {
	tests := []struct {
		name      string
		value     string
		fieldName string
		wantErr   bool
	}{
		{"Valid value", "test", "field", false},
		{"Empty string", "", "field", true},
		{"Whitespace only", "   ", "field", true},
		{"Tab and newline", "\t\n", "field", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateRequired(tt.value, tt.fieldName)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateRequired() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateLength(t *testing.T) {
	tests := []struct {
		name      string
		value     string
		min       int
		max       int
		fieldName string
		wantErr   bool
	}{
		{"Valid length", "test", 2, 10, "field", false},
		{"At minimum", "ab", 2, 10, "field", false},
		{"At maximum", "1234567890", 2, 10, "field", false},
		{"Too short", "a", 2, 10, "field", true},
		{"Too long", "12345678901", 2, 10, "field", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateLength(tt.value, tt.min, tt.max, tt.fieldName)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateLength() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidatePositiveInt(t *testing.T) {
	tests := []struct {
		name      string
		value     int
		fieldName string
		wantErr   bool
	}{
		{"Positive value", 10, "quantity", false},
		{"Large positive", 999999, "quantity", false},
		{"Zero", 0, "quantity", true},
		{"Negative", -5, "quantity", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidatePositiveInt(tt.value, tt.fieldName)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidatePositiveInt() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateEnum(t *testing.T) {
	allowed := []string{"pending", "confirmed", "shipped"}

	tests := []struct {
		name      string
		value     string
		fieldName string
		wantErr   bool
	}{
		{"Valid value", "pending", "status", false},
		{"Another valid", "shipped", "status", false},
		{"Invalid value", "delivered", "status", true},
		{"Empty value", "", "status", true},
		{"Case sensitive", "Pending", "status", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateEnum(tt.value, allowed, tt.fieldName)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateEnum() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestSanitizeHTML(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"No HTML", "plain text", "plain text"},
		{"Simple tag", "<p>text</p>", "text"},
		{"Multiple tags", "<div><p>hello</p></div>", "hello"},
		{"With attributes", "<a href='test'>link</a>", "link"},
		{"Script tag", "<script>alert('xss')</script>", "alert('xss')"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := SanitizeHTML(tt.input)
			if got != tt.want {
				t.Errorf("SanitizeHTML() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestValidateCreateOrderRequest(t *testing.T) {
	validItems := []OrderItem{
		{ProductID: "prod-1", Quantity: 2, Price: 100.0},
		{ProductID: "prod-2", Quantity: 1, Price: 50.0},
	}

	tests := []struct {
		name    string
		userID  string
		items   []OrderItem
		wantErr bool
	}{
		{"Valid order", "user-123", validItems, false},
		{"Empty user ID", "", validItems, true},
		{"No items", "user-123", []OrderItem{}, true},
		{"Invalid quantity", "user-123", []OrderItem{{ProductID: "prod-1", Quantity: 0, Price: 100}}, true},
		{"Negative price", "user-123", []OrderItem{{ProductID: "prod-1", Quantity: 1, Price: -10}}, true},
		{"Excessive quantity", "user-123", []OrderItem{{ProductID: "prod-1", Quantity: 2000, Price: 100}}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateCreateOrderRequest(tt.userID, tt.items)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateCreateOrderRequest() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateCreateProductRequest(t *testing.T) {
	tests := []struct {
		name        string
		prodName    string
		description string
		categoryID  string
		price       float64
		stock       int32
		wantErr     bool
	}{
		{"Valid product", "Test Product", "A good product", "cat-1", 99.99, 100, false},
		{"Empty name", "", "Description", "cat-1", 99.99, 100, true},
		{"Short name", "AB", "Description", "cat-1", 99.99, 100, true},
		{"Long name", string(make([]byte, 201)), "Description", "cat-1", 99.99, 100, true},
		{"Zero price", "Product", "Description", "cat-1", 0, 100, true},
		{"Negative price", "Product", "Description", "cat-1", -10, 100, true},
		{"Excessive price", "Product", "Description", "cat-1", 99999999, 100, true},
		{"Negative stock", "Product", "Description", "cat-1", 99.99, -1, true},
		{"Empty category", "Product", "Description", "", 99.99, 100, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateCreateProductRequest(tt.prodName, tt.description, tt.categoryID, tt.price, tt.stock)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateCreateProductRequest() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidatePaginationParams(t *testing.T) {
	tests := []struct {
		name     string
		page     int
		pageSize int
		wantErr  bool
	}{
		{"Valid params", 1, 20, false},
		{"Max page size", 1, 100, false},
		{"Page zero", 0, 20, true},
		{"Negative page", -1, 20, true},
		{"Page size zero", 1, 0, true},
		{"Excessive page size", 1, 101, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidatePaginationParams(tt.page, tt.pageSize)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidatePaginationParams() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
