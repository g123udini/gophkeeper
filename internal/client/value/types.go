package value

import "fmt"

type vType byte

const (
	typeLoginPassword vType = iota + 1
	typeText
	typeBinary
	typeCard
)

func newValueTypeFromByte(b byte) (vType, error) {
	if b < 1 || b > 4 {
		return 0, fmt.Errorf("invalid value type: %d", b)
	}

	return vType(b), nil
}

func newValueTypeFromString(s string) (vType, error) {
	switch s {
	case "login_password":
		return typeLoginPassword, nil
	case "text":
		return typeText, nil
	case "binary":
		return typeBinary, nil
	case "card":
		return typeCard, nil
	default:
		return 0, fmt.Errorf("invalid value type: %s", s)
	}
}
