package usecase

import (
	"context"
	"errors"

	"motrava/iam-service/internal/core/dto"
)

var (
	ErrEmailAlreadyExists  = errors.New("email already registered")
	ErrInvalidCredentials  = errors.New("invalid credentials")
	ErrInvalidRefreshToken = errors.New("invalid refresh token")
)

// AuthUsecase defines authentication flows exposed to handlers.
type AuthUsecase interface {
	Register(ctx context.Context, req dto.AuthRegisterRequest) (*dto.AuthResponse, error)
	Login(ctx context.Context, req dto.AuthLoginRequest) (*dto.AuthResponse, error)
	Refresh(ctx context.Context, refreshToken string) (*dto.AuthResponse, error)
	Me(ctx context.Context, accessToken string) (*dto.UserResponse, error)
	GetGoogleAuthURL(state string) (string, error)
	HandleGoogleCallback(ctx context.Context, code string) (*dto.UserResponse, error)
	HandleGoogleMobileLogin(ctx context.Context, idToken string) (*dto.AuthResponse, error)
}
