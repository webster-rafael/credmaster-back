package contact

import "context"

type Service interface {
	ListContacts(ctx context.Context) ([]Contact, error)
	GetContact(ctx context.Context, id uint) (*Contact, error)
	CreateContact(ctx context.Context, contact *Contact) error
	UpdateContact(ctx context.Context, contact *Contact) error
	DeleteContact(ctx context.Context, id uint) error
}
