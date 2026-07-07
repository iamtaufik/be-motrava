package usecase

import (
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"motrava/core/dto"
	"motrava/core/models"
	portUsecase "motrava/core/port/usecase"
	"motrava/core/repository"
)

type userUsecase struct {
	userRepo repository.UserRepository
}

func NewUserUsecase(userRepo repository.UserRepository) portUsecase.UserUsecase {
	return &userUsecase{userRepo: userRepo}
}

func (u *userUsecase) ListUsers() ([]dto.UserResponse, error) {
	users, err := u.userRepo.FindAll()
	if err != nil {
		return nil, err
	}

	responses := make([]dto.UserResponse, 0, len(users))
	for _, user := range users {
		responses = append(responses, toUserResponse(user))
	}

	return responses, nil
}

func (u *userUsecase) GetUserByID(id string) (*dto.UserResponse, error) {
	userID, err := uuid.Parse(id)
	if err != nil {
		return nil, fmt.Errorf("invalid user id: %w", err)
	}

	user, err := u.userRepo.FindByID(userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}

	res := toUserResponse(*user)
	return &res, nil
}

func (u *userUsecase) CreateUser(input dto.CreateUserRequest) (*dto.UserResponse, error) {
	fullName := strings.TrimSpace(input.FullName)
	phoneNumber := strings.TrimSpace(input.PhoneNumber)
	password := strings.TrimSpace(input.Password)
	if password == "" {
		password = ""
	}

	newUser := models.User{
		ID:           uuid.New(),
		Email:        strings.TrimSpace(input.Email),
		FullName:     optionalString(fullName),
		Password:     optionalString(password),
		AuthProvider: models.AuthProviderLocal,
		PhoneNumber:  optionalString(phoneNumber),
		IsActive:     true,
		IsVerified:   false,
	}

	if err := u.userRepo.Create(&newUser); err != nil {
		return nil, err
	}

	res := toUserResponse(newUser)
	return &res, nil
}

func toUserResponse(user models.User) dto.UserResponse {
	return dto.UserResponse{
		ID:           user.ID.String(),
		Email:        user.Email,
		FullName:     user.FullName,
		AuthProvider: user.AuthProvider,
		GoogleSub:    user.GoogleSub,
		AvatarURL:    user.AvatarURL,
		PhoneNumber:  user.PhoneNumber,
		IsActive:     user.IsActive,
		IsVerified:   user.IsVerified,
		LastLoginAt:  user.LastLoginAt,
		CreatedAt:    user.CreatedAt,
		UpdatedAt:    user.UpdatedAt,
	}
}

func optionalString(value string) *string {
	if strings.TrimSpace(value) == "" {
		return nil
	}

	trimmed := strings.TrimSpace(value)
	return &trimmed
}
