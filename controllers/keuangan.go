package controllers

import (
	"bufio"
	// "fmt"
	"mime/multipart"
	// "net/http"
	"strconv"
	"strings"
	// "ynb-backend/config"

	// "github.com/gofiber/fiber/v2"
)

type Penjualan struct {
	Tanggal     string
	Jam         string
	Kasir       string
	MetodeBayar string
	Subtotal    float64
	Bayar       float64
	Kembalian   float64
	BarangList  []PenjualanItem
}

type PenjualanItem struct {
	KodeBarang  string
	NamaBarang  string
	Jumlah      int
	HargaSatuan float64
	Total       float64
}

func parseXJD(file multipart.File) ([]Penjualan, error) {
	scanner := bufio.NewScanner(file)
	var transactions []Penjualan
	var current Penjualan

	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "TANGGAL") {
			// Simpan transaksi sebelumnya jika ada
			if current.Tanggal != "" {
				transactions = append(transactions, current)
				current = Penjualan{}
			}
			parts := strings.Split(line, ";")
			if len(parts) >= 2 {
				current.Tanggal = strings.TrimSpace(parts[1])
			}
		} else if strings.HasPrefix(line, "JAM") {
			parts := strings.Split(line, ";")
			if len(parts) >= 2 {
				current.Jam = strings.TrimSpace(parts[1])
			}
		} else if strings.HasPrefix(line, "KASIR") {
			parts := strings.Split(line, ";")
			if len(parts) >= 2 {
				current.Kasir = strings.TrimSpace(parts[1])
			}
		} else if strings.HasPrefix(line, "METODE") {
			parts := strings.Split(line, ";")
			if len(parts) >= 2 {
				current.MetodeBayar = strings.TrimSpace(parts[1])
			}
		} else if strings.HasPrefix(line, "SUBTOTAL") {
			parts := strings.Split(line, ";")
			if len(parts) >= 2 {
				current.Subtotal, _ = strconv.ParseFloat(strings.TrimSpace(parts[1]), 64)
			}
		} else if strings.HasPrefix(line, "BAYAR") {
			parts := strings.Split(line, ";")
			if len(parts) >= 2 {
				current.Bayar, _ = strconv.ParseFloat(strings.TrimSpace(parts[1]), 64)
			}
		} else if strings.HasPrefix(line, "KEMBALIAN") {
			parts := strings.Split(line, ";")
			if len(parts) >= 2 {
				current.Kembalian, _ = strconv.ParseFloat(strings.TrimSpace(parts[1]), 64)
			}
		} else if strings.HasPrefix(line, "BARANG") {
			// BARANG;KODE;NAMA;QTY;HARGA;TOTAL
			parts := strings.Split(line, ";")
			if len(parts) >= 6 {
				jumlah, _ := strconv.Atoi(parts[3])
				harga, _ := strconv.ParseFloat(parts[4], 64)
				total, _ := strconv.ParseFloat(parts[5], 64)
				item := PenjualanItem{
					KodeBarang:  parts[1],
					NamaBarang:  parts[2],
					Jumlah:      jumlah,
					HargaSatuan: harga,
					Total:       total,
				}
				current.BarangList = append(current.BarangList, item)
			}
		}
	}
	// Tambahkan transaksi terakhir
	if current.Tanggal != "" {
		transactions = append(transactions, current)
	}
	return transactions, nil
}
