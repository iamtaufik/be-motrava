package repository

import (
	"github.com/google/uuid"

	"motrava/core-service/internal/core/models"
)

type VehicleRepository interface {
	FindByUserID(userID uuid.UUID) ([]models.Vehicle, error)
	FindByIDAndUserID(id uuid.UUID, userID uuid.UUID) (*models.Vehicle, error)
	Create(vehicle *models.Vehicle) error
	Save(vehicle *models.Vehicle) error
	Delete(id uuid.UUID, userID uuid.UUID) error
	FindDefaultByUserID(userID uuid.UUID) (*models.Vehicle, error)
}
