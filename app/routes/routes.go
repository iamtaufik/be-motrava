package routes

import (
	"log/slog"
	"time"

	"github.com/gofiber/fiber/v2"

	"motrava/app/handlers"
	"motrava/core/utils/response"
)

type Handlers struct {
	AuthHandler            *handlers.AuthHandler
	UserHandler            *handlers.UserHandler
	VehicleHandler         *handlers.VehicleHandler
	TripHandler            *handlers.TripHandler
	WSHandler              *handlers.WSHandler
	ServiceReminderHandler *handlers.ServiceReminderHandler
	DeviceHandler          *handlers.DeviceHandler
}

type Router struct {
	app            *fiber.App
	logger         *slog.Logger
	handlers       Handlers
	authMiddleware fiber.Handler
}

func NewRoutes(app *fiber.App, logger *slog.Logger, handlers Handlers, authMiddleware fiber.Handler) *Router {
	return &Router{
		app:            app,
		logger:         logger,
		handlers:       handlers,
		authMiddleware: authMiddleware,
	}
}

func (r *Router) SetupRouters() {
	r.app.Use(func(c *fiber.Ctx) error {
		start := time.Now()
		body := c.Body()
		err := c.Next()
		attrs := []any{
			"module", "routes",
			"method", c.Method(),
			"path", c.Path(),
			"status", c.Response().StatusCode(),
			"latency_ms", time.Since(start).Milliseconds(),
		}
		if len(body) > 0 {
			attrs = append(attrs, "body", string(body))
		}
		r.logger.Info("http request", attrs...)
		return err
	})

	api := r.app.Group("/api")

	api.Get("/health", func(c *fiber.Ctx) error {
		r.logger.Info("health endpoint called", "module", "routes")
		return response.OK(c, "health check success", fiber.Map{
			"status": "ok",
		})
	})

	auth := api.Group("/auth")
	auth.Post("/register", r.handlers.AuthHandler.Register)
	auth.Post("/login", r.handlers.AuthHandler.Login)
	auth.Post("/refresh", r.handlers.AuthHandler.Refresh)
	auth.Get("/me", r.authMiddleware, r.handlers.AuthHandler.Me)
	auth.Get("/google/login", r.handlers.AuthHandler.GoogleLogin)
	auth.Get("/google/callback", r.handlers.AuthHandler.GoogleCallback)
	auth.Post("/google/mobile", r.handlers.AuthHandler.GoogleMobileLogin)

	vehicle := api.Group("/vehicles", r.authMiddleware)
	vehicle.Post("/", r.handlers.VehicleHandler.CreateVehicle)
	vehicle.Get("/", r.handlers.VehicleHandler.ListVehicles)
	vehicle.Get("/:id", r.handlers.VehicleHandler.GetVehicleByID)
	vehicle.Put("/:id", r.handlers.VehicleHandler.UpdateVehicle)
	vehicle.Delete("/:id", r.handlers.VehicleHandler.DeleteVehicle)
	vehicle.Put("/:id/default", r.handlers.VehicleHandler.SetDefaultVehicle)

	trip := api.Group("/trips", r.authMiddleware)
	trip.Get("/", r.handlers.TripHandler.GetTripHistory)
	trip.Get("/:id", r.handlers.TripHandler.GetTripDetail)
	trip.Post("/start", r.handlers.TripHandler.StartTrip)
	trip.Post("/:id/end", r.handlers.TripHandler.EndTrip)
	trip.Delete("/:id", r.handlers.TripHandler.DeleteTrip)

	api.Get("/ws/trip/location", r.handlers.WSHandler.Upgrade)

	device := api.Group("/devices", r.authMiddleware)
	device.Post("/register", r.handlers.DeviceHandler.RegisterDevice)

	serviceReminder := vehicle.Group("/:vehicleId/service-reminders")
	serviceReminder.Post("/", r.handlers.ServiceReminderHandler.CreateReminder)
	serviceReminder.Get("/", r.handlers.ServiceReminderHandler.ListReminders)
	serviceReminder.Get("/:reminderId/progress", r.handlers.ServiceReminderHandler.GetReminderProgress)
	serviceReminder.Put("/:reminderId", r.handlers.ServiceReminderHandler.UpdateReminder)
	serviceReminder.Delete("/:reminderId", r.handlers.ServiceReminderHandler.DeleteReminder)
	serviceReminder.Post("/:reminderId/reset", r.handlers.ServiceReminderHandler.ResetReminder)
	serviceReminder.Post("/:reminderId/manual-distance", r.handlers.ServiceReminderHandler.AddManualDistance)
}
