package karta

import (
	"context"
	"fmt"

	"github.com/rs/zerolog/log"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// Migrate converges each Karta with the RI it represents. Three buckets per RI:
//   - missing  -> create the Karta and seed its conditions
//   - stale    -> the Karta exists but does not cover the RI (e.g. an earlier
//     migration run set CRDExists=False before the CRD was installed). Just
//     recompute and patch its conditions; do not touch the spec.
//   - covered  -> nothing to do, skip
//
// Migrate is idempotent: a rerun with no drift is a no-op.
func Migrate(ctx context.Context, c client.Client) error {
	ris, err := listRIs(ctx, c)
	if err != nil {
		return err
	}

	var created, reconciled, covered int
	for i := range ris.Items {
		riObj := &ris.Items[i]
		name, err := generateKartaName(riObj)
		if err != nil {
			return fmt.Errorf("generate karta name for RI %s: %w", riObj.GetName(), err)
		}

		kartaObj, err := getKarta(ctx, c, name)
		if err != nil && !apierrors.IsNotFound(err) {
			return fmt.Errorf("get karta %s: %w", name, err)
		}

		if apierrors.IsNotFound(err) {
			if err := createKartaFromRI(ctx, c, name, riObj); err != nil {
				return fmt.Errorf("create karta %s: %w", name, err)
			}
			created++
			log.Info().Str("name", name).Msg("created karta")
			continue
		}

		if ok, _ := kartaCoversRI(kartaObj, riObj); ok {
			covered++
			log.Debug().Str("name", name).Msg("karta already covers RI, skipping")
			continue
		}

		if err := copyRIConditionsToKarta(ctx, c, name, riObj); err != nil {
			return fmt.Errorf("reconcile conditions for karta %s: %w", name, err)
		}
		reconciled++
		log.Info().Str("name", name).Msg("reconciled stale karta conditions")
	}

	log.Info().Int("ris", len(ris.Items)).Int("created", created).Int("reconciled_stale", reconciled).Int("covered", covered).Msg("migrate complete")
	return nil
}
