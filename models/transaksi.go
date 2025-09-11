package models

import (
	"database/sql"
	"fmt"
	"log"
	"math"
	"time"
	"ynb-backend/utils"
)

const (
	KODE_KAS        = "1-101"
	KODE_PENJUALAN  = "4-101"
	KODE_PERSEDIAAN = "1-104"
	KODE_HPP        = "5-100"
)

var DB *sql.DB // diset di main.go

func ProcessTransaksiFIFO(trx utils.Transaksi) error {
	fmt.Println("DB status OK. Mulai proses transaksi FIFO...")

	if DB == nil {
		return fmt.Errorf("Koneksi DB nil di models.ProcessTransaksiFIFO")
	}

	if err := DB.Ping(); err != nil {
		return fmt.Errorf("Ping DB gagal: %v", err)
	}

	log.Println("⏳ Memulai transaksi FIFO...")

	tx, err := DB.Begin()
	if err != nil {
		return fmt.Errorf("gagal mulai transaksi: %v", err)
	}
	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	if trx.Tanggal == "" {
		return fmt.Errorf("Tanggal transaksi kosong, tidak bisa disimpan")
	}

	// 1. Insert ke penjualan
	penjualanQuery := `INSERT INTO penjualan (tanggal, jam, kasir, metode_bayar, subtotal, bayar, kembalian) VALUES (?, ?, ?, ?, ?, ?, ?)`
	// Konversi jam string ke time.Time
	parsedJam, err := time.Parse("15:04", trx.Jam)
	if err != nil {
		return fmt.Errorf("gagal parsing jam: %v", err)
	}

	//gak jadi dipake
	// // Parse tanggal ke time.Time untuk stok_riwayat
	// tglTx, err := time.Parse("2006-01-02", trx.Tanggal)
	// if err != nil {
	// 	return fmt.Errorf("format tanggal transaksi tidak valid (harus YYYY-MM-DD): %v", err)
	// }

	res, err := tx.Exec(penjualanQuery, trx.Tanggal, parsedJam.Format("15:04:05"), "KASIR 01", trx.Metode, trx.Subtotal, trx.Subtotal, 0)
	if err != nil {
		return fmt.Errorf("insert penjualan error: %v", err)
	}
	penjualanID, _ := res.LastInsertId()

	log.Printf("✅ Penjualan #%d berhasil dimasukkan\n", penjualanID)

	var totalHPP float64
	// parse tanggal transaksi (YYYY-MM-DD)
	trxDate, _ := time.Parse("2006-01-02", trx.Tanggal)

	if err := ensureMonthlyCarryForward(tx, trxDate.Year(), int(trxDate.Month())); err != nil {
		return fmt.Errorf("carry-forward stok gagal: %v", err)
	}

	for _, item := range trx.Items {
		log.Printf("📦 Proses barang: %s x%d\n", item.Nama, item.Jumlah)
		// Cari barang_id
		var barangID int
		// err := tx.QueryRow("SELECT barang_id FROM barang WHERE nama_barang LIKE ?", "%"+item.Nama+"%").Scan(&barangID)
		err := tx.QueryRow(`SELECT barang_id FROM barang WHERE is_active = 1 AND nama_barang LIKE ? LIMIT 1`, "%"+item.Nama+"%").Scan(&barangID)
		if err != nil {
			fmt.Println("Barang tidak ditemukan di DB:", item.Nama)
			continue // Skip jika barang tidak ditemukan
		}

		// Insert ke detail_penjualan
		_, err = tx.Exec(`INSERT INTO detail_penjualan (penjualan_id, barang_id, jumlah, harga_satuan, total) VALUES (?, ?, ?, ?, ?)`,
			penjualanID, barangID, item.Jumlah, item.Harga, item.Harga*float64(item.Jumlah),
		)
		if err != nil {
			return err
		}

		// Sinkronkan harga_jual barang dengan harga di transaksi (update ke latest price)
		rounded := int64(math.Round(item.Harga))
		var current sql.NullInt64
		if err := tx.QueryRow(`SELECT CAST(harga_jual AS SIGNED) FROM barang WHERE barang_id = ?`, barangID).Scan(&current); err != nil {
			return fmt.Errorf("gagal baca harga_jual barang_id %d: %v", barangID, err)
		}
		if !current.Valid || current.Int64 != rounded {
			if _, err := tx.Exec(`UPDATE barang SET harga_jual = ? WHERE barang_id = ?`, rounded, barangID); err != nil {
				return fmt.Errorf("gagal update harga_jual barang_id %d: %v", barangID, err)
			}
			log.Printf("🔁 Update harga_jual barang_id %d → %d\n", barangID, rounded)
		}

		// FIFO: ambil stok lama dari barang_masuk
		jumlahSisa := item.Jumlah
		// HPP akumulasi khusus untuk barang ini (untuk stok_riwayat)
		var hppItem float64
		rows, err := DB.Query(`SELECT masuk_id, harga_beli, sisa_stok FROM barang_masuk 
			WHERE barang_id = ? AND sisa_stok > 0 ORDER BY tanggal ASC`, barangID)
		if err != nil {
			return err
		}
		defer rows.Close() // 🔥 Tambahkan ini segera setelah Query()

		for rows.Next() && jumlahSisa > 0 {
			var masukID, sisa int
			var hargaBeli float64
			// rows.Scan(&masukID, &hargaBeli, &sisa)
			if err := rows.Scan(&masukID, &hargaBeli, &sisa); err != nil {
				return err
			}

			jml := min(jumlahSisa, sisa)

			// Insert barang_keluar
			_, err = tx.Exec(`INSERT INTO barang_keluar (penjualan_id, barang_id, masuk_id, jumlah, harga_beli) VALUES (?, ?, ?, ?, ?)`,
				penjualanID, barangID, masukID, jml, hargaBeli)
			if err != nil {
				return err
			}

			// Update barang_masuk sisa_stok
			_, err = tx.Exec(`UPDATE barang_masuk SET sisa_stok = sisa_stok - ? WHERE masuk_id = ?`, jml, masukID)
			if err != nil {
				return fmt.Errorf("update stok masuk gagal: %v", err)
			}

			totalHPP += float64(jml) * hargaBeli
			hppItem += float64(jml) * hargaBeli
			jumlahSisa -= jml
		}
		rows.Close()
		if err := rows.Err(); err != nil {
			return fmt.Errorf("error saat iterasi stok masuk: %v", err)
		}

		// === Rekap ke stok_riwayat untuk barang ini ===
		if err := UpdateStokPenjualan(tx, barangID, trxDate, hppItem, float64(item.Jumlah)); err != nil {
			return fmt.Errorf("gagal update stok_riwayat (penjualan) barang_id %d: %v", barangID, err)
		}

		if err := CascadeCarryForward(tx, barangID, trxDate); err != nil {
			return fmt.Errorf("gagal cascade carry-forward (penjualan) barang_id %d: %v", barangID, err)
		}
	}

	log.Printf("✅ Total HPP transaksi: %.2f\n", totalHPP)

	// 3. Insert jurnal transaksi
	jurnalQuery := `INSERT INTO jurnal (tanggal, referensi, tipe_jurnal, user_id) VALUES (?, ?, ?, ?)`
	res, err = tx.Exec(jurnalQuery, trx.Tanggal, fmt.Sprintf("PJ-%d", penjualanID), "Penjualan", 1)
	if err != nil {
		return err
	}
	jurnalID, _ := res.LastInsertId()

	// Ambil akun_id
	var akunKas, akunPenjualan, akunPersediaan, akunHPP int

	if err := tx.QueryRow(`SELECT akun_id FROM akun WHERE kode_akun = ?`, KODE_KAS).Scan(&akunKas); err != nil {
		return fmt.Errorf("akun '1-101' (Kas) tidak ditemukan: %v", err)
	}

	if err := tx.QueryRow(`SELECT akun_id FROM akun WHERE kode_akun = ?`, KODE_PENJUALAN).Scan(&akunPenjualan); err != nil {
		return fmt.Errorf("akun 'Penjualan Produk' tidak ditemukan: %v", err)
	}

	if err := tx.QueryRow(`SELECT akun_id FROM akun WHERE kode_akun = ?`, KODE_PERSEDIAAN).Scan(&akunPersediaan); err != nil {
		return fmt.Errorf("akun 'Persediaan' tidak ditemukan: %v", err)
	}

	if err := tx.QueryRow(`SELECT akun_id FROM akun WHERE kode_akun = ?`, KODE_HPP).Scan(&akunHPP); err != nil {
		return fmt.Errorf("akun 'Harga Pokok Penjualan' tidak ditemukan: %v", err)
	}

	// Entri jurnal: Penjualan & HPP
	jurnalDetailQuery := `INSERT INTO jurnal_detail (jurnal_id, akun_id, debit, kredit, keterangan) VALUES (?, ?, ?, ?, ?)`

	// Kas (Dr)
	tx.Exec(jurnalDetailQuery, jurnalID, akunKas, trx.Subtotal, 0, "Penjualan tunai")
	// Penjualan (Cr)
	tx.Exec(jurnalDetailQuery, jurnalID, akunPenjualan, 0, trx.Subtotal, "Penjualan tunai")
	// HPP (Dr)
	tx.Exec(jurnalDetailQuery, jurnalID, akunHPP, totalHPP, 0, "Pengeluaran barang")
	// Persediaan (Cr)
	tx.Exec(jurnalDetailQuery, jurnalID, akunPersediaan, 0, totalHPP, "Pengeluaran barang")

	// return tx.Commit()
	// ✅ Commit transaksi
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit transaksi gagal: %v", err)
	}

	log.Println("🎉 Transaksi berhasil disimpan dan dicatat di jurnal.")
	return nil
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
