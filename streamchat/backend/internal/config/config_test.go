package config

import "testing"

func TestLoadRequiresDatabaseAndStrongJWTSecret(t *testing.T) {
	t.Setenv("DATABASE_URL", "")
	t.Setenv("JWT_SECRET", "short")
	if _, err := Load(); err == nil {
		t.Fatal("expected missing database error")
	}
	t.Setenv("DATABASE_URL", "postgres://localhost/test")
	if _, err := Load(); err == nil {
		t.Fatal("expected weak JWT secret error")
	}
}

func TestLoadRejectsInvalidMediaLimits(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://localhost/test")
	t.Setenv("JWT_SECRET", "01234567890123456789012345678901")
	t.Setenv("MEDIA_MAX_SIZE_MB", "0")
	if _, err := Load(); err == nil {
		t.Fatal("expected invalid media limit error")
	}
}

func TestLoadAcceptsValidConfiguration(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://localhost/test")
	t.Setenv("JWT_SECRET", "01234567890123456789012345678901")
	cfg, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.APIPort != "8080" || cfg.RedisAddr != "localhost:6379" {
		t.Fatalf("unexpected defaults: %+v", cfg)
	}
}
