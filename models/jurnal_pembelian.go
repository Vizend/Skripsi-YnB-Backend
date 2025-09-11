package models

import (
	"database/sql"
	"fmt"
	"time"
)

const (
// KODE_PERSEDIAAN = "1-104"
// KODE_KAS        = "1-101"
)

func CreateJurnalPembelian(tx *sql.Tx, t time.Time, total float64, ref, kreditKode string) error {
	if total <= 0 {
		return nil
	}
	if kreditKode == "" {
		kreditKode = KODE_KAS
	}

	var akunPersediaan, akunKredit int
	if err := tx.QueryRow(`SELECT akun_id FROM akun WHERE kode_akun=?`, KODE_PERSEDIAAN).Scan(&akunPersediaan); err != nil {
		return fmt.Errorf("akun persediaan tidak ditemukan: %w", err)
	}
	if err := tx.QueryRow(`SELECT akun_id FROM akun WHERE kode_akun=?`, kreditKode).Scan(&akunKredit); err != nil {
		return fmt.Errorf("akun kredit (%s) tidak ditemukan: %w", kreditKode, err)
	}

	res, err := tx.Exec(
		`INSERT INTO jurnal (tanggal, referensi, tipe_jurnal, user_id)
		VALUES (?,?, 'Pembelian', 1)`,
		t.Format("2006-01-02"), ref,
	)
	if err != nil {
		return err
	}
	jid, _ := res.LastInsertId()

	// Dr Persediaan
	if _, err := tx.Exec(
		`INSERT INTO jurnal_detail (jurnal_id, akun_id, debit, kredit, keterangan)
		VALUES (?,?,?,?,?)`,
		jid, akunPersediaan, total, 0, "Pembelian persediaan",
	); err != nil {
		return err
	}
	// Cr Kas / Utang
	if _, err := tx.Exec(
		`INSERT INTO jurnal_detail (jurnal_id, akun_id, debit, kredit, keterangan)
		VALUES (?,?,?,?,?)`,
		jid, akunKredit, 0, total, "Pembelian persediaan",
	); err != nil {
		return err
	}
	return nil
}


// Buat/ubah jurnal pembelian untuk 1 batch barang_masuk
func UpsertJurnalPembelianForMasuk(tx *sql.Tx, masukID int, t time.Time, total float64, kreditKode string) error {
	if kreditKode == "" {
		kreditKode = KODE_KAS
	}

	var akunPersediaan, akunKredit int
	if err := tx.QueryRow(`SELECT akun_id FROM akun WHERE kode_akun=?`, KODE_PERSEDIAAN).Scan(&akunPersediaan); err != nil {
		return fmt.Errorf("akun persediaan tidak ditemukan: %w", err)
	}
	if err := tx.QueryRow(`SELECT akun_id FROM akun WHERE kode_akun=?`, kreditKode).Scan(&akunKredit); err != nil {
		return fmt.Errorf("akun kredit (%s) tidak ditemukan: %w", kreditKode, err)
	}

	ref := fmt.Sprintf("PMB-BM-%d", masukID)

	var jurnalID int
	err := tx.QueryRow(`SELECT jurnal_id FROM jurnal WHERE referensi=?`, ref).Scan(&jurnalID)
	switch {
	case err == sql.ErrNoRows:
		if total <= 0 {
			return nil 
		}
		res, err := tx.Exec(
			`INSERT INTO jurnal (tanggal, referensi, tipe_jurnal, user_id)
			VALUES (?,?, 'Pembelian', 1)`,
			t.Format("2006-01-02"), ref,
		)
		if err != nil {
			return err
		}
		id64, _ := res.LastInsertId()
		jurnalID = int(id64)

		// Dr Persediaan
		if _, err := tx.Exec(
			`INSERT INTO jurnal_detail (jurnal_id, akun_id, debit, kredit, keterangan)
			VALUES (?,?,?,?,?)`,
			jurnalID, akunPersediaan, total, 0, "Pembelian persediaan",
		); err != nil {
			return err
		}
		// Cr Kas/Utang
		if _, err := tx.Exec(
			`INSERT INTO jurnal_detail (jurnal_id, akun_id, debit, kredit, keterangan)
			VALUES (?,?,?,?,?)`,
			jurnalID, akunKredit, 0, total, "Pembelian persediaan",
		); err != nil {
			return err
		}
		return nil

	case err != nil:
		return err

	default:
		if total <= 0 {
			// hapus jurnal jika sekarang nilainya 0
			if _, err := tx.Exec(`DELETE FROM jurnal_detail WHERE jurnal_id=?`, jurnalID); err != nil {
				return err
			}
			if _, err := tx.Exec(`DELETE FROM jurnal WHERE jurnal_id=?`, jurnalID); err != nil {
				return err
			}
			return nil
		}
		// update tanggal & detail (hapus lalu buat ulang agar sederhana/aman)
		if _, err := tx.Exec(`UPDATE jurnal SET tanggal=? WHERE jurnal_id=?`,
			t.Format("2006-01-02"), jurnalID); err != nil {
			return err
		}
		if _, err := tx.Exec(`DELETE FROM jurnal_detail WHERE jurnal_id=?`, jurnalID); err != nil {
			return err
		}
		if _, err := tx.Exec(
			`INSERT INTO jurnal_detail (jurnal_id, akun_id, debit, kredit, keterangan)
			VALUES (?,?,?,?,?)`,
			jurnalID, akunPersediaan, total, 0, "Pembelian persediaan",
		); err != nil {
			return err
		}
		if _, err := tx.Exec(
			`INSERT INTO jurnal_detail (jurnal_id, akun_id, debit, kredit, keterangan)
			VALUES (?,?,?,?,?)`,
			jurnalID, akunKredit, 0, total, "Pembelian persediaan",
		); err != nil {
			return err
		}
		return nil
	}
}
