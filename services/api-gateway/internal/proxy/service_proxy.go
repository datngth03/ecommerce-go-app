package proxy

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/ecommerce/proto/user_service"
	"github.com/ecommerce/services/api-gateway/internal/config"
)

type ServiceProxy struct {
	config     *config.Config
	httpClient *http.Client
	grpcConns  map[string]*grpc.ClientConn
}

func NewServiceProxy(cfg *config.Config) *ServiceProxy {
	return &ServiceProxy{
		config: cfg,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		grpcConns: make(map[string]*grpc.ClientConn),
	}
}

// HTTP Proxy Methods
func (p *ServiceProxy) ProxyHTTPRequest(serviceName, path string, r *http.Request) (*http.Response, error) {
	targetURL := p.getServiceURL(serviceName, path)

	// Create new request
	proxyReq, err := http.NewRequestWithContext(r.Context(), r.Method, targetURL, r.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to create proxy request: %w", err)
	}

	// Copy headers
	p.copyHeaders(r.Header, proxyReq.Header)

	// Add service identification headers
	proxyReq.Header.Set("X-Gateway-Service", serviceName)
	proxyReq.Header.Set("X-Original-Path", r.URL.Path)

	// Make the request
	resp, err := p.httpClient.Do(proxyReq)
	if err != nil {
		return nil, fmt.Errorf("proxy request failed: %w", err)
	}

	return resp, nil
}

func (p *ServiceProxy) getServiceURL(serviceName, path string) string {
	var endpoint config.ServiceEndpoint

	switch serviceName {
	case "user":
		endpoint = p.config.Services.UserService
	case "product":
		endpoint = p.config.Services.ProductService
	case "order":
		endpoint = p.config.Services.OrderService
	case "payment":
		endpoint = p.config.Services.PaymentService
	case "inventory":
		endpoint = p.config.Services.InventoryService
	case "notification":
		endpoint = p.config.Services.NotificationService
	default:
		endpoint = p.config.Services.UserService // fallback
	}

	return endpoint.GetHTTPURL() + path
}

func (p *ServiceProxy) copyHeaders(src, dst http.Header) {
	for key, values := range src {
		// Skip hop-by-hop headers
		if p.isHopByHopHeader(key) {
			continue
		}
		for _, value := range values {
			dst.Add(key, value)
		}
	}
}

func (p *ServiceProxy) isHopByHopHeader(header string) bool {
	hopByHop := []string{
		"Connection",
		"Keep-Alive",
		"Proxy-Authenticate",
		"Proxy-Authorization",
		"Te",
		"Trailers",
		"Transfer-Encoding",
		"Upgrade",
	}

	header = strings.ToLower(header)
	for _, h := range hopByHop {
		if strings.ToLower(h) == header {
			return true
		}
	}
	return false
}

// gRPC Client Methods
func (p *ServiceProxy) GetUserServiceGRPCClient(ctx context.Context) (user_service.UserServiceClient, error) {
	conn, err := p.getGRPCConnection("user", p.config.Services.UserService.GetGRPCAddr())
	if err != nil {
		return nil, err
	}

	return user_service.NewUserServiceClient(conn), nil
}

func (p *ServiceProxy) getGRPCConnection(serviceName, addr string) (*grpc.ClientConn, error) {
	if conn, exists := p.grpcConns[serviceName]; exists {
		return conn, nil
	}

	conn, err := grpc.Dial(addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithTimeout(p.config.Services.UserService.Timeout),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to %s service: %w", serviceName, err)
	}

	p.grpcConns[serviceName] = conn
	return conn, nil
}

// Authentication via gRPC
func (p *ServiceProxy) ValidateTokenViaGRPC(ctx context.Context, token string) (*user_service.User, error) {
	client, err := p.GetUserServiceGRPCClient(ctx)
	if err != nil {
		return nil, err
	}

	resp, err := client.ValidateToken(ctx, &user_service.ValidateTokenRequest{
		Token: token,
	})
	if err != nil {
		return nil, fmt.Errorf("token validation failed: %w", err)
	}

	if !resp.Success {
		return nil, fmt.Errorf("token validation failed: %s", resp.Message)
	}

	return resp.User, nil
}

// Authentication via HTTP
func (p *ServiceProxy) ValidateTokenViaHTTP(ctx context.Context, token string) (map[string]interface{}, error) {
	reqBody := map[string]string{
		"token": token,
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, "POST",
		p.config.Services.UserService.GetHTTPURL()+"/auth/validate",
		bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := p.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("token validation failed with status: %d", resp.StatusCode)
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return result, nil
}

// Route mapping
func (p *ServiceProxy) MapRouteToServicePath(originalPath string) (string, string) {
	// Remove /api/v1 prefix
	path := strings.TrimPrefix(originalPath, "/api/v1")

	if strings.HasPrefix(path, "/users") || strings.HasPrefix(path, "/auth") {
		// Map gateway routes to service routes
		if strings.HasPrefix(path, "/users/register") {
			return "user", "/users"
		}
		if strings.HasPrefix(path, "/users/login") || strings.HasPrefix(path, "/auth/login") {
			return "user", "/auth/login"
		}
		if strings.HasPrefix(path, "/users/profile") {
			return "user", "/auth/profile"
		}
		if strings.HasPrefix(path, "/auth") {
			return "user", path
		}
		return "user", path
	}

	if strings.HasPrefix(path, "/products") {
		return "product", path
	}

	if strings.HasPrefix(path, "/orders") {
		return "order", path
	}

	if strings.HasPrefix(path, "/payments") {
		return "payment", path
	}

	// Default fallback
	return "user", path
}

// Cleanup connections
func (p *ServiceProxy) Close() {
	for _, conn := range p.grpcConns {
		conn.Close()
	}
}
