package models

type Cliente struct {
	ID       uint   `gorm:"primaryKey" json:"id"`
	Nombre   string `json:"nombre"`
	Email    string `gorm:"unique" json:"email"`
	Telefono string `json:"telefono"`
	Empresa  string `json:"empresa"`
}
