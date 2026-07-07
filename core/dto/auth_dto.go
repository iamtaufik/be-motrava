package dto

// AuthRegisterRequest is the payload for email/password registration.
type AuthRegisterRequest struct {
	FullName string `json:"full_name"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

// AuthLoginRequest is the payload for email/password login.
type AuthLoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// AuthRefreshRequest is the payload for refresh token exchange.
type AuthRefreshRequest struct {
	RefreshToken string `json:"refresh_token"`
}

// AuthResponse is returned after register, login, refresh, and Google mobile auth.
type AuthResponse struct {
	User         UserResponse `json:"user"`
	AccessToken  string       `json:"access_token"`
	RefreshToken string       `json:"refresh_token"`
	TokenType    string       `json:"token_type"`
	ExpiresIn    int64        `json:"expires_in"`
}
