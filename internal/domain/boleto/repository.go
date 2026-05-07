package boleto

import "context"

type Repository interface {
	Create(ctx context.Context, boleto *Boleto) error
	BulkCreate(ctx context.Context, boletos []Boleto) error
	FindAllByCompany(ctx context.Context, companyID uint) ([]Boleto, error)
	FindByCPFs(ctx context.Context, cpfs []string, companyID uint) ([]Boleto, error)
	DeleteByIDs(ctx context.Context, ids []uint, companyID uint) error
}
