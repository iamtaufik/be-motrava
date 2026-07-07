package repository

import (
	"log/slog"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"motrava/core/models"
	coreRepo "motrava/core/repository"
)

type userRepositoryGorm struct {
	db  *gorm.DB
	log *slog.Logger
}

func NewUserRepositoryGorm(db *gorm.DB, logger *slog.Logger) coreRepo.UserRepository {
	return &userRepositoryGorm{db: db, log: logger}
}

func (r *userRepositoryGorm) FindAll() ([]models.User, error) {
	var users []models.User
	if err := r.db.Find(&users).Error; err != nil {
		r.log.Error("failed to query users", "module", "user_repository", "error", err)
		return nil, err
	}

	r.log.Info("users queried", "module", "user_repository", "count", len(users))
	return users, nil
}

func (r *userRepositoryGorm) FindByID(id uuid.UUID) (*models.User, error) {
	var user models.User
	if err := r.db.First(&user, id).Error; err != nil {
		r.log.Error("failed to query user by id", "module", "user_repository", "error", err, "user_id", id)
		return nil, err
	}

	return &user, nil
}

func (r *userRepositoryGorm) FindByEmail(email string) (*models.User, error) {
	var user models.User
	if err := r.db.Where("email = ?", email).First(&user).Error; err != nil {
		r.log.Error("failed to query user by email", "module", "user_repository", "error", err, "email", email)
		return nil, err
	}

	return &user, nil
}

func (r *userRepositoryGorm) FindByGoogleSub(googleSub string) (*models.User, error) {
	var user models.User
	if err := r.db.Where("google_sub = ?", googleSub).First(&user).Error; err != nil {
		r.log.Error("failed to query user by google sub", "module", "user_repository", "error", err, "google_sub", googleSub)
		return nil, err
	}

	return &user, nil
}

func (r *userRepositoryGorm) Create(user *models.User) error {
	if err := r.db.Create(user).Error; err != nil {
		r.log.Error("failed to create user", "module", "user_repository", "error", err)
		return err
	}

	r.log.Info("user created", "module", "user_repository", "user_id", user.ID)
	return nil
}

func (r *userRepositoryGorm) Save(user *models.User) error {
	if err := r.db.Save(user).Error; err != nil {
		r.log.Error("failed to save user", "module", "user_repository", "error", err, "user_id", user.ID)
		return err
	}

	r.log.Info("user saved", "module", "user_repository", "user_id", user.ID)
	return nil
}
