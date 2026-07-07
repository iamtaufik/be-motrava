package app

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"

	"motrava/app/handlers"
	"motrava/app/middleware"
	"motrava/app/routes"
	"motrava/config"
	"motrava/core/usecase"
	"motrava/infra/database"
	infraRepo "motrava/infra/repository"
)

// Application holds app-level dependencies.
type Application struct {
	fiber *fiber.App
	db    *gorm.DB
	cfg   config.Config
	log   *slog.Logger
}

func New(cfg config.Config, logger *slog.Logger) (*Application, error) {
	db, err := database.Connect(cfg.DatabaseURL, logger)
	if err != nil {
		return nil, fmt.Errorf("connect database: %w", err)
	}

	// if err := database.AutoMigrate(db, logger); err != nil {
	// 	return nil, fmt.Errorf("automigrate models: %w", err)
	// }

	fiberApp := fiber.New()
	userRepository := infraRepo.NewUserRepositoryGorm(db, logger)
	refreshTokenRepository := infraRepo.NewRefreshTokenRepositoryGorm(db, logger)
	authMiddleware := middleware.ValidateToken(cfg, userRepository, logger)
	userUsecase := usecase.NewUserUsecase(userRepository)
	userHandler := handlers.NewUserHandler(userUsecase, logger)
	authUsecase := usecase.NewAuthUsecase(cfg, userRepository, refreshTokenRepository, logger)
	authHandler := handlers.NewAuthHandler(authUsecase, logger)

	routes.Register(fiberApp, logger, authMiddleware, userHandler, authHandler)
	logger.Info("application modules wired", "module", "app")

	return &Application{
		fiber: fiberApp,
		db:    db,
		cfg:   cfg,
		log:   logger,
	}, nil
}

func (a *Application) Run() error {
	a.log.Info("starting fiber server", "module", "app", "port", a.cfg.AppPort)
	return a.fiber.Listen(":" + a.cfg.AppPort)
}

func (a *Application) Shutdown(ctx context.Context) error {
	a.log.Info("shutting down fiber server", "module", "app")
	return a.fiber.ShutdownWithContext(ctx)
}
