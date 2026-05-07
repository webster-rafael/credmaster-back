package router

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"gorm.io/gorm"

	"github.com/websterdev/cred-master/internal/config"
	domainauth "github.com/websterdev/cred-master/internal/domain/auth"
	domainboleto "github.com/websterdev/cred-master/internal/domain/boleto"
	domaincampaign "github.com/websterdev/cred-master/internal/domain/campaign"
	domainclient "github.com/websterdev/cred-master/internal/domain/client"
	domainmessage "github.com/websterdev/cred-master/internal/domain/message"
	domainwhatsapp "github.com/websterdev/cred-master/internal/domain/whatsapp"
	domainws "github.com/websterdev/cred-master/internal/domain/ws"
	appmiddleware "github.com/websterdev/cred-master/internal/middleware"
)

func New(db *gorm.DB, cfg *config.Config, authHandler *domainauth.Handler, msgHandler *domainmessage.Handler, clientHandler *domainclient.Handler, wsHandler *domainws.Handler, whatsHandler *domainwhatsapp.Handler, campaignHandler *domaincampaign.Handler, boletoHandler *domainboleto.Handler) http.Handler {
	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"http://localhost:5173"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	r.Handle("/uploads/*", http.StripPrefix("/uploads/", http.FileServer(http.Dir("uploads"))))

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"ok"}`))
	})

	r.Route("/api/v1", func(r chi.Router) {
		r.Route("/auth", func(r chi.Router) {
			r.Post("/login", authHandler.Login)
			r.Post("/register", authHandler.Register)
			r.Post("/forgot-password", authHandler.ForgotPassword)
		})

		// WebSocket — auth via ?token= query param (no middleware needed)
		r.Get("/ws", wsHandler.ServeWS)

		r.Group(func(r chi.Router) {
			r.Use(appmiddleware.Auth(cfg.JWTSecret))
			_ = db

			r.Route("/messages", func(r chi.Router) {
				r.Get("/conversations", msgHandler.GetConversations)
				r.Get("/thread", msgHandler.GetThread)
				r.Post("/send", msgHandler.Send)
				r.Post("/send-document", msgHandler.SendDocument)
				r.Delete("/conversation", msgHandler.DeleteConversation)
			})

			r.Route("/clients", func(r chi.Router) {
				r.Get("/", clientHandler.List)
				r.Post("/", clientHandler.Create)
				r.Delete("/{id}", clientHandler.Delete)
			})

			r.Route("/whatsapp", func(r chi.Router) {
				r.Get("/templates", whatsHandler.GetTemplates)
				r.Post("/templates", whatsHandler.CreateTemplate)
			})

			r.Route("/campaigns", func(r chi.Router) {
				r.Get("/", campaignHandler.List)
				r.Post("/", campaignHandler.Create)
				r.Patch("/{id}/dispatch", campaignHandler.Dispatch)
			})

			r.Route("/boletos", func(r chi.Router) {
				r.Get("/", boletoHandler.List)
				r.Post("/", boletoHandler.Create)
				r.Post("/bulk", boletoHandler.BulkCreate)
				r.Delete("/", boletoHandler.DeleteBulk)
			})
		})
	})

	return r
}
