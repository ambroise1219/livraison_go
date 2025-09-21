package config

import (
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	// Database Configuration (PostgreSQL)
	DatabaseURL string

	// LMDB Configuration
	LMDBPath string

	// Server Configuration
	ServerPort string
	ServerHost string

	// JWT Configuration
	JWTSecret     string
	JWTExpiration int // hours

	// OTP Configuration
	OTPExpiration    int // minutes
	OTPLength        int
	OTPMaxPerWindow  int // per phone
	OTPWindowMinutes int // minutes

	// SMS Configuration (pour notifications)
	SMSAPIKey    string
	SMSAPISecret string

	// Email Configuration
	SMTPHost     string
	SMTPPort     string
	SMTPUsername string
	SMTPPassword string

	// Application Settings
	Environment string
	Debug       bool

	// Platform Commission Settings
	DefaultCommissionRate float64
	DefaultServiceFee     float64

	// Referral Settings
	ReferralRewardAmount float64
	ReferralExpiration   int // days

	// Wanotifier (WhatsApp) Webhook
	WanotifierWebhookURL string
	WanotifierDebug      bool
}

var AppConfig *Config

func LoadConfig() *Config {
	// Load .env file if exists
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}

	config := &Config{
		// Database
		DatabaseURL: getEnv("DATABASE_URL", "postgresql://livraison_user:livraison_pass@localhost:5432/livraison_db?sslmode=disable"),

		// LMDB
		LMDBPath: getEnv("LMDB_PATH", "/home/ubuntu/www/livraison_go/data/livraison_cache"),

		// Server
		ServerPort: getEnv("SERVER_PORT", "8080"),
		ServerHost: getEnv("SERVER_HOST", "0.0.0.0"),

		// JWT
		JWTSecret:     getEnv("JWT_SECRET", "ilex-secret-key-2024"),
		JWTExpiration: getEnvInt("JWT_EXPIRATION", 24), // 24 hours

		// OTP
		OTPExpiration:    getEnvInt("OTP_EXPIRATION", 5), // 5 minutes
		OTPLength:        getEnvInt("OTP_LENGTH", 4),
		OTPMaxPerWindow:  getEnvInt("OTP_MAX_PER_WINDOW", 3),
		OTPWindowMinutes: getEnvInt("OTP_WINDOW_MINUTES", 15),

		// SMS
		SMSAPIKey:    getEnv("SMS_API_KEY", ""),
		SMSAPISecret: getEnv("SMS_API_SECRET", ""),

		// Email
		SMTPHost:     getEnv("SMTP_HOST", ""),
		SMTPPort:     getEnv("SMTP_PORT", "587"),
		SMTPUsername: getEnv("SMTP_USERNAME", ""),
		SMTPPassword: getEnv("SMTP_PASSWORD", ""),

		// App
		Environment: getEnv("ENVIRONMENT", "development"),
		Debug:       getEnvBool("DEBUG", true),

		// Platform
		DefaultCommissionRate: getEnvFloat("DEFAULT_COMMISSION_RATE", 0.15), // 15%
		DefaultServiceFee:     getEnvFloat("DEFAULT_SERVICE_FEE", 500.0),    // 500 FCFA

		// Referral
		ReferralRewardAmount: getEnvFloat("REFERRAL_REWARD_AMOUNT", 1000.0), // 1000 FCFA
		ReferralExpiration:   getEnvInt("REFERRAL_EXPIRATION", 30),          // 30 days

		// Wanotifier
		WanotifierWebhookURL: getEnv("WANOTIFIER_WEBHOOK_URL", ""),
		WanotifierDebug:      getEnvBool("WANOTIFIER_DEBUG", false),
	}

	AppConfig = config
	return config
}

// GetConfig returns the loaded configuration or loads it if not loaded
func GetConfig() *Config {
	if AppConfig == nil {
		return LoadConfig()
	}
	return AppConfig
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := parseInt(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

func getEnvFloat(key string, defaultValue float64) float64 {
	if value := os.Getenv(key); value != "" {
		if floatValue, err := parseFloat(value); err == nil {
			return floatValue
		}
	}
	return defaultValue
}

func getEnvBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		return value == "true" || value == "1"
	}
	return defaultValue
}

// Helper functions
func parseInt(s string) (int, error) {
	var result int
	_, err := fmt.Sscanf(s, "%d", &result)
	return result, err
}

func parseFloat(s string) (float64, error) {
	var result float64
	_, err := fmt.Sscanf(s, "%f", &result)
	return result, err
}
