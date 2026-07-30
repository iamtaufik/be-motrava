package middleware

import (
	"log/slog"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"gorm.io/gorm"

	"motrava/iam-service/internal/config"
	"motrava/iam-service/internal/core/models"
	"motrava/iam-service/internal/core/repository"
	"motrava/iam-service/internal/core/utils/response"
)

const currentUserKey = "current_user"

type AuthClaims struct {
	TokenType    string `json:"token_type"`
	Email        string `json:"email"`
	AuthProvider string `json:"auth_provider"`
	jwt.RegisteredClaims
}

func NewAuthRequired(cfg config.Config, userRepo repository.UserRepository, logger *slog.Logger) fiber.Handler {
	return func(c *fiber.Ctx) error {
		token := bearerToken(c.Get("Authorization"))
		if token == "" {
			return response.Error(c, fiber.StatusUnauthorized, "authorization bearer token is required", nil)
		}

		claims, err := parseAccessToken(token, cfg.JWTSecret, cfg.JWTIssuer)
		if err != nil {
			logger.Error("invalid access token", "module", "auth_middleware", "error", err)
			return response.Error(c, fiber.StatusUnauthorized, "invalid access token", nil)
		}

		userID, err := uuid.Parse(claims.Subject)
		if err != nil {
			return response.Error(c, fiber.StatusUnauthorized, "invalid access token", nil)
		}

		user, err := userRepo.FindByID(userID)
		if err != nil {
			if err == gorm.ErrRecordNotFound {
				return response.Error(c, fiber.StatusUnauthorized, "invalid access token", nil)
			}
			logger.Error("load current user failed", "module", "auth_middleware", "error", err)
			return response.Error(c, fiber.StatusInternalServerError, "failed to load current user", nil)
		}

		if !user.IsActive {
			return response.Error(c, fiber.StatusUnauthorized, "user is inactive", nil)
		}

		c.Locals(currentUserKey, user)
		return c.Next()
	}
}

func AuthMiddleware(cfg config.Config, userRepo repository.UserRepository, logger *slog.Logger) fiber.Handler {
	return NewAuthRequired(cfg, userRepo, logger)
}

func CurrentUser(c *fiber.Ctx) (*models.User, bool) {
	user, ok := c.Locals(currentUserKey).(*models.User)
	return user, ok
}

func ParseAccessToken(accessToken, jwtSecret, jwtIssuer string) (*AuthClaims, error) {
	return parseAccessToken(accessToken, jwtSecret, jwtIssuer)
}

func parseAccessToken(accessToken, jwtSecret, jwtIssuer string) (*AuthClaims, error) {
	claims := &AuthClaims{}
	token, err := jwt.ParseWithClaims(accessToken, claims, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fiber.ErrUnauthorized
		}
		return []byte(jwtSecret), nil
	})
	if err != nil || !token.Valid {
		return nil, fiber.ErrUnauthorized
	}

	if strings.TrimSpace(claims.TokenType) != "access" || claims.Subject == "" {
		return nil, fiber.ErrUnauthorized
	}

	if jwtIssuer != "" && claims.Issuer != jwtIssuer {
		return nil, fiber.ErrUnauthorized
	}

	return claims, nil
}

func BearerToken(header string) string {
	return bearerToken(header)
}

func bearerToken(header string) string {
	parts := strings.Fields(header)
	if len(parts) != 2 {
		return ""
	}
	if strings.ToLower(parts[0]) != "bearer" {
		return ""
	}
	return strings.TrimSpace(parts[1])
}
