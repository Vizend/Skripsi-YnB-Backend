package controllers

import (
	"database/sql"
	"fmt"
	"github.com/gofiber/fiber/v2"
	"strings"
	"time"
	"ynb-backend/config"
	"ynb-backend/models"
)

func UploadBarangCSV(c *fiber.Ctx) error {
	var items []models.BarangCSVInput
	if err := c.BodyParser(&items); err != nil {
		return c.Status(400).JSON(fiber.Map{"message": "Format JSON tidak valid"})
	}

	fmt.Println("🚀 Menerima upload barang CSV")
	fmt.Println("Jumlah items:", len(items))

	db := config.DB
	// melakukan validasi agar csv aja yang diinput
	fileName := c.Get("X-Filename")
	if fileName == "" {
		fileName = fmt.Sprintf("upload-%d.csv", time.Now().Unix())
	}
	if !strings.HasSuffix(strings.ToLower(fileName), ".csv") {
		return c.Status(400).JSON(fiber.Map{"message": "Hanya file .csv yang diperbolehkan"})
	}

	// Validasi duplikat di payload
	kodeMap := make(map[string]bool)
	for _, item := range items {
		if kodeMap[item.KodeBarang] {
			return c.Status(400).JSON(fiber.Map{
				"message": fmt.Sprintf("Duplikat kode_barang dalam file: %s", item.KodeBarang),
			})
		}
		kodeMap[item.KodeBarang] = true
	}

	// Mulai transaksi
	tx, err := db.Begin()
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"message": "Gagal mulai transaksi"})
	}
	defer tx.Rollback() // aman-kan transaksi bila ada error sebelum Commit

	var hasil []map[string]string
	mode := strings.ToLower(c.Query("mode", c.Get("X-Import-Mode"))) // "opening" atau ""
	now := time.Now()

	for _, item := range items {
		var existingID int
		var isActive int
		err := tx.QueryRow("SELECT barang_id, is_active FROM barang WHERE kode_barang = ?", item.KodeBarang).
			Scan(&existingID, &isActive)

		switch {
		case err == sql.ErrNoRows:
			// Barang baru
			resB, err := tx.Exec(`
				INSERT INTO barang (kode_barang, nama_barang, harga_jual, harga_beli, jumlah_stock)
				VALUES (?, ?, ?, ?, ?)`,
				item.KodeBarang, item.NamaBarang, item.HargaJual, item.HargaBeli, item.JumlahStock,
			)
			if err != nil {
				return c.Status(500).JSON(fiber.Map{"message": "Gagal insert barang"})
			}
			lastID, _ := resB.LastInsertId()

			resBM, err := tx.Exec(`
				INSERT INTO barang_masuk (barang_id, tanggal, jumlah, harga_beli, sisa_stok, keterangan)
				VALUES (?,?,?,?,?,?)`,
				lastID, now, item.JumlahStock, item.HargaBeli, item.JumlahStock, "Upload CSV Baru",
			)
			if err != nil {
				return c.Status(500).JSON(fiber.Map{"message": "Gagal insert barang_masuk"})
			}
			mid, _ := resBM.LastInsertId()

			// insert ke stok_riwayat
			qty := float64(item.JumlahStock)
			nilai := qty * item.HargaBeli // jika harga_beli 0 → hanya qty yang ter-rekap
			if mode == "opening" {
				if err := models.UpsertOpeningBalance(tx, int(lastID), now, nilai, qty); err != nil {
					return c.Status(500).JSON(fiber.Map{"message": "Gagal update stok_riwayat (opening)"})
				}
			} else {
				if err := models.UpdateStokPembelian(tx, int(lastID), now, nilai, qty); err != nil {
					return c.Status(500).JSON(fiber.Map{"message": "Gagal update stok_riwayat (pembelian)"})
				}
				// jurnal per batch (akan skip bila nilai==0)
				if err := models.UpsertJurnalPembelianForMasuk(tx, int(mid), now, nilai, ""); err != nil {
					return c.Status(500).JSON(fiber.Map{"message": "Gagal upsert jurnal pembelian", "detail": err.Error()})
				}
			}

			hasil = append(hasil, map[string]string{
				"kode_barang": item.KodeBarang,
				"nama_barang": item.NamaBarang,
				"status":      "Barang baru ditambahkan",
			})

		case err == nil:
			if isActive == 0 {
				if _, err := tx.Exec(`UPDATE barang SET is_active=1 WHERE barang_id=?`, existingID); err != nil {
					return c.Status(500).JSON(fiber.Map{"message": "Gagal re-activate barang"})
				}
			}

			// Barang sudah ada
			if _, err := tx.Exec(`UPDATE barang SET jumlah_stock = jumlah_stock + ? WHERE barang_id = ?`,
				item.JumlahStock, existingID,
			); err != nil {
				return c.Status(500).JSON(fiber.Map{"message": "Gagal update stok barang"})
			}

			resBM, err := tx.Exec(`
				INSERT INTO barang_masuk (barang_id, tanggal, jumlah, harga_beli, sisa_stok, keterangan)
				VALUES (?,?,?,?,?,?)`,
				existingID, now, item.JumlahStock, item.HargaBeli, item.JumlahStock, "Upload CSV Tambahan",
			)
			if err != nil {
				return c.Status(500).JSON(fiber.Map{"message": "Gagal insert barang_masuk"})
			}
			mid, _ := resBM.LastInsertId()

			// insert ke stok_riwayat
			qty := float64(item.JumlahStock)
			nilai := qty * item.HargaBeli
			if mode == "opening" {
				if err := models.UpsertOpeningBalance(tx, existingID, now, nilai, qty); err != nil {
					return c.Status(500).JSON(fiber.Map{"message": "Gagal update stok_riwayat (opening)"})
				}
			} else {
				if err := models.UpdateStokPembelian(tx, existingID, now, nilai, qty); err != nil {
					return c.Status(500).JSON(fiber.Map{"message": "Gagal update stok_riwayat (pembelian)"})
				}
				// jurnal per batch (akan skip bila nilai==0)
				if err := models.UpsertJurnalPembelianForMasuk(tx, int(mid), now, nilai, ""); err != nil {
					return c.Status(500).JSON(fiber.Map{"message": "Gagal upsert jurnal pembelian", "detail": err.Error()})
				}
			}

			hasil = append(hasil, map[string]string{
				"kode_barang": item.KodeBarang,
				"nama_barang": item.NamaBarang,
				"status":      "Stok barang ditambahkan",
			})

		default:
			return c.Status(500).JSON(fiber.Map{
				"message": fmt.Sprintf("Error membaca barang: %s", item.KodeBarang),
			})
		}
	}

	if err := tx.Commit(); err != nil {
		return c.Status(500).JSON(fiber.Map{"message": "Gagal commit transaksi"})
	}

	return c.JSON(fiber.Map{
		"message": "Upload berhasil",
		"result":  hasil,
	})
}
