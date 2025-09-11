package models

import (
	// "database/sql"
	"time"
)

type User struct {
	ID           int       `json:"id"`
	Username     string    `json:"username"`
	Password     string    `json:"-"`
	Email        string    `json:"email"`
	FullName     string    `json:"full_name"`
	Phone        string    `json:"phone"`
	Address      string    `json:"address"`
	Role         string    `json:"role"`
	RoleID       int       `json:"role_id"`
	LastLogin    *time.Time `json:"last_login"`
	TanggalLahir string    `json:"tanggal_lahir"` // format: "YYYY-MM-DD"
}
