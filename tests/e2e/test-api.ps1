# E-Commerce API Test Suite
# PowerShell script for automated testing of all microservices via API Gateway

# Configuration
$BASE_URL = "http://localhost:8000/api/v1"
$TEST_EMAIL = "test_$(Get-Date -Format 'yyyyMMddHHmmss')@example.com"
$TEST_PASSWORD = "SecurePass123!"
$TOKEN = ""
$USER_ID = 0
$PRODUCT_ID = 0
$ORDER_ID = 0
$PAYMENT_ID = ""

# Colors for output
$GREEN = "Green"
$RED = "Red"
$YELLOW = "Yellow"
$CYAN = "Cyan"

# Helper function to make HTTP requests
function Invoke-APITest {
    param(
        [string]$Method,
        [string]$Endpoint,
        [object]$Body = $null,
        [hashtable]$Headers = @{}
    )
    
    $url = "$BASE_URL$Endpoint"
    
    if ($TOKEN) {
        $Headers["Authorization"] = "Bearer $TOKEN"
    }
    
    $Headers["Content-Type"] = "application/json"
    
    try {
        $params = @{
            Uri     = $url
            Method  = $Method
            Headers = $Headers
        }
        
        if ($Body) {
            $params["Body"] = ($Body | ConvertTo-Json -Depth 10)
        }
        
        $response = Invoke-RestMethod @params
        return @{
            Success = $true
            Data    = $response
        }
    }
    catch {
        return @{
            Success  = $false
            Error    = $_.Exception.Message
            Response = $_.Exception.Response
        }
    }
}

# Test result tracking
$script:TestResults = @{
    Passed = 0
    Failed = 0
    Tests  = @()
}

function Test-API {
    param(
        [string]$TestName,
        [scriptblock]$TestBlock
    )
    
    Write-Host "`n▶️  Testing: $TestName" -ForegroundColor $CYAN
    
    try {
        & $TestBlock
        $script:TestResults.Passed++
        $script:TestResults.Tests += @{
            Name   = $TestName
            Status = "PASSED"
        }
        Write-Host "✅ PASSED: $TestName" -ForegroundColor $GREEN
    }
    catch {
        $script:TestResults.Failed++
        $script:TestResults.Tests += @{
            Name   = $TestName
            Status = "FAILED"
            Error  = $_.Exception.Message
        }
        Write-Host "❌ FAILED: $TestName" -ForegroundColor $RED
        Write-Host "   Error: $($_.Exception.Message)" -ForegroundColor $RED
    }
}

# Check if API Gateway is running
Write-Host "🔍 Checking API Gateway health..." -ForegroundColor $YELLOW
try {
    $healthCheck = Invoke-RestMethod -Uri "http://localhost:8000/health" -Method Get
    if ($healthCheck.status -eq "healthy") {
        Write-Host "✅ API Gateway is healthy" -ForegroundColor $GREEN
    }
}
catch {
    Write-Host "❌ API Gateway is not responding. Please start the services first." -ForegroundColor $RED
    Write-Host "   Run: docker-compose up -d" -ForegroundColor $YELLOW
    exit 1
}

Write-Host "`n========================================" -ForegroundColor $CYAN
Write-Host "🚀 Starting E-Commerce API Test Suite" -ForegroundColor $CYAN
Write-Host "========================================`n" -ForegroundColor $CYAN

# ==========================================
# Test 1: User Service
# ==========================================
Write-Host "`n📋 Test Suite 1: User Service" -ForegroundColor $YELLOW

Test-API "Register New User" {
    $body = @{
        email     = $script:TEST_EMAIL
        password  = $script:TEST_PASSWORD
        username  = "testuser_$(Get-Date -Format 'HHmmss')"
        full_name = "Test User"
    }
    
    $result = Invoke-APITest -Method POST -Endpoint "/auth/register" -Body $body
    
    if (-not $result.Success) {
        throw "Registration failed: $($result.Error)"
    }
    
    if (-not $result.Data.user) {
        throw "No user data in response"
    }
    
    $script:USER_ID = $result.Data.user.id
    Write-Host "   User ID: $($script:USER_ID)" -ForegroundColor Gray
}

Test-API "Login User" {
    $body = @{
        email    = $script:TEST_EMAIL
        password = $script:TEST_PASSWORD
    }
    
    $result = Invoke-APITest -Method POST -Endpoint "/auth/login" -Body $body
    
    if (-not $result.Success) {
        throw "Login failed: $($result.Error)"
    }
    
    if (-not $result.Data.access_token) {
        throw "No access token in response"
    }
    
    $script:TOKEN = $result.Data.access_token
    Write-Host "   Token obtained: $($script:TOKEN.Substring(0, 20))..." -ForegroundColor Gray
}

Test-API "Get User Profile" {
    $result = Invoke-APITest -Method GET -Endpoint "/users/me"
    
    if (-not $result.Success) {
        throw "Get profile failed: $($result.Error)"
    }
    
    if (-not $result.Data.data.email) {
        throw "No profile data in response"
    }
    
    Write-Host "   Email: $($result.Data.data.email)" -ForegroundColor Gray
}

# ==========================================
# Test 2: Product Service
# ==========================================
Write-Host "`n📋 Test Suite 2: Product Service" -ForegroundColor $YELLOW

Test-API "Create Product" {
    $body = @{
        name        = "Test Product $(Get-Date -Format 'HHmmss')"
        description = "A test product for automated testing"
        price       = 99.99
        sku         = "TEST-$(Get-Date -Format 'yyyyMMddHHmmss')"
        category_id = 1
        stock       = 100
    }
    
    $result = Invoke-APITest -Method POST -Endpoint "/products" -Body $body
    
    if (-not $result.Success) {
        throw "Create product failed: $($result.Error)"
    }
    
    if (-not $result.Data.product) {
        throw "No product data in response"
    }
    
    $script:PRODUCT_ID = $result.Data.product.id
    Write-Host "   Product ID: $($script:PRODUCT_ID)" -ForegroundColor Gray
    Write-Host "   Product Name: $($result.Data.product.name)" -ForegroundColor Gray
}

Test-API "List Products" {
    $result = Invoke-APITest -Method GET -Endpoint "/products?page=1&page_size=10"
    
    if (-not $result.Success) {
        throw "List products failed: $($result.Error)"
    }
    
    if (-not $result.Data.data) {
        throw "No products data in response"
    }
    
    Write-Host "   Total products: $($result.Data.total)" -ForegroundColor Gray
}

Test-API "Get Product Details" {
    $result = Invoke-APITest -Method GET -Endpoint "/products/$($script:PRODUCT_ID)"
    
    if (-not $result.Success) {
        throw "Get product failed: $($result.Error)"
    }
    
    if (-not $result.Data.data) {
        throw "No product data in response"
    }
    
    Write-Host "   Product: $($result.Data.data.name)" -ForegroundColor Gray
    Write-Host "   Price: $($result.Data.data.price)" -ForegroundColor Gray
}

# ==========================================
# Test 3: Inventory Service
# ==========================================
Write-Host "`n📋 Test Suite 3: Inventory Service" -ForegroundColor $YELLOW

Test-API "Check Product Stock" {
    $result = Invoke-APITest -Method GET -Endpoint "/inventory/$($script:PRODUCT_ID)"
    
    if (-not $result.Success) {
        throw "Check stock failed: $($result.Error)"
    }
    
    if (-not $result.Data.data) {
        throw "No stock data in response"
    }
    
    Write-Host "   Available: $($result.Data.data.available)" -ForegroundColor Gray
    Write-Host "   Reserved: $($result.Data.data.reserved)" -ForegroundColor Gray
}

Test-API "Check Availability" {
    $body = @{
        items = @(
            @{
                product_id = $script:PRODUCT_ID
                quantity   = 2
            }
        )
    }
    
    $result = Invoke-APITest -Method POST -Endpoint "/inventory/check-availability" -Body $body
    
    if (-not $result.Success) {
        throw "Check availability failed: $($result.Error)"
    }
    
    if ($null -eq $result.Data.data.available) {
        throw "No availability data in response"
    }
    
    Write-Host "   Available: $($result.Data.data.available)" -ForegroundColor Gray
}

# ==========================================
# Test 4: Order Service
# ==========================================
Write-Host "`n📋 Test Suite 4: Order Service" -ForegroundColor $YELLOW

Test-API "Add Item to Cart" {
    $body = @{
        product_id = $script:PRODUCT_ID
        quantity   = 2
        price      = 99.99
    }
    
    $result = Invoke-APITest -Method POST -Endpoint "/cart" -Body $body
    
    if (-not $result.Success) {
        throw "Add to cart failed: $($result.Error)"
    }
    
    Write-Host "   Added $($body.quantity) items to cart" -ForegroundColor Gray
}

Test-API "Get Cart" {
    $result = Invoke-APITest -Method GET -Endpoint "/cart"
    
    if (-not $result.Success) {
        throw "Get cart failed: $($result.Error)"
    }
    
    if (-not $result.Data.data) {
        throw "No cart data in response"
    }
    
    Write-Host "   Cart total: $($result.Data.data.total)" -ForegroundColor Gray
    Write-Host "   Items count: $($result.Data.data.items.Count)" -ForegroundColor Gray
}

Test-API "Create Order" {
    $body = @{
        shipping_address = "123 Test Street, San Francisco, CA 94105"
        payment_method   = "stripe"
    }
    
    $result = Invoke-APITest -Method POST -Endpoint "/orders" -Body $body
    
    if (-not $result.Success) {
        throw "Create order failed: $($result.Error)"
    }
    
    if (-not $result.Data.data) {
        throw "No order data in response"
    }
    
    $script:ORDER_ID = $result.Data.data.id
    Write-Host "   Order ID: $($script:ORDER_ID)" -ForegroundColor Gray
    Write-Host "   Total: $($result.Data.data.total_amount)" -ForegroundColor Gray
}

Test-API "Get Order Details" {
    $result = Invoke-APITest -Method GET -Endpoint "/orders/$($script:ORDER_ID)"
    
    if (-not $result.Success) {
        throw "Get order failed: $($result.Error)"
    }
    
    if (-not $result.Data.data) {
        throw "No order data in response"
    }
    
    Write-Host "   Status: $($result.Data.data.status)" -ForegroundColor Gray
}

Test-API "List Orders" {
    $result = Invoke-APITest -Method GET -Endpoint "/orders?page=1&page_size=10"
    
    if (-not $result.Success) {
        throw "List orders failed: $($result.Error)"
    }
    
    if (-not $result.Data.data) {
        throw "No orders data in response"
    }
    
    Write-Host "   Total orders: $($result.Data.total)" -ForegroundColor Gray
}

# ==========================================
# Test 5: Payment Service
# ==========================================
Write-Host "`n📋 Test Suite 5: Payment Service" -ForegroundColor $YELLOW

Test-API "Process Payment" {
    $body = @{
        order_id = $script:ORDER_ID.ToString()
        amount   = 199.98
        method   = "stripe"
        currency = "USD"
    }
    
    $result = Invoke-APITest -Method POST -Endpoint "/payments" -Body $body
    
    if (-not $result.Success) {
        throw "Process payment failed: $($result.Error)"
    }
    
    if (-not $result.Data.data) {
        throw "No payment data in response"
    }
    
    $script:PAYMENT_ID = $result.Data.data.id
    Write-Host "   Payment ID: $($script:PAYMENT_ID)" -ForegroundColor Gray
    Write-Host "   Status: $($result.Data.data.status)" -ForegroundColor Gray
}

Test-API "Confirm Payment" {
    $body = @{
        payment_intent_id = "pi_test_$(Get-Date -Format 'yyyyMMddHHmmss')"
    }
    
    $result = Invoke-APITest -Method POST -Endpoint "/payments/$($script:PAYMENT_ID)/confirm" -Body $body
    
    if (-not $result.Success) {
        throw "Confirm payment failed: $($result.Error)"
    }
    
    if (-not $result.Data.success) {
        throw "Payment confirmation unsuccessful"
    }
    
    Write-Host "   Payment confirmed" -ForegroundColor Gray
}

Test-API "Get Payment Details" {
    $result = Invoke-APITest -Method GET -Endpoint "/payments/$($script:PAYMENT_ID)"
    
    if (-not $result.Success) {
        throw "Get payment failed: $($result.Error)"
    }
    
    if (-not $result.Data.data) {
        throw "No payment data in response"
    }
    
    Write-Host "   Amount: $($result.Data.data.amount)" -ForegroundColor Gray
    Write-Host "   Status: $($result.Data.data.status)" -ForegroundColor Gray
}

Test-API "Get Payment History" {
    $result = Invoke-APITest -Method GET -Endpoint "/payments?page=1&page_size=10"
    
    if (-not $result.Success) {
        throw "Get payment history failed: $($result.Error)"
    }
    
    if (-not $result.Data.data) {
        throw "No payment history in response"
    }
    
    Write-Host "   Total payments: $($result.Data.total)" -ForegroundColor Gray
}

Test-API "Save Payment Method" {
    $body = @{
        method_type       = "card"
        gateway_method_id = "pm_test_visa4242_$(Get-Date -Format 'HHmmss')"
        is_default        = $true
    }
    
    $result = Invoke-APITest -Method POST -Endpoint "/payment-methods" -Body $body
    
    if (-not $result.Success) {
        throw "Save payment method failed: $($result.Error)"
    }
    
    if (-not $result.Data.success) {
        throw "Payment method save unsuccessful"
    }
    
    Write-Host "   Payment method saved" -ForegroundColor Gray
}

Test-API "Get Payment Methods" {
    $result = Invoke-APITest -Method GET -Endpoint "/payment-methods"
    
    if (-not $result.Success) {
        throw "Get payment methods failed: $($result.Error)"
    }
    
    if (-not $result.Data.data) {
        throw "No payment methods in response"
    }
    
    Write-Host "   Payment methods count: $($result.Data.data.Count)" -ForegroundColor Gray
}

# ==========================================
# Test 6: Verify Inventory Updates
# ==========================================
Write-Host "`n📋 Test Suite 6: Inventory Verification" -ForegroundColor $YELLOW

Test-API "Verify Stock After Order" {
    $result = Invoke-APITest -Method GET -Endpoint "/inventory/$($script:PRODUCT_ID)"
    
    if (-not $result.Success) {
        throw "Check stock failed: $($result.Error)"
    }
    
    if (-not $result.Data.data) {
        throw "No stock data in response"
    }
    
    Write-Host "   Available: $($result.Data.data.available)" -ForegroundColor Gray
    Write-Host "   Reserved: $($result.Data.data.reserved)" -ForegroundColor Gray
    
    # Verify stock was reserved
    if ($result.Data.data.reserved -lt 2) {
        throw "Stock not properly reserved (expected >= 2, got $($result.Data.data.reserved))"
    }
}

# ==========================================
# Test Summary
# ==========================================
Write-Host "`n========================================" -ForegroundColor $CYAN
Write-Host "📊 Test Summary" -ForegroundColor $CYAN
Write-Host "========================================" -ForegroundColor $CYAN

$totalTests = $script:TestResults.Passed + $script:TestResults.Failed
$passRate = if ($totalTests -gt 0) { [math]::Round(($script:TestResults.Passed / $totalTests) * 100, 2) } else { 0 }

Write-Host "`nTotal Tests:  $totalTests" -ForegroundColor White
Write-Host "Passed:       $($script:TestResults.Passed)" -ForegroundColor $GREEN
Write-Host "Failed:       $($script:TestResults.Failed)" -ForegroundColor $RED
Write-Host "Pass Rate:    $passRate%" -ForegroundColor $(if ($passRate -eq 100) { $GREEN } else { $YELLOW })

if ($script:TestResults.Failed -gt 0) {
    Write-Host "`n❌ Failed Tests:" -ForegroundColor $RED
    foreach ($test in $script:TestResults.Tests | Where-Object { $_.Status -eq "FAILED" }) {
        Write-Host "   • $($test.Name)" -ForegroundColor $RED
        if ($test.Error) {
            Write-Host "     Error: $($test.Error)" -ForegroundColor Gray
        }
    }
}

Write-Host "`n========================================" -ForegroundColor $CYAN

if ($script:TestResults.Failed -eq 0) {
    Write-Host "🎉 All tests passed! System is working correctly." -ForegroundColor $GREEN
    exit 0
}
else {
    Write-Host "⚠️  Some tests failed. Please check the errors above." -ForegroundColor $YELLOW
    exit 1
}
