package ioprovider

type IIOProvider interface {
	GetInput() (string, error)
	Print(data string)
}
