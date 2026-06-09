package response

import (
	"log"

	"github.com/gofiber/fiber/v2"
)

type Envelope struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
	Error   string `json:"error,omitempty"`
}

func Success(c *fiber.Ctx, status int, message string, data any) error {
	return c.Status(status).JSON(Envelope{
		Success: true,
		Message: message,
		Data:    data,
	})
}

func Fail(c *fiber.Ctx, status int, message string, err error) error {
	payload := Envelope{
		Success: false,
		Message: message,
	}
	if err != nil {
		log.Printf("request failed: %s: %v", message, err)
		payload.Error = err.Error()
	}

	return c.Status(status).JSON(payload)
}

func FiberErrorHandler(c *fiber.Ctx, err error) error {
	status := fiber.StatusInternalServerError
	message := "internal server error"
	if fe, ok := err.(*fiber.Error); ok {
		status = fe.Code
		message = fe.Message
	}
	payload := Envelope{
		Success: false,
		Message: message,
	}
	if err != nil {
		log.Printf("fiber error: %v", err)
		payload.Error = message
	}
	return c.Status(status).JSON(payload)
}
