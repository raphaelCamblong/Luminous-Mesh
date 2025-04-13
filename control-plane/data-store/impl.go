package dataStore

import (
	"sync"

	"github.com/raphaelCamblong/Luminous-Mesh/control-plane/shared/interfaces"
)

const PluginSymbolName = "DataStore"

var _ interfaces.DataStore = &dataStore{}

type dataStore struct {
	mu sync.Mutex
}

func (d *dataStore) Start() error {
	d.mu.Lock()
	defer d.mu.Unlock()

	return nil
}

func (d *dataStore) Stop() error {
	d.mu.Lock()
	defer d.mu.Unlock()

	return nil
}
