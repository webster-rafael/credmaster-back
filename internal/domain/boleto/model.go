package boleto

import (
	"time"

	"gorm.io/gorm"
)

type Boleto struct {
	ID            uint           `gorm:"primarykey" json:"id"`
	CreatedAt     time.Time      `json:"createdAt"`
	UpdatedAt     time.Time      `json:"updatedAt"`
	DeletedAt     gorm.DeletedAt `gorm:"index" json:"-"`
	CompanyID     uint           `gorm:"not null;index" json:"companyId"`
	ClienteNome   string         `json:"clienteNome"`
	ClienteCPF    string         `json:"clienteCPF"`
	Valor         float64        `json:"valor"`
	Vencimento    string         `json:"vencimento"`
	Arquivo       string         `json:"arquivo"`
	Tamanho       int64          `json:"tamanho"`
	Status        string         `gorm:"default:'pendente'" json:"status"`
	DisparoStatus string         `gorm:"default:'nao_enviado'" json:"disparoStatus"`
	ArquivoURL    string         `json:"arquivoUrl"`
}
