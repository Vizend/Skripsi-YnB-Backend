package models

import (
	"database/sql"
	"strings"
)

func CheckDuplicateOnDate(date, ref string) (bool, error) {
	if strings.TrimSpace(ref) == "" {
		return false, nil
	}
	var n int
	err := DB.QueryRow(`SELECT COUNT(*) FROM penjualan WHERE tanggal=? AND referensi_xjd=?`, date, ref).Scan(&n)
	return n > 0, err
}
func FindBarangIDForPreview(tx *sql.Tx, name string, price float64) (int, *barangCandidate, error) {
	return findBarangID(tx, name, price)
}
