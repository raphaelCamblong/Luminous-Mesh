package process

import (
	"fmt"

	"github.com/raphaelCamblong/Luminous-Mesh/control-plane/core/internal/infra"
)

type Process struct {
	infra *infra.Infra
}

func NewProcess() *Process {
	return &Process{
		infra: infra.NewInfra(),
	}
}

func (p *Process) Launch() {
	p.infra.LoadPlugins()
	p.infra.IntegrityCheck()

	fmt.Println(p.infra.Plugins.ApiGateway.GetName())
	fmt.Println(p.infra.Plugins.ApiGateway.GetVersion())
	fmt.Println(p.infra.Plugins.ApiGateway.Start())
	fmt.Println(p.infra.Plugins.DataStore.GetName())
	fmt.Println(p.infra.Plugins.DataStore.GetVersion())
	fmt.Println(p.infra.Plugins.DataStore.Start())
}
