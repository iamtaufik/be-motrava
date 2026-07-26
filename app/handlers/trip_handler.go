package handlers

import (
	"log/slog"
	"strconv"
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

func (h *TripHandler) GetTripDetail(c *fiber.Ctx) error {
	user, ok := middleware.CurrentUser(c)
	if !ok {
		return response.Error(c, fiber.StatusUnauthorized, "unauthorized", nil)
	}

	trip, err := h.tripUsecase.GetTripDetail(user.ID.String(), c.Params("id"))
	if err != nil {
		h.log.Error("get trip detail failed", "module", "trip_handler", "error", err)
		msg := err.Error()
		if strings.Contains(msg, "trip not found") {
			return response.Error(c, fiber.StatusNotFound, msg, nil)
		}
		if strings.Contains(msg, "invalid trip id") {
			return response.ValidationError(c, fiber.Map{"id": "must be a valid UUID"})
		}
		return response.Error(c, fiber.StatusInternalServerError, "failed to get trip detail", nil)
	}

	return response.OK(c, "trip detail retrieved", trip)
}

func (h *TripHandler) GetTripHistory(c *fiber.Ctx) error {
	user, ok := middleware.CurrentUser(c)
	if !ok {
		return response.Error(c, fiber.StatusUnauthorized, "unauthorized", nil)
	}

	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "10"))

	req := dto.TripHistoryRequest{
		Page:     page,
		Limit:    limit,
		Search:   c.Query("search"),
		DateFrom: c.Query("date_from"),
		DateTo:   c.Query("date_to"),
	}

	trips, meta, err := h.tripUsecase.GetTripHistory(user.ID.String(), req)
	if err != nil {
		h.log.Error("get trip history failed", "module", "trip_handler", "error", err)
		return response.Error(c, fiber.StatusInternalServerError, "failed to get trip history", nil)
	}

	return c.Status(fiber.StatusOK).JSON(response.APIResponse{
		Success: true,
		Message: "trip history retrieved",
		Data:    trips,
		Meta:    meta,
	})
}

func (h *TripHandler) DeleteTrip(c *fiber.Ctx) error {
	user, ok := middleware.CurrentUser(c)
	if !ok {
		return response.Error(c, fiber.StatusUnauthorized, "unauthorized", nil)
	}

	err := h.tripUsecase.DeleteTrip(user.ID.String(), c.Params("id"))
	if err != nil {
		h.log.Error("delete trip failed", "module", "trip_handler", "error", err)
		msg := err.Error()
		if strings.Contains(msg, "trip not found") {
			return response.Error(c, fiber.StatusNotFound, msg, nil)
		}
		if strings.Contains(msg, "invalid trip id") {
			return response.ValidationError(c, fiber.Map{"id": "must be a valid UUID"})
		}
		return response.Error(c, fiber.StatusInternalServerError, "failed to delete trip", nil)
	}

	return response.OK(c, "trip deleted", nil)
}
