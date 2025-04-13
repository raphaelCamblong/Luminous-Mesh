package apiGateway

import (
	"sync"

	"github.com/luminous-mesh/control-plane/shared/interfaces"
)

var _ interfaces.ApiGateway = &apiGateway{}

type apiGateway struct {
	mu sync.Mutex
}

func (a *apiGateway) Start() error {
	a.mu.Lock()
	defer a.mu.Unlock()

	return nil
}

func (a *apiGateway) Stop() error {
	a.mu.Lock()
	defer a.mu.Unlock()

	return nil
}
