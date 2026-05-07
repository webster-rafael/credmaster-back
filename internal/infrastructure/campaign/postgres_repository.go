package campaign

import (
	"context"

	domain "github.com/websterdev/cred-master/internal/domain/campaign"
	"gorm.io/gorm"
)

type postgresRepository struct {
	db *gorm.DB
}

func NewPostgresRepository(db *gorm.DB) domain.Repository {
	return &postgresRepository{db: db}
}

func (r *postgresRepository) Create(ctx context.Context, c *domain.Campaign) error {
	return r.db.WithContext(ctx).Create(c).Error
}

func (r *postgresRepository) FindAllByCompany(ctx context.Context, companyID uint) ([]domain.Campaign, error) {
	var campaigns []domain.Campaign
	err := r.db.WithContext(ctx).
		Where("company_id = ?", companyID).
		Order("created_at DESC").
		Find(&campaigns).Error
	return campaigns, err
}

func (r *postgresRepository) FindByID(ctx context.Context, id uint, companyID uint) (*domain.Campaign, error) {
	var c domain.Campaign
	err := r.db.WithContext(ctx).
		Where("id = ? AND company_id = ?", id, companyID).
		First(&c).Error
	return &c, err
}

func (r *postgresRepository) Dispatch(ctx context.Context, id uint, companyID uint, clientes int) error {
	result := r.db.WithContext(ctx).
		Model(&domain.Campaign{}).
		Where("id = ? AND company_id = ?", id, companyID).
		Updates(map[string]any{
			"status":    domain.StatusEmAndamento,
			"clientes":  clientes,
			"enviados":  clientes,
			"entregues": 0,
		})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}
