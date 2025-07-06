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
