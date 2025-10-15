// services/product-service/internal/middleware/logging.go

package middleware

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// RequestIDKey là key để lưu trữ request ID trong context của Gin.
const RequestIDKey = "requestID"

// StructuredLogger tạo một middleware mới để ghi log có cấu trúc cho mỗi request.
// Nó nhận vào một instance của slog.Logger để sử dụng.
func StructuredLogger(logger *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		requestID := uuid.New().String()
		c.Set(RequestIDKey, requestID)

		c.Next()

		latency := time.Since(start)
		status := c.Writer.Status()
		method := c.Request.Method
		path := c.Request.URL.Path
		clientIP := c.ClientIP()
		userAgent := c.Request.UserAgent()
		errorMessage := c.Errors.ByType(gin.ErrorTypePrivate).String()

		// Tạo các thuộc tính (attributes) cho structured log
		logAttributes := []slog.Attr{
			slog.Int("status", status),
			slog.String("method", method),
			slog.String("path", path),
			slog.String("ip", clientIP),
			slog.Duration("latency", latency),
			slog.String("user_agent", userAgent),
			slog.String("request_id", requestID),
		}

		if errorMessage != "" {
			logAttributes = append(logAttributes, slog.String("error", errorMessage))
		}

		// =======================================================
		// THAY ĐỔI Ở ĐÂY: Sử dụng logger.LogAttrs thay vì logger.Info/Warn/Error
		// =======================================================
		switch {
		case status >= http.StatusInternalServerError:
			// Truyền cả context của request vào log là một good practice
			logger.LogAttrs(c.Request.Context(), slog.LevelError, "Server Error", logAttributes...)
		case status >= http.StatusBadRequest:
			logger.LogAttrs(c.Request.Context(), slog.LevelWarn, "Client Error", logAttributes...)
		default:
			logger.LogAttrs(c.Request.Context(), slog.LevelInfo, "Request Handled", logAttributes...)
		}
	}
}
