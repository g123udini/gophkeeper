package value

import (
	"bytes"
	"errors"
	"testing"
)

func TestBinaryValue_vType(t *testing.T) {
	v := BinaryValue{Data: []byte("test data")}
	if v.vType() != typeBinary {
		t.Errorf("vType() = %v, want %v", v.vType(), typeBinary)
	}
}

func TestBinaryValue_ToBytes(t *testing.T) {
	tests := []struct {
		name    string
		data    []byte
		want    byte
		wantErr bool
	}{
		{
			name:    "valid binary data",
			data:    []byte("test data"),
			want:    byte(typeBinary),
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := &BinaryValue{Data: tt.data}
			got, err := v.ToBytes()
			if (err != nil) != tt.wantErr {
				t.Errorf("ToBytes() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if len(got) == 0 || got[0] != tt.want {
				t.Errorf("ToBytes() got = %v, want type byte %v", got, tt.want)
			}
			if !bytes.Equal(got[1:], tt.data) {
				t.Errorf("ToBytes() data mismatch, got = %v, want %v", got[1:], tt.data)
			}
		})
	}
}

func TestBinaryValue_Validate(t *testing.T) {
	tests := []struct {
		name    string
		data    []byte
		wantErr error
	}{
		{
			name:    "valid data",
			data:    []byte("test data"),
			wantErr: nil,
		},
		{
			name:    "empty data",
			data:    []byte{},
			wantErr: errors.New("data is empty"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := &BinaryValue{Data: tt.data}
			err := v.Validate()
			if (err == nil && tt.wantErr != nil) || (err != nil && tt.wantErr == nil) {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err != nil && tt.wantErr != nil && err.Error() != tt.wantErr.Error() {
				t.Errorf("Validate() error message = %v, wantErr message %v", err.Error(), tt.wantErr.Error())
			}
		})
	}
}

func TestBinaryValue_String(t *testing.T) {
	testData := []byte("test data")
	v := &BinaryValue{Data: testData}

	result := v.String()
	expected := "Binary[9]: dGVzdCBkYXRh..."

	if result != expected {
		t.Errorf("String() = %v, want %v", result, expected)
	}
}

func TestBinaryValue_ToBytes_FromBytes(t *testing.T) {
	original := &BinaryValue{
		Data: []byte{0x00, 0x01, 0x02, 0xFF, 0x10},
	}

	raw, err := original.ToBytes()
	if err != nil {
		t.Fatalf("ToBytes() error = %v", err)
	}

	got, err := FromBytes(raw)
	if err != nil {
		t.Fatalf("FromBytes() error = %v", err)
	}

	bin, ok := got.(*BinaryValue)
	if !ok {
		t.Fatalf("expected *BinaryValue, got %T", got)
	}

	if !bytes.Equal(bin.Data, original.Data) {
		t.Errorf("binary data mismatch: got %v, want %v", bin.Data, original.Data)
	}
}
