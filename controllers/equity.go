package controllers

import (
	"database/sql"
	"fmt"
	"time"
	"ynb-backend/config"

	"github.com/gofiber/fiber/v2"
)

type EquityInput struct {
	Tanggal    string  `json:"tanggal"` // YYYY-MM-DD
	Tipe       string  `json:"tipe"`    // "modal" | "prive"
	Metode     string  `json:"metode"`  // "kas" | "bank"
	Jumlah     float64 `json:"jumlah"`
	Keterangan string  `json:"keterangan,omitempty"`
	UserID     *int    `json:"user_id,omitempty"`
}

func getAkunIDEquity(tx *sql.Tx, kode string) (int, error) {
	var id int
	err := tx.QueryRow(`SELECT akun_id FROM akun WHERE kode_akun=?`, kode).Scan(&id)
	if err == sql.ErrNoRows {
		return 0, fmt.Errorf("akun %s tidak ditemukan", kode)
	}
	return id, err
}

func ensureAkunPrive(tx *sql.Tx) (int, error) {
	// cek sudah ada?
	var id int
	err := tx.QueryRow(`SELECT akun_id FROM akun WHERE kode_akun='3-101'`).Scan(&id)
	if err == nil && id > 0 {
		return id, nil
	}

	// jika belum ada, buat akun Prive
	// cari parent Ekuitas (3-000)
	var parent int
	if err := tx.QueryRow(`SELECT akun_id FROM akun WHERE kode_akun='3-000'`).Scan(&parent); err != nil {
		return 0, fmt.Errorf("akun header Ekuitas (3-000) tidak ditemukan")
	}
	res, err := tx.Exec(`
		INSERT INTO akun (kode_akun, nama_akun, jenis, parent_id, is_header)
		VALUES ('3-101','Prive','Equity',?,0)`, parent)
	if err != nil {
		return 0, err
	}
	newID, _ := res.LastInsertId()
	return int(newID), nil
}

// POST /api/akuntansi/equity
// Body: {tanggal:"2025-09-05", tipe:"modal"|"prive", metode:"kas"|"bank", jumlah:1000000, keterangan?}
func CreateEquity(c *fiber.Ctx) error {
	var in EquityInput
	if err := c.BodyParser(&in); err != nil {
		return c.Status(400).JSON(fiber.Map{"message": "Payload tidak valid"})
	}
	if in.Tanggal == "" || in.Tipe == "" || in.Metode == "" || in.Jumlah <= 0 {
		return c.Status(400).JSON(fiber.Map{"message": "tanggal, tipe, metode wajib & jumlah > 0"})
	}
	if _, err := time.Parse("2006-01-02", in.Tanggal); err != nil {
		return c.Status(400).JSON(fiber.Map{"message": "Format tanggal harus YYYY-MM-DD"})
	}
	if in.Tipe != "modal" && in.Tipe != "prive" {
		return c.Status(400).JSON(fiber.Map{"message": "tipe harus salah satu dari: modal|prive"})
	}
	if in.Metode != "kas" && in.Metode != "bank" {
		return c.Status(400).JSON(fiber.Map{"message": "metode harus salah satu dari: kas|bank"})
	}

	tx, err := config.DB.Begin()
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"message": "Gagal mulai transaksi"})
	}
	defer tx.Rollback()

	//akun sisi aset
	assetKode := map[string]string{"kas": "1-101", "bank": "1-102"}[in.Metode]
	assetID, err := getAkunIDEquity(tx, assetKode)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"message": err.Error()})
	}

	//akun equitas
	modalID, err := getAkunIDEquity(tx, "3-100") //Modal Pemilik
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"message": err.Error()})
	}

	//akun prive (dibuat jika belum ada)
	priveID, err := ensureAkunPrive(tx)
	if in.Tipe == "prive" && err != nil {
		return c.Status(500).JSON(fiber.Map{"message": "gagal memastikan akun prive", "detail": err.Error()})
	}

	//user
	userID := 1
	if in.UserID != nil && *in.UserID > 0 {
		userID = *in.UserID
	}

	// header jurnal
	ref := fmt.Sprintf("%s-%s", map[string]string{"modal": "MDL", "prive": "PRV"}[in.Tipe], time.Now().Format("060102150405"))
	res, err := tx.Exec(`INSERT INTO jurnal (tanggal, referensi, tipe_jurnal, user_id) VALUES (?, ?, ?, ?)`,
		in.Tanggal, ref, map[string]string{"modal": "Modal", "prive": "Prive"}[in.Tipe], userID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"message": "Gagal insert jurnal"})
	}
	jid, _ := res.LastInsertId()

	ket := in.Keterangan
	if ket == "" {
		if in.Tipe == "modal" {
			ket = "Setoran Modal"
		} else {
			ket = "Pengambilan Prive"
		}
	}

	// detail
	if in.Tipe == "modal" {
		// Debit Kas/Bank, Kredit Modal
		if _, err := tx.Exec(`INSERT INTO jurnal_detail (jurnal_id, akun_id, debit, kredit, keterangan) VALUES (?, ?, ?, 0, ?)`, jid, assetID, in.Jumlah, ket); err != nil {
			return c.Status(500).JSON(fiber.Map{"message": "Gagal insert debit"})
		}
		if _, err := tx.Exec(`INSERT INTO jurnal_detail (jurnal_id, akun_id, debit, kredit, keterangan) VALUES (?, ?, 0, ?, ?)`, jid, modalID, in.Jumlah, ket); err != nil {
			return c.Status(500).JSON(fiber.Map{"message": "Gagal insert kredit"})
		}
	} else {
		// PRIVE: Debit Prive, Kredit Kas/Bank
		if _, err := tx.Exec(`INSERT INTO jurnal_detail (jurnal_id, akun_id, debit, kredit, keterangan) VALUES (?, ?, ?, 0, ?)`, jid, priveID, in.Jumlah, ket); err != nil {
			return c.Status(500).JSON(fiber.Map{"message": "Gagal insert debit"})
		}
		if _, err := tx.Exec(`INSERT INTO jurnal_detail (jurnal_id, akun_id, debit, kredit, keterangan) VALUES (?, ?, 0, ?, ?)`, jid, assetID, in.Jumlah, ket); err != nil {
			return c.Status(500).JSON(fiber.Map{"message": "Gagal insert kredit"})
		}
	}

	if err := tx.Commit(); err != nil {
		return c.Status(500).JSON(fiber.Map{"message": "Gagal commit"})
	}

	return c.JSON(fiber.Map{
		"message":   "Transaksi ekuitas tersimpan",
		"jurnal_id": jid,
		"referensi": ref,
		"tanggal":   in.Tanggal,
		"tipe":      in.Tipe,
		"metode":    in.Metode,
		"jumlah":    in.Jumlah,
	})
}
