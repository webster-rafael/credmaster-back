package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/joho/godotenv"

	"github.com/websterdev/cred-master/internal/config"
	"github.com/websterdev/cred-master/internal/database"
	domainauth "github.com/websterdev/cred-master/internal/domain/auth"
	infraauth "github.com/websterdev/cred-master/internal/infrastructure/auth"
	"github.com/websterdev/cred-master/internal/router"
	authservice "github.com/websterdev/cred-master/internal/service/auth"
)

func main() {
	_ = godotenv.Load()

	cfg := config.Load()

	db, err := database.NewPostgresConnection(cfg)
	if err != nil {
		log.Fatalf("database connection failed: %v", err)
	}

	if err := db.AutoMigrate(&domainauth.User{}); err != nil {
		log.Fatalf("migration failed: %v", err)
	}

	authRepo := infraauth.NewPostgresRepository(db)
	authSvc := authservice.New(authRepo, cfg.JWTSecret)
	authHandler := domainauth.NewHandler(authSvc)

	r := router.New(db, cfg, authHandler)

	addr := fmt.Sprintf(":%s", cfg.Port)
	log.Printf("server listening on %s", addr)

	if err := http.ListenAndServe(addr, r); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
