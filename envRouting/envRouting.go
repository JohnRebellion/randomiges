// Package envRouting that provides environment variable access with static and secure bindings/configuration
package envRouting

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

var (
	// Port ...
	Port string

	// SecretKey ...
	SecretKey string

	// StaticWebLocation ...
	StaticWebLocation string

	// DatabaseName ...
	DatabaseName string

	// SQLiteFilename ...
	SQLiteFilename string

	// PostgresUsername ...
	PostgresUsername string

	// PostgresPassword ...
	PostgresPassword string

	// PostgresHost ...
	PostgresHost string

	// PostgresPort ...
	PostgresPort string

	// PostgresSSLMode ...
	PostgresSSLMode string

	// PostgresTimezone ...
	PostgresTimezone string

	// PostgresURL ...
	PostgresURL string

	// MailEmail ...
	MailEmail string

	// SendGridAPIKey ...
	SendGridAPIKey string

	// DevelopmentCloudName ...
	DevelopmentCloudName string

	// DevelopmentApiKey ...
	DevelopmentApiKey string

	// DevelopmentApiSecret ...
	DevelopmentApiSecret string

	// DevelopmentSecure ...
	DevelopmentSecure string

	// ProductionCloudName ...
	ProductionCloudName string

	// ProductionApiKey ...
	ProductionApiKey string

	// ProductionApiSecret ...
	ProductionApiSecret string

	// ProductionSecure ...
	ProductionSecure string

	// PexelsAPIKey ...
	PexelsAPIKey string
)

// LoadEnv Staticly load environment variables
func LoadEnv() {
	Port = getEnv("PORT")
	SecretKey = getEnv("SECRET_KEY")
	StaticWebLocation = getEnv("STATIC_WEB_LOCATION")
	DatabaseName = getEnv("DATABASE_NAME")

	SQLiteFilename = getEnv("SQLITE_FILENAME")

	PostgresUsername = getEnv("POSTGRES_USERNAME")
	PostgresPassword = getEnv("POSTGRES_PASSWORD")
	PostgresHost = getEnv("POSTGRES_HOST")
	PostgresPort = getEnv("POSTGRES_PORT")
	PostgresSSLMode = getEnv("POSTGRES_SSL_MODE")
	PostgresTimezone = getEnv("POSTGRES_TIMEZONE")

	MailEmail = getEnv("MAIL_EMAIL")

	PostgresURL = getEnv("POSTGRES_URL")

	SendGridAPIKey = getEnv("SENDGRID_API_KEY")

	DevelopmentCloudName = getEnv("DEVELOPMENT_CLOUD_NAME")
	DevelopmentApiKey = getEnv("DEVELOPMENT_API_KEY")
	DevelopmentApiSecret = getEnv("DEVELOPMENT_API_SECRET")
	DevelopmentSecure = getEnv("DEVELOPMENT_SECURE")

	ProductionCloudName = getEnv("PRODUCTION_CLOUD_NAME")
	ProductionApiKey = getEnv("PRODUCTION_API_KEY")
	ProductionApiSecret = getEnv("PRODUCTION_API_SECRET")
	ProductionSecure = getEnv("PRODUCTION_SECURE")

	PexelsAPIKey = getEnv("PEXELS_API_KEY")
}

func getEnv(key string) string {
	err := godotenv.Load(".env")

	if err != nil {
		log.Fatalf("Error loading .env file")
	}

	return os.Getenv(key)
}
