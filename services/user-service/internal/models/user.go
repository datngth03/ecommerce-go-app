package models

import (
	"time"
)

type User struct {
	ID        int64     `json:"id" db:"id"`
	Email     string    `json:"email" db:"email" validate:"required,email"`
	Password  string    `json:"-" db:"password" validate:"required,min=6"`
	Name      string    `json:"name" db:"name" validate:"required,min=2"`
	Phone     string    `json:"phone" db:"phone"`
	Role      string    `json:"role" db:"role" validate:"required,oneof=admin customer"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

type CreateUserRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=6"`
	Name     string `json:"name" validate:"required,min=2"`
	Phone    string `json:"phone"`
	Role     string `json:"role" validate:"required,oneof=admin customer"`
}

type UpdateUserRequest struct {
	Name  string `json:"name" validate:"required,min=2"`
	Phone string `json:"phone"`
	Role  string `json:"role" validate:"required,oneof=admin customer"`
}

type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

type LoginResponse struct {
	Token string `json:"token"`
	User  *User  `json:"user"`
}

type UserResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Data    *User  `json:"data,omitempty"`
}

type UsersResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Data    []User `json:"data,omitempty"`
	Total   int    `json:"total"`
}
