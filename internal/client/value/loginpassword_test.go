package value

import (
	"bytes"
	"encoding/json"
	"errors"
	"testing"
)

func TestLoginPassword_vType(t *testing.T) {
	v := LoginPassword{Login: "testuser", Password: "testpass"}
	if v.vType() != typeLoginPassword {
		t.Errorf("vType() = %v, want %v", v.vType(), typeLoginPassword)
	}
}

func TestLoginPassword_ToBytes(t *testing.T) {
	lp := LoginPassword{
		Login:    "testuser",
		Password: "testpass",
	}

	got, err := lp.ToBytes()
	if err != nil {
		t.Errorf("ToBytes() error = %v", err)
		return
	}

	if len(got) == 0 || got[0] != byte(typeLoginPassword) {
		t.Errorf("ToBytes() got = %v, want type byte %v", got, byte(typeLoginPassword))
	}

	expectedJSON, _ := json.Marshal(lp)
	if !bytes.Equal(got[1:], expectedJSON) {
		t.Errorf("ToBytes() data mismatch, got = %v, want %v", got[1:], expectedJSON)
	}
}

func TestLoginPassword_Validate(t *testing.T) {
	tests := []struct {
		name     string
		login    string
		password string
		wantErr  error
	}{
		{
			name:     "valid login and password",
			login:    "testuser",
			password: "testpass",
			wantErr:  nil,
		},
		{
			name:     "empty login",
			login:    "",
			password: "testpass",
			wantErr:  errors.New("login is empty"),
		},
		{
			name:     "empty password",
			login:    "testuser",
			password: "",
			wantErr:  errors.New("password is empty"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := &LoginPassword{
				Login:    tt.login,
				Password: tt.password,
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

func TestLoginPassword_String(t *testing.T) {
	v := &LoginPassword{
		Login:    "testuser",
		Password: "testpass",
	}
	expected := "Login: testuser, Password: testpass"
	if v.String() != expected {
		t.Errorf("String() = %v, want %v", v.String(), expected)
	}
}
