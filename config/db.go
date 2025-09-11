package config

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	// "time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/joho/godotenv"
)

var DB *sql.DB

func LoadEnv() {
	godotenv.Load()
}

func ConnectDB() *sql.DB {
	// dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true",
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true&timeout=5s&readTimeout=10s&writeTimeout=10s&multiStatements=true",
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_HOST"),
		os.Getenv("DB_PORT"),
		os.Getenv("DB_NAME"),
	)
	var err error
	DB, err = sql.Open("mysql", dsn)
	if err != nil {
		log.Fatal("Gagal konek DB:", err)
	}
	
	DB.SetMaxOpenConns(20)
	DB.SetMaxIdleConns(10)
	// DB.SetConnMaxLifetime(time.Minute * 5)

	// ✅ Tes koneksi langsung
	if err := DB.Ping(); err != nil {
		log.Fatal("Ping ke database gagal:", err)
	}
	return DB
}
