package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
)

type contextKey string

const (
	contextKeyUserID    contextKey = "user_id"
	contextKeyCompanyID contextKey = "company_id"
	contextKeyRole      contextKey = "role"
)

func GetUserID(ctx context.Context) uint {
	v, _ := ctx.Value(contextKeyUserID).(uint)
	return v
}

func GetCompanyID(ctx context.Context) uint {
	v, _ := ctx.Value(contextKeyCompanyID).(uint)
	return v
}

func GetRole(ctx context.Context) string {
	v, _ := ctx.Value(contextKeyRole).(string)
	return v
}

func Auth(jwtSecret string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
				return
			}

			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) != 2 || parts[0] != "Bearer" {
				http.Error(w, `{"error":"invalid authorization header"}`, http.StatusUnauthorized)
				return
			}

			token, err := jwt.Parse(parts[1], func(t *jwt.Token) (interface{}, error) {
				if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, jwt.ErrSignatureInvalid
				}
				return []byte(jwtSecret), nil
			})
			if err != nil || !token.Valid {
				http.Error(w, `{"error":"invalid token"}`, http.StatusUnauthorized)
				return
			}

			claims, ok := token.Claims.(jwt.MapClaims)
			if !ok {
				http.Error(w, `{"error":"invalid token claims"}`, http.StatusUnauthorized)
				return
			}

			var userID uint
			if v, ok := claims["sub"].(float64); ok {
				userID = uint(v)
			}

			var companyID uint
			if v, ok := claims["company_id"].(float64); ok {
				companyID = uint(v)
			}

			role, _ := claims["role"].(string)

			ctx := context.WithValue(r.Context(), contextKeyUserID, userID)
			ctx = context.WithValue(ctx, contextKeyCompanyID, companyID)
			ctx = context.WithValue(ctx, contextKeyRole, role)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
