package main

import (
	"fmt"

	"github.com/raphaelCamblong/Luminous-Mesh/control-plane/shared/interfaces"
)

const PluginSymbolName = "DataStore"

var _ interfaces.DataStore = &dataStore{}

type dataStore struct {
}

func (d *dataStore) GetName() string {
	return "data-store"
}

func (d *dataStore) GetVersion() string {
	return "0.0.1"
}

func (d *dataStore) Start() error {
	fmt.Println("Starting data-store")
	return nil
}

func (d *dataStore) Stop() error {
	fmt.Println("Stopping data-store")
	return nil
}
