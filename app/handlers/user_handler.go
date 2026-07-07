package handlers

import (
	"log/slog"
	"strings"

	"github.com/gofiber/fiber/v2"

	"motrava/core/dto"
	portUsecase "motrava/core/port/usecase"
	"motrava/core/utils/response"
)

// UserHandler handles user HTTP requests.
type UserHandler struct {
	userUsecase portUsecase.UserUsecase
	log         *slog.Logger
}

func NewUserHandler(userUsecase portUsecase.UserUsecase, logger *slog.Logger) *UserHandler {
	return &UserHandler{userUsecase: userUsecase, log: logger}
}

func (h *UserHandler) ListUsers(c *fiber.Ctx) error {
	users, err := h.userUsecase.ListUsers()
	if err != nil {
		h.log.Error("list users failed", "module", "user_handler", "error", err)
		return response.Error(c, fiber.StatusInternalServerError, "failed to fetch users", nil)
	}

	return response.OK(c, "users fetched", users)
}

func (h *UserHandler) GetUserByID(c *fiber.Ctx) error {
	user, err := h.userUsecase.GetUserByID(c.Params("id"))
	if err != nil {
		if strings.Contains(err.Error(), "invalid user id") {
			return response.ValidationError(c, fiber.Map{"id": "must be a valid UUID"})
		}

		h.log.Error("get user by id failed", "module", "user_handler", "error", err)
		return response.Error(c, fiber.StatusInternalServerError, "failed to fetch user", nil)
	}

	if user == nil {
		return response.Error(c, fiber.StatusNotFound, "user not found", nil)
	}

	return response.OK(c, "user fetched", user)
}

func (h *UserHandler) CreateUser(c *fiber.Ctx) error {
	var req dto.CreateUserRequest
	if err := c.BodyParser(&req); err != nil {
		return response.ValidationError(c, fiber.Map{"body": "invalid json payload"})
	}

	req.FullName = strings.TrimSpace(req.FullName)
	req.Email = strings.TrimSpace(req.Email)
	if req.FullName == "" || req.Email == "" {
		return response.ValidationError(c, fiber.Map{
			"full_name": "required",
			"email":     "required",
		})
	}

	created, err := h.userUsecase.CreateUser(req)
	if err != nil {
		h.log.Error("create user failed", "module", "user_handler", "error", err)
		return response.Error(c, fiber.StatusInternalServerError, "failed to create user", nil)
	}

	return response.Created(c, "user created", created)
}
