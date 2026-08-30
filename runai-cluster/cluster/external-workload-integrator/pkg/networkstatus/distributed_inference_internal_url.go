package networkstatus

import (
	"context"
	"fmt"
	"net/url"
	"slices"

	endpoints "github.com/run-ai/runai/runai-cluster/cluster/common/endpoints"
	dromosv1alpha1 "github.com/run-ai/runai/runai-cluster/cluster/dromos/api/v1alpha1"
	"github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1alpha1"
	commonconstants "github.com/run-ai/runai/runai-cluster/common/constants"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	distributedInferenceWorkloadType = "DistributedInference"
	distributedInferenceServicePort  = "serving-port"
	defaultHTTPServicePort           = int32(80)
)

// reconcileDistributedInferenceInternalURL publishes the cluster-local URL owned by an
// internal-only distributed-inference networking instruction. ExternalWorkload.spec.urls
// is the EWI-owned URL carrier consumed by cluster-sync and endpoint discovery.
func (n *NetworkStatusUpdater) reconcileDistributedInferenceInternalURL(
	ctx context.Context,
	externalWorkload *v1alpha1.ExternalWorkload,
) error {
	if externalWorkload == nil || externalWorkload.Spec.Type != distributedInferenceWorkloadType {
		return nil
	}

	workloadRunnerName := externalWorkload.Labels[commonconstants.WorkloadRunnerNameLabel]
	if workloadRunnerName == "" {
		return nil
	}

	workloadRunner := &dromosv1alpha1.WorkloadRunner{}
	if err := n.cachedClient.Get(ctx, types.NamespacedName{
		Name: workloadRunnerName, Namespace: externalWorkload.Namespace,
	}, workloadRunner); err != nil {
		if apierrors.IsNotFound(err) {
			return nil
		}
		return fmt.Errorf("get WorkloadRunner %s/%s: %w", externalWorkload.Namespace, workloadRunnerName, err)
	}

	serviceNames, hasExternalAccess, err := distributedInferenceNetworkingInput(workloadRunner)
	if err != nil {
		return fmt.Errorf("read WorkloadRunner %s/%s networking input: %w",
			externalWorkload.Namespace, workloadRunnerName, err)
	}

	desiredURLs := preserveHigherPriorityDistributedInferenceURLs(
		externalWorkload.Spec.Urls,
		externalWorkload.Namespace,
		serviceNames,
	)
	if len(desiredURLs) == 0 && !hasExternalAccess {
		desiredURL, err := n.distributedInferenceInternalServiceURL(
			ctx, externalWorkload.Namespace, serviceNames)
		if err != nil {
			return err
		}
		if desiredURL != "" {
			desiredURLs = []string{desiredURL}
		}
	}
	if slices.Equal(externalWorkload.Spec.Urls, desiredURLs) {
		return nil
	}

	updatedExternalWorkload := externalWorkload.DeepCopy()
	updatedExternalWorkload.Spec.Urls = desiredURLs
	if err := n.cachedClient.Patch(ctx, updatedExternalWorkload, client.MergeFrom(externalWorkload)); err != nil {
		return fmt.Errorf("patch ExternalWorkload %s/%s internal URL: %w",
			externalWorkload.Namespace, externalWorkload.Name, err)
	}
	externalWorkload.Spec.Urls = desiredURLs
	return nil
}

// distributedInferenceNetworkingInput normalizes the two networking input shapes
// supported by the WorkloadRunner networking plugin. Generated-native workloads use
// the run.ai/endpoints annotation, while older clients use spec.instructions.networking.
func distributedInferenceNetworkingInput(
	workloadRunner *dromosv1alpha1.WorkloadRunner,
) ([]string, bool, error) {
	var legacyNetworking *dromosv1alpha1.NetworkingConfig
	if workloadRunner.Spec.Instructions != nil {
		legacyNetworking = workloadRunner.Spec.Instructions.Networking
	}

	endpointConfig, err := endpoints.Get(workloadRunner)
	if err != nil {
		return nil, false, err
	}
	if legacyNetworking != nil && endpointConfig != nil {
		return nil, false, fmt.Errorf("spec.instructions.networking and the %s annotation are mutually exclusive",
			endpoints.Annotation)
	}

	if endpointConfig != nil {
		serviceNames := make([]string, 0)
		hasExternalAccess := false
		for i := range endpointConfig.Endpoints {
			endpoint := &endpointConfig.Endpoints[i]
			hasExternalAccess = hasExternalAccess || endpoint.ExternalAccess != nil
			for j := range endpoint.Services {
				service := &endpoint.Services[j]
				serviceName := ""
				if service.Create != nil {
					serviceName = service.Create.Name
				} else if service.Ref != nil {
					serviceName = service.Ref.Name
				}
				serviceNames = appendUniqueServiceName(serviceNames, serviceName)
			}
		}
		return serviceNames, hasExternalAccess, nil
	}

	if legacyNetworking == nil {
		return nil, false, nil
	}
	serviceNames := make([]string, 0, len(legacyNetworking.Services))
	for i := range legacyNetworking.Services {
		serviceNames = appendUniqueServiceName(serviceNames, configuredServiceName(legacyNetworking.Services[i]))
	}
	return serviceNames, legacyNetworking.ExternalAccess != nil, nil
}

func appendUniqueServiceName(serviceNames []string, serviceName string) []string {
	if serviceName == "" || slices.Contains(serviceNames, serviceName) {
		return serviceNames
	}
	return append(serviceNames, serviceName)
}

func (n *NetworkStatusUpdater) distributedInferenceInternalServiceURL(
	ctx context.Context,
	namespace string,
	serviceNames []string,
) (string, error) {
	for _, serviceName := range serviceNames {
		service := &corev1.Service{}
		if err := n.cachedClient.Get(ctx, types.NamespacedName{Name: serviceName, Namespace: namespace}, service); err != nil {
			if apierrors.IsNotFound(err) {
				continue
			}
			return "", fmt.Errorf("get Service %s/%s: %w", namespace, serviceName, err)
		}
		if service.Spec.Type != corev1.ServiceTypeClusterIP {
			continue
		}

		for _, port := range service.Spec.Ports {
			if port.Name != distributedInferenceServicePort || port.Port <= 0 {
				continue
			}
			host := fmt.Sprintf("%s.%s.svc.cluster.local", service.Name, namespace)
			if port.Port != defaultHTTPServicePort {
				host = fmt.Sprintf("%s:%d", host, port.Port)
			}
			return "http://" + host, nil
		}
	}

	return "", nil
}

func preserveHigherPriorityDistributedInferenceURLs(
	existingURLs []string,
	namespace string,
	serviceNames []string,
) []string {
	preservedURLs := make([]string, 0, len(existingURLs))
	for _, existingURL := range existingURLs {
		if !isManagedDistributedInferenceURL(existingURL, namespace, serviceNames) {
			preservedURLs = append(preservedURLs, existingURL)
		}
	}

	if len(preservedURLs) == 0 {
		return nil
	}
	return preservedURLs
}

func configuredServiceName(configuredService dromosv1alpha1.NetworkingServiceItem) string {
	if configuredService.Create != nil {
		return configuredService.Create.Name
	}
	if configuredService.Ref != nil {
		return configuredService.Ref.Name
	}
	return ""
}

func isManagedDistributedInferenceURL(
	rawURL string,
	namespace string,
	serviceNames []string,
) bool {
	parsedURL, err := url.Parse(rawURL)
	if err != nil || parsedURL.Scheme != protocolHTTP || parsedURL.User != nil ||
		(parsedURL.Path != "" && parsedURL.Path != "/") || parsedURL.RawQuery != "" || parsedURL.Fragment != "" {
		return false
	}

	for _, serviceName := range serviceNames {
		if parsedURL.Hostname() == fmt.Sprintf("%s.%s.svc.cluster.local", serviceName, namespace) {
			return true
		}
	}
	return false
}
