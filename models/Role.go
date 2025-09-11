package models

type Role struct {
	ID   int    `json:"role_id"`
	Name string `json:"nama_role"`
	Desc string `json:"deskripsi"`
}
