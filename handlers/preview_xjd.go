package handlers

import (
	"github.com/gofiber/fiber/v2"
	"ynb-backend/models"
	"ynb-backend/utils"
)

type previewItem struct {
	SrcNama    string   `json:"src_nama"`
	Qty        int      `json:"qty"`
	Price      float64  `json:"price"`
	BarangID   *int     `json:"barang_id,omitempty"`
	MatchName  string   `json:"match_name,omitempty"`
	MatchPrice *int64   `json:"match_price,omitempty"`
	Score      *float64 `json:"score,omitempty"`
	Note       string   `json:"note,omitempty"`
}

type previewTrx struct {
	Tanggal   string        `json:"tanggal"`
	Jam       string        `json:"jam"`
	Metode    string        `json:"metode"`
	Subtotal  float64       `json:"subtotal"`
	Bayar     float64       `json:"bayar"`
	Kembalian float64       `json:"kembalian"`
	RefNo     string        `json:"ref_no"`
	Items     []previewItem `json:"items"`
	Duplicate bool          `json:"duplicate"`
}

func PreviewXJDHandler(c *fiber.Ctx) error {
	fh, err := c.FormFile("file")
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "file tidak ditemukan"})
	}

	path := "./uploads/" + fh.Filename
	if err := c.SaveFile(fh, path); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "gagal simpan file"})
	}

	trxs, err := utils.ParseXJDFile(path)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "gagal parse XJD"})
	}

	// pakai DryRun: buka tx lalu rollback di dalam ProcessTransaksiFIFOWithOpts
	res := make([]previewTrx, 0, len(trxs))

	for _, t := range trxs {
		pv := previewTrx{
			Tanggal: t.Tanggal, Jam: t.Jam, Metode: t.Metode,
			Subtotal: t.Subtotal, Bayar: t.Bayar, Kembalian: t.Kembalian, RefNo: t.RefNo,
		}

		// cek duplikat
		dup, _ := models.CheckDuplicateOnDate(t.Tanggal, t.RefNo) // helper sederhana
		pv.Duplicate = dup

		// coba match tiap item TANPA tulis DB
		for _, it := range t.Items {
			itemPV := previewItem{SrcNama: it.Nama, Qty: it.Jumlah, Price: it.Harga}
			// open a sub-tx for pure reads
			subTx, _ := models.DB.Begin()
			id, cand, err := models.FindBarangIDForPreview(subTx, it.Nama, it.Harga)
			_ = subTx.Rollback()
			if err == nil && cand != nil {
				itemPV.BarangID = &id
				itemPV.MatchName = cand.Nama
				itemPV.MatchPrice = &cand.HargaJual
				itemPV.Score = &cand.Score
			} else {
				itemPV.Note = "Tidak ditemukan barang yang cocok"
			}
			pv.Items = append(pv.Items, itemPV)
		}
		res = append(res, pv)
	}

	return c.JSON(fiber.Map{"transactions": res})
}
