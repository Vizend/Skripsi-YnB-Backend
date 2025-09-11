package handlers

import (
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"ynb-backend/models"
)

// ====== 1) KPI Summary (MTD) ======
func GetDashboardSummary(c *fiber.Ctx) error {
	year := c.Query("year")
	month := c.Query("month")
	if year == "" {
		year = strconv.Itoa(time.Now().Year())
	}
	if month == "" {
		month = strconv.Itoa(int(time.Now().Month()))
	}
	y, _ := strconv.Atoi(year)
	m, _ := strconv.Atoi(month)

	// Revenue (MTD)
	var revenue float64
	_ = models.DB.QueryRow(`
		SELECT COALESCE(SUM(jd.kredit - jd.debit),0)
		FROM jurnal_detail jd
		JOIN akun a   ON a.akun_id = jd.akun_id
		JOIN jurnal j ON j.jurnal_id = jd.jurnal_id
		WHERE a.jenis='Revenue' AND YEAR(j.tanggal)=? AND MONTH(j.tanggal)=?`,
		y, m).Scan(&revenue)

	// Expense (MTD)
	var expense float64
	_ = models.DB.QueryRow(`
		SELECT COALESCE(SUM(jd.debit - jd.kredit),0)
		FROM jurnal_detail jd
		JOIN akun a   ON a.akun_id = jd.akun_id
		JOIN jurnal j ON j.jurnal_id = jd.jurnal_id
		WHERE a.jenis='Expense' AND YEAR(j.tanggal)=? AND MONTH(j.tanggal)=?`,
		y, m).Scan(&expense)

	// COGS (MTD) dari stok_riwayat: begin(prev), purchases(cur), ending(cur)
	prevY, prevM := y, m-1
	if m == 1 {
		prevY, prevM = y-1, 12
	}
	var beginInv, purchases, endInv float64
	_ = models.DB.QueryRow(`
		SELECT
			COALESCE((SELECT SUM(stok_akhir) FROM stok_riwayat WHERE tahun=? AND bulan=?),0),
			COALESCE((SELECT SUM(pembelian) FROM stok_riwayat WHERE tahun=? AND bulan=?),0),
			COALESCE((SELECT SUM(stok_akhir) FROM stok_riwayat WHERE tahun=? AND bulan=?),0)
	`, prevY, prevM, y, m, y, m).Scan(&beginInv, &purchases, &endInv)
	cogs := beginInv + purchases - endInv
	if cogs < 0 {
		cogs = 0
	} // jaga-jaga

	// Gross margin %
	var grossMarginPct float64
	if revenue > 0 {
		grossMarginPct = (revenue - cogs) / revenue * 100
	}

	// Net Income (sesuai permintaan: Revenue − Expense)
	netIncome := revenue - expense

	// Kas & Bank (contoh: nama_akun 'Kas' atau 'Bank' => sesuaikan dengan COA kamu)
	var cashBank float64
	_ = models.DB.QueryRow(`
		SELECT COALESCE(SUM(jd.debit - jd.kredit),0)
		FROM jurnal_detail jd
		JOIN akun a   ON a.akun_id = jd.akun_id
		JOIN jurnal j ON j.jurnal_id = jd.jurnal_id
		WHERE a.jenis='Asset'
		  AND a.nama_akun IN ('Kas','Bank')
		  AND YEAR(j.tanggal)=? AND MONTH(j.tanggal)=?`,
		y, m).Scan(&cashBank)

	// Persediaan akhir (nilai) bulan berjalan
	var inventory float64
	_ = models.DB.QueryRow(`
		SELECT COALESCE(SUM(stok_akhir),0)
		FROM stok_riwayat WHERE tahun=? AND bulan=?`, y, m).Scan(&inventory)

	return c.JSON(fiber.Map{
		"revenue":        revenue,
		"expense":        expense,
		"cogs":           cogs,
		"grossMarginPct": grossMarginPct,
		"netIncome":      netIncome, // sesuai permintaan
		"cashBank":       cashBank,
		"inventory":      inventory,
		"beginInv":       beginInv, // berguna untuk kartu COGS
		"purchases":      purchases,
		"endInv":         endInv,
	})
}

// ====== 2) Income Trend (1 tahun) ======
func GetIncomeTrend(c *fiber.Ctx) error {
	year := c.Query("year")
	if year == "" {
		year = strconv.Itoa(time.Now().Year())
	}
	y, _ := strconv.Atoi(year)

	// Revenue & Expense per bulan
	type rec struct {
		M        int
		Rev, Exp float64
	}
	reMap := map[int]*rec{}
	rows, err := models.DB.Query(`
		SELECT MONTH(j.tanggal) AS m,
			SUM(CASE WHEN a.jenis='Revenue' THEN jd.kredit - jd.debit ELSE 0 END) AS revenue,
			SUM(CASE WHEN a.jenis='Expense' THEN jd.debit - jd.kredit ELSE 0 END) AS expense
		FROM jurnal_detail jd
		JOIN akun a   ON a.akun_id = jd.akun_id
		JOIN jurnal j ON j.jurnal_id = jd.jurnal_id
		WHERE YEAR(j.tanggal)=?
		GROUP BY m
	`, y)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var m int
			var rev, exp float64
			_ = rows.Scan(&m, &rev, &exp)
			reMap[m] = &rec{M: m, Rev: rev, Exp: exp}
		}
	}

	// Purchases & Ending dari stok_riwayat per bulan
	type invRec struct{ Purch, End float64 }
	invMap := map[int]*invRec{}
	rows2, err2 := models.DB.Query(`
		SELECT bulan, COALESCE(SUM(pembelian),0), COALESCE(SUM(stok_akhir),0)
		FROM stok_riwayat WHERE tahun=? GROUP BY bulan`, y)
	if err2 == nil {
		defer rows2.Close()
		for rows2.Next() {
			var m int
			var p, e float64
			_ = rows2.Scan(&m, &p, &e)
			invMap[m] = &invRec{Purch: p, End: e}
		}
	}
	// Ending Des tahun sebelumnya untuk Begin Jan (opsional)
	var prevDecEnd float64
	_ = models.DB.QueryRow(`
		SELECT COALESCE(SUM(stok_akhir),0)
		FROM stok_riwayat WHERE tahun=? AND bulan=12`, y-1).Scan(&prevDecEnd)

	// Rakit 12 bulan
	type out struct {
		Month   int     `json:"month"`
		Revenue float64 `json:"revenue"`
		Expense float64 `json:"expense"`
	}

	results := make([]out, 0, 12)
	// saat mengisi results, tidak perlu hitung cogs/gp/ni
	for m := 1; m <= 12; m++ {
		var rev, exp float64
		if r := reMap[m]; r != nil {
			rev, exp = r.Rev, r.Exp
		}
		results = append(results, out{Month: m, Revenue: rev, Expense: exp})
	}
	return c.JSON(results)
}

// ====== 3) Expense Breakdown (bulan ini) ======
func GetExpenseBreakdown(c *fiber.Ctx) error {
	year := c.Query("year")
	month := c.Query("month")
	if year == "" {
		year = strconv.Itoa(time.Now().Year())
	}
	if month == "" {
		month = strconv.Itoa(int(time.Now().Month()))
	}
	y, _ := strconv.Atoi(year)
	m, _ := strconv.Atoi(month)

	rows, err := models.DB.Query(`
		SELECT a.nama_akun,
		    COALESCE(SUM(jd.debit - jd.kredit),0) AS amount
		FROM jurnal_detail jd
		JOIN akun a   ON a.akun_id = jd.akun_id
		JOIN jurnal j ON j.jurnal_id = jd.jurnal_id
		WHERE a.jenis='Expense' AND YEAR(j.tanggal)=? AND MONTH(j.tanggal)=?
		GROUP BY a.akun_id, a.nama_akun
		HAVING amount <> 0
		ORDER BY amount DESC
	`, y, m)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	defer rows.Close()

	type item struct {
		Name   string  `json:"name"`
		Amount float64 `json:"amount"`
	}
	var result []item
	for rows.Next() {
		var name string
		var amt float64
		_ = rows.Scan(&name, &amt)
		result = append(result, item{Name: name, Amount: amt})
	}
	return c.JSON(result)
}

// ====== 4) Top Produk Terjual (qty) bulan ini ======
// NOTE: Pastikan tabel barang_keluar punya kolom tanggal.
// Jika tidak, sesuaikan join/tanggal sesuai skema Anda.
func GetTopProducts(c *fiber.Ctx) error {
	year := c.Query("year")
	month := c.Query("month")
	limit := c.Query("limit", "5")
	if year == "" {
		year = strconv.Itoa(time.Now().Year())
	}
	if month == "" {
		month = strconv.Itoa(int(time.Now().Month()))
	}
	y, _ := strconv.Atoi(year)
	m, _ := strconv.Atoi(month)
	lim, _ := strconv.Atoi(limit)
	if lim <= 0 {
		lim = 5
	}

	rows, err := models.DB.Query(`
		SELECT b.nama_barang, COALESCE(SUM(k.jumlah),0) AS qty
    	FROM barang_keluar k
    	JOIN penjualan p ON p.penjualan_id = k.penjualan_id
    	JOIN barang b     ON b.barang_id = k.barang_id
    	WHERE YEAR(p.tanggal)=? AND MONTH(p.tanggal)=?
    	GROUP BY b.barang_id, b.nama_barang
    	ORDER BY qty DESC
    	LIMIT ?`, y, m, lim)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	defer rows.Close()

	type row struct {
		Name string  `json:"name"`
		Qty  float64 `json:"qty"`
	}
	var result []row
	for rows.Next() {
		var name string
		var qty float64
		_ = rows.Scan(&name, &qty)
		result = append(result, row{Name: name, Qty: qty})
	}
	return c.JSON(result)
}

// ====== 5) Qty Purchases vs Sales per bulan (1 tahun) ======
func GetQtyInOutTrend(c *fiber.Ctx) error {
	year := c.Query("year")
	if year == "" {
		year = strconv.Itoa(time.Now().Year())
	}
	y, _ := strconv.Atoi(year)

	inMap := map[int]float64{}
	outMap := map[int]float64{}

	// Pembelian (barang_masuk)
	rows1, err1 := models.DB.Query(`
		SELECT MONTH(tanggal) AS m, COALESCE(SUM(jumlah),0) AS qty
		FROM barang_masuk WHERE YEAR(tanggal)=? GROUP BY m`, y)
	if err1 == nil {
		defer rows1.Close()
		for rows1.Next() {
			var m int
			var q float64
			_ = rows1.Scan(&m, &q)
			inMap[m] = q
		}
	}
	// Penjualan (barang_keluar)
	rows2, err2 := models.DB.Query(`
		SELECT MONTH(p.tanggal) AS m, COALESCE(SUM(k.jumlah),0) AS qty
    	FROM barang_keluar k
    	JOIN penjualan p ON p.penjualan_id = k.penjualan_id
    	WHERE YEAR(p.tanggal)=?
    	GROUP BY m`, y)
	if err2 == nil {
		defer rows2.Close()
		for rows2.Next() {
			var m int
			var q float64
			_ = rows2.Scan(&m, &q)
			outMap[m] = q
		}
	}

	type out struct {
		Month  int     `json:"month"`
		InQty  float64 `json:"inQty"`
		OutQty float64 `json:"outQty"`
	}
	res := make([]out, 0, 12)
	for m := 1; m <= 12; m++ {
		res = append(res, out{
			Month:  m,
			InQty:  inMap[m],
			OutQty: outMap[m],
		})
	}
	return c.JSON(res)
}
