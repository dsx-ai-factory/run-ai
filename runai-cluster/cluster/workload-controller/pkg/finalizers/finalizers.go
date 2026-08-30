package finalizers

import (
	"context"
	"fmt"
	"net/http"

	"github.com/run-ai/runai/runai-cluster/cluster/common/constants"
	workloadCommon "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/common/constants"
	logf "sigs.k8s.io/controller-runtime/pkg/log"

	kubeflowv1 "github.com/kubeflow/training-operator/pkg/apis/kubeflow.org/v1"
	openshiftV1 "github.com/openshift/api/route/v1"
	runaijobv1 "github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1"
	"github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/common"
	mpiv2beta1 "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/mpijob/v2beta1"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var fLog = logf.Log.WithName("finalizer")

// HasWorkloadFinalizer check if a resource contains our finalizer
func HasWorkloadFinalizer(om metav1.ObjectMeta) bool {
	if om.Finalizers == nil {
		return false
	}
	for _, finalizer := range om.Finalizers {
		if finalizer == workloadCommon.FinalizerName {
			return true
		}
	}
	return false
}

// RemoveResources remove all the resources which were created by the workload when removing it
func RemoveResources(ctx context.Context, status common.WorkloadStatus, k8sClient client.Client) error {
	log := logf.FromContext(ctx)
	for i := len(status.CreatedResources) - 1; i >= 0; i-- {

		resource := status.CreatedResources[i]
		err := RemoveResource(ctx, resource, k8sClient)
		if err != nil {
			if err.Error() == fmt.Errorf("unsupported resource %+v", resource).Error() {
				log.Error(err, "skipped")
				continue
			}
			return err
		}
	}

	return nil
}

func RemoveResource(ctx context.Context, resource common.ResourceTypeAndName, k8sClient client.Client) error {

	var err error

	resrcId := metav1.ObjectMeta{
		Name:      resource.Name,
		Namespace: resource.Namespace,
	}

	var toDelete client.Object

	switch resource.Kind {
	case constants.KindRunaiJob:
		toDelete = &runaijobv1.RunaiJob{
			ObjectMeta: resrcId,
		}
	case constants.KindMpiJob:
		toDelete = &mpiv2beta1.MPIJob{
			ObjectMeta: resrcId,
		}
	case constants.KindPyTorchJob:
		toDelete = &kubeflowv1.PyTorchJob{
			ObjectMeta: resrcId,
		}
	case constants.KindTFJob:
		toDelete = &kubeflowv1.TFJob{
			ObjectMeta: resrcId,
		}
	case constants.KindXGBoostJob:
		toDelete = &kubeflowv1.XGBoostJob{
			ObjectMeta: resrcId,
		}
	case constants.KindPvc:
		fLog.Info("bypass PVC removal", "id", resource)
		return nil
	case constants.KindIngress:
		toDelete = &networkingv1.Ingress{
			ObjectMeta: resrcId,
		}
	case constants.KindRoute:
		toDelete = &openshiftV1.Route{
			ObjectMeta: resrcId,
		}
	case constants.KindConfigMap:
		toDelete = &corev1.ConfigMap{
			ObjectMeta: resrcId,
		}
	default:
		return fmt.Errorf("unsupported resource %+v", resource)
	}

	err = k8sClient.Delete(ctx, toDelete)
	if err != nil {
		if apiErr, ok := err.(*errors.StatusError); ok && (apiErr.ErrStatus.Code == http.StatusNotFound) {
			fLog.Info("resource already deleted", "id", resource)
		} else {
			fLog.Error(err, "failed to remove resource", "id", resource)
			return err
		}
	}
	fLog.Info("deleted resource", "id", resource)
	return nil
}
