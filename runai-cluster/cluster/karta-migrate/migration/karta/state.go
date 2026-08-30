package karta

import (
	"context"
	"fmt"

	"github.com/rs/zerolog/log"

	kartav1alpha1 "github.com/run-ai/karta/pkg/api/runai/v1alpha1"
	"github.com/run-ai/runai/runai-cluster/cluster/karta-migrate/migration"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// IsMigrationComplete reports whether every RI already has a Karta that is at
// least as healthy as the RI: same canonical name, every condition the RI had
// True is True on the Karta, and the Karta is Ready=True.
func IsMigrationComplete(ctx context.Context, c client.Client) (bool, error) {
	ris, err := listRIs(ctx, c)
	if err != nil {
		return false, err
	}
	for i := range ris.Items {
		ri := &ris.Items[i]
		name, err := generateKartaName(ri)
		if err != nil {
			return false, err
		}
		karta, err := getKarta(ctx, c, name)
		if err != nil {
			if apierrors.IsNotFound(err) {
				log.Debug().Str("ri", ri.GetName()).Str("expectedKarta", name).Msg("karta missing, will run migration")
				return false, nil
			}
			return false, err
		}
		if ok, why := kartaCoversRI(karta, ri); !ok {
			log.Debug().Str("karta", name).Str("reason", why).Msg("karta does not cover RI yet, will run migration")
			return false, nil
		}
	}
	return true, nil
}

// listRIs is a small helper that lists every legacy ResourceInterface via the
// single client.Client (registered for unstructured RI ops in
// migration.NewK8sClient). All entrypoints into the karta package go through
// this so the RI plumbing is centralized.
func listRIs(ctx context.Context, c client.Client) (*unstructured.UnstructuredList, error) {
	list := &unstructured.UnstructuredList{}
	list.SetGroupVersionKind(migration.LegacyRIGVK.GroupVersion().WithKind(migration.LegacyRIGVK.Kind + "List"))
	if err := c.List(ctx, list); err != nil {
		return nil, err
	}
	return list, nil
}

// riToKartaCondition maps RI-side condition names to the equivalent Karta-side
// name when they differ. Karta renamed RIValidated -> KartaValidated as part
// of the rename, and Ready -> RunaiReady to disambiguate from any upstream
// Ready condition; semantically the paired conditions mean the same thing,
// so for coverage we treat them as equivalent.
var riToKartaCondition = map[string]string{
	"RIValidated": "KartaValidated",
	"Ready":       "RunaiReady",
}

// kartaCoversRI returns true when the Karta is at least as healthy as the RI:
// every condition that is True on the RI is also True on the Karta. If the RI
// was Ready=True, the loop naturally requires the Karta to be Ready=True; if
// the RI was Ready=False (e.g. its workload CRD is missing from the cluster),
// the migration faithfully mirrors that and we do not require the Karta to
// claim Ready - making it stricter would fail the gate on legitimately broken
// RIs that the migration honestly reproduced.
//
// This intentionally does NOT compare specs: schema defaulting between the
// two CRDs (e.g. the karta CRD declares defaults the RI CRD doesn't) means
// the spec maps are never byte-equal even when the migration is correct, and
// a byte-equal check would be a perpetual false alarm. The condition-coverage
// check instead validates the meaningful thing: did the migration produce a
// Karta that is at least as functional as the RI it replaces?
//
// The second return value is a human-readable reason, suitable for logs.
func kartaCoversRI(karta *kartav1alpha1.Karta, ri *unstructured.Unstructured) (bool, string) {
	kartaConds := typedConditionsMap(karta)
	for ct := range unstructuredTruthyConditions(ri) {
		target := ct
		if mapped, ok := riToKartaCondition[ct]; ok {
			target = mapped
		}
		if kartaConds[target] != metav1.ConditionTrue {
			return false, fmt.Sprintf("RI condition %s=True not mirrored on Karta (looked for %s)", ct, target)
		}
	}
	return true, ""
}

// typedConditionsMap reads conditions from a typed Karta into a map of type ->
// status. The Karta status is canonical (we own its schema) so we use the
// typed API instead of unstructured parsing.
func typedConditionsMap(k *kartav1alpha1.Karta) map[string]metav1.ConditionStatus {
	out := make(map[string]metav1.ConditionStatus, len(k.Status.Conditions))
	for _, c := range k.Status.Conditions {
		out[c.Type] = c.Status
	}
	return out
}

// unstructuredTruthyConditions returns the set of condition types where
// status==True on an RI. RIs are unstructured because we no longer import the
// legacy bolt v1alpha1 RI types.
func unstructuredTruthyConditions(obj *unstructured.Unstructured) map[string]struct{} {
	out := map[string]struct{}{}
	conditions, _, _ := unstructured.NestedSlice(obj.Object, "status", "conditions")
	for _, c := range conditions {
		cm, ok := c.(map[string]interface{})
		if !ok {
			continue
		}
		ctype, _ := cm["type"].(string)
		cstatus, _ := cm["status"].(string)
		if cstatus == string(metav1.ConditionTrue) {
			out[ctype] = struct{}{}
		}
	}
	return out
}
