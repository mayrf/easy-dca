// Package config provides configuration loading and validation for the easy-dca application.
package config

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/robfig/cron/v3"
)

// Config holds all configuration values for the application.
type Config struct {
	PublicKey           string  // Kraken API public key
	PrivateKey          string  // Kraken API private key
	Pair                string  // Trading pair, e.g., BTC/EUR
	DryRun              bool    // If true, only validate orders (dry run); if false, actually place orders
	PriceFactor         float32 // Price factor for limit orders
	MonthlyFiatSpending float32 // Monthly fiat spending (optional, used if FiatAmountPerBuy is not set)
	FiatAmountPerBuy    float32 // Fixed fiat amount to spend each run (optional, takes precedence over MonthlyFiatSpending)
	AutoAdjustMinOrder  bool    // If true, automatically adjust orders below minimum size; if false, let them fail
	SchedulerMode       string  // Scheduler mode: "cron", "systemd", or "manual" (default: "cron" if EASY_DCA_CRON is set, otherwise "manual")

	CronExpr     string // Cron expression for scheduling (optional)
	BuysPerMonth int    // Number of buys per month (calculated from cron expression)

	NotifyMethod    string // Notification method (ntfy, slack, email, etc.)
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

// calculateBuysPerMonth calculates how many times the cron expression will run in a typical month
func calculateBuysPerMonth(cronExpr string) (int, error) {
	if cronExpr == "" {
		return 1, nil // If no cron, assume single run
	}

	parser := cron.NewParser(cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow)
	schedule, err := parser.Parse(cronExpr)
	if err != nil {
		return 0, fmt.Errorf("invalid cron expression: %w", err)
	}

	// Calculate runs for the next 31 days (to cover a full month)
	now := time.Now()
	endDate := now.AddDate(0, 0, 31)

	var runCount int
	next := schedule.Next(now)
	for next.Before(endDate) {
		runCount++
		next = schedule.Next(next)
	}

	// Check for intervals longer than a month
	if runCount == 0 {
		return 0, fmt.Errorf("cron schedule results in no runs within 31 days (interval longer than a month)")
	}

	// Warn for irregular schedules (less than 2 runs or more than 31 runs in a month)
	if runCount < 2 {
		log.Printf("Warning: Cron schedule results in only %d run(s) per month - this may not be optimal for DCA", runCount)
	} else if runCount > 31 {
		log.Printf("Warning: Cron schedule results in %d runs per month - this may be more frequent than intended", runCount)
	}

	return runCount, nil
}

// LoadConfig loads configuration from environment variables and files, validates it, and returns a Config struct.
// Returns an error if required configuration is missing or invalid.
func LoadConfig() (Config, error) {
	var cfg Config
	cfg.Pair = "BTC/EUR"

	// 1. Load and validate required API keys first (fail fast)
	publicKey, err := loadFileToString(os.Getenv("EASY_DCA_PUBLIC_KEY_PATH"))
	if err != nil {
		publicKey = os.Getenv("EASY_DCA_PUBLIC_KEY")
		if publicKey == "" {
			return cfg, fmt.Errorf("No PUBLIC_KEY found, neither via EASY_DCA_PUBLIC_KEY_PATH nor EASY_DCA_PUBLIC_KEY")
		}
	}
	cfg.PublicKey = strings.TrimSpace(publicKey)

	privateKey, err := loadFileToString(os.Getenv("EASY_DCA_PRIVATE_KEY_PATH"))
	if err != nil {
		privateKey = os.Getenv("EASY_DCA_PRIVATE_KEY")
		if privateKey == "" {
			return cfg, fmt.Errorf("No PRIVATE_KEY found, neither via EASY_DCA_PRIVATE_KEY_PATH nor EASY_DCA_PRIVATE_KEY")
		}
	}
	cfg.PrivateKey = strings.TrimSpace(privateKey)

	// 2. Load basic configuration
	cfg.DryRun = getEnvAsBool("EASY_DCA_DRY_RUN", true)
	cfg.PriceFactor = getEnvAsFloat32("EASY_DCA_PRICEFACTOR", 0.998)
	cfg.MonthlyFiatSpending = getEnvAsFloat32("EASY_DCA_MONTHLY_FIAT_SPENDING", 0.0)
	cfg.FiatAmountPerBuy = getEnvAsFloat32("EASY_DCA_FIAT_AMOUNT_PER_BUY", 0.0)
	cfg.AutoAdjustMinOrder = getEnvAsBool("EASY_DCA_AUTO_ADJUST_MIN_ORDER", false)
	cfg.CronExpr = os.Getenv("EASY_DCA_CRON")
	cfg.SchedulerMode = os.Getenv("EASY_DCA_SCHEDULER_MODE")

	// 3. Validate constraints immediately (fail fast)
	if cfg.PriceFactor > 0.9999 {
		return cfg, fmt.Errorf("priceFactor must be smaller than 0.9999 (99.99%% of ask price) to ensure maker orders")
	}
	if cfg.PriceFactor < 0.95 {
		return cfg, fmt.Errorf("priceFactor must be at least 0.95 (95%% of ask price) to ensure reasonable fill probability")
	}

	// 4. Validate amount configuration before complex calculations
	if cfg.FiatAmountPerBuy == 0 && cfg.MonthlyFiatSpending == 0 {
		return cfg, fmt.Errorf("either EASY_DCA_FIAT_AMOUNT_PER_BUY or EASY_DCA_MONTHLY_FIAT_SPENDING must be set")
	}

	if cfg.FiatAmountPerBuy > 0 && cfg.MonthlyFiatSpending > 0 {
		log.Printf("Warning: Both EASY_DCA_FIAT_AMOUNT_PER_BUY (%.2f) and EASY_DCA_MONTHLY_FIAT_SPENDING (%.2f) are set. Amount per buy takes precedence.", cfg.FiatAmountPerBuy, cfg.MonthlyFiatSpending)
	}

	// 5. Do complex calculations (cron parsing) only after basic validation passes
	buysPerMonth, err := calculateBuysPerMonth(cfg.CronExpr)
	if err != nil {
		return cfg, err
	}
	cfg.BuysPerMonth = buysPerMonth

	// Set default scheduler mode based on configuration
	if cfg.SchedulerMode == "" {
		if cfg.CronExpr != "" {
			cfg.SchedulerMode = "cron"
		} else {
			cfg.SchedulerMode = "manual"
		}
	}

	// 6. Load optional notification configuration
	cfg.NotifyMethod = os.Getenv("NOTIFY_METHOD")
	cfg.NotifyNtfyTopic = os.Getenv("NOTIFY_NTFY_TOPIC")
	cfg.NotifyNtfyURL = os.Getenv("NOTIFY_NTFY_URL")
	// Add more notification config as needed

	return cfg, nil
}
