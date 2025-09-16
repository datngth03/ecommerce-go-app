package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/datngth03/ecommerce-go-app/services/user-service/internal/service"
	"github.com/datngth03/ecommerce-go-app/services/user-service/pkg/utils"
)

// contextKey là một kiểu dữ liệu riêng để tránh xung đột key trong context.
type contextKey string

const (
	// UserContextKey là key để lưu và truy xuất thông tin claims của user từ context.
	UserContextKey contextKey = "claims"
)

// AuthMiddleware chứa các dependency cần thiết cho middleware xác thực.
type AuthMiddleware struct {
	authService service.AuthServiceInterface // Sử dụng interface thay vì struct cụ thể
}

// NewAuthMiddleware tạo một instance mới của AuthMiddleware.
func NewAuthMiddleware(authService service.AuthServiceInterface) *AuthMiddleware {
	return &AuthMiddleware{
		authService: authService,
	}
}

// RequireAuth là middleware yêu cầu request phải có token hợp lệ.
func (m *AuthMiddleware) RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Lấy token từ header "Authorization"
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			utils.WriteErrorResponse(w, http.StatusUnauthorized, "Authorization header is required")
			return
		}

		// Tách chuỗi "Bearer <token>"
		tokenParts := strings.Split(authHeader, " ")
		if len(tokenParts) != 2 || tokenParts[0] != "Bearer" {
			utils.WriteErrorResponse(w, http.StatusUnauthorized, "Invalid authorization header format (must be Bearer token)")
			return
		}

		tokenString := tokenParts[1]

		// Xác thực access token
		claims, err := m.authService.ValidateAccessToken(r.Context(), tokenString)
		if err != nil {
			utils.WriteErrorResponse(w, http.StatusUnauthorized, "Invalid or expired token")
			return
		}

		// Thêm thông tin claims của user vào context của request
		ctx := context.WithValue(r.Context(), UserContextKey, claims)

		// Chuyển request đã chứa context mới cho handler tiếp theo
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// OptionalAuth là middleware sẽ xác thực token nếu có, nhưng không bắt buộc.
func (m *AuthMiddleware) OptionalAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader != "" {
			tokenParts := strings.Split(authHeader, " ")
			if len(tokenParts) == 2 && tokenParts[0] == "Bearer" {
				tokenString := tokenParts[1]

				// Cố gắng xác thực token
				claims, err := m.authService.ValidateAccessToken(r.Context(), tokenString)
				// Nếu thành công, thêm claims vào context
				if err == nil && claims != nil {
					ctx := context.WithValue(r.Context(), UserContextKey, claims)
					r = r.WithContext(ctx) // Cập nhật request với context mới
				}
			}
		}

		// Luôn cho request đi tiếp dù có token hay không
		next.ServeHTTP(w, r)
	})
}

// GetUserFromContext là hàm hỗ trợ để lấy thông tin claims từ context một cách an toàn.
// Trả về claims và một boolean cho biết có tìm thấy hay không.
func GetUserFromContext(ctx context.Context) (*utils.JWTClaims, bool) {
	value := ctx.Value(UserContextKey)
	if value == nil {
		return nil, false
	}

	claims, ok := value.(*utils.JWTClaims)
	return claims, ok
}
