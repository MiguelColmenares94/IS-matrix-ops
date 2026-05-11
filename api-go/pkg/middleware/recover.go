package middleware

import (
	"log"

	"github.com/gofiber/fiber/v2"
)

func Recovery() fiber.Handler {
	return func(c *fiber.Ctx) (err error) {
		defer func() {
			if r := recover(); r != nil {
				log.Printf("panic recovered: %v", r)
				err = c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "internal server error"})
			}
		}()
		return c.Next()
	}
}
