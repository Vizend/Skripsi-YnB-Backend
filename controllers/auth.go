package controllers

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"math/rand"
	"os"
	"time"
	"ynb-backend/config"
	"ynb-backend/models"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

var (
	// window & lockout policy
	loginWindow       = 15 * time.Minute //jendela hitung attempt
	maxAttemptsWindow = 5                //5x gagal dalam 15 menit => lock
	lockDurationShort = 15 * time.Minute //durasi lock awal
	dummyHash         []byte             //untuk samakan timing user-not-found
)

func init() {
	rand.Seed(time.Now().UnixNano())
	var err error
	//dummy hash untuk menyamakan timing (anti user enumeration timing)
	dummyHash, err = bcrypt.GenerateFromPassword([]byte("dummy-password-to-match-timing"), 12)
	if err != nil {
		log.Fatal("failed to init dummy hash:", err)
	}
}

// ========== helpers untuk login attempts ==========

type attemptRow struct {
	FailedCount int
	WindowStart time.Time
	LockedUntil sql.NullTime
}

func getAttempt(db *sql.DB, key string) (*attemptRow, error) {
	row := db.QueryRow("SELECT failed_count, window_start, locked_until FROM login_attempts WHERE key_name=?", key)
	var r attemptRow
	err := row.Scan(&r.FailedCount, &r.WindowStart, &r.LockedUntil)
	if errors.Is(err, sql.ErrNoRows) {
		//buat baris awal
		_, _ = db.Exec("INSERT INTO login_attempts (key_name, failed_count, window_start) VALUES (?, 0, NOW())", key)
		//ambil lagi
		row2 := db.QueryRow("SELECT failed_count, window_start, locked_until FROM login_attempts WHERE key_name=?", key)
		err2 := row2.Scan(&r.FailedCount, &r.WindowStart, &r.LockedUntil)
		return &r, err2
	}
	return &r, err
}

func isLocked(r *attemptRow) (bool, time.Duration) {
	if r.LockedUntil.Valid {
		until := r.LockedUntil.Time
		if time.Now().Before(until) {
			return true, time.Until(until)
		}
	}
	return false, 0
}

func bumpFailure(db *sql.DB, key string) (*attemptRow, error) {
	r, err := getAttempt(db, key)
	if err != nil {
		return nil, err
	}
	now := time.Now()

	// reset window kalau sudah lewat
	if now.Sub(r.WindowStart) > loginWindow {
		r.FailedCount = 0
		r.WindowStart = now
	}

	r.FailedCount++
	lockedUntil := sql.NullTime{}

	// lock kalau melewati ambang
	if r.FailedCount >= maxAttemptsWindow {
		lockedUntil = sql.NullTime{Time: now.Add(lockDurationShort), Valid: true}
	}

	_, err = db.Exec(`
		UPDATE login_attempts
		SET failed_count=?, window_start=?, last_failed_at=NOW(), locked_until=IF(?=1, ?, NULLIF(locked_until, locked_until))
		WHERE key_name=?`,
		r.FailedCount, r.WindowStart, // count & window
		boolToInt(lockedUntil.Valid), lockedUntil.Time, // lock
		key,
	)
	if err != nil {
		return nil, err
	}

	// refresh state yang kita pegang
	r.LockedUntil = lockedUntil
	return r, nil
}

func resetKey(db *sql.DB, key string) {
	_, _ = db.Exec("DELETE FROM login_attempts WHERE key_name=?", key)
}

func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}

// ========== SIGNUP tetap, tapi hilangkan debug/password log jika ada ==========

func SignUp(c *fiber.Ctx) error {
	var u models.User
	if err := c.BodyParser(&u); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid input"})
	}

	hash, _ := bcrypt.GenerateFromPassword([]byte(u.Password), 14)
	_, err := config.DB.Exec("INSERT INTO user (username, password, email) VALUES (?, ?, ?)", u.Username, hash, u.Email)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to register"})
	}

	return c.JSON(fiber.Map{"message": "Signup berhasil"})
}

// ========== LOGIN dengan hardening anti-bruteforce ==========
func Login(c *fiber.Ctx) error {
	var input struct {
		Username string `json:"username"`
		Password string `json:"password"`
		// CaptchaToken string `json:"captcha_token,omitempty"` // kalau nanti mau pakai hCaptcha/recaptcha
	}
	if err := c.BodyParser(&input); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request"})
	}

	db := config.DB
	ipKey := "ip:" + c.IP()
	userKey := "user:" + input.Username

	// 🔒 Cek lock untuk IP & Username (kalau salah satu lock → tolak)
	if r, err := getAttempt(db, ipKey); err == nil {
		if locked, wait := isLocked(r); locked {
			c.Set("Retry-After", fmt.Sprintf("%d", int(wait.Seconds())))
			return c.Status(429).JSON(fiber.Map{"error": "Terlalu banyak percobaan. Coba lagi nanti."})
		}
	}
	if r, err := getAttempt(db, userKey); err == nil {
		if locked, wait := isLocked(r); locked {
			c.Set("Retry-After", fmt.Sprintf("%d", int(wait.Seconds())))
			return c.Status(429).JSON(fiber.Map{"error": "Terlalu banyak percobaan. Coba lagi nanti."})
		}
	}

	// 🔕 Jangan pernah log password ke console/log file
	var u models.User
	var tanggalLahir sql.NullTime

	err := config.DB.QueryRow(`
    SELECT u.user_id, u.username, u.password, u.email, u.nama_lengkap, u.no_telp, u.alamat, u.tanggal_lahir,
	        r.role_id, r.nama_role
    FROM user u
    LEFT JOIN user_role ur ON u.user_id = ur.user_id
    LEFT JOIN role r ON ur.role_id = r.role_id
    WHERE u.username = ?`, input.Username).
		Scan(&u.ID, &u.Username, &u.Password, &u.Email, &u.FullName, &u.Phone, &u.Address, &tanggalLahir,
			&u.RoleID, &u.Role)

	// Jitter untuk samakan timing (anti timing attack)
	time.Sleep(200*time.Millisecond + time.Duration(rand.Intn(600))*time.Millisecond)

	invalid := func() error {
		// samakan timing bila user tidak ada
		_ = bcrypt.CompareHashAndPassword(dummyHash, []byte(input.Password))
		// bump counter untuk IP & username
		if _, err := bumpFailure(db, ipKey); err != nil {
			log.Println("bump ip fail:", err)
		}
		if _, err := bumpFailure(db, userKey); err != nil {
			log.Println("bump user fail:", err)
		}
		// pesan generik
		return c.Status(401).JSON(fiber.Map{"error": "Invalid username or password"})
	}

	if errors.Is(err, sql.ErrNoRows) {
		return invalid()
	}
	if err != nil {
		log.Println("DB error:", err)
		return c.Status(500).JSON(fiber.Map{"error": "Server error"})
	}

	if tanggalLahir.Valid {
		u.TanggalLahir = tanggalLahir.Time.Format("2006-01-02")
	} else {
		u.TanggalLahir = ""
	}

	// Cek password
	if bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(input.Password)) != nil {
		return invalid()
	}

	// ✅ Berhasil → reset counter IP & username
	resetKey(db, ipKey)
	resetKey(db, userKey)

	// Update last_login (best-effort)
	now := time.Now()
	if _, err = db.Exec("UPDATE user SET last_login = ? WHERE user_id = ?", now, u.ID); err != nil {
		log.Println("Failed to update last login:", err)
	}
	u.LastLogin = &now

	// if err != nil {
	// 	log.Println("Failed to update last login:", err)
	// }

	// Generate JWT
	claims := jwt.MapClaims{
		"user_id": u.ID,
		"role":    u.Role,
		"exp":     time.Now().Add(72 * time.Hour).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString([]byte(os.Getenv("JWT_SECRET")))
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to generate token"})
	}

	// ⚠️ Saran kuat: kirim JWT sebagai HttpOnly cookie, bukan disimpan di localStorage (anti XSS)
	// Uncomment ini kalau siap migrasi di frontend:
	/*
		c.Cookie(&fiber.Cookie{
			Name:     "access_token",
			Value:    signed,
			HTTPOnly: true,
			Secure:   true,              // aktifkan di production (HTTPS)
			SameSite: "Lax",
			Expires:  time.Now().Add(72 * time.Hour),
		})
		return c.JSON(fiber.Map{
			"message": "Login berhasil",
			"user":    ... // tanpa token di body
		})
	*/

	// return c.JSON(fiber.Map{"token": signed})
	return c.JSON(fiber.Map{
		"message": "Login berhasil",
		"token":   signed,
		"user": fiber.Map{
			"id":        u.ID,
			"username":  u.Username,
			"email":     u.Email,
			"full_name": u.FullName,
			"phone":     u.Phone,
			"address":   u.Address,
			"role":      u.Role,
			"role_id":   u.RoleID,
			"last_login": func() string {
				if u.LastLogin != nil {
					return u.LastLogin.Format("2006-01-02 15:04:05")
				}
				return ""
			}(),
			"tanggal_lahir": u.TanggalLahir,
		},
	})
}
