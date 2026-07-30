package handlers

import (
	"log/slog"
	"strings"

	"github.com/gofiber/fiber/v2"

	"motrava/core-service/internal/app/middleware"
	"motrava/core-service/internal/core/dto"
	portUsecase "motrava/core-service/internal/core/port/usecase"
	"motrava/core-service/internal/core/utils/response"
)

type ServiceReminderHandler struct {
	reminderUsecase portUsecase.ServiceReminderUsecase
	log             *slog.Logger
}

func NewServiceReminderHandler(reminderUsecase portUsecase.ServiceReminderUsecase, logger *slog.Logger) *ServiceReminderHandler {
	return &ServiceReminderHandler{reminderUsecase: reminderUsecase, log: logger}
}

func (h *ServiceReminderHandler) CreateReminder(c *fiber.Ctx) error {
	user, ok := middleware.CurrentUser(c)
	if !ok {
		return response.Error(c, fiber.StatusUnauthorized, "unauthorized", nil)
	}

	var req dto.CreateServiceReminderRequest
	if err := c.BodyParser(&req); err != nil {
		return response.ValidationError(c, fiber.Map{"body": "invalid json payload"})
	}

	req.ServiceName = strings.TrimSpace(req.ServiceName)
	if req.ServiceName == "" {
		return response.ValidationError(c, fiber.Map{"service_name": "required"})
	}
	if req.IntervalKM <= 0 {
		return response.ValidationError(c, fiber.Map{"interval_km": "must be greater than 0"})
	}

	reminder, err := h.reminderUsecase.CreateReminder(user.ID.String(), c.Params("vehicleId"), req)
	if err != nil {
		h.log.Error("create reminder failed", "module", "service_reminder_handler", "error", err)
		msg := err.Error()
		if strings.Contains(msg, "vehicle not found") {
			return response.Error(c, fiber.StatusNotFound, msg, nil)
		}
		if strings.Contains(msg, "invalid user id") || strings.Contains(msg, "invalid vehicle id") {
			return response.ValidationError(c, fiber.Map{"vehicle_id": "must be a valid UUID"})
		}
		return response.Error(c, fiber.StatusInternalServerError, "failed to create reminder", nil)
	}

	return response.Created(c, "service reminder created", reminder)
}

func (h *ServiceReminderHandler) GetReminderProgress(c *fiber.Ctx) error {
	user, ok := middleware.CurrentUser(c)
	if !ok {
		return response.Error(c, fiber.StatusUnauthorized, "unauthorized", nil)
	}

	reminder, err := h.reminderUsecase.GetReminderProgress(user.ID.String(), c.Params("vehicleId"), c.Params("reminderId"))
	if err != nil {
		h.log.Error("get reminder progress failed", "module", "service_reminder_handler", "error", err)
		msg := err.Error()
		if strings.Contains(msg, "not found") {
			return response.Error(c, fiber.StatusNotFound, msg, nil)
		}
		if strings.Contains(msg, "invalid") {
			return response.ValidationError(c, fiber.Map{"id": "must be a valid UUID"})
		}
		return response.Error(c, fiber.StatusInternalServerError, "failed to get reminder progress", nil)
	}

	return response.OK(c, "service reminder progress retrieved", reminder)
}

func (h *ServiceReminderHandler) ListReminders(c *fiber.Ctx) error {
	user, ok := middleware.CurrentUser(c)
	if !ok {
		return response.Error(c, fiber.StatusUnauthorized, "unauthorized", nil)
	}

	reminders, err := h.reminderUsecase.ListRemindersByVehicle(user.ID.String(), c.Params("vehicleId"))
	if err != nil {
		h.log.Error("list reminders failed", "module", "service_reminder_handler", "error", err)
		msg := err.Error()
		if strings.Contains(msg, "vehicle not found") {
			return response.Error(c, fiber.StatusNotFound, msg, nil)
		}
		if strings.Contains(msg, "invalid") {
			return response.ValidationError(c, fiber.Map{"vehicle_id": "must be a valid UUID"})
		}
		return response.Error(c, fiber.StatusInternalServerError, "failed to list reminders", nil)
	}

	return response.OK(c, "service reminders retrieved", reminders)
}

func (h *ServiceReminderHandler) UpdateReminder(c *fiber.Ctx) error {
	user, ok := middleware.CurrentUser(c)
	if !ok {
		return response.Error(c, fiber.StatusUnauthorized, "unauthorized", nil)
	}

	var req dto.UpdateServiceReminderRequest
	if err := c.BodyParser(&req); err != nil {
		return response.ValidationError(c, fiber.Map{"body": "invalid json payload"})
	}

	reminder, err := h.reminderUsecase.UpdateReminder(user.ID.String(), c.Params("vehicleId"), c.Params("reminderId"), req)
	if err != nil {
		h.log.Error("update reminder failed", "module", "service_reminder_handler", "error", err)
		msg := err.Error()
		if strings.Contains(msg, "not found") {
			return response.Error(c, fiber.StatusNotFound, msg, nil)
		}
		if strings.Contains(msg, "invalid") || strings.Contains(msg, "must be") {
			return response.ValidationError(c, fiber.Map{"message": msg})
		}
		return response.Error(c, fiber.StatusInternalServerError, "failed to update reminder", nil)
	}

	return response.OK(c, "service reminder updated", reminder)
}

func (h *ServiceReminderHandler) DeleteReminder(c *fiber.Ctx) error {
	user, ok := middleware.CurrentUser(c)
	if !ok {
		return response.Error(c, fiber.StatusUnauthorized, "unauthorized", nil)
	}

	err := h.reminderUsecase.DeleteReminder(user.ID.String(), c.Params("vehicleId"), c.Params("reminderId"))
	if err != nil {
		h.log.Error("delete reminder failed", "module", "service_reminder_handler", "error", err)
		msg := err.Error()
		if strings.Contains(msg, "not found") {
			return response.Error(c, fiber.StatusNotFound, msg, nil)
		}
		if strings.Contains(msg, "invalid") {
			return response.ValidationError(c, fiber.Map{"id": "must be a valid UUID"})
		}
		return response.Error(c, fiber.StatusInternalServerError, "failed to delete reminder", nil)
	}

	return response.OK(c, "service reminder deleted", nil)
}

func (h *ServiceReminderHandler) ResetReminder(c *fiber.Ctx) error {
	user, ok := middleware.CurrentUser(c)
	if !ok {
		return response.Error(c, fiber.StatusUnauthorized, "unauthorized", nil)
	}

	reminder, err := h.reminderUsecase.ResetReminder(user.ID.String(), c.Params("vehicleId"), c.Params("reminderId"))
	if err != nil {
		h.log.Error("reset reminder failed", "module", "service_reminder_handler", "error", err)
		msg := err.Error()
		if strings.Contains(msg, "not found") {
			return response.Error(c, fiber.StatusNotFound, msg, nil)
		}
		if strings.Contains(msg, "invalid") {
			return response.ValidationError(c, fiber.Map{"id": "must be a valid UUID"})
		}
		return response.Error(c, fiber.StatusInternalServerError, "failed to reset reminder", nil)
	}

	return response.OK(c, "service reminder reset", reminder)
}

func (h *ServiceReminderHandler) AddManualDistance(c *fiber.Ctx) error {
	user, ok := middleware.CurrentUser(c)
	if !ok {
		return response.Error(c, fiber.StatusUnauthorized, "unauthorized", nil)
	}

	var req dto.ManualDistanceRequest
	if err := c.BodyParser(&req); err != nil {
		return response.ValidationError(c, fiber.Map{"body": "invalid json payload"})
	}

	if req.DistanceKM <= 0 {
		return response.ValidationError(c, fiber.Map{"distance_km": "must be greater than 0"})
	}

	reminder, err := h.reminderUsecase.AddManualDistance(user.ID.String(), c.Params("vehicleId"), c.Params("reminderId"), req)
	if err != nil {
		h.log.Error("add manual distance failed", "module", "service_reminder_handler", "error", err)
		msg := err.Error()
		if strings.Contains(msg, "not found") {
			return response.Error(c, fiber.StatusNotFound, msg, nil)
		}
		if strings.Contains(msg, "invalid") || strings.Contains(msg, "must be") {
			return response.ValidationError(c, fiber.Map{"message": msg})
		}
		return response.Error(c, fiber.StatusInternalServerError, "failed to add manual distance", nil)
	}

	return response.OK(c, "manual distance added", reminder)
}
