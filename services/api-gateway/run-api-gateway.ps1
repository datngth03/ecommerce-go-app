# Build and Run Script for API Gateway
# Usage: .\run-api-gateway.ps1 [build|run|docker|clean]

param(
    [Parameter(Position = 0)]
    [string]$Action = "run"
)

$ErrorActionPreference = "Stop"
$ServicePath = "d:\WorkSpace\Personal\Go\ecommerce-go-app\services\api-gateway"

function Show-Header {
    Write-Host ""
    Write-Host "================================" -ForegroundColor Cyan
    Write-Host "   API Gateway Manager" -ForegroundColor Cyan
    Write-Host "================================" -ForegroundColor Cyan
    Write-Host ""
}

function Build-Service {
    Write-Host "🔨 Building API Gateway..." -ForegroundColor Yellow
    Set-Location $ServicePath
    
    go build -o api-gateway.exe ./cmd/main.go
    
    if ($LASTEXITCODE -eq 0) {
        $size = (Get-Item api-gateway.exe).Length / 1MB
        Write-Host "✅ Build successful! Binary size: $([math]::Round($size, 2)) MB" -ForegroundColor Green
    }
    else {
        Write-Host "❌ Build failed!" -ForegroundColor Red
        exit 1
    }
}

function Run-Service {
    Write-Host "🚀 Starting API Gateway..." -ForegroundColor Yellow
    Set-Location $ServicePath
    
    if (-not (Test-Path "api-gateway.exe")) {
        Write-Host "⚠️  Binary not found. Building first..." -ForegroundColor Yellow
        Build-Service
    }
    
    Write-Host ""
    Write-Host "📡 API Gateway will start on:" -ForegroundColor Cyan
    Write-Host "   - HTTP: http://localhost:8000" -ForegroundColor White
    Write-Host "   - Health: http://localhost:8000/health" -ForegroundColor White
    Write-Host "   - API v1: http://localhost:8000/api/v1" -ForegroundColor White
    Write-Host ""
    Write-Host "Press Ctrl+C to stop" -ForegroundColor Gray
    Write-Host ""
    
    .\api-gateway.exe
}

function Build-Docker {
    Write-Host "🐳 Building Docker image..." -ForegroundColor Yellow
    Set-Location $ServicePath
    
    docker build -t ecommerce-api-gateway:latest .
    
    if ($LASTEXITCODE -eq 0) {
        Write-Host "✅ Docker image built successfully!" -ForegroundColor Green
        Write-Host ""
        Write-Host "Run with: docker run -p 8000:8000 ecommerce-api-gateway:latest" -ForegroundColor Cyan
    }
    else {
        Write-Host "❌ Docker build failed!" -ForegroundColor Red
        exit 1
    }
}

function Clean-Build {
    Write-Host "🧹 Cleaning build artifacts..." -ForegroundColor Yellow
    Set-Location $ServicePath
    
    if (Test-Path "api-gateway.exe") {
        Remove-Item "api-gateway.exe" -Force
        Write-Host "✅ Removed api-gateway.exe" -ForegroundColor Green
    }
    
    if (Test-Path "api-gateway") {
        Remove-Item "api-gateway" -Force
        Write-Host "✅ Removed api-gateway (Linux binary)" -ForegroundColor Green
    }
    
    Write-Host "✅ Clean completed!" -ForegroundColor Green
}

Show-Header

switch ($Action.ToLower()) {
    "build" {
        Build-Service
    }
    "run" {
        Run-Service
    }
    "docker" {
        Build-Docker
    }
    "clean" {
        Clean-Build
    }
    default {
        Write-Host "Usage: .\run-api-gateway.ps1 [build|run|docker|clean]" -ForegroundColor Yellow
        Write-Host ""
        Write-Host "Commands:" -ForegroundColor Cyan
        Write-Host "  build   - Build the API Gateway binary" -ForegroundColor White
        Write-Host "  run     - Build and run the API Gateway" -ForegroundColor White
        Write-Host "  docker  - Build Docker image" -ForegroundColor White
        Write-Host "  clean   - Remove build artifacts" -ForegroundColor White
        exit 1
    }
}
