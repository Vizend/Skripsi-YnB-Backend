package models

import (
	"database/sql"
	"fmt"
)

type JurnalDetail struct {
	AkunID     int     `json:"akun_id"`
	NamaAkun   string  `json:"nama_akun"`
	Debit      float64 `json:"debit"`
	Kredit     float64 `json:"kredit"`
	Keterangan string  `json:"keterangan"`
}

type Jurnal struct {
	JurnalID   int            `json:"jurnal_id"`
	Tanggal    string         `json:"tanggal"`
	Referensi  string         `json:"referensi"`
	TipeJurnal string         `json:"tipe_jurnal"`
	UserID     int            `json:"user_id"`
	Details    []JurnalDetail `json:"details"`
}

func ensureTimeFormat(timeStr string) string {
	// Jika formatnya hanya HH:MM, tambahkan :00 (detik)
	if len(timeStr) == 5 {
		return timeStr + ":00"
	}
	return timeStr
}

func GetJurnalList() ([]Jurnal, error) {
	rows, err := DB.Query(`SELECT j.jurnal_id, j.tanggal, j.referensi, j.tipe_jurnal, j.user_id, p.jam
		FROM jurnal j
		LEFT JOIN penjualan p ON j.referensi = CONCAT('PJ-', p.penjualan_id)
		ORDER BY j.tanggal ASC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []Jurnal

	for rows.Next() {
		var j Jurnal
		var rawDate string
		var rawTime sql.NullString
		// if err := rows.Scan(&j.JurnalID, &j.Tanggal, &j.Referensi, &j.TipeJurnal, &j.UserID); err != nil {
		// 	return nil, err
		// }

		if err := rows.Scan(&j.JurnalID, &rawDate, &j.Referensi, &j.TipeJurnal, &j.UserID, &rawTime); err != nil {
			return nil, err
		}

		// Format ISO 8601 hanya jika rawTime valid
		if rawTime.Valid {
			// Ambil hanya bagian tanggal tanpa waktu jika rawDate sudah dalam format "2025-06-01T00:00:00Z"
			datePart := rawDate
			if split := len(rawDate) > 10 && rawDate[10] == 'T'; split {
				datePart = rawDate[:10]
			}
			j.Tanggal = fmt.Sprintf("%sT%sZ", datePart, ensureTimeFormat(rawTime.String))

		} else {
			j.Tanggal = rawDate
		}

		// if rawTime.Valid {
		// 	j.Tanggal = fmt.Sprintf("%sT%s:00Z", rawDate, rawTime.String)
		// } else {
		// 	j.Tanggal = rawDate
		// }

		// Ambil detail untuk jurnal ini
		detailRows, err := DB.Query(`
			SELECT jd.akun_id, a.nama_akun, jd.debit, jd.kredit, jd.keterangan
			FROM jurnal_detail jd
			JOIN akun a ON jd.akun_id = a.akun_id
			WHERE jd.jurnal_id = ?
		`, j.JurnalID)
		if err != nil {
			return nil, err
		}

		for detailRows.Next() {
			var d JurnalDetail
			if err := detailRows.Scan(&d.AkunID, &d.NamaAkun, &d.Debit, &d.Kredit, &d.Keterangan); err != nil {
				detailRows.Close()
				return nil, err
			}
			j.Details = append(j.Details, d)
		}
		detailRows.Close()

		result = append(result, j)
	}

	return result, nil
}

func GetJurnalListFiltered(year, month string) ([]Jurnal, error) {
	query := `
		SELECT j.jurnal_id, j.tanggal, j.referensi, j.tipe_jurnal, j.user_id, p.jam
		FROM jurnal j
		LEFT JOIN penjualan p ON j.referensi = CONCAT('PJ-', p.penjualan_id)
		WHERE 1=1`

	params := []interface{}{}
	if year != "" {
		query += " AND YEAR(j.tanggal) = ?"
		params = append(params, year)
	}
	if month != "" {
		query += " AND MONTH(j.tanggal) = ?"
		params = append(params, month)
	}
	query += " ORDER BY j.tanggal ASC"

	rows, err := DB.Query(query, params...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []Jurnal
	for rows.Next() {
		var j Jurnal
		var rawDate string
		var rawTime sql.NullString
		if err := rows.Scan(&j.JurnalID, &rawDate, &j.Referensi, &j.TipeJurnal, &j.UserID, &rawTime); err != nil {
			return nil, err
		}

		if rawTime.Valid {
			// Ambil hanya bagian tanggal tanpa waktu jika rawDate sudah dalam format "2025-06-01T00:00:00Z"
			datePart := rawDate
			if split := len(rawDate) > 10 && rawDate[10] == 'T'; split {
				datePart = rawDate[:10]
			}
			j.Tanggal = fmt.Sprintf("%sT%sZ", datePart, ensureTimeFormat(rawTime.String))

		} else {
			j.Tanggal = rawDate
		}

		detailRows, err := DB.Query(`
			SELECT jd.akun_id, a.nama_akun, jd.debit, jd.kredit, jd.keterangan
			FROM jurnal_detail jd
			JOIN akun a ON jd.akun_id = a.akun_id
			WHERE jd.jurnal_id = ?
		`, j.JurnalID)
		if err != nil {
			return nil, err
		}
		for detailRows.Next() {
			var d JurnalDetail
			if err := detailRows.Scan(&d.AkunID, &d.NamaAkun, &d.Debit, &d.Kredit, &d.Keterangan); err != nil {
				detailRows.Close()
				return nil, err
			}
			j.Details = append(j.Details, d)
		}
		detailRows.Close()
		result = append(result, j)
	}
	return result, nil
}
