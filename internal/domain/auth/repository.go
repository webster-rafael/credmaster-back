package auth

import "context"

type Repository interface {
	FindByEmail(ctx context.Context, email string) (*User, error)
	Create(ctx context.Context, user *User) error
	CreateCompany(ctx context.Context, company *Company) error
}
