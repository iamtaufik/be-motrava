package dto

import "time"

// CreateUserRequest is a local user creation payload.
type CreateUserRequest struct {
	Email       string `json:"email"`
	FullName    string `json:"full_name"`
	Password    string `json:"password"`
	PhoneNumber string `json:"phone_number"`
}

// UserResponse is the public representation of a user entity.
type UserResponse struct {
	ID           string     `json:"id"`
	Email        string     `json:"email"`
	FullName     *string    `json:"full_name,omitempty"`
	AuthProvider string     `json:"auth_provider"`
	GoogleSub    *string    `json:"google_sub,omitempty"`
	AvatarURL    *string    `json:"avatar_url,omitempty"`
	PhoneNumber  *string    `json:"phone_number,omitempty"`
	IsActive     bool       `json:"is_active"`
	IsVerified   bool       `json:"is_verified"`
	LastLoginAt  *time.Time `json:"last_login_at,omitempty"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
}
