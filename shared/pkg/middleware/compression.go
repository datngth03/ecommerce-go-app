package middleware

import (
	"compress/gzip"
	"io"
	"net/http"
	"strings"
	"sync"

	"github.com/gin-gonic/gin"
)

// gzipResponseWriter wraps http.ResponseWriter to provide gzip compression
type gzipResponseWriter struct {
	gin.ResponseWriter
	gzipWriter *gzip.Writer
	minSize    int
	buffer     []byte
	status     int
}

// Write implements http.ResponseWriter
func (g *gzipResponseWriter) Write(data []byte) (int, error) {
	// If we haven't written headers yet, buffer the data
	if g.status == 0 {
		g.buffer = append(g.buffer, data...)

		// Check if we have enough data to decide whether to compress
		if len(g.buffer) >= g.minSize {
			g.writeHeaders()
			return g.gzipWriter.Write(g.buffer)
		}
		return len(data), nil
	}

	// Headers already written, write through gzip
	if g.gzipWriter != nil {
		return g.gzipWriter.Write(data)
	}

	// Not using compression, write directly
	return g.ResponseWriter.Write(data)
}

// WriteHeader implements http.ResponseWriter
func (g *gzipResponseWriter) WriteHeader(code int) {
	g.status = code

	// If we have buffered data, decide on compression now
	if len(g.buffer) > 0 {
		g.writeHeaders()
	}

	if g.gzipWriter == nil {
		g.ResponseWriter.WriteHeader(code)
	}
}

// writeHeaders decides whether to use compression and writes headers
func (g *gzipResponseWriter) writeHeaders() {
	// Check if response is large enough to compress
	if len(g.buffer) < g.minSize {
		g.gzipWriter = nil
		g.ResponseWriter.WriteHeader(g.status)
		g.ResponseWriter.Write(g.buffer)
		return
	}

	// Use compression
	g.ResponseWriter.Header().Set("Content-Encoding", "gzip")
	g.ResponseWriter.Header().Del("Content-Length") // gzip changes the length
	g.ResponseWriter.WriteHeader(g.status)

	if g.gzipWriter != nil {
		g.gzipWriter.Write(g.buffer)
	}
}

// Close closes the gzip writer
func (g *gzipResponseWriter) Close() error {
	// If we haven't written headers yet, write the buffer uncompressed
	if g.status == 0 && len(g.buffer) > 0 {
		g.status = http.StatusOK
		g.ResponseWriter.WriteHeader(g.status)
		g.ResponseWriter.Write(g.buffer)
		return nil
	}

	if g.gzipWriter != nil {
		return g.gzipWriter.Close()
	}
	return nil
}

// gzipWriterPool pools gzip writers for reuse
var gzipWriterPool = sync.Pool{
	New: func() interface{} {
		gz, err := gzip.NewWriterLevel(io.Discard, gzip.BestSpeed)
		if err != nil {
			panic(err)
		}
		return gz
	},
}

// CompressionConfig holds compression middleware configuration
type CompressionConfig struct {
	// MinSize is the minimum response size to compress (default: 1024 bytes = 1KB)
	MinSize int
	// Level is the compression level (1-9, default: gzip.BestSpeed)
	Level int
	// ExcludedTypes are content types to exclude from compression
	ExcludedTypes []string
	// ExcludedPaths are URL paths to exclude from compression
	ExcludedPaths []string
}

// DefaultCompressionConfig returns default compression configuration
func DefaultCompressionConfig() CompressionConfig {
	return CompressionConfig{
		MinSize: 1024, // 1KB
		Level:   gzip.BestSpeed,
		ExcludedTypes: []string{
			"image/jpeg",
			"image/png",
			"image/gif",
			"video/",
			"audio/",
		},
		ExcludedPaths: []string{
			"/metrics",
			"/health",
		},
	}
}

// CompressionMiddleware enables gzip compression for HTTP responses
func CompressionMiddleware() gin.HandlerFunc {
	return CompressionMiddlewareWithConfig(DefaultCompressionConfig())
}

// CompressionMiddlewareWithConfig enables gzip compression with custom config
func CompressionMiddlewareWithConfig(config CompressionConfig) gin.HandlerFunc {
	if config.MinSize == 0 {
		config.MinSize = 1024
	}
	if config.Level == 0 {
		config.Level = gzip.BestSpeed
	}

	return func(c *gin.Context) {
		// Check if client accepts gzip
		if !strings.Contains(c.GetHeader("Accept-Encoding"), "gzip") {
			c.Next()
			return
		}

		// Check if path should be excluded
		for _, path := range config.ExcludedPaths {
			if strings.HasPrefix(c.Request.URL.Path, path) {
				c.Next()
				return
			}
		}

		// Get gzip writer from pool
		gz := gzipWriterPool.Get().(*gzip.Writer)
		defer gzipWriterPool.Put(gz)

		// Reset the writer to write to the response
		gz.Reset(c.Writer)

		// Create custom response writer
		gzipWriter := &gzipResponseWriter{
			ResponseWriter: c.Writer,
			gzipWriter:     gz,
			minSize:        config.MinSize,
			buffer:         make([]byte, 0, config.MinSize),
		}

		c.Writer = gzipWriter
		c.Header("Vary", "Accept-Encoding")

		// Process request
		c.Next()

		// Check if we should exclude based on content type
		contentType := c.Writer.Header().Get("Content-Type")
		for _, excludedType := range config.ExcludedTypes {
			if strings.HasPrefix(contentType, excludedType) {
				// Don't compress this content type
				if gzipWriter.gzipWriter != nil {
					gzipWriter.gzipWriter = nil
				}
				break
			}
		}

		// Close the gzip writer
		gzipWriter.Close()
	}
}

// CompressionStats holds compression statistics
type CompressionStats struct {
	OriginalSize   int64
	CompressedSize int64
	Ratio          float64
}

// CalculateCompressionRatio calculates the compression ratio
func CalculateCompressionRatio(original, compressed int64) float64 {
	if original == 0 {
		return 0
	}
	return float64(compressed) / float64(original) * 100
}
