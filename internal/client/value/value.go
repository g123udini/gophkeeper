package value

type Value interface {
	ToBytes() ([]byte, error)
	Validate() error
	vType() vType
	String() string
}
