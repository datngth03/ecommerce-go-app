package handler

import (
	"io"
	"log"
	"net/http"
	"strings"

	"github.com/ecommerce/services/api-gateway/internal/proxy"
	"github.com/ecommerce/shared/pkg/response"
)

type ProxyHandler struct {
	serviceProxy *proxy.ServiceProxy
}

func NewProxyHandler(serviceProxy *proxy.ServiceProxy) *ProxyHandler {
	return &ProxyHandler{
		serviceProxy: serviceProxy,
	}
}

// ProxyToUserService proxies requests to user service
func (h *ProxyHandler) ProxyToUserService(w http.ResponseWriter, r *http.Request) {
	serviceName, servicePath := h.serviceProxy.MapRouteToServicePath(r.URL.Path)
	h.proxyRequest(w, r, serviceName, servicePath)
}

// ProxyToUserServiceWithAuth proxies requests to user service with authentication
func (h *ProxyHandler) ProxyToUserServiceWithAuth(w http.ResponseWriter, r *http.Request) {
	// Extract and validate token
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		response.Unauthorized(w, "Authorization header required")
		return
	}

	tokenParts := strings.Split(authHeader, " ")
	if len(tokenParts) != 2 || tokenParts[0] != "Bearer" {
		response.Unauthorized(w, "Invalid authorization header format")
		return
	}

	token := tokenParts[1]

	// Validate token via gRPC (more efficient) or HTTP
	_, err := h.serviceProxy.ValidateTokenViaGRPC(r.Context(), token)
	if err != nil {
		response.Unauthorized(w, "Invalid or expired token")
		return
	}

	// Token is valid, proceed with proxying
	serviceName, servicePath := h.serviceProxy.MapRouteToServicePath(r.URL.Path)
	h.proxyRequest(w, r, serviceName, servicePath)
}

// ProxyToProductService proxies requests to product service
func (h *ProxyHandler) ProxyToProductService(w http.ResponseWriter, r *http.Request) {
	serviceName, servicePath := h.serviceProxy.MapRouteToServicePath(r.URL.Path)
	h.proxyRequest(w, r, serviceName, servicePath)
}

// ProxyToOrderService proxies requests to order service
func (h *ProxyHandler) ProxyToOrderService(w http.ResponseWriter, r *http.Request) {
	serviceName, servicePath := h.serviceProxy.MapRouteToServicePath(r.URL.Path)
	h.proxyRequest(w, r, serviceName, servicePath)
}

// Generic proxy method
func (h *ProxyHandler) proxyRequest(w http.ResponseWriter, r *http.Request, serviceName, servicePath string) {
	// Update the request URL path for the target service
	r.URL.Path = servicePath

	// Include query parameters
	if r.URL.RawQuery != "" {
		r.URL.Path += "?" + r.URL.RawQuery
	}

	// Make the proxy request
	resp, err := h.serviceProxy.ProxyHTTPRequest(serviceName, r.URL.Path, r)
	if err != nil {
		log.Printf("Proxy request failed: %v", err)
		response.InternalServerError(w, "Service temporarily unavailable")
		return
	}
	defer resp.Body.Close()

	// Copy response headers
	for key, values := range resp.Header {
		for _, value := range values {
			w.Header().Add(key, value)
		}
	}

	// Set response status code
	w.WriteHeader(resp.StatusCode)

	// Copy response body
	_, err = io.Copy(w, resp.Body)
	if err != nil {
		log.Printf("Failed to copy response body: %v", err)
	}
}

// HealthCheck for API Gateway
func (h *ProxyHandler) HealthCheck(w http.ResponseWriter, r *http.Request) {
	healthStatus := map[string]interface{}{
		"status": "ok",
		"services": map[string]string{
			"user_service":         "unknown",
			"product_service":      "unknown",
			"order_service":        "unknown",
			"payment_service":      "unknown",
			"inventory_service":    "unknown",
			"notification_service": "unknown",
		},
	}

	// TODO: Implement actual health checks for each service
	// For now, just return basic status

	response.OK(w, "API Gateway is healthy", healthStatus)
}

// Route information endpoint
func (h *ProxyHandler) GetRoutes(w http.ResponseWriter, r *http.Request) {
	routes := map[string]interface{}{
		"version": "v1",
		"routes": map[string][]string{
			"auth": {
				"POST /api/v1/auth/login",
				"POST /api/v1/auth/register",
				"POST /api/v1/auth/refresh",
				"POST /api/v1/auth/validate",
				"POST /api/v1/auth/logout",
			},
			"users": {
				"GET /api/v1/users",
				"POST /api/v1/users",
				"GET /api/v1/users/{id}",
				"PUT /api/v1/users/{id}",
				"DELETE /api/v1/users/{id}",
				"GET /api/v1/users/profile",
			},
			"products": {
				"GET /api/v1/products",
				"POST /api/v1/products",
				"GET /api/v1/products/{id}",
				"PUT /api/v1/products/{id}",
				"DELETE /api/v1/products/{id}",
			},
			"orders": {
				"GET /api/v1/orders",
				"POST /api/v1/orders",
				"GET /api/v1/orders/{id}",
				"PUT /api/v1/orders/{id}",
				"DELETE /api/v1/orders/{id}",
			},
		},
	}

	response.OK(w, "Available routes", routes)
}
