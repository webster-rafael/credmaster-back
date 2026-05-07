package client

import (
	"time"

	"gorm.io/gorm"
)

type Status string
type PaymentStatus string

const (
	StatusAtivo        Status = "ativo"
	StatusInativo      Status = "inativo"
	StatusInadimplente Status = "inadimplente"
)

const (
	PaymentPago     PaymentStatus = "pago"
	PaymentPendente PaymentStatus = "pendente"
	PaymentAtrasado PaymentStatus = "atrasado"
)

type Client struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	CreatedAt time.Time      `json:"createdAt"`
	UpdatedAt time.Time      `json:"updatedAt"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
	CompanyID uint           `gorm:"not null;index" json:"companyId"`
	Nome      string         `gorm:"not null" json:"nome"`
	CPF       string         `gorm:"not null" json:"cpf"`
	Telefone  string         `json:"telefone"`
	Whatsapp  string         `json:"whatsapp"`
	Email     string         `json:"email"`
	Status    Status         `gorm:"not null;default:'ativo'" json:"status"`
	Pagamento PaymentStatus  `gorm:"not null;default:'pendente'" json:"pagamento"`
	Contrato  string         `json:"contrato"`
	Boletos   int            `gorm:"default:0" json:"boletos"`
	WaID          string `gorm:"index" json:"waId"`
	ContactUserID string `json:"contactUserId"`
}
