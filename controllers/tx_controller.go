package controllers

import (
	"bufio"
	"bytes"
	"encoding/csv"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"

	"regexp"

	"github.com/gofiber/fiber/v2"
)

type Transaction struct {
	Cashier       string    `json:"cashier"`
	Items         []Item    `json:"items"`
	Subtotal      float64   `json:"subtotal"`
	PaymentType   string    `json:"payment_type"`
	AmountPaid    float64   `json:"amount_paid"`
	Change        float64   `json:"change"`
	TransactionID string    `json:"transaction_id"`
	Date          time.Time `json:"date"`
}

type Item struct {
	Quantity    int     `json:"quantity"`
	Name        string  `json:"name"`
	Price       float64 `json:"price"`
	IsCancelled bool    `json:"is_cancelled"`
}

func ConvertTXTToCSV(c *fiber.Ctx) error {
	file, err := c.FormFile("file")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "No file uploaded",
		})
	}

	src, err := file.Open()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to open file",
		})
	}
	defer src.Close()

	transactions, err := parseTXT(src)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to parse file",
		})
	}

	csvData, err := convertToCSV(transactions)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to convert to CSV",
		})
	}

	c.Set("Content-Type", "text/csv")
	c.Set("Content-Disposition", "attachment; filename=transactions.csv")

	return c.Send(csvData)
}

func GetTransactionData(c *fiber.Ctx) error {
	file, err := c.FormFile("file")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "No file uploaded",
		})
	}

	src, err := file.Open()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to open file",
		})
	}
	defer src.Close()

	transactions, err := parseTXT(src)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to parse file",
		})
	}

	return c.JSON(transactions)
}

func parseTXT(reader io.Reader) ([]Transaction, error) {
	var transactions []Transaction
	var currentTransaction *Transaction
	var isInTransaction bool

	scanner := bufio.NewScanner(reader)
	dateRegex := regexp.MustCompile(`(\d{2}/\d{2}/\d{4} \d{2}:\d{2})`)
	itemRegex := regexp.MustCompile(`^\s*(\d+)\s+([A-Za-z\s\.\'\(\)\d]+[A-Za-z\s\.\'\(\)])\s+(\d+\.\d{3}|\d+\.\d{2}|\d+\.?\d*)$`)
	subtotalRegex := regexp.MustCompile(`SUBTOTAL\s+(\d+\.\d{3}|\d+\.\d{2}|\d+\.?\d*)`)
	paymentRegex := regexp.MustCompile(`(CASH|BCA|QRIS|DEBIT|CREDIT)\s+(\d+\.\d{3}|\d+\.\d{2}|\d+\.?\d*)`)
	changeRegex := regexp.MustCompile(`CHANGE\s+(\d+\.\d{3}|\d+\.\d{2}|\d+\.?\d*)`)
	transactionIDRegex := regexp.MustCompile(`(\d{3} \d{3} \d{6} \d{4})`)
	cancelRegex := regexp.MustCompile(`BATAL`)

	for scanner.Scan() {
		line := scanner.Text()
		line = strings.TrimRight(line, " ")

		// Check for new transaction
		if strings.HasPrefix(line, "KASIR") && strings.Contains(line, "0.000") && !isInTransaction {
			currentTransaction = &Transaction{
				Cashier: strings.TrimSpace(line[:strings.Index(line, "0.000")]),
				Items:   []Item{},
			}
			isInTransaction = true
			// itemCountBeforeBatal = 0
			continue
		}

		// Check for transaction end (date line)
		if dateRegex.MatchString(line) && isInTransaction {
			dateStr := dateRegex.FindString(line)
			date, err := time.Parse("02/01/2006 15:04", dateStr)
			if err == nil {
				currentTransaction.Date = date
			}

			if transactionIDRegex.MatchString(line) {
				currentTransaction.TransactionID = strings.ReplaceAll(transactionIDRegex.FindString(line), " ", "")
			}

			transactions = append(transactions, *currentTransaction)
			isInTransaction = false
			continue
		}

		// Parse items
		if isInTransaction {
			// Check for cancelled line
			if cancelRegex.MatchString(line) {
				// Hapus 2 item terakhir dari transaksi (item sebelum BATAL)
				if len(currentTransaction.Items) >= 2 {
					currentTransaction.Items = currentTransaction.Items[:len(currentTransaction.Items)-2]
				} else if len(currentTransaction.Items) == 1 {
					// Jika hanya ada 1 item, hapus semuanya
					currentTransaction.Items = []Item{}
				}
				continue
			}

			// Parse item lines
			if itemRegex.MatchString(line) {
				matches := itemRegex.FindStringSubmatch(line)
				if len(matches) >= 4 {
					quantity, _ := strconv.Atoi(matches[1])
					name := strings.TrimSpace(matches[2])

					// Parse harga - hilangkan titik sebagai pemisah ribuan
					priceStr := strings.Replace(matches[3], ".", "", -1)
					price, _ := strconv.ParseFloat(priceStr, 64)

					item := Item{
						Quantity:    quantity,
						Name:        name,
						Price:       price,
						IsCancelled: false,
					}

					currentTransaction.Items = append(currentTransaction.Items, item)
				}
				continue
			}

			// Parse subtotal
			if subtotalRegex.MatchString(line) {
				matches := subtotalRegex.FindStringSubmatch(line)
				if len(matches) >= 2 {
					subtotalStr := strings.Replace(matches[1], ".", "", -1)
					subtotal, _ := strconv.ParseFloat(subtotalStr, 64)
					currentTransaction.Subtotal = subtotal
				}
				continue
			}

			// Parse payment
			if paymentRegex.MatchString(line) {
				matches := paymentRegex.FindStringSubmatch(line)
				if len(matches) >= 3 {
					currentTransaction.PaymentType = matches[1]
					amountStr := strings.Replace(matches[2], ".", "", -1)
					amount, _ := strconv.ParseFloat(amountStr, 64)
					currentTransaction.AmountPaid = amount
				}
				continue
			}

			// Parse change
			if changeRegex.MatchString(line) {
				matches := changeRegex.FindStringSubmatch(line)
				if len(matches) >= 2 {
					changeStr := strings.Replace(matches[1], ".", "", -1)
					change, _ := strconv.ParseFloat(changeStr, 64)
					currentTransaction.Change = change
				}
				continue
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return transactions, nil
}

func convertToCSV(transactions []Transaction) ([]byte, error) {
	var csvData bytes.Buffer
	writer := csv.NewWriter(&csvData)

	// Write header
	header := []string{
		"TransactionID", "Date", "Cashier",
		"ItemQuantity", "ItemName", "ItemPrice",
		"Subtotal", "PaymentType", "AmountPaid", "Change",
	}
	if err := writer.Write(header); err != nil {
		return nil, err
	}

	// Write data
	for _, t := range transactions {
		for _, item := range t.Items {
			// Hanya tulis item yang tidak dibatalkan
			if !item.IsCancelled {
				record := []string{
					t.TransactionID,
					t.Date.Format("2006-01-02 15:04:05"),
					t.Cashier,
					strconv.Itoa(item.Quantity),
					item.Name,
					fmt.Sprintf("%.2f", item.Price),
					fmt.Sprintf("%.2f", t.Subtotal),
					t.PaymentType,
					fmt.Sprintf("%.2f", t.AmountPaid),
					fmt.Sprintf("%.2f", t.Change),
				}
				if err := writer.Write(record); err != nil {
					return nil, err
				}
			}
		}

		// If no items, still write the transaction
		if len(t.Items) == 0 {
			record := []string{
				t.TransactionID,
				t.Date.Format("2006-01-02 15:04:05"),
				t.Cashier,
				"",
				"",
				"",
				"",
				fmt.Sprintf("%.2f", t.Subtotal),
				t.PaymentType,
				fmt.Sprintf("%.2f", t.AmountPaid),
				fmt.Sprintf("%.2f", t.Change),
			}
			if err := writer.Write(record); err != nil {
				return nil, err
			}
		}
	}

	writer.Flush()
	if err := writer.Error(); err != nil {
		return nil, err
	}

	return csvData.Bytes(), nil
}
