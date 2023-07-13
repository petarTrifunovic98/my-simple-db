package inputprovider

type IInputProvider interface {
	GetInput() (string, error)
}
