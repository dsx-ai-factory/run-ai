package distributed_status

import (
	"context"
	"fmt"
	nvapi "github.com/NVIDIA/k8s-dra-driver-gpu/api/nvidia.com/resource/v1beta1"
	"net/url"
	"reflect"
	"sort"
	"time"

	lwsv1 "sigs.k8s.io/lws/api/leaderworkerset/v1"

	"github.com/run-ai/runai/runai-cluster/cluster/common/constants"
	"github.com/run-ai/runai/runai-cluster/cluster/inference-workload-controller/internal/builders"
	distresources "github.com/run-ai/runai/runai-cluster/cluster/inference-workload-controller/internal/resources"
	inferenceaggregator "github.com/run-ai/runai/runai-cluster/cluster/inference-workload-controller/internal/status/aggregator"
	statuscontract "github.com/run-ai/runai/runai-cluster/cluster/workload-common/status/contract"
	commondistributedinference "github.com/run-ai/runai/runai-cluster/cluster/workload-common/status/distributedinference"
	wccommon "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/common"
	dinfw "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/distributedinferenceworkload/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logger "sigs.k8s.io/controller-runtime/pkg/log"
)

const (
	PhaseReady        = string(statuscontract.Ready)
	PhasePending      = string(statuscontract.Pending)
	PhaseInitializing = string(statuscontract.Initializing)
	PhaseCreating     = string(statuscontract.Creating)
	PhaseDeleting     = string(statuscontract.Deleting)
	PhaseFailed       = string(statuscontract.Failed)
	PhaseDegraded     = string(statuscontract.Degraded)
)

const (
	HttpPort = 80
)

var log = logger.Log.WithName("distributed-status")

type DistributedStatusUpdaterInterface interface {
	Update(ctx context.Context, workload *dinfw.DistributedInferenceWorkload,
		originalWorkload *dinfw.DistributedInferenceWorkload, builtResources []client.Object, buildErr *error) error
}

type DistributedStatusUpdater struct {
	k8sClient           client.Client
	lwsStatusCalculator LwsStatusCalculatorInterface
	resourcesManager    distresources.DistributedResourcesManagerInterface
}

// NewDistributedStatusUpdater creates a new DistributedStatusUpdater instance
func NewDistributedStatusUpdater(k8sClient client.Client) *DistributedStatusUpdater {
	return &DistributedStatusUpdater{
		k8sClient:           k8sClient,
		lwsStatusCalculator: NewLwsStatusCalculator(k8sClient),
		resourcesManager:    distresources.NewDistributedResourcesManager(k8sClient),
	}
}

// Update updates the status of the distributed inference workload
func (s *DistributedStatusUpdater) Update(ctx context.Context, workload *dinfw.DistributedInferenceWorkload,
	originalWorkload *dinfw.DistributedInferenceWorkload, builtResources []client.Object, buildErr *error) error {
	log := logger.FromContext(ctx)
	log.Info("updating workload status", "workload", workload.Name)

	// update the status
	err := s.updateStatusFields(ctx, workload, builtResources, buildErr)
	if err != nil {
		log.Error(err, "failed to update status fields", "workload", workload.Name)
		return err
	}

	// compare to previous status
	if !s.statusChanged(&originalWorkload.Status, &workload.Status) {
		log.V(4).Info("status unchanged, skipping patch", "workload", workload.Name)
		return nil
	}

	// patch the status
	patch := client.MergeFrom(originalWorkload)
	err = s.k8sClient.Status().Patch(ctx, workload, patch)
	if err != nil {
		log.Error(err, "failed to patch workload status", "workload", workload.Name)
		return err
	}

	log.V(4).Info("workload status patched successfully", "workload", workload.Name, "phase", workload.Status.Phase)
	return nil
}

func (s *DistributedStatusUpdater) updateStatusFields(ctx context.Context, workload *dinfw.DistributedInferenceWorkload,
	builtResources []client.Object, buildErr *error) error {
	// reset all the status fields that we're responsible for, in order not to keep old values with us,
	// in case we're only updating some of them
	workload.Status.DistributedInferenceStatus = wccommon.DistributedInferenceStatus{}
	workload.Status.NetworkStatus = wccommon.NetworkStatus{}

	// handle first acknowledgement
	if workload.Status.FirstAcknowledgedTime.IsZero() {
		log.V(4).Info("setting status first acknowledged time", "workload", workload.Name)
		workload.Status.FirstAcknowledgedTime = &metav1.Time{Time: time.Now()}
	}

	// set workload generation
	log.V(4).Info("setting status observed generation", "workload", workload.Name, "generation", workload.Generation)
	workload.Status.ObservedGeneration = workload.Generation

	// set last status change time
	now := metav1.Now()
	log.V(6).Info("setting last status change time", "workload", workload.Name, "time", now.Format(time.RFC3339))
	workload.Status.LastStatusChangeTime = &now

	// get distributed resources
	resources, missingResources, err := s.resourcesManager.GetResources(ctx, builtResources)
	if err != nil {
		log.Error(err, "failed to get distributed resources", "workload", workload.Name)
		return err
	}

	var lwsPhase string
	var replicas, readyReplicas int32
	var phaseMessage string
	if resources.Lws != nil {
		lwsPhase, replicas, readyReplicas, phaseMessage, err = s.lwsStatusCalculator.GetStatus(ctx, workload, resources.Lws)
		if err != nil {
			log.Error(err, "failed to get lws status", "workload", workload.Name)
			return err
		}
		log.V(4).Info("lws status", "workload", workload.Name, "phase", lwsPhase, "replicas", replicas, "readyReplicas", readyReplicas)
	}

	// set phase and phase message
	s.setPhase(workload, buildErr, missingResources, resources.Lws, lwsPhase, resources.ComputeDomains, phaseMessage)

	// set other status fields
	if resources.Lws != nil {
		workload.Status.Replicas = replicas
		workload.Status.ReadyReplicas = readyReplicas
	}
	s.setServices(workload, resources)

	// publish the urls only if there are ready replicas
	if readyReplicas > 0 {
		s.setUrls(workload, resources)
	}

	return nil
}

// setPhase sets the phase and phase message for the workload
func (s *DistributedStatusUpdater) setPhase(workload *dinfw.DistributedInferenceWorkload, buildErr *error,
	missingResources []client.Object, lws *lwsv1.LeaderWorkerSet, lwsPhase string, computeDomains []*nvapi.ComputeDomain, phaseMessage string) {
	mnnvlReady := isComputeDomainsReady(computeDomains)
	result := commondistributedinference.ApplyWorkloadState(
		commondistributedinference.Result{
			Phase:   statuscontract.Phase(lwsPhase),
			Message: phaseMessage,
		},
		commondistributedinference.WorkloadState{
			Exists:            lws != nil,
			BuildError:        buildErrorMessage(buildErr),
			MissingResources:  missingResourceNames(missingResources),
			MNNVLEnabled:      len(computeDomains) > 0,
			MNNVLReady:        mnnvlReady,
			MarkedForDeletion: workload.DeletionTimestamp != nil,
		},
	)
	workload.Status.Phase = string(result.Phase)
	workload.Status.PhaseMessage = result.Message
}

func buildErrorMessage(buildErr *error) *string {
	if buildErr == nil {
		return nil
	}
	message := (*buildErr).Error()
	return &message
}

func missingResourceNames(resources []client.Object) []string {
	names := make([]string, 0, len(resources))
	for _, resource := range resources {
		kind := resource.GetObjectKind().GroupVersionKind().Kind
		names = append(names, fmt.Sprintf("%s/%s", kind, resource.GetName()))
	}
	return names
}

// statusChanged compares two status objects, ignoring LastStatusChangeTime and SyncStatus
func (s *DistributedStatusUpdater) statusChanged(previous, current *dinfw.DistributedInferenceWorkloadStatus) bool {
	// Copy current LastStatusChangeTime and SyncStatus to previous, to ignore them in comparison
	originalLastStatusChangeTime := previous.LastStatusChangeTime
	originalSyncStatus := previous.SyncStatus
	previous.LastStatusChangeTime = current.LastStatusChangeTime
	previous.SyncStatus = current.SyncStatus

	// Use DeepEqual to compare the whole structs
	changed := !reflect.DeepEqual(previous, current)

	// Restore previous LastStatusChangeTime and SyncStatus
	previous.LastStatusChangeTime = originalLastStatusChangeTime
	previous.SyncStatus = originalSyncStatus

	return changed
}

// buildURLFromService builds cluster-local URLs for all ports of a service
func (s *DistributedStatusUpdater) buildURLFromService(service *corev1.Service) ([]string, error) {
	if service == nil {
		return nil, fmt.Errorf("service is nil")
	}

	if len(service.Spec.Ports) == 0 {
		return nil, fmt.Errorf("service has no ports")
	}

	var urls []string
	for _, port := range service.Spec.Ports {
		var clusterUrl string
		clusterUrl = fmt.Sprintf("http://%s.%s.svc.cluster.local", service.Name, service.Namespace)
		if port.Port != HttpPort {
			// Include port number for non-default ports. This case is only
			// relevant for services with multiple ports - something we currently
			// don't do.
			clusterUrl = fmt.Sprintf("%s:%d", clusterUrl, port.Port)
		}
		urls = append(urls, clusterUrl)
	}

	return urls, nil
}

func (s *DistributedStatusUpdater) setServices(workload *dinfw.DistributedInferenceWorkload,
	resources distresources.DistributedResources) {
	serviceStatuses := make([]wccommon.ServiceStatus, 0)

	for _, service := range resources.Services {
		serviceStatus := wccommon.ServiceStatus{
			Name: service.Name,
			Type: service.Spec.Type,
		}
		for _, port := range service.Spec.Ports {
			externalPort := port.Port
			if port.NodePort != 0 {
				externalPort = port.NodePort
			}
			exposedServicePort := wccommon.ExposePort{
				Name:       port.Name,
				Protocol:   port.Protocol,
				Port:       externalPort,
				TargetPort: port.TargetPort,
			}
			serviceStatus.Ports = append(serviceStatus.Ports, exposedServicePort)
		}
		// Sort Ports by Port, then Name for determinism
		sort.Slice(serviceStatus.Ports, func(i, j int) bool {
			if serviceStatus.Ports[i].Port != serviceStatus.Ports[j].Port {
				return serviceStatus.Ports[i].Port < serviceStatus.Ports[j].Port
			}
			return serviceStatus.Ports[i].Name < serviceStatus.Ports[j].Name
		})
		serviceStatuses = append(serviceStatuses, serviceStatus)
	}

	// Sort serviceStatuses by Name for determinism
	sort.Slice(serviceStatuses, func(i, j int) bool {
		return serviceStatuses[i].Name < serviceStatuses[j].Name
	})

	if len(serviceStatuses) > 0 {
		workload.Status.NetworkStatus.ServiceStatuses = serviceStatuses
	}
}

func (s *DistributedStatusUpdater) setUrls(workload *dinfw.DistributedInferenceWorkload,
	resources distresources.DistributedResources) {
	servingUrlSet := make(map[string]struct{})
	externalUrlSet := make(map[string]struct{})
	servingUrls := make([]string, 0)
	externalUrls := make([]string, 0)

	// for servingUrls, we prioritize ingress/route/httproute over service.
	// that means we're first adding the external urls to it (from ingress/route/httproute), and only if no such, we add the internal url (from service)
	addServingURL := func(u string) {
		if _, exists := servingUrlSet[u]; !exists {
			servingUrlSet[u] = struct{}{}
			servingUrls = append(servingUrls, u)
		}
	}
	addExternalURL := func(u string) {
		if _, exists := externalUrlSet[u]; !exists {
			externalUrlSet[u] = struct{}{}
			externalUrls = append(externalUrls, u)
		}
	}

	servingNetworkResourceName := builders.GetServingPortNetworkResourceName(workload.Name)

	for _, ingress := range resources.Ingresses {
		if len(ingress.Spec.Rules) > 0 {
			ingressUrl, err := inferenceaggregator.BuildURLFromIngress(ingress)
			if err != nil {
				log.Error(err, "failed to build URL from ingress", "ingress", ingress.Name)
				ingressUrl = &url.URL{Host: "http://"}
			}

			if ingress.Name == servingNetworkResourceName {
				addServingURL(ingressUrl.String())
			} else {
				addExternalURL(ingressUrl.String())
			}
		}
	}

	for _, route := range resources.Routes {
		routeUrl, err := inferenceaggregator.BuildURLFromRoute(route)
		if err != nil {
			log.Error(err, "failed to build URL from route", "route", route.Name)
			routeUrl = &url.URL{Host: "http://"}
		}

		if route.Name == servingNetworkResourceName {
			addServingURL(routeUrl.String())
		} else {
			addExternalURL(routeUrl.String())
		}
	}

	for _, httpRoute := range resources.HTTPRoutes {
		if httpRoute == nil {
			continue
		}
		httpRouteUrl, err := inferenceaggregator.BuildURLFromHTTPRoute(httpRoute)
		if err != nil {
			log.Error(err, "failed to build URL from httpRoute", "httpRoute", httpRoute.Name)
			httpRouteUrl = &url.URL{Host: "http://"}
		}

		if httpRoute.Name == servingNetworkResourceName {
			addServingURL(httpRouteUrl.String())
		} else {
			addExternalURL(httpRouteUrl.String())
		}
	}

	// if no external serving url was found, add the internal url from service
	if len(servingUrls) == 0 {
		for _, service := range resources.Services {
			if service.Name == servingNetworkResourceName {
				serviceUrls, err := s.buildURLFromService(service)
				if err != nil {
					log.Error(err, "failed to build URL from service", "service", service.Name)
					continue
				}

				servingUrls = append(servingUrls, serviceUrls...)
			}
		}
	}

	// Sort the urls for determinism
	sort.Strings(servingUrls)
	sort.Strings(externalUrls)

	if len(servingUrls) > 0 {
		workload.Status.DistributedInferenceStatus.ServingUrls = servingUrls
	}

	if len(externalUrls) > 0 {
		workload.Status.DistributedInferenceStatus.ExternalUrls = externalUrls
	}
}

// isComputeDomainsReady checks if all compute domains are in status ready
func isComputeDomainsReady(computeDomains []*nvapi.ComputeDomain) bool {
	for _, cd := range computeDomains {
		if cd.Status.Status != constants.ComputeDomainStatusReady {
			return false
		}
	}
	return true
}
