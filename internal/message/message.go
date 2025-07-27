package message

type Message interface {
	Read(buf []byte) (Message, error)
	String() string
}
