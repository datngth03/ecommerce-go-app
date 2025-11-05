# Simple E2E Test
$BASE_URL = "http://localhost:8000/api/v1"
$TOKEN = ""
$PRODUCT_ID = ""

Write-Host "=== E-Commerce Simple Test ===" -ForegroundColor Cyan

# Step 1: Login
Write-Host "`n1. Testing Login..." -ForegroundColor Yellow
$loginBody = @{
    email    = "testfix@example.com"
    password = "TestFix123!"
} | ConvertTo-Json

$loginResult = Invoke-RestMethod -Uri "$BASE_URL/auth/login" -Method Post -Body $loginBody -ContentType "application/json"
$TOKEN = $loginResult.data.access_token
Write-Host "   Login successful. Token: $($TOKEN.Substring(0,20))..." -ForegroundColor Green

# Step 2: Create Product
Write-Host "`n2. Testing Create Product..." -ForegroundColor Yellow
$headers = @{
    Authorization  = "Bearer $TOKEN"
    "Content-Type" = "application/json"
}

$productBody = @{
    name        = "Test Product $(Get-Date -Format 'HHmmss')"
    description = "Automated test product"
    price       = 99.99
    category_id = "63b957bf-0f16-4f32-8c34-8215ccc5bc46"
} | ConvertTo-Json

$productResult = Invoke-RestMethod -Uri "$BASE_URL/products" -Method Post -Body $productBody -Headers $headers
$PRODUCT_ID = $productResult.data.id
Write-Host "   Product created. ID: $PRODUCT_ID" -ForegroundColor Green
Write-Host "   Name: $($productResult.data.name)" -ForegroundColor Gray

# Step 3: Check Inventory
Write-Host "`n3. Testing Check Inventory..." -ForegroundColor Yellow
$inventoryResult = Invoke-RestMethod -Uri "$BASE_URL/inventory/$PRODUCT_ID" -Method Get -Headers $headers
Write-Host "   Inventory checked" -ForegroundColor Green
Write-Host "   Available: $($inventoryResult.data.available)" -ForegroundColor Gray
Write-Host "   Reserved: $($inventoryResult.data.reserved)" -ForegroundColor Gray

# Step 4: Check Availability
Write-Host "`n4. Testing Check Availability..." -ForegroundColor Yellow
$availabilityBody = @{
    items = @(
        @{
            product_id = $PRODUCT_ID
            quantity   = 2
        }
    )
} | ConvertTo-Json -Depth 10

try {
    $availabilityResult = Invoke-RestMethod -Uri "$BASE_URL/inventory/check-availability" -Method Post -Body $availabilityBody -Headers $headers
    Write-Host "   Availability checked" -ForegroundColor Green
    Write-Host "   Available: $($availabilityResult.available)" -ForegroundColor Gray
    if (-not $availabilityResult.available) {
        Write-Host "   Note: Product has no stock" -ForegroundColor Yellow
    }
}
catch {
    Write-Host "   ❌ Check availability failed: $($_.Exception.Message)" -ForegroundColor Red
    Write-Host "   Request Body: $availabilityBody" -ForegroundColor Gray
}

# Step 5: Add to Cart
Write-Host "`n5. Testing Add to Cart..." -ForegroundColor Yellow
$cartBody = @{
    product_id = $PRODUCT_ID
    quantity   = 2
} | ConvertTo-Json

try {
    $cartResult = Invoke-RestMethod -Uri "$BASE_URL/cart" -Method Post -Body $cartBody -Headers $headers
    Write-Host "   Added to cart" -ForegroundColor Green
}
catch {
    Write-Host "   ❌ Add to cart failed: $($_.Exception.Message)" -ForegroundColor Red
}

# Step 6: Create Order
Write-Host "`n6. Testing Create Order..." -ForegroundColor Yellow
$orderBody = @{
    shipping_address = "123 Test Street, San Francisco, CA 94105"
    payment_method   = "stripe"
} | ConvertTo-Json

try {
    $orderResult = Invoke-RestMethod -Uri "$BASE_URL/orders" -Method Post -Body $orderBody -Headers $headers
    $ORDER_ID = $orderResult.data.id
    Write-Host "   Order created. ID: $ORDER_ID" -ForegroundColor Green
    Write-Host "   Total: $($orderResult.data.total_amount)" -ForegroundColor Gray
    
    # Step 7: Process Payment
    Write-Host "`n7. Testing Process Payment..." -ForegroundColor Yellow
    $paymentBody = @{
        order_id = $ORDER_ID.ToString()
        amount   = 199.98
        method   = "stripe"
        currency = "USD"
    } | ConvertTo-Json
    
    $paymentResult = Invoke-RestMethod -Uri "$BASE_URL/payments" -Method Post -Body $paymentBody -Headers $headers
    Write-Host "   Payment processed. ID: $($paymentResult.data.id)" -ForegroundColor Green
    Write-Host "   Status: $($paymentResult.data.status)" -ForegroundColor Gray
}
catch {
    Write-Host "   ❌ Order/Payment failed: $($_.Exception.Message)" -ForegroundColor Red
}

Write-Host "`n=== Test Complete ===" -ForegroundColor Cyan
