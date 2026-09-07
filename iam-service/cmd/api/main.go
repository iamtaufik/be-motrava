package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"motrava/iam-service/internal/app"
	"motrava/iam-service/internal/config"
	"motrava/iam-service/internal/infra/logger"
)

func main() {
	cfg := config.Load()
	appLogger, logFile, err := logger.NewJSONFileLogger(cfg.LogFile)
	if err != nil {
		panic(fmt.Errorf("failed to initialize logger: %w", err))
	}
	defer logFile.Close()

	application, err := app.New(cfg, appLogger)
	if err != nil {
		appLogger.Error("failed to initialize iam service", "error", err)
		os.Exit(1)
	}

	appLogger.Info("iam service initialized", "module", "main")

	serverErr := make(chan error, 1)
	go func() {
		serverErr <- application.Run()
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)

	select {
	case err := <-serverErr:
		if err != nil {
			appLogger.Error("iam server stopped with error", "error", err, "module", "main")
			os.Exit(1)
		}
		appLogger.Info("iam service shutdown complete", "module", "main")
		return
	case <-quit:
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := application.Shutdown(ctx); err != nil {
		appLogger.Error("graceful shutdown failed", "error", err, "module", "main")
	}

	appLogger.Info("iam service shutdown complete", "module", "main")
}
