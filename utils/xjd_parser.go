package utils

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strings"
)

type Item struct {
	Nama   string
	Jumlah int
	Harga  float64
}

type Transaksi struct {
	Tanggal  string // format: 2025-03-11
	Jam      string // format: 10:38
	Metode   string // CASH, BCA, QRIS
	Subtotal float64
	Items    []Item
}

func ParseXJDFile(path string) ([]Transaksi, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var hasil []Transaksi
	scanner := bufio.NewScanner(file)

	var trx Transaksi
	itemRegex := regexp.MustCompile(`^\s+(\d+)\s+(.+?)\s+([\d.]+)\s*$`)
	tanggalRegex := regexp.MustCompile(`\d{2}/\d{2}/\d{4}`)
	metodePembayaran := []string{"CASH", "BCA", "QRIS"}
	// inTransaksi := false

	fmt.Println("Mulai parsing XJD...")

	for scanner.Scan() {
		line := scanner.Text()

		if strings.Contains(line, "SUBTOTAL") {
			fmt.Println(">> SUBTOTAL ditemukan:", line)
			parts := strings.Fields(line)
			trx.Subtotal = parseRupiah(parts[len(parts)-1])
		} else if strings.Contains(line, "CHANGE") || containsAny(line, metodePembayaran) {
			fmt.Println(">> Pembayaran:", line)
			for _, metode := range metodePembayaran {
				if strings.Contains(line, metode) {
					trx.Metode = metode
					break
				}
			}
		} else if tanggalRegex.MatchString(line) {
			fmt.Println(">> Baris tanggal:", line)
			parts := strings.Fields(line)
			found := false
			for i := 0; i < len(parts)-1; i++ {
				if tanggalRegex.MatchString(parts[i]) {
					tgl := parts[i]
					jam := parts[i+1]
					trx.Tanggal = convertTanggal(tgl)
					fmt.Println("📦 Baris ditemukan:", line)
					trx.Jam = jam
					hasil = append(hasil, trx)
					trx = Transaksi{}
					found = true
					break
				}
			}
			if found {
				fmt.Println("📅 Parsed:", trx.Tanggal, trx.Jam)
			}

			if !found {
				fmt.Println("⚠️ Gagal parsing tanggal dari:", line)
			}
		} else if itemRegex.MatchString(line) {
			fmt.Println(">> Item:", line)
			m := itemRegex.FindStringSubmatch(line)
			item := Item{
				Jumlah: parseInt(m[1]),
				Nama:   strings.TrimSpace(m[2]),
				Harga:  parseRupiah(m[3]),
			}
			trx.Items = append(trx.Items, item)
		}
	}

	return hasil, nil
}

func parseInt(s string) int {
	var i int
	fmt.Sscanf(s, "%d", &i)
	return i
}

func parseRupiah(s string) float64 {
	var f float64
	fmt.Sscanf(strings.ReplaceAll(s, ".", ""), "%f", &f)
	return f
}

func convertTanggal(s string) string {
	// parts := strings.Split(s, "/")
	// return "20" + parts[2] + "-" + parts[1] + "-" + parts[0]
	parts := strings.Split(s, "/")
	if len(parts) != 3 {
		fmt.Println("❌ Tanggal tidak valid:", s)
		return "" // atau log.Println("Tanggal tidak valid:", s)
	}
	return parts[2] + "-" + parts[1] + "-" + parts[0]
}

func containsAny(s string, arr []string) bool {
	for _, a := range arr {
		if strings.Contains(s, a) {
			return true
		}
	}
	return false
}



