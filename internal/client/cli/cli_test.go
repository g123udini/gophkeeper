package cli

import (
	"bytes"
	"context"
	"errors"
	"os"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest/observer"

	"github.com/g123udini/gophkeeper/internal/client/command"
	"github.com/g123udini/gophkeeper/internal/common/logger"
)

func TestRun(t *testing.T) {
	tests := []struct {
		name              string
		input             string
		registry          CommandRegistry
		expectedOutput    string
		expectedLogChecks func(t *testing.T, logs []observer.LoggedEntry)
	}{
		{
			name:  "login success",
			input: "login user pass\n",
			registry: CommandRegistry{
				"login": func(ctx context.Context, args []string) tea.Cmd {
					if len(args) != 2 || args[0] != "user" || args[1] != "pass" {
						t.Errorf("unexpected args: %v", args)
					}

					return func() tea.Msg {
						return command.LoginSuccessMsg{}
					}
				},
			},
			expectedOutput: "login successful\n> ",
			expectedLogChecks: func(t *testing.T, logs []observer.LoggedEntry) {
			},
		},
		{
			name:  "register success",
			input: "register user pass\n",
			registry: CommandRegistry{
				"register": func(ctx context.Context, args []string) tea.Cmd {
					if len(args) != 2 || args[0] != "user" || args[1] != "pass" {
						t.Errorf("unexpected args: %v", args)
					}

					return func() tea.Msg {
						return command.RegisterSuccessMsg{}
					}
				},
			},
			expectedOutput: "register successful\n> ",
			expectedLogChecks: func(t *testing.T, logs []observer.LoggedEntry) {
			},
		},
		{
			name:  "get success",
			input: "get secret\n",
			registry: CommandRegistry{
				"get": func(ctx context.Context, args []string) tea.Cmd {
					if len(args) != 1 || args[0] != "secret" {
						t.Errorf("unexpected args: %v", args)
					}

					return func() tea.Msg {
						return command.GetSuccessMsg{Value: "my-value"}
					}
				},
			},
			expectedOutput: "my-value\n> ",
			expectedLogChecks: func(t *testing.T, logs []observer.LoggedEntry) {
			},
		},
		{
			name:  "set success",
			input: "set key text value\n",
			registry: CommandRegistry{
				"set": func(ctx context.Context, args []string) tea.Cmd {
					if len(args) != 3 || args[0] != "key" || args[1] != "text" || args[2] != "value" {
						t.Errorf("unexpected args: %v", args)
					}

					return func() tea.Msg {
						return command.SetSuccessMsg{}
					}
				},
			},
			expectedOutput: "ok\n> ",
			expectedLogChecks: func(t *testing.T, logs []observer.LoggedEntry) {
			},
		},
		{
			name:           "unknown command",
			input:          "unknowncmd\n",
			registry:       CommandRegistry{},
			expectedOutput: "> ",
			expectedLogChecks: func(t *testing.T, logs []observer.LoggedEntry) {
				errorFound := false
				for _, entry := range logs {
					if entry.Message == "Unknown command" && entry.Level == zapcore.ErrorLevel {
						errorFound = true
					}
				}
				if !errorFound {
					t.Errorf("expected error log 'Unknown command', not found")
				}
			},
		},
		{
			name:  "command execution error",
			input: "login error\n",
			registry: CommandRegistry{
				"login": func(ctx context.Context, args []string) tea.Cmd {
					return func() tea.Msg {
						return command.ErrorMsg{Err: errors.New("execution failed")}
					}
				},
			},
			expectedOutput: "> ",
			expectedLogChecks: func(t *testing.T, logs []observer.LoggedEntry) {
				errorFound := false
				for _, entry := range logs {
					if entry.Message == "Command execution error" && entry.Level == zapcore.ErrorLevel {
						errorFound = true
					}
				}
				if !errorFound {
					t.Errorf("expected error log 'Command execution error', not found")
				}
			},
		},
		{
			name:  "empty line",
			input: "\n",
			registry: CommandRegistry{
				"login": func(ctx context.Context, args []string) tea.Cmd {
					return func() tea.Msg {
						return command.LoginSuccessMsg{}
					}
				},
			},
			expectedOutput: "> ",
			expectedLogChecks: func(t *testing.T, logs []observer.LoggedEntry) {
			},
		},
		{
			name:  "unknown message type",
			input: "custom\n",
			registry: CommandRegistry{
				"custom": func(ctx context.Context, args []string) tea.Cmd {
					return func() tea.Msg {
						return "custom result"
					}
				},
			},
			expectedOutput: "custom result\n> ",
			expectedLogChecks: func(t *testing.T, logs []observer.LoggedEntry) {
			},
		},
		{
			name:  "nil message",
			input: "noop\n",
			registry: CommandRegistry{
				"noop": func(ctx context.Context, args []string) tea.Cmd {
					return nil
				},
			},
			expectedOutput: "> ",
			expectedLogChecks: func(t *testing.T, logs []observer.LoggedEntry) {
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rIn, wIn, err := os.Pipe()
			if err != nil {
				t.Fatalf("pipe stdin: %v", err)
			}

			rOut, wOut, err := os.Pipe()
			if err != nil {
				t.Fatalf("pipe stdout: %v", err)
			}

			stdin := os.Stdin
			stdout := os.Stdout
			defer func() { os.Stdin = stdin }()
			defer func() { os.Stdout = stdout }()

			os.Stdin = rIn
			os.Stdout = wOut

			go func() {
				_, _ = wIn.Write([]byte(tt.input))
				_ = wIn.Close()
			}()

			core, obs := observer.New(zapcore.DebugLevel)
			logger.Logger = zap.New(core)
			defer func() { _ = logger.Logger.Sync() }()

			Run(context.Background(), tt.registry)

			_ = wOut.Close()

			var writer bytes.Buffer
			_, _ = writer.ReadFrom(rOut)

			actualOutput := writer.String()
			if !strings.HasSuffix(actualOutput, tt.expectedOutput) {
				t.Errorf("unexpected output, got: %q, want suffix: %q", actualOutput, tt.expectedOutput)
			}

			tt.expectedLogChecks(t, obs.All())
		})
	}
}
