package routes

import (
	"log/slog"
	"time"

	"github.com/gofiber/fiber/v2"

	"motrava/app/handlers"
	"motrava/core/utils/response"
)

func Register(app *fiber.App, logger *slog.Logger, userHandler *handlers.UserHandler, authHandler *handlers.AuthHandler) {
	app.Use(func(c *fiber.Ctx) error {
		start := time.Now()
		err := c.Next()
		logger.Info("http request",
			"module", "routes",
			"method", c.Method(),
			"path", c.Path(),
			"status", c.Response().StatusCode(),
			"latency_ms", time.Since(start).Milliseconds(),
		)
		return err
	})

	api := app.Group("/api")

	api.Get("/health", func(c *fiber.Ctx) error {
		logger.Info("health endpoint called", "module", "routes")
		return response.OK(c, "health check success", fiber.Map{
			"status": "ok",
		})
	})

	api.Get("/users", userHandler.ListUsers)
	api.Get("/users/:id", userHandler.GetUserByID)
	api.Post("/users", userHandler.CreateUser)

	auth := api.Group("/auth")
	auth.Post("/register", authHandler.Register)
	auth.Post("/login", authHandler.Login)
	auth.Post("/refresh", authHandler.Refresh)
	auth.Get("/me", authHandler.Me)
	auth.Get("/google/login", authHandler.GoogleLogin)
	auth.Get("/google/callback", authHandler.GoogleCallback)
	auth.Post("/google/mobile", authHandler.GoogleMobileLogin)
}
