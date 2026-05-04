package deal

import "context"

type Service interface {
	ListDeals(ctx context.Context) ([]Deal, error)
	GetDeal(ctx context.Context, id uint) (*Deal, error)
	CreateDeal(ctx context.Context, deal *Deal) error
	UpdateDeal(ctx context.Context, deal *Deal) error
	DeleteDeal(ctx context.Context, id uint) error
}
