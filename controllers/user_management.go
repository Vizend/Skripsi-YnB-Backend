package controllers

import (
	"database/sql"
	"log"
	"strconv"
	"strings"
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

	user.Username = strings.TrimSpace(user.Username)
	user.Email = strings.TrimSpace(user.Email)
	user.FullName = strings.TrimSpace(user.FullName)
	user.Phone = strings.TrimSpace(user.Phone)
	user.Address = strings.TrimSpace(user.Address)

	// Validasi wajib field
	if user.Username == "" || user.Email == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Username & Email wajib diisi"})
	}

	if user.TanggalLahir == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Tanggal lahir dibutuhkan untuk generate password"})
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

	defer func() {
		_ = tx.Rollback() // aman; kalau sudah Commit
	}()

	// === CEK DUPLIKAT ===
	dupU, dupE, err := checkDupUser(tx, user.Username, user.Email, nil)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "DB error saat cek duplikat"})
	}
	if dupU || dupE {
		fields := fiber.Map{}
		if dupU {
			fields["username"] = "Username sudah digunakan"
		}
		if dupE {
			fields["email"] = "Email sudah digunakan"
		}
		return c.Status(409).JSON(fiber.Map{
			"error":  "Duplicate fields",
			"fields": fields,
		})
	}

	// Insert user
	res, err := tx.Exec(`
		INSERT INTO user (username, password, email, nama_lengkap, no_telp, alamat, tanggal_lahir)
		VALUES (?, ?, ?, ?, ?, ?, ?)`,
		user.Username, string(hashed), user.Email, user.FullName, user.Phone, user.Address, user.TanggalLahir,
	)
	if err != nil {
		// tx.Rollback()
		// return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to save user"})
		return c.Status(409).JSON(fiber.Map{"error": "Username/Email sudah digunakan"})
	}

	userID, err := res.LastInsertId()
	if err != nil {
		// tx.Rollback()
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to retrieve user ID"})
	}

	// Assign role
	_, err = tx.Exec(`INSERT INTO user_role (user_id, role_id) VALUES (?, ?)`, userID, user.RoleID)
	if err != nil {
		// tx.Rollback()
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to assign role"})
	}

	if err := tx.Commit(); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "DB commit error"})
	}

	return c.JSON(fiber.Map{"message": "User created with default password from birthdate"})
}

func UpdateUser(c *fiber.Ctx) error {
	id := c.Params("id")
	var user models.User

	if err := c.BodyParser(&user); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid input"})
	}

	user.Username = strings.TrimSpace(user.Username)
	user.Email = strings.TrimSpace(user.Email)

	db := config.DB
	tx, err := db.Begin()
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to begin transaction"})
	}

	defer func() {
		_ = tx.Rollback()
	}()

	// excludeID untuk cek duplikat
	var exID int64
	if v, err := strconv.ParseInt(id, 10, 64); err == nil {
		exID = v
	}
	dupU, dupE, err := checkDupUser(tx, user.Username, user.Email, &exID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "DB error saat cek duplikat"})
	}
	if dupU || dupE {
		fields := fiber.Map{}
		if dupU {
			fields["username"] = "Username sudah digunakan"
		}
		if dupE {
			fields["email"] = "Email sudah digunakan"
		}
		return c.Status(409).JSON(fiber.Map{
			"error":  "Duplicate fields",
			"fields": fields,
		})
	}

	// Update data user
	_, err = tx.Exec(`
		UPDATE user 
		SET username = ?, email = ?, nama_lengkap = ?, no_telp = ?, alamat = ?, tanggal_lahir = ? 
		WHERE user_id = ?`,
		user.Username, user.Email, user.FullName, user.Phone, user.Address, user.TanggalLahir, id,
	)
	if err != nil {
		// tx.Rollback()
		return c.Status(500).JSON(fiber.Map{"error": "Failed to update user"})
	}

	// Update role (hapus lama → insert baru)
	_, err = tx.Exec(`DELETE FROM user_role WHERE user_id = ?`, id)
	if err != nil {
		// tx.Rollback()
		return c.Status(500).JSON(fiber.Map{"error": "Failed to clear old role"})
	}
	_, err = tx.Exec(`INSERT INTO user_role (user_id, role_id) VALUES (?, ?)`, id, user.RoleID)
	if err != nil {
		// tx.Rollback()
		return c.Status(500).JSON(fiber.Map{"error": "Failed to update role"})
	}

	if err := tx.Commit(); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "DB commit error"})
	}
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

func checkDupUser(tx *sql.Tx, username, email string, excludeID *int64) (dupU, dupE bool, err error) {
	var cnt int

	qU := "SELECT COUNT(*) FROM user WHERE username = ?"
	argsU := []any{username}
	qE := "SELECT COUNT(*) FROM user WHERE email = ?"
	argsE := []any{email}

	if excludeID != nil {
		qU += " AND user_id <> ?"
		argsU = append(argsU, *excludeID)
		qE += " AND user_id <> ?"
		argsE = append(argsE, *excludeID)
	}

	if err = tx.QueryRow(qU, argsU...).Scan(&cnt); err != nil {
		return
	}
	dupU = cnt > 0

	if err = tx.QueryRow(qE, argsE...).Scan(&cnt); err != nil {
		return
	}
	dupE = cnt > 0
	return
}

func CheckAvailability(c *fiber.Ctx) error {
	username := strings.TrimSpace(c.Query("username"))
	email := strings.TrimSpace(c.Query("email"))
	exclude := c.Query("excludeId")

	var exID *int64
	if exclude != "" {
		if v, err := strconv.ParseInt(exclude, 10, 64); err == nil {
			exID = &v
		}
	}

	db := config.DB
	tx, err := db.Begin()
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "DB error"})
	}
	defer func() { _ = tx.Rollback() }()

	dupU, dupE, err := checkDupUser(tx, username, email, exID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "DB error"})
	}

	_ = tx.Commit()
	return c.JSON(fiber.Map{
		"username_taken": dupU,
		"email_taken":    dupE,
	})
}
