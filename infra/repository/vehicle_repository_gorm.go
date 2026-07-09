package repository

import (
	"log/slog"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"motrava/core/models"
	coreRepo "motrava/core/repository"
)

type vehicleRepositoryGorm struct {
	db  *gorm.DB
	log *slog.Logger
}

func NewVehicleRepositoryGorm(db *gorm.DB, logger *slog.Logger) coreRepo.VehicleRepository {
	return &vehicleRepositoryGorm{db: db, log: logger}
}

func (r *vehicleRepositoryGorm) FindByUserID(userID uuid.UUID) ([]models.Vehicle, error) {
	var vehicles []models.Vehicle
	if err := r.db.Where("user_id = ?", userID).Order("is_default desc, created_at desc").Find(&vehicles).Error; err != nil {
		r.log.Error("failed to query vehicles by user", "module", "vehicle_repository", "error", err, "user_id", userID)
		return nil, err
	}
	return vehicles, nil
}

func (r *vehicleRepositoryGorm) FindByIDAndUserID(id uuid.UUID, userID uuid.UUID) (*models.Vehicle, error) {
	var vehicle models.Vehicle
	if err := r.db.Where("id = ? AND user_id = ?", id, userID).First(&vehicle).Error; err != nil {
		r.log.Error("failed to query vehicle by id and user", "module", "vehicle_repository", "error", err, "vehicle_id", id, "user_id", userID)
		return nil, err
	}
	return &vehicle, nil
}

func (r *vehicleRepositoryGorm) Create(vehicle *models.Vehicle) error {
	if err := r.db.Create(vehicle).Error; err != nil {
		r.log.Error("failed to create vehicle", "module", "vehicle_repository", "error", err)
		return err
	}
	r.log.Info("vehicle created", "module", "vehicle_repository", "vehicle_id", vehicle.ID)
	return nil
}

func (r *vehicleRepositoryGorm) Save(vehicle *models.Vehicle) error {
	if err := r.db.Save(vehicle).Error; err != nil {
		r.log.Error("failed to save vehicle", "module", "vehicle_repository", "error", err, "vehicle_id", vehicle.ID)
		return err
	}
	r.log.Info("vehicle saved", "module", "vehicle_repository", "vehicle_id", vehicle.ID)
	return nil
}

func (r *vehicleRepositoryGorm) Delete(id uuid.UUID, userID uuid.UUID) error {
	result := r.db.Where("id = ? AND user_id = ?", id, userID).Delete(&models.Vehicle{})
	if result.Error != nil {
		r.log.Error("failed to delete vehicle", "module", "vehicle_repository", "error", result.Error, "vehicle_id", id)
		return result.Error
	}
	if result.RowsAffected == 0 {
		r.log.Warn("vehicle not found for deletion", "module", "vehicle_repository", "vehicle_id", id, "user_id", userID)
		return gorm.ErrRecordNotFound
	}
	r.log.Info("vehicle deleted", "module", "vehicle_repository", "vehicle_id", id)
	return nil
}

func (r *vehicleRepositoryGorm) FindDefaultByUserID(userID uuid.UUID) (*models.Vehicle, error) {
	var vehicle models.Vehicle
	if err := r.db.Where("user_id = ? AND is_default = ?", userID, true).First(&vehicle).Error; err != nil {
		return nil, err
	}
	return &vehicle, nil
}
