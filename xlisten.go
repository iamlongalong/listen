package listen

type Encoder interface {
	Encode() ([]byte, error)
}
