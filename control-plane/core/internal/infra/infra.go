package infra

import (
	"fmt"
	"path/filepath"

	"github.com/raphaelCamblong/Luminous-Mesh/control-plane/core/init/config"
	"github.com/raphaelCamblong/Luminous-Mesh/control-plane/core/pkg/plugins"
	"github.com/raphaelCamblong/Luminous-Mesh/control-plane/shared/interfaces"
)

type Infra struct {
	Plugins pluginRegistry
}

type pluginRegistry struct {
	ApiGateway interfaces.ApiGateway
	DataStore  interfaces.DataStore
}

func NewInfra() *Infra {
	return &Infra{}
}

func (i *Infra) IntegrityCheck() {
	if i.Plugins.ApiGateway == nil {
		panic("ApiGateway is not initialized")
	}
	if i.Plugins.DataStore == nil {
		panic("DataStore is not initialized")
	}
}

func (i *Infra) LoadPlugins() {
	cfg := config.Get()

	pluginDefs := []plugins.Definition{
		{
			Name:        "api-gateway",
			Loader:      plugins.LoadPlugin[interfaces.ApiGateway],
			Destination: &i.Plugins.ApiGateway,
			Required:    true,
		},
		{
			Name:        "data-store",
			Loader:      plugins.LoadPlugin[interfaces.DataStore],
			Destination: &i.Plugins.DataStore,
			Required:    true,
		},
	}
	pluginMap := plugins.NewDefinitionMap(pluginDefs)

	for _, pluginName := range cfg.Plugins.Load {
		def, exists := pluginMap[pluginName]
		if !exists {
			panic(fmt.Errorf("❌ Unknown plugin type: %s", pluginName))
		}

		filePath := filepath.Join(cfg.Plugins.Path, pluginName+".so")
		if err := def.LoadPlugin(filePath); err != nil {
			panic(fmt.Errorf("❌ Failed to load plugin %s: %w", pluginName, err))
		}
	}
}
