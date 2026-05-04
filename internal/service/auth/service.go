package authservice

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	domain "github.com/websterdev/cred-master/internal/domain/auth"
)

type service struct {
	repo      domain.Repository
	jwtSecret []byte
}

func New(repo domain.Repository, jwtSecret string) domain.Service {
	return &service{
		repo:      repo,
		jwtSecret: []byte(jwtSecret),
	}
}

type accessClaims struct {
	UserID uint   `json:"sub"`
	Email  string `json:"email"`
	Name   string `json:"name"`
	Type   string `json:"type"`
	jwt.RegisteredClaims
}

type refreshClaims struct {
	UserID uint   `json:"sub"`
	Type   string `json:"type"`
	jwt.RegisteredClaims
}

type resetClaims struct {
	UserID uint   `json:"sub"`
	Email  string `json:"email"`
	Type   string `json:"type"`
	jwt.RegisteredClaims
}

func (s *service) generateTokenPair(user *domain.User) (*domain.TokenPair, error) {
	now := time.Now()

	access := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims{
		UserID: user.ID,
		Email:  user.Email,
		Name:   user.Name,
		Type:   "access",
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(15 * time.Minute)),
		},
	})

	refresh := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims{
		UserID: user.ID,
		Type:   "refresh",
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(7 * 24 * time.Hour)),
		},
	})

	accessSigned, err := access.SignedString(s.jwtSecret)
	if err != nil {
		return nil, fmt.Errorf("signing access token: %w", err)
	}

	refreshSigned, err := refresh.SignedString(s.jwtSecret)
	if err != nil {
		return nil, fmt.Errorf("signing refresh token: %w", err)
	}

	return &domain.TokenPair{
		AccessToken:  accessSigned,
		RefreshToken: refreshSigned,
	}, nil
}

func (s *service) Register(ctx context.Context, user *domain.User) error {
	existing, err := s.repo.FindByEmail(ctx, user.Email)
	if err == nil && existing != nil {
		return domain.ErrEmailAlreadyRegistered
	}
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return fmt.Errorf("checking email: %w", err)
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("hashing password: %w", err)
	}

	user.Password = string(hashed)
	return s.repo.Create(ctx, user)
}

func (s *service) Login(ctx context.Context, creds domain.Credentials) (*domain.User, *domain.TokenPair, error) {
	user, err := s.repo.FindByEmail(ctx, creds.Email)
	if err != nil {
		return nil, nil, errors.New("invalid credentials")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(creds.Password)); err != nil {
		return nil, nil, errors.New("invalid credentials")
	}

	pair, err := s.generateTokenPair(user)
	if err != nil {
		return nil, nil, fmt.Errorf("generating tokens: %w", err)
	}

	return user, pair, nil
}

func (s *service) ForgotPassword(ctx context.Context, email string) (*domain.ForgotPasswordResult, error) {
	user, err := s.repo.FindByEmail(ctx, email)
	if err != nil {
		return &domain.ForgotPasswordResult{ResetToken: ""}, nil
	}

	now := time.Now()
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, resetClaims{
		UserID: user.ID,
		Email:  user.Email,
		Type:   "reset",
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(30 * time.Minute)),
		},
	})

	signed, err := token.SignedString(s.jwtSecret)
	if err != nil {
		return nil, fmt.Errorf("signing reset token: %w", err)
	}

	return &domain.ForgotPasswordResult{ResetToken: signed}, nil
}
