package controllers

import (
	"database/sql"
	"log"
	"time"
	"ynb-backend/config"
	"ynb-backend/models"

	"github.com/gofiber/fiber/v2"
	"golang.org/x/crypto/bcrypt"
)

func CreateUser(c *fiber.Ctx) error {
	var user models.User
	if err := c.BodyParser(&user); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid input"})
	}

	// Password default dari tanggal lahir: DDMMYYYY
	password := ""
	if user.TanggalLahir != "" {
		t, err := time.Parse("2006-01-02", user.TanggalLahir)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid date format"})
		}
		password = t.Format("02012006")
	} else {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Tanggal lahir dibutuhkan untuk generate password"})
	}

	// Hash password
	hashed, err := bcrypt.GenerateFromPassword([]byte(password), 12)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Could not hash password"})
	}

	db := config.DB
	tx, err := db.Begin()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "DB error"})
	}

	// Insert user
	res, err := tx.Exec(`
		INSERT INTO user (username, password, email, nama_lengkap, no_telp, alamat, tanggal_lahir)
		VALUES (?, ?, ?, ?, ?, ?, ?)`,
		user.Username, string(hashed), user.Email, user.FullName, user.Phone, user.Address, user.TanggalLahir,
	)
	if err != nil {
		tx.Rollback()
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to save user"})
	}

	userID, err := res.LastInsertId()
	if err != nil {
		tx.Rollback()
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to retrieve user ID"})
	}

	// Assign role
	_, err = tx.Exec(`INSERT INTO user_role (user_id, role_id) VALUES (?, ?)`, userID, user.RoleID)
	if err != nil {
		tx.Rollback()
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to assign role"})
	}

	tx.Commit()

	return c.JSON(fiber.Map{"message": "User created with default password from birthdate"})
}

func UpdateUser(c *fiber.Ctx) error {
	id := c.Params("id")
	var user models.User

	if err := c.BodyParser(&user); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid input"})
	}

	db := config.DB
	tx, err := db.Begin()
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to begin transaction"})
	}

	// Update data user
	_, err = tx.Exec(`
		UPDATE user 
		SET username = ?, email = ?, nama_lengkap = ?, no_telp = ?, alamat = ?, tanggal_lahir = ? 
		WHERE user_id = ?`,
		user.Username, user.Email, user.FullName, user.Phone, user.Address, user.TanggalLahir, id,
	)
	if err != nil {
		tx.Rollback()
		return c.Status(500).JSON(fiber.Map{"error": "Failed to update user"})
	}

	// Update role (hapus lama → insert baru)
	_, err = tx.Exec(`DELETE FROM user_role WHERE user_id = ?`, id)
	if err != nil {
		tx.Rollback()
		return c.Status(500).JSON(fiber.Map{"error": "Failed to clear old role"})
	}
	_, err = tx.Exec(`INSERT INTO user_role (user_id, role_id) VALUES (?, ?)`, id, user.RoleID)
	if err != nil {
		tx.Rollback()
		return c.Status(500).JSON(fiber.Map{"error": "Failed to update role"})
	}

	tx.Commit()
	return c.JSON(fiber.Map{"message": "User updated successfully"})
}

func DeleteUser(c *fiber.Ctx) error {
	id := c.Params("id")
	db := config.DB

	// Delete user role
	_, err := db.Exec(`DELETE FROM user_role WHERE user_id = ?`, id)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Gagal hapus role user"})
	}

	// Delete usernya
	_, err = db.Exec(`DELETE FROM user WHERE user_id = ?`, id)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Gagal hapus user"})
	}

	return c.JSON(fiber.Map{"message": "User berhasil dihapus"})
}

func GetAllUsers(c *fiber.Ctx) error {
	db := config.DB
	rows, err := db.Query(`
SELECT u.user_id, u.username, u.email, u.nama_lengkap, u.no_telp, u.alamat,
       u.tanggal_lahir, u.last_login, r.role_id, r.nama_role
FROM user u
		LEFT JOIN user_role ur ON u.user_id = ur.user_id
		LEFT JOIN role r ON ur.role_id = r.role_id
`)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Database error"})
	}
	defer rows.Close()

	var users []models.User
	for rows.Next() {
		var user models.User
		var lastLogin sql.NullTime
		// err := rows.Scan(&user.ID, &user.Username, &user.Email, &user.FullName, &user.LastLogin, &user.Role)
		err := rows.Scan(
			&user.ID, &user.Username, &user.Email, &user.FullName, &user.Phone, &user.Address,
			&user.TanggalLahir, &lastLogin, &user.RoleID, &user.Role,
		)
		if err != nil {
			log.Println("Error scan:", err)
			continue
		}
		if lastLogin.Valid {
			user.LastLogin = &lastLogin.Time
		} else {
			user.LastLogin = nil
		}
		// user.LastLogin = lastLogin
		users = append(users, user)
	}

	return c.JSON(users)
}

func GetAllRoles(c *fiber.Ctx) error {
	db := config.DB
	rows, err := db.Query("SELECT role_id, nama_role FROM role")
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Gagal ambil data role"})
	}
	defer rows.Close()

	var roles []models.Role
	for rows.Next() {
		var role models.Role
		if err := rows.Scan(&role.ID, &role.Name); err != nil {
			continue
		}
		roles = append(roles, role)
	}

	return c.JSON(roles)
}
