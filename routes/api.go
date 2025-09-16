package routes

import (
	"time"
	"ynb-backend/controllers"
	"ynb-backend/handlers"

	// "ynb-backend/middlewares"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/limiter"
)

func SetupRoutes(app *fiber.App) {
	api := app.Group("/api")

	// api.Post("/login", controllers.Login)

	// Rate limit khusus /login
	api.Post("/login",
		limiter.New(limiter.Config{
			Max:        30, // max 30 req/menit per IP
			Expiration: 1 * time.Minute,
			// KeyGenerator default = IP
		}),
		controllers.Login,
	)

	api.Post("/signup", controllers.SignUp)
	app.Get("/api/users", controllers.GetAllUsers)
	app.Post("/api/users", controllers.CreateUser)
	app.Put("/api/users/:id", controllers.UpdateUser)
	app.Delete("/api/users/:id", controllers.DeleteUser)
	app.Get("/api/roles", controllers.GetAllRoles)

	app.Put("/api/users/:id/profile", controllers.UpdateProfile)
	app.Put("/api/users/:id/change-password", controllers.ChangePassword)

	//upload data
	app.Post("/api/xjd/upload", handlers.UploadXJDHandler)
	app.Post("/api/xjd/preview", handlers.PreviewXJDHandler)
	app.Post("/api/upload-barang-csv", controllers.UploadBarangCSV)

	//inventory
	app.Post("/api/barang/manual", controllers.AddProductManual)
	app.Get("/api/barang", controllers.GetBarang)
	app.Post("/api/barang/export-csv", controllers.ExportBarangCSV)
	app.Put("/api/barang/:kode_barang", controllers.UpdateBarang)
	app.Delete("/api/barang/:kode_barang", controllers.DeleteBarang)
	app.Post("/api/barang/bulk-delete", controllers.BulkDeleteBarang)
	app.Post("/api/barang/:kode_barang/restore", controllers.RestoreBarang)
	app.Get("/api/barang/:kode_barang/fifo-harga", controllers.GetFIFOHarga)
	app.Get("/api/barang/:kode_barang/masuk", controllers.GetBarangMasukByKode)
	app.Put("/api/barang-masuk/:masuk_id", controllers.UpdateBarangMasuk)
	//autofill
	app.Get("/api/barang/prefill", controllers.PrefillBarang)
	app.Get("/api/barang/search", controllers.SearchBarang)

	//akuntansi
	app.Get("/api/akuntansi/journal-entries", handlers.GetJurnalHandler)
	app.Get("/api/akuntansi/journal-adjustments", handlers.GetJournalAdjustments) //belum dibuat
	app.Get("/api/akuntansi/trial-balance", handlers.GetTrialBalance)
	app.Get("/api/akuntansi/income-statement", handlers.GetIncomeStatement)
	app.Get("/api/akuntansi/balance-sheet", handlers.GetBalanceSheet)
	app.Get("/api/akuntansi/cogs", handlers.GetCOGS) //masih error
	app.Get("/api/akuntansi/inventory-calculation", handlers.GetInventoryCalculation)
	app.Post("/api/akuntansi/expenses", controllers.CreateExpense)
	app.Get("/api/akuntansi/expenses", controllers.ListExpenses)
	// modal & prive
	app.Post("/api/akuntansi/equity", controllers.CreateEquity) // body: {tanggal, tipe: "modal"|"prive", metode: "kas"|"bank", jumlah, keterangan?}
	app.Post("/api/tools/backfill-pembelian", handlers.BackfillPembelian)
	app.Get("/api/akuntansi/years", handlers.GetAvailableYears)
	app.Get("/api/akuntansi/months", handlers.GetAvailableMonths)

	//pembelian
	app.Post("/api/pembelian/manual", handlers.CreatePembelianManual)

	//beban

	//penjualan

	//dashboard
	app.Get("/api/dashboard/summary", handlers.GetDashboardSummary)
	app.Get("/api/dashboard/income-trend", handlers.GetIncomeTrend)
	app.Get("/api/dashboard/expense-breakdown", handlers.GetExpenseBreakdown)
	app.Get("/api/dashboard/top-products", handlers.GetTopProducts)
	app.Get("/api/dashboard/qty-inout-trend", handlers.GetQtyInOutTrend)

	// Transaction routes
	txRoutes := app.Group("/api")
	txRoutes.Post("/convert", controllers.ConvertTXTToCSV)
	txRoutes.Post("/transactions", controllers.GetTransactionData)

	// Tambahkan route untuk menampilkan stok_riwayat per bulan/barang
	app.Get("/api/stok-riwayat", handlers.GetStokRiwayatHandler)
}
