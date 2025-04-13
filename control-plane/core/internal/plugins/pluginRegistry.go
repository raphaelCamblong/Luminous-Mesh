// plugin/manager.go
package plugin

import (
	"fmt"
	"os"
	"path/filepath"
	"plugin"

	"github.com/luminous-mesh/control-plane/shared/interfaces"
)

const PluginSymbolName = "PluginRegistry"

type Manager struct {
	processors map[string]interfaces.Processor
}

func NewManager() *Manager {
	return &Manager{
		processors: make(map[string]interfaces.Processor),
	}
}

func (m *Manager) LoadPlugins(pluginDir string) error {
	files, err := os.ReadDir(pluginDir)
	if err != nil {
		return err
	}

	for _, file := range files {
		if !file.IsDir() && filepath.Ext(file.Name()) == ".so" {
			if err := m.loadPlugin(filepath.Join(pluginDir, file.Name())); err != nil {
				fmt.Printf("Error loading plugin %s: %v\n", file.Name(), err)
				// Continue loading other plugins
			}
		}
	}
	return nil
}

func (m *Manager) loadPlugin(path string) error {
	p, err := plugin.Open(path)
	if err != nil {
		return err
	}

	symPlugin, err := p.Lookup(PluginSymbolName)
	if err != nil {
		return fmt.Errorf("plugin doesn't export %q symbol: %w", PluginSymbolName, err)
	}

	registry, ok := symPlugin.(PluginRegistry)
	if !ok {
		return fmt.Errorf("plugin symbol doesn't implement PluginRegistry")
	}

	for id, component := range registry.Register() {
		if processor, ok := component.(interfaces.Processor); ok {
			m.processors[id] = processor
			fmt.Printf("Registered processor: %s (version %s)\n",
				processor.GetName(), processor.GetVersion())
		}
	}

	return nil
}

func (m *Manager) GetProcessor(id string) (interfaces.Processor, bool) {
	p, exists := m.processors[id]
	return p, exists
}
