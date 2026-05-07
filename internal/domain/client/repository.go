package client

import "context"

type Repository interface {
	Create(ctx context.Context, client *Client) error
	FindAllByCompany(ctx context.Context, companyID uint) ([]Client, error)
	FindByIDs(ctx context.Context, ids []uint, companyID uint) ([]Client, error)
	FindByWaID(ctx context.Context, waID string, companyID uint) (*Client, error)
	Delete(ctx context.Context, id uint, companyID uint) error
}
