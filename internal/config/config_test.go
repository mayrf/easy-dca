package config

import (
	"testing"
)

func TestLoadConfig_Success(t *testing.T) {
	t.Setenv("EASY_DCA_PUBLIC_KEY", "test-public")
	t.Setenv("EASY_DCA_PRIVATE_KEY", "test-private")
	t.Setenv("EASY_DCA_DRY_RUN", "false")
	t.Setenv("EASY_DCA_PRICE_FACTOR", "0.95")
	t.Setenv("EASY_DCA_MONTHLY_FIAT_SPENDING", "60.0")

	cfg, err := LoadConfig()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if cfg.PublicKey != "test-public" {
		t.Errorf("expected public key 'test-public', got '%s'", cfg.PublicKey)
	}
	if cfg.PrivateKey != "test-private" {
		t.Errorf("expected private key 'test-private', got '%s'", cfg.PrivateKey)
	}
	if cfg.DryRun != false {
		t.Errorf("expected DryRun false, got %v", cfg.DryRun)
	}
	if cfg.PriceFactor != 0.95 {
		t.Errorf("expected PriceFactor 0.95, got %v", cfg.PriceFactor)
	}
	if cfg.MonthlyFiatSpending != 60.0 {
		t.Errorf("expected MonthlyFiatSpending 60.0, got %v", cfg.MonthlyFiatSpending)
	}
	if cfg.BuysPerMonth != 1 {
		t.Errorf("expected BuysPerMonth 1 (no cron), got %v", cfg.BuysPerMonth)
	}
}

func TestLoadConfig_AmountPerBuy(t *testing.T) {
	t.Setenv("EASY_DCA_PUBLIC_KEY", "test-public")
	t.Setenv("EASY_DCA_PRIVATE_KEY", "test-private")
	t.Setenv("EASY_DCA_FIAT_AMOUNT_PER_BUY", "10.0")

	cfg, err := LoadConfig()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if cfg.FiatAmountPerBuy != 10.0 {
		t.Errorf("expected FiatAmountPerBuy 10.0, got %v", cfg.FiatAmountPerBuy)
	}
}

func TestLoadConfig_MissingKeys(t *testing.T) {
	t.Setenv("EASY_DCA_PUBLIC_KEY", "")
	t.Setenv("EASY_DCA_PRIVATE_KEY", "")
	_, err := LoadConfig()
	if err == nil {
		t.Fatal("expected error for missing keys, got nil")
	}
}

func TestLoadConfig_PriceFactorAtMax(t *testing.T) {
	t.Setenv("EASY_DCA_PUBLIC_KEY", "test-public")
	t.Setenv("EASY_DCA_PRIVATE_KEY", "test-private")
	t.Setenv("EASY_DCA_FIAT_AMOUNT_PER_BUY", "10.0")
	t.Setenv("EASY_DCA_PRICE_FACTOR", "0.9999")
	cfg, err := LoadConfig()
	if err != nil {
		t.Fatalf("expected no error for price factor at maximum (0.9999), got %v", err)
	}
	if cfg.PriceFactor != 0.9999 {
		t.Errorf("expected PriceFactor 0.9999, got %v", cfg.PriceFactor)
	}
}

func TestLoadConfig_PriceFactorAboveMax(t *testing.T) {
	t.Setenv("EASY_DCA_PUBLIC_KEY", "test-public")
	t.Setenv("EASY_DCA_PRIVATE_KEY", "test-private")
	t.Setenv("EASY_DCA_FIAT_AMOUNT_PER_BUY", "10.0")
	t.Setenv("EASY_DCA_PRICE_FACTOR", "1.0")
	_, err := LoadConfig()
	if err == nil {
		t.Fatal("expected error for price factor above maximum (1.0), got nil")
	}
}

func TestLoadConfig_PriceFactorJustBelowMax(t *testing.T) {
	t.Setenv("EASY_DCA_PUBLIC_KEY", "test-public")
	t.Setenv("EASY_DCA_PRIVATE_KEY", "test-private")
	t.Setenv("EASY_DCA_FIAT_AMOUNT_PER_BUY", "10.0")
	t.Setenv("EASY_DCA_PRICE_FACTOR", "0.9998")
	cfg, err := LoadConfig()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if cfg.PriceFactor != 0.9998 {
		t.Errorf("expected PriceFactor 0.9998, got %v", cfg.PriceFactor)
	}
}

func TestLoadConfig_PriceFactorTooLow(t *testing.T) {
	t.Setenv("EASY_DCA_PUBLIC_KEY", "test-public")
	t.Setenv("EASY_DCA_PRIVATE_KEY", "test-private")
	t.Setenv("EASY_DCA_FIAT_AMOUNT_PER_BUY", "10.0")
	t.Setenv("EASY_DCA_PRICE_FACTOR", "0.94")
	_, err := LoadConfig()
	if err == nil {
		t.Fatal("expected error for price factor too low, got nil")
	}
}

func TestLoadConfig_ValidPriceFactor(t *testing.T) {
	t.Setenv("EASY_DCA_PUBLIC_KEY", "test-public")
	t.Setenv("EASY_DCA_PRIVATE_KEY", "test-private")
	t.Setenv("EASY_DCA_FIAT_AMOUNT_PER_BUY", "10.0")
	t.Setenv("EASY_DCA_PRICE_FACTOR", "0.97")
	cfg, err := LoadConfig()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if cfg.PriceFactor != 0.97 {
		t.Errorf("expected PriceFactor 0.97, got %v", cfg.PriceFactor)
	}
}

func TestLoadConfig_NoAmountSet(t *testing.T) {
	t.Setenv("EASY_DCA_PUBLIC_KEY", "test-public")
	t.Setenv("EASY_DCA_PRIVATE_KEY", "test-private")
	// Don't set any amount variables
	_, err := LoadConfig()
	if err == nil {
		t.Fatal("expected error for missing amount configuration, got nil")
	}
}

func TestLoadConfig_AutoAdjustMinOrderDefault(t *testing.T) {
	t.Setenv("EASY_DCA_PUBLIC_KEY", "test-public")
	t.Setenv("EASY_DCA_PRIVATE_KEY", "test-private")
	t.Setenv("EASY_DCA_FIAT_AMOUNT_PER_BUY", "10.0")
	
	cfg, err := LoadConfig()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if cfg.AutoAdjustMinOrder != false {
		t.Errorf("expected AutoAdjustMinOrder false (default), got %v", cfg.AutoAdjustMinOrder)
	}
}

func TestLoadConfig_AutoAdjustMinOrderEnabled(t *testing.T) {
	t.Setenv("EASY_DCA_PUBLIC_KEY", "test-public")
	t.Setenv("EASY_DCA_PRIVATE_KEY", "test-private")
	t.Setenv("EASY_DCA_FIAT_AMOUNT_PER_BUY", "10.0")
	t.Setenv("EASY_DCA_AUTO_ADJUST_MIN_ORDER", "true")
	
	cfg, err := LoadConfig()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if cfg.AutoAdjustMinOrder != true {
		t.Errorf("expected AutoAdjustMinOrder true, got %v", cfg.AutoAdjustMinOrder)
	}
}

func TestLoadConfig_AutoAdjustMinOrderDisabled(t *testing.T) {
	t.Setenv("EASY_DCA_PUBLIC_KEY", "test-public")
	t.Setenv("EASY_DCA_PRIVATE_KEY", "test-private")
	t.Setenv("EASY_DCA_FIAT_AMOUNT_PER_BUY", "10.0")
	t.Setenv("EASY_DCA_AUTO_ADJUST_MIN_ORDER", "false")
	
	cfg, err := LoadConfig()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if cfg.AutoAdjustMinOrder != false {
		t.Errorf("expected AutoAdjustMinOrder false, got %v", cfg.AutoAdjustMinOrder)
	}
}

func TestLoadConfig_SchedulerModeDefaultManual(t *testing.T) {
	t.Setenv("EASY_DCA_PUBLIC_KEY", "test-public")
	t.Setenv("EASY_DCA_PRIVATE_KEY", "test-private")
	t.Setenv("EASY_DCA_FIAT_AMOUNT_PER_BUY", "10.0")
	// Don't set EASY_DCA_CRON or EASY_DCA_SCHEDULER_MODE
	
	cfg, err := LoadConfig()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if cfg.SchedulerMode != "manual" {
		t.Errorf("expected SchedulerMode 'manual' (default), got '%s'", cfg.SchedulerMode)
	}
}

func TestLoadConfig_SchedulerModeDefaultCron(t *testing.T) {
	t.Setenv("EASY_DCA_PUBLIC_KEY", "test-public")
	t.Setenv("EASY_DCA_PRIVATE_KEY", "test-private")
	t.Setenv("EASY_DCA_FIAT_AMOUNT_PER_BUY", "10.0")
	t.Setenv("EASY_DCA_CRON", "0 8 * * *")
	// Don't set EASY_DCA_SCHEDULER_MODE
	
	cfg, err := LoadConfig()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if cfg.SchedulerMode != "cron" {
		t.Errorf("expected SchedulerMode 'cron' (default with cron), got '%s'", cfg.SchedulerMode)
	}
}

func TestLoadConfig_SchedulerModeExplicit(t *testing.T) {
	t.Setenv("EASY_DCA_PUBLIC_KEY", "test-public")
	t.Setenv("EASY_DCA_PRIVATE_KEY", "test-private")
	t.Setenv("EASY_DCA_FIAT_AMOUNT_PER_BUY", "10.0")
	t.Setenv("EASY_DCA_SCHEDULER_MODE", "systemd")
	
	cfg, err := LoadConfig()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if cfg.SchedulerMode != "systemd" {
		t.Errorf("expected SchedulerMode 'systemd', got '%s'", cfg.SchedulerMode)
	}
}

func TestCalculateBuysPerMonth_NoCron(t *testing.T) {
	buys, err := calculateBuysPerMonth("")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if buys != 1 {
		t.Errorf("expected 1 buy for no cron, got %d", buys)
	}
}

func TestCalculateBuysPerMonth_DailyCron(t *testing.T) {
	buys, err := calculateBuysPerMonth("0 8 * * *") // Daily at 8 AM
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if buys < 28 || buys > 31 {
		t.Errorf("expected ~30 buys for daily cron, got %d", buys)
	}
}

func TestCalculateBuysPerMonth_WeeklyCron(t *testing.T) {
	buys, err := calculateBuysPerMonth("0 8 * * 1") // Weekly on Monday at 8 AM
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if buys < 4 || buys > 5 {
		t.Errorf("expected ~4-5 buys for weekly cron, got %d", buys)
	}
}

func TestFormatBTC_SatsWithSeparators(t *testing.T) {
	cfg := &Config{DisplaySats: true}
	
	tests := []struct {
		amount float32
		want   string
	}{
		{0.00001, "1,000"},           // 1000 sats
		{0.0001, "10,000"},           // 10000 sats
		{0.001, "100,000"},           // 100000 sats
		{0.01, "1,000,000"},          // 1000000 sats
		{0.1, "10,000,000"},          // 10000000 sats
		{1.0, "100,000,000"},         // 100000000 sats
		{0.00005, "5,000"},           // 5000 sats
		{0.000123, "12,300"},         // 12300 sats
		{0.000999, "99,900"},         // 99900 sats
		{0.000001, "100"},            // 100 sats (no separator needed)
		{0.000009, "900"},            // 900 sats (no separator needed)
	}
	
	for _, tt := range tests {
		got := cfg.FormatBTC(tt.amount)
		if got != tt.want {
			t.Errorf("FormatBTC(%f) = %s, want %s", tt.amount, got, tt.want)
		}
	}
}

func TestFormatBTC_BTCWithoutSeparators(t *testing.T) {
	cfg := &Config{DisplaySats: false}
	
	tests := []struct {
		amount float32
		want   string
	}{
		{0.00001, "0.00001000"},
		{0.0001, "0.00010000"},
		{0.001, "0.00100000"},
		{0.01, "0.01000000"},
		{0.1, "0.10000000"},
		{1.0, "1.00000000"},
		{0.00005, "0.00005000"},
		{0.000123, "0.00012300"},
		{0.000999, "0.00099900"},
	}
	
	for _, tt := range tests {
		got := cfg.FormatBTC(tt.amount)
		if got != tt.want {
			t.Errorf("FormatBTC(%f) = %s, want %s", tt.amount, got, tt.want)
		}
	}
} 
