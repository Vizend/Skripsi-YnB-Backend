package controllers

import (
	"database/sql"
	// "fmt"
	"strconv"
	"strings"
	"time"
	"ynb-backend/config"
	"ynb-backend/models"

	"github.com/gofiber/fiber/v2"
)

type BarangInput struct {
	NamaBarang  string  `json:"nama_barang"`
	HargaJual   float64 `json:"harga_jual"`
	HargaBeli   float64 `json:"harga_beli"`
	JumlahStock int     `json:"jumlah_stock"`
}

type UpdateBarangInput struct {
	NamaBarang  string  `json:"nama_barang"`
	HargaJual   float64 `json:"harga_jual"`
	HargaBeli   float64 `json:"harga_beli"`
	JumlahStock int     `json:"jumlah_stock"`
}

func AddProductManual(c *fiber.Ctx) error {
	var input struct {
		KodeBarang  string  `json:"kode_barang"`
		NamaBarang  string  `json:"name"`
		HargaJual   float64 `json:"selling"`
		HargaBeli   float64 `json:"purchase"`
		JumlahStock int     `json:"stock"`
		AsOpening   bool    `json:"as_opening"` // optional: treat as opening balance
	}

	if err := c.BodyParser(&input); err != nil {
		return c.Status(400).JSON(fiber.Map{"message": "Data tidak valid"})
	}

	if input.KodeBarang == "" {
		return c.Status(400).JSON(fiber.Map{"message": "Kode barang wajib diisi"})
	}

	// Coba cari barang (bisa aktif / nonaktif)
	var barangID int
	var isActive int
	err := config.DB.QueryRow(
		`SELECT barang_id, is_active FROM barang WHERE kode_barang=?`,
		input.KodeBarang,
	).Scan(&barangID, &isActive)

	// ==== CASE 1: barang sudah ada ====
	if err == nil {
		tx, err := config.DB.Begin()
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"message": "Gagal mulai transaksi"})
		}
		defer tx.Rollback()

		if isActive == 0 {
			// Reactivate
			if _, err := tx.Exec(
				`UPDATE barang SET is_active=1, nama_barang=?, harga_jual=?, harga_beli=? WHERE barang_id=?`,
				input.NamaBarang, input.HargaJual, input.HargaBeli, barangID,
			); err != nil {
				return c.Status(500).JSON(fiber.Map{"message": "Gagal re-activate barang"})
			}

			// Tambah stok awal jika diisi
			if input.JumlahStock > 0 {
				resBM, err := tx.Exec(
					`INSERT INTO barang_masuk (barang_id, tanggal, jumlah, harga_beli, sisa_stok, keterangan)
					VALUES (?,?,?,?,?,?)`,
					barangID, time.Now(), input.JumlahStock, input.HargaBeli, input.JumlahStock, "Input Manual (reactivate)",
				)
				if err != nil {
					return c.Status(500).JSON(fiber.Map{"message": "Gagal insert barang_masuk"})
				}
				mid, _ := resBM.LastInsertId()

				// Rekap ke stok_riwayat
				now := time.Now()
				nilai := float64(input.JumlahStock) * input.HargaBeli
				qty := float64(input.JumlahStock)
				if input.AsOpening || input.HargaBeli == 0 {
					if err := models.UpsertOpeningBalance(tx, barangID, now, nilai, qty); err != nil {
						return c.Status(500).JSON(fiber.Map{"message": "Gagal update stok_riwayat (opening)", "detail": err.Error()})
					}
				} else {
					if err := models.UpdateStokPembelian(tx, barangID, now, nilai, qty); err != nil {
						return c.Status(500).JSON(fiber.Map{"message": "Gagal update stok_riwayat (pembelian)", "detail": err.Error()})
					}
					if err := models.UpsertJurnalPembelianForMasuk(tx, int(mid), now, nilai, ""); err != nil {
						return c.Status(500).JSON(fiber.Map{"message": "Gagal upsert jurnal pembelian", "detail": err.Error()})
					}
				}
			}

			if err := tx.Commit(); err != nil {
				return c.Status(500).JSON(fiber.Map{"message": "Gagal commit"})
			}
			return c.JSON(fiber.Map{"message": "Barang nonaktif diaktifkan kembali & stok ditambahkan"})
		}

		// Sudah aktif → tolak duplikat
		return c.Status(400).JSON(fiber.Map{"message": "Kode barang sudah digunakan"})
	}

	// Bila error selain "tidak ada baris", balikan error DB
	if err != sql.ErrNoRows {
		return c.Status(500).JSON(fiber.Map{"message": "Gagal cek kode barang"})
	}

	// ==== CASE 2: barang baru ====
	tx, err := config.DB.Begin()
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"message": "Gagal memulai transaksi"})
	}
	defer tx.Rollback()

	res, err := tx.Exec(
		`INSERT INTO barang (kode_barang, nama_barang, harga_jual, harga_beli, jumlah_stock)
		VALUES (?, ?, ?, ?, ?)`,
		input.KodeBarang, input.NamaBarang, input.HargaJual, input.HargaBeli,
		input.JumlahStock,
	)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"message": "Gagal simpan ke tabel barang"})
	}

	lastID64, err := res.LastInsertId()
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"message": "Gagal mendapatkan ID barang"})
	}
	barangID = int(lastID64)

	resBM, err := tx.Exec(
		`INSERT INTO barang_masuk (barang_id, tanggal, jumlah, harga_beli, sisa_stok, keterangan)
		VALUES (?,?,?,?,?,?)`,
		barangID, time.Now(), input.JumlahStock, input.HargaBeli, input.JumlahStock, "Input Manual",
	)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"message": "Gagal simpan ke barang_masuk"})
	}
	mid, _ := resBM.LastInsertId()

	// Rekap ke stok_riwayat
	now := time.Now()
	nilai := float64(input.JumlahStock) * input.HargaBeli
	qty := float64(input.JumlahStock)
	if input.AsOpening || input.HargaBeli == 0 {
		if err := models.UpsertOpeningBalance(tx, barangID, now, nilai, qty); err != nil {
			return c.Status(500).JSON(fiber.Map{"message": "Gagal update stok_riwayat (opening)", "detail": err.Error()})
		}
	} else {
		if err := models.UpdateStokPembelian(tx, barangID, now, nilai, qty); err != nil {
			return c.Status(500).JSON(fiber.Map{"message": "Gagal update stok_riwayat (pembelian)", "detail": err.Error()})
		}

		//input ke jurnal persediaan
		if err := models.UpsertJurnalPembelianForMasuk(tx, int(mid), now, nilai, ""); err != nil {
			return c.Status(500).JSON(fiber.Map{"message": "Gagal upsert jurnal pembelian", "detail": err.Error()})
		}
	}

	if err := tx.Commit(); err != nil {
		return c.Status(500).JSON(fiber.Map{"message": "Gagal commit transaksi"})
	}
	return c.JSON(fiber.Map{"message": "Produk berhasil ditambahkan secara manual"})
}

// GET /api/barang
func GetBarang(c *fiber.Ctx) error {
	// ?include_archived=1 → tampilkan semua (aktif + arsip)
	includeArchived := c.Query("include_archived") == "1"
	base := `
		SELECT 
			b.kode_barang, b.nama_barang, b.harga_jual, b.harga_beli,
			IFNULL(SUM(bm.sisa_stok), 0) AS total_stock
    	FROM barang b
    	LEFT JOIN barang_masuk bm ON b.barang_id = bm.barang_id
		`
	where := " WHERE b.is_active = 1 "
	if includeArchived {
		where = "" // tampilkan semua
	}
	query := base + where + " GROUP BY b.barang_id"

	rows, err := config.DB.Query(query)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Gagal mengambil data barang"})
	}
	defer rows.Close()

	var results []map[string]interface{}

	for rows.Next() {
		var kode, nama string
		var jual, beli float64
		var stock int

		err := rows.Scan(&kode, &nama, &jual, &beli, &stock)
		if err != nil {
			continue
		}
		results = append(results, map[string]interface{}{
			"kode_barang":  kode,
			"nama_barang":  nama,
			"harga_jual":   jual,
			"harga_beli":   beli,
			"jumlah_stock": stock,
		})
	}

	return c.JSON(results)
}

func UpdateBarang(c *fiber.Ctx) error {
	kode := c.Params("kode_barang")
	if kode == "" {
		return c.Status(400).JSON(fiber.Map{"message": "Kode barang diperlukan"})
	}

	var input UpdateBarangInput
	if err := c.BodyParser(&input); err != nil {
		return c.Status(400).JSON(fiber.Map{"message": "Data tidak valid"})
	}

	db := config.DB
	res, err := db.Exec(`UPDATE barang SET nama_barang = ?, harga_jual = ?, harga_beli = ?, jumlah_stock = ? WHERE kode_barang = ?`,
		input.NamaBarang, input.HargaJual, input.HargaBeli, input.JumlahStock, kode)

	if err != nil {
		return c.Status(500).JSON(fiber.Map{"message": "Gagal update data barang"})
	}

	rows, _ := res.RowsAffected()
	if rows == 0 {
		return c.Status(404).JSON(fiber.Map{"message": "Barang tidak ditemukan atau tidak ada perubahan"})
	}

	return c.JSON(fiber.Map{
		"message": "Barang berhasil diupdate",
	})
}

func DeleteBarang(c *fiber.Ctx) error {
	kode := c.Params("kode_barang")
	if kode == "" {
		return c.Status(400).JSON(fiber.Map{"message": "Kode barang diperlukan"})
	}
	tx, err := config.DB.Begin()
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"message": "Gagal mulai transaksi"})
	}
	defer tx.Rollback()

	var barangID int
	err = tx.QueryRow(`SELECT barang_id FROM barang WHERE kode_barang = ?`, kode).Scan(&barangID)
	if err == sql.ErrNoRows {
		return c.Status(404).JSON(fiber.Map{"message": "Barang tidak ditemukan"})
	}
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"message": "DB error"})
	}

	// Soft delete → nonaktifkan
	if _, err := tx.Exec(`UPDATE barang SET is_active = 0 WHERE barang_id = ?`, barangID); err != nil {
		return c.Status(500).JSON(fiber.Map{"message": "Gagal mengarsipkan barang"})
	}
	if err := tx.Commit(); err != nil {
		return c.Status(500).JSON(fiber.Map{"message": "Gagal commit"})
	}
	return c.JSON(fiber.Map{"message": "Barang berhasil diarsipkan (nonaktif)"})
}

func BulkDeleteBarang(c *fiber.Ctx) error {
	var body struct {
		KodeList []string `json:"kode_list"`
	}
	if err := c.BodyParser(&body); err != nil {
		return c.Status(400).JSON(fiber.Map{"message": "Payload tidak valid"})
	}
	if len(body.KodeList) == 0 {
		return c.Status(400).JSON(fiber.Map{"message": "kode_list kosong"})
	}

	// Normalisasi & buang duplikat
	set := make(map[string]struct{})
	clean := make([]string, 0, len(body.KodeList))
	for _, k := range body.KodeList {
		k = strings.TrimSpace(k)
		if k == "" {
			continue
		}
		if _, ok := set[k]; !ok {
			set[k] = struct{}{}
			clean = append(clean, k)
		}
	}
	if len(clean) == 0 {
		return c.Status(400).JSON(fiber.Map{"message": "kode_list tidak valid"})
	}

	tx, err := config.DB.Begin()
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"message": "Gagal mulai transaksi"})
	}
	defer tx.Rollback()

	deleted := 0
	notFound := []string{}
	fail := []string{}

	for _, kode := range clean {
		var barangID int
		err := tx.QueryRow(`SELECT barang_id FROM barang WHERE kode_barang = ?`, kode).Scan(&barangID)
		if err == sql.ErrNoRows {
			notFound = append(notFound, kode)
			continue
		}
		if err != nil {
			fail = append(fail, kode)
			continue
		}

		if _, err := tx.Exec(`UPDATE barang SET is_active=0 WHERE barang_id = ?`, barangID); err != nil {
			fail = append(fail, kode)
			continue
		}

		deleted++
	}

	if err := tx.Commit(); err != nil {
		return c.Status(500).JSON(fiber.Map{"message": "Gagal commit transaksi"})
	}

	return c.JSON(fiber.Map{
		"message":   "Bulk delete selesai",
		"deleted":   deleted,
		"not_found": notFound,
		"failed":    fail,
		"requested": len(clean),
	})
}

// PATCH /api/barang/:kode/restore
func RestoreBarang(c *fiber.Ctx) error {
	kode := c.Params("kode_barang")
	tx, err := config.DB.Begin()
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"message": "Gagal mulai transaksi"})
	}
	defer tx.Rollback()

	res, err := tx.Exec(`UPDATE barang SET is_active=1 WHERE kode_barang=?`, kode)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"message": "Gagal restore"})
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return c.Status(404).JSON(fiber.Map{"message": "Barang tidak ditemukan"})
	}

	if err := tx.Commit(); err != nil {
		return c.Status(500).JSON(fiber.Map{"message": "Gagal commit"})
	}
	return c.JSON(fiber.Map{"message": "Barang berhasil diaktifkan kembali"})
}

// untuk mendapatkan harga beli barang berdasarkan FIFO
func GetFIFOHarga(c *fiber.Ctx) error {
	kode := c.Params("kode_barang")
	if kode == "" {
		return c.Status(400).JSON(fiber.Map{"message": "Kode barang diperlukan"})
	}

	var barangID int
	err := config.DB.QueryRow(`SELECT barang_id FROM barang WHERE kode_barang = ?`, kode).Scan(&barangID)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"message": "Barang tidak ditemukan"})
	}

	var hargaBeli float64
	err = config.DB.QueryRow(`
		SELECT harga_beli FROM barang_masuk 
		WHERE barang_id = ? AND sisa_stok > 0 
		ORDER BY tanggal ASC, masuk_id ASC LIMIT 1`, barangID).Scan(&hargaBeli)

	if err != nil {
		return c.Status(200).JSON(fiber.Map{"harga_beli": 0}) // tidak ada stok tersisa
	}

	return c.JSON(fiber.Map{"harga_beli": hargaBeli})
}

// }

func GetBarangMasukByKode(c *fiber.Ctx) error {
	kode := c.Params("kode_barang")
	if kode == "" {
		return c.Status(400).JSON(fiber.Map{"message": "Kode barang wajib diisi"})
	}

	rows, err := config.DB.Query(`
		SELECT bm.masuk_id, bm.tanggal, bm.jumlah, bm.harga_beli, bm.sisa_stok, bm.keterangan
		FROM barang_masuk bm
		JOIN barang b ON b.barang_id = bm.barang_id
		WHERE b.kode_barang = ?
		ORDER BY bm.tanggal ASC
	`, kode)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"message": "Gagal mengambil data barang_masuk"})
	}
	defer rows.Close()

	var result []map[string]interface{}
	for rows.Next() {
		var id, jumlah, sisa int
		var tgl time.Time
		var harga float64
		var ket string
		rows.Scan(&id, &tgl, &jumlah, &harga, &sisa, &ket)
		result = append(result, map[string]interface{}{
			"masuk_id":   id,
			"tanggal":    tgl.Format("2006-01-02"),
			"jumlah":     jumlah,
			"harga_beli": harga,
			"sisa_stok":  sisa,
			"keterangan": ket,
		})
	}
	return c.JSON(result)
}

func UpdateBarangMasuk(c *fiber.Ctx) error {
	id := c.Params("masuk_id")
	var input struct {
		Tanggal    string  `json:"tanggal"`
		Jumlah     int     `json:"jumlah"`
		HargaBeli  float64 `json:"harga_beli"`
		SisaStok   int     `json:"sisa_stok"`
		Keterangan string  `json:"keterangan"`
	}
	if err := c.BodyParser(&input); err != nil {
		return c.Status(400).JSON(fiber.Map{"message": "Data tidak valid"})
	}

	tx, err := config.DB.Begin()
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"message": "Gagal mulai transaksi"})
	}
	defer tx.Rollback()

	// Ambil kondisi lama: untuk delta stok_riwayat & validasi
	var barangID, oldJumlah int
	var oldHarga float64
	var oldTanggal time.Time
	if err := tx.QueryRow(`
		SELECT bm.barang_id, bm.tanggal, bm.jumlah, bm.harga_beli
		FROM barang_masuk bm
		WHERE bm.masuk_id = ?`, id).Scan(&barangID, &oldTanggal, &oldJumlah, &oldHarga); err != nil {
		return c.Status(404).JSON(fiber.Map{"message": "Data barang_masuk tidak ditemukan"})
	}

	// 🔒 VALIDASI: berapa unit dari batch ini yang sudah terjual?
	var soldQty int
	if err := tx.QueryRow(`
        SELECT COALESCE(SUM(jumlah),0)
        FROM barang_keluar
        WHERE masuk_id = ?`, id).Scan(&soldQty); err != nil {
		return c.Status(500).JSON(fiber.Map{"message": "Gagal membaca penjualan batch", "detail": err.Error()})
	}

	// 1) jumlah baru tidak boleh < yang sudah terjual
	if input.Jumlah < soldQty {
		return c.Status(400).JSON(fiber.Map{
			"message":         "Jumlah tidak boleh lebih kecil dari total yang sudah terjual dari batch ini",
			"sold_from_batch": soldQty,
			"min_allowed":     soldQty,
		})
	}

	// 2) sisa_stok harus = jumlah_baru - terjual
	expectedSisa := input.Jumlah - soldQty
	if input.SisaStok != expectedSisa {
		return c.Status(400).JSON(fiber.Map{
			"message":         "sisa_stok tidak konsisten dengan jumlah dan penjualan batch",
			"sold_from_batch": soldQty,
			"expected_sisa":   expectedSisa,
		})
	}

	// — UPDATE barang_masuk
	if _, err := tx.Exec(`
        UPDATE barang_masuk
        SET tanggal=?, jumlah=?, harga_beli=?, sisa_stok=?, keterangan=?
        WHERE masuk_id=?`,
		input.Tanggal, input.Jumlah, input.HargaBeli, input.SisaStok, input.Keterangan, id); err != nil {
		return c.Status(500).JSON(fiber.Map{"message": "Gagal update barang_masuk"})
	}

	// — Sinkron stok_riwayat (pembelian & qty) via delta seperti sebelumnya
	newTanggal, _ := time.Parse("2006-01-02", input.Tanggal)
	oldNilai := float64(oldJumlah) * oldHarga
	newNilai := float64(input.Jumlah) * input.HargaBeli
	oldQty := float64(oldJumlah)
	newQty := float64(input.Jumlah)

	samePeriod := oldTanggal.Year() == newTanggal.Year() && oldTanggal.Month() == newTanggal.Month()
	if samePeriod {
		deltaNilai := newNilai - oldNilai
		deltaQty := newQty - oldQty
		if err := models.UpdateStokPembelian(tx, barangID, newTanggal, deltaNilai, deltaQty); err != nil {
			return c.Status(500).JSON(fiber.Map{"message": "Gagal update stok_riwayat (delta)", "detail": err.Error()})
		}
	} else {
		if err := models.UpdateStokPembelian(tx, barangID, oldTanggal, -oldNilai, -oldQty); err != nil {
			return c.Status(500).JSON(fiber.Map{"message": "Gagal revert stok_riwayat (periode lama)", "detail": err.Error()})
		}
		if err := models.UpdateStokPembelian(tx, barangID, newTanggal, newNilai, newQty); err != nil {
			return c.Status(500).JSON(fiber.Map{"message": "Gagal apply stok_riwayat (periode baru)", "detail": err.Error()})
		}
	}
	
	// Upsert jurnal pembelian untuk batch ini sesuai nilai baru
	mid, _ := strconv.Atoi(id)
	if err := models.UpsertJurnalPembelianForMasuk(tx, mid, newTanggal, newNilai, ""); err != nil {
		return c.Status(500).JSON(fiber.Map{"message": "Gagal upsert jurnal pembelian", "detail": err.Error()})
	}

	if err := tx.Commit(); err != nil {
		return c.Status(500).JSON(fiber.Map{"message": "Gagal commit"})
	}
	return c.JSON(fiber.Map{"message": "Berhasil update barang masuk"})
}

