package workloads

import (
	"context"
	"fmt"
	"github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/jobs"
	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

const (
	pvcNotFoundTemplate    = "PVC '%s' does not exist in namespace '%s', please create it first"
	failedToGetPvcTemplate = "Failed to get PVC '%s' in namespace '%s', please try again"
)

type ExistingPvcsValidator struct {
	k8sClient client.Client
}

func NewExistingPvcsValidator(k8sClient client.Client) *ExistingPvcsValidator {
	return &ExistingPvcsValidator{
		k8sClient: k8sClient,
	}
}

func (v ExistingPvcsValidator) Validate(ctx context.Context, workload *jobs.JobWorkload) error {
	log := logf.FromContext(ctx)

	pvcsFromSpec := workload.Spec.Pvcs
	if pvcsFromSpec == nil {
		log.V(4).Info("got no PVCs to validate")
		return nil
	}

	log.Info("checking if expected PVCs exist", "workload", workload.Name)

	for _, pvcFromSpec := range pvcsFromSpec.Deref() {
		pvcValue := pvcFromSpec.Value
		if pvcValue == nil {
			log.Info("got pvc with nil value")
			continue
		}
		if !ptr.Deref(pvcValue.ExistingPvc, false) {
			continue
		}

		var pvc corev1.PersistentVolumeClaim
		pvcName := ptr.Deref(pvcValue.ClaimName, "")
		err := v.k8sClient.Get(ctx, client.ObjectKey{Name: pvcName, Namespace: workload.Namespace}, &pvc)
		if err != nil {
			if k8serrors.IsNotFound(err) {
				log.Info("pvc does not exist in namespace", "pvc", pvcName, "namespace", workload.Namespace)
				return fmt.Errorf(pvcNotFoundTemplate, pvcName, workload.Namespace)
			} else {
				log.Error(err, "failed to get pvc", "pvc", pvcName, "namespace", workload.Namespace)
				return fmt.Errorf(failedToGetPvcTemplate, pvcName, workload.Namespace)
			}
		}
	}

	log.Info("all expected PVCs exist", "workload", workload.Name)
	return nil
}
