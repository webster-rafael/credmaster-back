package message

import "context"

// ConversationSummary is the shape returned by GetConversations.
// Each row represents the latest message per unique contact (grouped by "from").
// The WaID field is populated from the "from" column (same value — E.164 phone).
type ConversationSummary struct {
	WaID        string `json:"waId"        gorm:"column:wa_id"`
	ContactName string `json:"contactName" gorm:"column:contact_name"`
	Body        string `json:"body"        gorm:"column:body"`
	WaTimestamp int64  `json:"waTimestamp" gorm:"column:wa_timestamp"`
	Direction   string `json:"direction"   gorm:"column:direction"`
	Status      string `json:"status"      gorm:"column:status"`
	MessageType string `json:"messageType" gorm:"column:message_type"`
	SenderPhone   string `json:"senderPhone"   gorm:"column:sender_phone"`
	ClientPhone   string `json:"clientPhone"   gorm:"column:client_phone"`
	ClientWaID    string `json:"clientWaId"    gorm:"column:client_wa_id"`
	UnreadCount   int64  `json:"unreadCount"   gorm:"column:unread_count"`
}

type Repository interface {
	Create(ctx context.Context, msg *Message) error
	FindByID(ctx context.Context, id string) (*Message, error)
	FindByCompanyID(ctx context.Context, companyID uint, limit, offset int) ([]Message, error)
	FindByClientID(ctx context.Context, clientID uint) ([]Message, error)
	FindByWaID(ctx context.Context, waID string, companyID uint) ([]Message, error)
	GetConversations(ctx context.Context, companyID uint) ([]ConversationSummary, error)
}
