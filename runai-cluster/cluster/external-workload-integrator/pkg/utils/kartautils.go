package utils

import (
	"context"
	"fmt"

	kartav1alpha1 "github.com/run-ai/karta/pkg/api/runai/v1alpha1"
	"github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1alpha1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func GetGvkOfKarta(karta *kartav1alpha1.Karta) *schema.GroupVersionKind {
	gvk := karta.Spec.StructureDefinition.RootComponent.Kind
	if gvk == nil {
		return nil
	}
	return &schema.GroupVersionKind{Group: gvk.Group, Version: gvk.Version, Kind: gvk.Kind}
}

func GetAdditionalGvksOfKarta(karta *kartav1alpha1.Karta) []schema.GroupVersionKind {
	additionalGvks := make([]schema.GroupVersionKind, 0, len(karta.Spec.StructureDefinition.AdditionalChildKinds))
	for _, gvk := range karta.Spec.StructureDefinition.AdditionalChildKinds {
		additionalGvks = append(additionalGvks, schema.GroupVersionKind{Group: gvk.Group, Version: gvk.Version, Kind: gvk.Kind})
	}
	return additionalGvks
}

// GetKartaOfGvk returns the Karta matching the given GVK.
// Also allows empty version in "gvk" - in case we don't know the version;
// in this case, we will return the first ready Karta with matching group and kind,
// or if there is no ready Karta - the first non-ready one.
func GetKartaOfGvk(ctx context.Context, client client.Client, gvk schema.GroupVersionKind) (*kartav1alpha1.Karta, error) {
	kartaPerGvkLabelSelector := v1alpha1.GetKartaPerGvkLabelSelectors(gvk)

	kartas := &kartav1alpha1.KartaList{}
	err := client.List(ctx, kartas, kartaPerGvkLabelSelector)
	if err != nil {
		return nil, err
	}

	var fallback *kartav1alpha1.Karta
	for i := range kartas.Items {
		karta := kartas.Items[i]
		kartaGvk := GetGvkOfKarta(&karta)
		if kartaGvk == nil {
			continue
		}

		if gvk.Version != "" {
			if *kartaGvk == gvk {
				return &kartas.Items[i], nil
			}
			continue
		}

		if kartaGvk.Group == gvk.Group && kartaGvk.Kind == gvk.Kind {
			if isKartaReady(&karta) {
				return &karta, nil
			}
			if fallback == nil {
				fallback = &kartas.Items[i]
			}
		}
	}

	if fallback != nil {
		return fallback, nil
	}

	return nil, errors.NewNotFound(schema.GroupResource{Group: kartav1alpha1.GroupVersion.Group, Resource: "Karta"},
		fmt.Sprintf("for gvk %s", gvk.String()))
}

func isKartaReady(karta *kartav1alpha1.Karta) bool {
	for _, condition := range karta.Status.Conditions {
		if condition.Type == string(v1alpha1.RunaiReady) &&
			condition.Status == metav1.ConditionTrue {
			return true
		}
	}
	return false
}
