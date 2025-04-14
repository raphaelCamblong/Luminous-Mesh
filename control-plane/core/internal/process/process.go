package process

import (
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
	p.infra.LoadGrpcServer()
	p.infra.IntegrityCheck()

	p.infra.Server.Start(p.infra.Ctx)
}
