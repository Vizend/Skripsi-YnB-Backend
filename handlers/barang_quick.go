package handlers

import (
	"database/sql"
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"
	"ynb-backend/models"

	"github.com/gofiber/fiber/v2"
)

type CreateBarangReq struct {
	NamaBarang string   `json:"nama_barang"`
	HargaJual  float64  `json:"harga_jual"`
	HargaBeli  *float64 `json:"harga_beli,omitempty"`
	KodeBarang string   `json:"kode_barang,omitempty"`
}

func CreateBarangQuick(c *fiber.Ctx) error {
	var in CreateBarangReq
	if err := c.BodyParser(&in); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "JSON tidak valid"})
	}
	in.NamaBarang = strings.TrimSpace(in.NamaBarang)
	if in.NamaBarang == "" || in.HargaJual <= 0 {
		return c.Status(400).JSON(fiber.Map{"error": "nama_barang & harga_jual wajib"})
	}

	// default rasio harga_beli = 10% dari jual (bisa diubah via query ?buy_ratio=0.10)
	ratio := 0.10
	if s := c.Query("buy_ratio"); s != "" {
		if f, err := strconv.ParseFloat(s, 64); err == nil && f >= 0 && f <= 1 {
			ratio = f
		}
	}
	hb := in.HargaJual * ratio
	if in.HargaBeli != nil && *in.HargaBeli > 0 {
		hb = *in.HargaBeli
	}
	// pembulatan sederhana ke rupiah bulat
	hargaJual := int64(math.Round(in.HargaJual))
	hargaBeli := int64(math.Round(hb))

	// siapkan kode_barang jika wajib (tergantung schema)
	kode := strings.TrimSpace(in.KodeBarang)
	if kode == "" {
		kode = fmt.Sprintf("AUT-%d", time.Now().UnixNano())
	}

	db := models.DB
	if db == nil {
		return c.Status(500).JSON(fiber.Map{"error": "DB belum diinisialisasi"})
	}

	res, err := db.Exec(`
		INSERT INTO barang (kode_barang, nama_barang, harga_jual, harga_beli, jumlah_stock, is_active)
		VALUES (?,?,?,?,0,1)
	`, kode, in.NamaBarang, hargaJual, hargaBeli)
	if err != nil {
		// contoh tangani unik constraint
		if err == sql.ErrNoRows {
			return c.Status(400).JSON(fiber.Map{"error": "gagal insert barang"})
		}
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	id, _ := res.LastInsertId()

	return c.JSON(fiber.Map{
		"barang_id":   id,
		"nama_barang": in.NamaBarang,
		"harga_jual":  hargaJual,
		"harga_beli":  hargaBeli,
		"kode_barang": kode,
	})
}
