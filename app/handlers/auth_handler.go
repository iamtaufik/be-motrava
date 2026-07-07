package handlers

import (
	"crypto/rand"
	"encoding/base64"
	"log/slog"
	"strings"

	"github.com/gofiber/fiber/v2"

	"motrava/core/dto"
	"motrava/core/port/usecase"
	"motrava/core/utils/response"
)

const googleOAuthStateCookie = "google_oauth_state"

// AuthHandler handles authentication flows.
type AuthHandler struct {
	authUsecase usecase.AuthUsecase
	log         *slog.Logger
}

func NewAuthHandler(authUsecase usecase.AuthUsecase, logger *slog.Logger) *AuthHandler {
	return &AuthHandler{authUsecase: authUsecase, log: logger}
}

func (h *AuthHandler) GoogleLogin(c *fiber.Ctx) error {
	state, err := generateState()
	if err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "failed to create oauth state", nil)
	}

	authURL, err := h.authUsecase.GetGoogleAuthURL(state)
	if err != nil {
		h.log.Error("get google auth url failed", "module", "auth_handler", "error", err)
		return response.Error(c, fiber.StatusInternalServerError, "google auth is not configured", nil)
	}

	c.Cookie(&fiber.Cookie{
		Name:     googleOAuthStateCookie,
		Value:    state,
		HTTPOnly: true,
		Secure:   false,
		Path:     "/",
		SameSite: "Lax",
	})

	return c.Redirect(authURL, fiber.StatusFound)
}

func (h *AuthHandler) GoogleCallback(c *fiber.Ctx) error {
	state := c.Query("state")
	cookieState := c.Cookies(googleOAuthStateCookie)
	if state == "" || cookieState == "" || state != cookieState {
		return response.ValidationError(c, fiber.Map{"state": "invalid oauth state"})
	}

	code := c.Query("code")
	if code == "" {
		return response.ValidationError(c, fiber.Map{"code": "required"})
	}

	user, err := h.authUsecase.HandleGoogleCallback(c.Context(), code)
	if err != nil {
		h.log.Error("google callback failed", "module", "auth_handler", "error", err)
		return response.Error(c, fiber.StatusInternalServerError, "google authentication failed", nil)
	}

	c.ClearCookie(googleOAuthStateCookie)

	return response.OK(c, "google authentication success", user)
}

func (h *AuthHandler) GoogleMobileLogin(c *fiber.Ctx) error {
	var req dto.GoogleMobileAuthRequest
	if err := c.BodyParser(&req); err != nil {
		return response.ValidationError(c, fiber.Map{"body": "invalid json payload"})
	}

	req.IDToken = strings.TrimSpace(req.IDToken)
	if req.IDToken == "" {
		return response.ValidationError(c, fiber.Map{"id_token": "required"})
	}

	user, err := h.authUsecase.HandleGoogleMobileLogin(c.Context(), req.IDToken)
	if err != nil {
		h.log.Error("google mobile login failed", "module", "auth_handler", "error", err)
		return response.Error(c, fiber.StatusUnauthorized, "google authentication failed", nil)
	}

	return response.OK(c, "google mobile authentication success", user)
}

func generateState() (string, error) {
	buffer := make([]byte, 32)
	if _, err := rand.Read(buffer); err != nil {
		return "", err
	}

	return base64.RawURLEncoding.EncodeToString(buffer), nil
}
