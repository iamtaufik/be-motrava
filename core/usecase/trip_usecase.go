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
	movingSpeedThreshold = 2.0 // km/h — below this is considered idle
	earthRadiusKm        = 6371.0
)

type tripUsecase struct {
	tripRepo      repository.TripRepository
	tripPointRepo repository.TripPointRepository
	vehicleRepo   repository.VehicleRepository
}

func NewTripUsecase(
	tripRepo repository.TripRepository,
	tripPointRepo repository.TripPointRepository,
	vehicleRepo repository.VehicleRepository,
) portUsecase.TripUsecase {
	return &tripUsecase{
		tripRepo:      tripRepo,
		tripPointRepo: tripPointRepo,
		vehicleRepo:   vehicleRepo,
	}
}

func (u *tripUsecase) StartTrip(userID string, input dto.StartTripRequest) (*dto.TripResponse, error) {
	uid, err := uuid.Parse(userID)
	if err != nil {
		return nil, fmt.Errorf("invalid user id: %w", err)
	}

	vid, err := uuid.Parse(input.VehicleID)
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

	ongoing, _ := u.tripRepo.FindOngoingByUserID(uid)
	if ongoing != nil {
		return nil, fmt.Errorf("there is already an ongoing trip")
	}

	trip := models.Trip{
		ID:        uuid.New(),
		UserID:    uid,
		VehicleID: vehicle.ID,
		StartTime: time.Now().UTC(),
		Status:    models.TripStatusOngoing,
	}

	if err := u.tripRepo.Create(&trip); err != nil {
		return nil, err
	}

	res := toTripResponse(trip)
	return &res, nil
}

func (u *tripUsecase) GetTripHistory(userID string, req dto.TripHistoryRequest) ([]dto.TripHistoryItem, *dto.PaginationMeta, error) {
	uid, err := uuid.Parse(userID)
	if err != nil {
		return nil, nil, fmt.Errorf("invalid user id: %w", err)
	}

	page := req.Page
	if page < 1 {
		page = 1
	}
	limit := req.Limit
	if limit < 1 || limit > 100 {
		limit = 10
	}

	trips, total, err := u.tripRepo.FindByUserID(uid, page, limit, req.Search, req.DateFrom, req.DateTo)
	if err != nil {
		return nil, nil, err
	}

	totalPages := int(total) / limit
	if int(total)%limit != 0 {
		totalPages++
	}

	items := make([]dto.TripHistoryItem, len(trips))
	for i, t := range trips {
		vehicleName := ""
		plateNumber := ""
		if t.Vehicle.VehicleName != "" {
			vehicleName = t.Vehicle.VehicleName
			plateNumber = t.Vehicle.PlateNumber
		}

		items[i] = dto.TripHistoryItem{
			ID:            t.ID.String(),
			VehicleName:   vehicleName,
			PlateNumber:   plateNumber,
			StartTime:     t.StartTime,
			EndTime:       t.EndTime,
			StartAddress:  t.StartAddress,
			EndAddress:    t.EndAddress,
			TotalDistance: t.TotalDistance,
			Duration:      t.Duration,
			MovingTime:    t.MovingTime,
			IdleTime:      t.IdleTime,
			AverageSpeed:  t.AverageSpeed,
			MaximumSpeed:  t.MaximumSpeed,
			FuelConsumed:  t.FuelConsumed,
			Status:        t.Status,
			CreatedAt:     t.CreatedAt,
		}
	}

	meta := &dto.PaginationMeta{
		Page:       page,
		Limit:      limit,
		TotalItems: total,
		TotalPages: totalPages,
	}

	return items, meta, nil
}

func (u *tripUsecase) ProcessLocation(userID string, tripID string, point models.TripPoint, speed float64) error {
	uid, err := uuid.Parse(userID)
	if err != nil {
		return fmt.Errorf("invalid user id: %w", err)
	}

	tid, err := uuid.Parse(tripID)
	if err != nil {
		return fmt.Errorf("invalid trip id: %w", err)
	}

	trip, err := u.tripRepo.FindByID(tid)
	if err != nil {
		return err
	}

	if trip.UserID != uid {
		return fmt.Errorf("trip not found")
	}

	if trip.Status != models.TripStatusOngoing {
		return fmt.Errorf("trip is not ongoing")
	}

	if err := u.tripPointRepo.Create(&point); err != nil {
		return err
	}

	lastPoint, _ := u.tripPointRepo.FindLastByTripID(tid)
	if lastPoint != nil && lastPoint.ID != point.ID {
		dist := haversine(lastPoint.Latitude, lastPoint.Longitude, point.Latitude, point.Longitude)
		trip.TotalDistance += dist

		timeDiff := point.RecordedAt.Sub(lastPoint.RecordedAt).Seconds()
		if timeDiff > 0 {
			if speed > movingSpeedThreshold {
				trip.MovingTime += int(timeDiff)
			} else {
				trip.IdleTime += int(timeDiff)
			}
		}
	}

	if speed > trip.MaximumSpeed {
		trip.MaximumSpeed = speed
	}

	if trip.StartLatitude == 0 && trip.StartLongitude == 0 {
		trip.StartLatitude = point.Latitude
		trip.StartLongitude = point.Longitude
	}

	trip.EndLatitude = &point.Latitude
	trip.EndLongitude = &point.Longitude

	return u.tripRepo.Save(trip)
}

func (u *tripUsecase) EndTrip(userID string, tripID string) (*dto.TripResponse, error) {
	uid, err := uuid.Parse(userID)
	if err != nil {
		return nil, fmt.Errorf("invalid user id: %w", err)
	}

	tid, err := uuid.Parse(tripID)
	if err != nil {
		return nil, fmt.Errorf("invalid trip id: %w", err)
	}

	trip, err := u.tripRepo.FindByID(tid)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("trip not found")
		}
		return nil, err
	}

	if trip.UserID != uid {
		return nil, fmt.Errorf("trip not found")
	}

	if trip.Status != models.TripStatusOngoing {
		return nil, fmt.Errorf("trip is not ongoing")
	}

	now := time.Now().UTC()
	trip.EndTime = &now

	elapsed := now.Sub(trip.StartTime).Seconds()
	trip.Duration = int(elapsed)

	if trip.TotalDistance > 0 && trip.Duration > 0 {
		avgSpeed := (trip.TotalDistance / 1000.0) / (float64(trip.Duration) / 3600.0)
		trip.AverageSpeed = math.Round(avgSpeed*100) / 100
	}

	// Calculate estimated fuel consumed
	if trip.TotalDistance > 0 {
		vehicle, err := u.vehicleRepo.FindByIDAndUserID(trip.VehicleID, uid)
		if err == nil && vehicle != nil && vehicle.FuelEfficiencyKmPerLiter != nil && *vehicle.FuelEfficiencyKmPerLiter > 0 {
			distanceKm := trip.TotalDistance / 1000.0
			consumed := distanceKm / *vehicle.FuelEfficiencyKmPerLiter
			consumed = math.Round(consumed*100) / 100
			trip.FuelConsumed = &consumed
		}
	}

	trip.MaximumSpeed = math.Round(trip.MaximumSpeed*100) / 100
	trip.TotalDistance = math.Round(trip.TotalDistance*100) / 100
	trip.Status = models.TripStatusCompleted

	if err := u.tripRepo.Save(trip); err != nil {
		return nil, err
	}

	res := toTripResponse(*trip)
	return &res, nil
}

func toTripResponse(t models.Trip) dto.TripResponse {
	return dto.TripResponse{
		ID:             t.ID.String(),
		UserID:         t.UserID.String(),
		VehicleID:      t.VehicleID.String(),
		StartTime:      t.StartTime,
		EndTime:        t.EndTime,
		StartLatitude:  t.StartLatitude,
		StartLongitude: t.StartLongitude,
		EndLatitude:    t.EndLatitude,
		EndLongitude:   t.EndLongitude,
		StartAddress:   t.StartAddress,
		EndAddress:     t.EndAddress,
		TotalDistance:  t.TotalDistance,
		Duration:       t.Duration,
		MovingTime:     t.MovingTime,
		IdleTime:       t.IdleTime,
		AverageSpeed:   t.AverageSpeed,
		MaximumSpeed:   t.MaximumSpeed,
		FuelConsumed:   t.FuelConsumed,
		Status:         t.Status,
		CreatedAt:      t.CreatedAt,
		UpdatedAt:      t.UpdatedAt,
	}
}

func haversine(lat1, lon1, lat2, lon2 float64) float64 {
	dLat := (lat2 - lat1) * math.Pi / 180.0
	dLon := (lon2 - lon1) * math.Pi / 180.0

	a := math.Sin(dLat/2)*math.Sin(dLat/2) +
		math.Cos(lat1*math.Pi/180.0)*math.Cos(lat2*math.Pi/180.0)*
			math.Sin(dLon/2)*math.Sin(dLon/2)

	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))

	return earthRadiusKm * c * 1000.0 // result in meters
}
