package usecase

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/mail"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/idtoken"
	"gorm.io/gorm"

	"motrava/iam-service/internal/config"
	"motrava/iam-service/internal/core/dto"
	"motrava/iam-service/internal/core/models"
	portUsecase "motrava/iam-service/internal/core/port/usecase"
	"motrava/iam-service/internal/core/repository"
)

const (
	accessTokenType  = "access"
	refreshTokenType = "refresh"
	tokenTypeBearer  = "Bearer"
)

type googleAuthUsecase struct {
	userRepo          repository.UserRepository
	refreshTokenRepo  repository.RefreshTokenRepository
	googleOAuthConfig oauth2.Config
	jwtSecret         string
	jwtIssuer         string
	accessTokenTTL    time.Duration
	refreshTokenTTL   time.Duration
	log               *slog.Logger
}

type googleUserInfo struct {
	Sub           string `json:"sub"`
	Email         string `json:"email"`
	EmailVerified bool   `json:"email_verified"`
	Name          string `json:"name"`
	Picture       string `json:"picture"`
}

type authClaims struct {
	TokenType    string `json:"token_type"`
	Email        string `json:"email"`
	AuthProvider string `json:"auth_provider"`
	jwt.RegisteredClaims
}

func NewAuthUsecase(cfg config.Config, userRepo repository.UserRepository, refreshTokenRepo repository.RefreshTokenRepository, logger *slog.Logger) portUsecase.AuthUsecase {
	return &googleAuthUsecase{
		userRepo:         userRepo,
		refreshTokenRepo: refreshTokenRepo,
		googleOAuthConfig: oauth2.Config{
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
		jwtSecret:       cfg.JWTSecret,
		jwtIssuer:       cfg.JWTIssuer,
		accessTokenTTL:  time.Duration(cfg.AccessTokenTTLMinutes) * time.Minute,
		refreshTokenTTL: time.Duration(cfg.RefreshTokenTTLHours) * time.Hour,
		log:             logger,
	}
}

func NewGoogleAuthUsecase(cfg config.Config, userRepo repository.UserRepository, refreshTokenRepo repository.RefreshTokenRepository, logger *slog.Logger) portUsecase.AuthUsecase {
	return NewAuthUsecase(cfg, userRepo, refreshTokenRepo, logger)
}

func (u *googleAuthUsecase) Register(ctx context.Context, req dto.AuthRegisterRequest) (*dto.AuthResponse, error) {
	if err := u.ensureAuthConfig(); err != nil {
		return nil, err
	}

	email := normalizeEmail(req.Email)
	fullName := strings.TrimSpace(req.FullName)
	password := req.Password

	if fullName == "" || email == "" || strings.TrimSpace(password) == "" {
		return nil, fmt.Errorf("full_name, email, and password are required")
	}

	if _, err := mail.ParseAddress(email); err != nil {
		return nil, fmt.Errorf("invalid email format")
	}

	if _, err := u.userRepo.FindByEmail(email); err == nil {
		return nil, portUsecase.ErrEmailAlreadyExists
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("hash password: %w", err)
	}

	user := &models.User{
		ID:           uuid.New(),
		Email:        email,
		FullName:     optionalStringPtr(fullName),
		Password:     optionalStringPtr(string(passwordHash)),
		AuthProvider: models.AuthProviderLocal,
		IsActive:     true,
		IsVerified:   false,
	}

	if err := u.userRepo.Create(user); err != nil {
		return nil, err
	}

	return u.issueAuthResponse(user)
}

func (u *googleAuthUsecase) Login(ctx context.Context, req dto.AuthLoginRequest) (*dto.AuthResponse, error) {
	if err := u.ensureAuthConfig(); err != nil {
		return nil, err
	}

	email := normalizeEmail(req.Email)
	password := req.Password
	if email == "" || strings.TrimSpace(password) == "" {
		return nil, portUsecase.ErrInvalidCredentials
	}

	user, err := u.userRepo.FindByEmail(email)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, portUsecase.ErrInvalidCredentials
		}
		return nil, err
	}

	if user.Password == nil || bcrypt.CompareHashAndPassword([]byte(*user.Password), []byte(password)) != nil {
		return nil, portUsecase.ErrInvalidCredentials
	}

	user.AuthProvider = models.AuthProviderLocal
	return u.issueAuthResponse(user)
}

func (u *googleAuthUsecase) Refresh(ctx context.Context, refreshToken string) (*dto.AuthResponse, error) {
	if err := u.ensureAuthConfig(); err != nil {
		return nil, err
	}

	claims, err := u.parseRefreshToken(refreshToken)
	if err != nil {
		return nil, portUsecase.ErrInvalidRefreshToken
	}

	tokenHash := hashToken(refreshToken)
	storedToken, err := u.refreshTokenRepo.FindByHash(tokenHash)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, portUsecase.ErrInvalidRefreshToken
		}
		return nil, err
	}

	now := time.Now().UTC()
	if storedToken.RevokedAt != nil || now.After(storedToken.ExpiresAt) || storedToken.UserID.String() != claims.Subject {
		return nil, portUsecase.ErrInvalidRefreshToken
	}

	if err := u.refreshTokenRepo.RevokeByHash(tokenHash); err != nil {
		return nil, err
	}

	user, err := u.userRepo.FindByID(storedToken.UserID)
	if err != nil {
		return nil, err
	}

	return u.issueAuthResponse(user)
}

func (u *googleAuthUsecase) Me(ctx context.Context, accessToken string) (*dto.UserResponse, error) {
	if err := u.ensureAuthConfig(); err != nil {
		return nil, err
	}

	claims, err := u.parseAccessToken(accessToken)
	if err != nil {
		return nil, portUsecase.ErrInvalidCredentials
	}

	userID, err := uuid.Parse(claims.Subject)
	if err != nil {
		return nil, portUsecase.ErrInvalidCredentials
	}

	user, err := u.userRepo.FindByID(userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, portUsecase.ErrInvalidCredentials
		}
		return nil, err
	}

	response := toUserResponse(*user)
	return &response, nil
}

func (u *googleAuthUsecase) GetGoogleAuthURL(state string) (string, error) {
	if strings.TrimSpace(u.googleOAuthConfig.ClientID) == "" || strings.TrimSpace(u.googleOAuthConfig.ClientSecret) == "" {
		return "", fmt.Errorf("google oauth config is not set")
	}

	return u.googleOAuthConfig.AuthCodeURL(state, oauth2.AccessTypeOffline, oauth2.SetAuthURLParam("prompt", "consent")), nil
}

func (u *googleAuthUsecase) HandleGoogleCallback(ctx context.Context, code string) (*dto.UserResponse, error) {
	token, err := u.googleOAuthConfig.Exchange(ctx, code)
	if err != nil {
		return nil, fmt.Errorf("exchange google code: %w", err)
	}

	client := u.googleOAuthConfig.Client(ctx, token)
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

func (u *googleAuthUsecase) HandleGoogleMobileLogin(ctx context.Context, idToken string) (*dto.AuthResponse, error) {
	if err := u.ensureAuthConfig(); err != nil {
		return nil, err
	}

	validator, err := idtoken.Validate(ctx, idToken, u.googleOAuthConfig.ClientID)
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
		Sub:           valueAsString(validator.Claims["sub"]),
		Email:         valueAsString(validator.Claims["email"]),
		EmailVerified: valueAsBool(validator.Claims["email_verified"]),
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

	return u.issueAuthResponse(user)
}

func (u *googleAuthUsecase) issueAuthResponse(user *models.User) (*dto.AuthResponse, error) {
	now := time.Now().UTC()
	user.LastLoginAt = &now
	if err := u.userRepo.Save(user); err != nil {
		return nil, err
	}

	accessToken, _, err := u.createSignedToken(user, now, u.accessTokenTTL, accessTokenType)
	if err != nil {
		return nil, err
	}

	refreshToken, refreshClaims, err := u.createSignedToken(user, now, u.refreshTokenTTL, refreshTokenType)
	if err != nil {
		return nil, err
	}

	refreshRecord := &models.RefreshToken{
		ID:        uuid.New(),
		UserID:    user.ID,
		TokenHash: hashToken(refreshToken),
		ExpiresAt: refreshClaims.ExpiresAt.Time,
	}

	if err := u.refreshTokenRepo.Create(refreshRecord); err != nil {
		return nil, err
	}

	return &dto.AuthResponse{
		User:         toUserResponse(*user),
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		TokenType:    tokenTypeBearer,
		ExpiresIn:    int64(u.accessTokenTTL.Seconds()),
	}, nil
}

func (u *googleAuthUsecase) createSignedToken(user *models.User, now time.Time, ttl time.Duration, tokenType string) (string, *authClaims, error) {
	claims := &authClaims{
		TokenType:    tokenType,
		Email:        user.Email,
		AuthProvider: user.AuthProvider,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    u.jwtIssuer,
			Subject:   user.ID.String(),
			ID:        uuid.NewString(),
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(ttl)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString([]byte(u.jwtSecret))
	if err != nil {
		return "", nil, fmt.Errorf("sign jwt: %w", err)
	}

	return signed, claims, nil
}

func (u *googleAuthUsecase) parseRefreshToken(refreshToken string) (*authClaims, error) {
	claims := &authClaims{}
	token, err := jwt.ParseWithClaims(refreshToken, claims, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method")
		}

		return []byte(u.jwtSecret), nil
	})
	if err != nil || !token.Valid {
		return nil, portUsecase.ErrInvalidRefreshToken
	}

	if claims.TokenType != refreshTokenType || claims.Subject == "" {
		return nil, portUsecase.ErrInvalidRefreshToken
	}

	return claims, nil
}

func (u *googleAuthUsecase) parseAccessToken(accessToken string) (*authClaims, error) {
	claims := &authClaims{}
	token, err := jwt.ParseWithClaims(accessToken, claims, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method")
		}

		return []byte(u.jwtSecret), nil
	})
	if err != nil || !token.Valid {
		return nil, portUsecase.ErrInvalidCredentials
	}

	if claims.TokenType != accessTokenType || claims.Subject == "" {
		return nil, portUsecase.ErrInvalidCredentials
	}

	return claims, nil
}

func (u *googleAuthUsecase) ensureAuthConfig() error {
	if strings.TrimSpace(u.jwtSecret) == "" {
		return fmt.Errorf("jwt secret is not configured")
	}

	return nil
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

func valueAsBool(value interface{}) bool {
	if value == nil {
		return false
	}

	result, ok := value.(bool)
	if !ok {
		return false
	}

	return result
}

func hashToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}

func normalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}

func optionalStringPtr(value string) *string {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return nil
	}

	return &trimmed
}

func (u *googleAuthUsecase) upsertGoogleUser(profile *googleUserInfo, now time.Time) (*models.User, error) {
	email := normalizeEmail(profile.Email)
	googleSub := strings.TrimSpace(profile.Sub)

	if googleSub != "" {
		if user, err := u.userRepo.FindByGoogleSub(googleSub); err == nil {
			return u.updateGoogleUser(user, profile, now)
		} else if !errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("find user by google sub: %w", err)
		}
	}

	if user, err := u.userRepo.FindByEmail(email); err == nil {
		return u.updateGoogleUser(user, profile, now)
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, fmt.Errorf("find user by email: %w", err)
	}

	newUser := &models.User{
		ID:           uuid.New(),
		Email:        email,
		FullName:     optionalStringPtr(profile.Name),
		AuthProvider: models.AuthProviderGoogle,
		GoogleSub:    optionalStringPtr(profile.Sub),
		AvatarURL:    optionalStringPtr(profile.Picture),
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
	user.Email = normalizeEmail(profile.Email)
	user.FullName = optionalStringPtr(profile.Name)
	user.AuthProvider = models.AuthProviderGoogle
	user.GoogleSub = optionalStringPtr(profile.Sub)
	user.AvatarURL = optionalStringPtr(profile.Picture)
	user.IsActive = true
	user.IsVerified = true
	user.LastLoginAt = &now

	if err := u.userRepo.Save(user); err != nil {
		return nil, err
	}

	return user, nil
}
