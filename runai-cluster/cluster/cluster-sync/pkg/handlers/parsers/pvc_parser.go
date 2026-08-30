package parsers

import (
	"fmt"

	"github.com/run-ai/runai/runai-application-packages/convutils"
	commonconstants "github.com/run-ai/runai/runai-cluster/cluster/common/constants"

	k8sobjectstracker "github.com/run-ai/runai/runai-clients/k8s-objects-tracker-client/generated"
	"github.com/run-ai/runai/runai-cluster/cluster/cluster-sync/pkg/util"
	corev1 "k8s.io/api/core/v1"
)

type ParsedPvc struct {
	Scope           ScopeTypeAndId
	CreationRequest k8sobjectstracker.PVCCreationRequest
}

func ParsePvcToServerScheme(pvc *corev1.PersistentVolumeClaim) (*ParsedPvc, error) {
	var origin k8sobjectstracker.NullableOrigin

	if util.IsAssetRelatedResource(&pvc.ObjectMeta) {
		origin.Set(&originAssets)
	} else if util.IsUserExposedPvcForRunai(pvc) {
		origin.Set(&originExisting)
	} else {
		return nil, fmt.Errorf("could not determine the origin of the PVC")
	}

	issues := util.ExtractIssuesFromObject(pvc.GetAnnotations())
	if pvc.Status.Phase == corev1.ClaimLost {
		issues = append(issues, k8sobjectstracker.Issue{
			Namespace: pvc.Namespace,
			Issue:     string(k8sobjectstracker.CLUSTEROBJECTISSUETYPE_LOST_PHASE),
		})
	}

	dataSharing := pvc.Labels[commonconstants.RunaiDataSharingLabel] == commonconstants.BoolTrueAsString

	result := ParsedPvc{
		CreationRequest: k8sobjectstracker.PVCCreationRequest{
			Origin:      origin,
			Issues:      issues,
			Phase:       *k8sobjectstracker.NewNullablePvcPhaseOptional(convutils.Ref(k8sobjectstracker.PvcPhaseOptional(pvc.Status.Phase))),
			DataSharing: *k8sobjectstracker.NewNullableBool(&dataSharing),
		},
	}

	if pvc.Spec.StorageClassName != nil {
		result.CreationRequest.StorageClassName = *k8sobjectstracker.NewNullableString(pvc.Spec.StorageClassName)
	}

	if pvc.Spec.Resources.Requests != nil {
		quantity, storageRequested := pvc.Spec.Resources.Requests[corev1.ResourceStorage]
		if storageRequested {
			result.CreationRequest.RequestedCapacity = *k8sobjectstracker.NewNullableString(convutils.Ref(quantity.String()))
		}
	}

	if pvc.Status.Capacity != nil {
		quantity, storageAllocated := pvc.Status.Capacity[corev1.ResourceStorage]
		if storageAllocated {
			result.CreationRequest.ActualCapacity = *k8sobjectstracker.NewNullableString(convutils.Ref(quantity.String()))
		}
	}

	err := parseScope(&result.Scope, &pvc.ObjectMeta)
	if err != nil {
		return nil, err
	}

	return &result, nil
}
