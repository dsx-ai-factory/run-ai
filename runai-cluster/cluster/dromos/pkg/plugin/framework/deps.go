package framework

import (
	"github.com/go-logr/logr"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type PluginDeps struct {
	Client client.Reader
	Logger logr.Logger
}
