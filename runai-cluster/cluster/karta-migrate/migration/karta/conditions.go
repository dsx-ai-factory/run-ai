package karta

import (
	"context"

	kartav1alpha1 "github.com/run-ai/karta/pkg/api/runai/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// copyRIConditionsToKarta copies the RI's status.conditions onto the Karta
// verbatim, with two minimal interventions:
//
//  1. The condition renames via riToKartaCondition (RIValidated ->
//     KartaValidated, Ready -> RunaiReady).
//  2. KartaValidated's status is recomputed using karta's own validator
//     (kartav1alpha1.NewKartaValidator) instead of carrying RIValidated's
//     truth. The RI validator and the karta validator enforce different
//     rules (we know from mgkrt that an RI can pass RIValidated and still
//     fail KartaValidator after migration). Computing KartaValidated here
//     means the migration writes the same truth EWI's first reconcile
//     would write, so the pre-upgrade gate is consistent with the
//     post-upgrade state.
//
// Per natasharomm PR review: the migration is not a reconciler. It does
// not check the workload CRD, it does not reconcile RBAC, and it does not
// derive RunaiReady from other conditions. CRDExists / RBACReady /
// SyncEnabled / RunaiReady come straight from the RI - whatever it
// claimed is what the Karta claims (RI Ready becomes Karta RunaiReady).
func copyRIConditionsToKarta(ctx context.Context, c client.Client, name string, ri *unstructured.Unstructured) error {
	k, err := getKarta(ctx, c, name)
	if err != nil {
		return err
	}

	out := renamedConditionsFromRI(ri)
	out = setKartaValidatedFromValidator(out, k)
	out = setReadyFromConditions(out)

	patch := client.MergeFrom(k.DeepCopy())
	k.Status.Conditions = out
	return c.Status().Patch(ctx, k, patch)
}

// renamedConditionsFromRI reads every condition from the RI's
// status.conditions and returns them as a typed slice, with the only
// rename (RIValidated -> KartaValidated) applied. Conditions whose type or
// status cannot be parsed are dropped silently - the migration writes only
// what it can faithfully read.
func renamedConditionsFromRI(ri *unstructured.Unstructured) []metav1.Condition {
	conditions, _, _ := unstructured.NestedSlice(ri.Object, "status", "conditions")
	out := make([]metav1.Condition, 0, len(conditions))
	now := metav1.Now()
	for _, c := range conditions {
		cm, ok := c.(map[string]interface{})
		if !ok {
			continue
		}
		ctype, _ := cm["type"].(string)
		cstatus, _ := cm["status"].(string)
		if ctype == "" || cstatus == "" {
			continue
		}
		if mapped, ok := riToKartaCondition[ctype]; ok {
			ctype = mapped
		}
		creason, _ := cm["reason"].(string)
		cmsg, _ := cm["message"].(string)
		// metav1.Condition requires a non-empty Reason. Default to the
		// canonical "<Type>" pattern EWI uses when the RI did not have one.
		if creason == "" {
			creason = ctype
		}
		out = append(out, metav1.Condition{
			Type:               ctype,
			Status:             metav1.ConditionStatus(cstatus),
			Reason:             creason,
			Message:            cmsg,
			LastTransitionTime: now,
		})
	}
	return out
}

// setKartaValidatedFromValidator overrides (or appends) the KartaValidated
// condition based on karta's own validator. All other conditions in the
// input slice are passed through untouched.
func setKartaValidatedFromValidator(conds []metav1.Condition, k *kartav1alpha1.Karta) []metav1.Condition {
	status := metav1.ConditionTrue
	message := ""
	if err := kartav1alpha1.NewKartaValidator(k).Validate(); err != nil {
		status = metav1.ConditionFalse
		message = err.Error()
	}
	for i := range conds {
		if conds[i].Type == "KartaValidated" {
			conds[i].Status = status
			conds[i].Message = message
			return conds
		}
	}
	return append(conds, metav1.Condition{
		Type:               "KartaValidated",
		Status:             status,
		Reason:             "KartaValidated",
		Message:            message,
		LastTransitionTime: metav1.Now(),
	})
}

// setReadyFromConditions derives the aggregate Ready condition (True only when
// both CRDExists and KartaValidated are True) from the conditions already in
// the slice, and appends or overrides it.
func setReadyFromConditions(conds []metav1.Condition) []metav1.Condition {
	isTrue := func(condType string) bool {
		for i := range conds {
			if conds[i].Type == condType {
				return conds[i].Status == metav1.ConditionTrue
			}
		}
		return false
	}

	status := metav1.ConditionFalse
	message := "The Karta is not ready (CRDExists and KartaValidated must both be true)"
	if isTrue("CRDExists") && isTrue("KartaValidated") {
		status = metav1.ConditionTrue
		message = ""
	}

	for i := range conds {
		if conds[i].Type == "Ready" {
			conds[i].Status = status
			conds[i].Reason = "Ready"
			conds[i].Message = message
			return conds
		}
	}
	return append(conds, metav1.Condition{
		Type:               "Ready",
		Status:             status,
		Reason:             "Ready",
		Message:            message,
		LastTransitionTime: metav1.Now(),
	})
}
