package app

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"

	"motrava/core-service/internal/app/handlers"
	"motrava/core-service/internal/app/middleware"
	"motrava/core-service/internal/app/routes"
	"motrava/core-service/internal/config"
	"motrava/core-service/internal/core/repository"
	"motrava/core-service/internal/core/usecase"
	"motrava/core-service/internal/infra/database"
	fcmInfra "motrava/core-service/internal/infra/fcm"
	"motrava/core-service/internal/infra/iamclient"
	infraRepo "motrava/core-service/internal/infra/repository"
	wsInfra "motrava/core-service/internal/infra/ws"
	portUsecase "motrava/core-service/internal/core/port/usecase"
)

type Application struct {
	fiber     *fiber.App
	rdb       *redis.Client
	cfg       config.Config
	log       *slog.Logger
	scheduler *usecase.ReminderScheduler
}

type repositories struct {
	vehicleRepo           repository.VehicleRepository
	tripRepo              repository.TripRepository
	tripPointRepo         repository.TripPointRepository
	serviceReminderRepo   repository.ServiceReminderRepository
	manualDistanceLogRepo repository.ManualDistanceLogRepository
	userDeviceRepo        repository.UserDeviceRepository
}

type usecases struct {
	vehicleUsecase         portUsecase.VehicleUsecase
	tripUsecase            portUsecase.TripUsecase
	serviceReminderUsecase portUsecase.ServiceReminderUsecase
	deviceUsecase          portUsecase.DeviceUsecase
	reminderNotifier       *usecase.ReminderNotifier
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

	iamClient, err := iamclient.New(cfg.IAMServiceAddress, logger)
	if err != nil {
		return nil, fmt.Errorf("connect iam gRPC: %w", err)
	}

	repos := newRepositories(db, logger)
	useCases := newUsecases(cfg, repos, logger)
	wsHub := wsInfra.NewHub(logger, rdb, repos.tripPointRepo)
	authMiddleware := middleware.CoreAuthMiddleware(cfg, iamClient, logger)

	scheduler := usecase.StartReminderScheduler(30*time.Second, useCases.reminderNotifier, repos.serviceReminderRepo)

	routes.NewRoutes(fiberApp, logger, routes.Handlers{
		VehicleHandler:         handlers.NewVehicleHandler(useCases.vehicleUsecase, logger),
		TripHandler:            handlers.NewTripHandler(useCases.tripUsecase, logger),
		WSHandler:              handlers.NewWSHandler(useCases.tripUsecase, wsHub, rdb, logger, cfg.JWTSecret, cfg.JWTIssuer),
		ServiceReminderHandler: handlers.NewServiceReminderHandler(useCases.serviceReminderUsecase, logger),
		DeviceHandler:          handlers.NewDeviceHandler(useCases.deviceUsecase, logger),
	}, authMiddleware).SetupRouters()

	logger.Info("core application modules wired", "module", "core_app")

	return &Application{
		fiber:     fiberApp,
		rdb:       rdb,
		cfg:       cfg,
		log:       logger,
		scheduler: scheduler,
	}, nil
}

func newRepositories(db *gorm.DB, logger *slog.Logger) *repositories {
	return &repositories{
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
		logger.Warn("fcm client not available, using no-op", "module", "core_app", "error", err)
		fcmClient = fcmInfra.NewNoopFCMClient(logger)
	}

	reminderNotifier := usecase.NewReminderNotifier(fcmClient, repos.userDeviceRepo)

	return &usecases{
		vehicleUsecase:         usecase.NewVehicleUsecase(repos.vehicleRepo, repos.tripRepo, repos.serviceReminderRepo),
		tripUsecase:            usecase.NewTripUsecase(repos.tripRepo, repos.tripPointRepo, repos.vehicleRepo, repos.serviceReminderRepo, reminderNotifier),
		serviceReminderUsecase: usecase.NewServiceReminderUsecase(repos.serviceReminderRepo, repos.manualDistanceLogRepo, repos.vehicleRepo, reminderNotifier),
		deviceUsecase:          usecase.NewDeviceUsecase(repos.userDeviceRepo),
		reminderNotifier:       reminderNotifier,
	}
}

func (a *Application) Run() error {
	a.log.Info("starting core fiber server", "module", "core_app", "port", a.cfg.AppPort)
	return a.fiber.Listen(":" + a.cfg.AppPort)
}

func (a *Application) Shutdown(ctx context.Context) error {
	a.log.Info("shutting down core service", "module", "core_app")

	if a.scheduler != nil {
		a.scheduler.Stop()
	}

	if err := a.rdb.Close(); err != nil {
		a.log.Error("redis close error", "module", "core_app", "error", err)
	}
	return a.fiber.ShutdownWithContext(ctx)
}
