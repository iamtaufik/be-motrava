package usecase

import "motrava/core/dto"

type DeviceUsecase interface {
	RegisterDevice(userID string, input dto.RegisterDeviceRequest) error
}
