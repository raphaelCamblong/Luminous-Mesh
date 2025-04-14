package interfaces

type ApiGateway interface {
	Plugin
	Start() error
	Stop() error
}
