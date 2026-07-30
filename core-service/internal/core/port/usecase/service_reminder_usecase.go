package usecase

import "motrava/core-service/internal/core/dto"

type ServiceReminderUsecase interface {
	CreateReminder(userID string, vehicleID string, input dto.CreateServiceReminderRequest) (*dto.ServiceReminderResponse, error)
	GetReminderProgress(userID string, vehicleID string, reminderID string) (*dto.ServiceReminderResponse, error)
	ListRemindersByVehicle(userID string, vehicleID string) ([]dto.ServiceReminderResponse, error)
	UpdateReminder(userID string, vehicleID string, reminderID string, input dto.UpdateServiceReminderRequest) (*dto.ServiceReminderResponse, error)
	DeleteReminder(userID string, vehicleID string, reminderID string) error
	ResetReminder(userID string, vehicleID string, reminderID string) (*dto.ServiceReminderResponse, error)
	AddManualDistance(userID string, vehicleID string, reminderID string, input dto.ManualDistanceRequest) (*dto.ServiceReminderResponse, error)
}
