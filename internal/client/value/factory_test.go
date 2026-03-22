package value

import (
	"encoding/base64"
	"os"
	"testing"
)

func TestFromBytes(t *testing.T) {
	tests := []struct {
		name     string
		setup    func() []byte
		validate func(t *testing.T, v Value, err error)
	}{
		{
			name: "empty data",
			setup: func() []byte {
				return []byte{}
			},
			validate: func(t *testing.T, v Value, err error) {
				if err == nil {
					t.Errorf("Expected error for empty data, got nil")
				}
			},
		},
		{
			name: "invalid type",
			setup: func() []byte {
				return []byte{255, 1, 2, 3}
			},
			validate: func(t *testing.T, v Value, err error) {
				if err == nil {
					t.Errorf("Expected error for invalid type, got nil")
				}
			},
		},
		{
			name: "valid login password",
			setup: func() []byte {
				lp := LoginPassword{
					Login:    "testuser",
					Password: "testpass",
				}
				data, _ := lp.ToBytes()
				return data
			},
			validate: func(t *testing.T, v Value, err error) {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
					return
				}
				lp, ok := v.(*LoginPassword)
				if !ok {
					t.Errorf("Expected LoginPassword, got %T", v)
					return
				}
				if lp.Login != "testuser" || lp.Password != "testpass" {
					t.Errorf("Wrong data: %+v", lp)
				}
			},
		},
		{
			name: "valid text",
			setup: func() []byte {
				tv := TextValue{
					Text: "test text",
				}
				data, _ := tv.ToBytes()
				return data
			},
			validate: func(t *testing.T, v Value, err error) {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
					return
				}
				tv, ok := v.(*TextValue)
				if !ok {
					t.Errorf("Expected TextValue, got %T", v)
					return
				}
				if tv.Text != "test text" {
					t.Errorf("Wrong data: %+v", tv)
				}
			},
		},
		{
			name: "valid binary",
			setup: func() []byte {
				binData := []byte("test binary data")
				encoded := base64.StdEncoding.EncodeToString(binData)
				return append([]byte{byte(typeBinary)}, encoded...)
			},
			validate: func(t *testing.T, v Value, err error) {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
					return
				}
				bv, ok := v.(*BinaryValue)
				if !ok {
					t.Errorf("Expected BinaryValue, got %T", v)
					return
				}
				if string(bv.Data) != "test binary data" {
					t.Errorf("Wrong data: %s", string(bv.Data))
				}
			},
		},
		{
			name: "valid card",
			setup: func() []byte {
				cv := CardValue{
					Number:      "4111111111111111",
					Holder:      "Test User",
					ExpireMonth: 12,
					ExpireYear:  2030,
					CVC:         "123",
				}
				data, _ := cv.ToBytes()
				return data
			},
			validate: func(t *testing.T, v Value, err error) {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
					return
				}
				cv, ok := v.(*CardValue)
				if !ok {
					t.Errorf("Expected CardValue, got %T", v)
					return
				}
				if cv.Number != "4111111111111111" || cv.Holder != "Test User" ||
					cv.ExpireMonth != 12 || cv.ExpireYear != 2030 || cv.CVC != "123" {
					t.Errorf("Wrong data: %+v", cv)
				}
			},
		},
		{
			name: "invalid login password json",
			setup: func() []byte {
				return append([]byte{byte(typeLoginPassword)}, []byte("invalid json")...)
			},
			validate: func(t *testing.T, v Value, err error) {
				if err == nil {
					t.Errorf("Expected error for invalid json, got nil")
				}
			},
		},
		{
			name: "invalid text json",
			setup: func() []byte {
				return append([]byte{byte(typeText)}, []byte("invalid json")...)
			},
			validate: func(t *testing.T, v Value, err error) {
				if err == nil {
					t.Errorf("Expected error for invalid json, got nil")
				}
			},
		},
		{
			name: "invalid binary data",
			setup: func() []byte {
				return append([]byte{byte(typeBinary)}, []byte("invalid base64")...)
			},
			validate: func(t *testing.T, v Value, err error) {
				if err == nil {
					t.Errorf("Expected error for invalid base64, got nil")
				}
			},
		},
		{
			name: "invalid card json",
			setup: func() []byte {
				return append([]byte{byte(typeCard)}, []byte("invalid json")...)
			},
			validate: func(t *testing.T, v Value, err error) {
				if err == nil {
					t.Errorf("Expected error for invalid json, got nil")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data := tt.setup()
			v, err := FromBytes(data)
			tt.validate(t, v, err)
		})
	}
}

func TestFromUserInput(t *testing.T) {
	// Create a temporary file for binary tests
	binaryFile, err := os.CreateTemp("", "binary_test")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(binaryFile.Name())

	binaryData := []byte("test binary data")
	if _, err := binaryFile.Write(binaryData); err != nil {
		t.Fatalf("Failed to write to temp file: %v", err)
	}
	binaryFile.Close()

	tests := []struct {
		name     string
		typeStr  string
		data     []string
		validate func(t *testing.T, v Value, err error)
	}{
		{
			name:    "invalid type",
			typeStr: "unknown_type",
			data:    []string{},
			validate: func(t *testing.T, v Value, err error) {
				if err == nil {
					t.Errorf("Expected error for invalid type, got nil")
				}
			},
		},
		{
			name:    "login_password with missing data",
			typeStr: "login_password",
			data:    []string{"testuser"},
			validate: func(t *testing.T, v Value, err error) {
				if err == nil {
					t.Errorf("Expected error for missing data, got nil")
				}
			},
		},
		{
			name:    "valid login_password",
			typeStr: "login_password",
			data:    []string{"testuser", "testpass"},
			validate: func(t *testing.T, v Value, err error) {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
					return
				}
				lp, ok := v.(*LoginPassword)
				if !ok {
					t.Errorf("Expected LoginPassword, got %T", v)
					return
				}
				if lp.Login != "testuser" || lp.Password != "testpass" {
					t.Errorf("Wrong data: %+v", lp)
				}
			},
		},
		{
			name:    "text with missing data",
			typeStr: "text",
			data:    []string{},
			validate: func(t *testing.T, v Value, err error) {
				if err == nil {
					t.Errorf("Expected error for missing data, got nil")
				}
			},
		},
		{
			name:    "valid text",
			typeStr: "text",
			data:    []string{"test text"},
			validate: func(t *testing.T, v Value, err error) {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
					return
				}
				tv, ok := v.(*TextValue)
				if !ok {
					t.Errorf("Expected TextValue, got %T", v)
					return
				}
				if tv.Text != "test text" {
					t.Errorf("Wrong data: %+v", tv)
				}
			},
		},
		{
			name:    "binary with missing data",
			typeStr: "binary",
			data:    []string{},
			validate: func(t *testing.T, v Value, err error) {
				if err == nil {
					t.Errorf("Expected error for missing data, got nil")
				}
			},
		},
		{
			name:    "binary with invalid file",
			typeStr: "binary",
			data:    []string{"nonexistent_file.bin"},
			validate: func(t *testing.T, v Value, err error) {
				if err == nil {
					t.Errorf("Expected error for invalid file, got nil")
				}
			},
		},
		{
			name:    "valid binary",
			typeStr: "binary",
			data:    []string{binaryFile.Name()},
			validate: func(t *testing.T, v Value, err error) {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
					return
				}
				bv, ok := v.(*BinaryValue)
				if !ok {
					t.Errorf("Expected BinaryValue, got %T", v)
					return
				}
				if string(bv.Data) != string(binaryData) {
					t.Errorf("Wrong data: %s", string(bv.Data))
				}
			},
		},
		{
			name:    "card with missing data",
			typeStr: "card",
			data:    []string{"4111111111111111", "Test User", "12", "2030"},
			validate: func(t *testing.T, v Value, err error) {
				if err == nil {
					t.Errorf("Expected error for missing data, got nil")
				}
			},
		},
		{
			name:    "card with invalid expire month",
			typeStr: "card",
			data:    []string{"4111111111111111", "Test User", "invalid", "2030", "123"},
			validate: func(t *testing.T, v Value, err error) {
				if err == nil {
					t.Errorf("Expected error for invalid expire month, got nil")
				}
			},
		},
		{
			name:    "card with invalid expire year",
			typeStr: "card",
			data:    []string{"4111111111111111", "Test User", "12", "invalid", "123"},
			validate: func(t *testing.T, v Value, err error) {
				if err == nil {
					t.Errorf("Expected error for invalid expire year, got nil")
				}
			},
		},
		{
			name:    "valid card",
			typeStr: "card",
			data:    []string{"4111111111111111", "Test User", "12", "2030", "123"},
			validate: func(t *testing.T, v Value, err error) {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
					return
				}
				cv, ok := v.(*CardValue)
				if !ok {
					t.Errorf("Expected CardValue, got %T", v)
					return
				}
				if cv.Number != "4111111111111111" || cv.Holder != "Test User" ||
					cv.ExpireMonth != 12 || cv.ExpireYear != 2030 || cv.CVC != "123" {
					t.Errorf("Wrong data: %+v", cv)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v, err := FromUserInput(tt.typeStr, tt.data)
			tt.validate(t, v, err)
		})
	}
}

func TestParseInt(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    int
		wantErr bool
	}{
		{
			name:    "valid integer",
			input:   "42",
			want:    42,
			wantErr: false,
		},
		{
			name:    "negative integer",
			input:   "-10",
			want:    -10,
			wantErr: false,
		},
		{
			name:    "invalid integer",
			input:   "not a number",
			want:    0,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseInt(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseInt() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("parseInt() = %v, want %v", got, tt.want)
			}
		})
	}
}
