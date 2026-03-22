package value

import (
	"encoding/json"
	"errors"
	"fmt"
)

type LoginPassword struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

func (v *LoginPassword) vType() vType {
	return typeLoginPassword
}

func (v *LoginPassword) ToBytes() ([]byte, error) {
	payload, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}
	return append([]byte{byte(typeLoginPassword)}, payload...), nil
}

func (v *LoginPassword) Validate() error {
	if v.Login == "" {
		return errors.New("login is empty")
	}
	if v.Password == "" {
		return errors.New("password is empty")
	}
	return nil
}

func (v *LoginPassword) String() string {
	return fmt.Sprintf("Login: %s, Password: %s", v.Login, v.Password)
}
