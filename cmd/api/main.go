package main

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/joho/godotenv"

	"github.com/websterdev/cred-master/internal/config"
	"github.com/websterdev/cred-master/internal/database"
	domainauth "github.com/websterdev/cred-master/internal/domain/auth"
	domainboleto "github.com/websterdev/cred-master/internal/domain/boleto"
	domaincampaign "github.com/websterdev/cred-master/internal/domain/campaign"
	domainclient "github.com/websterdev/cred-master/internal/domain/client"
	domainmessage "github.com/websterdev/cred-master/internal/domain/message"
	domainwhatsapp "github.com/websterdev/cred-master/internal/domain/whatsapp"
	domainws "github.com/websterdev/cred-master/internal/domain/ws"
	infraauth "github.com/websterdev/cred-master/internal/infrastructure/auth"
	infroboleto "github.com/websterdev/cred-master/internal/infrastructure/boleto"
	infracampaign "github.com/websterdev/cred-master/internal/infrastructure/campaign"
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

	if err := db.AutoMigrate(&domainauth.Company{}, &domainauth.User{}, &domainmessage.Message{}, &domainclient.Client{}, &domainwhatsapp.WhatsConfig{}, &domaincampaign.Campaign{}, &domainboleto.Boleto{}); err != nil {
		log.Fatalf("migration failed: %v", err)
	}

	authRepo := infraauth.NewPostgresRepository(db)
	authSvc := authservice.New(authRepo, cfg.JWTSecret)
	authHandler := domainauth.NewHandler(authSvc)

	clientRepo := infraclient.NewPostgresRepository(db)
	clientHandler := domainclient.NewHandler(clientRepo)

	msgRepo := inframessage.NewPostgresRepository(db)
	whatsRepo := infrawhatsapp.NewPostgresRepository(db)
	msgHandler := domainmessage.NewHandler(msgRepo, whatsRepo, clientRepo, cfg.WebhookSendMessage, cfg.BaseURL)

	wsHub := hub.New()
	go wsHub.Run()
	hub.StartPGListener(context.Background(), cfg.DatabaseURL, wsHub)

	wsHandler := domainws.NewHandler(wsHub, cfg.JWTSecret)

	whatsHandler := domainwhatsapp.NewHandler(whatsRepo)

	boletoRepo := infroboleto.NewPostgresRepository(db)
	boletoHandler := domainboleto.NewHandler(boletoRepo, cfg.BaseURL)

	campaignRepo := infracampaign.NewPostgresRepository(db)
	campaignHandler := domaincampaign.NewHandler(campaignRepo, clientRepo, boletoRepo, whatsRepo, cfg.TriggerTemplatesWebhook, cfg.BaseURL)

	r := router.New(db, cfg, authHandler, msgHandler, clientHandler, wsHandler, whatsHandler, campaignHandler, boletoHandler)

	addr := fmt.Sprintf(":%s", cfg.Port)
	log.Printf("server listening on %s", addr)

	if err := http.ListenAndServe(addr, r); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
