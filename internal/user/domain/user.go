// internal/user/domain/user.go
package domain

import (
	"time"
)

// User represents the core user entity in the domain.
// User đại diện cho thực thể người dùng cốt lõi trong domain.
type User struct {
	ID          string    `json:"id"`
	Email       string    `json:"email"`
	Password    string    `json:"-"` // Password should not be marshaled to JSON
	FullName    string    `json:"full_name"`
	PhoneNumber string    `json:"phone_number,omitempty"`
	Address     string    `json:"address,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// NewUser creates a new User instance.
// NewUser tạo một thể hiện User mới.
func NewUser(id, email, password, fullName string) *User {
	now := time.Now()
	return &User{
		ID:        id,
		Email:     email,
		Password:  password,
		FullName:  fullName,
		CreatedAt: now,
		UpdatedAt: now,
	}
}

// UpdateProfile updates user's profile information.
// UpdateProfile cập nhật thông tin hồ sơ người dùng.
func (u *User) UpdateProfile(fullName, phoneNumber, address string) {
	u.FullName = fullName
	u.PhoneNumber = phoneNumber
	u.Address = address
	u.UpdatedAt = time.Now()
}

// HashPassword is a placeholder for password hashing logic.
// In a real application, this would use a strong hashing algorithm like bcrypt.
// HashPassword là một placeholder cho logic băm mật khẩu.
// Trong ứng dụng thực tế, điều này sẽ sử dụng thuật toán băm mạnh như bcrypt.
func (u *User) HashPassword() error {
	// TODO: Implement actual password hashing using bcrypt or similar
	// Ví dụ: hashedPassword, err := bcrypt.GenerateFromPassword([]byte(u.Password), bcrypt.DefaultCost)
	// u.Password = string(hashedPassword)
	// return err
	return nil // Placeholder
}

// CheckPassword is a placeholder for password checking logic.
// CheckPassword là một placeholder cho logic kiểm tra mật khẩu.
func (u *User) CheckPassword(password string) bool {
	// TODO: Implement actual password comparison
	// Ví dụ: err := bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(password))
	// return err == nil
	return u.Password == password // Placeholder
}
