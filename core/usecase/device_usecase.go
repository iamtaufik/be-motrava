package usecase

import (
	"fmt"

	"github.com/google/uuid"

	"motrava/core/dto"
	"motrava/core/models"
	portUsecase "motrava/core/port/usecase"
	"motrava/core/repository"
)

type deviceUsecase struct {
	deviceRepo repository.UserDeviceRepository
}

func NewDeviceUsecase(deviceRepo repository.UserDeviceRepository) portUsecase.DeviceUsecase {
	return &deviceUsecase{deviceRepo: deviceRepo}
}

func (u *deviceUsecase) RegisterDevice(userID string, input dto.RegisterDeviceRequest) error {
	uid, err := uuid.Parse(userID)
	if err != nil {
		return fmt.Errorf("invalid user id: %w", err)
	}

	if input.DeviceToken == "" {
		return fmt.Errorf("device_token is required")
	}

	platform := input.Platform
	if platform == "" {
		platform = "android"
	}

	device := models.UserDevice{
		ID:          uuid.New(),
		UserID:      uid,
		DeviceToken: input.DeviceToken,
		Platform:    platform,
	}

	return u.deviceRepo.Upsert(&device)
}
