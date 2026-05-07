package whatsapp

import (
	"time"

	"gorm.io/gorm"
)

type WhatsConfig struct {
	ID                 uint           `gorm:"primarykey" json:"id"`
	CreatedAt          time.Time      `json:"createdAt"`
	UpdatedAt          time.Time      `json:"updatedAt"`
	DeletedAt          gorm.DeletedAt `gorm:"index" json:"-"`
	CompanyID          uint           `gorm:"uniqueIndex;not null" json:"companyId"`
	DisplayPhoneNumber string         `gorm:"not null" json:"displayPhoneNumber"`
	PhoneNumberID      string         `gorm:"not null" json:"phoneNumberId"`
}
