package handlers

import (
	// "fmt"
	"fmt"
	"time"
	// "time"
	"ynb-backend/models"
	"ynb-backend/utils"

	"github.com/gofiber/fiber/v2"
)

func UploadXJDHandler(c *fiber.Ctx) error {
	fileHeader, err := c.FormFile("file")
	if err != nil {
		fmt.Println("Gagal ambil file:", err)
		return c.Status(400).SendString("File tidak ditemukan")
	}

	filePath := "./uploads/" + fileHeader.Filename
	if err := c.SaveFile(fileHeader, filePath); err != nil {
		fmt.Println("Gagal simpan file:", err)
		return c.Status(500).SendString("Gagal menyimpan file")
	}

	fmt.Println("File berhasil diupload ke:", filePath)

	transaksiList, err := utils.ParseXJDFile(filePath)
	if err != nil {
		fmt.Println("Gagal parse file:", err)
		return c.Status(500).SendString("Gagal parsing XJD")
	}

	fmt.Printf("Ditemukan %d transaksi\n", len(transaksiList))

	for i, trx := range transaksiList {
		fmt.Printf("Transaksi #%d: %s %s, %d item\n", i+1, trx.Tanggal, trx.Jam, len(trx.Items))

		err := models.ProcessTransaksiFIFO(trx)
		if err != nil {
			fmt.Printf("Gagal proses transaksi %d: %v\n", i+1, err)
			// return c.Status(500).SendString(fmt.Sprintf("Gagal memproses transaksi %d", i+1))
			return c.Status(500).JSON(fiber.Map{
				"error": "Gagal memproses transaksi 1",
			})
		}

		// Tambahan: beri waktu untuk driver SQL
		time.Sleep(10 * time.Millisecond)
	}


	return c.JSON(fiber.Map{
		"message": "Transaksi berhasil diproses",
	})
}
