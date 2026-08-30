package handlers

import (
	"context"

	"github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1alpha1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type PvHandler struct {
	client.Client
}

func NewPvHandler(client client.Client) *PvHandler {
	return &PvHandler{
		Client: client,
	}
}

// ConstructCopiedPv constructs a new PersistentVolume object in the target namespace;
// the name should already be validated to be unique.
func (pvHandler *PvHandler) ConstructCopiedPv(
	originalPv *v1.PersistentVolume, datavolume *v1alpha1.DataVolume,
	pvName, pvcName, pvcNamespace string, isFirstCopyOfOriginal bool) *v1.PersistentVolume {
	newPv := &v1.PersistentVolume{
		TypeMeta: metav1.TypeMeta{
			Kind:       "PersistentVolume",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: pvName,
			Labels: map[string]string{
				v1alpha1.DataVolumeIdLabel:             datavolume.Spec.DatavolumeId,
				v1alpha1.DataVolumeNameLabel:           datavolume.Name,
				v1alpha1.DataVolumePvForNamespaceLabel: pvcNamespace,
			},
		},
	}
	datavolume.SetOwner(newPv)

	pvHandler.deepCopySpec(originalPv, newPv)

	newPv.Spec.ClaimRef = &v1.ObjectReference{
		Name:      pvcName,
		Namespace: pvcNamespace,
	}

	newPv.Spec.PersistentVolumeReclaimPolicy = v1.PersistentVolumeReclaimRetain
	if isFirstCopyOfOriginal && datavolume.Spec.ShouldDeleteOriginalVolume {
		newPv.Spec.PersistentVolumeReclaimPolicy = v1.PersistentVolumeReclaimDelete
	}

	return newPv
}

func (pvHandler *PvHandler) deepCopySpec(fromPvc, toPvc *v1.PersistentVolume) {
	fromPvc.Spec.DeepCopyInto(&toPvc.Spec)
}

// GetPvsOfDatavolume returns a map of pvName -> pv.
// If excludeNs is not nil, it will exclude the PVs that are for this namespace.
func (pvHandler *PvHandler) GetPvsOfDatavolume(ctx context.Context, datavolume *v1alpha1.DataVolume, excludeNs *string) (map[string]*v1.PersistentVolume, error) {
	pvs, err := pvHandler.ListPvsOfDatavolume(ctx, datavolume)
	if err != nil {
		return nil, err
	}

	result := map[string]*v1.PersistentVolume{}
	for i := range pvs {
		pv := &pvs[i]

		if datavolume.IsOwner(pv) == -1 {
			continue
		}
		if excludeNs != nil && pvHandler.GetNamespaceOfPv(pv) == *excludeNs {
			continue
		}

		result[pv.Name] = pv
	}

	return result, nil
}

// GetPvOfPvc - returns the PV of the given PVC.
func (pvHandler *PvHandler) GetPvOfPvc(ctx context.Context, pvc *v1.PersistentVolumeClaim) (*v1.PersistentVolume, error) {
	if pvc.Spec.VolumeName == "" {
		return nil, nil
	}

	pv := &v1.PersistentVolume{}
	err := pvHandler.Get(ctx, client.ObjectKey{Name: pvc.Spec.VolumeName}, pv)
	if err != nil {
		return nil, err
	}

	return pv, nil
}

// GetAllDatavolumePvsForNamespace - PVs are not namespaced; but the ones we create are always to match a namespaced pvc;
// we have a label on them, so this func will return the datavolme pvs that are "for" a specific namespace.
func (pvHandler *PvHandler) GetAllDatavolumePvsForNamespace(ctx context.Context, namespace string) ([]v1.PersistentVolume, error) {
	return pvHandler.listPvs(ctx,
		client.MatchingLabels{v1alpha1.DataVolumePvForNamespaceLabel: namespace},
		client.HasLabels{v1alpha1.DataVolumeIdLabel})
}

func (pvHandler *PvHandler) listPvs(ctx context.Context, opts ...client.ListOption) ([]v1.PersistentVolume, error) {
	pvs := &v1.PersistentVolumeList{}
	err := pvHandler.List(ctx, pvs, opts...)
	if err != nil {
		return nil, err
	}

	return pvs.Items, nil
}

// ListPvsOfDatavolume - returns all PVs that are owned by the given DataVolume;
// using the label selector v1alpha1.DataVolumeNameLabel
func (pvHandler *PvHandler) ListPvsOfDatavolume(ctx context.Context, datavolume *v1alpha1.DataVolume) ([]v1.PersistentVolume, error) {
	return pvHandler.listPvs(ctx, client.MatchingLabels{
		v1alpha1.DataVolumeIdLabel: datavolume.Spec.DatavolumeId,
	})
}

func (pvHandler *PvHandler) ListPvsOfDatavolumeForNamespace(ctx context.Context, namespace string, datavolumeId string) ([]v1.PersistentVolume, error) {
	return pvHandler.listPvs(ctx, client.MatchingLabels{
		v1alpha1.DataVolumeIdLabel:             datavolumeId,
		v1alpha1.DataVolumePvForNamespaceLabel: namespace,
	})
}

func (pvHandler *PvHandler) GetNamespaceOfPv(pv *v1.PersistentVolume) string {
	return pv.Labels[v1alpha1.DataVolumePvForNamespaceLabel]
}
