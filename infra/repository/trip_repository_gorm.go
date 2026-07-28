package repository

import (
	"log/slog"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"motrava/core/models"
	coreRepo "motrava/core/repository"
)

type tripRepositoryGorm struct {
	db  *gorm.DB
	log *slog.Logger
}

func NewTripRepositoryGorm(db *gorm.DB, logger *slog.Logger) coreRepo.TripRepository {
	return &tripRepositoryGorm{db: db, log: logger}
}

func (r *tripRepositoryGorm) FindByID(id uuid.UUID) (*models.Trip, error) {
	var trip models.Trip
	if err := r.db.Preload("Vehicle").First(&trip, id).Error; err != nil {
		r.log.Error("failed to query trip by id", "module", "trip_repository", "error", err, "trip_id", id)
		return nil, err
	}
	return &trip, nil
}

func (r *tripRepositoryGorm) FindOngoingByUserID(userID uuid.UUID) (*models.Trip, error) {
	var trip models.Trip
	if err := r.db.Where("user_id = ? AND status = ?", userID, models.TripStatusOngoing).First(&trip).Error; err != nil {
		return nil, err
	}
	return &trip, nil
}

func (r *tripRepositoryGorm) Create(trip *models.Trip) error {
	if err := r.db.Create(trip).Error; err != nil {
		r.log.Error("failed to create trip", "module", "trip_repository", "error", err)
		return err
	}
	r.log.Info("trip created", "module", "trip_repository", "trip_id", trip.ID)
	return nil
}

func (r *tripRepositoryGorm) FindByUserID(userID uuid.UUID, page, limit int, search, dateFrom, dateTo string) ([]models.Trip, int64, error) {
	var trips []models.Trip
	query := r.db.Model(&models.Trip{}).Where("user_id = ?", userID)

	if search != "" {
		like := "%" + search + "%"
		query = query.Where("start_address ILIKE ? OR end_address ILIKE ?", like, like)
	}
	if dateFrom != "" {
		query = query.Where("start_time >= ?", dateFrom)
	}
	if dateTo != "" {
		query = query.Where("start_time <= ?", dateTo+" 23:59:59")
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		r.log.Error("failed to count trips", "module", "trip_repository", "error", err)
		return nil, 0, err
	}

	offset := (page - 1) * limit
	if err := query.Offset(offset).Limit(limit).Order("start_time desc").Preload("Vehicle").Find(&trips).Error; err != nil {
		r.log.Error("failed to query trips by user", "module", "trip_repository", "error", err)
		return nil, 0, err
	}

	return trips, total, nil
}

func (r *tripRepositoryGorm) SumDistanceByVehicleID(vehicleID uuid.UUID) (float64, error) {
	var total struct {
		Sum float64
	}
	if err := r.db.Model(&models.Trip{}).
		Select("COALESCE(SUM(total_distance), 0) as sum").
		Where("vehicle_id = ? AND status = ?", vehicleID, models.TripStatusCompleted).
		Scan(&total).Error; err != nil {
		r.log.Error("failed to sum trip distance", "module", "trip_repository", "error", err, "vehicle_id", vehicleID)
		return 0, err
	}
	return total.Sum, nil
}

func (r *tripRepositoryGorm) FindByVehicleID(vehicleID uuid.UUID) ([]models.Trip, error) {
	var trips []models.Trip
	if err := r.db.Where("vehicle_id = ?", vehicleID).Find(&trips).Error; err != nil {
		r.log.Error("failed to query trips by vehicle", "module", "trip_repository", "error", err, "vehicle_id", vehicleID)
		return nil, err
	}
	return trips, nil
}

func (r *tripRepositoryGorm) Delete(id uuid.UUID) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("trip_id = ?", id).Delete(&models.TripPoint{}).Error; err != nil {
			return err
		}
		if err := tx.Delete(&models.Trip{}, id).Error; err != nil {
			return err
		}
		return nil
	})
}

func (r *tripRepositoryGorm) DeleteByVehicleID(vehicleID uuid.UUID) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		var tripIDs []uuid.UUID
		if err := tx.Model(&models.Trip{}).Where("vehicle_id = ?", vehicleID).Pluck("id", &tripIDs).Error; err != nil {
			return err
		}
		if len(tripIDs) > 0 {
			if err := tx.Where("trip_id IN ?", tripIDs).Delete(&models.TripPoint{}).Error; err != nil {
				return err
			}
			if err := tx.Where("id IN ?", tripIDs).Delete(&models.Trip{}).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func (r *tripRepositoryGorm) Save(trip *models.Trip) error {
	if err := r.db.Save(trip).Error; err != nil {
		r.log.Error("failed to save trip", "module", "trip_repository", "error", err, "trip_id", trip.ID)
		return err
	}
	r.log.Info("trip saved", "module", "trip_repository", "trip_id", trip.ID)
	return nil
}

type tripPointRepositoryGorm struct {
	db  *gorm.DB
	log *slog.Logger
}

func NewTripPointRepositoryGorm(db *gorm.DB, logger *slog.Logger) coreRepo.TripPointRepository {
	return &tripPointRepositoryGorm{db: db, log: logger}
}

func (r *tripPointRepositoryGorm) Create(point *models.TripPoint) error {
	if err := r.db.Create(point).Error; err != nil {
		r.log.Error("failed to create trip point", "module", "trip_point_repository", "error", err)
		return err
	}
	return nil
}

func (r *tripPointRepositoryGorm) FindAllByTripID(tripID uuid.UUID) ([]models.TripPoint, error) {
	var points []models.TripPoint
	if err := r.db.Where("trip_id = ?", tripID).Order("recorded_at asc").Find(&points).Error; err != nil {
		return nil, err
	}
	return points, nil
}

func (r *tripPointRepositoryGorm) FindLastByTripID(tripID uuid.UUID) (*models.TripPoint, error) {
	var point models.TripPoint
	if err := r.db.Where("trip_id = ?", tripID).Order("recorded_at desc").First(&point).Error; err != nil {
		return nil, err
	}
	return &point, nil
}

func (r *tripPointRepositoryGorm) CountByTripID(tripID uuid.UUID) (int64, error) {
	var count int64
	if err := r.db.Model(&models.TripPoint{}).Where("trip_id = ?", tripID).Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}

func (r *tripPointRepositoryGorm) DeleteByTripID(tripID uuid.UUID) error {
	if err := r.db.Where("trip_id = ?", tripID).Delete(&models.TripPoint{}).Error; err != nil {
		r.log.Error("failed to delete trip points by trip", "module", "trip_point_repository", "error", err, "trip_id", tripID)
		return err
	}
	return nil
}
