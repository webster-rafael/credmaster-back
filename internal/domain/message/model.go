package message

import (
	"time"

	"gorm.io/gorm"
)

type Direction string

const (
	DirectionInbound  Direction = "inbound"
	DirectionOutbound Direction = "outbound"
)

type Status string

const (
	StatusReceived  Status = "received"
	StatusSent      Status = "sent"
	StatusDelivered Status = "delivered"
	StatusRead      Status = "read"
	StatusFailed    Status = "failed"
)

type Message struct {
	ID        string         `gorm:"primarykey" json:"id"`
	CreatedAt time.Time      `json:"createdAt"`
	UpdatedAt time.Time      `json:"updatedAt"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	CompanyID uint  `gorm:"not null;index" json:"companyId"`
	ClientID  *uint `gorm:"index"          json:"clientId"`

	MessagingProduct string `gorm:"not null;default:'whatsapp'" json:"messagingProduct"`
	ContactName      string `gorm:"not null;default:''"         json:"contactName"`
	From             string `gorm:"not null;index"              json:"from"`
	FromUserID       string `gorm:"not null;default:''"         json:"fromUserId"`
	WaTimestamp      int64  `gorm:"not null"                    json:"waTimestamp"`
	Body             string `gorm:"type:text;not null;default:''" json:"body"`
	MessageType      string `gorm:"not null;default:'text'"       json:"messageType"`
	TemplateName     string `gorm:"default:''"                    json:"templateName"`
	Filename         string `gorm:"default:''"                    json:"filename"`
	MimeType         string `gorm:"default:''"                    json:"mimeType"`
	MediaURL         string `gorm:"default:''"                    json:"mediaUrl"`

	Direction Direction `gorm:"not null;default:'inbound'"  json:"direction"`
	Status    Status    `gorm:"not null;default:'received'" json:"status"`

	RawPayload string `gorm:"type:jsonb;not null;default:'{}'" json:"-"`
}
