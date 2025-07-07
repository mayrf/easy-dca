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

// TradingPair represents a supported trading pair with validation
type TradingPair struct {
	value string
}

// Supported trading pairs
var supportedPairs = map[string]string{
	"BTC/EUR": "EUR",
	"BTC/GBP": "GBP", 
	"BTC/CHF": "CHF",
	"BTC/AUD": "AUD",
	"BTC/CAD": "CAD",
	"BTC/USD": "USD",
}

// NewTradingPair creates a new TradingPair with validation
func NewTradingPair(pair string) (TradingPair, error) {
	if _, exists := supportedPairs[pair]; !exists {
		return TradingPair{}, fmt.Errorf("unsupported trading pair: %s. Supported pairs: %s", 
			pair, strings.Join(GetSupportedPairs(), ", "))
	}
	return TradingPair{value: pair}, nil
}

// String returns the string representation of the trading pair
func (tp TradingPair) String() string {
	return tp.value
}

// GetFiatCurrency returns the fiat currency code for this trading pair
func (tp TradingPair) GetFiatCurrency() string {
	return supportedPairs[tp.value]
}

// GetFiatCurrencyName returns the full name of the fiat currency
func (tp TradingPair) GetFiatCurrencyName() string {
	currencyNames := map[string]string{
		"EUR": "Euro",
		"GBP": "British Pound",
		"CHF": "Swiss Franc", 
		"AUD": "Australian Dollar",
		"CAD": "Canadian Dollar",
		"USD": "US Dollar",
	}
	return currencyNames[tp.GetFiatCurrency()]
}

// GetSupportedPairs returns a slice of all supported trading pairs
func GetSupportedPairs() []string {
	pairs := make([]string, 0, len(supportedPairs))
	for pair := range supportedPairs {
		pairs = append(pairs, pair)
	}
	return pairs
}

// Config holds all configuration values for the application.
type Config struct {
	PublicKey           string      // Kraken API public key
	PrivateKey          string      // Kraken API private key
	Pair                TradingPair // Trading pair, e.g., BTC/EUR
	DryRun              bool        // If true, only validate orders (dry run); if false, actually place orders
	PriceFactor         float32     // Price factor for limit orders
	MonthlyFiatSpending float32     // Monthly fiat spending (optional, used if FiatAmountPerBuy is not set)
	FiatAmountPerBuy    float32     // Fixed fiat amount to spend each run (optional, takes precedence over MonthlyFiatSpending)
	AutoAdjustMinOrder  bool        // If true, automatically adjust orders below minimum size; if false, let them fail
	SchedulerMode       string      // Scheduler mode: "cron", "systemd", or "manual" (default: "cron" if EASY_DCA_CRON is set, otherwise "manual")

	DisplaySats         bool        // If true, display all BTC amounts in satoshi

	CronExpr     string // Cron expression for scheduling (optional)
	BuysPerMonth int    // Number of buys per month (calculated from cron expression)

	NotifyMethod    string // Notification method (ntfy, slack, email, etc.)
	NotifyNtfyTopic string // ntfy topic (if using ntfy)
	NotifyNtfyURL   string // ntfy server URL (if using ntfy)
	// Add more fields for other notification methods as needed
}

// ConfigureLogging sets up the log format based on environment variables
func ConfigureLogging() {
	logFormat := os.Getenv("EASY_DCA_LOG_FORMAT")
	
	switch strings.ToLower(logFormat) {
	case "timestamp", "time":
		// Standard timestamp format (2006/01/02 15:04:05)
		log.SetFlags(log.Ldate | log.Ltime)
	case "microseconds", "micro":
		// Full datetime with microseconds (2006/01/02 15:04:05.000000)
		log.SetFlags(log.Ldate | log.Ltime | log.Lmicroseconds)
	default:
		// Default: no timestamp prefix
		log.SetFlags(0)
	}
}

// logConfiguration prints a user-friendly summary of the loaded configuration
func logConfiguration(cfg Config) {
	log.Print("=== easy-dca Configuration Summary ===")
	
	// Trading pair
	log.Printf("📊 Trading pair: %s", cfg.Pair.String())

	// BTC unit
	log.Printf("🪙  BTC unit: %s", cfg.GetBTCUnit())
	
	// Execution mode
	if cfg.DryRun {
		log.Print("🔍 DRY RUN MODE: Orders will be validated but not executed")
	} else {
		log.Print("🚀 LIVE TRADING MODE: Orders will be placed on Kraken")
	}
	
	// Buy amount configuration
	if cfg.FiatAmountPerBuy > 0 {
		log.Printf("💰 Fixed amount per buy: %.2f %s", cfg.FiatAmountPerBuy, cfg.Pair.GetFiatCurrency())
	} else if cfg.MonthlyFiatSpending > 0 {
		log.Printf("💰 Monthly budget: %.2f %s (%.2f %s per buy, %d buys/month)", 
			cfg.MonthlyFiatSpending, cfg.Pair.GetFiatCurrency(), 
			cfg.MonthlyFiatSpending/float32(cfg.BuysPerMonth), cfg.Pair.GetFiatCurrency(), 
			cfg.BuysPerMonth)
	}
	
	// Price factor explanation
	log.Printf("📈 Price factor: %.4f (%.2f%% of ask price)", cfg.PriceFactor, cfg.PriceFactor*100)
	if cfg.PriceFactor >= 0.999 {
		log.Print("   → Very conservative: High fill probability, minimal savings")
	} else if cfg.PriceFactor >= 0.995 {
		log.Print("   → Conservative: Good fill probability, small savings")
	} else if cfg.PriceFactor >= 0.99 {
		log.Print("   → Balanced: Moderate fill probability, good savings")
	} else {
		log.Print("   → Aggressive: Lower fill probability, higher potential savings")
	}
	
	// Scheduling
	if cfg.SchedulerMode == "systemd" {
		log.Print("⏰ Schedule: Managed by systemd timer")
	} else if cfg.CronExpr != "" {
		log.Printf("⏰ Schedule: %s (%s mode)", cfg.CronExpr, cfg.SchedulerMode)
	} else {
		log.Printf("⏰ Schedule: Run once (%s mode)", cfg.SchedulerMode)
	}
	
	// Order behavior
	if cfg.AutoAdjustMinOrder {
		log.Print("🔧 Auto-adjustment: Enabled (orders below minimum will be increased)")
	} else {
		log.Print("🔧 Auto-adjustment: Disabled (orders below minimum may fail)")
	}
	
	// Notifications
	if cfg.NotifyMethod != "" {
		log.Printf("🔔 Notifications: %s", cfg.NotifyMethod)
		if cfg.NotifyMethod == "ntfy" {
			if cfg.NotifyNtfyURL != "" {
				log.Printf("   → ntfy server: %s", cfg.NotifyNtfyURL)
			}
			if cfg.NotifyNtfyTopic != "" {
				log.Printf("   → ntfy topic: %s", cfg.NotifyNtfyTopic)
			}
		}
	} else {
		log.Print("🔔 Notifications: Disabled")
	}
	
	// API key source
	if os.Getenv("EASY_DCA_PUBLIC_KEY_PATH") != "" {
		log.Print("🔑 API keys: Loaded from file paths (secure)")
	} else {
		log.Print("🔑 API keys: Loaded from environment variables")
	}
	
	log.Print("=====================================")
}

func getEnvAsFloat32(key string, defaultValue float32) float32 {
	if value := os.Getenv(key); value != "" {
		if floatValue, err := strconv.ParseFloat(value, 32); err == nil {
			return float32(floatValue)
		}
	}
	return defaultValue
}

func getEnvAsString(key string, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
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
	pairStr := getEnvAsString("EASY_DCA_PAIR", "BTC/EUR")
	pair, err := NewTradingPair(pairStr)
	if err != nil {
		return cfg, err
	}
	cfg.Pair = pair
	
	cfg.DryRun = getEnvAsBool("EASY_DCA_DRY_RUN", true)
	cfg.PriceFactor = getEnvAsFloat32("EASY_DCA_PRICE_FACTOR", 0.998)
	cfg.MonthlyFiatSpending = getEnvAsFloat32("EASY_DCA_MONTHLY_FIAT_SPENDING", 0.0)
	cfg.FiatAmountPerBuy = getEnvAsFloat32("EASY_DCA_FIAT_AMOUNT_PER_BUY", 0.0)
	cfg.AutoAdjustMinOrder = getEnvAsBool("EASY_DCA_AUTO_ADJUST_MIN_ORDER", false)
	cfg.CronExpr = os.Getenv("EASY_DCA_CRON")
	cfg.SchedulerMode = os.Getenv("EASY_DCA_SCHEDULER_MODE")
	cfg.DisplaySats = getEnvAsBool("EASY_DCA_DISPLAY_SATS", false)

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

	logConfiguration(cfg)

	return cfg, nil
}

// formatNumberWithSeparators formats a number with thousands separators
func formatNumberWithSeparators(n int64) string {
	if n < 1000 {
		return fmt.Sprintf("%d", n)
	}
	
	// Convert to string and add separators from right to left
	str := fmt.Sprintf("%d", n)
	result := ""
	for i, digit := range str {
		if i > 0 && (len(str)-i)%3 == 0 {
			result += "," // Use comma as separator
		}
		result += string(digit)
	}
	return result
}

// FormatBTC formats a BTC amount according to the display configuration
func (c *Config) FormatBTC(amount float32) string {
	if c.DisplaySats {
		sats := int64(amount * 100000000)
		return formatNumberWithSeparators(sats)
	}
	return fmt.Sprintf("%.8f", amount)
}

// GetBTCUnit returns the appropriate unit string for BTC amounts
func (c *Config) GetBTCUnit() string {
	if c.DisplaySats {
		return "sats"
	}
	return "BTC"
}
