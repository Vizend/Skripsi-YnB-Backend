package controllers

import (
	"database/sql"
	"time"
	"ynb-backend/models"

	"github.com/gofiber/fiber/v2"
)

// Hitung saldo (debit - kredit)
func getSaldoAkunUpTo(tx *sql.Tx, akunID int, tanggal string) (float64, error) {
	var saldo float64
	err := tx.QueryRow(`
		SELECT COALESCE(SUM(jd.debit - jd.kredit), 0)
		FROM jurnal_detail jd
		JOIN jurnal j ON j.jurnal_id = jd.jurnal_id
		WHERE jd.akun_id = ? AND DATE(j.tanggal) <= DATE(?)
	`, akunID, tanggal).Scan(&saldo)
	return saldo, err
}

// handlers/akuntansi_saldo.go
func GetSaldoAkun(c *fiber.Ctx) error {
	kode := c.Query("kode")       // "1-101" untuk Kas, "1-102" untuk Bank
	tanggal := c.Query("tanggal") // optional; default hari ini
	if tanggal == "" {
		tanggal = time.Now().Format("2006-01-02")
	}

	tx, err := models.DB.Begin()
	if err != nil {
		return fiber.NewError(500, "DB error")
	}
	defer tx.Rollback()

	var akunID int
	if err := tx.QueryRow(`SELECT akun_id FROM akun WHERE kode_akun=?`, kode).Scan(&akunID); err != nil {
		return fiber.NewError(400, "Akun tidak ditemukan")
	}
	saldo, err := getSaldoAkunUpTo(tx, akunID, tanggal)
	if err != nil {
		return fiber.NewError(500, "Gagal hitung saldo")
	}
	return c.JSON(fiber.Map{"kode": kode, "tanggal": tanggal, "saldo": saldo})
}
