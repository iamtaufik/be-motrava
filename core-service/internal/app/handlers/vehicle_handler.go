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

type VehicleHandler struct {
	vehicleUsecase portUsecase.VehicleUsecase
	log            *slog.Logger
}

func NewVehicleHandler(vehicleUsecase portUsecase.VehicleUsecase, logger *slog.Logger) *VehicleHandler {
	return &VehicleHandler{vehicleUsecase: vehicleUsecase, log: logger}
}

func (h *VehicleHandler) CreateVehicle(c *fiber.Ctx) error {
	user, ok := middleware.CurrentUser(c)
	if !ok {
		return response.Error(c, fiber.StatusUnauthorized, "unauthorized", nil)
	}

	var req dto.CreateVehicleRequest
	if err := c.BodyParser(&req); err != nil {
		return response.ValidationError(c, fiber.Map{"body": "invalid json payload"})
	}

	req.VehicleName = strings.TrimSpace(req.VehicleName)
	req.PlateNumber = strings.TrimSpace(req.PlateNumber)
	req.Brand = strings.TrimSpace(req.Brand)
	req.Model = strings.TrimSpace(req.Model)
	req.Color = strings.TrimSpace(req.Color)

	if req.VehicleName == "" || req.PlateNumber == "" || req.Brand == "" || req.Model == "" || req.Color == "" {
		return response.ValidationError(c, fiber.Map{
			"vehicle_name": "required",
			"plate_number": "required",
			"brand":        "required",
			"model":        "required",
			"color":        "required",
		})
	}

	created, err := h.vehicleUsecase.CreateVehicle(user.ID.String(), req)
	if err != nil {
		h.log.Error("create vehicle failed", "module", "vehicle_handler", "error", err)
		return response.Error(c, fiber.StatusInternalServerError, "failed to create vehicle", nil)
	}

	return response.Created(c, "vehicle created", created)
}

func (h *VehicleHandler) ListVehicles(c *fiber.Ctx) error {
	user, ok := middleware.CurrentUser(c)
	if !ok {
		return response.Error(c, fiber.StatusUnauthorized, "unauthorized", nil)
	}

	vehicles, err := h.vehicleUsecase.ListVehicles(user.ID.String())
	if err != nil {
		h.log.Error("list vehicles failed", "module", "vehicle_handler", "error", err)
		return response.Error(c, fiber.StatusInternalServerError, "failed to fetch vehicles", nil)
	}

	return response.OK(c, "vehicles fetched", vehicles)
}

func (h *VehicleHandler) GetVehicleByID(c *fiber.Ctx) error {
	user, ok := middleware.CurrentUser(c)
	if !ok {
		return response.Error(c, fiber.StatusUnauthorized, "unauthorized", nil)
	}

	vehicle, err := h.vehicleUsecase.GetVehicleByID(c.Params("id"), user.ID.String())
	if err != nil {
		if strings.Contains(err.Error(), "invalid vehicle id") {
			return response.ValidationError(c, fiber.Map{"id": "must be a valid UUID"})
		}
		h.log.Error("get vehicle by id failed", "module", "vehicle_handler", "error", err)
		return response.Error(c, fiber.StatusInternalServerError, "failed to fetch vehicle", nil)
	}

	if vehicle == nil {
		return response.Error(c, fiber.StatusNotFound, "vehicle not found", nil)
	}

	return response.OK(c, "vehicle fetched", vehicle)
}

func (h *VehicleHandler) UpdateVehicle(c *fiber.Ctx) error {
	user, ok := middleware.CurrentUser(c)
	if !ok {
		return response.Error(c, fiber.StatusUnauthorized, "unauthorized", nil)
	}

	var req dto.UpdateVehicleRequest
	if err := c.BodyParser(&req); err != nil {
		return response.ValidationError(c, fiber.Map{"body": "invalid json payload"})
	}

	updated, err := h.vehicleUsecase.UpdateVehicle(c.Params("id"), user.ID.String(), req)
	if err != nil {
		if strings.Contains(err.Error(), "invalid vehicle id") {
			return response.ValidationError(c, fiber.Map{"id": "must be a valid UUID"})
		}
		h.log.Error("update vehicle failed", "module", "vehicle_handler", "error", err)
		return response.Error(c, fiber.StatusInternalServerError, "failed to update vehicle", nil)
	}

	if updated == nil {
		return response.Error(c, fiber.StatusNotFound, "vehicle not found", nil)
	}

	return response.OK(c, "vehicle updated", updated)
}

func (h *VehicleHandler) DeleteVehicle(c *fiber.Ctx) error {
	user, ok := middleware.CurrentUser(c)
	if !ok {
		return response.Error(c, fiber.StatusUnauthorized, "unauthorized", nil)
	}

	err := h.vehicleUsecase.DeleteVehicle(c.Params("id"), user.ID.String())
	if err != nil {
		if strings.Contains(err.Error(), "invalid vehicle id") {
			return response.ValidationError(c, fiber.Map{"id": "must be a valid UUID"})
		}
		if err.Error() == "record not found" {
			return response.Error(c, fiber.StatusNotFound, "vehicle not found", nil)
		}
		h.log.Error("delete vehicle failed", "module", "vehicle_handler", "error", err)
		return response.Error(c, fiber.StatusInternalServerError, "failed to delete vehicle", nil)
	}

	return response.OK(c, "vehicle deleted", nil)
}

func (h *VehicleHandler) SetDefaultVehicle(c *fiber.Ctx) error {
	user, ok := middleware.CurrentUser(c)
	if !ok {
		return response.Error(c, fiber.StatusUnauthorized, "unauthorized", nil)
	}

	vehicle, err := h.vehicleUsecase.SetDefaultVehicle(c.Params("id"), user.ID.String())
	if err != nil {
		if strings.Contains(err.Error(), "invalid vehicle id") {
			return response.ValidationError(c, fiber.Map{"id": "must be a valid UUID"})
		}
		h.log.Error("set default vehicle failed", "module", "vehicle_handler", "error", err)
		return response.Error(c, fiber.StatusInternalServerError, "failed to set default vehicle", nil)
	}

	if vehicle == nil {
		return response.Error(c, fiber.StatusNotFound, "vehicle not found", nil)
	}

	return response.OK(c, "default vehicle updated", vehicle)
}
