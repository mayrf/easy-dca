// Package config provides configuration loading and validation for the easy-dca application.
package config

import (
	"fmt"
	"os"
	"strconv"
)

// Config holds all configuration values for the application.
type Config struct {
	PublicKey    string // Kraken API public key
	PrivateKey   string // Kraken API private key
	Pair         string // Trading pair, e.g., BTC/EUR
	DryRun       bool   // If true, only validate orders (dry run); if false, actually place orders
	PriceFactor  float32 // Price factor for limit orders
	MonthlyVolume float32 // Monthly trading volume
	DailyVolume   float32 // Daily trading volume (derived)

	CronExpr      string // Cron expression for scheduling (optional)

	NotifyMethod  string // Notification method (ntfy, slack, email, etc.)
	NotifyNtfyTopic string // ntfy topic (if using ntfy)
	NotifyNtfyURL   string // ntfy server URL (if using ntfy)
	// Add more fields for other notification methods as needed
}

func getEnvAsFloat32(key string, defaultValue float32) float32 {
	if value := os.Getenv(key); value != "" {
		if floatValue, err := strconv.ParseFloat(value, 32); err == nil {
			return float32(floatValue)
		}
	}
	return defaultValue
}

func getEnvAsBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if boolValue, err := strconv.ParseBool(value); err == nil {
			return boolValue
		}
	}
	return defaultValue
}

func loadFileToString(filepath string) (string, error) {
	content, err := os.ReadFile(filepath)
	if err != nil {
		return "", fmt.Errorf("failed to read file %s: %w", filepath, err)
	}
	return string(content), nil
}

// LoadConfig loads configuration from environment variables and files, validates it, and returns a Config struct.
// Returns an error if required configuration is missing or invalid.
func LoadConfig() (Config, error) {
	var cfg Config
	cfg.Pair = "BTC/EUR"

	publicKey, err := loadFileToString(os.Getenv("EASY_DCA_PUBLIC_KEY_PATH"))
	if err != nil {
		publicKey = os.Getenv("EASY_DCA_PUBLIC_KEY")
		if publicKey == "" {
			return cfg, fmt.Errorf("No PUBLIC_KEY found, neither via EASY_DCA_PUBLIC_KEY_PATH nor EASY_DCA_PUBLIC_KEY")
		}
	}
	cfg.PublicKey = publicKey

	privateKey, err := loadFileToString(os.Getenv("EASY_DCA_PRIVATE_KEY_PATH"))
	if err != nil {
		privateKey = os.Getenv("EASY_DCA_PRIVATE_KEY")
		if privateKey == "" {
			return cfg, fmt.Errorf("No PRIVATE_KEY found, neither via EASY_DCA_PRIVATE_KEY_PATH nor EASY_DCA_PRIVATE_KEY")
		}
	}
	cfg.PrivateKey = privateKey

	cfg.DryRun = getEnvAsBool("EASY_DCA_DRY_RUN", true)
	cfg.PriceFactor = getEnvAsFloat32("EASY_DCA_PRICEFACTOR", 0.998)
	cfg.MonthlyVolume = getEnvAsFloat32("EASY_DCA_MONTHLY_VOLUME", 140.0)
	cfg.DailyVolume = cfg.MonthlyVolume / 30.0

	if cfg.PriceFactor >= 1 {
		return cfg, fmt.Errorf("priceFactor must be smaller than 1 in order to place a limit order as a maker")
	}

	cfg.CronExpr = os.Getenv("EASY_DCA_CRON")

	cfg.NotifyMethod = os.Getenv("NOTIFY_METHOD")
	cfg.NotifyNtfyTopic = os.Getenv("NOTIFY_NTFY_TOPIC")
	cfg.NotifyNtfyURL = os.Getenv("NOTIFY_NTFY_URL")
	// Add more notification config as needed

	return cfg, nil
} 