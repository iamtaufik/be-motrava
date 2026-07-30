package repository

import (
	"github.com/google/uuid"

	"motrava/iam-service/internal/core/models"
)

// UserRepository defines persistence operations for users.
type UserRepository interface {
	FindAll() ([]models.User, error)
	FindByID(id uuid.UUID) (*models.User, error)
	FindByEmail(email string) (*models.User, error)
	FindByGoogleSub(googleSub string) (*models.User, error)
	Create(user *models.User) error
	Save(user *models.User) error
}
