package kartacontroller

import (
	"context"
	"encoding/json"

	"github.com/rs/zerolog/log"

	kartav1alpha1 "github.com/run-ai/karta/pkg/api/runai/v1alpha1"
	"github.com/run-ai/runai/runai-cluster/cluster/external-workload-integrator/pkg/kartacontroller/handlers"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func (rc *KartaController) patchStatusWithNewConditions(
	ctx context.Context,
	karta *kartav1alpha1.Karta,
	originalStatus *kartav1alpha1.KartaStatus) error {
	if !handlers.WasStatusUpdated(&karta.Status, originalStatus) {
		return nil
	}

	patchBytes, err := updateConditionsStatusPatchBytes(karta.Status.Conditions)
	if err != nil {
		log.Error().Msgf("Failed to json.Marshal patch - "+
			"update init conditions of Karta <%s>, error: %s",
			karta.Name, err.Error())
		return err
	}

	patch := client.RawPatch(types.MergePatchType, patchBytes)
	err = rc.Client.Status().Patch(ctx, karta, patch)
	if err != nil {
		log.Error().Msgf("Failed to update status for Karta <%s>, error: %v", karta.Name, err)
		return err
	}

	return nil
}

func updateConditionsStatusPatchBytes(conditions []metav1.Condition) ([]byte, error) {
	return json.Marshal(map[string]interface{}{"status": map[string]interface{}{
		"conditions": conditions}})
}
