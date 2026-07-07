package usecase

import "motrava/core/dto"

// UserUsecase defines business operations exposed to handlers.
type UserUsecase interface {
	ListUsers() ([]dto.UserResponse, error)
	GetUserByID(id string) (*dto.UserResponse, error)
	CreateUser(input dto.CreateUserRequest) (*dto.UserResponse, error)
}
