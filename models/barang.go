package models

type Barang struct {
	ID        int     `json:"id"`
	Kode      string  `json:"kode_barang"`
	Nama      string  `json:"nama_barang"`
	HargaJual float64 `json:"harga_jual"`
	HargaBeli float64 `json:"harga_beli"`
	Stok      int     `json:"jumlah_stock"`
}

// BarangCSVInput digunakan untuk parsing upload dari frontend
type BarangCSVInput struct {
	KodeBarang  string  `json:"kode_barang"`
	NamaBarang  string  `json:"nama_barang"`
	HargaJual   float64 `json:"harga_jual"`
	HargaBeli   float64 `json:"harga_beli"`
	JumlahStock int     `json:"jumlah_stock"`
}
