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
	"motrava/core/repository"
	"motrava/core/usecase"
	portUsecase "motrava/core/port/usecase"
	"motrava/infra/database"
	infraRepo "motrava/infra/repository"
)

type repositories struct {
	userRepo         repository.UserRepository
	refreshTokenRepo repository.RefreshTokenRepository
	vehicleRepo      repository.VehicleRepository
}

type usecases struct {
	authUsecase   portUsecase.AuthUsecase
	userUsecase   portUsecase.UserUsecase
	vehicleUsecase portUsecase.VehicleUsecase
}

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

	fiberApp := fiber.New()
	repos := newRepositories(db, logger)
	usecases := newUsecases(cfg, repos, logger)
	authMiddleware := middleware.AuthMiddleware(cfg, repos.userRepo, logger)

	routes.NewRoutes(fiberApp, logger, routes.Handlers{
		AuthHandler:    handlers.NewAuthHandler(usecases.authUsecase, logger),
		UserHandler:    handlers.NewUserHandler(usecases.userUsecase, logger),
		VehicleHandler: handlers.NewVehicleHandler(usecases.vehicleUsecase, logger),
	}, authMiddleware).SetupRouters()

	logger.Info("application modules wired", "module", "app")

	return &Application{
		fiber: fiberApp,
		db:    db,
		cfg:   cfg,
		log:   logger,
	}, nil
}

func newRepositories(db *gorm.DB, logger *slog.Logger) *repositories {
	return &repositories{
		userRepo:         infraRepo.NewUserRepositoryGorm(db, logger),
		refreshTokenRepo: infraRepo.NewRefreshTokenRepositoryGorm(db, logger),
		vehicleRepo:      infraRepo.NewVehicleRepositoryGorm(db, logger),
	}
}

func newUsecases(cfg config.Config, repos *repositories, logger *slog.Logger) *usecases {
	return &usecases{
		authUsecase:    usecase.NewAuthUsecase(cfg, repos.userRepo, repos.refreshTokenRepo, logger),
		userUsecase:    usecase.NewUserUsecase(repos.userRepo),
		vehicleUsecase: usecase.NewVehicleUsecase(repos.vehicleRepo),
	}
}

func (a *Application) Run() error {
	a.log.Info("starting fiber server", "module", "app", "port", a.cfg.AppPort)
	return a.fiber.Listen(":" + a.cfg.AppPort)
}

func (a *Application) Shutdown(ctx context.Context) error {
	a.log.Info("shutting down fiber server", "module", "app")
	return a.fiber.ShutdownWithContext(ctx)
}
