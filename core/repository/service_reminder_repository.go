package repository

import (
	"github.com/google/uuid"

	"motrava/core/models"
)

type ServiceReminderRepository interface {
	FindByID(id uuid.UUID) (*models.ServiceReminder, error)
	FindByVehicleID(vehicleID uuid.UUID) ([]models.ServiceReminder, error)
	FindByUserID(userID uuid.UUID) ([]models.ServiceReminder, error)
	FindAllActive() ([]models.ServiceReminder, error)
	Create(reminder *models.ServiceReminder) error
	Save(reminder *models.ServiceReminder) error
	Delete(id uuid.UUID) error
}

type ManualDistanceLogRepository interface {
	Create(log *models.ManualDistanceLog) error
	FindByReminderID(reminderID uuid.UUID) ([]models.ManualDistanceLog, error)
}
