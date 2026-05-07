package config

import "os"

type Config struct {
	Port               string
	DatabaseURL        string
	JWTSecret          string
	Env                string
	WebhookSendMessage string
}

func Load() *Config {
	return &Config{
		Port:               getEnv("PORT", "8080"),
		DatabaseURL:        getEnv("DATABASE_URL", ""),
		JWTSecret:          getEnv("JWT_SECRET", ""),
		Env:                getEnv("ENV", "development"),
		WebhookSendMessage: getEnv("WEBHOOK_SENDMESSAGE", ""),
	}
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}
