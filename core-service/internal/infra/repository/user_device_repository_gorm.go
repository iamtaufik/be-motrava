package repository

import (
	"log/slog"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"motrava/core-service/internal/core/models"
	coreRepo "motrava/core-service/internal/core/repository"
)

type userDeviceRepositoryGorm struct {
	db  *gorm.DB
	log *slog.Logger
}

func NewUserDeviceRepositoryGorm(db *gorm.DB, logger *slog.Logger) coreRepo.UserDeviceRepository {
	return &userDeviceRepositoryGorm{db: db, log: logger}
}

func (r *userDeviceRepositoryGorm) FindByUserID(userID uuid.UUID) ([]models.UserDevice, error) {
	var devices []models.UserDevice
	if err := r.db.Where("user_id = ?", userID).Find(&devices).Error; err != nil {
		r.log.Error("failed to query devices by user", "module", "user_device_repository", "error", err, "user_id", userID)
		return nil, err
	}
	return devices, nil
}

func (r *userDeviceRepositoryGorm) FindByDeviceToken(token string) (*models.UserDevice, error) {
	var device models.UserDevice
	if err := r.db.Where("device_token = ?", token).First(&device).Error; err != nil {
		return nil, err
	}
	return &device, nil
}

func (r *userDeviceRepositoryGorm) Upsert(device *models.UserDevice) error {
	if err := r.db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "device_token"}},
		DoUpdates: clause.AssignmentColumns([]string{"user_id", "platform", "updated_at"}),
	}).Create(device).Error; err != nil {
		r.log.Error("failed to upsert device", "module", "user_device_repository", "error", err)
		return err
	}
	return nil
}

func (r *userDeviceRepositoryGorm) DeleteByUserID(userID uuid.UUID) error {
	if err := r.db.Where("user_id = ?", userID).Delete(&models.UserDevice{}).Error; err != nil {
		r.log.Error("failed to delete devices by user", "module", "user_device_repository", "error", err)
		return err
	}
	return nil
}

func (r *userDeviceRepositoryGorm) Delete(id uuid.UUID) error {
	if err := r.db.Delete(&models.UserDevice{}, id).Error; err != nil {
		r.log.Error("failed to delete device", "module", "user_device_repository", "error", err)
		return err
	}
	return nil
}
