package grpcpool

import (
	"context"
	"fmt"
	"sync"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/connectivity"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/keepalive"
)

// ConnectionPool manages a pool of gRPC connections
type ConnectionPool struct {
	target      string
	poolSize    int
	connections []*grpc.ClientConn
	current     int
	mu          sync.RWMutex
	dialOpts    []grpc.DialOption
}

// PoolConfig contains configuration for the connection pool
type PoolConfig struct {
	Target   string // Target address (e.g., "localhost:50051")
	PoolSize int    // Number of connections in the pool (default: 5)

	// Keepalive settings
	KeepaliveTime    time.Duration // Time after which keepalive ping is sent (default: 30s)
	KeepaliveTimeout time.Duration // Timeout for keepalive ping (default: 10s)

	// Connection settings
	MaxConnectionIdle     time.Duration // Max idle time before connection closes (default: 5min)
	MaxConnectionAge      time.Duration // Max connection lifetime (default: 30min)
	MaxConnectionAgeGrace time.Duration // Grace period for connection close (default: 5min)

	// Additional dial options
	DialOptions []grpc.DialOption
}

// DefaultPoolConfig returns a configuration with sensible defaults
func DefaultPoolConfig(target string) *PoolConfig {
	return &PoolConfig{
		Target:                target,
		PoolSize:              5,
		KeepaliveTime:         30 * time.Second,
		KeepaliveTimeout:      10 * time.Second,
		MaxConnectionIdle:     5 * time.Minute,
		MaxConnectionAge:      30 * time.Minute,
		MaxConnectionAgeGrace: 5 * time.Minute,
	}
}

// NewConnectionPool creates a new gRPC connection pool
func NewConnectionPool(config *PoolConfig) (*ConnectionPool, error) {
	if config.PoolSize <= 0 {
		config.PoolSize = 5 // Default pool size
	}

	// Build dial options
	dialOpts := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithKeepaliveParams(keepalive.ClientParameters{
			Time:                config.KeepaliveTime,
			Timeout:             config.KeepaliveTimeout,
			PermitWithoutStream: true,
		}),
		grpc.WithDefaultCallOptions(
			grpc.MaxCallRecvMsgSize(10*1024*1024), // 10MB
			grpc.MaxCallSendMsgSize(10*1024*1024), // 10MB
		),
	}

	// Add custom dial options
	if len(config.DialOptions) > 0 {
		dialOpts = append(dialOpts, config.DialOptions...)
	}

	pool := &ConnectionPool{
		target:      config.Target,
		poolSize:    config.PoolSize,
		connections: make([]*grpc.ClientConn, config.PoolSize),
		dialOpts:    dialOpts,
	}

	// Create all connections
	for i := 0; i < config.PoolSize; i++ {
		conn, err := grpc.NewClient(config.Target, dialOpts...)
		if err != nil {
			// Close already created connections
			pool.Close()
			return nil, fmt.Errorf("failed to create connection %d: %w", i, err)
		}
		pool.connections[i] = conn
	}

	// Start health check goroutine
	go pool.healthCheck()

	return pool, nil
}

// Get returns the next available connection from the pool (round-robin)
func (p *ConnectionPool) Get() *grpc.ClientConn {
	p.mu.Lock()
	defer p.mu.Unlock()

	conn := p.connections[p.current]
	p.current = (p.current + 1) % p.poolSize

	return conn
}

// GetHealthy returns a healthy connection from the pool
func (p *ConnectionPool) GetHealthy(ctx context.Context) (*grpc.ClientConn, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	// Try to find a ready connection
	for i := 0; i < p.poolSize; i++ {
		idx := (p.current + i) % p.poolSize
		conn := p.connections[idx]

		state := conn.GetState()
		if state == connectivity.Ready || state == connectivity.Idle {
			return conn, nil
		}
	}

	// If no ready connection, wait for one to become ready
	for i := 0; i < p.poolSize; i++ {
		idx := (p.current + i) % p.poolSize
		conn := p.connections[idx]

		if conn.WaitForStateChange(ctx, connectivity.TransientFailure) {
			state := conn.GetState()
			if state == connectivity.Ready || state == connectivity.Idle {
				return conn, nil
			}
		}
	}

	return nil, fmt.Errorf("no healthy connection available")
}

// GetAll returns all connections in the pool
func (p *ConnectionPool) GetAll() []*grpc.ClientConn {
	p.mu.RLock()
	defer p.mu.RUnlock()

	return p.connections
}

// Size returns the size of the connection pool
func (p *ConnectionPool) Size() int {
	return p.poolSize
}

// Target returns the target address
func (p *ConnectionPool) Target() string {
	return p.target
}

// healthCheck periodically checks and repairs unhealthy connections
func (p *ConnectionPool) healthCheck() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		p.mu.Lock()
		for i, conn := range p.connections {
			state := conn.GetState()

			// If connection is in a bad state, try to reconnect
			if state == connectivity.Shutdown || state == connectivity.TransientFailure {
				fmt.Printf("Connection %d to %s is unhealthy (state: %s), attempting to reconnect...\n",
					i, p.target, state)

				// Close old connection
				conn.Close()

				// Create new connection
				newConn, err := grpc.NewClient(p.target, p.dialOpts...)
				if err != nil {
					fmt.Printf("Failed to recreate connection %d: %v\n", i, err)
					continue
				}

				p.connections[i] = newConn
				fmt.Printf("Successfully recreated connection %d\n", i)
			}
		}
		p.mu.Unlock()
	}
}

// Close closes all connections in the pool
func (p *ConnectionPool) Close() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	var errs []error
	for i, conn := range p.connections {
		if conn != nil {
			if err := conn.Close(); err != nil {
				errs = append(errs, fmt.Errorf("failed to close connection %d: %w", i, err))
			}
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("errors closing connections: %v", errs)
	}

	return nil
}

// GetStats returns statistics about the connection pool
func (p *ConnectionPool) GetStats() *PoolStats {
	p.mu.RLock()
	defer p.mu.RUnlock()

	stats := &PoolStats{
		PoolSize:    p.poolSize,
		Target:      p.target,
		Connections: make([]ConnectionState, p.poolSize),
	}

	for i, conn := range p.connections {
		state := conn.GetState()
		stats.Connections[i] = ConnectionState{
			Index: i,
			State: state.String(),
		}

		switch state {
		case connectivity.Ready:
			stats.ReadyCount++
		case connectivity.Idle:
			stats.IdleCount++
		case connectivity.Connecting:
			stats.ConnectingCount++
		case connectivity.TransientFailure:
			stats.FailureCount++
		case connectivity.Shutdown:
			stats.ShutdownCount++
		}
	}

	return stats
}

// PoolStats contains statistics about the connection pool
type PoolStats struct {
	PoolSize        int
	Target          string
	ReadyCount      int
	IdleCount       int
	ConnectingCount int
	FailureCount    int
	ShutdownCount   int
	Connections     []ConnectionState
}

// ConnectionState represents the state of a single connection
type ConnectionState struct {
	Index int
	State string
}

// IsHealthy returns true if the pool has at least one healthy connection
func (s *PoolStats) IsHealthy() bool {
	return s.ReadyCount > 0 || s.IdleCount > 0
}

// HealthyPercentage returns the percentage of healthy connections
func (s *PoolStats) HealthyPercentage() float64 {
	if s.PoolSize == 0 {
		return 0
	}
	healthy := s.ReadyCount + s.IdleCount
	return float64(healthy) / float64(s.PoolSize) * 100
}
