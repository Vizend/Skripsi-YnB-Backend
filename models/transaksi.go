package models

import (
	"database/sql"
	"fmt"
	"math"
	"strings"
	"time"
	"ynb-backend/utils"
)

type barangCandidate struct {
	ID        int
	Nama      string
	HargaJual int64 // asumsi kolom integer/decimal dibaca ke int64
	Score     float64
	PriceDiff int64
}

type ProcessOpts struct {
	DryRun bool // untuk preview
}

const (
	KODE_KAS        = "1-101"
	KODE_BANK       = "1-102"
	KODE_PENJUALAN  = "4-101"
	KODE_PERSEDIAAN = "1-104"
	KODE_HPP        = "5-100"
)

var DB *sql.DB // diset di main.go

func ProcessTransaksiFIFO(trx utils.Transaksi) error {
	return ProcessTransaksiFIFOWithOpts(trx, ProcessOpts{DryRun: false})
}

func ProcessTransaksiFIFOWithOpts(trx utils.Transaksi, opts ProcessOpts) (retErr error) {
	if DB == nil {
		return fmt.Errorf("DB nil")
	}
	if err := DB.Ping(); err != nil {
		return fmt.Errorf("ping DB gagal: %v", err)
	}
	if trx.Tanggal == "" {
		return fmt.Errorf("Tanggal kosong")
	}

	tx, err := DB.Begin()
	if err != nil {
		return fmt.Errorf("begin tx gagal: %v", err)
	}
	defer func() {
		if retErr != nil || opts.DryRun {
			_ = tx.Rollback()
		} else {
			if err := tx.Commit(); err != nil {
				retErr = fmt.Errorf("commit gagal: %v", err)
			}
		}
	}()

	// --- Cek duplikat per (tanggal, refno) ---
	if trx.RefNo != "" {
		var n int
		if err := tx.QueryRow(`SELECT COUNT(*) FROM penjualan WHERE tanggal = ? AND referensi_xjd = ?`, trx.Tanggal, trx.RefNo).Scan(&n); err != nil {
			return fmt.Errorf("cek duplikat gagal: %v", err)
		}
		if n > 0 {
			return fmt.Errorf("duplikat: transaksi dengan refno %s pada %s sudah ada", trx.RefNo, trx.Tanggal)
		}
	}

	// --- Insert penjualan ---
	// fallback bayar/kembalian bila parser tidak mengisi
	bayar := trx.Bayar
	if bayar <= 0 {
		bayar = trx.Subtotal
	}
	kemb := trx.Kembalian
	if kemb < 0 {
		kemb = 0
	}

	parsedJam, err := time.Parse("15:04", trx.Jam)
	if err != nil {
		return fmt.Errorf("jam invalid: %v", err)
	}

	penjualanQuery := `INSERT INTO penjualan (tanggal, jam, kasir, metode_bayar, subtotal, bayar, kembalian, referensi_xjd)
                    	VALUES (?, ?, ?, ?, ?, ?, ?, ?)`
	res, err := tx.Exec(penjualanQuery, trx.Tanggal, parsedJam.Format("15:04:05"), "KASIR 01", trx.Metode, trx.Subtotal, bayar, kemb, nullIfEmpty(trx.RefNo))
	if err != nil {
		return fmt.Errorf("insert penjualan: %v", err)
	}
	penjualanID, _ := res.LastInsertId()

	trxDate, _ := time.Parse("2006-01-02", trx.Tanggal)
	if err := ensureMonthlyCarryForward(tx, trxDate.Year(), int(trxDate.Month())); err != nil {
		return fmt.Errorf("carry-forward stok gagal: %v", err)
	}

	var totalHPP float64

	for _, it := range trx.Items {

		// --- SEBELUM cari barang, tentukan harga_satuan yang benar ---
		unit := it.Harga
		if it.Jumlah > 1 {
			unit = it.Harga / float64(it.Jumlah) // TXT pakai total baris
		}

		// --- cari barang ---
		barangID, _, err := findBarangID(tx, it.Nama, unit)
		if err != nil {
			// barang tidak ketemu => untuk DryRun: cukup warning, untuk commit: gagal atau lewati?
			if opts.DryRun {
				// cukup teruskan; preview akan laporkan item unmatched
				continue
			}
			// kamu bisa pilih: return error (fail-fast) atau skip.
			// Di sini aku pilih fail-fast agar HPP tidak salah.
			return fmt.Errorf("barang tidak ditemukan: %s (harga %.0f)", it.Nama, it.Harga)
		}

		// sinkron harga_jual → set ke harga terbaru dari TXT
		rounded := int64(math.Round(unit))
		var current sql.NullInt64
		if err := tx.QueryRow(`SELECT CAST(harga_jual AS SIGNED) FROM barang WHERE barang_id = ?`, barangID).Scan(&current); err != nil {
			return fmt.Errorf("read harga_jual: %v", err)
		}

		priceDiff := abs64(current.Int64 - rounded)
		tolAbs := int64(2000)
		tolPct := int64(math.Round(float64(rounded) * 0.05))
		tol := tolAbs
		if tolPct > tol {
			tol = tolPct
		}
		if priceDiff <= tol {
			_, _ = tx.Exec(`UPDATE barang SET harga_jual = ? WHERE barang_id = ?`, rounded, barangID)
		}

		if opts.DryRun {
			// preview tidak memodifikasi stok & detail
			continue
		}

		// detail_penjualan
		if _, err := tx.Exec(`INSERT INTO detail_penjualan (penjualan_id, barang_id, jumlah, harga_satuan, total)
                            	VALUES (?,?,?,?,?)`,
			penjualanID, barangID, it.Jumlah, unit, unit*float64(it.Jumlah)); err != nil {
			return fmt.Errorf("insert detail_penjualan: %v", err)
		}

		// --- Ambil semua lot stok dulu, tutup rows, baru mutasi (hindari bad connection) ---
		type lot struct {
			masukID   int
			hargaBeli float64
			sisa      int
		}

		q := `SELECT masuk_id, harga_beli, sisa_stok
				FROM barang_masuk
				WHERE barang_id = ? AND sisa_stok > 0
				ORDER BY tanggal ASC
				`
		rows, err := tx.Query(q, barangID)
		if err != nil {
			return fmt.Errorf("query stok masuk: %w", err)
		}

		var lots []lot
		for rows.Next() {
			var l lot
			if err := rows.Scan(&l.masukID, &l.hargaBeli, &l.sisa); err != nil {
				rows.Close()
				return fmt.Errorf("scan stok masuk: %w", err)
			}
			lots = append(lots, l)
		}
		if err := rows.Close(); err != nil {
			return fmt.Errorf("close rows stok masuk: %w", err)
		}
		if err := rows.Err(); err != nil {
			return fmt.Errorf("iter stok masuk: %w", err)
		}

		// --- Mutasi stok & catat HPP setelah rows tertutup ---
		jumlahSisa := it.Jumlah
		var hppItem float64
		for _, l := range lots {
			if jumlahSisa <= 0 {
				break
			}
			jml := min(jumlahSisa, l.sisa)
			if jml <= 0 {
				continue
			}

			if _, err := tx.Exec(
				`INSERT INTO barang_keluar (penjualan_id, barang_id, masuk_id, jumlah, harga_beli)
        		VALUES (?,?,?,?,?)`,
				penjualanID, barangID, l.masukID, jml, l.hargaBeli,
			); err != nil {
				return fmt.Errorf("insert barang_keluar: %w", err)
			}

			if _, err := tx.Exec(
				`UPDATE barang_masuk SET sisa_stok = sisa_stok - ? WHERE masuk_id = ?`,
				jml, l.masukID,
			); err != nil {
				return fmt.Errorf("update sisa_stok: %w", err)
			}

			totalHPP += float64(jml) * l.hargaBeli
			hppItem += float64(jml) * l.hargaBeli
			jumlahSisa -= jml
		}

		if err := UpdateStokPenjualan(tx, barangID, trxDate, hppItem, float64(it.Jumlah)); err != nil {
			return fmt.Errorf("stok_riwayat update: %v", err)
		}
		if err := CascadeCarryForward(tx, barangID, trxDate); err != nil {
			return fmt.Errorf("cascade carry-forward: %v", err)
		}
	}

	if opts.DryRun {
		return nil
	}

	// jurnal
	var akunKas, akunBank, akunPenj, akunPers, akunHPP int
	if err := tx.QueryRow(`SELECT akun_id FROM akun WHERE kode_akun=?`, KODE_KAS).Scan(&akunKas); err != nil {
		return err
	}
	if err := tx.QueryRow(`SELECT akun_id FROM akun WHERE kode_akun=?`, KODE_BANK).Scan(&akunBank); err != nil {
		return err
	}
	if err := tx.QueryRow(`SELECT akun_id FROM akun WHERE kode_akun=?`, KODE_PENJUALAN).Scan(&akunPenj); err != nil {
		return err
	}
	if err := tx.QueryRow(`SELECT akun_id FROM akun WHERE kode_akun=?`, KODE_PERSEDIAAN).Scan(&akunPers); err != nil {
		return err
	}
	if err := tx.QueryRow(`SELECT akun_id FROM akun WHERE kode_akun=?`, KODE_HPP).Scan(&akunHPP); err != nil {
		return err
	}

	// Net cash-in: bayar - kembalian; fallback ke subtotal jika tidak valid/negatif
	kasMasuk := bayar - kemb
	if kasMasuk <= 0 {
		kasMasuk = trx.Subtotal
	}

	// Pilih akun debit berdasarkan metode bayar
	pay := strings.ToUpper(strings.TrimSpace(trx.Metode))
	debitAkun := akunKas
	switch pay {
	case "CASH":
		debitAkun = akunKas
	case "BCA", "QRIS", "DEBIT", "TRANSFER":
		debitAkun = akunBank
	default:
		// fallback kalau ada metode baru yang belum ditangani
		debitAkun = akunKas
	}

	res, err = tx.Exec(`INSERT INTO jurnal (tanggal, referensi, tipe_jurnal, user_id)
                        VALUES (?,?,?,?)`,
		trx.Tanggal, fmt.Sprintf("PJ-%d", penjualanID), "Penjualan", 1)
	if err != nil {
		return err
	}
	jurnalID, _ := res.LastInsertId()

	det := `INSERT INTO jurnal_detail (jurnal_id, akun_id, debit, kredit, keterangan) VALUES (?,?,?,?,?)`
	// Debit kas
	tx.Exec(det, jurnalID, debitAkun, kasMasuk, 0, "Penjualan "+pay)
	// Kredit penjualan
	tx.Exec(det, jurnalID, akunPenj, 0, trx.Subtotal, "Penjualan -"+pay)
	// Debit HPP
	tx.Exec(det, jurnalID, akunHPP, totalHPP, 0, "HPP FIFO")
	// Kredit persediaan
	tx.Exec(det, jurnalID, akunPers, 0, totalHPP, "Persediaan - FIFO")

	return nil
}

func nullIfEmpty(s string) interface{} {
	if strings.TrimSpace(s) == "" {
		return nil
	}
	return s
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

const (
	tolAbsDefault    = int64(2000)
	tolPctNearest    = 0.05 // dulu 0.08 -> diperkecil supaya 46.5k tidak nyangkut ke 50k
	minSimPriceMatch = 0.50 // minimal similarity untuk exact/nearest price
	minSimLike       = 0.60 // minimal similarity untuk fallback LIKE
)

func findBarangID(tx *sql.Tx, rawName string, price float64) (int, *barangCandidate, error) {
	name := strings.TrimSpace(rawName)
	rounded := int64(math.Round(price))

	// 1) Exact name (case-insensitive)
	var exactID int
	if err := tx.QueryRow(
		`SELECT barang_id FROM barang WHERE is_active=1 AND LOWER(nama_barang)=LOWER(?) LIMIT 1`, name,
	).Scan(&exactID); err == nil {
		return exactID, &barangCandidate{ID: exactID, Nama: name, HargaJual: rounded, Score: 1.0, PriceDiff: 0}, nil
	}

	// helper untuk memilih kandidat terbaik dari hasil query
	pickBest := func(rows *sql.Rows, preferByPrice bool) (*barangCandidate, error) {
		defer rows.Close()
		var best *barangCandidate
		var bestAbsDiff int64 = 1 << 62
		for rows.Next() {
			var cand barangCandidate
			if err := rows.Scan(&cand.ID, &cand.Nama, &cand.HargaJual); err != nil {
				return nil, err
			}
			cand.PriceDiff = cand.HargaJual - rounded
			cand.Score = simRatio(name, cand.Nama)
			absd := abs64(cand.PriceDiff)

			if best == nil {
				tmp := cand
				best = &tmp
				bestAbsDiff = absd
				continue
			}
			if preferByPrice {
				// Utamakan harga lebih dekat; jika sama, utamakan kemiripan nama
				if absd < bestAbsDiff || (absd == bestAbsDiff && cand.Score > best.Score) {
					tmp := cand
					best = &tmp
					bestAbsDiff = absd
				}
			} else {
				// Utamakan kemiripan nama; jika sama, harga lebih dekat
				if cand.Score > best.Score || (cand.Score == best.Score && absd < bestAbsDiff) {
					tmp := cand
					best = &tmp
					bestAbsDiff = absd
				}
			}
		}
		if err := rows.Err(); err != nil {
			return nil, err
		}
		return best, nil
	}

	// 2) Exact price terlebih dahulu
	rows, err := tx.Query(
		`SELECT barang_id, nama_barang, CAST(harga_jual AS SIGNED)
        FROM barang
        WHERE is_active=1 AND CAST(harga_jual AS SIGNED) = ?
        LIMIT 50`, rounded)
	if err != nil {
		return 0, nil, err
	}
	if best, err := pickBest(rows, false /* preferByPrice? no, semua sama harga */); err != nil {
		return 0, nil, err
	} else if best != nil && best.Score >= minSimPriceMatch {
		return best.ID, best, nil
	}
	// catatan: kalau best==nil atau similarity < ambang, lanjut ke nearest price

	// 3) Nearest price dalam toleransi (max{±2.000, ±5%})
	tolPct := int64(math.Round(float64(rounded) * tolPctNearest))
	tol := tolAbsDefault
	if tolPct > tol {
		tol = tolPct
	}

	rows, err = tx.Query(
		`SELECT barang_id, nama_barang, CAST(harga_jual AS SIGNED)
        FROM barang
        WHERE is_active=1 
        AND ABS(CAST(harga_jual AS SIGNED) - ?) <= ?
        ORDER BY ABS(CAST(harga_jual AS SIGNED) - ?) ASC
        LIMIT 50`, rounded, tol, rounded)
	if err != nil {
		return 0, nil, err
	}
	if best, err := pickBest(rows, true /* preferByPrice */); err != nil {
		return 0, nil, err
	} else if best != nil && best.Score >= minSimPriceMatch {
		return best.ID, best, nil
	}

	// 4) Fallback terakhir: LIKE (tetap rank dengan nama > harga)
	rows, err = tx.Query(
		`SELECT barang_id, nama_barang, CAST(harga_jual AS SIGNED)
        FROM barang
        WHERE is_active=1 AND nama_barang LIKE ?
        LIMIT 50`, "%"+name+"%")
	if err != nil {
		return 0, nil, err
	}
	if best, err := pickBest(rows, false /* prefer name similarity */); err != nil {
		return 0, nil, err
	} else if best != nil && best.Score >= minSimLike {
		return best.ID, best, nil
	}

	return 0, nil, sql.ErrNoRows
}
func abs64(x int64) int64 {
	if x < 0 {
		return -x
	}
	return x
}
