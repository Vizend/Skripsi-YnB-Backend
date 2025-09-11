package controllers

import (
	"database/sql"
	"strings"

	"github.com/gofiber/fiber/v2"
	"ynb-backend/config"
)

func ifNullF(n sql.NullFloat64) float64 {
	if n.Valid {
		return n.Float64
	}
	return 0
}
func ifNullI(n sql.NullInt64) int64 {
	if n.Valid {
		return n.Int64
	}
	return 0
}

// GET /api/barang/prefill?kode=ABC   atau   /api/barang/prefill?nama=odol
// Balikkan data inti + harga pembelian saran (pakai harga beli terakhir; fallback: kolom harga_beli di tabel barang)
func PrefillBarang(c *fiber.Ctx) error {
	kode := strings.TrimSpace(c.Query("kode"))
	nama := strings.TrimSpace(c.Query("nama"))
	prefer := strings.ToLower(strings.TrimSpace(c.Query("prefer"))) // "", "fifo"
	if kode == "" && nama == "" {
		return c.Status(400).JSON(fiber.Map{"message": "Wajib isi ?kode= atau ?nama="})
	}

	db := config.DB
	var (
		barangID               int
		kodeBarang, namaBarang string
		hargaJual              float64
		isActive               int
		err                    error
	)

	if kode != "" {
		err = db.QueryRow(`
			SELECT barang_id, kode_barang, nama_barang, harga_jual, is_active
			FROM barang WHERE kode_barang = ?`,
			kode,
		).Scan(&barangID, &kodeBarang, &namaBarang, &hargaJual, &isActive)
	} else {
		// ambil yang paling mendekati nama
		err = db.QueryRow(`
			SELECT barang_id, kode_barang, nama_barang, harga_jual, is_active
			FROM barang
			WHERE nama_barang LIKE ?
			ORDER BY LENGTH(nama_barang) ASC
			LIMIT 1`,
			"%"+nama+"%",
		).Scan(&barangID, &kodeBarang, &namaBarang, &hargaJual, &isActive)
	}
	if err == sql.ErrNoRows {
		return c.Status(404).JSON(fiber.Map{"message": "Barang tidak ditemukan"})
	}
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"message": "DB error", "detail": err.Error()})
	}

	var stokTersisa sql.NullInt64
	_ = db.QueryRow(`SELECT COALESCE(SUM(sisa_stok),0) FROM barang_masuk WHERE barang_id=?`, barangID).
		Scan(&stokTersisa)

	// Harga FIFO (batch tertua yang masih punya sisa)
	var fifoHarga sql.NullFloat64
	_ = db.QueryRow(`
		SELECT harga_beli FROM barang_masuk
		WHERE barang_id=? AND sisa_stok>0
		ORDER BY tanggal ASC, masuk_id ASC
		LIMIT 1`, barangID).Scan(&fifoHarga)

	// Harga beli terakhir (transaksi pembelian paling akhir)
	var lastBeli sql.NullFloat64
	_ = db.QueryRow(`
		SELECT harga_beli FROM barang_masuk
		WHERE barang_id=? ORDER BY tanggal DESC, masuk_id DESC LIMIT 1`,
		barangID).Scan(&lastBeli)

	// 🎯 Saran harga pembelian HANYA dari barang_masuk
	var defaultHarga float64
	if prefer == "fifo" {
		defaultHarga = ifNullF(fifoHarga)
		if defaultHarga == 0 {
			defaultHarga = ifNullF(lastBeli)
		}
	} else {
		// default: pakai harga beli terakhir; fallback ke FIFO sisa
		defaultHarga = ifNullF(lastBeli)
		if defaultHarga == 0 {
			defaultHarga = ifNullF(fifoHarga)
		}
	}

	return c.JSON(fiber.Map{
		"barang_id":             barangID,
		"kode_barang":           kodeBarang,
		"nama_barang":           namaBarang,
		"harga_jual":            hargaJual,
		"saran_harga_pembelian": defaultHarga,
		"harga_beli_terakhir":   ifNullF(lastBeli),
		"harga_fifo_sisa":       ifNullF(fifoHarga),
		"stok_tersedia":         ifNullI(stokTersisa),
		"is_active":             isActive,
	})
}

// GET /api/barang/search?q=odol
// Untuk dropdown autocomplete nama/kode
func SearchBarang(c *fiber.Ctx) error {
	q := strings.TrimSpace(c.Query("q"))
	if q == "" {
		return c.Status(400).JSON(fiber.Map{"message": "q wajib diisi"})
	}
	rows, err := config.DB.Query(`
		SELECT barang_id, kode_barang, nama_barang, harga_jual, is_active
		FROM barang
		WHERE kode_barang LIKE ? OR nama_barang LIKE ?
		ORDER BY is_active DESC, nama_barang ASC
		LIMIT 10`,
		"%"+q+"%", "%"+q+"%",
	)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"message": "DB error"})
	}
	defer rows.Close()

	list := []fiber.Map{}
	for rows.Next() {
		var id, active int
		var kode, nama string
		var jual float64
		if err := rows.Scan(&id, &kode, &nama, &jual, &active); err == nil {
			list = append(list, fiber.Map{
				"barang_id":   id,
				"kode_barang": kode,
				"nama_barang": nama,
				"harga_jual":  jual,
				"is_active":   active,
			})
		}
	}
	return c.JSON(list)
}
