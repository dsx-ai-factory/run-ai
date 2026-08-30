package main

import (
	"context"
	"fmt"
	"log"

	"github.com/run-ai/runai/runai-cluster/cluster/helm-hooks-handler/internal/clients"
	"github.com/run-ai/runai/runai-cluster/cluster/helm-hooks-handler/internal/hooks/factory"
	"go.uber.org/zap/zapcore"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	// Import all Kubernetes client auth plugins (e.g. Azure, GCP, OIDC, etc.)
	// to ensure that exec-entrypoint and run can make use of them.

	_ "k8s.io/client-go/plugin/pkg/client/auth"

	ctrl "sigs.k8s.io/controller-runtime"

	"github.com/run-ai/runai/runai-cluster/cluster/helm-hooks-handler/internal/installation/util"

	_ "github.com/run-ai/runai/runai-common-packages/fips"
)

var (
	setupLog = ctrl.Log.WithName("setup")
)

func main() {
	ctrl.SetLogger(zap.New(zap.UseFlagOptions(&zap.Options{
		Level: zapcore.InfoLevel,
	})))
	opts := util.SetOptions()

	var k8sClients *clients.Clients
	var err error
	if opts.RequiresKubernetesAccess {
		k8sClients, err = clients.InitClients(opts)
		if err != nil {
			log.Fatalf("failed to init k8sClients needed to run hooks: %v", err)
		}
	}

	hook, err := factory.HookFactory(opts)
	if err != nil {
		log.Fatalf("Error creating hook: %v", err)
	}
	err = hook.Call(context.Background(), k8sClients)
	if err != nil {
		log.Fatalf("Failed to run %s tasks: %v", hook.GetIdentifier(), err)
	}

	setupLog.Info(fmt.Sprintf("Successfully ran %s tasks", hook.GetIdentifier()))
	return
}
