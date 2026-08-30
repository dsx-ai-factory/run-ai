// Run:ai's anyworkload-controller binary: a thin consumer of the dromos library
// that wires the Run:AI / KAI conventions the dromos plugins do not own.
package main

import (
	"fmt"
	"os"
	"time"

	"github.com/rs/zerolog/log"
	rcpconfig "github.com/run-ai/runai/runai-common-packages/config"
	ctrl "sigs.k8s.io/controller-runtime"

	anyworkloadconfig "github.com/run-ai/runai/runai-cluster/cluster/anyworkload-controller/internal/config"
	"github.com/run-ai/runai/runai-cluster/cluster/anyworkload-controller/internal/mnnvlconv"
	"github.com/run-ai/runai/runai-cluster/cluster/anyworkload-controller/internal/ownership"
	"github.com/run-ai/runai/runai-cluster/cluster/anyworkload-controller/internal/restart"
	"github.com/run-ai/runai/runai-cluster/cluster/anyworkload-controller/plugins/networking"
	"github.com/run-ai/runai/runai-cluster/cluster/anyworkload-controller/plugins/personalcreds"
	"github.com/run-ai/runai/runai-cluster/cluster/anyworkload-controller/plugins/restartrules"
	"github.com/run-ai/runai/runai-cluster/cluster/dromos/pkg/app"
	dromosconfig "github.com/run-ai/runai/runai-cluster/cluster/dromos/pkg/config"
	"github.com/run-ai/runai/runai-cluster/cluster/dromos/plugins/mnnvl"
	"github.com/run-ai/runai/runai-cluster/cluster/dromos/plugins/nodeaffinity"
	"github.com/run-ai/runai/runai-cluster/cluster/dromos/plugins/storage"
	scheme_init "github.com/run-ai/runai/runai-cluster/cluster/go-operator/pkg/scheme-init"

	_ "github.com/run-ai/runai/runai-common-packages/fips"
)

func main() {
	if err := run(); err != nil {
		log.Error().Err(err).Msg("anyworkload-controller exited with error")
		os.Exit(1)
	}
}

func run() error {
	if err := rcpconfig.Register[*dromosconfig.DromosConfig](); err != nil {
		return fmt.Errorf("register dromos config: %w", err)
	}
	if err := rcpconfig.Register[*anyworkloadconfig.Config](); err != nil {
		return fmt.Errorf("register anyworkload-controller config: %w", err)
	}
	cfg := rcpconfig.Get[*anyworkloadconfig.Config]()

	options := []app.Option{
		app.WithScheme(scheme_init.Scheme()),
		app.WithConfig(rcpconfig.Get[*dromosconfig.DromosConfig]()),
		app.WithPlugins(networking.New, personalcreds.New, storage.New, nodeaffinity.New, mnnvl.NewBuilder(
			mnnvl.WithPodAffinityProvider(mnnvlconv.ProvidePodAffinity),
			mnnvl.WithWorkloadLabelStampDisabled(),
		)),
		app.WithManagerSetup(ownership.NewSetup()),
	}
	if cfg.RestartEnabled {
		log.Info().Msg("restart reconcilers enabled")
		options = append(options,
			app.WithManagerSetup(restart.NewSetup(
				cfg.RestartMaxConcurrentReconciles,
				time.Duration(cfg.RestartFrameworkDeletionTimeoutSec)*time.Second,
			)),
			app.WithPlugins(restartrules.NewBuilder(cfg.RestartAllowedServiceAccounts)),
		)
	}
	return app.RunController(ctrl.SetupSignalHandler(), options...)
}
