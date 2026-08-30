package schema_upgrades

import (
	runaiv1 "github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type ownerReferenceUpdate struct {
	clusterCR *runaiv1.Cluster
}

func newOwnerReferenceUpgrader(clusterCR *runaiv1.Cluster) *ownerReferenceUpdate {
	return &ownerReferenceUpdate{
		clusterCR: clusterCR,
	}
}

func (o *ownerReferenceUpdate) shouldPerformUpgrade(resource *runaiv1.RunaiConfig) bool {
	shouldUpdateOwnerRef := true
	for _, ownerRef := range resource.ObjectMeta.OwnerReferences {
		if ownerRef.UID == o.clusterCR.UID &&
			ownerRef.Kind == runaiv1.ClusterKind &&
			ownerRef.Name == runaiv1.ClusterName &&
			ownerRef.APIVersion == runaiv1.GroupVersion.String() &&
			ownerRef.UID == o.clusterCR.UID {
			shouldUpdateOwnerRef = false
			break
		}
	}
	if shouldUpdateOwnerRef {
		return true
	}
	return false
}

func (o *ownerReferenceUpdate) updateRunaiConfig(resource *runaiv1.RunaiConfig) (*runaiv1.RunaiConfig, error) {
	resource.ObjectMeta.OwnerReferences = []metav1.OwnerReference{
		{
			APIVersion: runaiv1.GroupVersion.String(),
			Kind:       runaiv1.ClusterKind,
			Name:       runaiv1.ClusterName,
			UID:        o.clusterCR.UID,
		},
	}
	return resource, nil
}

func (o *ownerReferenceUpdate) upgradeHandlerIdentifier() string {
	return "affinity-and-tolerations-update"
}
