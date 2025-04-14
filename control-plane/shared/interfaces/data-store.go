package interfaces

type DataStore interface {
	Plugin
	Start() error
	Stop() error
}
