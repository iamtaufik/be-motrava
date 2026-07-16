package usecase

import (
	"errors"
	"fmt"
	"math"
	"strings"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"motrava/core/dto"
	"motrava/core/models"
	portUsecase "motrava/core/port/usecase"
	"motrava/core/repository"
)

var _ portUsecase.VehicleUsecase = (*vehicleUsecase)(nil)

type vehicleUsecase struct {
	vehicleRepo repository.VehicleRepository
}

func NewVehicleUsecase(vehicleRepo repository.VehicleRepository) portUsecase.VehicleUsecase {
	return &vehicleUsecase{vehicleRepo: vehicleRepo}
}

func (u *vehicleUsecase) ListVehicles(userID string) ([]dto.VehicleResponse, error) {
	uid, err := uuid.Parse(userID)
	if err != nil {
		return nil, fmt.Errorf("invalid user id: %w", err)
	}

	vehicles, err := u.vehicleRepo.FindByUserID(uid)
	if err != nil {
		return nil, err
	}

	responses := make([]dto.VehicleResponse, 0, len(vehicles))
	for _, v := range vehicles {
		res := toVehicleResponse(v)
		res.TotalDistanceKM = calcTotalDistanceKM(v)
		responses = append(responses, res)
	}

	return responses, nil
}

func (u *vehicleUsecase) GetVehicleByID(id string, userID string) (*dto.VehicleResponse, error) {
	vid, err := uuid.Parse(id)
	if err != nil {
		return nil, fmt.Errorf("invalid vehicle id: %w", err)
	}

	uid, err := uuid.Parse(userID)
	if err != nil {
		return nil, fmt.Errorf("invalid user id: %w", err)
	}

	vehicle, err := u.vehicleRepo.FindByIDAndUserID(vid, uid)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}

	res := toVehicleResponse(*vehicle)
	res.TotalDistanceKM = calcTotalDistanceKM(*vehicle)
	return &res, nil
}

func (u *vehicleUsecase) CreateVehicle(userID string, input dto.CreateVehicleRequest) (*dto.VehicleResponse, error) {
	uid, err := uuid.Parse(userID)
	if err != nil {
		return nil, fmt.Errorf("invalid user id: %w", err)
	}

	vehicleType := strings.ToUpper(strings.TrimSpace(input.VehicleType))
	if vehicleType == "" {
		vehicleType = string(models.VehicleTypeCar)
	}

	vehicle := models.Vehicle{
		ID:                      uuid.New(),
		UserID:                  uid,
		VehicleName:             strings.TrimSpace(input.VehicleName),
		PlateNumber:             strings.TrimSpace(input.PlateNumber),
		Brand:                   strings.TrimSpace(input.Brand),
		Model:                   strings.TrimSpace(input.Model),
		VehicleType:             vehicleType,
		Color:                   strings.TrimSpace(input.Color),
		Year:                    input.Year,
		FuelEfficiencyKmPerLiter: input.FuelEfficiencyKmPerLiter,
		InitialKM:               input.InitialKM,
		IsDefault:               false,
	}

	if input.InitialKM != nil {
		vehicle.LastRecordedOdometerKM = *input.InitialKM
	}

	if photo := strings.TrimSpace(input.Photo); photo != "" {
		vehicle.Photo = &photo
	}

	count, _ := u.vehicleRepo.FindByUserID(uid)
	if len(count) == 0 {
		vehicle.IsDefault = true
	}

	if err := u.vehicleRepo.Create(&vehicle); err != nil {
		return nil, err
	}

	res := toVehicleResponse(vehicle)
	return &res, nil
}

func (u *vehicleUsecase) UpdateVehicle(id string, userID string, input dto.UpdateVehicleRequest) (*dto.VehicleResponse, error) {
	vid, err := uuid.Parse(id)
	if err != nil {
		return nil, fmt.Errorf("invalid vehicle id: %w", err)
	}

	uid, err := uuid.Parse(userID)
	if err != nil {
		return nil, fmt.Errorf("invalid user id: %w", err)
	}

	vehicle, err := u.vehicleRepo.FindByIDAndUserID(vid, uid)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}

	if input.VehicleName != nil {
		vehicle.VehicleName = strings.TrimSpace(*input.VehicleName)
	}
	if input.PlateNumber != nil {
		vehicle.PlateNumber = strings.TrimSpace(*input.PlateNumber)
	}
	if input.Brand != nil {
		vehicle.Brand = strings.TrimSpace(*input.Brand)
	}
	if input.Model != nil {
		vehicle.Model = strings.TrimSpace(*input.Model)
	}
	if input.VehicleType != nil {
		vehicle.VehicleType = strings.ToUpper(strings.TrimSpace(*input.VehicleType))
	}
	if input.Color != nil {
		vehicle.Color = strings.TrimSpace(*input.Color)
	}
	if input.Year != nil {
		vehicle.Year = input.Year
	}
	if input.Photo != nil {
		photo := strings.TrimSpace(*input.Photo)
		if photo == "" {
			vehicle.Photo = nil
		} else {
			vehicle.Photo = &photo
		}
	}
	if input.FuelEfficiencyKmPerLiter != nil {
		vehicle.FuelEfficiencyKmPerLiter = input.FuelEfficiencyKmPerLiter
	}
	if input.InitialKM != nil {
		vehicle.InitialKM = input.InitialKM
	}

	if err := u.vehicleRepo.Save(vehicle); err != nil {
		return nil, err
	}

	res := toVehicleResponse(*vehicle)
	return &res, nil
}

func (u *vehicleUsecase) DeleteVehicle(id string, userID string) error {
	vid, err := uuid.Parse(id)
	if err != nil {
		return fmt.Errorf("invalid vehicle id: %w", err)
	}

	uid, err := uuid.Parse(userID)
	if err != nil {
		return fmt.Errorf("invalid user id: %w", err)
	}

	if err := u.vehicleRepo.Delete(vid, uid); err != nil {
		return err
	}

	return nil
}

func (u *vehicleUsecase) SetDefaultVehicle(id string, userID string) (*dto.VehicleResponse, error) {
	vid, err := uuid.Parse(id)
	if err != nil {
		return nil, fmt.Errorf("invalid vehicle id: %w", err)
	}

	uid, err := uuid.Parse(userID)
	if err != nil {
		return nil, fmt.Errorf("invalid user id: %w", err)
	}

	vehicle, err := u.vehicleRepo.FindByIDAndUserID(vid, uid)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}

	defaultVehicle, err := u.vehicleRepo.FindDefaultByUserID(uid)
	if err == nil && defaultVehicle != nil {
		defaultVehicle.IsDefault = false
		if err := u.vehicleRepo.Save(defaultVehicle); err != nil {
			return nil, err
		}
	}

	vehicle.IsDefault = true
	if err := u.vehicleRepo.Save(vehicle); err != nil {
		return nil, err
	}

	res := toVehicleResponse(*vehicle)
	return &res, nil
}

func calcTotalDistanceKM(v models.Vehicle) float64 {
	if v.InitialKM == nil {
		return 0
	}
	total := v.LastRecordedOdometerKM - *v.InitialKM
	if total < 0 {
		return 0
	}
	return math.Round(total*100) / 100
}

func toVehicleResponse(v models.Vehicle) dto.VehicleResponse {
	return dto.VehicleResponse{
		ID:                      v.ID.String(),
		UserID:                  v.UserID.String(),
		VehicleName:             v.VehicleName,
		PlateNumber:             v.PlateNumber,
		Brand:                   v.Brand,
		Model:                   v.Model,
		VehicleType:             v.VehicleType,
		Color:                   v.Color,
		Year:                    v.Year,
		Photo:                   v.Photo,
		IsDefault:               v.IsDefault,
		InitialKM:               v.InitialKM,
		LastRecordedOdometerKM:  v.LastRecordedOdometerKM,
		FuelEfficiencyKmPerLiter: v.FuelEfficiencyKmPerLiter,
		CreatedAt:               v.CreatedAt,
		UpdatedAt:               v.UpdatedAt,
	}
}
