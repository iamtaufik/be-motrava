package repository

import (
	"github.com/google/uuid"

	"motrava/core/models"
)

type UserDeviceRepository interface {
	FindByUserID(userID uuid.UUID) ([]models.UserDevice, error)
	FindByDeviceToken(token string) (*models.UserDevice, error)
	Upsert(device *models.UserDevice) error
	DeleteByUserID(userID uuid.UUID) error
	Delete(id uuid.UUID) error
}
