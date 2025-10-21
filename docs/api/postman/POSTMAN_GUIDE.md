# Postman Collection Guide

## E-Commerce Microservices API Collection

### Overview
This Postman collection provides a complete set of API endpoints for testing the E-Commerce microservices platform.

### Setup Instructions

1. **Import Collection**
   - Open Postman
   - Click "Import" button
   - Select `ecommerce.postman_collection.json`

2. **Import Environment**
   - Import `ecommerce-local.postman_environment.json`
   - Select "E-Commerce Local Environment" from the environment dropdown

3. **Start Services**
   ```bash
   docker-compose up -d
   ```

4. **Validate Files** (Optional)
   ```powershell
   # Validate JSON syntax
   Get-Content docs/api/postman/ecommerce.postman_collection.json | ConvertFrom-Json | Out-Null
   Get-Content docs/api/postman/ecommerce-local.postman_environment.json | ConvertFrom-Json | Out-Null
   ```

### Quick Start Test Flow

Execute requests in this order for a complete E2E test:

#### 1. Authentication
- **Register User** → Creates new user account
- **Login** → Returns access token (auto-saved to environment)

#### 2. User Management
- **Get Profile** → Retrieve current user details
- **Update Profile** → Modify user information

#### 3. Product Management
- **Create Product** → Add new product (saves product_id)
- **List Products** → Browse product catalog
- **Get Product** → View product details
- **Update Product** → Modify product information

#### 4. Inventory Management
- **Check Stock** → View product availability
- **Check Availability** → Validate stock for multiple items
- **Update Stock** → Adjust inventory levels (Admin)

#### 5. Shopping Cart
- **Add to Cart** → Add product to cart
- **Get Cart** → View cart contents
- **Update Cart Item** → Change quantity
- **Remove from Cart** → Delete cart item

#### 6. Order Management
- **Create Order** → Place order from cart (saves order_id)
- **Get Order** → View order details
- **List Orders** → View order history
- **Update Order Status** → Change order status (Admin)

#### 7. Payment Processing
- **Process Payment** → Create payment (saves payment_id)
- **Confirm Payment** → Finalize payment
- **Get Payment** → View payment details
- **Payment History** → View all payments
- **Save Payment Method** → Store payment method
- **List Payment Methods** → View saved methods

### Environment Variables

| Variable | Description | Auto-Set |
|----------|-------------|----------|
| `base_url` | API Gateway URL | Manual |
| `access_token` | JWT access token | Auto |
| `user_id` | Current user ID | Auto |
| `product_id` | Last created product | Auto |
| `category_id` | Product category UUID | Manual |
| `order_id` | Last created order | Auto |
| `payment_id` | Last payment ID | Auto |

### Automated Tests

Each request includes automated tests that:
- ✅ Verify response status codes
- ✅ Validate response structure
- ✅ Extract and save important IDs
- ✅ Check data consistency

Example test script (included in requests):
```javascript
pm.test("Status code is 200", function () {
    pm.response.to.have.status(200);
});

pm.test("Response has data", function () {
    var jsonData = pm.response.json();
    pm.expect(jsonData.data).to.exist;
});
```

### Collection Runner

To run all tests sequentially:

1. Click "Run Collection" button
2. Select "E-Commerce Local Environment"
3. Click "Run E-Commerce Microservices API"
4. View test results and coverage

### API Endpoints Summary

#### Authentication Service
- `POST /auth/register` - Register new user
- `POST /auth/login` - User login
- `POST /auth/refresh` - Refresh token

#### User Service
- `GET /users/me` - Get current user
- `PUT /users/me` - Update profile
- `GET /users/:id` - Get user by ID (Admin)

#### Product Service
- `POST /products` - Create product
- `GET /products` - List products (with pagination)
- `GET /products/:id` - Get product details
- `PUT /products/:id` - Update product
- `DELETE /products/:id` - Delete product

#### Inventory Service
- `GET /inventory/:product_id` - Get stock info
- `POST /inventory/check-availability` - Check availability
- `PUT /inventory/:product_id` - Update stock (Admin)

#### Order Service
- `POST /cart` - Add to cart
- `GET /cart` - Get cart
- `PUT /cart/:item_id` - Update cart item
- `DELETE /cart/:item_id` - Remove from cart
- `POST /orders` - Create order
- `GET /orders/:id` - Get order
- `GET /orders` - List orders
- `PUT /orders/:id/status` - Update status (Admin)

#### Payment Service
- `POST /payments` - Process payment
- `POST /payments/:id/confirm` - Confirm payment
- `GET /payments/:id` - Get payment details
- `GET /payments` - Payment history
- `POST /payment-methods` - Save payment method
- `GET /payment-methods` - List payment methods
- `DELETE /payment-methods/:id` - Remove payment method

### Testing Tips

1. **Start Fresh**: Run "Register User" to create a new test account
2. **Authentication**: Login generates a token valid for 24 hours
3. **Product IDs**: Use existing category_id or create categories first
4. **Order Flow**: Cart → Order → Payment (in sequence)
5. **Admin Actions**: Some endpoints require admin role

### Common Issues

**401 Unauthorized**
- Token expired → Run "Login" again
- Missing token → Check environment variable `access_token`

**404 Not Found**
- Invalid ID → Check product_id, order_id variables
- Service not running → Run `docker-compose ps`

**500 Internal Server Error**
- Check service logs: `docker logs ecommerce-<service-name>`
- Verify database connections

### Advanced Features

#### Pre-request Scripts
Automatically set timestamps and generate test data:
```javascript
pm.environment.set("timestamp", Date.now());
pm.environment.set("random_email", `test${Date.now()}@example.com`);
```

#### Collection Variables
Shared across all requests in the collection for consistent testing.

### Support

For issues or questions:
- Check service logs: `docker-compose logs -f <service-name>`
- Review API documentation: `docs/api/swagger.yaml`
- Run E2E tests: `./tests/e2e/test-simple.ps1`

---

**Version**: 2.0.0  
**Last Updated**: October 2025  
**Maintained By**: Development Team
