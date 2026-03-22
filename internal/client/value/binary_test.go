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
	if v.String() != string(testData) {
		t.Errorf("String() = %v, want %v", v.String(), string(testData))
	}
}
