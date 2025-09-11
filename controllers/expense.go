package controllers

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/gofiber/fiber/v2"
	"ynb-backend/config"
)

type ExpenseInput struct {
	Tanggal    string  `json:"tanggal"`              // "YYYY-MM-DD"
	Kategori   string  `json:"kategori"`             // "gaji" | "listrik" | "transport"
	Metode     string  `json:"metode"`               // "kas" | "bank" | "utang"
	Jumlah     float64 `json:"jumlah"`               // nominal
	Keterangan string  `json:"keterangan,omitempty"` // optional
	UserID     *int    `json:"user_id,omitempty"`    // optional
}

// helper ambil akun_id by kode
func getAkunID(tx *sql.Tx, kode string) (int, error) {
	var id int
	err := tx.QueryRow(`SELECT akun_id FROM akun WHERE kode_akun=?`, kode).Scan(&id)
	if err == sql.ErrNoRows {
		return 0, fmt.Errorf("akun %s tidak ditemukan", kode)
	}
	return id, err
}

// pastikan akun beban transport ada (kalau kamu jalankan migrasi SQL, ini opsional)
func ensureAkunTransport(tx *sql.Tx) error {
	var cnt int
	if err := tx.QueryRow(`SELECT COUNT(*) FROM akun WHERE kode_akun='5-204'`).Scan(&cnt); err != nil {
		return err
	}
	if cnt > 0 {
		return nil
	}
	var parentID int
	if err := tx.QueryRow(`SELECT akun_id FROM akun WHERE kode_akun='5-200'`).Scan(&parentID); err != nil {
		return fmt.Errorf("parent Beban Operasional (5-200) tidak ada")
	}
	_, err := tx.Exec(`
		INSERT INTO akun (kode_akun, nama_akun, jenis, parent_id, is_header)
		VALUES ('5-204','Beban Transportasi','Expense',?,0)`, parentID)
	return err
}

// POST /api/akuntansi/expenses
func CreateExpense(c *fiber.Ctx) error {
	var in ExpenseInput
	if err := c.BodyParser(&in); err != nil {
		return c.Status(400).JSON(fiber.Map{"message": "Payload tidak valid"})
	}
	if in.Tanggal == "" || in.Kategori == "" || in.Metode == "" || in.Jumlah <= 0 {
		return c.Status(400).JSON(fiber.Map{"message": "tanggal, kategori, metode, jumlah wajib diisi & jumlah > 0"})
	}
	// validasi tanggal
	if _, err := time.Parse("2006-01-02", in.Tanggal); err != nil {
		return c.Status(400).JSON(fiber.Map{"message": "Format tanggal harus YYYY-MM-DD"})
	}

	// mapping akun
	debitMap := map[string]string{
		"gaji":      "5-201", // Beban Gaji
		"listrik":   "5-202", // Beban Listrik dan Air
		"transport": "5-204", // Beban Transportasi
	}
	creditMap := map[string]string{
		"kas":   "1-101", // Kas
		"bank":  "1-102", // Bank
		"utang": "2-101", // Utang Usaha (jika dicatat akrual)
	}
	debitKode, ok := debitMap[in.Kategori]
	if !ok {
		return c.Status(400).JSON(fiber.Map{"message": "kategori harus salah satu dari: gaji|listrik|transport"})
	}
	creditKode, ok := creditMap[in.Metode]
	if !ok {
		return c.Status(400).JSON(fiber.Map{"message": "metode harus salah satu dari: kas|bank|utang"})
	}

	tx, err := config.DB.Begin()
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"message": "Gagal mulai transaksi"})
	}
	defer tx.Rollback()

	// jika belum ada akun transport → buat
	if in.Kategori == "transport" {
		if err := ensureAkunTransport(tx); err != nil {
			return c.Status(500).JSON(fiber.Map{"message": "Gagal memastikan akun transport", "detail": err.Error()})
		}
	}

	debitID, err := getAkunID(tx, debitKode)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"message": err.Error()})
	}
	creditID, err := getAkunID(tx, creditKode)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"message": err.Error()})
	}

	// default sementara: user_id = 1
	userID := 1
	if in.UserID != nil && *in.UserID > 0 {
		userID = *in.UserID
	}

	// buat header jurnal
	ref := fmt.Sprintf("EXP-%s", time.Now().Format("060102150405"))
	res, err := tx.Exec(`INSERT INTO jurnal (tanggal, referensi, tipe_jurnal, user_id) VALUES (?, ?, ?, ?)`,
		in.Tanggal, ref, "Beban", userID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"message": "Gagal insert jurnal"})
	}
	jid, _ := res.LastInsertId()

	ket := in.Keterangan
	if ket == "" {
		switch in.Kategori {
		case "gaji":
			ket = "Beban gaji"
		case "listrik":
			ket = "Beban listrik/air"
		default:
			ket = "Beban transportasi"
		}
	}

	// detail debit (beban)
	if _, err := tx.Exec(`
		INSERT INTO jurnal_detail (jurnal_id, akun_id, debit, kredit, keterangan)
		VALUES (?, ?, ?, 0, ?)`,
		jid, debitID, in.Jumlah, ket,
	); err != nil {
		return c.Status(500).JSON(fiber.Map{"message": "Gagal insert jurnal_detail debit"})
	}

	// detail kredit (kas/bank/utang)
	if _, err := tx.Exec(`
		INSERT INTO jurnal_detail (jurnal_id, akun_id, debit, kredit, keterangan)
		VALUES (?, ?, 0, ?, ?)`,
		jid, creditID, in.Jumlah, ket,
	); err != nil {
		return c.Status(500).JSON(fiber.Map{"message": "Gagal insert jurnal_detail kredit"})
	}

	if err := tx.Commit(); err != nil {
		return c.Status(500).JSON(fiber.Map{"message": "Gagal commit"})
	}

	return c.JSON(fiber.Map{
		"message":    "Beban berhasil dicatat",
		"jurnal_id":  jid,
		"referensi":  ref,
		"tanggal":    in.Tanggal,
		"kategori":   in.Kategori,
		"metode":     in.Metode,
		"jumlah":     in.Jumlah,
		"keterangan": ket,
	})
}

// GET /api/akuntansi/expenses?year=2025&month=8
// List ringkas beban per transaksi + agregat per akun
func ListExpenses(c *fiber.Ctx) error {
	year := c.Query("year")
	month := c.Query("month") // 1..12
	where := "a.jenis='Expense'"
	args := []interface{}{}
	if year != "" {
		where += " AND YEAR(j.tanggal)=?"
		args = append(args, year)
	}
	if month != "" {
		where += " AND MONTH(j.tanggal)=?"
		args = append(args, month)
	}

	db := config.DB
	// detail baris (join jurnal + detail + akun)
	rows, err := db.Query(`
		SELECT j.jurnal_id, j.tanggal, j.referensi, a.kode_akun, a.nama_akun,
		       jd.debit, jd.kredit, jd.keterangan
		FROM jurnal j
		JOIN jurnal_detail jd ON jd.jurnal_id=j.jurnal_id
		JOIN akun a ON a.akun_id=jd.akun_id
		WHERE `+where+`
		ORDER BY j.tanggal ASC, j.jurnal_id ASC`, args...)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"message": "DB error", "detail": err.Error()})
	}
	defer rows.Close()

	details := []fiber.Map{}
	for rows.Next() {
		var jid int
		var tgl, ref, kode, nama, ket string
		var debit, kredit float64
		if err := rows.Scan(&jid, &tgl, &ref, &kode, &nama, &debit, &kredit, &ket); err == nil {
			details = append(details, fiber.Map{
				"jurnal_id":  jid,
				"tanggal":    tgl,
				"referensi":  ref,
				"kode_akun":  kode,
				"nama_akun":  nama,
				"debit":      debit,
				"kredit":     kredit,
				"keterangan": ket,
			})
		}
	}

	// agregat per akun (debit - kredit)
	aggRows, err := db.Query(`
		SELECT a.kode_akun, a.nama_akun, SUM(jd.debit - jd.kredit) AS total
		FROM jurnal j
		JOIN jurnal_detail jd ON jd.jurnal_id=j.jurnal_id
		JOIN akun a ON a.akun_id=jd.akun_id
		WHERE ` + where + `
		GROUP BY a.kode_akun, a.nama_akun
		ORDER BY a.kode_akun`)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"message": "DB error", "detail": err.Error()})
	}
	defer aggRows.Close()

	aggs := []fiber.Map{}
	var grand float64
	for aggRows.Next() {
		var kode, nama string
		var total float64
		if err := aggRows.Scan(&kode, &nama, &total); err == nil {
			aggs = append(aggs, fiber.Map{"kode_akun": kode, "nama_akun": nama, "total": total})
			grand += total
		}
	}

	return c.JSON(fiber.Map{
		"summary":      aggs,
		"grand_total":  grand,
		"transactions": details,
	})
}
