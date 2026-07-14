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
			VehicleType:   t.Vehicle.VehicleType,
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

func (u *tripUsecase) ProcessLocation(userID string, tripID string, point models.TripPoint) error {
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

	// Calculate all metrics from trip points
	points, err := u.tripPointRepo.FindAllByTripID(tid)
	if err != nil {
		return nil, err
	}

	var totalDistance float64
	var maxSpeed float64
	var movingTime int
	var idleTime int

	for i, p := range points {
		if p.Speed > maxSpeed {
			maxSpeed = p.Speed
		}

		if i > 0 {
			prev := points[i-1]
			dist := haversine(prev.Latitude, prev.Longitude, p.Latitude, p.Longitude)
			totalDistance += dist

			timeDiff := p.RecordedAt.Sub(prev.RecordedAt).Seconds()
			if timeDiff > 0 {
				if p.Speed > movingSpeedThreshold {
					movingTime += int(timeDiff)
				} else {
					idleTime += int(timeDiff)
				}
			}
		}
	}

	trip.TotalDistance = math.Round(totalDistance*100) / 100
	trip.MaximumSpeed = math.Round(maxSpeed*100) / 100
	trip.MovingTime = movingTime
	trip.IdleTime = idleTime

	if len(points) > 0 {
		trip.StartLatitude = points[0].Latitude
		trip.StartLongitude = points[0].Longitude
		endLat := points[len(points)-1].Latitude
		endLon := points[len(points)-1].Longitude
		trip.EndLatitude = &endLat
		trip.EndLongitude = &endLon
	}

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

	trip.Status = models.TripStatusCompleted

	if err := u.tripRepo.Save(trip); err != nil {
		return nil, err
	}

	res := toTripResponse(*trip)
	return &res, nil
}

func (u *tripUsecase) GetTripDetail(userID string, tripID string) (*dto.TripDetailResponse, error) {
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

	points, err := u.tripPointRepo.FindAllByTripID(tid)
	if err != nil {
		return nil, err
	}

	route := make([]dto.PointPayload, len(points))
	for i, p := range points {
		route[i] = dto.PointPayload{
			Latitude:   p.Latitude,
			Longitude:  p.Longitude,
			Speed:      p.Speed,
			Heading:    p.Heading,
			Accuracy:   p.Accuracy,
			Altitude:   p.Altitude,
			Battery:    p.Battery,
			RecordedAt: p.RecordedAt.Format(time.RFC3339),
		}
	}

	vehicleName := ""
	plateNumber := ""
	vehicleType := ""
	if trip.Vehicle.VehicleName != "" {
		vehicleName = trip.Vehicle.VehicleName
		plateNumber = trip.Vehicle.PlateNumber
		vehicleType = trip.Vehicle.VehicleType
	}

	res := dto.TripDetailResponse{
		ID:             trip.ID.String(),
		UserID:         trip.UserID.String(),
		VehicleID:      trip.VehicleID.String(),
		VehicleName:    vehicleName,
		PlateNumber:    plateNumber,
		VehicleType:    vehicleType,
		StartTime:      trip.StartTime,
		EndTime:        trip.EndTime,
		StartLatitude:  trip.StartLatitude,
		StartLongitude: trip.StartLongitude,
		EndLatitude:    trip.EndLatitude,
		EndLongitude:   trip.EndLongitude,
		StartAddress:   trip.StartAddress,
		EndAddress:     trip.EndAddress,
		TotalDistance:  trip.TotalDistance,
		Duration:       trip.Duration,
		MovingTime:     trip.MovingTime,
		IdleTime:       trip.IdleTime,
		AverageSpeed:   trip.AverageSpeed,
		MaximumSpeed:   trip.MaximumSpeed,
		FuelConsumed:   trip.FuelConsumed,
		Status:         trip.Status,
		Route:          route,
		CreatedAt:      trip.CreatedAt,
		UpdatedAt:      trip.UpdatedAt,
	}

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
