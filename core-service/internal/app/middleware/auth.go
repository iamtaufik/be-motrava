package middleware

import (
	"log/slog"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"

	"motrava/core-service/internal/config"
	"motrava/core-service/internal/core/models"
	"motrava/core-service/internal/core/utils/response"
	"motrava/core-service/internal/infra/iamclient"
)

const currentUserKey = "current_user"

type AuthClaims struct {
	TokenType    string `json:"token_type"`
	Email        string `json:"email"`
	AuthProvider string `json:"auth_provider"`
	jwt.RegisteredClaims
}

func CoreAuthMiddleware(cfg config.Config, iamClient *iamclient.Client, logger *slog.Logger) fiber.Handler {
	return func(c *fiber.Ctx) error {
		token := bearerToken(c.Get("Authorization"))
		if token == "" {
			return response.Error(c, fiber.StatusUnauthorized, "authorization bearer token is required", nil)
		}

		resp, err := iamClient.ValidateToken(c.Context(), token)
		if err != nil {
			logger.Error("core auth: iam validation failed", "module", "core_auth_middleware", "error", err)
			return response.Error(c, fiber.StatusUnauthorized, "invalid access token", nil)
		}

		if !resp.Valid {
			return response.Error(c, fiber.StatusUnauthorized, "invalid access token", nil)
		}

		userID, err := uuid.Parse(resp.UserId)
		if err != nil {
			return response.Error(c, fiber.StatusUnauthorized, "invalid access token", nil)
		}

		user := &models.User{
			ID:       userID,
			Email:    resp.Email,
			IsActive: resp.IsActive,
		}
		if resp.FullName != "" {
			user.FullName = &resp.FullName
		}
		if resp.AvatarUrl != "" {
			user.AvatarURL = &resp.AvatarUrl
		}

		c.Locals(currentUserKey, user)
		return c.Next()
	}
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
