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
	AccessToken  string `json:"accessToken"`
	RefreshToken string `json:"refreshToken"`
}

type ForgotPasswordResult struct {
	ResetToken string
}

type Service interface {
	Login(ctx context.Context, credentials Credentials) (*User, *TokenPair, error)
	Register(ctx context.Context, user *User, companyName string) error
	ForgotPassword(ctx context.Context, email string) (*ForgotPasswordResult, error)
}
