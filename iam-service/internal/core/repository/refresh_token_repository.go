package repository

import (
	"motrava/iam-service/internal/core/models"
)

// RefreshTokenRepository defines persistence operations for refresh tokens.
type RefreshTokenRepository interface {
	Create(token *models.RefreshToken) error
	FindByHash(hash string) (*models.RefreshToken, error)
	RevokeByHash(hash string) error
}
