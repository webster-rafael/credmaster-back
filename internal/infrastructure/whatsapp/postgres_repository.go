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
