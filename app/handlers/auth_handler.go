package handlers

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"log/slog"
	"strings"

	"github.com/gofiber/fiber/v2"

	"motrava/app/middleware"
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

func (h *AuthHandler) Register(c *fiber.Ctx) error {
	var req dto.AuthRegisterRequest
	if err := c.BodyParser(&req); err != nil {
		return response.ValidationError(c, fiber.Map{"body": "invalid json payload"})
	}

	req.FullName = strings.TrimSpace(req.FullName)
	req.Email = strings.TrimSpace(req.Email)
	if req.FullName == "" || req.Email == "" || req.Password == "" {
		return response.ValidationError(c, fiber.Map{
			"full_name": "required",
			"email":     "required",
			"password":  "required",
		})
	}

	result, err := h.authUsecase.Register(c.Context(), req)
	if err != nil {
		return h.handleAuthError(c, err, "register failed")
	}

	return response.Created(c, "register success", result)
}

func (h *AuthHandler) Login(c *fiber.Ctx) error {
	var req dto.AuthLoginRequest
	if err := c.BodyParser(&req); err != nil {
		return response.ValidationError(c, fiber.Map{"body": "invalid json payload"})
	}

	req.Email = strings.TrimSpace(req.Email)
	if req.Email == "" || req.Password == "" {
		return response.ValidationError(c, fiber.Map{
			"email":    "required",
			"password": "required",
		})
	}

	result, err := h.authUsecase.Login(c.Context(), req)
	if err != nil {
		return h.handleAuthError(c, err, "login failed")
	}

	return response.OK(c, "login success", result)
}

func (h *AuthHandler) Refresh(c *fiber.Ctx) error {
	var req dto.AuthRefreshRequest
	if err := c.BodyParser(&req); err != nil {
		return response.ValidationError(c, fiber.Map{"body": "invalid json payload"})
	}

	req.RefreshToken = strings.TrimSpace(req.RefreshToken)
	if req.RefreshToken == "" {
		return response.ValidationError(c, fiber.Map{"refresh_token": "required"})
	}

	result, err := h.authUsecase.Refresh(c.Context(), req.RefreshToken)
	if err != nil {
		return h.handleAuthError(c, err, "refresh token failed")
	}

	return response.OK(c, "refresh token success", result)
}

func (h *AuthHandler) Me(c *fiber.Ctx) error {
	user, ok := middleware.CurrentUser(c)
	if !ok || user == nil {
		return response.Error(c, fiber.StatusUnauthorized, "unauthorized", nil)
	}

	return response.OK(c, "current user fetched", user)
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

func (h *AuthHandler) handleAuthError(c *fiber.Ctx, err error, fallbackMessage string) error {
	switch {
	case errors.Is(err, usecase.ErrEmailAlreadyExists):
		return response.Error(c, fiber.StatusConflict, "email already registered", nil)
	case errors.Is(err, usecase.ErrInvalidCredentials):
		return response.Error(c, fiber.StatusUnauthorized, "invalid credentials", nil)
	case errors.Is(err, usecase.ErrInvalidRefreshToken):
		return response.Error(c, fiber.StatusUnauthorized, "invalid refresh token", nil)
	default:
		h.log.Error(fallbackMessage, "module", "auth_handler", "error", err)
		return response.Error(c, fiber.StatusInternalServerError, fallbackMessage, nil)
	}
}

func generateState() (string, error) {
	buffer := make([]byte, 32)
	if _, err := rand.Read(buffer); err != nil {
		return "", err
	}

	return base64.RawURLEncoding.EncodeToString(buffer), nil
}
