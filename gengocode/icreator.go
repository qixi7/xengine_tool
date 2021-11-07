package main

// struct creator interface
type _creator interface {
	isTypeCreator()
}

// protocol buf creator interface
type _pbCreator interface {
	Reset()
	String() string
	ProtoMessage()
	Size() int
	MarshalTo([]byte) (int, error)
	Unmarshal([]byte) error
}
