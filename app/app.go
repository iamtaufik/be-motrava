package app

import (
	"context"
	"fmt"
	"log/slog"
	"time"

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
	fcmInfra "motrava/infra/fcm"
	infraRepo "motrava/infra/repository"
	wsInfra "motrava/infra/ws"
)

type repositories struct {
	userRepo              repository.UserRepository
	refreshTokenRepo      repository.RefreshTokenRepository
	vehicleRepo           repository.VehicleRepository
	tripRepo              repository.TripRepository
	tripPointRepo         repository.TripPointRepository
	serviceReminderRepo   repository.ServiceReminderRepository
	manualDistanceLogRepo repository.ManualDistanceLogRepository
	userDeviceRepo        repository.UserDeviceRepository
}

type usecases struct {
	authUsecase            portUsecase.AuthUsecase
	userUsecase            portUsecase.UserUsecase
	vehicleUsecase         portUsecase.VehicleUsecase
	tripUsecase            portUsecase.TripUsecase
	serviceReminderUsecase portUsecase.ServiceReminderUsecase
	deviceUsecase          portUsecase.DeviceUsecase
	reminderNotifier       *usecase.ReminderNotifier
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
	useCases := newUsecases(cfg, repos, logger)
	wsHub := wsInfra.NewHub(logger, rdb, repos.tripPointRepo)
	authMiddleware := middleware.AuthMiddleware(cfg, repos.userRepo, logger)

	usecase.StartReminderScheduler(30*time.Minute, useCases.reminderNotifier, repos.serviceReminderRepo)

	routes.NewRoutes(fiberApp, logger, routes.Handlers{
		AuthHandler:            handlers.NewAuthHandler(useCases.authUsecase, logger),
		UserHandler:            handlers.NewUserHandler(useCases.userUsecase, logger),
		VehicleHandler:         handlers.NewVehicleHandler(useCases.vehicleUsecase, logger),
		TripHandler:            handlers.NewTripHandler(useCases.tripUsecase, logger),
		WSHandler:              handlers.NewWSHandler(useCases.tripUsecase, wsHub, rdb, logger, cfg.JWTSecret, cfg.JWTIssuer),
		ServiceReminderHandler: handlers.NewServiceReminderHandler(useCases.serviceReminderUsecase, logger),
		DeviceHandler:          handlers.NewDeviceHandler(useCases.deviceUsecase, logger),
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
		userRepo:              infraRepo.NewUserRepositoryGorm(db, logger),
		refreshTokenRepo:      infraRepo.NewRefreshTokenRepositoryGorm(db, logger),
		vehicleRepo:           infraRepo.NewVehicleRepositoryGorm(db, logger),
		tripRepo:              infraRepo.NewTripRepositoryGorm(db, logger),
		tripPointRepo:         infraRepo.NewTripPointRepositoryGorm(db, logger),
		serviceReminderRepo:   infraRepo.NewServiceReminderRepositoryGorm(db, logger),
		manualDistanceLogRepo: infraRepo.NewManualDistanceLogRepositoryGorm(db, logger),
		userDeviceRepo:        infraRepo.NewUserDeviceRepositoryGorm(db, logger),
	}
}

func newUsecases(cfg config.Config, repos *repositories, logger *slog.Logger) *usecases {
	fcmClient, err := fcmInfra.NewFCMClient(context.Background(), cfg.FCMCredentialsFile, cfg.FCMProjectID, logger)
	if err != nil {
		logger.Warn("fcm client not available, using no-op", "module", "app", "error", err)
		fcmClient = fcmInfra.NewNoopFCMClient(logger)
	}

	reminderNotifier := usecase.NewReminderNotifier(fcmClient, repos.userDeviceRepo)

	return &usecases{
		authUsecase:            usecase.NewAuthUsecase(cfg, repos.userRepo, repos.refreshTokenRepo, logger),
		userUsecase:            usecase.NewUserUsecase(repos.userRepo),
		vehicleUsecase:         usecase.NewVehicleUsecase(repos.vehicleRepo, repos.tripRepo, repos.serviceReminderRepo),
		tripUsecase:            usecase.NewTripUsecase(repos.tripRepo, repos.tripPointRepo, repos.vehicleRepo, repos.serviceReminderRepo, reminderNotifier),
		serviceReminderUsecase: usecase.NewServiceReminderUsecase(repos.serviceReminderRepo, repos.manualDistanceLogRepo, repos.vehicleRepo, reminderNotifier),
		deviceUsecase:          usecase.NewDeviceUsecase(repos.userDeviceRepo),
		reminderNotifier:       reminderNotifier,
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
