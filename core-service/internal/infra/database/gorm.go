package database

import (
	"fmt"
	"log/slog"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"motrava/core-service/internal/core/models"
)

func Connect(dsn string, logger *slog.Logger) (*gorm.DB, error) {
	logger.Info("connecting to database", "module", "database")
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		logger.Error("database connection failed", "module", "database", "error", err)
		return nil, err
	}

	logger.Info("database connected", "module", "database")
	return db, nil
}

func AutoMigrate(db *gorm.DB, logger *slog.Logger) error {
	if err := ensureUUIDExtension(db); err != nil {
		logger.Error("ensure uuid extension failed", "module", "database", "error", err)
		return err
	}

	logger.Info("running database automigrate", "module", "database")
	err := db.AutoMigrate(&models.Vehicle{}, &models.Trip{}, &models.TripPoint{}, &models.ServiceReminder{}, &models.ManualDistanceLog{}, &models.UserDevice{})
	if err != nil {
		logger.Error("database automigrate failed", "module", "database", "error", err)
		return err
	}

	logger.Info("database automigrate completed", "module", "database")
	return nil
}

func ensureUUIDExtension(db *gorm.DB) error {
	if err := db.Exec(`CREATE EXTENSION IF NOT EXISTS "uuid-ossp"`).Error; err != nil {
		return fmt.Errorf("create uuid extension: %w", err)
	}

	return nil
}
