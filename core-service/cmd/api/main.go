package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"motrava/core-service/internal/app"
	"motrava/core-service/internal/config"
	"motrava/core-service/internal/infra/logger"
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
		appLogger.Error("failed to initialize core service", "error", err)
		os.Exit(1)
	}

	appLogger.Info("core service initialized", "module", "main")

	go func() {
		if err := application.Run(); err != nil {
			appLogger.Error("core server stopped with error", "error", err, "module", "main")
			os.Exit(1)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := application.Shutdown(ctx); err != nil {
		appLogger.Error("graceful shutdown failed", "error", err, "module", "main")
	}

	appLogger.Info("core service shutdown complete", "module", "main")
}
