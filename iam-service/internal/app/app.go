package app

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/gofiber/fiber/v2"
	"github.com/redis/go-redis/v9"

	"motrava/iam-service/internal/app/handlers"
	"motrava/iam-service/internal/app/middleware"
	"motrava/iam-service/internal/app/routes"
	"motrava/iam-service/internal/config"
	"motrava/iam-service/internal/core/usecase"
	"motrava/iam-service/internal/infra/database"
	grpcServer "motrava/iam-service/internal/infra/grpc"
	infraRepo "motrava/iam-service/internal/infra/repository"
)

type Application struct {
	fiber      *fiber.App
	grpcServer *grpcServer.Server
	rdb        *redis.Client
	cfg        config.Config
	log        *slog.Logger
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

	userRepo := infraRepo.NewUserRepositoryGorm(db, logger)
	refreshTokenRepo := infraRepo.NewRefreshTokenRepositoryGorm(db, logger)

	authMiddleware := middleware.AuthMiddleware(cfg, userRepo, logger)

	authUsecase := usecase.NewAuthUsecase(cfg, userRepo, refreshTokenRepo, logger)
	userUsecase := usecase.NewUserUsecase(userRepo)

	routes.NewRoutes(fiberApp, logger, routes.Handlers{
		AuthHandler: handlers.NewAuthHandler(authUsecase, logger),
		UserHandler: handlers.NewUserHandler(userUsecase, logger),
	}, authMiddleware).SetupRouters()

	grpcSrv := grpcServer.NewServer(userRepo, cfg.JWTSecret, cfg.JWTIssuer, logger)

	logger.Info("iam application modules wired", "module", "iam_app")

	return &Application{
		fiber:      fiberApp,
		grpcServer: grpcSrv,
		rdb:        rdb,
		cfg:        cfg,
		log:        logger,
	}, nil
}

func (a *Application) Run() error {
	go func() {
		a.log.Info("starting iam grpc server", "module", "iam_app", "port", a.cfg.GRPCPort)
		if err := a.grpcServer.Run(a.cfg.GRPCPort); err != nil {
			a.log.Error("grpc server stopped with error", "module", "iam_app", "error", err)
		}
	}()

	a.log.Info("starting iam fiber server", "module", "iam_app", "port", a.cfg.AppPort)
	return a.fiber.Listen(":" + a.cfg.AppPort)
}

func (a *Application) Shutdown(ctx context.Context) error {
	a.log.Info("shutting down iam service", "module", "iam_app")
	if err := a.rdb.Close(); err != nil {
		a.log.Error("redis close error", "module", "iam_app", "error", err)
	}
	return a.fiber.ShutdownWithContext(ctx)
}
