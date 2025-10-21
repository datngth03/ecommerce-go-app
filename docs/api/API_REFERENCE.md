# E-Commerce Microservices API Documentation

## Overview
Complete RESTful API documentation for the E-Commerce microservices platform accessible through the API Gateway.

**Base URL**: `http://localhost:8000/api/v1`  
**Authentication**: Bearer Token (JWT)  
**Content-Type**: `application/json`

---

## Table of Contents
1. [Authentication](#authentication)
2. [User Management](#user-management)
3. [Product Service](#product-service)
4. [Inventory Service](#inventory-service)
5. [Order Service](#order-service)
6. [Payment Service](#payment-service)
7. [Error Responses](#error-responses)
8. [Rate Limiting](#rate-limiting)

---

## Authentication

### Register User
Creates a new user account.

**Endpoint**: `POST /auth/register`  
**Auth Required**: No

**Request Body**:
```json
{
  "email": "user@example.com",
  "password": "SecurePass123!",
  "name": "John Doe",
  "phone": "+1234567890"
}
```

**Response** (201 Created):
```json
{
  "data": {
    "success": true,
    "message": "User created successfully"
  }
}
```

---

### Login
Authenticates user and returns JWT token.

**Endpoint**: `POST /auth/login`  
**Auth Required**: No

**Request Body**:
```json
{
  "email": "user@example.com",
  "password": "SecurePass123!"
}
```

**Response** (200 OK):
```json
{
  "data": {
    "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "expires_at": "2025-10-22T12:00:00Z"
  }
}
```

---

### Refresh Token
Generates new access token using refresh token.

**Endpoint**: `POST /auth/refresh`  
**Auth Required**: Yes (Refresh Token)

**Request Body**:
```json
{
  "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
}
```

**Response** (200 OK):
```json
{
  "data": {
    "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "expires_at": "2025-10-22T12:00:00Z"
  }
}
```

---

## User Management

### Get Current User Profile
Retrieves authenticated user's profile.

**Endpoint**: `GET /users/me`  
**Auth Required**: Yes

**Response** (200 OK):
```json
{
  "data": {
    "id": 123,
    "email": "user@example.com",
    "name": "John Doe",
    "phone": "+1234567890",
    "is_active": true,
    "created_at": "2025-10-21T10:00:00Z",
    "updated_at": "2025-10-21T10:00:00Z"
  }
}
```

---

### Update User Profile
Updates authenticated user's profile information.

**Endpoint**: `PUT /users/me`  
**Auth Required**: Yes

**Request Body**:
```json
{
  "name": "John Smith",
  "phone": "+1234567899"
}
```

**Response** (200 OK):
```json
{
  "message": "Profile updated successfully",
  "data": {
    "id": 123,
    "email": "user@example.com",
    "name": "John Smith",
    "phone": "+1234567899"
  }
}
```

---

## Product Service

### Create Product
Creates a new product (Admin only).

**Endpoint**: `POST /products`  
**Auth Required**: Yes

**Request Body**:
```json
{
  "name": "Wireless Headphones",
  "description": "Premium noise-canceling headphones",
  "price": 199.99,
  "category_id": "63b957bf-0f16-4f32-8c34-8215ccc5bc46",
  "image_url": "https://example.com/image.jpg"
}
```

**Response** (201 Created):
```json
{
  "data": {
    "id": "a1b2c3d4-e5f6-7890-abcd-ef1234567890",
    "name": "Wireless Headphones",
    "slug": "wireless-headphones",
    "description": "Premium noise-canceling headphones",
    "price": 199.99,
    "category_id": "63b957bf-0f16-4f32-8c34-8215ccc5bc46",
    "image_url": "https://example.com/image.jpg",
    "is_active": true,
    "created_at": "2025-10-21T10:00:00Z",
    "updated_at": "2025-10-21T10:00:00Z"
  }
}
```

---

### List Products
Retrieves paginated list of products.

**Endpoint**: `GET /products`  
**Auth Required**: No

**Query Parameters**:
- `page` (default: 1) - Page number
- `page_size` (default: 10, max: 100) - Items per page
- `category_id` (optional) - Filter by category
- `search` (optional) - Search by name/description

**Example**: `GET /products?page=1&page_size=20&category_id=63b957bf-0f16-4f32-8c34-8215ccc5bc46`

**Response** (200 OK):
```json
{
  "data": [
    {
      "id": "a1b2c3d4-e5f6-7890-abcd-ef1234567890",
      "name": "Wireless Headphones",
      "price": 199.99,
      "image_url": "https://example.com/image.jpg"
    }
  ],
  "pagination": {
    "page": 1,
    "page_size": 20,
    "total": 150,
    "total_pages": 8
  }
}
```

---

### Get Product Details
Retrieves detailed information about a specific product.

**Endpoint**: `GET /products/:id`  
**Auth Required**: No

**Response** (200 OK):
```json
{
  "data": {
    "id": "a1b2c3d4-e5f6-7890-abcd-ef1234567890",
    "name": "Wireless Headphones",
    "slug": "wireless-headphones",
    "description": "Premium noise-canceling headphones",
    "price": 199.99,
    "category_id": "63b957bf-0f16-4f32-8c34-8215ccc5bc46",
    "image_url": "https://example.com/image.jpg",
    "is_active": true,
    "created_at": "2025-10-21T10:00:00Z",
    "updated_at": "2025-10-21T10:00:00Z"
  }
}
```

---

### Update Product
Updates product information (Admin only).

**Endpoint**: `PUT /products/:id`  
**Auth Required**: Yes (Admin)

**Request Body**:
```json
{
  "name": "Premium Wireless Headphones",
  "price": 249.99,
  "description": "Updated description"
}
```

---

### Delete Product
Soft deletes a product (Admin only).

**Endpoint**: `DELETE /products/:id`  
**Auth Required**: Yes (Admin)

**Response** (200 OK):
```json
{
  "message": "Product deleted successfully"
}
```

---

## Inventory Service

### Get Product Stock
Retrieves inventory information for a product.

**Endpoint**: `GET /inventory/:product_id`  
**Auth Required**: Yes

**Response** (200 OK):
```json
{
  "data": {
    "product_id": "a1b2c3d4-e5f6-7890-abcd-ef1234567890",
    "available": 50,
    "reserved": 5,
    "total": 55,
    "warehouse_id": "default"
  }
}
```

---

### Check Availability
Checks if products are available in requested quantities.

**Endpoint**: `POST /inventory/check-availability`  
**Auth Required**: Yes

**Request Body**:
```json
{
  "items": [
    {
      "product_id": "a1b2c3d4-e5f6-7890-abcd-ef1234567890",
      "quantity": 2
    },
    {
      "product_id": "b2c3d4e5-f6g7-8901-bcde-fg2345678901",
      "quantity": 1
    }
  ]
}
```

**Response** (200 OK):
```json
{
  "available": true,
  "message": "availability checked successfully",
  "unavailable_items": []
}
```

**Response when unavailable**:
```json
{
  "available": false,
  "message": "availability checked successfully",
  "unavailable_items": [
    {
      "product_id": "a1b2c3d4-e5f6-7890-abcd-ef1234567890",
      "requested": 10,
      "available": 5
    }
  ]
}
```

---

### Update Stock (Admin)
Updates product stock levels.

**Endpoint**: `PUT /inventory/:product_id`  
**Auth Required**: Yes (Admin)

**Request Body**:
```json
{
  "quantity": 100,
  "operation": "add"
}
```

---

## Order Service

### Add to Cart
Adds a product to the user's shopping cart.

**Endpoint**: `POST /cart`  
**Auth Required**: Yes

**Request Body**:
```json
{
  "product_id": "a1b2c3d4-e5f6-7890-abcd-ef1234567890",
  "quantity": 2
}
```

**Response** (200 OK):
```json
{
  "message": "item added to cart successfully",
  "data": {
    "user_id": 123,
    "items": [
      {
        "product_id": "a1b2c3d4-e5f6-7890-abcd-ef1234567890",
        "quantity": 2,
        "price": 199.99
      }
    ],
    "total": 399.98
  }
}
```

---

### Get Cart
Retrieves the user's current shopping cart.

**Endpoint**: `GET /cart`  
**Auth Required**: Yes

**Response** (200 OK):
```json
{
  "data": {
    "user_id": 123,
    "items": [
      {
        "id": "cart-item-1",
        "product_id": "a1b2c3d4-e5f6-7890-abcd-ef1234567890",
        "product_name": "Wireless Headphones",
        "quantity": 2,
        "price": 199.99,
        "subtotal": 399.98
      }
    ],
    "total": 399.98,
    "item_count": 2
  }
}
```

---

### Update Cart Item
Updates quantity of an item in the cart.

**Endpoint**: `PUT /cart/:item_id`  
**Auth Required**: Yes

**Request Body**:
```json
{
  "quantity": 3
}
```

---

### Remove from Cart
Removes an item from the cart.

**Endpoint**: `DELETE /cart/:item_id`  
**Auth Required**: Yes

**Response** (200 OK):
```json
{
  "message": "item removed from cart successfully"
}
```

---

### Create Order
Creates an order from the user's cart.

**Endpoint**: `POST /orders`  
**Auth Required**: Yes

**Request Body**:
```json
{
  "shipping_address": "123 Main St, San Francisco, CA 94105",
  "payment_method": "stripe"
}
```

**Response** (201 Created):
```json
{
  "data": {
    "id": "order-uuid-1234",
    "user_id": 123,
    "status": "PENDING",
    "total_amount": 399.98,
    "shipping_address": "123 Main St, San Francisco, CA 94105",
    "payment_method": "stripe",
    "items": [
      {
        "product_id": "a1b2c3d4-e5f6-7890-abcd-ef1234567890",
        "quantity": 2,
        "price": 199.99,
        "subtotal": 399.98
      }
    ],
    "created_at": "2025-10-21T10:00:00Z"
  }
}
```

---

### Get Order Details
Retrieves details of a specific order.

**Endpoint**: `GET /orders/:id`  
**Auth Required**: Yes

**Response** (200 OK):
```json
{
  "data": {
    "id": "order-uuid-1234",
    "user_id": 123,
    "status": "CONFIRMED",
    "total_amount": 399.98,
    "shipping_address": "123 Main St, San Francisco, CA 94105",
    "items": [...],
    "created_at": "2025-10-21T10:00:00Z",
    "updated_at": "2025-10-21T10:05:00Z"
  }
}
```

---

### List Orders
Retrieves user's order history.

**Endpoint**: `GET /orders`  
**Auth Required**: Yes

**Query Parameters**:
- `page` (default: 1)
- `page_size` (default: 10)
- `status` (optional) - Filter by status

**Response** (200 OK):
```json
{
  "data": [
    {
      "id": "order-uuid-1234",
      "status": "CONFIRMED",
      "total_amount": 399.98,
      "created_at": "2025-10-21T10:00:00Z"
    }
  ],
  "pagination": {
    "page": 1,
    "page_size": 10,
    "total": 25
  }
}
```

---

## Payment Service

### Process Payment
Initiates payment for an order.

**Endpoint**: `POST /payments`  
**Auth Required**: Yes

**Request Body**:
```json
{
  "order_id": "order-uuid-1234",
  "amount": 399.98,
  "method": "stripe",
  "currency": "USD"
}
```

**Response** (201 Created):
```json
{
  "data": {
    "id": "payment-uuid-5678",
    "order_id": "order-uuid-1234",
    "amount": 399.98,
    "currency": "USD",
    "method": "stripe",
    "status": "COMPLETED",
    "created_at": "2025-10-21T10:00:00Z"
  }
}
```

---

### Confirm Payment
Confirms a payment transaction.

**Endpoint**: `POST /payments/:id/confirm`  
**Auth Required**: Yes

**Request Body**:
```json
{
  "payment_intent_id": "pi_stripe_123456"
}
```

**Response** (200 OK):
```json
{
  "success": true,
  "message": "payment confirmed successfully"
}
```

---

### Get Payment Details
Retrieves payment information.

**Endpoint**: `GET /payments/:id`  
**Auth Required**: Yes

**Response** (200 OK):
```json
{
  "data": {
    "id": "payment-uuid-5678",
    "order_id": "order-uuid-1234",
    "amount": 399.98,
    "currency": "USD",
    "method": "stripe",
    "status": "COMPLETED",
    "gateway_transaction_id": "tx_stripe_789",
    "created_at": "2025-10-21T10:00:00Z"
  }
}
```

---

### Payment History
Retrieves user's payment history.

**Endpoint**: `GET /payments`  
**Auth Required**: Yes

**Query Parameters**:
- `page` (default: 1)
- `page_size` (default: 10)

**Response** (200 OK):
```json
{
  "data": [
    {
      "id": "payment-uuid-5678",
      "order_id": "order-uuid-1234",
      "amount": 399.98,
      "status": "COMPLETED",
      "created_at": "2025-10-21T10:00:00Z"
    }
  ],
  "pagination": {
    "page": 1,
    "page_size": 10,
    "total": 15
  }
}
```

---

### Save Payment Method
Stores a payment method for future use.

**Endpoint**: `POST /payment-methods`  
**Auth Required**: Yes

**Request Body**:
```json
{
  "method_type": "card",
  "gateway_method_id": "pm_stripe_visa4242",
  "is_default": true
}
```

**Response** (201 Created):
```json
{
  "success": true,
  "message": "payment method saved successfully"
}
```

---

### List Payment Methods
Retrieves user's saved payment methods.

**Endpoint**: `GET /payment-methods`  
**Auth Required**: Yes

**Response** (200 OK):
```json
{
  "data": [
    {
      "id": "pm-uuid-9012",
      "method_type": "card",
      "gateway_method_id": "pm_stripe_visa4242",
      "is_default": true,
      "created_at": "2025-10-21T10:00:00Z"
    }
  ]
}
```

---

### Delete Payment Method
Removes a saved payment method.

**Endpoint**: `DELETE /payment-methods/:id`  
**Auth Required**: Yes

**Response** (200 OK):
```json
{
  "message": "payment method deleted successfully"
}
```

---

## Error Responses

All error responses follow this format:

```json
{
  "error": "Error message description"
}
```

### Common HTTP Status Codes

| Code | Meaning | Example |
|------|---------|---------|
| 200 | OK | Request successful |
| 201 | Created | Resource created |
| 400 | Bad Request | Invalid input data |
| 401 | Unauthorized | Missing/invalid token |
| 403 | Forbidden | Insufficient permissions |
| 404 | Not Found | Resource doesn't exist |
| 409 | Conflict | Duplicate resource |
| 500 | Internal Server Error | Server-side error |

---

## Rate Limiting

**Rate Limit**: 1000 requests per hour per IP  
**Headers**:
- `X-RateLimit-Limit`: Maximum requests allowed
- `X-RateLimit-Remaining`: Remaining requests
- `X-RateLimit-Reset`: Time when limit resets (Unix timestamp)

**Response when exceeded** (429 Too Many Requests):
```json
{
  "error": "Rate limit exceeded. Please try again later."
}
```

---

## Authentication Headers

All authenticated endpoints require:

```
Authorization: Bearer <access_token>
Content-Type: application/json
```

---

## Pagination

Paginated endpoints return:

```json
{
  "data": [...],
  "pagination": {
    "page": 1,
    "page_size": 10,
    "total": 150,
    "total_pages": 15
  }
}
```

---

## Timestamps

All timestamps are in ISO 8601 format (UTC):
```
2025-10-21T10:00:00Z
```

---

## Testing

Use the Postman collection in `docs/api/postman/` for interactive API testing.

---

**Version**: 2.0.0  
**Last Updated**: October 2025
