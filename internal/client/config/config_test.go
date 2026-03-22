package config

import (
	"bytes"
	"flag"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func resetFlags(output *bytes.Buffer) {
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ContinueOnError)
	flag.CommandLine.SetOutput(output)
}

func TestParseArgs_Defaults(t *testing.T) {
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	var buf bytes.Buffer
	resetFlags(&buf)

	os.Args = []string{"app", "test.db", "secret"}

	cfg, err := ParseArgs()
	require.NoError(t, err)

	assert.Equal(t, "test.db", cfg.DBPath)
	assert.Equal(t, "secret", cfg.MasterPassword)
	assert.Equal(t, "localhost:50501", cfg.ServerAddr)
	assert.Equal(t, 60, cfg.SyncIntervalSec)
}

func TestParseArgs_WithFlags(t *testing.T) {
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	var buf bytes.Buffer
	resetFlags(&buf)

	os.Args = []string{
		"app",
		"-addr", "127.0.0.1:9000",
		"-interval", "30",
		"custom.db",
		"pass123",
	}

	cfg, err := ParseArgs()
	require.NoError(t, err)

	assert.Equal(t, "custom.db", cfg.DBPath)
	assert.Equal(t, "pass123", cfg.MasterPassword)
	assert.Equal(t, "127.0.0.1:9000", cfg.ServerAddr)
	assert.Equal(t, 30, cfg.SyncIntervalSec)
}

func TestParseArgs_InvalidArgs(t *testing.T) {
	tests := []struct {
		name string
		args []string
	}{
		{"no args", []string{"app"}},
		{"one arg", []string{"app", "db"}},
		{"too many args", []string{"app", "a", "b", "c"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			oldArgs := os.Args
			defer func() { os.Args = oldArgs }()

			var buf bytes.Buffer
			resetFlags(&buf)

			os.Args = tt.args

			cfg, err := ParseArgs()

			assert.Error(t, err)
			assert.Nil(t, cfg)

			// проверяем что usage был вызван
			out := buf.String()
			assert.Contains(t, out, "usage:")
			assert.Contains(t, out, "<db_path> <master_password>")
		})
	}
}

func TestParseArgs_UsageOutput(t *testing.T) {
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	var buf bytes.Buffer
	resetFlags(&buf)

	os.Args = []string{"app"} // invalid

	_, _ = ParseArgs()

	out := buf.String()

	assert.Contains(t, out, "usage:")
	assert.Contains(t, out, "server address")
	assert.Contains(t, out, "synchronization interval")
}
