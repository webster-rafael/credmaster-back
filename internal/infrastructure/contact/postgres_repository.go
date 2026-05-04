package contact

import (
	"context"

	domain "github.com/websterdev/cred-master/internal/domain/contact"
	"gorm.io/gorm"
)

type postgresRepository struct {
	db *gorm.DB
}

func NewPostgresRepository(db *gorm.DB) domain.Repository {
	return &postgresRepository{db: db}
}

func (r *postgresRepository) FindAll(ctx context.Context) ([]domain.Contact, error) {
	var contacts []domain.Contact
	result := r.db.WithContext(ctx).Find(&contacts)
	return contacts, result.Error
}

func (r *postgresRepository) FindByID(ctx context.Context, id uint) (*domain.Contact, error) {
	var contact domain.Contact
	result := r.db.WithContext(ctx).First(&contact, id)
	return &contact, result.Error
}

func (r *postgresRepository) Create(ctx context.Context, contact *domain.Contact) error {
	return r.db.WithContext(ctx).Create(contact).Error
}

func (r *postgresRepository) Update(ctx context.Context, contact *domain.Contact) error {
	return r.db.WithContext(ctx).Save(contact).Error
}

func (r *postgresRepository) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&domain.Contact{}, id).Error
}
