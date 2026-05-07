package whatsapp

import "context"

type Repository interface {
	FindByCompanyID(ctx context.Context, companyID uint) (*WhatsConfig, error)
	UpdateWabaID(ctx context.Context, configID uint, wabaID string) error
}
