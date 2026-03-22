package value

import (
	"testing"
)

func TestNewValueTypeFromByte(t *testing.T) {
	tests := []struct {
		name    string
		b       byte
		want    vType
		wantErr bool
	}{
		{
			name:    "login_password type",
			b:       byte(typeLoginPassword),
			want:    typeLoginPassword,
			wantErr: false,
		},
		{
			name:    "text type",
			b:       byte(typeText),
			want:    typeText,
			wantErr: false,
		},
		{
			name:    "binary type",
			b:       byte(typeBinary),
			want:    typeBinary,
			wantErr: false,
		},
		{
			name:    "card type",
			b:       byte(typeCard),
			want:    typeCard,
			wantErr: false,
		},
		{
			name:    "invalid type - zero",
			b:       0,
			want:    0,
			wantErr: true,
		},
		{
			name:    "invalid type - too high",
			b:       5,
			want:    0,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := newValueTypeFromByte(tt.b)
			if (err != nil) != tt.wantErr {
				t.Errorf("newValueTypeFromByte() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("newValueTypeFromByte() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNewValueTypeFromString(t *testing.T) {
	tests := []struct {
		name    string
		s       string
		want    vType
		wantErr bool
	}{
		{
			name:    "login_password type",
			s:       "login_password",
			want:    typeLoginPassword,
			wantErr: false,
		},
		{
			name:    "text type",
			s:       "text",
			want:    typeText,
			wantErr: false,
		},
		{
			name:    "binary type",
			s:       "binary",
			want:    typeBinary,
			wantErr: false,
		},
		{
			name:    "card type",
			s:       "card",
			want:    typeCard,
			wantErr: false,
		},
		{
			name:    "invalid type",
			s:       "unknown_type",
			want:    0,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := newValueTypeFromString(tt.s)
			if (err != nil) != tt.wantErr {
				t.Errorf("newValueTypeFromString() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("newValueTypeFromString() = %v, want %v", got, tt.want)
			}
		})
	}
}
