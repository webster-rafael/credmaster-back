package message

import (
	"context"

	domain "github.com/websterdev/cred-master/internal/domain/message"
	"gorm.io/gorm"
)

type postgresRepository struct {
	db *gorm.DB
}

func NewPostgresRepository(db *gorm.DB) domain.Repository {
	return &postgresRepository{db: db}
}

func (r *postgresRepository) Create(ctx context.Context, msg *domain.Message) error {
	return r.db.WithContext(ctx).Create(msg).Error
}

func (r *postgresRepository) FindByID(ctx context.Context, id string) (*domain.Message, error) {
	var msg domain.Message
	err := r.db.WithContext(ctx).First(&msg, id).Error
	return &msg, err
}

func (r *postgresRepository) FindByCompanyID(ctx context.Context, companyID uint, limit, offset int) ([]domain.Message, error) {
	var msgs []domain.Message
	err := r.db.WithContext(ctx).
		Where("company_id = ?", companyID).
		Order("wa_timestamp DESC").
		Limit(limit).
		Offset(offset).
		Find(&msgs).Error
	return msgs, err
}

func (r *postgresRepository) FindByClientID(ctx context.Context, clientID uint) ([]domain.Message, error) {
	var msgs []domain.Message
	err := r.db.WithContext(ctx).
		Where("client_id = ?", clientID).
		Order("wa_timestamp ASC").
		Find(&msgs).Error
	return msgs, err
}

// FindByWaID returns all messages for a contact identified by their from_user_id.
func (r *postgresRepository) FindByWaID(ctx context.Context, waID string, companyID uint) ([]domain.Message, error) {
	var msgs []domain.Message
	err := r.db.WithContext(ctx).
		Where("from_user_id = ? AND company_id = ?", waID, companyID).
		Order("wa_timestamp ASC").
		Find(&msgs).Error
	return msgs, err
}

// GetConversations returns the latest message per unique contact (grouped by from_user_id),
// ordered by most recent activity, with unread count per contact.
func (r *postgresRepository) GetConversations(ctx context.Context, companyID uint) ([]domain.ConversationSummary, error) {
	var result []domain.ConversationSummary
	err := r.db.WithContext(ctx).Raw(`
		SELECT *
		FROM (
			SELECT DISTINCT ON (m.from_user_id)
				m.from_user_id  AS wa_id,
				m.contact_name,
				m.body,
				m.wa_timestamp,
				m.direction,
				m.status,
				m.message_type,
				m."from"        AS sender_phone,
				COALESCE(c.whatsapp, '')  AS client_phone,
				COALESCE(c.wa_id, '')     AS client_wa_id,
				(
					SELECT COUNT(*)
					FROM messages m2
					WHERE m2.from_user_id = m.from_user_id
					  AND m2.company_id   = m.company_id
					  AND m2.direction    = 'inbound'
					  AND m2.status       = 'received'
					  AND m2.deleted_at  IS NULL
				) AS unread_count
			FROM messages m
			LEFT JOIN clients c
				ON c.contact_user_id = m.from_user_id
				AND c.company_id = m.company_id
			WHERE m.company_id = @companyID
			  AND m.deleted_at IS NULL
			ORDER BY m.from_user_id, m.wa_timestamp DESC
		) latest
		ORDER BY wa_timestamp DESC
	`, map[string]interface{}{"companyID": companyID}).Scan(&result).Error
	return result, err
}
