package main

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/joho/godotenv"

	"github.com/websterdev/cred-master/internal/config"
	"github.com/websterdev/cred-master/internal/database"
	domainclient "github.com/websterdev/cred-master/internal/domain/client"
	domainauth "github.com/websterdev/cred-master/internal/domain/auth"
	domainmessage "github.com/websterdev/cred-master/internal/domain/message"
	domainwhatsapp "github.com/websterdev/cred-master/internal/domain/whatsapp"
	domainws "github.com/websterdev/cred-master/internal/domain/ws"
	infraauth "github.com/websterdev/cred-master/internal/infrastructure/auth"
	infraclient "github.com/websterdev/cred-master/internal/infrastructure/client"
	inframessage "github.com/websterdev/cred-master/internal/infrastructure/message"
	infrawhatsapp "github.com/websterdev/cred-master/internal/infrastructure/whatsapp"
	"github.com/websterdev/cred-master/internal/hub"
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

	if err := db.AutoMigrate(&domainauth.Company{}, &domainauth.User{}, &domainmessage.Message{}, &domainclient.Client{}, &domainwhatsapp.WhatsConfig{}); err != nil {
		log.Fatalf("migration failed: %v", err)
	}

	authRepo := infraauth.NewPostgresRepository(db)
	authSvc := authservice.New(authRepo, cfg.JWTSecret)
	authHandler := domainauth.NewHandler(authSvc)

	clientRepo := infraclient.NewPostgresRepository(db)
	clientHandler := domainclient.NewHandler(clientRepo)

	msgRepo := inframessage.NewPostgresRepository(db)
	whatsRepo := infrawhatsapp.NewPostgresRepository(db)
	msgHandler := domainmessage.NewHandler(msgRepo, whatsRepo, clientRepo, cfg.WebhookSendMessage)

	wsHub := hub.New()
	go wsHub.Run()
	hub.StartPGListener(context.Background(), cfg.DatabaseURL, wsHub)

	wsHandler := domainws.NewHandler(wsHub, cfg.JWTSecret)

	r := router.New(db, cfg, authHandler, msgHandler, clientHandler, wsHandler)

	addr := fmt.Sprintf(":%s", cfg.Port)
	log.Printf("server listening on %s", addr)

	if err := http.ListenAndServe(addr, r); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
