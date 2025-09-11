package handlers

import (
	"time"
	"ynb-backend/models"

	"github.com/gofiber/fiber/v2"
)

func BackfillPembelian(c *fiber.Ctx) error {
	kredit := c.Query("credit")
	var startPtr, endPtr *time.Time

	if s := c.Query("start"); s != "" {
		t, err := time.Parse("2006-01-02", s)
		if err != nil {
			return fiber.NewError(400, "start invalid (YYYY-MM-DD)")
		}
		startPtr = &t
	}
	if s := c.Query("end"); s != "" {
		t, err := time.Parse("2006-01-02", s)
		if err != nil {
			return fiber.NewError(400, "end invalid (YYYY-MM-DD)")
		}
		endPtr = &t
	}

	if err := models.BackfillPembelianFromBarangMasuk(models.DB, startPtr, endPtr, kredit); err != nil {
		return fiber.NewError(500, err.Error())
	}
	return c.JSON(fiber.Map{"ok": true})
}
