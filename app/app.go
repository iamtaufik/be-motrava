package app

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/gofiber/fiber/v2"
	"github.com/redis/go-redis/v9"
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
	wsInfra "motrava/infra/ws"
)

type repositories struct {
	userRepo         repository.UserRepository
	refreshTokenRepo repository.RefreshTokenRepository
	vehicleRepo      repository.VehicleRepository
	tripRepo         repository.TripRepository
	tripPointRepo    repository.TripPointRepository
}

type usecases struct {
	authUsecase    portUsecase.AuthUsecase
	userUsecase    portUsecase.UserUsecase
	vehicleUsecase portUsecase.VehicleUsecase
	tripUsecase    portUsecase.TripUsecase
}

type Application struct {
	fiber *fiber.App
	db    *gorm.DB
	rdb   *redis.Client
	cfg   config.Config
	log   *slog.Logger
}

func New(cfg config.Config, logger *slog.Logger) (*Application, error) {
	db, err := database.Connect(cfg.DatabaseURL, logger)
	if err != nil {
		return nil, fmt.Errorf("connect database: %w", err)
	}

	rdb, err := database.ConnectRedis(cfg.RedisURL, cfg.RedisPassword, cfg.RedisDB, logger)
	if err != nil {
		return nil, fmt.Errorf("connect redis: %w", err)
	}

	fiberApp := fiber.New()
	repos := newRepositories(db, logger)
	usecases := newUsecases(cfg, repos, logger)
	wsHub := wsInfra.NewHub(logger, rdb, repos.tripPointRepo)
	authMiddleware := middleware.AuthMiddleware(cfg, repos.userRepo, logger)

	routes.NewRoutes(fiberApp, logger, routes.Handlers{
		AuthHandler:    handlers.NewAuthHandler(usecases.authUsecase, logger),
		UserHandler:    handlers.NewUserHandler(usecases.userUsecase, logger),
		VehicleHandler: handlers.NewVehicleHandler(usecases.vehicleUsecase, logger),
		TripHandler:    handlers.NewTripHandler(usecases.tripUsecase, logger),
		WSHandler:      handlers.NewWSHandler(usecases.tripUsecase, wsHub, rdb, logger, cfg.JWTSecret, cfg.JWTIssuer),
	}, authMiddleware).SetupRouters()

	logger.Info("application modules wired", "module", "app")

	return &Application{
		fiber: fiberApp,
		db:    db,
		rdb:   rdb,
		cfg:   cfg,
		log:   logger,
	}, nil
}

func newRepositories(db *gorm.DB, logger *slog.Logger) *repositories {
	return &repositories{
		userRepo:         infraRepo.NewUserRepositoryGorm(db, logger),
		refreshTokenRepo: infraRepo.NewRefreshTokenRepositoryGorm(db, logger),
		vehicleRepo:      infraRepo.NewVehicleRepositoryGorm(db, logger),
		tripRepo:         infraRepo.NewTripRepositoryGorm(db, logger),
		tripPointRepo:    infraRepo.NewTripPointRepositoryGorm(db, logger),
	}
}

func newUsecases(cfg config.Config, repos *repositories, logger *slog.Logger) *usecases {
	return &usecases{
		authUsecase:    usecase.NewAuthUsecase(cfg, repos.userRepo, repos.refreshTokenRepo, logger),
		userUsecase:    usecase.NewUserUsecase(repos.userRepo),
		vehicleUsecase: usecase.NewVehicleUsecase(repos.vehicleRepo),
		tripUsecase:    usecase.NewTripUsecase(repos.tripRepo, repos.tripPointRepo, repos.vehicleRepo),
	}
}

func (a *Application) Run() error {
	a.log.Info("starting fiber server", "module", "app", "port", a.cfg.AppPort)
	return a.fiber.Listen(":" + a.cfg.AppPort)
}

func (a *Application) Shutdown(ctx context.Context) error {
	a.log.Info("shutting down fiber server", "module", "app")
	if err := a.rdb.Close(); err != nil {
		a.log.Error("redis close error", "module", "app", "error", err)
	}
	return a.fiber.ShutdownWithContext(ctx)
}
