package main

import (
	"github.com/raphaelCamblong/Luminous-Mesh/control-plane/shared/interfaces"
)

var _ interfaces.ApiGateway = &apiGateway{}

type apiGateway struct {
}

func (a *apiGateway) Start() error {
	return nil
}

func (a *apiGateway) Stop() error {
	return nil
}

func (a *apiGateway) GetName() string {
	return "api-gateway"
}

func (a *apiGateway) GetVersion() string {
	return "0.0.1"
}
