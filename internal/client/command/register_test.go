package command

import (
	"context"
	"errors"
	"testing"
)

type mockRegistrarClient struct {
	registerFunc func(ctx context.Context, login, password string, masterPassword []byte) (string, error)
}

func (m *mockRegistrarClient) Register(
	ctx context.Context,
	login, password string,
	masterPassword []byte,
) (string, error) {
	return m.registerFunc(ctx, login, password, masterPassword)
}

func TestRegister_Success(t *testing.T) {
	client := &mockRegistrarClient{
		registerFunc: func(ctx context.Context, login, password string, masterPassword []byte) (string, error) {
			return "token", nil
		},
	}

	msg := Run(Register(context.Background(), client, []byte("1234567890abcdef"), []string{"user", "pass"}))

	if _, ok := msg.(RegisterSuccessMsg); !ok {
		t.Fatalf("expected RegisterSuccessMsg, got %T", msg)
	}
}

func TestRegister_MissingArgs(t *testing.T) {
	tests := [][]string{
		{},
		{"user"},
	}

	for _, args := range tests {
		msg := Run(Register(context.Background(), nil, nil, args))

		errMsg, ok := msg.(ErrorMsg)
		if !ok {
			t.Fatalf("expected ErrorMsg, got %T", msg)
		}

		if errMsg.Err == nil {
			t.Fatal("expected error")
		}
	}
}

func TestRegister_Error(t *testing.T) {
	client := &mockRegistrarClient{
		registerFunc: func(ctx context.Context, login, password string, masterPassword []byte) (string, error) {
			return "", errors.New("user already exists")
		},
	}

	msg := Run(Register(context.Background(), client, []byte("1234567890abcdef"), []string{"user", "pass"}))

	errMsg, ok := msg.(ErrorMsg)
	if !ok {
		t.Fatalf("expected ErrorMsg, got %T", msg)
	}

	if errMsg.Err == nil || errMsg.Err.Error() != "user already exists" {
		t.Fatalf("unexpected error: %v", errMsg.Err)
	}
}
