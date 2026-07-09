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
	if err := r.db.First(&trip, id).Error; err != nil {
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
