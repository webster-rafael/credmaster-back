package boleto

import (
	"context"

	"gorm.io/gorm"

	domain "github.com/websterdev/cred-master/internal/domain/boleto"
)

type postgresRepository struct {
	db *gorm.DB
}

func NewPostgresRepository(db *gorm.DB) domain.Repository {
	return &postgresRepository{db: db}
}

func (r *postgresRepository) Create(ctx context.Context, b *domain.Boleto) error {
	return r.db.WithContext(ctx).Create(b).Error
}

func (r *postgresRepository) BulkCreate(ctx context.Context, boletos []domain.Boleto) error {
	return r.db.WithContext(ctx).Create(&boletos).Error
}

func (r *postgresRepository) FindAllByCompany(ctx context.Context, companyID uint) ([]domain.Boleto, error) {
	var boletos []domain.Boleto
	err := r.db.WithContext(ctx).Where("company_id = ?", companyID).Order("created_at DESC").Find(&boletos).Error
	return boletos, err
}

func (r *postgresRepository) DeleteByIDs(ctx context.Context, ids []uint, companyID uint) error {
	return r.db.WithContext(ctx).
		Where("id IN ? AND company_id = ?", ids, companyID).
		Delete(&domain.Boleto{}).Error
}

func (r *postgresRepository) FindByCPFs(ctx context.Context, cpfs []string, companyID uint) ([]domain.Boleto, error) {
	var boletos []domain.Boleto
	err := r.db.WithContext(ctx).
		Where("cliente_cpf IN ? AND company_id = ?", cpfs, companyID).
		Order("vencimento ASC").
		Find(&boletos).Error
	return boletos, err
}
