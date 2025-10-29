package handler

import (
	"net/http"

	"github.com/datngth03/ecommerce-go-app/services/api-gateway/internal/clients"
	"github.com/gin-gonic/gin"
)

// HealthHandler handles health check endpoints
type HealthHandler struct {
	clients *clients.Clients
}

// NewHealthHandler creates a new health handler
func NewHealthHandler(clients *clients.Clients) *HealthHandler {
	return &HealthHandler{
		clients: clients,
	}
}

// HealthCheck returns basic health status
func (h *HealthHandler) HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":  "healthy",
		"service": "api-gateway",
	})
}

// ReadinessCheck returns readiness status
func (h *HealthHandler) ReadinessCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status": "ready",
	})
}

// PoolsHealth returns connection pool health statistics
func (h *HealthHandler) PoolsHealth(c *gin.Context) {
	stats := h.clients.GetPoolStats()

	if stats == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"status": "unavailable",
			"error":  "connection pool manager not initialized",
		})
		return
	}

	response := gin.H{
		"status":   "healthy",
		"services": make(map[string]interface{}),
	}

	allHealthy := true
	for serviceName, stat := range stats {
		healthyPercentage := stat.HealthyPercentage()

		serviceStatus := gin.H{
			"healthy_percentage": healthyPercentage,
			"ready_connections":  stat.ReadyCount,
			"idle_connections":   stat.IdleCount,
			"connecting":         stat.ConnectingCount,
			"failed_connections": stat.FailureCount,
			"shutdown":           stat.ShutdownCount,
			"total_connections":  stat.PoolSize,
		}

		// Consider unhealthy if less than 50% connections are ready
		if !stat.IsHealthy() || healthyPercentage < 50.0 {
			allHealthy = false
			serviceStatus["status"] = "unhealthy"
			serviceStatus["warning"] = "low connection availability"
		} else if healthyPercentage < 80.0 {
			serviceStatus["status"] = "degraded"
			serviceStatus["warning"] = "some connections unavailable"
		} else {
			serviceStatus["status"] = "healthy"
		}

		response["services"].(map[string]interface{})[serviceName] = serviceStatus
	}

	// Set overall status
	if !allHealthy {
		response["status"] = "degraded"
		c.JSON(http.StatusServiceUnavailable, response)
		return
	}

	c.JSON(http.StatusOK, response)
}

// DetailedPoolsHealth returns detailed connection pool statistics
func (h *HealthHandler) DetailedPoolsHealth(c *gin.Context) {
	stats := h.clients.GetPoolStats()

	if stats == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"status": "unavailable",
			"error":  "connection pool manager not initialized",
		})
		return
	}

	response := gin.H{
		"status":   "healthy",
		"services": make(map[string]interface{}),
		"summary": gin.H{
			"total_services":     len(stats),
			"healthy_services":   0,
			"degraded_services":  0,
			"unhealthy_services": 0,
			"total_connections":  0,
			"ready_connections":  0,
			"failed_connections": 0,
		},
	}

	healthyServices := 0
	degradedServices := 0
	unhealthyServices := 0
	totalConnections := 0
	readyConnections := 0
	failedConnections := 0

	for serviceName, stat := range stats {
		healthyPercentage := stat.HealthyPercentage()

		serviceStatus := gin.H{
			"healthy_percentage": healthyPercentage,
			"ready_connections":  stat.ReadyCount,
			"idle_connections":   stat.IdleCount,
			"connecting":         stat.ConnectingCount,
			"failed_connections": stat.FailureCount,
			"shutdown":           stat.ShutdownCount,
			"total_connections":  stat.PoolSize,
			"is_healthy":         stat.IsHealthy(),
		}

		totalConnections += stat.PoolSize
		readyConnections += stat.ReadyCount
		failedConnections += stat.FailureCount

		// Categorize service health
		if !stat.IsHealthy() || healthyPercentage < 50.0 {
			unhealthyServices++
			serviceStatus["status"] = "unhealthy"
			serviceStatus["warning"] = "low connection availability"
		} else if healthyPercentage < 80.0 {
			degradedServices++
			serviceStatus["status"] = "degraded"
			serviceStatus["warning"] = "some connections unavailable"
		} else {
			healthyServices++
			serviceStatus["status"] = "healthy"
		}

		response["services"].(map[string]interface{})[serviceName] = serviceStatus
	}

	// Update summary
	summary := response["summary"].(gin.H)
	summary["healthy_services"] = healthyServices
	summary["degraded_services"] = degradedServices
	summary["unhealthy_services"] = unhealthyServices
	summary["total_connections"] = totalConnections
	summary["ready_connections"] = readyConnections
	summary["failed_connections"] = failedConnections

	// Calculate overall health percentage
	if totalConnections > 0 {
		summary["overall_health_percentage"] = float64(readyConnections) / float64(totalConnections) * 100.0
	} else {
		summary["overall_health_percentage"] = 0.0
	}

	// Set overall status
	if unhealthyServices > 0 {
		response["status"] = "unhealthy"
		c.JSON(http.StatusServiceUnavailable, response)
		return
	} else if degradedServices > 0 {
		response["status"] = "degraded"
		c.JSON(http.StatusOK, response)
		return
	}

	c.JSON(http.StatusOK, response)
}
