package models

type PembelianManual struct {
	Tanggal     string  `json:"tanggal"`
	KodeBarang  string  `json:"kode_barang"`
	NamaBarang  string  `json:"nama_barang"`
	Jumlah      int     `json:"jumlah"`
	HargaSatuan float64 `json:"harga_satuan"`
	Total       float64 `json:"total"`
}
