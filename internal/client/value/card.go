package value

import (
	"encoding/json"
	"errors"
	"fmt"
)

type CardValue struct {
	Number      string `json:"number"`
	Holder      string `json:"holder"`
	ExpireMonth int    `json:"expire_month"`
	ExpireYear  int    `json:"expire_year"`
	CVC         string `json:"cvc"`
}

func (v *CardValue) vType() vType { return typeCard }

func (v *CardValue) ToBytes() ([]byte, error) {
	payload, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}
	return append([]byte{byte(typeCard)}, payload...), nil
}

func (v *CardValue) Validate() error {
	if len(v.Number) < 13 || len(v.Number) > 19 {
		return errors.New("invalid card number length")
	}
	if v.Holder == "" {
		return errors.New("card holder is empty")
	}
	if v.ExpireMonth < 1 || v.ExpireMonth > 12 {
		return errors.New("invalid expire month")
	}
	if v.ExpireYear < 2000 || v.ExpireYear > 2100 {
		return errors.New("invalid expire year")
	}
	if len(v.CVC) < 3 || len(v.CVC) > 4 {
		return errors.New("invalid CVC")
	}
	return nil
}

func (v *CardValue) String() string {
	return fmt.Sprintf("Card: %s, %s, %d/%d, %s", v.Number, v.Holder, v.ExpireMonth, v.ExpireYear, v.CVC)
}
