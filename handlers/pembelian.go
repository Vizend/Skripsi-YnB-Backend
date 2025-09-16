package handlers

import (
	"fmt"
	"time"
	// "ynb-backend/controllers"
	"ynb-backend/models"

	"github.com/gofiber/fiber/v2"
)

func CreatePembelianManual(c *fiber.Ctx) error {
	var input models.PembelianManual
	if err := c.BodyParser(&input); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Input tidak valid"})
	}

	if input.HargaSatuan <= 0 {
		return c.Status(400).JSON(fiber.Map{"error": "Harga satuan harus lebih besar dari 0"})
	}
	if input.Jumlah <= 0 {
		return c.Status(400).JSON(fiber.Map{"error": "Jumlah harus lebih besar dari 0"})
	}

	tx, err := models.DB.Begin()
	if err != nil {
		return err
	}

	// Insert pembelian
	res, err := tx.Exec(`INSERT INTO pembelian (tanggal, total) VALUES (?, ?)`, input.Tanggal, input.Total)
	if err != nil {
		tx.Rollback()
		return c.Status(500).JSON(fiber.Map{"error": "Gagal insert pembelian", "detail": err.Error()})
	}
	pembelianID, _ := res.LastInsertId()

	var barangID int
	var isActive int
	err = tx.QueryRow(`SELECT barang_id, is_active FROM barang WHERE kode_barang = ?`, input.KodeBarang).Scan(&barangID, &isActive)

	if err != nil {
		// jika barang tidak ditemukan maka akan buat barang baru
		createRes, err := tx.Exec(`INSERT into barang (kode_barang, nama_barang, harga_beli, jumlah_stock) 
									VALUES (?, ?, ?, ?)`,
			input.KodeBarang, input.NamaBarang, input.HargaSatuan, input.Jumlah)
		if err != nil {
			tx.Rollback()
			return c.Status(500).JSON(fiber.Map{"error": "Gagal insert barang", "detail": err.Error()})
		}
		barangID64, _ := createRes.LastInsertId()
		barangID = int(barangID64)
	} else {
		if isActive == 0 {
			if _, err := tx.Exec(`UPDATE barang SET is_active=1 WHERE barang_id=?`, barangID); err != nil {
				tx.Rollback()
				return c.Status(500).JSON(fiber.Map{"error": "Gagal re-activate barang"})
			}
		}
	}

	// Insert detail pembelian
	_, err = tx.Exec(`
		INSERT INTO detail_pembelian (pembelian_id, kode_barang, nama_barang, barang_id, jumlah, harga_satuan, total)
		VALUES (?, ?, ?, ?, ?, ?, ?)`,
		pembelianID, input.KodeBarang, input.NamaBarang, barangID, input.Jumlah, input.HargaSatuan, input.Total)
	if err != nil {
		fmt.Println("ERROR:", err)
		tx.Rollback()
		return c.Status(500).JSON(fiber.Map{"error": "Internal Server Error", "detail": err.Error()})
	}

	Keterangan := "Pembelian manual " + input.NamaBarang
	// Insert ke barang_masuk
	_, err = tx.Exec(`
	INSERT INTO barang_masuk (barang_id, jumlah, sisa_stok, harga_beli, tanggal, pembelian_id, keterangan)
	VALUES (?, ?, ?, ?, ?, ?, ?)`,
		barangID, input.Jumlah, input.Jumlah, input.HargaSatuan, input.Tanggal, pembelianID, Keterangan)

	if err != nil {
		fmt.Println("ERROR:", err)
		tx.Rollback()
		return c.Status(500).JSON(fiber.Map{"error": "Internal Server Error", "detail": err.Error()})
	}

	// Insert ke stok_riwayat
	tgl, _ := time.Parse("2006-01-02", input.Tanggal)
	nilaiPembelian := float64(input.Jumlah) * input.HargaSatuan
	if err := models.UpdateStokPembelian(tx, barangID, tgl, nilaiPembelian, float64(input.Jumlah)); err != nil {
		tx.Rollback()
		return c.Status(500).JSON(fiber.Map{"error": "Gagal update stok_riwayat", "detail": err.Error()})
	}

	//user
	userID := 1

	jurnalRes, err := tx.Exec(`
		INSERT INTO jurnal (tanggal, referensi, tipe_jurnal, user_id)
		VALUES (?, ?, ?, ?)`,
		input.Tanggal, fmt.Sprintf("PB-%d", pembelianID), "Pembelian", userID)
	if err != nil {
		fmt.Println("ERROR insert jurnal:", err)
		tx.Rollback()
		return err
	}
	jurnalID, _ := jurnalRes.LastInsertId()

	// Ambil akun_id dari tabel akun
	var akunPersediaan, akunKas int
	err = tx.QueryRow(`SELECT akun_id FROM akun WHERE kode_akun = '1-104'`).Scan(&akunPersediaan)
	if err != nil {
		fmt.Println("ERROR ambil akun persediaan:", err)
		tx.Rollback()
		return err
	}
	err = tx.QueryRow(`SELECT akun_id FROM akun WHERE kode_akun = '1-101'`).Scan(&akunKas)
	if err != nil {
		fmt.Println("ERROR ambil akun kas:", err)
		tx.Rollback()
		return err
	}

	// Tambahkan detail jurnal
	_, err = tx.Exec(`
		INSERT INTO jurnal_detail (jurnal_id, akun_id, debit, kredit, keterangan)
		VALUES (?, ?, ?, 0, ?)`,
		jurnalID, akunPersediaan, input.Total, "Pembelian barang")
	if err != nil {
		tx.Rollback()
		return err
	}

	_, err = tx.Exec(`
		INSERT INTO jurnal_detail (jurnal_id, akun_id, debit, kredit, keterangan)
		VALUES (?, ?, 0, ?, ?)`,
		jurnalID, akunKas, input.Total, "Pembayaran pembelian")
	if err != nil {
		tx.Rollback()
		return err
	}

	tx.Commit()
	return c.JSON(fiber.Map{"message": "Pembelian berhasil disimpan"})
}
