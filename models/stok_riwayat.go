package models

import (
	"database/sql"
	"time"
)

// Pastikan baris stok_riwayat (barang, tahun, bulan) tersedia.
// Jika belum ada, insert dengan stok_awal = stok_akhir bulan lalu (kalau ada).
func ensureStokRiwayat(tx *sql.Tx, barangID int, t time.Time) error {
	y, m, _ := t.Date()

	var exists int
	if err := tx.QueryRow(`
		SELECT COUNT(*) FROM stok_riwayat
		WHERE barang_id = ? AND tahun = ? AND bulan = ?`,
		barangID, y, int(m)).Scan(&exists); err != nil {
		return err
	}
	if exists > 0 {
		return nil
	}

	// bawa stok akhir & qty bulan sebelumnya jadi stok_awal & qty awal
	prev := t.AddDate(0, -1, 0)
	var prevAkhir, prevQty sql.NullFloat64
	_ = tx.QueryRow(`
		SELECT stok_akhir, qty
		FROM stok_riwayat
		WHERE barang_id = ? AND tahun = ? AND bulan = ?`,
		barangID, prev.Year(), int(prev.Month()),
	).Scan(&prevAkhir, &prevQty)

	awal := 0.0
	qtyAwal := 0.0
	if prevAkhir.Valid {
		awal = prevAkhir.Float64
	}
	if prevQty.Valid {
		qtyAwal = prevQty.Float64
	}

	_, err := tx.Exec(`
		INSERT INTO stok_riwayat (barang_id, tahun, bulan, stok_awal, pembelian, penjualan, stok_akhir, qty)
		VALUES (?, ?, ?, ?, 0, 0, ?, ?)`,
		barangID, y, int(m), awal, awal, qtyAwal)
	return err
}

// models/stok_riwayat.go
func ensureMonthlyCarryForward(tx *sql.Tx, y, m int) error {
	prevY, prevM := y, m-1
	if m == 1 {
		prevY, prevM = y-1, 12
	}

	// ✅ Guard: cek apakah masih ada barang aktif yang BELUM punya baris stok_riwayat bulan ini
	var missing int
	if err := tx.QueryRow(`
        SELECT COUNT(*)
        FROM barang b
        LEFT JOIN stok_riwayat s
                ON s.barang_id = b.barang_id AND s.tahun = ? AND s.bulan = ?
        WHERE b.is_active = 1 AND s.barang_id IS NULL
    `, y, m).Scan(&missing); err != nil {
		return err
	}
	if missing == 0 {
		// Semua barang aktif sudah punya baris bulan ini → tidak perlu insert lagi
		return nil
	}

	// Insert baris bulan ini untuk SEMUA barang aktif yang belum punya baris.
	// stok_awal & stok_akhir awal = stok_akhir bulan sebelumnya (jika ada).
	_, err := tx.Exec(`
        INSERT INTO stok_riwayat (barang_id, tahun, bulan, stok_awal, pembelian, penjualan, stok_akhir, qty)
        SELECT b.barang_id, ?, ?, COALESCE(prev.stok_akhir, 0), 0, 0, COALESCE(prev.stok_akhir, 0), COALESCE(prev.qty, 0)
        FROM barang b
        LEFT JOIN stok_riwayat cur
                ON cur.barang_id = b.barang_id AND cur.tahun = ? AND cur.bulan = ?
        LEFT JOIN stok_riwayat prev
                ON prev.barang_id = b.barang_id AND prev.tahun = ? AND prev.bulan = ?
        WHERE b.is_active = 1 AND cur.barang_id IS NULL
    `, y, m, y, m, prevY, prevM)
	return err
}

// CascadeCarryForward menyesuaikan stok_awal, stok_akhir, dan qty
// untuk SEMUA bulan setelah 'start' (jika barisnya sudah ada),
// agar konsisten dengan stok_akhir & qty bulan sebelumnya yang telah berubah.
func CascadeCarryForward(tx *sql.Tx, barangID int, start time.Time) error {
	// cari periode terakhir yang ada untuk barang ini
	var lastY, lastM int
	err := tx.QueryRow(`
        SELECT tahun, bulan
        FROM stok_riwayat
        WHERE barang_id=?
        ORDER BY tahun DESC, bulan DESC
        LIMIT 1
    `, barangID).Scan(&lastY, &lastM)
	if err == sql.ErrNoRows {
		return nil
	}
	if err != nil {
		return err
	}

	y, mo, _ := start.Date()
	m := int(mo) // <<< cast ke int

	// mulai dari bulan SETELAH 'start'
	if m == 12 {
		y++
		m = 1
	} else {
		m++
	}

	for {
		// stop jika melewati periode terakhir
		if y > lastY || (y == lastY && m > lastM) {
			break
		}

		// cek apakah bulan ini ada barisnya
		var exists int
		if err := tx.QueryRow(`
            SELECT COUNT(*) FROM stok_riwayat
            WHERE barang_id=? AND tahun=? AND bulan=?`,
			barangID, y, m,
		).Scan(&exists); err != nil {
			return err
		}
		if exists == 0 {
			// kalau belum ada baris, berhenti.
			break
		}

		// prev (bulan sebelumnya)
		prevY, prevM := y, m-1
		if m == 1 {
			prevY, prevM = y-1, 12
		}

		// ambil stok_akhir (nilai) & qty bulan sebelumnya (SUDAH updated)
		var prevEndVal, prevQty float64
		if err := tx.QueryRow(`
            SELECT stok_akhir, qty
            FROM stok_riwayat
            WHERE barang_id=? AND tahun=? AND bulan=?`,
			barangID, prevY, prevM,
		).Scan(&prevEndVal, &prevQty); err != nil {
			return err
		}

		// hitung net flow qty bulan ini (masuk - keluar) dari tabel detail
		var masukQty, keluarQty float64
		if err := tx.QueryRow(`
            SELECT COALESCE(SUM(jumlah),0)
            FROM barang_masuk
            WHERE barang_id=? AND YEAR(tanggal)=? AND MONTH(tanggal)=?`,
			barangID, y, m,
		).Scan(&masukQty); err != nil {
			return err
		}
		if err := tx.QueryRow(`
            SELECT COALESCE(SUM(bk.jumlah),0)
            FROM barang_keluar bk
            JOIN penjualan p ON p.penjualan_id = bk.penjualan_id
            WHERE bk.barang_id=? AND YEAR(p.tanggal)=? AND MONTH(p.tanggal)=?`,
			barangID, y, m,
		).Scan(&keluarQty); err != nil {
			return err
		}

		// qty lama di bulan ini (sebelum penyesuaian)
		var curQty float64
		if err := tx.QueryRow(`
            SELECT qty FROM stok_riwayat
            WHERE barang_id=? AND tahun=? AND bulan=?`,
			barangID, y, m,
		).Scan(&curQty); err != nil {
			return err
		}

		// startingQtyOld = qty awal bulan ini sebelum flow (estimasi)
		// Karena: curQty ≈ startingQtyOld + masukQty - keluarQty
		startingQtyOld := curQty - masukQty + keluarQty
		qtyDelta := prevQty - startingQtyOld // selisih awal yang harus dibawa ke semua bulan ke depan

		// set stok_awal = stok_akhir bulan sebelumnya (nilai),
		// stok_akhir = stok_awal + pembelian - penjualan (nilai),
		// qty disesuaikan dengan qtyDelta
		if _, err := tx.Exec(`
            UPDATE stok_riwayat
            SET stok_awal=?,
                stok_akhir = ? + pembelian - penjualan,
                qty = qty + ?
            WHERE barang_id=? AND tahun=? AND bulan=?`,
			prevEndVal, prevEndVal, qtyDelta, barangID, y, int(m),
		); err != nil {
			return err
		}

		// next month
		if m == 12 {
			y++
			m = 1
		} else {
			m++
		}
	}

	return nil
}

// dipanggil saat PEMBELIAN: tambah nilai pembelian (Rp), update stok_akhir
func UpdateStokPembelian(tx *sql.Tx, barangID int, t time.Time, nilaiPembelian float64, jumlahPembelian float64) error {
	if err := ensureStokRiwayat(tx, barangID, t); err != nil {
		return err
	}
	y, mo, _ := t.Date()
	m := int(mo)

	if err := ensureMonthlyCarryForward(tx, y, m); err != nil {
		return err
	}

	// // MySQL mengevaluasi SET dari kiri ke kanan, jadi stok_akhir pakai angka pembelian terbaru
	// _, err := tx.Exec(`
	// 	UPDATE stok_riwayat
	// 	SET pembelian = pembelian + ?,
	// 		stok_akhir = stok_awal + pembelian - penjualan,
	// 		qty = qty + ?
	// 	WHERE barang_id=? AND tahun=? AND bulan=?`,
	// 	nilaiPembelian, jumlahPembelian, barangID, y, m)
	// return err

	// step 1: update akumulasi
	if _, err := tx.Exec(`
		UPDATE stok_riwayat
		SET pembelian = pembelian + ?,
		    qty = qty + ?
		WHERE barang_id = ? AND tahun = ? AND bulan = ?`,
		nilaiPembelian, jumlahPembelian, barangID, y, m); err != nil {
		return err
	}

	// step 2: hitung ulang stok_akhir
	if _, err := tx.Exec(`
		UPDATE stok_riwayat
		SET stok_akhir = stok_awal + pembelian - penjualan
		WHERE barang_id = ? AND tahun = ? AND bulan = ?`,
		barangID, y, m); err != nil {
		return err
	}

	// step 3: cascade ke bulan-bulan setelahnya
	if err := CascadeCarryForward(tx, barangID, t); err != nil {
		return err
	}

	return nil
}

// Rekap penjualan bulan berjalan (nilai HPP & qty terjual) → stok_akhir direcalculate.
func UpdateStokPenjualan(tx *sql.Tx, barangID int, t time.Time, nilaiHPP, jumlahPenjualan float64) error {
	if err := ensureStokRiwayat(tx, barangID, t); err != nil {
		return err
	}
	y, m, _ := t.Date()

	// (opsional) safety
	if err := ensureMonthlyCarryForward(tx, y, int(m)); err != nil {
		return err
	}

	// step 1: update akumulasi
	if _, err := tx.Exec(`
		UPDATE stok_riwayat
		SET penjualan = penjualan + ?,
		    qty = qty - ?
		WHERE barang_id = ? AND tahun = ? AND bulan = ?`,
		nilaiHPP, jumlahPenjualan, barangID, y, int(m)); err != nil {
		return err
	}

	// step 2: hitung ulang stok_akhir
	_, err := tx.Exec(`
		UPDATE stok_riwayat
		SET stok_akhir = stok_awal + pembelian - penjualan
		WHERE barang_id = ? AND tahun = ? AND bulan = ?`,
		barangID, y, int(m))
	return err
}

// (Opsional) untuk import mode "opening", tambahkan stok_awal/qty awal tanpa tercatat sbg pembelian.
func UpsertOpeningBalance(tx *sql.Tx, barangID int, t time.Time, tambahNilai, tambahQty float64) error {
	if err := ensureStokRiwayat(tx, barangID, t); err != nil {
		return err
	}
	y, m, _ := t.Date()

	// Tambah stok_awal & qty awal
	if _, err := tx.Exec(`
		UPDATE stok_riwayat
		SET stok_awal = stok_awal + ?, qty = qty + ?
		WHERE barang_id = ? AND tahun = ? AND bulan = ?`,
		tambahNilai, tambahQty, barangID, y, int(m)); err != nil {
		return err
	}

	_, err := tx.Exec(`
		UPDATE stok_riwayat
		SET stok_akhir = stok_awal + pembelian - penjualan
		WHERE barang_id = ? AND tahun = ? AND bulan = ?`,
		barangID, y, int(m))
	return err
}
