package handlers

import (
	// "database/sql"
	// "fmt"
	"strconv"
	"time"
	"ynb-backend/models"

	"github.com/gofiber/fiber/v2"
)

const HPPAccountName = "Harga Pokok Penjualan"

func GetJurnalHandler(c *fiber.Ctx) error {
	year := c.Query("year")
	month := c.Query("month")

	jurnals, err := models.GetJurnalListFiltered(year, month)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Gagal mengambil data jurnal",
		})
	}

	return c.JSON(jurnals)
}

func GetJournalAdjustments(c *fiber.Ctx) error {
	// masih Dummy
	adjustments := []fiber.Map{
		{"date": "2024-04-30", "account": "Supplies Expense", "adjustment": 20000, "description": "End of month adjustment"},
	}
	return c.JSON(adjustments)
}

func GetTrialBalance(c *fiber.Ctx) error {
	year := c.Query("year")
	month := c.Query("month")

	query := `
		SELECT a.nama_akun, a.jenis,
		    SUM(COALESCE(jd.debit,0))  AS total_debit,
		    SUM(COALESCE(jd.kredit,0)) AS total_kredit
		FROM jurnal_detail jd
		JOIN akun   a ON jd.akun_id   = a.akun_id
		JOIN jurnal j ON jd.jurnal_id = j.jurnal_id
		WHERE 1=1
	`
	args := []interface{}{}
	if year != "" && month != "" {
		query += " AND YEAR(j.tanggal)=? AND MONTH(j.tanggal)=? "
		args = append(args, year, month)
	}
	query += " GROUP BY a.akun_id "

	rows, err := models.DB.Query(query, args...)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "DB error: " + err.Error()})
	}
	defer rows.Close()

	var result []fiber.Map
	for rows.Next() {
		var nama, jenis string
		var debit, kredit float64
		if err := rows.Scan(&nama, &jenis, &debit, &kredit); err != nil {
			return err
		}
		result = append(result, fiber.Map{
			"account": nama, "debit": debit, "credit": kredit,
		})
	}
	return c.JSON(result)
}

// helper menghitung COGS (HPP) untuk year,month tertentu dari
// stok_riwayat untuk digunakan pada income statement
func calcCOGSByYM(yearStr, monthStr string) (float64, error) {
	if yearStr == "" || monthStr == "" {
		return 0, nil
	}
	y, _ := strconv.Atoi(yearStr)
	m, _ := strconv.Atoi(monthStr)

	prevY, prevM := y, m-1
	if m == 1 {
		prevY, prevM = y-1, 12
	}

	var awal, pembelian, akhir float64
	err := models.DB.QueryRow(`
        SELECT
            COALESCE((SELECT SUM(stok_akhir) FROM stok_riwayat WHERE tahun=? AND bulan=?), 0) AS awal,
            COALESCE((SELECT SUM(pembelian) FROM stok_riwayat WHERE tahun=? AND bulan=?), 0) AS pembelian,
            COALESCE((SELECT SUM(stok_akhir)  FROM stok_riwayat WHERE tahun=? AND bulan=?), 0) AS akhir
    `, prevY, prevM, y, m, y, m).Scan(&awal, &pembelian, &akhir)
	if err != nil {
		return 0, err
	}
	return awal + pembelian - akhir, nil
}

func GetIncomeStatement(c *fiber.Ctx) error {
	year := c.Query("year")
	month := c.Query("month")

	//ambil menghitung income statement (revenue dan expense) tanpa HPP
	query := `
		SELECT 
			a.nama_akun,
			a.jenis,
			SUM(
				CASE 
					WHEN a.jenis = 'Revenue' THEN COALESCE(jd.kredit,0) - COALESCE(jd.debit,0)
					WHEN a.jenis = 'Expense' THEN COALESCE(jd.debit,0) - COALESCE(jd.kredit,0)
					ELSE 0
				END
			) AS amount
		FROM jurnal_detail jd
		JOIN akun   a ON jd.akun_id   = a.akun_id
		JOIN jurnal j ON jd.jurnal_id = j.jurnal_id
		WHERE a.jenis IN ('Revenue','Expense')
			AND a.nama_akun <> ?
	`
	args := []interface{}{HPPAccountName}
	if year != "" && month != "" {
		query += " AND YEAR(j.tanggal)=? AND MONTH(j.tanggal)=? "
		args = append(args, year, month)
	}
	query += " GROUP BY a.akun_id, a.nama_akun, a.jenis "

	rows, err := models.DB.Query(query, args...)
	if err != nil {
		return err
	}
	defer rows.Close()

	var result []fiber.Map
	for rows.Next() {
		var nama, jenis string
		var amount float64
		if err := rows.Scan(&nama, &jenis, &amount); err != nil {
			return err
		}
		typ := "Revenue"
		if jenis == "Expense" {
			typ = "Expense"
		}
		result = append(result, fiber.Map{"type": typ, "name": nama, "amount": amount})
	}

	// menghitung COGS dari stok_riwayat & tambahkan sebagai HPP
	cogs, err := calcCOGSByYM(year, month)
	if err != nil {
		return fiber.NewError(500, "Gagal hitung COGS: "+err.Error())
	}
	result = append(result, fiber.Map{
		"type":   "Expense",
		"name":   HPPAccountName,
		"amount": cogs,
	})

	return c.JSON(result)
}

func GetBalanceSheet(c *fiber.Ctx) error {
	year := c.Query("year")
	month := c.Query("month")

	query := `
	SELECT a.nama_akun, a.jenis,
    	SUM(
    		CASE
        		WHEN a.jenis = 'Asset'
    				THEN COALESCE(jd.debit,0) - COALESCE(jd.kredit,0)
				WHEN a.jenis IN ('Liability','Equity')
					THEN COALESCE(jd.kredit,0) - COALESCE(jd.debit,0)
			END
    	) AS saldo
	FROM jurnal_detail jd
	JOIN akun   a ON jd.akun_id   = a.akun_id
	JOIN jurnal j ON jd.jurnal_id = j.jurnal_id
	WHERE a.jenis IN ('Asset','Liability','Equity')
	`
	args := []interface{}{}
	if year != "" && month != "" {
		query += " AND YEAR(j.tanggal)=? AND MONTH(j.tanggal)=? "
		args = append(args, year, month)
	}
	query += " GROUP BY a.akun_id, a.nama_akun, a.jenis "

	rows, err := models.DB.Query(query, args...)
	if err != nil {
		return err
	}
	defer rows.Close()

	var result []fiber.Map
	for rows.Next() {
		var nama, jenis string
		var saldo float64
		if err := rows.Scan(&nama, &jenis, &saldo); err != nil {
			return err
		}
		typ := "Liabilities & Equity"
		if jenis == "Asset" {
			typ = "Assets"
		}
		result = append(result, fiber.Map{"type": typ, "name": nama, "amount": saldo})
	}

	// 2) Hitung laba/rugi berjalan periode yang sama menggunakan cogs bukan hpp dari tabel jurnal
	plq := `
		SELECT
			COALESCE(SUM(CASE WHEN a.jenis='Revenue' THEN jd.kredit - jd.debit END),0) AS rev,
			COALESCE(SUM(CASE WHEN a.jenis='Expense' AND a.nama_akun <> ? THEN jd.debit - jd.kredit END),0) AS opex
		FROM jurnal_detail jd
		JOIN akun   a ON jd.akun_id   = a.akun_id
		JOIN jurnal j ON jd.jurnal_id = j.jurnal_id
		WHERE 1=1
	`
	plArgs := []interface{}{HPPAccountName}
	if year != "" && month != "" {
		plq += " AND YEAR(j.tanggal)=? AND MONTH(j.tanggal)=? "
		plArgs = append(plArgs, year, month)
	}

	var rev, opex float64

	if err := models.DB.QueryRow(plq, plArgs...).Scan(&rev, &opex); err != nil {
		return err
	}

	cogs, err := calcCOGSByYM(year, month)
	if err != nil {
		return fiber.NewError(500, "Gagal hitung COGS untuk BS: "+err.Error())
	}

	laba := rev - (opex + cogs)

	// 3) Tambahkan Net Income ke sisi Liabilities & Equity agar balance
	if laba != 0 {
		result = append(result, fiber.Map{
			"type":   "Liabilities & Equity",
			"name":   "Laba/Rugi Berjalan",
			"amount": laba,
		})
	}

	return c.JSON(result)
}

func GetCOGS(c *fiber.Ctx) error {
	year := c.Query("year")
	month := c.Query("month")

	// parse dan cari prevY, prevM
	y, _ := strconv.Atoi(year)
	m, _ := strconv.Atoi(month)
	prevY, prevM := y, m-1
	if m == 1 {
		prevY, prevM = y-1, 12
	}

	var awal, pembelian, akhir float64
	// awal = SUM stok_akhir bulan sebelumnya
	// pembelian & akhir ambil dari bulan berjalan seperti biasa
	if err := models.DB.QueryRow(`
        SELECT
            COALESCE((SELECT SUM(stok_akhir) FROM stok_riwayat WHERE tahun=? AND bulan=?), 0) AS awal,
            COALESCE((SELECT SUM(pembelian) FROM stok_riwayat WHERE tahun=? AND bulan=?), 0) AS pembelian,
            COALESCE((SELECT SUM(stok_akhir)  FROM stok_riwayat WHERE tahun=? AND bulan=?), 0) AS akhir
    `, prevY, prevM, y, m, y, m).Scan(&awal, &pembelian, &akhir); err != nil {
		return fiber.NewError(500, "Error ambil stok_riwayat: "+err.Error())
	}

	return c.JSON(fiber.Map{
		"beginningInventory": awal,
		"purchases":          pembelian,
		"endingInventory":    akhir,
	})
}

func GetInventoryCalculation(c *fiber.Ctx) error {
	rows, err := models.DB.Query(`
		SELECT b.nama_barang,
		    COALESCE(masuk.total_masuk, 0),
		    COALESCE(keluar.total_keluar, 0),
		    (COALESCE(masuk.total_masuk, 0) - COALESCE(keluar.total_keluar, 0)) as akhir
		FROM barang b
		LEFT JOIN (
			SELECT barang_id, SUM(jumlah) AS total_masuk FROM barang_masuk GROUP BY barang_id
		) AS masuk ON b.barang_id = masuk.barang_id
		LEFT JOIN (
			SELECT barang_id, SUM(jumlah) AS total_keluar FROM barang_keluar GROUP BY barang_id
		) AS keluar ON b.barang_id = keluar.barang_id
	`)
	if err != nil {
		return err
	}
	defer rows.Close()

	var result []fiber.Map
	for rows.Next() {
		var nama string
		var masuk, keluar, akhir int
		if err := rows.Scan(&nama, &masuk, &keluar, &akhir); err != nil {
			return err
		}
		result = append(result, fiber.Map{
			"item": nama, "beginning": 0, "purchased": masuk, "sold": keluar, "ending": akhir,
		})
	}
	return c.JSON(result)
}

func GetAvailableYears(c *fiber.Ctx) error {
	rows, err := models.DB.Query(`
		SELECT DISTINCT YEAR(tanggal) AS y
		FROM jurnal
		ORDER BY y DESC`)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	defer rows.Close()

	var years []int
	for rows.Next() {
		var y int
		if err := rows.Scan(&y); err != nil {
			return err
		}
		years = append(years, y)
	}
	// fallback kalau kosong → pakai tahun berjalan
	if len(years) == 0 {
		years = []int{time.Now().Year()}
	}
	return c.JSON(years)
}

func GetAvailableMonths(c *fiber.Ctx) error {
	year := c.Query("year")
	if year == "" {
		year = strconv.Itoa(time.Now().Year())
	}

	rows, err := models.DB.Query(`
		SELECT DISTINCT MONTH(tanggal) AS m
		FROM jurnal
		WHERE YEAR(tanggal)=?
		ORDER BY m ASC`, year)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	defer rows.Close()

	var months []int
	for rows.Next() {
		var m int
		if err := rows.Scan(&m); err != nil {
			return err
		}
		months = append(months, m)
	}
	// fallback kalau kosong → pakai bulan berjalan
	if len(months) == 0 {
		months = []int{int(time.Now().Month())}
	}
	return c.JSON(months)
}
