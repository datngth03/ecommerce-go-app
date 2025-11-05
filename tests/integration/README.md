# Integration Tests

## Overview
Integration tests for the E-Commerce microservices platform. These tests verify end-to-end functionality across all services through the API Gateway.

## Prerequisites

1. **Running Services**
   ```bash
   docker-compose up -d
   ```

2. **Go Dependencies**
   ```bash
   go mod download
   go get github.com/stretchr/testify/assert
   go get github.com/stretchr/testify/require
   ```

## Running Tests

### Run All Integration Tests
```bash
cd tests/integration
go test -v
```

### Run Specific Test
```bash
go test -v -run TestUserRegistrationAndLogin
```

### Run with Coverage
```bash
go test -v -cover -coverprofile=coverage.out
go tool cover -html=coverage.out
```

### Skip Integration Tests (for CI/CD)
```bash
go test -v -short
```

## Test Suites

### 1. User Registration and Login
**File**: `ecommerce_test.go`  
**Function**: `TestUserRegistrationAndLogin`

Tests:
- User registration with unique email
- User login with credentials
- Get user profile with JWT token
- Token-based authentication

### 2. Product Creation and Retrieval
**Function**: `TestProductCreationAndRetrieval`

Tests:
- Create product with valid data
- Retrieve product by ID
- List products with pagination
- Product data consistency

### 3. Complete Order Flow
**Function**: `TestCompleteOrderFlow`

Tests:
- Add product to cart
- View cart contents
- Create order from cart
- Process payment
- Verify order status
- End-to-end order lifecycle

### 4. Inventory Management
**Function**: `TestInventoryManagement`

Tests:
- Check product stock levels
- Verify product availability
- Multi-item availability check
- Stock reservation tracking

## Test Structure

```go
func TestFeatureName(t *testing.T) {
    if testing.Short() {
        t.Skip("Skipping integration test")
    }
    
    // Setup
    token := getTestToken(t)
    
    // Test cases
    t.Run("Specific Test Case", func(t *testing.T) {
        // Test implementation
    })
}
```

## Helper Functions

### makeRequest
```go
makeRequest(method, endpoint string, payload interface{}, token string) (*http.Response, error)
```
Makes HTTP requests to the API Gateway with automatic JSON handling.

### getTestToken
```go
getTestToken(t *testing.T) string
```
Registers a test user and returns a valid JWT token.

### createTestProduct
```go
createTestProduct(t *testing.T, token string) string
```
Creates a test product and returns the product ID.

## Test Data

### Default Category ID
```
63b957bf-0f16-4f32-8c34-8215ccc5bc46
```
(Electronics category - ensure this exists in your database)

### Test User Format
```
email: test_{timestamp}@example.com
password: Test123!
```

## Configuration

### Base URL
```go
const baseURL = "http://localhost:8000/api/v1"
```

### Timeout
```go
client := &http.Client{Timeout: 10 * time.Second}
```

## Assertions

Uses `testify` package for assertions:

```go
assert.Equal(t, expected, actual)
assert.NotEmpty(t, value)
assert.True(t, condition)
require.NoError(t, err) // Fails immediately if error
```

## Common Issues

### Services Not Running
```bash
Error: connection refused
Solution: docker-compose up -d
```

### Invalid Category ID
```bash
Error: category not found
Solution: Update category_id in createTestProduct()
```

### Token Expired
```bash
Error: 401 Unauthorized
Solution: Tests automatically generate fresh tokens
```

## Best Practices

1. **Isolation**: Each test should be independent
2. **Cleanup**: Tests create unique data (no cleanup needed)
3. **Timeouts**: Set reasonable HTTP timeouts (10s)
4. **Error Handling**: Use `require.NoError` for critical errors
5. **Parallel Tests**: Avoid parallel execution for integration tests

## CI/CD Integration

### GitHub Actions Example
```yaml
name: Integration Tests

on: [push, pull_request]

jobs:
  test:
    runs-on: ubuntu-latest
    
    services:
      postgres:
        image: postgres:15
        env:
          POSTGRES_PASSWORD: postgres
        ports:
          - 5432:5432
          
    steps:
      - uses: actions/checkout@v3
      
      - name: Set up Go
        uses: actions/setup-go@v4
        with:
          go-version: '1.24'
          
      - name: Start Services
        run: docker-compose up -d
        
      - name: Wait for Services
        run: sleep 30
        
      - name: Run Integration Tests
        run: |
          cd tests/integration
          go test -v -timeout 5m
```

## Extending Tests

### Adding New Test Suite

1. Create test function:
```go
func TestNewFeature(t *testing.T) {
    if testing.Short() {
        t.Skip("Skipping integration test")
    }
    
    token := getTestToken(t)
    
    t.Run("Test Case Name", func(t *testing.T) {
        // Test implementation
    })
}
```

2. Add helper functions if needed
3. Document in this README
4. Run and verify

### Adding New Helper Function

```go
func createTestOrder(t *testing.T, token string, productID string) string {
    // Add to cart
    payload := map[string]interface{}{
        "product_id": productID,
        "quantity":   1,
    }
    makeRequest("POST", "/cart", payload, token)
    
    // Create order
    orderPayload := map[string]interface{}{
        "shipping_address": "123 Test St",
        "payment_method":   "stripe",
    }
    resp, _ := makeRequest("POST", "/orders", orderPayload, token)
    
    // Extract order ID
    var result map[string]interface{}
    json.NewDecoder(resp.Body).Decode(&result)
    return result["data"].(map[string]interface{})["id"].(string)
}
```

## Performance Testing

For load testing, use tools like:
- Apache JMeter
- k6
- Locust

Integration tests focus on correctness, not performance.

## Troubleshooting

### Debug Mode
Add verbose logging:
```go
resp, err := makeRequest("POST", "/orders", payload, token)
if err != nil {
    t.Logf("Request failed: %v", err)
}
body, _ := ioutil.ReadAll(resp.Body)
t.Logf("Response: %s", string(body))
```

### Check Service Logs
```bash
docker logs ecommerce-api-gateway
docker logs ecommerce-user-service
docker logs ecommerce-product-service
```

### Database State
```bash
docker exec -it ecommerce-postgres psql -U postgres -d users_db -c "SELECT * FROM users;"
```

## Test Coverage Goals

- **Critical Paths**: 100% (Auth, Orders, Payments)
- **User Flows**: 90%
- **Edge Cases**: 70%
- **Overall**: 85%+

---

**Note**: These are integration tests that require running services. They are slower than unit tests but provide higher confidence in system functionality.
