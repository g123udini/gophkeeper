package value

import (
	"encoding/base64"
	"errors"
	"fmt"
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
	const max = 16

	data := v.Data
	if len(data) > max {
		data = data[:max]
	}

	return fmt.Sprintf("Binary[%d]: %s...", len(v.Data), base64.StdEncoding.EncodeToString(data))
}
