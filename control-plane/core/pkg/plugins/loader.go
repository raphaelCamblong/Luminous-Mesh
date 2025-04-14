package plugins

import (
	"fmt"
	"plugin"
)

// LoadPlugin loads a plugin from a specific .so file path.
// ! T must be an interface type. The plugin must export: `func New() T`
func LoadPlugin[T any](path string) (T, error) {
	var zero T

	p, err := plugin.Open(path)
	if err != nil {
		return zero, fmt.Errorf("failed to open plugin %s: %w", path, err)
	}

	sym, err := p.Lookup("New")
	if err != nil {
		return zero, fmt.Errorf("symbol 'New' not found in %s: %w", path, err)
	}

	newFunc, ok := sym.(func() T)
	if !ok {
		return zero, fmt.Errorf("symbol 'New' has wrong signature in %s", path)
	}

	instance := newFunc()
	return instance, nil
}
