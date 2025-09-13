package main

import (
	"log"
	"ynb-backend/config"
	"ynb-backend/models"
	"ynb-backend/routes"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/helmet"
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	config.LoadEnv()
	db := config.ConnectDB()
	defer db.Close()

	// Tambahkan ini agar bisa digunakan di models/*
	models.DB = db //WAJIB agar models.ProcessTransaksiFIFO() bisa pakai DB

	app := fiber.New()

	app.Use(helmet.New())

	// Tambahkan ini agar bisa digunakan dari frontend
	app.Use(cors.New(cors.Config{
		AllowOrigins: "http://localhost:5173", //frontend
		// AllowOrigins:     "*", // izinkan semua
		AllowHeaders:     "Origin, Content-Type, Accept, Authorization, X-Filename",
		ExposeHeaders:    "Content-Disposition", // berguna untuk download CSV
		AllowCredentials: true,
	}))

	routes.SetupRoutes(app)
	app.Listen(":8080")
}
