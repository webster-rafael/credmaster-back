package contact

import "context"

type Repository interface {
	FindAll(ctx context.Context) ([]Contact, error)
	FindByID(ctx context.Context, id uint) (*Contact, error)
	Create(ctx context.Context, contact *Contact) error
	Update(ctx context.Context, contact *Contact) error
	Delete(ctx context.Context, id uint) error
}
