package command

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/g123udini/gophkeeper/internal/client/model"
)

type mockDataUpserter struct {
	upsertFunc func(ctx context.Context, data *model.UserData) error
}

func (m *mockDataUpserter) Upsert(ctx context.Context, data *model.UserData) error {
	return m.upsertFunc(ctx, data)
}

func TestSet_Success(t *testing.T) {
	dataManager := &mockDataUpserter{
		upsertFunc: func(ctx context.Context, data *model.UserData) error {
			if data.DataKey != "test-key" {
				t.Fatalf("got key %q, want test-key", data.DataKey)
			}
			if data.DataValue == nil {
				t.Fatal("expected data value")
			}
			if data.UpdatedAt.IsZero() {
				t.Fatal("expected UpdatedAt")
			}
			if data.DeletedAt != time.Unix(0, 0) {
				t.Fatalf("unexpected DeletedAt: %v", data.DeletedAt)
			}

			return nil
		},
	}

	msg := Run(Set(
		context.Background(),
		dataManager,
		[]byte("1234567890abcdef"),
		[]string{"test-key", "text", "test-value"},
	))

	if _, ok := msg.(SetSuccessMsg); !ok {
		t.Fatalf("expected SetSuccessMsg, got %T", msg)
	}
}

func TestSet_MissingArgs(t *testing.T) {
	tests := [][]string{
		{},
		{"key"},
		{"key", "type"},
	}

	for _, args := range tests {
		msg := Run(Set(context.Background(), nil, nil, args))

		errMsg, ok := msg.(ErrorMsg)
		if !ok {
			t.Fatalf("expected ErrorMsg, got %T", msg)
		}

		if errMsg.Err == nil {
			t.Fatal("expected error")
		}
	}
}

func TestSet_InvalidType(t *testing.T) {
	msg := Run(Set(
		context.Background(),
		nil,
		[]byte("1234567890abcdef"),
		[]string{"key", "invalid-type", "value"},
	))

	errMsg, ok := msg.(ErrorMsg)
	if !ok {
		t.Fatalf("expected ErrorMsg, got %T", msg)
	}

	if errMsg.Err == nil {
		t.Fatal("expected error")
	}
}

func TestSet_UpsertError(t *testing.T) {
	dataManager := &mockDataUpserter{
		upsertFunc: func(ctx context.Context, data *model.UserData) error {
			return errors.New("upsert error")
		},
	}

	msg := Run(Set(
		context.Background(),
		dataManager,
		[]byte("1234567890abcdef"),
		[]string{"key", "text", "value"},
	))

	errMsg, ok := msg.(ErrorMsg)
	if !ok {
		t.Fatalf("expected ErrorMsg, got %T", msg)
	}

	if errMsg.Err == nil || errMsg.Err.Error() != "upsert error" {
		t.Fatalf("unexpected error: %v", errMsg.Err)
	}
}
