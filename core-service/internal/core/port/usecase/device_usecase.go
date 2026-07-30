package usecase

import "motrava/core-service/internal/core/dto"

type DeviceUsecase interface {
	RegisterDevice(userID string, input dto.RegisterDeviceRequest) error
}
