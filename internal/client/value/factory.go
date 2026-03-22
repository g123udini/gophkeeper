package value

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"os"
)

func FromBytes(data []byte) (Value, error) {
	if len(data) < 2 {
		return nil, errors.New("empty value data")
	}

	typ, err := newValueTypeFromByte(data[0])
	if err != nil {
		return nil, err
	}
	payload := data[1:]

	switch typ {
	case typeLoginPassword:
		var v LoginPassword
		if err := json.Unmarshal(payload, &v); err != nil {
			return nil, fmt.Errorf("invalid login/password: %w", err)
		}
		return &v, nil
	case typeText:
		var v TextValue
		if err := json.Unmarshal(payload, &v); err != nil {
			return nil, fmt.Errorf("invalid text: %w", err)
		}
		return &v, nil
	case typeBinary:
		data, err := base64.StdEncoding.DecodeString(string(payload))
		if err != nil {
			return nil, fmt.Errorf("invalid binary: %w", err)
		}

		return &BinaryValue{Data: data}, nil
	case typeCard:
		var v CardValue
		if err := json.Unmarshal(payload, &v); err != nil {
			return nil, fmt.Errorf("invalid card: %w", err)
		}
		return &v, nil
	default:
		return nil, errors.New("unknown value type")
	}
}

func FromUserInput(typeString string, data []string) (Value, error) {
	typ, err := newValueTypeFromString(typeString)
	if err != nil {
		return nil, err
	}

	switch typ {
	case typeLoginPassword:
		if len(data) < 2 {
			return nil, errors.New("login_password requires 2 arguments: login, password")
		}
		v := &LoginPassword{
			Login:    data[0],
			Password: data[1],
		}
		if err := v.Validate(); err != nil {
			return nil, err
		}
		return v, nil
	case typeText:
		if len(data) < 1 {
			return nil, errors.New("text requires 1 argument: text")
		}
		v := &TextValue{Text: data[0]}
		if err := v.Validate(); err != nil {
			return nil, err
		}
		return v, nil
	case typeBinary:
		if len(data) < 1 {
			return nil, errors.New("binary requires 1 argument: file path")
		}
		fileData, err := os.ReadFile(data[0])
		if err != nil {
			return nil, fmt.Errorf("cannot read binary file: %w", err)
		}
		v := &BinaryValue{Data: fileData}
		if err := v.Validate(); err != nil {
			return nil, err
		}
		return v, nil
	case typeCard:
		if len(data) < 5 {
			return nil, errors.New("card requires 5 arguments: number, holder, expireMonth, expireYear, cvc")
		}
		expireMonth, err := parseInt(data[2])
		if err != nil {
			return nil, fmt.Errorf("expireMonth should be int: %w", err)
		}
		expireYear, err := parseInt(data[3])
		if err != nil {
			return nil, fmt.Errorf("expireYear should be int: %w", err)
		}
		v := &CardValue{
			Number:      data[0],
			Holder:      data[1],
			ExpireMonth: expireMonth,
			ExpireYear:  expireYear,
			CVC:         data[4],
		}
		if err := v.Validate(); err != nil {
			return nil, err
		}
		return v, nil
	default:
		return nil, fmt.Errorf("unknown value type: %s", typeString)
	}
}

func parseInt(str string) (int, error) {
	var val int
	_, err := fmt.Sscan(str, &val)
	return val, err
}
