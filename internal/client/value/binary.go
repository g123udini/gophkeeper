package value

import (
	"errors"
)

type BinaryValue struct {
	Data []byte `json:"data"`
}

func (v *BinaryValue) vType() vType { return typeBinary }

func (v *BinaryValue) ToBytes() ([]byte, error) {
	return append([]byte{byte(typeBinary)}, v.Data...), nil
}

func (v *BinaryValue) Validate() error {
	if len(v.Data) == 0 {
		return errors.New("data is empty")
	}
	return nil
}

func (v *BinaryValue) String() string {
	return string(v.Data)
}
