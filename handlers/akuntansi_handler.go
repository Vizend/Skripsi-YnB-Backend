package handlers

import (
	// "database/sql"
	"bytes"
	// "encoding/csv"
	"fmt"
	"strconv"
	"time"
	"ynb-backend/models"

	"github.com/gofiber/fiber/v2"
	"github.com/xuri/excelize/v2"
)

const HPPAccountName = "Harga Pokok Penjualan"

func sendXLSX(c *fiber.Ctx, f *excelize.File, filename string) error {
	buf, err := f.WriteToBuffer()
	if err != nil {
		return fiber.NewError(500, "Gagal membuat XLSX: "+err.Error())
	}
	c.Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	c.Attachment(filename)
	return c.SendStream(bytes.NewReader(buf.Bytes()))
}

func makeStyles(f *excelize.File) (header, number int, err error) {
	header, err = f.NewStyle(&excelize.Style{Font: &excelize.Font{Bold: true}})
	if err != nil {
		return
	}
	// Built-in NumFmt 3 = #,##0
	custom := "[$Rp-421] #,##0" //untuk rupiah
	number, err = f.NewStyle(&excelize.Style{CustomNumFmt: &custom})
	return
}

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
	yearStr := c.Query("year")
	monthStr := c.Query("month")

	// fallback kalau kosong -> bulan berjalan
	now := time.Now()
	y, _ := strconv.Atoi(yearStr)
	m, _ := strconv.Atoi(monthStr)
	if y == 0 {
		y = now.Year()
	}
	if m == 0 {
		m = int(now.Month())
	}

	// bulan sebelumnya untuk beginning
	prevY, prevM := y, m-1
	if m == 1 {
		prevY, prevM = y-1, 12
	}

	rows, err := models.DB.Query(`
		SELECT 
			b.nama_barang,
			COALESCE(beg.qty, 0) AS beginning,
			COALESCE(pur.qty, 0) AS purchased,
			COALESCE(sld.qty, 0) AS sold,
			(COALESCE(beg.qty,0) + COALESCE(pur.qty,0) - COALESCE(sld.qty,0)) AS ending
		FROM barang b
		LEFT JOIN (
			SELECT barang_id, qty 
			FROM stok_riwayat 
			WHERE tahun = ? AND bulan = ?
		) beg ON beg.barang_id = b.barang_id
		LEFT JOIN (
			SELECT barang_id, SUM(jumlah) AS qty
			FROM barang_masuk
			WHERE YEAR(tanggal) = ? AND MONTH(tanggal) = ?
			GROUP BY barang_id
		) pur ON pur.barang_id = b.barang_id
		LEFT JOIN (
			SELECT bk.barang_id, SUM(bk.jumlah) AS qty
			FROM barang_keluar bk
			JOIN penjualan p ON p.penjualan_id = bk.penjualan_id
			WHERE YEAR(p.tanggal) = ? AND MONTH(p.tanggal) = ?
			GROUP BY bk.barang_id
		) sld ON sld.barang_id = b.barang_id
		WHERE b.is_active = 1
		ORDER BY b.nama_barang ASC
	`, prevY, prevM, y, m, y, m)
	if err != nil {
		return err
	}
	defer rows.Close()

	var result []fiber.Map
	for rows.Next() {
		var nama string
		var beginning, purchased, sold, ending float64
		if err := rows.Scan(&nama, &beginning, &purchased, &sold, &ending); err != nil {
			return err
		}
		result = append(result, fiber.Map{
			"item":      nama,
			"beginning": beginning,
			"purchased": purchased,
			"sold":      sold,
			"ending":    ending,
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

// === Excel: Journal Entries ===
func ExportJournalEntriesXLSX(c *fiber.Ctx) error {
	year := c.Query("year")
	month := c.Query("month")

	list, err := models.GetJurnalListFiltered(year, month)
	if err != nil {
		return fiber.NewError(500, "Gagal ambil jurnal: "+err.Error())
	}

	f := excelize.NewFile()
	sheet := "Journal Entries"
	f.SetSheetName("Sheet1", sheet)
	header, number, _ := makeStyles(f)

	cols := []string{"Date", "Reference", "Journal Type", "Account", "Debit", "Credit", "Description"}
	for i, h := range cols {
		col, _ := excelize.ColumnNumberToName(i + 1)
		f.SetCellValue(sheet, col+"1", h)
	}
	f.SetCellStyle(sheet, "A1", "G1", header)

	r := 2
	for _, j := range list {
		date := j.Tanggal
		if len(date) >= 10 {
			date = date[:10]
		} // YYYY-MM-DD
		for _, d := range j.Details {
			f.SetCellValue(sheet, fmt.Sprintf("A%d", r), date)
			f.SetCellValue(sheet, fmt.Sprintf("B%d", r), j.Referensi)
			f.SetCellValue(sheet, fmt.Sprintf("C%d", r), j.TipeJurnal)
			f.SetCellValue(sheet, fmt.Sprintf("D%d", r), d.NamaAkun)
			f.SetCellValue(sheet, fmt.Sprintf("E%d", r), d.Debit)
			f.SetCellValue(sheet, fmt.Sprintf("F%d", r), d.Kredit)
			f.SetCellValue(sheet, fmt.Sprintf("G%d", r), d.Keterangan)
			r++
		}
	}
	if r > 2 {
		f.SetCellStyle(sheet, "E2", fmt.Sprintf("F%d", r-1), number)
	}
	f.SetColWidth(sheet, "A", "D", 20)
	f.SetColWidth(sheet, "E", "F", 16)
	f.SetColWidth(sheet, "G", "G", 40)
	f.SetPanes(sheet, &excelize.Panes{Freeze: true, Split: true, YSplit: 1})

	return sendXLSX(c, f, fmt.Sprintf("journal_%s_%s.xlsx", year, month))
}

// === Excel: Trial Balance ===
func ExportTrialBalanceXLSX(c *fiber.Ctx) error {
	year := c.Query("year")
	month := c.Query("month")

	q := `
        SELECT a.nama_akun, a.jenis,
            SUM(COALESCE(jd.debit,0))  AS total_debit,
            SUM(COALESCE(jd.kredit,0)) AS total_kredit
        FROM jurnal_detail jd
        JOIN akun a   ON jd.akun_id = a.akun_id
        JOIN jurnal j ON jd.jurnal_id = j.jurnal_id
        WHERE 1=1`
	args := []interface{}{}
	if year != "" && month != "" {
		q += " AND YEAR(j.tanggal)=? AND MONTH(j.tanggal)=? "
		args = append(args, year, month)
	}
	q += " GROUP BY a.akun_id ORDER BY a.nama_akun"

	rows, err := models.DB.Query(q, args...)
	if err != nil {
		return fiber.NewError(500, "DB error: "+err.Error())
	}
	defer rows.Close()

	f := excelize.NewFile()
	sheet := "Trial Balance"
	f.SetSheetName("Sheet1", sheet)
	header, number, _ := makeStyles(f)

	// header
	f.SetCellValue(sheet, "A1", "Account")
	f.SetCellValue(sheet, "B1", "Debit")
	f.SetCellValue(sheet, "C1", "Credit")
	f.SetCellStyle(sheet, "A1", "C1", header)

	r := 2
	for rows.Next() {
		var nama, jenis string
		var debit, kredit float64
		if err := rows.Scan(&nama, &jenis, &debit, &kredit); err != nil {
			return err
		}
		f.SetCellValue(sheet, fmt.Sprintf("A%d", r), nama)
		f.SetCellValue(sheet, fmt.Sprintf("B%d", r), debit)
		f.SetCellValue(sheet, fmt.Sprintf("C%d", r), kredit)
		r++
	}

	// style kolom angka
	if r > 2 {
		f.SetCellStyle(sheet, "B2", fmt.Sprintf("B%d", r-1), number)
		f.SetCellStyle(sheet, "C2", fmt.Sprintf("C%d", r-1), number)
	}

	// kosmetik
	f.SetColWidth(sheet, "A", "A", 28)
	f.SetColWidth(sheet, "B", "C", 18)
	f.SetPanes(sheet, &excelize.Panes{Freeze: true, Split: true, XSplit: 0, YSplit: 1})

	return sendXLSX(c, f, fmt.Sprintf("trial_balance_%s_%s.xlsx", year, month))
}

// === Excel: Income Statement ===
func ExportIncomeStatementXLSX(c *fiber.Ctx) error {
	year := c.Query("year")
	month := c.Query("month")

	q := `
    	SELECT a.nama_akun, a.jenis,
            SUM(CASE
                WHEN a.jenis='Revenue' THEN COALESCE(jd.kredit,0)-COALESCE(jd.debit,0)
                WHEN a.jenis='Expense' THEN COALESCE(jd.debit,0)-COALESCE(jd.kredit,0)
                ELSE 0 END) AS amount
    	FROM jurnal_detail jd
    	JOIN akun a   ON jd.akun_id=a.akun_id
    	JOIN jurnal j ON jd.jurnal_id=j.jurnal_id
    	WHERE a.jenis IN ('Revenue','Expense') AND a.nama_akun <> ?
    `
	args := []interface{}{HPPAccountName}
	if year != "" && month != "" {
		q += " AND YEAR(j.tanggal)=? AND MONTH(j.tanggal)=? "
		args = append(args, year, month)
	}
	q += " GROUP BY a.akun_id,a.nama_akun,a.jenis ORDER BY a.jenis,a.nama_akun"

	rows, err := models.DB.Query(q, args...)
	if err != nil {
		return err
	}
	defer rows.Close()

	f := excelize.NewFile()
	sheet := "Income Statement"
	f.SetSheetName("Sheet1", sheet)
	header, number, _ := makeStyles(f)

	f.SetCellValue(sheet, "A1", "Type")
	f.SetCellValue(sheet, "B1", "Account")
	f.SetCellValue(sheet, "C1", "Amount")
	f.SetCellStyle(sheet, "A1", "C1", header)

	r := 2
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
		f.SetCellValue(sheet, fmt.Sprintf("A%d", r), typ)
		f.SetCellValue(sheet, fmt.Sprintf("B%d", r), nama)
		f.SetCellValue(sheet, fmt.Sprintf("C%d", r), amount)
		r++
	}

	// Tambah baris COGS (HPP)
	cogs, err := calcCOGSByYM(year, month)
	if err != nil {
		return fiber.NewError(500, "Gagal hitung COGS: "+err.Error())
	}
	f.SetCellValue(sheet, fmt.Sprintf("A%d", r), "Expense")
	f.SetCellValue(sheet, fmt.Sprintf("B%d", r), HPPAccountName)
	f.SetCellValue(sheet, fmt.Sprintf("C%d", r), cogs)

	// Format angka untuk semua data
	lastData := r // baris terakhir data (termasuk COGS)
	f.SetCellStyle(sheet, "C2", fmt.Sprintf("C%d", lastData), number)

	// Baris ringkasan
	r = lastData + 2 // beri 1 baris kosong lalu ringkasan di bawahnya

	// Total Revenue
	f.SetCellValue(sheet, fmt.Sprintf("B%d", r), "Total Revenue")
	f.SetCellFormula(sheet, fmt.Sprintf("C%d", r),
		fmt.Sprintf(`SUMIF(A2:A%d,"Revenue",C2:C%d)`, lastData, lastData))
	f.SetCellStyle(sheet, fmt.Sprintf("B%d", r), fmt.Sprintf("B%d", r), header)
	f.SetCellStyle(sheet, fmt.Sprintf("C%d", r), fmt.Sprintf("C%d", r), number)
	totalRevRow := r
	r++

	// Total Expense (termasuk COGS)
	f.SetCellValue(sheet, fmt.Sprintf("B%d", r), "Total Expense")
	f.SetCellFormula(sheet, fmt.Sprintf("C%d", r),
		fmt.Sprintf(`SUMIF(A2:A%d,"Expense",C2:C%d)`, lastData, lastData))
	f.SetCellStyle(sheet, fmt.Sprintf("B%d", r), fmt.Sprintf("B%d", r), header)
	f.SetCellStyle(sheet, fmt.Sprintf("C%d", r), fmt.Sprintf("C%d", r), number)
	totalExpRow := r
	r++

	// Net Income
	f.SetCellValue(sheet, fmt.Sprintf("B%d", r), "Net Income")
	f.SetCellFormula(sheet, fmt.Sprintf("C%d", r),
		fmt.Sprintf("C%d-C%d", totalRevRow, totalExpRow))
	f.SetCellStyle(sheet, fmt.Sprintf("B%d", r), fmt.Sprintf("B%d", r), header)
	f.SetCellStyle(sheet, fmt.Sprintf("C%d", r), fmt.Sprintf("C%d", r), number)

	// kosmetik
	f.SetColWidth(sheet, "A", "B", 28)
	f.SetColWidth(sheet, "C", "C", 18)
	f.SetPanes(sheet, &excelize.Panes{Freeze: true, Split: true, YSplit: 1})

	return sendXLSX(c, f, fmt.Sprintf("income_statement_%s_%s.xlsx", year, month))
}

// === Excel: Balance Sheet ===
func ExportBalanceSheetXLSX(c *fiber.Ctx) error {
	year := c.Query("year")
	month := c.Query("month")

	q := `
    	SELECT a.nama_akun, a.jenis,
        	SUM(CASE WHEN a.jenis='Asset' THEN COALESCE(jd.debit,0)-COALESCE(jd.kredit,0)
                WHEN a.jenis IN ('Liability','Equity') THEN COALESCE(jd.kredit,0)-COALESCE(jd.debit,0)
            END) AS saldo
    	FROM jurnal_detail jd
    	JOIN akun a   ON jd.akun_id=a.akun_id
    	JOIN jurnal j ON jd.jurnal_id=j.jurnal_id
    	WHERE a.jenis IN ('Asset','Liability','Equity')`
	args := []interface{}{}
	if year != "" && month != "" {
		q += " AND YEAR(j.tanggal)=? AND MONTH(j.tanggal)=? "
		args = append(args, year, month)
	}
	q += " GROUP BY a.akun_id,a.nama_akun,a.jenis ORDER BY a.jenis,a.nama_akun"

	rows, err := models.DB.Query(q, args...)
	if err != nil {
		return err
	}
	defer rows.Close()

	f := excelize.NewFile()
	sheet := "Balance Sheet"
	f.SetSheetName("Sheet1", sheet)
	header, number, _ := makeStyles(f)

	f.SetCellValue(sheet, "A1", "Category")
	f.SetCellValue(sheet, "B1", "Account")
	f.SetCellValue(sheet, "C1", "Amount")
	f.SetCellStyle(sheet, "A1", "C1", header)

	r := 2
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
		f.SetCellValue(sheet, fmt.Sprintf("A%d", r), typ)
		f.SetCellValue(sheet, fmt.Sprintf("B%d", r), nama)
		f.SetCellValue(sheet, fmt.Sprintf("C%d", r), saldo)
		r++
	}

	// Hitung Net Income periode ini dan tambahkan
	plq := `
        SELECT
        	COALESCE(SUM(CASE WHEN a.jenis='Revenue' THEN jd.kredit - jd.debit END),0) AS rev,
        	COALESCE(SUM(CASE WHEN a.jenis='Expense' AND a.nama_akun <> ? THEN jd.debit - jd.kredit END),0) AS opex
        FROM jurnal_detail jd
        JOIN akun a   ON jd.akun_id=a.akun_id
        JOIN jurnal j ON jd.jurnal_id=j.jurnal_id
        WHERE 1=1`
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
		return fiber.NewError(500, "Gagal hitung COGS: "+err.Error())
	}
	laba := rev - (opex + cogs)
	if laba != 0 {
		f.SetCellValue(sheet, fmt.Sprintf("A%d", r), "Liabilities & Equity")
		f.SetCellValue(sheet, fmt.Sprintf("B%d", r), "Laba/Rugi Berjalan")
		f.SetCellValue(sheet, fmt.Sprintf("C%d", r), laba)
		r++
	}

	if r > 2 {
		f.SetCellStyle(sheet, "C2", fmt.Sprintf("C%d", r-1), number)
	}
	f.SetColWidth(sheet, "A", "B", 28)
	f.SetColWidth(sheet, "C", "C", 18)
	f.SetPanes(sheet, &excelize.Panes{Freeze: true, Split: true, YSplit: 1})

	return sendXLSX(c, f, fmt.Sprintf("balance_sheet_%s_%s.xlsx", year, month))
}

// === Excel: COGS ===
func ExportCOGSXLSX(c *fiber.Ctx) error {
	year := c.Query("year")
	month := c.Query("month")

	y, _ := strconv.Atoi(year)
	m, _ := strconv.Atoi(month)
	prevY, prevM := y, m-1
	if m == 1 {
		prevY, prevM = y-1, 12
	}

	var awal, pembelian, akhir float64
	if err := models.DB.QueryRow(`
        SELECT
			COALESCE((SELECT SUM(stok_akhir) FROM stok_riwayat WHERE tahun=? AND bulan=?),0),
			COALESCE((SELECT SUM(pembelian)  FROM stok_riwayat WHERE tahun=? AND bulan=?),0),
        	COALESCE((SELECT SUM(stok_akhir) FROM stok_riwayat WHERE tahun=? AND bulan=?),0)
    `, prevY, prevM, y, m, y, m).Scan(&awal, &pembelian, &akhir); err != nil {
		return fiber.NewError(500, "Error ambil stok_riwayat: "+err.Error())
	}
	cogs := awal + pembelian - akhir

	f := excelize.NewFile()
	sheet := "COGS"
	f.SetSheetName("Sheet1", sheet)
	header, number, _ := makeStyles(f)

	headers := []string{"Beginning Inventory", "Purchases", "Ending Inventory", "COGS"}
	for i, h := range headers {
		col, _ := excelize.ColumnNumberToName(i + 1)
		f.SetCellValue(sheet, col+"1", h)
	}
	f.SetCellStyle(sheet, "A1", "D1", header)

	f.SetCellValue(sheet, "A2", awal)
	f.SetCellValue(sheet, "B2", pembelian)
	f.SetCellValue(sheet, "C2", akhir)
	f.SetCellValue(sheet, "D2", cogs)
	f.SetCellStyle(sheet, "A2", "D2", number)
	f.SetColWidth(sheet, "A", "D", 22)

	return sendXLSX(c, f, fmt.Sprintf("cogs_%s_%s.xlsx", year, month))
}

// === Excel: Inventory Calculation (qty) ===
func ExportInventoryCalculationXLSX(c *fiber.Ctx) error {
	yearStr := c.Query("year")
	monthStr := c.Query("month")

	now := time.Now()
	y, _ := strconv.Atoi(yearStr)
	m, _ := strconv.Atoi(monthStr)
	if y == 0 {
		y = now.Year()
	}
	if m == 0 {
		m = int(now.Month())
	}
	prevY, prevM := y, m-1
	if m == 1 {
		prevY, prevM = y-1, 12
	}

	rows, err := models.DB.Query(`
        SELECT b.nama_barang,
			COALESCE(beg.qty,0) AS beginning,
            COALESCE(pur.qty,0) AS purchased,
            COALESCE(sld.qty,0) AS sold,
            (COALESCE(beg.qty,0)+COALESCE(pur.qty,0)-COALESCE(sld.qty,0)) AS ending
        FROM barang b
        LEFT JOIN (SELECT barang_id, qty FROM stok_riwayat WHERE tahun=? AND bulan=?) beg ON beg.barang_id=b.barang_id
        LEFT JOIN (SELECT barang_id, SUM(jumlah) AS qty FROM barang_masuk WHERE YEAR(tanggal)=? AND MONTH(tanggal)=? GROUP BY barang_id) pur ON pur.barang_id=b.barang_id
        LEFT JOIN (SELECT bk.barang_id, SUM(bk.jumlah) AS qty FROM barang_keluar bk JOIN penjualan p ON p.penjualan_id=bk.penjualan_id
            WHERE YEAR(p.tanggal)=? AND MONTH(p.tanggal)=? GROUP BY bk.barang_id) sld ON sld.barang_id=b.barang_id
        WHERE b.is_active=1
        ORDER BY b.nama_barang ASC
    `, prevY, prevM, y, m, y, m)
	if err != nil {
		return err
	}
	defer rows.Close()

	f := excelize.NewFile()
	sheet := "Inventory Calc"
	f.SetSheetName("Sheet1", sheet)
	header, number, _ := makeStyles(f)

	heads := []string{"Item", "Beginning", "Purchased", "Sold", "Ending"}
	for i, h := range heads {
		col, _ := excelize.ColumnNumberToName(i + 1)
		f.SetCellValue(sheet, col+"1", h)
	}
	f.SetCellStyle(sheet, "A1", "E1", header)

	r := 2
	for rows.Next() {
		var nama string
		var beginning, purchased, sold, ending float64
		if err := rows.Scan(&nama, &beginning, &purchased, &sold, &ending); err != nil {
			return err
		}
		f.SetCellValue(sheet, fmt.Sprintf("A%d", r), nama)
		f.SetCellValue(sheet, fmt.Sprintf("B%d", r), beginning)
		f.SetCellValue(sheet, fmt.Sprintf("C%d", r), purchased)
		f.SetCellValue(sheet, fmt.Sprintf("D%d", r), sold)
		f.SetCellValue(sheet, fmt.Sprintf("E%d", r), ending)
		r++
	}

	if r > 2 {
		f.SetCellStyle(sheet, "B2", fmt.Sprintf("E%d", r-1), number)
	}
	f.SetColWidth(sheet, "A", "A", 30)
	f.SetColWidth(sheet, "B", "E", 16)
	f.SetPanes(sheet, &excelize.Panes{Freeze: true, Split: true, YSplit: 1})

	return sendXLSX(c, f, fmt.Sprintf("inventory_calc_%d_%02d.xlsx", y, m))
}
