package routes

import (
	"log/slog"
	"time"

	"github.com/gofiber/fiber/v2"

	"motrava/app/handlers"
	"motrava/core/utils/response"
)

type Handlers struct {
	AuthHandler *handlers.AuthHandler
	UserHandler *handlers.UserHandler
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
		err := c.Next()
		r.logger.Info("http request",
			"module", "routes",
			"method", c.Method(),
			"path", c.Path(),
			"status", c.Response().StatusCode(),
			"latency_ms", time.Since(start).Milliseconds(),
		)
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
}
