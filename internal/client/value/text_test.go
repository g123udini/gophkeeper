package value

import (
	"bytes"
	"encoding/json"
	"errors"
	"testing"
)

func TestTextValue_vType(t *testing.T) {
	v := TextValue{Text: "test text"}
	if v.vType() != typeText {
		t.Errorf("vType() = %v, want %v", v.vType(), typeText)
	}
}

func TestTextValue_ToBytes(t *testing.T) {
	tests := []struct {
		name    string
		text    string
		want    byte
		wantErr bool
	}{
		{
			name:    "valid text",
			text:    "test text",
			want:    byte(typeText),
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := &TextValue{Text: tt.text}
			got, err := v.ToBytes()
			if (err != nil) != tt.wantErr {
				t.Errorf("ToBytes() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if len(got) == 0 || got[0] != tt.want {
				t.Errorf("ToBytes() got = %v, want type byte %v", got, tt.want)
			}

			expectedJSON, _ := json.Marshal(v)
			if !bytes.Equal(got[1:], expectedJSON) {
				t.Errorf("ToBytes() data mismatch, got = %v, want %v", got[1:], expectedJSON)
			}
		})
	}
}

func TestTextValue_Validate(t *testing.T) {
	tests := []struct {
		name    string
		text    string
		wantErr error
	}{
		{
			name:    "valid text",
			text:    "test text",
			wantErr: nil,
		},
		{
			name:    "empty text",
			text:    "",
			wantErr: errors.New("text is empty"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := &TextValue{Text: tt.text}
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

func TestTextValue_String(t *testing.T) {
	testText := "test text"
	v := &TextValue{Text: testText}
	if v.String() != testText {
		t.Errorf("String() = %v, want %v", v.String(), testText)
	}
}
