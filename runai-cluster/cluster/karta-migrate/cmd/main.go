// karta-migrate is a helm-hook binary that migrates legacy ResourceInterface
// CRs to Karta CRs, and (with --remove-rid) deletes the legacy RIs and CRD.
package main

import (
	"context"
	"flag"
	"os"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/run-ai/runai/runai-cluster/cluster/karta-migrate/migration"
	"github.com/run-ai/runai/runai-cluster/cluster/karta-migrate/migration/karta"
	"github.com/run-ai/runai/runai-cluster/cluster/karta-migrate/migration/ri"

	_ "github.com/run-ai/runai/runai-common-packages/fips"
)

func init() {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: time.RFC3339})
}

func main() {
	var removeRID bool
	flag.BoolVar(&removeRID, "remove-rid", false, "delete legacy RIs and the RI CRD after migration")
	flag.Parse()

	ctx := ctrl.SetupSignalHandler()

	cfg, err := ctrl.GetConfig()
	if err != nil {
		log.Error().Err(err).Msg("get kubeconfig")
		os.Exit(1)
	}

	k8sClient, err := migration.NewK8sClient(cfg)
	if err != nil {
		log.Error().Err(err).Msg("build k8s client")
		os.Exit(1)
	}

	if err := run(ctx, k8sClient, removeRID); err != nil {
		log.Error().Err(err).Msg("migration failed")
		os.Exit(1)
	}
	log.Info().Msg("migration complete")
}

// run is the helm-hook orchestrator. See README.md for the flow.
//
// removeRID=false (pre-upgrade): strict gate. Create / reconcile Kartas,
// then verify every Karta covers its RI. Non-zero exit aborts the helm
// upgrade.
//
// removeRID=true (post-upgrade): backfill any RI the agent created during
// the upgrade window, then strip legacy finalizers and delete the RI CRD.
// Does not verify status - EWI owns it by then.
func run(ctx context.Context, c client.Client, removeRID bool) error {
	installed, err := ri.CRDInstalled(ctx, c)
	if err != nil {
		return err
	}
	if !installed {
		log.Info().Msg("RI CRD not installed, nothing to migrate")
		return nil
	}

	if removeRID {
		if err := karta.Migrate(ctx, c); err != nil {
			return err
		}
		log.Info().Msg("--remove-rid set, removing legacy RI CRs and CRD")
		if err := ri.Cleanup(ctx, c); err != nil {
			return err
		}
		if err := ri.DeleteCRD(ctx, c); err != nil {
			return err
		}
		log.Info().Msg("legacy RI CRs and CRD removed")
		return nil
	}

	done, err := karta.IsMigrationComplete(ctx, c)
	if err != nil {
		return err
	}
	if done {
		log.Info().Msg("all RIs already have matching Ready Karta CRs, skipping migration phase")
		return nil
	}

	if err := karta.Migrate(ctx, c); err != nil {
		return err
	}
	return karta.VerifyHealth(ctx, c)
}
