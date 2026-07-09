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

type TripHandler struct {
	tripUsecase portUsecase.TripUsecase
	log         *slog.Logger
}

func NewTripHandler(tripUsecase portUsecase.TripUsecase, logger *slog.Logger) *TripHandler {
	return &TripHandler{tripUsecase: tripUsecase, log: logger}
}

func (h *TripHandler) StartTrip(c *fiber.Ctx) error {
	user, ok := middleware.CurrentUser(c)
	if !ok {
		return response.Error(c, fiber.StatusUnauthorized, "unauthorized", nil)
	}

	var req dto.StartTripRequest
	if err := c.BodyParser(&req); err != nil {
		return response.ValidationError(c, fiber.Map{"body": "invalid json payload"})
	}

	req.VehicleID = strings.TrimSpace(req.VehicleID)
	if req.VehicleID == "" {
		return response.ValidationError(c, fiber.Map{"vehicle_id": "required"})
	}

	trip, err := h.tripUsecase.StartTrip(user.ID.String(), req)
	if err != nil {
		h.log.Error("start trip failed", "module", "trip_handler", "error", err)
		msg := err.Error()
		if strings.Contains(msg, "vehicle not found") {
			return response.Error(c, fiber.StatusNotFound, msg, nil)
		}
		if strings.Contains(msg, "ongoing trip") {
			return response.Error(c, fiber.StatusConflict, msg, nil)
		}
		if strings.Contains(msg, "invalid vehicle id") || strings.Contains(msg, "invalid user id") {
			return response.ValidationError(c, fiber.Map{"vehicle_id": "must be a valid UUID"})
		}
		return response.Error(c, fiber.StatusInternalServerError, "failed to start trip", nil)
	}

	return response.Created(c, "trip started", trip)
}

func (h *TripHandler) EndTrip(c *fiber.Ctx) error {
	user, ok := middleware.CurrentUser(c)
	if !ok {
		return response.Error(c, fiber.StatusUnauthorized, "unauthorized", nil)
	}

	trip, err := h.tripUsecase.EndTrip(user.ID.String(), c.Params("id"))
	if err != nil {
		h.log.Error("end trip failed", "module", "trip_handler", "error", err)
		msg := err.Error()
		if strings.Contains(msg, "trip not found") {
			return response.Error(c, fiber.StatusNotFound, msg, nil)
		}
		if strings.Contains(msg, "not ongoing") {
			return response.Error(c, fiber.StatusBadRequest, msg, nil)
		}
		if strings.Contains(msg, "invalid trip id") {
			return response.ValidationError(c, fiber.Map{"id": "must be a valid UUID"})
		}
		return response.Error(c, fiber.StatusInternalServerError, "failed to end trip", nil)
	}

	return response.OK(c, "trip completed", trip)
}
