package config

import (
	"os"
	"testing"
	"time"
)

func TestGetEnv(t *testing.T) {
	tests := []struct {
		name         string
		key          string
		defaultValue string
		envValue     string
		expected     string
	}{
		{
			name:         "returns default when env not set",
			key:          "TEST_GET_ENV_NOT_SET",
			defaultValue: "default_value",
			envValue:     "",
			expected:     "default_value",
		},
		{
			name:         "returns env value when set",
			key:          "TEST_GET_ENV_SET",
			defaultValue: "default_value",
			envValue:     "env_value",
			expected:     "env_value",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.envValue != "" {
				os.Setenv(tt.key, tt.envValue)
				defer os.Unsetenv(tt.key)
			} else {
				os.Unsetenv(tt.key)
			}

			result := getEnv(tt.key, tt.defaultValue)
			if result != tt.expected {
				t.Errorf("getEnv(%s, %s) = %s, expected %s", tt.key, tt.defaultValue, result, tt.expected)
			}
		})
	}
}

func TestGetBoolEnv(t *testing.T) {
	tests := []struct {
		name     string
		key      string
		envValue string
		expected bool
	}{
		{
			name:     "returns false when env not set",
			key:      "TEST_GET_BOOL_ENV_NOT_SET",
			envValue: "",
			expected: false,
		},
		{
			name:     "returns true when env is 'true'",
			key:      "TEST_GET_BOOL_ENV_TRUE",
			envValue: "true",
			expected: true,
		},
		{
			name:     "returns true when env is '1'",
			key:      "TEST_GET_BOOL_ENV_ONE",
			envValue: "1",
			expected: true,
		},
		{
			name:     "returns false when env is not 'true' or '1'",
			key:      "TEST_GET_BOOL_ENV_OTHER",
			envValue: "yes",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.envValue != "" {
				os.Setenv(tt.key, tt.envValue)
				defer os.Unsetenv(tt.key)
			} else {
				os.Unsetenv(tt.key)
			}

			result := getBoolEnv(tt.key)
			if result != tt.expected {
				t.Errorf("getBoolEnv(%s) = %v, expected %v", tt.key, result, tt.expected)
			}
		})
	}
}

func TestParseDuration(t *testing.T) {
	tests := []struct {
		name         string
		key          string
		defaultValue time.Duration
		envValue     string
		expected     time.Duration
	}{
		{
			name:         "returns default when env not set",
			key:          "TEST_PARSE_DURATION_NOT_SET",
			defaultValue: 5 * time.Second,
			envValue:     "",
			expected:     5 * time.Second,
		},
		{
			name:         "returns parsed duration when env is valid",
			key:          "TEST_PARSE_DURATION_VALID",
			defaultValue: 5 * time.Second,
			envValue:     "10s",
			expected:     10 * time.Second,
		},
		{
			name:         "returns default when env is invalid",
			key:          "TEST_PARSE_DURATION_INVALID",
			defaultValue: 5 * time.Second,
			envValue:     "invalid",
			expected:     5 * time.Second,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.envValue != "" {
				os.Setenv(tt.key, tt.envValue)
				defer os.Unsetenv(tt.key)
			} else {
				os.Unsetenv(tt.key)
			}

			result := parseDuration(tt.key, tt.defaultValue)
			if result != tt.expected {
				t.Errorf("parseDuration(%s, %v) = %v, expected %v", tt.key, tt.defaultValue, result, tt.expected)
			}
		})
	}
}

func TestLoad(t *testing.T) {
	tests := []struct {
		name          string
		envs          map[string]string
		expectedError bool
		checkConfig   func(*Config) bool
	}{
		{
			name: "returns error when DATABASE_DSN is not set",
			envs: map[string]string{
				"APP_SECRET": "secret",
			},
			expectedError: true,
		},
		{
			name: "returns error when APP_SECRET is not set",
			envs: map[string]string{
				"DATABASE_DSN": "postgres://user:pass@localhost:5432/db",
			},
			expectedError: true,
		},
		{
			name: "returns config with defaults when required fields are set",
			envs: map[string]string{
				"DATABASE_DSN": "postgres://user:pass@localhost:5432/db",
				"APP_SECRET":   "secret",
			},
			expectedError: false,
			checkConfig: func(cfg *Config) bool {
				return cfg.Env == "dev" &&
					cfg.DatabaseDSN == "postgres://user:pass@localhost:5432/db" &&
					cfg.AppSecret == "secret" &&
					cfg.Listen == defaultListen &&
					cfg.ReadTimeout == defaultReadTimeout &&
					cfg.WriteTimeout == defaultWriteTimeout
			},
		},
		{
			name: "returns config with custom values when all fields are set",
			envs: map[string]string{
				"APP_ENV":       "prod",
				"DATABASE_DSN":  "postgres://user:pass@localhost:5432/prod_db",
				"APP_SECRET":    "prod_secret",
				"LISTEN":        ":8080",
				"READ_TIMEOUT":  "15s",
				"WRITE_TIMEOUT": "30s",
			},
			expectedError: false,
			checkConfig: func(cfg *Config) bool {
				return cfg.Env == "prod" &&
					cfg.DatabaseDSN == "postgres://user:pass@localhost:5432/prod_db" &&
					cfg.AppSecret == "prod_secret" &&
					cfg.Listen == ":8080" &&
					cfg.ReadTimeout == 15*time.Second &&
					cfg.WriteTimeout == 30*time.Second
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clear environment before each test
			os.Clearenv()

			// Set environment variables for this test
			for k, v := range tt.envs {
				os.Setenv(k, v)
			}

			cfg, err := Load()
			if tt.expectedError && err == nil {
				t.Error("Load() did not return expected error")
			}

			if !tt.expectedError && err != nil {
				t.Errorf("Load() returned unexpected error: %v", err)
			}

			if !tt.expectedError && err == nil {
				if tt.checkConfig != nil && !tt.checkConfig(cfg) {
					t.Error("Load() returned config with unexpected values")
				}
			}
		})
	}
}
