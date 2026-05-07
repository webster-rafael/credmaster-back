package campaign

import (
	"time"

	"gorm.io/gorm"
)

type Tipo   string
type Status string

const (
	TipoTexto     Tipo = "texto"
	TipoImagem    Tipo = "imagem"
	TipoVideo     Tipo = "video"
	TipoDocumento Tipo = "documento"
)

const (
	StatusRascunho    Status = "rascunho"
	StatusEmAndamento Status = "em_andamento"
	StatusEnviada     Status = "enviada"
	StatusFalha       Status = "falha"
)

type Campaign struct {
	ID           uint           `gorm:"primarykey"        json:"id"`
	CreatedAt    time.Time      `                         json:"criadoEm"`
	UpdatedAt    time.Time      `                         json:"updatedAt"`
	DeletedAt    gorm.DeletedAt `gorm:"index"             json:"-"`
	CompanyID    uint           `gorm:"not null;index"    json:"-"`
	Name         string         `gorm:"not null"          json:"name"`
	Tipo         Tipo           `gorm:"not null;default:'texto'"    json:"tipo"`
	Status       Status         `gorm:"not null;default:'rascunho'" json:"status"`
	TemplateName string         `json:"templateName"`
	Descricao    string         `json:"descricao"`
	Clientes     int            `gorm:"default:0" json:"clientes"`
	Enviados     int            `gorm:"default:0" json:"enviados"`
	Entregues    int            `gorm:"default:0" json:"entregues"`
}
