package repository

import (
	"log/slog"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"motrava/core/models"
	coreRepo "motrava/core/repository"
)

type serviceReminderRepositoryGorm struct {
	db  *gorm.DB
	log *slog.Logger
}

func NewServiceReminderRepositoryGorm(db *gorm.DB, logger *slog.Logger) coreRepo.ServiceReminderRepository {
	return &serviceReminderRepositoryGorm{db: db, log: logger}
}

func (r *serviceReminderRepositoryGorm) FindByID(id uuid.UUID) (*models.ServiceReminder, error) {
	var reminder models.ServiceReminder
	if err := r.db.First(&reminder, id).Error; err != nil {
		r.log.Error("failed to query service reminder by id", "module", "service_reminder_repository", "error", err, "reminder_id", id)
		return nil, err
	}
	return &reminder, nil
}

func (r *serviceReminderRepositoryGorm) FindByVehicleID(vehicleID uuid.UUID) ([]models.ServiceReminder, error) {
	var reminders []models.ServiceReminder
	if err := r.db.Where("vehicle_id = ?", vehicleID).Order("created_at desc").Find(&reminders).Error; err != nil {
		r.log.Error("failed to query reminders by vehicle", "module", "service_reminder_repository", "error", err, "vehicle_id", vehicleID)
		return nil, err
	}
	return reminders, nil
}

func (r *serviceReminderRepositoryGorm) FindByUserID(userID uuid.UUID) ([]models.ServiceReminder, error) {
	var reminders []models.ServiceReminder
	if err := r.db.Where("user_id = ?", userID).Order("created_at desc").Find(&reminders).Error; err != nil {
		r.log.Error("failed to query reminders by user", "module", "service_reminder_repository", "error", err, "user_id", userID)
		return nil, err
	}
	return reminders, nil
}

func (r *serviceReminderRepositoryGorm) FindAllActive() ([]models.ServiceReminder, error) {
	var reminders []models.ServiceReminder
	if err := r.db.Where("is_active = ?", true).Find(&reminders).Error; err != nil {
		r.log.Error("failed to query active reminders", "module", "service_reminder_repository", "error", err)
		return nil, err
	}
	return reminders, nil
}

func (r *serviceReminderRepositoryGorm) Create(reminder *models.ServiceReminder) error {
	if err := r.db.Create(reminder).Error; err != nil {
		r.log.Error("failed to create service reminder", "module", "service_reminder_repository", "error", err)
		return err
	}
	r.log.Info("service reminder created", "module", "service_reminder_repository", "reminder_id", reminder.ID)
	return nil
}

func (r *serviceReminderRepositoryGorm) Save(reminder *models.ServiceReminder) error {
	if err := r.db.Save(reminder).Error; err != nil {
		r.log.Error("failed to save service reminder", "module", "service_reminder_repository", "error", err, "reminder_id", reminder.ID)
		return err
	}
	r.log.Info("service reminder saved", "module", "service_reminder_repository", "reminder_id", reminder.ID)
	return nil
}

func (r *serviceReminderRepositoryGorm) DeleteByVehicleID(vehicleID uuid.UUID) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		var reminderIDs []uuid.UUID
		if err := tx.Model(&models.ServiceReminder{}).Where("vehicle_id = ?", vehicleID).Pluck("id", &reminderIDs).Error; err != nil {
			return err
		}
		if len(reminderIDs) > 0 {
			if err := tx.Where("reminder_id IN ?", reminderIDs).Delete(&models.ManualDistanceLog{}).Error; err != nil {
				return err
			}
			if err := tx.Where("id IN ?", reminderIDs).Delete(&models.ServiceReminder{}).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func (r *serviceReminderRepositoryGorm) Delete(id uuid.UUID) error {
	if err := r.db.Delete(&models.ServiceReminder{}, id).Error; err != nil {
		r.log.Error("failed to delete service reminder", "module", "service_reminder_repository", "error", err, "reminder_id", id)
		return err
	}
	r.log.Info("service reminder deleted", "module", "service_reminder_repository", "reminder_id", id)
	return nil
}

type manualDistanceLogRepositoryGorm struct {
	db  *gorm.DB
	log *slog.Logger
}

func NewManualDistanceLogRepositoryGorm(db *gorm.DB, logger *slog.Logger) coreRepo.ManualDistanceLogRepository {
	return &manualDistanceLogRepositoryGorm{db: db, log: logger}
}

func (r *manualDistanceLogRepositoryGorm) Create(log *models.ManualDistanceLog) error {
	if err := r.db.Create(log).Error; err != nil {
		r.log.Error("failed to create manual distance log", "module", "manual_distance_log_repository", "error", err)
		return err
	}
	return nil
}

func (r *manualDistanceLogRepositoryGorm) FindByReminderID(reminderID uuid.UUID) ([]models.ManualDistanceLog, error) {
	var logs []models.ManualDistanceLog
	if err := r.db.Where("reminder_id = ?", reminderID).Order("created_at desc").Find(&logs).Error; err != nil {
		r.log.Error("failed to query manual distance logs", "module", "manual_distance_log_repository", "error", err, "reminder_id", reminderID)
		return nil, err
	}
	return logs, nil
}

func (r *manualDistanceLogRepositoryGorm) DeleteByReminderID(reminderID uuid.UUID) error {
	if err := r.db.Where("reminder_id = ?", reminderID).Delete(&models.ManualDistanceLog{}).Error; err != nil {
		r.log.Error("failed to delete manual distance logs by reminder", "module", "manual_distance_log_repository", "error", err, "reminder_id", reminderID)
		return err
	}
	return nil
}
