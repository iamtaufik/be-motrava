package repository

import (
	"log/slog"
	"time"

	"gorm.io/gorm"

	"motrava/iam-service/internal/core/models"
	coreRepo "motrava/iam-service/internal/core/repository"
)

type refreshTokenRepositoryGorm struct {
	db  *gorm.DB
	log *slog.Logger
}

func NewRefreshTokenRepositoryGorm(db *gorm.DB, logger *slog.Logger) coreRepo.RefreshTokenRepository {
	return &refreshTokenRepositoryGorm{db: db, log: logger}
}

func (r *refreshTokenRepositoryGorm) Create(token *models.RefreshToken) error {
	if err := r.db.Create(token).Error; err != nil {
		r.log.Error("failed to create refresh token", "module", "refresh_token_repository", "error", err)
		return err
	}

	return nil
}

func (r *refreshTokenRepositoryGorm) FindByHash(hash string) (*models.RefreshToken, error) {
	var token models.RefreshToken
	if err := r.db.Where("token_hash = ?", hash).First(&token).Error; err != nil {
		r.log.Error("failed to query refresh token", "module", "refresh_token_repository", "error", err)
		return nil, err
	}

	return &token, nil
}

func (r *refreshTokenRepositoryGorm) RevokeByHash(hash string) error {
	now := time.Now().UTC()
	result := r.db.Model(&models.RefreshToken{}).
		Where("token_hash = ? AND revoked_at IS NULL", hash).
		Update("revoked_at", now)
	if result.Error != nil {
		r.log.Error("failed to revoke refresh token", "module", "refresh_token_repository", "error", result.Error)
		return result.Error
	}

	return nil
}
