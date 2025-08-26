// internal/user/domain/user.go
package domain

import (
	"errors"
	"time"
)

var (
	ErrAddressNotFound   = errors.New("address not found")
	ErrUserNotFound      = errors.New("user not found")
	ErrUserAlreadyExists = errors.New("user already exists")
)

// Address represents a user's address in the domain.
// Address đại diện cho một địa chỉ của người dùng trong domain.
type Address struct {
	ID            string    `json:"id"`
	FullName      string    `json:"full_name"`
	StreetAddress string    `json:"street_address"`
	City          string    `json:"city"`
	PostalCode    string    `json:"postal_code"`
	Country       string    `json:"country"`
	IsDefault     bool      `json:"is_default"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// User represents the core user entity in the domain.
// User đại diện cho thực thể người dùng cốt lõi trong domain.
type User struct {
	ID           string    `json:"id"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"-"` // Password should not be marshaled to JSON
	FirstName    string    `json:"first_name"`
	LastName     string    `json:"last_name"`
	PhoneNumber  string    `json:"phone_number,omitempty"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
	Addresses    []Address `json:"addresses"` // Đã thay thế Address string bằng slice của Address struct
}

// NewUser creates a new User instance.
// NewUser tạo một thể hiện User mới.
func NewUser(id, email, passwordHash, firstName, lastName string) *User {
	now := time.Now()
	return &User{
		ID:           id,
		Email:        email,
		PasswordHash: passwordHash,
		FirstName:    firstName,
		LastName:     lastName,
		CreatedAt:    now,
		UpdatedAt:    now,
		Addresses:    []Address{}, // Khởi tạo slice địa chỉ
	}
}

// NewAddress creates a new Address instance.
// NewAddress tạo một thể hiện Address mới.
func NewAddress(id, fullName, streetAddress, city, postalCode, country string, isDefault bool) Address {
	now := time.Now()
	return Address{
		ID:            id,
		FullName:      fullName,
		StreetAddress: streetAddress,
		City:          city,
		PostalCode:    postalCode,
		Country:       country,
		IsDefault:     isDefault,
		CreatedAt:     now,
		UpdatedAt:     now,
	}
}

// AddAddress adds a new address to the user's profile.
// AddAddress thêm một địa chỉ mới vào hồ sơ của người dùng.
func (u *User) AddAddress(address Address) {
	if address.IsDefault {
		// Đảm bảo chỉ có một địa chỉ mặc định
		for i := range u.Addresses {
			u.Addresses[i].IsDefault = false
		}
	}
	u.Addresses = append(u.Addresses, address)
	u.UpdatedAt = time.Now()
}

// UpdateAddress updates an existing address in the user's profile.
// UpdateAddress cập nhật một địa chỉ hiện có trong hồ sơ của người dùng.
func (u *User) UpdateAddress(updated Address) error {
	found := false
	for i, addr := range u.Addresses {
		if addr.ID == updated.ID {
			// Cập nhật thông tin địa chỉ
			u.Addresses[i].FullName = updated.FullName
			u.Addresses[i].StreetAddress = updated.StreetAddress
			u.Addresses[i].City = updated.City
			u.Addresses[i].PostalCode = updated.PostalCode
			u.Addresses[i].Country = updated.Country
			u.Addresses[i].IsDefault = updated.IsDefault
			u.Addresses[i].UpdatedAt = time.Now()
			found = true
			break
		}
	}
	if !found {
		return ErrAddressNotFound
	}
	// Nếu địa chỉ được cập nhật là mặc định, đảm bảo không có địa chỉ nào khác mặc định.
	if updated.IsDefault {
		for i := range u.Addresses {
			if u.Addresses[i].ID != updated.ID {
				u.Addresses[i].IsDefault = false
			}
		}
	}
	u.UpdatedAt = time.Now()
	return nil
}

// RemoveAddress removes an address from the user's profile by ID.
// RemoveAddress xóa một địa chỉ khỏi hồ sơ của người dùng theo ID.
func (u *User) RemoveAddress(addressID string) error {
	for i, addr := range u.Addresses {
		if addr.ID == addressID {
			// Xóa địa chỉ khỏi slice
			u.Addresses = append(u.Addresses[:i], u.Addresses[i+1:]...)
			u.UpdatedAt = time.Now()
			return nil
		}
	}
	return ErrAddressNotFound
}
