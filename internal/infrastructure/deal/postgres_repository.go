package deal

import (
	"context"

	domain "github.com/websterdev/cred-master/internal/domain/deal"
	"gorm.io/gorm"
)

type postgresRepository struct {
	db *gorm.DB
}

func NewPostgresRepository(db *gorm.DB) domain.Repository {
	return &postgresRepository{db: db}
}

func (r *postgresRepository) FindAll(ctx context.Context) ([]domain.Deal, error) {
	var deals []domain.Deal
	result := r.db.WithContext(ctx).Find(&deals)
	return deals, result.Error
}

func (r *postgresRepository) FindByID(ctx context.Context, id uint) (*domain.Deal, error) {
	var deal domain.Deal
	result := r.db.WithContext(ctx).First(&deal, id)
	return &deal, result.Error
}

func (r *postgresRepository) Create(ctx context.Context, deal *domain.Deal) error {
	return r.db.WithContext(ctx).Create(deal).Error
}

func (r *postgresRepository) Update(ctx context.Context, deal *domain.Deal) error {
	return r.db.WithContext(ctx).Save(deal).Error
}

func (r *postgresRepository) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&domain.Deal{}, id).Error
}
