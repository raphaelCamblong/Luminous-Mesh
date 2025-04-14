package main

import "github.com/raphaelCamblong/Luminous-Mesh/control-plane/shared/interfaces"

func New() interfaces.ApiGateway {
	return &apiGateway{}
}
