package value

import (
	"encoding/json"
	"errors"
)

type TextValue struct {
	Text string `json:"text"`
}

func (v *TextValue) vType() vType { return typeText }

func (v *TextValue) ToBytes() ([]byte, error) {
	payload, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}
	return append([]byte{byte(typeText)}, payload...), nil
}

func (v *TextValue) Validate() error {
	if v.Text == "" {
		return errors.New("text is empty")
	}
	return nil
}

func (v *TextValue) String() string {
	return v.Text
}
