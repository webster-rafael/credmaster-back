package whatsapp

import (
	"context"

	domain "github.com/websterdev/cred-master/internal/domain/whatsapp"
	"gorm.io/gorm"
)

type postgresRepository struct {
	db *gorm.DB
}

func NewPostgresRepository(db *gorm.DB) domain.Repository {
	return &postgresRepository{db: db}
}

func (r *postgresRepository) FindByCompanyID(ctx context.Context, companyID uint) (*domain.WhatsConfig, error) {
	var cfg domain.WhatsConfig
	err := r.db.WithContext(ctx).Where("company_id = ?", companyID).First(&cfg).Error
	return &cfg, err
}

func (r *postgresRepository) UpdateWabaID(ctx context.Context, configID uint, wabaID string) error {
	return r.db.WithContext(ctx).Model(&domain.WhatsConfig{}).
		Where("id = ?", configID).
		Update("waba_id", wabaID).Error
}
