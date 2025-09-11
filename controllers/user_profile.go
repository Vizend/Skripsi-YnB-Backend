package controllers

import (
	"database/sql"
	"log"
	// "time"
	"ynb-backend/config"
	"ynb-backend/models"

	"github.com/gofiber/fiber/v2"
	"golang.org/x/crypto/bcrypt"
)

func UpdateProfile(c *fiber.Ctx) error {
	id := c.Params("id")
	db := config.DB

	var input struct {
		FullName     string `json:"full_name"`
		Username     string `json:"username"`
		Email        string `json:"email"`
		Phone        string `json:"phone"`
		Address      string `json:"address"`
		TanggalLahir string `json:"tanggal_lahir"`
	}

	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Input tidak valid"})
	}

	// Update data profil
	// _, err = db.Exec(`
	// 	UPDATE user SET nama_lengkap = ?, username = ?, email = ?, password = ?
	// 	WHERE user_id = ?`,
	// 	input.FullName, input.Username, input.Email, passwordToSave, id)

	_, err := db.Exec(`
		UPDATE user SET 
			nama_lengkap = ?, username = ?, email = ?, no_telp = ?, alamat = ?, tanggal_lahir = ?
		WHERE user_id = ?`,
		input.FullName, input.Username, input.Email, input.Phone, input.Address, input.TanggalLahir, id,
	)
	if err != nil {
		log.Println("Gagal update profil:", err)
		return c.Status(500).JSON(fiber.Map{"error": "Gagal update data"})
	}

	// Ambil data terbaru untuk dikembalikan
	var user models.User
	var lastLogin sql.NullTime
	var tgl sql.NullTime

	err = db.QueryRow(`
		SELECT u.user_id, u.username, u.email, u.nama_lengkap, u.no_telp, u.alamat,
		       u.tanggal_lahir, u.last_login, r.role_id, r.nama_role
		FROM user u
		LEFT JOIN user_role ur ON u.user_id = ur.user_id
		LEFT JOIN role r ON ur.role_id = r.role_id
		WHERE u.user_id = ?
	`, id).Scan(
		&user.ID, &user.Username, &user.Email, &user.FullName, &user.Phone, &user.Address,
		&tgl, &lastLogin, &user.RoleID, &user.Role,
	)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Gagal mengambil data baru"})
	}

	// Format tanggal lahir
	if tgl.Valid {
		user.TanggalLahir = tgl.Time.Format("2006-01-02")
	} else {
		user.TanggalLahir = ""
	}

	if lastLogin.Valid {
		user.LastLogin = &lastLogin.Time
	} else {
		user.LastLogin = nil
	}

	// Kirim user terbaru ke frontend
	return c.JSON(fiber.Map{
		"message": "Profil berhasil diperbarui",
		"user": fiber.Map{
			"id":        user.ID,
			"username":  user.Username,
			"email":     user.Email,
			"full_name": user.FullName,
			"phone":     user.Phone,
			"address":   user.Address,
			"role":      user.Role,
			"role_id":   user.RoleID,
			"last_login": func() string {
				if user.LastLogin != nil {
					return user.LastLogin.Format("2006-01-02 15:04:05")
				}
				return ""
			}(),
			"tanggal_lahir": user.TanggalLahir,
		},
	})
}

func ChangePassword(c *fiber.Ctx) error {
	id := c.Params("id")
	db := config.DB

	var input struct {
		PasswordLama string `json:"password_lama"`
		PasswordBaru string `json:"password_baru"`
	}

	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Input tidak valid"})
	}

	if input.PasswordBaru == "" || len(input.PasswordBaru) < 6 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Password baru minimal 6 karakter"})
	}

	// Ambil password lama dari database
	var existingPassword string
	err := db.QueryRow("SELECT password FROM user WHERE user_id = ?", id).Scan(&existingPassword)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "User tidak ditemukan"})
	}

	// Validasi password lama
	if bcrypt.CompareHashAndPassword([]byte(existingPassword), []byte(input.PasswordLama)) != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Password lama salah"})
	}

	// Hash password baru
	hashed, err := bcrypt.GenerateFromPassword([]byte(input.PasswordBaru), 12)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Gagal mengenkripsi password baru"})
	}

	// Simpan password baru
	_, err = db.Exec("UPDATE user SET password = ? WHERE user_id = ?", string(hashed), id)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Gagal menyimpan password baru"})
	}

	return c.JSON(fiber.Map{"message": "Password berhasil diperbarui"})
}
