package deal

import "context"

type Repository interface {
	FindAll(ctx context.Context) ([]Deal, error)
	FindByID(ctx context.Context, id uint) (*Deal, error)
	Create(ctx context.Context, deal *Deal) error
	Update(ctx context.Context, deal *Deal) error
	Delete(ctx context.Context, id uint) error
}
