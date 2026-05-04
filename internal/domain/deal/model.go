package deal

import (
	"time"

	"gorm.io/gorm"
)

type Stage string

const (
	StageProspect    Stage = "prospect"
	StageQualified   Stage = "qualified"
	StageProposal    Stage = "proposal"
	StageNegotiation Stage = "negotiation"
	StageClosed      Stage = "closed"
	StageLost        Stage = "lost"
)

type Deal struct {
	ID          uint           `gorm:"primarykey" json:"id"`
	CreatedAt   time.Time      `json:"createdAt"`
	UpdatedAt   time.Time      `json:"updatedAt"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
	Title       string         `gorm:"not null" json:"title"`
	Value       float64        `gorm:"not null" json:"value"`
	Stage       Stage          `gorm:"not null;default:'prospect'" json:"stage"`
	ContactID   uint           `json:"contactId"`
	ClosingDate *time.Time     `json:"closingDate"`
}
