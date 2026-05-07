package config

import "os"

type Config struct {
	Port               string
	DatabaseURL        string
	JWTSecret          string
	Env                string
	WebhookSendMessage string
	BaseURL            string
}

func Load() *Config {
	return &Config{
		Port:               getEnv("PORT", "8080"),
		DatabaseURL:        getEnv("DATABASE_URL", ""),
		JWTSecret:          getEnv("JWT_SECRET", ""),
		Env:                getEnv("ENV", "development"),
		WebhookSendMessage: getEnv("WEBHOOK_SENDMESSAGE", ""),
		BaseURL:            getEnv("BASE_URL", "http://localhost:8080"),
	}
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}
