package usecase

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/idtoken"
	"gorm.io/gorm"

	"motrava/config"
	"motrava/core/dto"
	"motrava/core/models"
	portUsecase "motrava/core/port/usecase"
	"motrava/core/repository"
)

type googleAuthUsecase struct {
	userRepo repository.UserRepository
	config   oauth2.Config
	log      *slog.Logger
}

type googleUserInfo struct {
	Sub           string `json:"sub"`
	Email         string `json:"email"`
	EmailVerified bool   `json:"email_verified"`
	Name          string `json:"name"`
	Picture       string `json:"picture"`
}

func NewGoogleAuthUsecase(cfg config.Config, userRepo repository.UserRepository, logger *slog.Logger) portUsecase.AuthUsecase {
	return &googleAuthUsecase{
		userRepo: userRepo,
		config: oauth2.Config{
			ClientID:     cfg.GoogleClientID,
			ClientSecret: cfg.GoogleClientSecret,
			RedirectURL:  cfg.GoogleRedirectURL,
			Scopes: []string{
				"openid",
				"email",
				"profile",
			},
			Endpoint: google.Endpoint,
		},
		log: logger,
	}
}

func (u *googleAuthUsecase) GetGoogleAuthURL(state string) (string, error) {
	if strings.TrimSpace(u.config.ClientID) == "" || strings.TrimSpace(u.config.ClientSecret) == "" {
		return "", fmt.Errorf("google oauth config is not set")
	}

	return u.config.AuthCodeURL(state, oauth2.AccessTypeOffline, oauth2.SetAuthURLParam("prompt", "consent")), nil
}

func (u *googleAuthUsecase) HandleGoogleCallback(ctx context.Context, code string) (*dto.UserResponse, error) {
	token, err := u.config.Exchange(ctx, code)
	if err != nil {
		return nil, fmt.Errorf("exchange google code: %w", err)
	}

	client := u.config.Client(ctx, token)
	profile, err := fetchGoogleUserInfo(client)
	if err != nil {
		return nil, err
	}

	now := time.Now().UTC()
	user, err := u.upsertGoogleUser(profile, now)
	if err != nil {
		return nil, err
	}

	response := toUserResponse(*user)
	return &response, nil
}

func (u *googleAuthUsecase) HandleGoogleMobileLogin(ctx context.Context, idToken string) (*dto.UserResponse, error) {
	if strings.TrimSpace(u.config.ClientID) == "" {
		return nil, fmt.Errorf("google client id is not configured")
	}

	validator, err := idtoken.Validate(ctx, idToken, u.config.ClientID)
	if err != nil {
		return nil, fmt.Errorf("validate google id token: %w", err)
	}

	claims := struct {
		Sub           string `json:"sub"`
		Email         string `json:"email"`
		EmailVerified bool   `json:"email_verified"`
		Name          string `json:"name"`
		Picture       string `json:"picture"`
	}{
		Sub:           validator.Claims["sub"].(string),
		Email:         validator.Claims["email"].(string),
		EmailVerified: validator.Claims["email_verified"].(bool),
		Name:          valueAsString(validator.Claims["name"]),
		Picture:       valueAsString(validator.Claims["picture"]),
	}

	if strings.TrimSpace(claims.Sub) == "" || strings.TrimSpace(claims.Email) == "" {
		return nil, fmt.Errorf("google id token is missing required claims")
	}

	now := time.Now().UTC()
	user, err := u.upsertGoogleUser(&googleUserInfo{
		Sub:           claims.Sub,
		Email:         claims.Email,
		EmailVerified: claims.EmailVerified,
		Name:          claims.Name,
		Picture:       claims.Picture,
	}, now)
	if err != nil {
		return nil, err
	}

	response := toUserResponse(*user)
	return &response, nil
}

func fetchGoogleUserInfo(client *http.Client) (*googleUserInfo, error) {
	request, err := http.NewRequest(http.MethodGet, "https://www.googleapis.com/oauth2/v3/userinfo", nil)
	if err != nil {
		return nil, fmt.Errorf("build google userinfo request: %w", err)
	}

	response, err := client.Do(request)
	if err != nil {
		return nil, fmt.Errorf("request google userinfo: %w", err)
	}
	defer response.Body.Close()

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, fmt.Errorf("read google userinfo response: %w", err)
	}

	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("google userinfo returned status %d: %s", response.StatusCode, string(body))
	}

	var profile googleUserInfo
	if err := json.Unmarshal(body, &profile); err != nil {
		return nil, fmt.Errorf("decode google userinfo: %w", err)
	}

	if strings.TrimSpace(profile.Sub) == "" || strings.TrimSpace(profile.Email) == "" {
		return nil, fmt.Errorf("google userinfo is missing required fields")
	}

	return &profile, nil
}

func valueAsString(value interface{}) string {
	if value == nil {
		return ""
	}

	text, ok := value.(string)
	if !ok {
		return ""
	}

	return text
}

func (u *googleAuthUsecase) upsertGoogleUser(profile *googleUserInfo, now time.Time) (*models.User, error) {
	if googleSub := strings.TrimSpace(profile.Sub); googleSub != "" {
		if user, err := u.userRepo.FindByGoogleSub(googleSub); err == nil {
			return u.updateGoogleUser(user, profile, now)
		} else if !errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("find user by google sub: %w", err)
		}
	}

	if user, err := u.userRepo.FindByEmail(strings.TrimSpace(profile.Email)); err == nil {
		return u.updateGoogleUser(user, profile, now)
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, fmt.Errorf("find user by email: %w", err)
	}

	newUser := &models.User{
		ID:           uuid.New(),
		Email:        strings.TrimSpace(profile.Email),
		FullName:     optionalString(profile.Name),
		AuthProvider: models.AuthProviderGoogle,
		GoogleSub:    optionalString(profile.Sub),
		AvatarURL:    optionalString(profile.Picture),
		IsActive:     true,
		IsVerified:   profile.EmailVerified,
		LastLoginAt:  &now,
	}

	if err := u.userRepo.Create(newUser); err != nil {
		return nil, err
	}

	return newUser, nil
}

func (u *googleAuthUsecase) updateGoogleUser(user *models.User, profile *googleUserInfo, now time.Time) (*models.User, error) {
	user.Email = strings.TrimSpace(profile.Email)
	user.FullName = optionalString(profile.Name)
	user.AuthProvider = models.AuthProviderGoogle
	user.GoogleSub = optionalString(profile.Sub)
	user.AvatarURL = optionalString(profile.Picture)
	user.IsActive = true
	user.IsVerified = true
	user.LastLoginAt = &now

	if err := u.userRepo.Save(user); err != nil {
		return nil, err
	}

	return user, nil
}
