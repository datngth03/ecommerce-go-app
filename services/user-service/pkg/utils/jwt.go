// pkg/utils/jwt.go
package utils

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// JWTClaims represents JWT token claims
type JWTClaims struct {
	UserID int64  `json:"user_id"`
	Email  string `json:"email"`
	// ExpiresAt int64  `json:"exp"`
	// IssuedAt  int64  `json:"iat"`
	jwt.RegisteredClaims
}

// TokenPair represents access and refresh token pair
type TokenPair struct {
	AccessToken      string    `json:"access_token"`
	RefreshToken     string    `json:"refresh_token"`
	AccessExpiresAt  time.Time `json:"access_expires_at"`
	RefreshExpiresAt time.Time `json:"refresh_expires_at"`
}

// RefreshTokenData represents refresh token data from storage
type RefreshTokenData struct {
	UserID    int64     `json:"user_id"`
	Token     string    `json:"token"`
	ExpiresAt time.Time `json:"expires_at"`
	CreatedAt time.Time `json:"created_at"`
}

// PasswordResetTokenData represents password reset token data
type PasswordResetTokenData struct {
	UserID    int64     `json:"user_id"`
	Token     string    `json:"token"`
	ExpiresAt time.Time `json:"expires_at"`
	CreatedAt time.Time `json:"created_at"`
}

// GenerateJWT generates a JWT access token
func GenerateJWT(userID int64, email string, expiresAt time.Time, secret string) (string, error) {
	now := time.Now()

	// Sử dụng các trường chuẩn từ jwt.RegisteredClaims
	claims := &JWTClaims{
		UserID: userID,
		Email:  email,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
		},
	}

	// Tạo token với claims và phương thức ký HS256
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Ký token với secret key và trả về chuỗi token
	signedToken, err := token.SignedString([]byte(secret))
	if err != nil {
		return "", fmt.Errorf("could not sign token: %w", err)
	}

	return signedToken, nil
}

// ValidateJWT validates a JWT token and returns its claims if valid
func ValidateJWT(tokenString string, secret string) (*JWTClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		// Kiểm tra phương thức ký, đảm bảo là HMAC
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(secret), nil
	})

	if err != nil {
		return nil, fmt.Errorf("could not parse token: %w", err)
	}

	// Trích xuất claims và kiểm tra tính hợp lệ
	claims, ok := token.Claims.(*JWTClaims)
	if !ok || !token.Valid {
		return nil, errors.New("invalid token")
	}

	return claims, nil
}

// GenerateRefreshToken creates a secure, random string for a refresh token
func GenerateRefreshToken() (string, error) {
	bytes := make([]byte, 32) // Tạo 32 bytes ngẫu nhiên
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

// GenerateTokenPair creates a new access and refresh token pair
func GenerateTokenPair(userID int64, email, secret string, accessDuration, refreshDuration time.Duration) (*TokenPair, error) {
	accessExpiresAt := time.Now().Add(accessDuration)
	refreshExpiresAt := time.Now().Add(refreshDuration)

	// Tạo access token
	accessToken, err := GenerateJWT(userID, email, accessExpiresAt, secret)
	if err != nil {
		return nil, fmt.Errorf("could not generate access token: %w", err)
	}

	// Tạo refresh token
	refreshToken, err := GenerateRefreshToken()
	if err != nil {
		return nil, fmt.Errorf("could not generate refresh token: %w", err)
	}

	return &TokenPair{
		AccessToken:      accessToken,
		RefreshToken:     refreshToken,
		AccessExpiresAt:  accessExpiresAt,
		RefreshExpiresAt: refreshExpiresAt,
	}, nil
}
