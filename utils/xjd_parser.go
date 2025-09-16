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
	Tanggal   string
	Jam       string
	Metode    string
	Subtotal  float64
	Bayar     float64
	Kembalian float64
	RefNo     string // nomor struk (reset harian)
	Items     []Item
}

var (
	itemRegex        = regexp.MustCompile(`^\s+(\d+)\s+(.+?)\s+([\d.]+)\s*$`)
	tanggalRegex     = regexp.MustCompile(`\d{2}/\d{2}/\d{4}`)
	rupiahRegex      = regexp.MustCompile(`([\d.]+)`) // ambil angka terakhir di baris
	metodePembayaran = []string{"CASH", "BCA", "QRIS"}
)

func ParseXJDFile(path string) ([]Transaksi, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var hasil []Transaksi
	scanner := bufio.NewScanner(f)
	var trx Transaksi

	for scanner.Scan() {
		line := scanner.Text()

		switch {
		case strings.Contains(line, "SUBTOTAL"):
			// contoh: "SUBTOTAL                53.500"
			if m := rupiahRegex.FindAllString(line, -1); len(m) > 0 {
				trx.Subtotal = parseRupiah(m[len(m)-1])
			}

		case strings.Contains(line, "CHANGE"):
			// contoh: "CHANGE                  1.500"
			if m := rupiahRegex.FindAllString(line, -1); len(m) > 0 {
				trx.Kembalian = parseRupiah(m[len(m)-1])
			}

		case containsAny(line, metodePembayaran):
			// contoh: "CASH                    55.000"
			for _, metode := range metodePembayaran {
				if strings.Contains(line, metode) {
					trx.Metode = metode
					// jika ada angka di baris ini, jadikan Bayar
					if m := rupiahRegex.FindAllString(line, -1); len(m) > 0 {
						trx.Bayar = parseRupiah(m[len(m)-1])
					}
					break
				}
			}

		case tanggalRegex.MatchString(line):
			// contoh: "001 001 000019 0001 10/08/2025 10:38"
			parts := strings.Fields(line)
			// ambil tanggal + jam
			iTgl := -1
			for i := 0; i < len(parts); i++ {
				if tanggalRegex.MatchString(parts[i]) {
					iTgl = i
					break
				}
			}
			if iTgl >= 0 && iTgl+1 < len(parts) {
				tgl := parts[iTgl]
				jam := parts[iTgl+1]
				trx.Tanggal = convertTanggal(tgl)
				trx.Jam = jam

				// cari token numerik panjang yg tampak seperti nomor struk di sisi kiri tanggal ( refno terpanjang )
				ref := "" 
				for j := iTgl - 1; j >= 0; j-- {
					tok := parts[j]
					if !isAllDigits(tok) {
						continue
					}
					if len(tok) >= 6 {
						ref = tok
						break
					}
					if len(tok) > len(ref) {
						ref = tok
					}
				}
				trx.RefNo = ref

				// tutup transaksi saat ini & reset
				hasil = append(hasil, trx)
				trx = Transaksi{}
			}

		case itemRegex.MatchString(line):
			m := itemRegex.FindStringSubmatch(line)
			trx.Items = append(trx.Items, Item{
				Jumlah: parseInt(m[1]),
				Nama:   strings.TrimSpace(m[2]),
				Harga:  parseRupiah(m[3]),
			})
		}
	}
	return hasil, nil
}

func isAllDigits(s string) bool {
	for _, r := range s {
		if r < '0' || r > '9' {
			return false
		}
	}
	return len(s) > 0
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
	parts := strings.Split(s, "/")
	if len(parts) != 3 {
		fmt.Println("❌ Tanggal tidak valid:", s)
		return ""
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
