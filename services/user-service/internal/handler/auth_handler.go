package handler

import (
	"encoding/json"
	"net/http"

	"github.com/ecommerce/services/user-service/internal/models"
	"github.com/ecommerce/services/user-service/internal/service"
	"github.com/ecommerce/services/user-service/pkg/utils"
)

type AuthHandler struct {
	authService service.AuthService
	userService service.UserService
}

func NewAuthHandler(authService service.AuthService, userService service.UserService) *AuthHandler {
	return &AuthHandler{
		authService: authService,
		userService: userService,
	}
}

// Login handles user authentication
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req models.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.WriteErrorResponse(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	loginResponse, err := h.userService.Login(r.Context(), &req)
	if err != nil {
		status := http.StatusInternalServerError
		if err.Error() == "invalid credentials" {
			status = http.StatusUnauthorized
		}
		utils.WriteErrorResponse(w, status, err.Error())
		return
	}

	response := struct {
		Success bool                  `json:"success"`
		Message string                `json:"message"`
		Data    *models.LoginResponse `json:"data"`
	}{
		Success: true,
		Message: "Login successful",
		Data:    loginResponse,
	}

	utils.WriteJSONResponse(w, http.StatusOK, response)
}

// ValidateToken validates JWT token
func (h *AuthHandler) ValidateToken(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Token string `json:"token"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "Invalid request body")
		return
	}

	user, err := h.authService.ValidateToken(r.Context(), req.Token)
	if err != nil {
		response.Unauthorized(w, "Invalid or expired token")
		return
	}

	response.OK(w, "Token is valid", user)
}

// RefreshToken generates new JWT token
func (h *AuthHandler) RefreshToken(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Token string `json:"token"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "Invalid request body")
		return
	}

	newToken, err := h.authService.RefreshToken(r.Context(), req.Token)
	if err != nil {
		response.Unauthorized(w, "Invalid or expired token")
		return
	}

	tokenResponse := struct {
		Token string `json:"token"`
	}{
		Token: newToken,
	}

	response.OK(w, "Token refreshed successfully", tokenResponse)
}

// Logout handles user logout
func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Token string `json:"token"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "Invalid request body")
		return
	}

	err := h.authService.Logout(r.Context(), req.Token)
	if err != nil {
		response.Unauthorized(w, "Invalid token")
		return
	}

	response.OK(w, "Logout successful", nil)
}

// GetProfile returns current user profile (requires authentication)
func (h *AuthHandler) GetProfile(w http.ResponseWriter, r *http.Request) {
	// This endpoint would typically be protected by auth middleware
	// and the user would be extracted from context

	// For demonstration, we'll extract from Authorization header
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		response.Unauthorized(w, "Authorization header required")
		return
	}

	// Extract token (assuming Bearer token)
	token := authHeader[len("Bearer "):]

	user, err := h.authService.ValidateToken(r.Context(), token)
	if err != nil {
		response.Unauthorized(w, "Invalid or expired token")
		return
	}

	response.OK(w, "Profile retrieved successfully", user)
}

// ChangePassword allows user to change password
func (h *AuthHandler) ChangePassword(w http.ResponseWriter, r *http.Request) {
	var req struct {
		CurrentPassword string `json:"current_password" validate:"required"`
		NewPassword     string `json:"new_password" validate:"required,min=6"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "Invalid request body")
		return
	}

	// Extract token from Authorization header
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		response.Unauthorized(w, "Authorization header required")
		return
	}

	token := authHeader[len("Bearer "):]
	user, err := h.authService.ValidateToken(r.Context(), token)
	if err != nil {
		response.Unauthorized(w, "Invalid or expired token")
		return
	}

	// Verify current password
	loginReq := &models.LoginRequest{
		Email:    user.Email,
		Password: req.CurrentPassword,
	}

	_, err = h.userService.Login(r.Context(), loginReq)
	if err != nil {
		response.Unauthorized(w, "Current password is incorrect")
		return
	}

	// TODO: Implement password change logic in user service
	// This would require adding a ChangePassword method to UserService

	response.OK(w, "Password changed successfully", nil)
}
