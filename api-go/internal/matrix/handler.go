package matrix

import (
	"github.com/gofiber/fiber/v2"
)

type Handler struct {
	svc  *Service
	repo *Repository
}

func NewHandler(svc *Service, repo *Repository) *Handler {
	return &Handler{svc: svc, repo: repo}
}

func (h *Handler) ComputeQR(c *fiber.Ctx) error {
	userID, _ := c.Locals("userID").(string)

	var body struct {
		Matrix [][]interface{} `json:"matrix"`
	}
	if err := c.BodyParser(&body); err != nil || body.Matrix == nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "matrix must not be empty",
			"example": exampleMatrix,
		})
	}

	data, err := h.svc.ValidateMatrix(body.Matrix)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   err.Error(),
			"example": exampleMatrix,
		})
	}

	result, err := h.svc.FactorizeQR(data)
	if err != nil {
		_ = h.repo.SaveQRComputation(userID, data, nil, nil, false, err.Error())
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "computation failed"})
	}

	_ = h.repo.SaveQRComputation(userID, data, result.Q, result.R, true, "")
	return c.JSON(result)
}
