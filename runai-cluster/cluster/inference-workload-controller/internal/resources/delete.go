package resources

import (
	"context"
	"fmt"

	openshiftV1 "github.com/openshift/api/route/v1"
	"github.com/run-ai/runai/runai-cluster/cluster/common/constants"

	"github.com/rs/zerolog/log"
	"github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/common"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	servicev1 "knative.dev/serving/pkg/apis/serving/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// RemoveResources remove all the resources which were created by the workload when removing it
func RemoveResources(ctx context.Context, status common.WorkloadStatus, k8sClient client.Client) error {
	for i := len(status.CreatedResources) - 1; i >= 0; i-- {

		resource := status.CreatedResources[i]
		err := RemoveResource(ctx, resource, k8sClient)
		if err != nil {
			if err.Error() == fmt.Errorf("unsupported resource %+v", resource).Error() {
				log.Error().Msgf("got unsupported resource %+v, skiping", resource)
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
	case constants.KindPvc:
		log.Info().Msgf("By pass PVC deletion for now id %+v", resource)
		return nil
	case constants.KindService:
		//
		//   need to distinguish b/w networking service and knative service
		//
		if resource.APIVersion == constants.KnativeServiceApi {
			toDelete = &servicev1.Service{
				ObjectMeta: resrcId,
			}
		} else {
			toDelete = &corev1.Service{
				ObjectMeta: resrcId,
			}
		}
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
		if errors.IsNotFound(err) {
			log.Info().Msgf("resource already deleted id %+v", resource)
		} else {

			log.Error().Msgf("failed to remove resource %+v", resource)
			return err
		}
	}
	log.Info().Msgf("deleted resource %+v", resource)
	return nil
}
