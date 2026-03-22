package value

import (
	"bytes"
	"encoding/json"
	"errors"
	"testing"
)

func TestCardValue_vType(t *testing.T) {
	v := CardValue{
		Number:      "4111111111111111",
		Holder:      "Test User",
		ExpireMonth: 12,
		ExpireYear:  2030,
		CVC:         "123",
	}
	if v.vType() != typeCard {
		t.Errorf("vType() = %v, want %v", v.vType(), typeCard)
	}
}

func TestCardValue_ToBytes(t *testing.T) {
	card := CardValue{
		Number:      "4111111111111111",
		Holder:      "Test User",
		ExpireMonth: 12,
		ExpireYear:  2030,
		CVC:         "123",
	}

	got, err := card.ToBytes()
	if err != nil {
		t.Errorf("ToBytes() error = %v", err)
		return
	}

	if len(got) == 0 || got[0] != byte(typeCard) {
		t.Errorf("ToBytes() got = %v, want type byte %v", got, byte(typeCard))
	}

	expectedJSON, _ := json.Marshal(card)
	if !bytes.Equal(got[1:], expectedJSON) {
		t.Errorf("ToBytes() data mismatch, got = %v, want %v", got[1:], expectedJSON)
	}
}

func TestCardValue_Validate(t *testing.T) {
	tests := []struct {
		name        string
		number      string
		holder      string
		expireMonth int
		expireYear  int
		cvc         string
		wantErr     error
	}{
		{
			name:        "valid card",
			number:      "4111111111111111",
			holder:      "Test User",
			expireMonth: 12,
			expireYear:  2030,
			cvc:         "123",
			wantErr:     nil,
		},
		{
			name:        "invalid card number - too short",
			number:      "41111",
			holder:      "Test User",
			expireMonth: 12,
			expireYear:  2030,
			cvc:         "123",
			wantErr:     errors.New("invalid card number length"),
		},
		{
			name:        "invalid card number - too long",
			number:      "41111111111111111111111",
			holder:      "Test User",
			expireMonth: 12,
			expireYear:  2030,
			cvc:         "123",
			wantErr:     errors.New("invalid card number length"),
		},
		{
			name:        "empty holder",
			number:      "4111111111111111",
			holder:      "",
			expireMonth: 12,
			expireYear:  2030,
			cvc:         "123",
			wantErr:     errors.New("card holder is empty"),
		},
		{
			name:        "invalid expire month - too low",
			number:      "4111111111111111",
			holder:      "Test User",
			expireMonth: 0,
			expireYear:  2030,
			cvc:         "123",
			wantErr:     errors.New("invalid expire month"),
		},
		{
			name:        "invalid expire month - too high",
			number:      "4111111111111111",
			holder:      "Test User",
			expireMonth: 13,
			expireYear:  2030,
			cvc:         "123",
			wantErr:     errors.New("invalid expire month"),
		},
		{
			name:        "invalid expire year - too low",
			number:      "4111111111111111",
			holder:      "Test User",
			expireMonth: 12,
			expireYear:  1999,
			cvc:         "123",
			wantErr:     errors.New("invalid expire year"),
		},
		{
			name:        "invalid expire year - too high",
			number:      "4111111111111111",
			holder:      "Test User",
			expireMonth: 12,
			expireYear:  2101,
			cvc:         "123",
			wantErr:     errors.New("invalid expire year"),
		},
		{
			name:        "invalid CVC - too short",
			number:      "4111111111111111",
			holder:      "Test User",
			expireMonth: 12,
			expireYear:  2030,
			cvc:         "12",
			wantErr:     errors.New("invalid CVC"),
		},
		{
			name:        "invalid CVC - too long",
			number:      "4111111111111111",
			holder:      "Test User",
			expireMonth: 12,
			expireYear:  2030,
			cvc:         "12345",
			wantErr:     errors.New("invalid CVC"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := &CardValue{
				Number:      tt.number,
				Holder:      tt.holder,
				ExpireMonth: tt.expireMonth,
				ExpireYear:  tt.expireYear,
				CVC:         tt.cvc,
			}
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

func TestCardValue_String(t *testing.T) {
	v := &CardValue{
		Number:      "4111111111111111",
		Holder:      "Test User",
		ExpireMonth: 12,
		ExpireYear:  2030,
		CVC:         "123",
	}
	expected := "Card: 4111111111111111, Test User, 12/2030, 123"
	if v.String() != expected {
		t.Errorf("String() = %v, want %v", v.String(), expected)
	}
}
