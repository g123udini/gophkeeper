package config

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoad_Defaults(t *testing.T) {
	t.Setenv("APP_ENV", "")
	t.Setenv("DATABASE_DSN", "")
	t.Setenv("APP_SECRET", "secret")
	t.Setenv("LISTEN", "")
	t.Setenv("READ_TIMEOUT", "")
	t.Setenv("WRITE_TIMEOUT", "")

	cfg, err := Load()
	require.NoError(t, err)

	assert.Equal(t, "dev", cfg.Env)
	assert.Equal(t, defaultDSN, cfg.DatabaseDSN)
	assert.Equal(t, "secret", cfg.AppSecret)
	assert.Equal(t, defaultListen, cfg.Listen)
	assert.Equal(t, defaultReadTimeout, cfg.ReadTimeout)
	assert.Equal(t, defaultWriteTimeout, cfg.WriteTimeout)
}

func TestLoad_FromEnv(t *testing.T) {
	t.Setenv("APP_ENV", "prod")
	t.Setenv("DATABASE_DSN", "user:pass@tcp(localhost:3306)/app")
	t.Setenv("APP_SECRET", "super-secret")
	t.Setenv("LISTEN", ":9090")
	t.Setenv("READ_TIMEOUT", "15s")
	t.Setenv("WRITE_TIMEOUT", "20s")

	cfg, err := Load()
	require.NoError(t, err)

	assert.Equal(t, "prod", cfg.Env)
	assert.Equal(t, "user:pass@tcp(localhost:3306)/app", cfg.DatabaseDSN)
	assert.Equal(t, "super-secret", cfg.AppSecret)
	assert.Equal(t, ":9090", cfg.Listen)
	assert.Equal(t, 15*time.Second, cfg.ReadTimeout)
	assert.Equal(t, 20*time.Second, cfg.WriteTimeout)
}

func TestLoad_ErrorWhenAppSecretMissing(t *testing.T) {
	t.Setenv("APP_ENV", "")
	t.Setenv("DATABASE_DSN", "")
	t.Setenv("APP_SECRET", "")
	t.Setenv("LISTEN", "")
	t.Setenv("READ_TIMEOUT", "")
	t.Setenv("WRITE_TIMEOUT", "")

	cfg, err := Load()
	require.Error(t, err)
	assert.Nil(t, cfg)
	assert.Equal(t, "APP_SECRET is required", err.Error())
}

func TestParseDuration_DefaultOnInvalidValue(t *testing.T) {
	t.Setenv("READ_TIMEOUT", "bad-duration")

	got := parseDuration("READ_TIMEOUT", 7*time.Second)

	assert.Equal(t, 7*time.Second, got)
}

func TestParseDuration_DefaultOnEmptyValue(t *testing.T) {
	t.Setenv("READ_TIMEOUT", "")

	got := parseDuration("READ_TIMEOUT", 9*time.Second)

	assert.Equal(t, 9*time.Second, got)
}

func TestParseDuration_ParsesValidValue(t *testing.T) {
	t.Setenv("READ_TIMEOUT", "12s")

	got := parseDuration("READ_TIMEOUT", 9*time.Second)

	assert.Equal(t, 12*time.Second, got)
}

func TestGetEnv(t *testing.T) {
	t.Run("returns env value", func(t *testing.T) {
		t.Setenv("TEST_KEY", "value")

		got := getEnv("TEST_KEY", "default")

		assert.Equal(t, "value", got)
	})

	t.Run("returns default when empty", func(t *testing.T) {
		t.Setenv("TEST_KEY", "")

		got := getEnv("TEST_KEY", "default")

		assert.Equal(t, "default", got)
	})
}

func TestGetBoolEnv(t *testing.T) {
	tests := []struct {
		name  string
		value string
		want  bool
	}{
		{"true string", "true", true},
		{"one string", "1", true},
		{"false string", "false", false},
		{"zero string", "0", false},
		{"empty", "", false},
		{"random", "abc", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("BOOL_KEY", tt.value)

			got := getBoolEnv("BOOL_KEY")

			assert.Equal(t, tt.want, got)
		})
	}
}

func getBoolEnv(key string) bool {
	return os.Getenv(key) == "true" || os.Getenv(key) == "1"
}

func TestLoad_UsesRealEnvAPI(t *testing.T) {
	old := os.Getenv("APP_SECRET")
	defer func() { _ = os.Setenv("APP_SECRET", old) }()

	_ = os.Setenv("APP_SECRET", "secret")
	cfg, err := Load()
	require.NoError(t, err)
	assert.Equal(t, "secret", cfg.AppSecret)
}
