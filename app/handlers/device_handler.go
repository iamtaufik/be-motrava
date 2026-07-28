package handlers

import (
	"log/slog"
	"strings"

	"github.com/gofiber/fiber/v2"

	"motrava/app/middleware"
	"motrava/core/dto"
	portUsecase "motrava/core/port/usecase"
	"motrava/core/utils/response"
)

type DeviceHandler struct {
	deviceUsecase portUsecase.DeviceUsecase
	log           *slog.Logger
}

func NewDeviceHandler(deviceUsecase portUsecase.DeviceUsecase, logger *slog.Logger) *DeviceHandler {
	return &DeviceHandler{deviceUsecase: deviceUsecase, log: logger}
}

func (h *DeviceHandler) RegisterDevice(c *fiber.Ctx) error {
	user, ok := middleware.CurrentUser(c)
	if !ok {
		return response.Error(c, fiber.StatusUnauthorized, "unauthorized", nil)
	}

	var req dto.RegisterDeviceRequest
	if err := c.BodyParser(&req); err != nil {
		return response.ValidationError(c, fiber.Map{"body": "invalid json payload"})
	}

	req.DeviceToken = strings.TrimSpace(req.DeviceToken)
	if req.DeviceToken == "" {
		return response.ValidationError(c, fiber.Map{"device_token": "required"})
	}

	if err := h.deviceUsecase.RegisterDevice(user.ID.String(), req); err != nil {
		h.log.Error("register device failed", "module", "device_handler", "error", err)
		return response.Error(c, fiber.StatusInternalServerError, "failed to register device", nil)
	}

	return response.OK(c, "device registered", nil)
}
