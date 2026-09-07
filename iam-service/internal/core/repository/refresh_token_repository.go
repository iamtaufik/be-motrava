package repository

import (
	"errors"

	"motrava/iam-service/internal/core/models"
)

// ErrRefreshTokenReuse is returned when a refresh token that was already
// revoked is being rotated again (reuse detection).
var ErrRefreshTokenReuse = errors.New("refresh token already revoked")

// RefreshTokenRepository defines persistence operations for refresh tokens.
type RefreshTokenRepository interface {
	Create(token *models.RefreshToken) error
	FindByHash(hash string) (*models.RefreshToken, error)
	RevokeByHash(hash string) error
	// Rotate atomically revokes the token identified by oldHash and creates
	// the replacement token in a single transaction. It returns
	// ErrRefreshTokenReuse if oldHash was already revoked.
	Rotate(oldHash string, newToken *models.RefreshToken) error
}
