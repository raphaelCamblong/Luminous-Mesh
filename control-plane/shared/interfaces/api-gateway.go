package interfaces

type ApiGateway interface {
	Start() error
	Stop() error
}
