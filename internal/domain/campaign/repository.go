package campaign

import "context"

type Repository interface {
	Create(ctx context.Context, c *Campaign) error
	FindAllByCompany(ctx context.Context, companyID uint) ([]Campaign, error)
	FindByID(ctx context.Context, id uint, companyID uint) (*Campaign, error)
	Dispatch(ctx context.Context, id uint, companyID uint, clientes int) error
}
