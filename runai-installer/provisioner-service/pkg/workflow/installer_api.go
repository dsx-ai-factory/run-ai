package workflow

import (
	"github.com/run-ai/runai/runai-installer/installer/pkg/api"
)

type InstallerInterface interface {
	RunInstallationWithEventListner() (<-chan *api.FlowStatusReportingMsg, error)
}

type InstallerAPIInterface interface {
	Init(config *api.Config) error
	Get() InstallerInterface
}

type InstallerAPI struct{}

func (r *InstallerAPI) Init(config *api.Config) error {
	return api.Init(config)
}

func (r *InstallerAPI) Get() InstallerInterface {
	return &InstallerWrapper{installer: api.Get()}
}

type InstallerWrapper struct {
	installer *api.Installer
}

func (w *InstallerWrapper) RunInstallationWithEventListner() (<-chan *api.FlowStatusReportingMsg, error) {
	return w.installer.RunInstallationWithEventListner()
}
