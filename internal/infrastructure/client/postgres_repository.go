package client

import (
	"context"
	"fmt"
	"time"

	domain "github.com/websterdev/cred-master/internal/domain/client"
	"gorm.io/gorm"
)

type postgresRepository struct {
	db *gorm.DB
}

func NewPostgresRepository(db *gorm.DB) domain.Repository {
	return &postgresRepository{db: db}
}

func (r *postgresRepository) Create(ctx context.Context, client *domain.Client) error {
	if err := r.db.WithContext(ctx).Create(client).Error; err != nil {
		return err
	}
	client.Contrato = fmt.Sprintf("CTR-%d-%04d", time.Now().Year(), client.ID)
	return r.db.WithContext(ctx).Model(client).Update("contrato", client.Contrato).Error
}

func (r *postgresRepository) FindAllByCompany(ctx context.Context, companyID uint) ([]domain.Client, error) {
	var clients []domain.Client
	err := r.db.WithContext(ctx).
		Where("company_id = ?", companyID).
		Order("created_at DESC").
		Find(&clients).Error
	return clients, err
}

func (r *postgresRepository) FindByWaID(ctx context.Context, waID string, companyID uint) (*domain.Client, error) {
	var client domain.Client
	err := r.db.WithContext(ctx).
		Where("(wa_id = ? OR contact_user_id = ?) AND company_id = ?", waID, waID, companyID).
		First(&client).Error
	return &client, err
}

func (r *postgresRepository) Delete(ctx context.Context, id uint, companyID uint) error {
	result := r.db.WithContext(ctx).
		Where("id = ? AND company_id = ?", id, companyID).
		Delete(&domain.Client{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}
