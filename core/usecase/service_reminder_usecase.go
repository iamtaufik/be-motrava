package usecase

import (
	"errors"
	"fmt"
	"math"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"motrava/core/dto"
	"motrava/core/models"
	portUsecase "motrava/core/port/usecase"
	"motrava/core/repository"
)

const (
	ServiceReminderThresholdPercent = 90.0
)

type serviceReminderUsecase struct {
	reminderRepo  repository.ServiceReminderRepository
	manualLogRepo repository.ManualDistanceLogRepository
	vehicleRepo   repository.VehicleRepository
	notifier     *ReminderNotifier
}

func NewServiceReminderUsecase(
	reminderRepo repository.ServiceReminderRepository,
	manualLogRepo repository.ManualDistanceLogRepository,
	vehicleRepo repository.VehicleRepository,
	notifier *ReminderNotifier,
) portUsecase.ServiceReminderUsecase {
	return &serviceReminderUsecase{
		reminderRepo:  reminderRepo,
		manualLogRepo: manualLogRepo,
		vehicleRepo:   vehicleRepo,
		notifier:     notifier,
	}
}

func (u *serviceReminderUsecase) CreateReminder(userID string, vehicleID string, input dto.CreateServiceReminderRequest) (*dto.ServiceReminderResponse, error) {
	uid, err := uuid.Parse(userID)
	if err != nil {
		return nil, fmt.Errorf("invalid user id: %w", err)
	}

	vid, err := uuid.Parse(vehicleID)
	if err != nil {
		return nil, fmt.Errorf("invalid vehicle id: %w", err)
	}

	vehicle, err := u.vehicleRepo.FindByIDAndUserID(vid, uid)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("vehicle not found")
		}
		return nil, err
	}

	if input.ServiceName == "" {
		return nil, fmt.Errorf("service_name is required")
	}
	if input.IntervalKM <= 0 {
		return nil, fmt.Errorf("interval_km must be greater than 0")
	}

	reminder := models.ServiceReminder{
		ID:            uuid.New(),
		VehicleID:     vehicle.ID,
		UserID:        uid,
		ServiceName:   input.ServiceName,
		IntervalKM:    input.IntervalKM,
		AccumulatedKM: 0,
		IsActive:      true,
	}

	if err := u.reminderRepo.Create(&reminder); err != nil {
		return nil, err
	}

	res := toServiceReminderResponse(reminder)
	return &res, nil
}

func (u *serviceReminderUsecase) GetReminderProgress(userID string, vehicleID string, reminderID string) (*dto.ServiceReminderResponse, error) {
	reminder, err := u.validateOwnership(userID, vehicleID, reminderID)
	if err != nil {
		return nil, err
	}

	res := toServiceReminderResponse(*reminder)
	return &res, nil
}

func (u *serviceReminderUsecase) ListRemindersByVehicle(userID string, vehicleID string) ([]dto.ServiceReminderResponse, error) {
	uid, err := uuid.Parse(userID)
	if err != nil {
		return nil, fmt.Errorf("invalid user id: %w", err)
	}

	vid, err := uuid.Parse(vehicleID)
	if err != nil {
		return nil, fmt.Errorf("invalid vehicle id: %w", err)
	}

	_, err = u.vehicleRepo.FindByIDAndUserID(vid, uid)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("vehicle not found")
		}
		return nil, err
	}

	reminders, err := u.reminderRepo.FindByVehicleID(vid)
	if err != nil {
		return nil, err
	}

	res := make([]dto.ServiceReminderResponse, len(reminders))
	for i, r := range reminders {
		res[i] = toServiceReminderResponse(r)
	}

	return res, nil
}

func (u *serviceReminderUsecase) UpdateReminder(userID string, vehicleID string, reminderID string, input dto.UpdateServiceReminderRequest) (*dto.ServiceReminderResponse, error) {
	reminder, err := u.validateOwnership(userID, vehicleID, reminderID)
	if err != nil {
		return nil, err
	}

	if input.ServiceName != nil {
		reminder.ServiceName = *input.ServiceName
	}
	if input.IntervalKM != nil {
		if *input.IntervalKM <= 0 {
			return nil, fmt.Errorf("interval_km must be greater than 0")
		}
		reminder.IntervalKM = *input.IntervalKM
	}

	if err := u.reminderRepo.Save(reminder); err != nil {
		return nil, err
	}

	res := toServiceReminderResponse(*reminder)
	return &res, nil
}

func (u *serviceReminderUsecase) DeleteReminder(userID string, vehicleID string, reminderID string) error {
	reminder, err := u.validateOwnership(userID, vehicleID, reminderID)
	if err != nil {
		return err
	}

	return u.reminderRepo.Delete(reminder.ID)
}

func (u *serviceReminderUsecase) ResetReminder(userID string, vehicleID string, reminderID string) (*dto.ServiceReminderResponse, error) {
	reminder, err := u.validateOwnership(userID, vehicleID, reminderID)
	if err != nil {
		return nil, err
	}

	now := time.Now().UTC()
	reminder.AccumulatedKM = 0
	reminder.LastServiceAt = &now
	reminder.NotifiedAt = nil

	if err := u.reminderRepo.Save(reminder); err != nil {
		return nil, err
	}

	res := toServiceReminderResponse(*reminder)
	return &res, nil
}

func (u *serviceReminderUsecase) AddManualDistance(userID string, vehicleID string, reminderID string, input dto.ManualDistanceRequest) (*dto.ServiceReminderResponse, error) {
	reminder, err := u.validateOwnership(userID, vehicleID, reminderID)
	if err != nil {
		return nil, err
	}

	if input.DistanceKM <= 0 {
		return nil, fmt.Errorf("distance_km must be greater than 0")
	}

	log := models.ManualDistanceLog{
		ID:         uuid.New(),
		ReminderID: reminder.ID,
		DistanceKM: input.DistanceKM,
		Note:       input.Note,
	}

	if err := u.manualLogRepo.Create(&log); err != nil {
		return nil, err
	}

	reminder.AccumulatedKM = math.Round((reminder.AccumulatedKM+input.DistanceKM)*100) / 100

	if err := u.reminderRepo.Save(reminder); err != nil {
		return nil, err
	}

	if u.notifier != nil {
		u.notifier.CheckAndNotify(reminder)
	}

	res := toServiceReminderResponse(*reminder)
	return &res, nil
}

func (u *serviceReminderUsecase) validateOwnership(userID string, vehicleID string, reminderID string) (*models.ServiceReminder, error) {
	uid, err := uuid.Parse(userID)
	if err != nil {
		return nil, fmt.Errorf("invalid user id: %w", err)
	}

	vid, err := uuid.Parse(vehicleID)
	if err != nil {
		return nil, fmt.Errorf("invalid vehicle id: %w", err)
	}

	rid, err := uuid.Parse(reminderID)
	if err != nil {
		return nil, fmt.Errorf("invalid reminder id: %w", err)
	}

	vehicle, err := u.vehicleRepo.FindByIDAndUserID(vid, uid)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("vehicle not found")
		}
		return nil, err
	}

	reminder, err := u.reminderRepo.FindByID(rid)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("reminder not found")
		}
		return nil, err
	}

	if reminder.VehicleID != vehicle.ID {
		return nil, fmt.Errorf("reminder not found")
	}

	return reminder, nil
}

func toServiceReminderResponse(r models.ServiceReminder) dto.ServiceReminderResponse {
	var progressPercent float64
	if r.IntervalKM > 0 {
		progressPercent = math.Round((r.AccumulatedKM/r.IntervalKM)*100*100) / 100
	}

	return dto.ServiceReminderResponse{
		ID:              r.ID.String(),
		VehicleID:       r.VehicleID.String(),
		ServiceName:     r.ServiceName,
		IntervalKM:      r.IntervalKM,
		AccumulatedKM:   r.AccumulatedKM,
		ProgressPercent: progressPercent,
		NeedsService:    progressPercent >= 100,
		NotifiedAt:      r.NotifiedAt,
		LastServiceAt:   r.LastServiceAt,
		IsActive:        r.IsActive,
		CreatedAt:       r.CreatedAt,
		UpdatedAt:       r.UpdatedAt,
	}
}
