package interfaces

type DataStore interface {
	Start() error
	Stop() error
}
