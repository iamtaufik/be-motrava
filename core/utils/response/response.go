package response

import "github.com/gofiber/fiber/v2"

// APIResponse is the standard response envelope for all endpoints.
type APIResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
	Error   interface{} `json:"error,omitempty"`
	Meta    interface{} `json:"meta,omitempty"`
}

func OK(c *fiber.Ctx, message string, data interface{}) error {
	return c.Status(fiber.StatusOK).JSON(APIResponse{
		Success: true,
		Message: message,
		Data:    data,
	})
}

func Created(c *fiber.Ctx, message string, data interface{}) error {
	return c.Status(fiber.StatusCreated).JSON(APIResponse{
		Success: true,
		Message: message,
		Data:    data,
	})
}

func Error(c *fiber.Ctx, status int, message string, details interface{}) error {
	return c.Status(status).JSON(APIResponse{
		Success: false,
		Message: message,
		Error:   details,
	})
}

func ValidationError(c *fiber.Ctx, details interface{}) error {
	return Error(c, fiber.StatusBadRequest, "validation failed", details)
}
