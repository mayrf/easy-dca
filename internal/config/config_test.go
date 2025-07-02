package config

import (
	"testing"
)

func TestLoadConfig_Success(t *testing.T) {
	t.Setenv("EASY_DCA_PUBLIC_KEY", "test-public")
	t.Setenv("EASY_DCA_PRIVATE_KEY", "test-private")
	t.Setenv("EASY_DCA_VALIDATION_ON", "false")
	t.Setenv("EASY_DCA_PRICEFACTOR", "0.95")
	t.Setenv("EASY_DCA_MONTHLY_VOLUME", "60.0")

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
	if cfg.OrderValidation != false {
		t.Errorf("expected OrderValidation false, got %v", cfg.OrderValidation)
	}
	if cfg.PriceFactor != 0.95 {
		t.Errorf("expected PriceFactor 0.95, got %v", cfg.PriceFactor)
	}
	if cfg.MonthlyVolume != 60.0 {
		t.Errorf("expected MonthlyVolume 60.0, got %v", cfg.MonthlyVolume)
	}
	if cfg.DailyVolume != 2.0 {
		t.Errorf("expected DailyVolume 2.0, got %v", cfg.DailyVolume)
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

func TestLoadConfig_InvalidPriceFactor(t *testing.T) {
	t.Setenv("EASY_DCA_PUBLIC_KEY", "test-public")
	t.Setenv("EASY_DCA_PRIVATE_KEY", "test-private")
	t.Setenv("EASY_DCA_PRICEFACTOR", "1.5")
	_, err := LoadConfig()
	if err == nil {
		t.Fatal("expected error for invalid price factor, got nil")
	}
} 