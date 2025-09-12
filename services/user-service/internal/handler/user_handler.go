package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"

	"github.com/ecommerce/services/user-service/internal/models"
	"github.com/ecommerce/services/user-service/internal/service"
	"github.com/ecommerce/services/user-service/pkg/utils"
)

type UserHandler struct {
	userService service.UserService
}

func NewUserHandler(userService service.UserService) *UserHandler {
	return &UserHandler{
		userService: userService,
	}
}

func (h *UserHandler) CreateUser(w http.ResponseWriter, r *http.Request) {
	var req models.CreateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "Invalid request body")
		return
	}

	user, err := h.userService.CreateUser(r.Context(), &req)
	if err != nil {
		if err.Error() == "email already exists" {
			response.Conflict(w, err.Error())
			return
		}
		response.InternalServerError(w, err.Error())
		return
	}

	response.Created(w, "User created successfully", user)
}

func (h *UserHandler) GetUser(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.ParseInt(vars["id"], 10, 64)
	if err != nil {
		utils.WriteErrorResponse(w, http.StatusBadRequest, "Invalid user ID")
		return
	}

	user, err := h.userService.GetUserByID(r.Context(), id)
	if err != nil {
		status := http.StatusInternalServerError
		if err.Error() == "user not found" {
			status = http.StatusNotFound
		}
		utils.WriteErrorResponse(w, status, err.Error())
		return
	}

	response := models.UserResponse{
		Success: true,
		Message: "User retrieved successfully",
		Data:    user,
	}

	utils.WriteJSONResponse(w, http.StatusOK, response)
}

func (h *UserHandler) UpdateUser(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.ParseInt(vars["id"], 10, 64)
	if err != nil {
		utils.WriteErrorResponse(w, http.StatusBadRequest, "Invalid user ID")
		return
	}

	var req models.UpdateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.WriteErrorResponse(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	user, err := h.userService.UpdateUser(r.Context(), id, &req)
	if err != nil {
		status := http.StatusInternalServerError
		if err.Error() == "user not found" {
			status = http.StatusNotFound
		}
		utils.WriteErrorResponse(w, status, err.Error())
		return
	}

	response := models.UserResponse{
		Success: true,
		Message: "User updated successfully",
		Data:    user,
	}

	utils.WriteJSONResponse(w, http.StatusOK, response)
}

func (h *UserHandler) DeleteUser(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.ParseInt(vars["id"], 10, 64)
	if err != nil {
		response.BadRequest(w, "Invalid user ID")
		return
	}

	err = h.userService.DeleteUser(r.Context(), id)
	if err != nil {
		if err.Error() == "user not found" {
			response.NotFound(w, err.Error())
			return
		}
		response.InternalServerError(w, err.Error())
		return
	}

	response.OK(w, "User deleted successfully", nil)
}

func (h *UserHandler) ListUsers(w http.ResponseWriter, r *http.Request) {
	// Parse query parameters
	pageStr := r.URL.Query().Get("page")
	limitStr := r.URL.Query().Get("limit")
	role := r.URL.Query().Get("role")

	page := 1
	if pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}

	limit := 10
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 100 {
			limit = l
		}
	}

	users, total, err := h.userService.ListUsers(r.Context(), page, limit, role)
	if err != nil {
		response.InternalServerError(w, err.Error())
		return
	}

	response.WriteList(w, http.StatusOK, "Users retrieved successfully", users, total, page, limit)
}
