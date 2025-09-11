package models

import (
	"database/sql"
	"fmt"
	"time"
)

const (
	//(tambahin ini jika tidak ada di list yang di transaksi.go)
// KODE_PERSEDIAAN = "1-104"
// KODE_KAS        = "1-101"
// KODE_UTANG    = "2-201"
)

func BackfillPembelianFromBarangMasuk(db *sql.DB, start, end *time.Time, kreditKode string) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	// ambil akun_id
	var akunPersediaan, akunKredit int
	if err = tx.QueryRow(`SELECT akun_id FROM akun WHERE kode_akun=?`, KODE_PERSEDIAAN).Scan(&akunPersediaan); err != nil {
		return fmt.Errorf("akun persediaan tidak ditemukan: %w", err)
	}
	if kreditKode == "" {
		kreditKode = KODE_KAS
	}
	if err = tx.QueryRow(`SELECT akun_id FROM akun WHERE kode_akun=?`, kreditKode).Scan(&akunKredit); err != nil {
		return fmt.Errorf("akun kredit (%s) tidak ditemukan: %w", kreditKode, err)
	}

	// rentang tanggal default = min/max barang_masuk
	var minTgl, maxTgl time.Time
	if err = tx.QueryRow(`SELECT MIN(tanggal), MAX(tanggal) FROM barang_masuk`).Scan(&minTgl, &maxTgl); err != nil {
		return fmt.Errorf("ambil rentang barang_masuk: %w", err)
	}
	from := minTgl
	to := maxTgl
	if start != nil {
		from = *start
	}
	if end != nil {
		to = *end
	}

	for d := from; !d.After(to); d = d.AddDate(0, 0, 1) {
		tgl := d.Format("2006-01-02")
		var total float64
		if err = tx.QueryRow(`
			SELECT COALESCE(SUM(jumlah*harga_beli),0)
			FROM barang_masuk WHERE DATE(tanggal)=?`, tgl).Scan(&total); err != nil {
			return fmt.Errorf("SUM barang_masuk %s: %w", tgl, err)
		}
		if total <= 0 {
			continue
		}

		ref := "BACKFILL-BM-" + d.Format("20060102")

		// idempotent guard
		var cnt int
		if err = tx.QueryRow(`SELECT COUNT(*) FROM jurnal WHERE referensi=?`, ref).Scan(&cnt); err != nil {
			return err
		}
		if cnt > 0 {
			continue
		} // sudah pernah dibuat

		// insert header jurnal
		res, err2 := tx.Exec(
			`INSERT INTO jurnal (tanggal, referensi, tipe_jurnal, user_id)
			VALUES (?,?, 'Pembelian', 1)`,
			tgl, ref,
		)
		if err2 != nil {
			return fmt.Errorf("insert jurnal %s: %w", ref, err2)
		}
		jid, _ := res.LastInsertId()

		// detail: Dr Persediaan, Cr Kas/Utang
		if _, err2 = tx.Exec(
			`INSERT INTO jurnal_detail (jurnal_id, akun_id, debit, kredit, keterangan)
			VALUES (?,?,?,?,?)`,
			jid, akunPersediaan, total, 0, "Backfill pembelian dari barang_masuk",
		); err2 != nil {
			return err2
		}

		if _, err2 = tx.Exec(
			`INSERT INTO jurnal_detail (jurnal_id, akun_id, debit, kredit, keterangan)
			VALUES (?,?,?,?,?)`,
			jid, akunKredit, 0, total, "Backfill pembelian dari barang_masuk",
		); err2 != nil {
			return err2
		}
	}

	if err = tx.Commit(); err != nil {
		return err
	}
	return nil
}
