package auth

import (
	"context"
	"errors"
)

var ErrEmailAlreadyRegistered = errors.New("email already registered")

type Credentials struct {
	Email    string
	Password string
}

type TokenPair struct {
	AccessToken  string
	RefreshToken string
}

type ForgotPasswordResult struct {
	ResetToken string
}

type Service interface {
	Login(ctx context.Context, credentials Credentials) (*User, *TokenPair, error)
	Register(ctx context.Context, user *User) error
	ForgotPassword(ctx context.Context, email string) (*ForgotPasswordResult, error)
}
