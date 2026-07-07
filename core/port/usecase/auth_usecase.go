package usecase

import (
	"context"

	"motrava/core/dto"
)

// AuthUsecase defines authentication flows exposed to handlers.
type AuthUsecase interface {
	GetGoogleAuthURL(state string) (string, error)
	HandleGoogleCallback(ctx context.Context, code string) (*dto.UserResponse, error)
	HandleGoogleMobileLogin(ctx context.Context, idToken string) (*dto.UserResponse, error)
}
