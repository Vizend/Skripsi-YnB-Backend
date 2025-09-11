// handlers/stok_riwayat.go
package handlers

import (
	"github.com/gofiber/fiber/v2"
	"ynb-backend/models"
)

func GetStokRiwayatHandler(c *fiber.Ctx) error {
	year := c.Query("year")
	month := c.Query("month")

	rows, err := models.DB.Query(`
		SELECT sr.barang_id, b.kode_barang, b.nama_barang,
		        sr.stok_awal, sr.pembelian, sr.penjualan, sr.stok_akhir, sr.qty
		FROM stok_riwayat sr
		JOIN barang b ON b.barang_id = sr.barang_id
		WHERE sr.tahun = ? AND sr.bulan = ?
		ORDER BY b.nama_barang`,
		year, month)
	if err != nil {
		return fiber.NewError(500, "Error ambil stok_riwayat: "+err.Error())
	}
	defer rows.Close()

	var out []fiber.Map
	for rows.Next() {
		var id int
		var kode, nama string
		var awal, beli, jual, akhir, qty float64
		if err := rows.Scan(&id, &kode, &nama, &awal, &beli, &jual, &akhir, &qty); err != nil {
			return err
		}
		out = append(out, fiber.Map{
			"barang_id":  id,
			"kode":       kode,
			"nama":       nama,
			"stok_awal":  awal,
			"pembelian":  beli,
			"penjualan":  jual,
			"stok_akhir": akhir,
			"qty":        qty,
		})
	}
	return c.JSON(out)
}
