package plugins

import (
	"fmt"
	"reflect"
)

type Loader[T any] func(string) (T, error)

type Definition struct {
	Name        string
	Loader      any
	Destination any
	Required    bool
}

type DefinitionMap map[string]Definition

func NewDefinitionMap(defs []Definition) DefinitionMap {
	pluginMap := make(map[string]Definition)
	for _, def := range defs {
		pluginMap[def.Name] = def
	}
	return pluginMap
}

func (def *Definition) LoadPlugin(filePath string) error {
	loaderType := reflect.TypeOf(def.Loader)
	if loaderType.Kind() != reflect.Func {
		return fmt.Errorf("loader must be a function for plugin %s", def.Name)
	}

	typeArg := loaderType.Out(0)
	destValue := reflect.ValueOf(def.Destination)
	if destValue.Kind() != reflect.Ptr {
		return fmt.Errorf("destination must be a pointer for plugin %s", def.Name)
	}

	destElem := destValue.Elem()
	if !typeArg.AssignableTo(destElem.Type()) {
		return fmt.Errorf("loader return type %v not assignable to destination type %v for plugin %s",
			typeArg, destElem.Type(), def.Name)
	}

	loaderValue := reflect.ValueOf(def.Loader)
	results := loaderValue.Call([]reflect.Value{reflect.ValueOf(filePath)})

	if !results[1].IsNil() {
		return results[1].Interface().(error)
	}

	destElem.Set(results[0])
	return nil
}
