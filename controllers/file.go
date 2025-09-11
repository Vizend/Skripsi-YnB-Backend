package controllers

import (
	"github.com/gofiber/fiber/v2"
)

func UploadFile(c *fiber.Ctx) error {
	file, err := c.FormFile("file")
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "File not found"})
	}

	err = c.SaveFile(file, "./uploads/"+file.Filename)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Upload failed"})
	}

	return c.JSON(fiber.Map{"filename": file.Filename})
}
