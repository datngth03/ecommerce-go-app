// internal/user/domain/user.go
package domain

import (
	"errors"
	"time"
)

var (
	ErrAddressNotFound = errors.New("address not found")
)

// Address represents a user's address in the domain.
// Address đại diện cho một địa chỉ của người dùng trong domain.
type Address struct {
	ID            string `json:"id"`
	FullName      string `json:"full_name"`
	PhoneNumber   string `json:"phone_number"`
	StreetAddress string `json:"street_address"`
	City          string `json:"city"`
	PostalCode    string `json:"postal_code"`
	Country       string `json:"country"`
	IsDefault     bool   `json:"is_default"`
}

// User represents the core user entity in the domain.
// User đại diện cho thực thể người dùng cốt lõi trong domain.
type User struct {
	ID          string    `json:"id"`
	Email       string    `json:"email"`
	Password    string    `json:"-"` // Password should not be marshaled to JSON
	FullName    string    `json:"full_name"`
	PhoneNumber string    `json:"phone_number,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	Addresses   []Address `json:"addresses"` // Đã thay thế Address string bằng slice của Address struct
}

// UpdatePassword updates the user's password.
// UpdatePassword cập nhật mật khẩu của người dùng.
func (u *User) UpdatePassword(newPassword string) {
	u.Password = newPassword
	u.UpdatedAt = time.Now()
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
		Addresses: []Address{}, // Khởi tạo slice địa chỉ
	}
}

// UpdateProfile updates user's profile information.
// UpdateProfile cập nhật thông tin hồ sơ người dùng.
func (u *User) UpdateProfile(fullName, phoneNumber string, address string) {
	u.FullName = fullName
	u.PhoneNumber = phoneNumber
	u.UpdatedAt = time.Now()
}

// AddAddress adds a new address to the user's profile.
func (u *User) AddAddress(addr Address) error {
	if len(u.Addresses) >= 5 {
		return errors.New("too many addresses")
	}
	u.Addresses = append(u.Addresses, addr)
	u.UpdatedAt = time.Now()
	return nil
}

// UpdateAddress updates an existing address for the user.
// UpdateAddress cập nhật một địa chỉ đã có của người dùng.
func (u *User) UpdateAddress(updatedAddress Address) error {
	for i, addr := range u.Addresses {
		if addr.ID == updatedAddress.ID {
			u.Addresses[i] = updatedAddress
			u.UpdatedAt = time.Now()
			return nil
		}
	}
	return ErrAddressNotFound
}

// DeleteAddress deletes a user's address by ID.
// DeleteAddress xóa một địa chỉ của người dùng bằng ID.
func (u *User) DeleteAddress(addressID string) error {
	for i, addr := range u.Addresses {
		if addr.ID == addressID {
			u.Addresses = append(u.Addresses[:i], u.Addresses[i+1:]...)
			u.UpdatedAt = time.Now()
			return nil
		}
	}
	return ErrAddressNotFound
}

// GetAddresses returns all addresses for the user.
// GetAddresses trả về tất cả địa chỉ của người dùng.
func (u *User) GetAddress() []Address {
	return u.Addresses
}

// HashPassword is a placeholder for password hashing logic.
// HashPassword là một placeholder cho logic băm mật khẩu.
func (u *User) HashPassword() error {
	// TODO: Implement actual password hashing using bcrypt or similar
	return nil // Placeholder
}

// CheckPassword is a placeholder for password checking logic.
// CheckPassword là một placeholder cho logic kiểm tra mật khẩu.
func (u *User) CheckPassword(password string) bool {
	// TODO: Implement actual password comparison
	return u.Password == password // Placeholder
}
