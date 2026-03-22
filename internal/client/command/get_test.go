package command

import (
	"context"
	"errors"
	"testing"

	"github.com/g123udini/gophkeeper/internal/client/aes"
	"github.com/g123udini/gophkeeper/internal/client/model"
	"github.com/g123udini/gophkeeper/internal/client/value"
)

type mockDataManager struct {
	getFunc func(ctx context.Context, key string) (*model.UserData, error)
}

func (m *mockDataManager) Get(ctx context.Context, key string) (*model.UserData, error) {
	return m.getFunc(ctx, key)
}

func TestGet_Success(t *testing.T) {
	password := []byte("1234567890abcdef")
	want := "some-value"

	val, err := value.FromUserInput("text", []string{want})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	bytes, err := val.ToBytes()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	cipherBytes, err := aes.Encrypt(password, bytes)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	dataManager := &mockDataManager{
		getFunc: func(ctx context.Context, key string) (*model.UserData, error) {
			return &model.UserData{
				DataKey:   key,
				DataValue: cipherBytes,
			}, nil
		},
	}

	msg := Run(Get(context.Background(), dataManager, password, []string{"some-key"}))

	got, ok := msg.(GetSuccessMsg)
	if !ok {
		t.Fatalf("expected GetSuccessMsg, got %T", msg)
	}

	if got.Value != want {
		t.Fatalf("got %q, want %q", got.Value, want)
	}
}

func TestGet_MissingArgs(t *testing.T) {
	msg := Run(Get(context.Background(), nil, nil, []string{}))

	errMsg, ok := msg.(ErrorMsg)
	if !ok {
		t.Fatalf("expected ErrorMsg, got %T", msg)
	}

	if errMsg.Err == nil {
		t.Fatal("expected error")
	}
}

func TestGet_GetError(t *testing.T) {
	dataManager := &mockDataManager{
		getFunc: func(ctx context.Context, key string) (*model.UserData, error) {
			return nil, errors.New("not found")
		},
	}

	msg := Run(Get(context.Background(), dataManager, []byte("1234567890abcdef"), []string{"some-key"}))

	errMsg, ok := msg.(ErrorMsg)
	if !ok {
		t.Fatalf("expected ErrorMsg, got %T", msg)
	}

	if errMsg.Err == nil || errMsg.Err.Error() != "not found" {
		t.Fatalf("unexpected error: %v", errMsg.Err)
	}
}

func TestGet_DecryptError(t *testing.T) {
	dataManager := &mockDataManager{
		getFunc: func(ctx context.Context, key string) (*model.UserData, error) {
			return &model.UserData{
				DataKey:   key,
				DataValue: []byte("corrupted-cipher"),
			}, nil
		},
	}

	msg := Run(Get(context.Background(), dataManager, []byte("1234567890abcdef"), []string{"test"}))

	errMsg, ok := msg.(ErrorMsg)
	if !ok {
		t.Fatalf("expected ErrorMsg, got %T", msg)
	}

	if errMsg.Err == nil {
		t.Fatal("expected error")
	}
}
