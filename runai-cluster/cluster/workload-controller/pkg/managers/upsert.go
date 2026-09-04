/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package managers

import (
	"bytes"
	"context"
	"fmt"
	"sync"

	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/constants"
	runaiJobV1 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-common/config"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-common/issue"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-controller/pkg/api/common"
	job "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-controller/pkg/api/jobs"
	mpi "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-controller/pkg/api/mpijob/v2beta1"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-controller/pkg/builders/common/resource"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-controller/pkg/utils/resources"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-controller/pkg/utils/workloads/status"
	kubeflowv1 "github.com/kubeflow/training-operator/pkg/apis/kubeflow.org/v1"
	nvapi "github.com/NVIDIA/k8s-dra-driver-gpu/api/nvidia.com/resource/v1beta1"
	routev1 "github.com/openshift/api/route/v1"
	v1 "k8s.io/api/core/v1"
	network "k8s.io/api/networking/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/util/errors"
	"knative.dev/pkg/logging"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	gatewayv1 "sigs.k8s.io/gateway-api/apis/v1"
)

type ResourceOptions struct {
	// DryRun - if true, the resource manager will not create the resource
	DryRun bool
}

type K8sResourceManager struct {
	client        client.Client
	apiReadClient client.Reader
	rg            issue.ResourceGrouper
}

func NewResourceManager(client client.Client, apiReadClient client.Reader, rg issue.ResourceGrouper) *K8sResourceManager {
	return &K8sResourceManager{
		client:        client,
		apiReadClient: apiReadClient,
		rg:            rg,
	}
}

func (krm *K8sResourceManager) getServices(ctx context.Context, labelMatch *map[string]string, workloadNS string, waitGroup *sync.WaitGroup, serviceListCh chan v1.ServiceList, errChannel chan error) {
	defer waitGroup.Done()
	serviceList := &v1.ServiceList{}
	err := krm.client.List(ctx, serviceList,
		client.InNamespace(workloadNS),
		client.MatchingLabels(*labelMatch))
	if err != nil {
		errChannel <- err
		return
	}
	serviceListCh <- *serviceList
}

func (krm *K8sResourceManager) getIngresses(ctx context.Context, labelMatch *map[string]string, workloadNS string, waitGroup *sync.WaitGroup, ingressListCh chan network.IngressList, errChannel chan error) {
	defer waitGroup.Done()
	ingressList := &network.IngressList{}
	err := krm.client.List(ctx, ingressList,
		client.InNamespace(workloadNS),
		client.MatchingLabels(*labelMatch))
	if err != nil {
		errChannel <- err
		return
	}
	ingressListCh <- *ingressList
}

func (krm *K8sResourceManager) getRoutes(ctx context.Context, labelMatch *map[string]string, workloadNS string, waitGroup *sync.WaitGroup, routeChan chan routev1.RouteList, errChannel chan error) {
	defer waitGroup.Done()
	routeList := &routev1.RouteList{}
	err := krm.client.List(ctx, routeList,
		client.InNamespace(workloadNS),
		client.MatchingLabels(*labelMatch))
	if err != nil {
		errChannel <- err
		return
	}
	routeChan <- *routeList
}

func (krm *K8sResourceManager) getHTTPRoutes(ctx context.Context, labelMatch *map[string]string, workloadNS string, waitGroup *sync.WaitGroup, httpRouteChan chan gatewayv1.HTTPRouteList, errChannel chan error) {
	defer waitGroup.Done()
	httpRouteList := &gatewayv1.HTTPRouteList{}
	err := krm.client.List(ctx, httpRouteList,
		client.InNamespace(workloadNS),
		client.MatchingLabels(*labelMatch))
	if err != nil {
		if meta.IsNoMatchError(err) {
			log := logf.FromContext(ctx)
			log.V(2).Info("HTTPRoute CRD not installed, skipping HTTPRoute listing")
			httpRouteChan <- gatewayv1.HTTPRouteList{}
			return
		}
		errChannel <- err
		return
	}
	httpRouteChan <- *httpRouteList
}

func (krm *K8sResourceManager) GetAllResources(ctx context.Context, aggr *resource.Aggregator, kindList []common.Resource) ([]client.Object, error) {
	workload := aggr.GetWorkload()
	log := logf.FromContext(ctx)

	if workload == nil {
		return nil, fmt.Errorf("workload is nil")
	}

	var (
		serviceListCh   chan v1.ServiceList
		ingressListCh   chan network.IngressList
		routeListCh     chan routev1.RouteList
		httpRouteListCh chan gatewayv1.HTTPRouteList
		pvcListCh       chan v1.PersistentVolumeClaimList
		jobChannel      chan client.Object
		computeDomainCh chan *nvapi.ComputeDomain
		secretListCh    chan v1.SecretList
		labelMatch      = map[string]string{constants.LabelRelease: workload.Name}
	)

	getResources := krm.groupResourceByType(ctx, kindList)

	// Always list networking resource types to detect orphans from mode transitions
	// (e.g., leftover Ingresses after switching to Gateway API mode).
	// When a type is disabled, its kindMap entry is empty, so any found resources
	// are correctly identified as dangling by getResourcesFromChannels.
	for _, kind := range []string{constants.KindIngress, constants.KindHTTPRoute} {
		if _, exists := getResources[kind]; !exists {
			getResources[kind] = make(map[string]common.Resource)
		}
	}
	if config.IsOpenshift() {
		if _, exists := getResources[constants.KindRoute]; !exists {
			getResources[constants.KindRoute] = make(map[string]common.Resource)
		}
	}

	wg := sync.WaitGroup{}
	errChan := make(chan error, len(getResources))

	for kind, resourceList := range getResources {
		switch kind {
		case constants.KindService:
			wg.Add(1)
			serviceListCh = make(chan v1.ServiceList, 1)
			go krm.getServices(ctx, &labelMatch, workload.Namespace, &wg, serviceListCh, errChan)
		case constants.KindIngress:
			wg.Add(1)
			ingressListCh = make(chan network.IngressList, 1)
			go krm.getIngresses(ctx, &labelMatch, workload.Namespace, &wg, ingressListCh, errChan)
		case constants.KindRoute:
			wg.Add(1)
			routeListCh = make(chan routev1.RouteList, 1)
			go krm.getRoutes(ctx, &labelMatch, workload.Namespace, &wg, routeListCh, errChan)
		case constants.KindHTTPRoute:
			wg.Add(1)
			httpRouteListCh = make(chan gatewayv1.HTTPRouteList, 1)
			go krm.getHTTPRoutes(ctx, &labelMatch, workload.Namespace, &wg, httpRouteListCh, errChan)
		case constants.KindPvc:
			wg.Add(1)
			pvcListCh = make(chan v1.PersistentVolumeClaimList, 1)
			go krm.getPvc(ctx, &labelMatch, workload.Namespace, &wg, pvcListCh, errChan)
		case constants.JobType:
			wg.Add(1)
			jobChannel = make(chan client.Object, 1)
			go krm.checkAndGetJobType(ctx, resourceList, &wg, jobChannel, errChan)
		case constants.KindComputeDomain:
			wg.Add(1)
			computeDomainCh = make(chan *nvapi.ComputeDomain, 1)
			go krm.getComputeDomain(ctx, &labelMatch, workload.Name, workload.Namespace, &wg, computeDomainCh, errChan)
		case constants.KindSecret:
			wg.Add(1)
			secretListCh = make(chan v1.SecretList, 1)
			go krm.getSecret(ctx, &labelMatch, workload.Namespace, &wg, secretListCh, errChan)
		default:
			log.V(2).Info("resource type was not recognized, bypassing", "typeKind", kind)
		}
	}

	wg.Wait()
	closeChannels(serviceListCh, ingressListCh, routeListCh, httpRouteListCh, pvcListCh, jobChannel, computeDomainCh, secretListCh, errChan)

	return krm.getResourcesFromChannels(aggr, &getResources, jobChannel, serviceListCh, ingressListCh, routeListCh, httpRouteListCh, pvcListCh, computeDomainCh, secretListCh, errChan)
}

func closeChannels(serviceListCh chan v1.ServiceList, ingressListCh chan network.IngressList, routeListCh chan routev1.RouteList,
	httpRouteListCh chan gatewayv1.HTTPRouteList, pvcListCh chan v1.PersistentVolumeClaimList,
	jobChannel chan client.Object, computeDomainCh chan *nvapi.ComputeDomain, secretListCh chan v1.SecretList, errChan chan error) {

	if serviceListCh != nil {
		close(serviceListCh)
	}
	if ingressListCh != nil {
		close(ingressListCh)
	}
	if routeListCh != nil {
		close(routeListCh)
	}
	if httpRouteListCh != nil {
		close(httpRouteListCh)
	}
	if pvcListCh != nil {
		close(pvcListCh)
	}
	if jobChannel != nil {
		close(jobChannel)
	}
	if computeDomainCh != nil {
		close(computeDomainCh)
	}
	if secretListCh != nil {
		close(secretListCh)
	}
	close(errChan)
}

func (krm *K8sResourceManager) checkAndGetJobType(ctx context.Context, resourceMap map[string]common.Resource, wg *sync.WaitGroup, jobChannel chan client.Object, errChan chan error) {
	defer wg.Done()
	log := logging.FromContext(ctx)
	var jobResource *common.Resource

	var resListStr []string
	for _, res := range resourceMap {
		// either our pointer is nil or we found a resource that is not the basic RunaiJob
		if jobResource == nil || res.Kind.Kind != constants.KindRunaiJob {
			jobResource = &res
		}

		resListStr = append(resListStr, res.Kind.Kind)
	}
	if len(resListStr) > 1 {
		log.Warn("resource count more than 1, fetching the most relevant", "resourceMap", resListStr)
	}

	var jobObj client.Object
	switch jobResource.Kind.Kind {
	case constants.KindRunaiJob:
		jobObj = &runaiJobV1.RunaiJob{}
	case constants.KindPyTorchJob:
		jobObj = &kubeflowv1.PyTorchJob{}
	case constants.KindTFJob:
		jobObj = &kubeflowv1.TFJob{}
	case constants.KindXGBoostJob:
		jobObj = &kubeflowv1.XGBoostJob{}
	case constants.KindMpiJob:
		jobObj = &mpi.MPIJob{}
	case constants.KindJAXJob:
		jobObj = &kubeflowv1.JAXJob{}

	}

	if jobObj.GetObjectKind() != nil {
		krm.getJob(ctx, *jobResource, &jobObj, jobChannel, errChan)
	}
}

func (krm *K8sResourceManager) getJob(ctx context.Context, job common.Resource, targetJobObj *client.Object, jobChannel chan client.Object, errChan chan error) {
	err := krm.client.Get(ctx, client.ObjectKey{Name: job.Name, Namespace: job.Namespace}, *targetJobObj)

	if err != nil {
		errChan <- err
		return
	}
	jobChannel <- *targetJobObj
}

func (krm *K8sResourceManager) getResourcesFromChannels(aggr *resource.Aggregator,
	kindMap *map[string]map[string]common.Resource, jobChannel chan client.Object, serviceChan chan v1.ServiceList,
	ingressListCh chan network.IngressList, routeListCh chan routev1.RouteList,
	httpRouteListCh chan gatewayv1.HTTPRouteList,
	pvcListCh chan v1.PersistentVolumeClaimList, computeDomainCh chan *nvapi.ComputeDomain, secretListCh chan v1.SecretList, errChannel chan error) ([]client.Object, error) {

	if kindMap == nil {
		return nil, fmt.Errorf("kind map is nil")
	}

	var (
		errList      []error
		danglingList []client.Object
	)

	if jobChannel != nil {
		currentJob := <-jobChannel
		aggr.SetJob(currentJob)
	}

	if serviceChan != nil {
		currentService := <-serviceChan
		for _, item := range currentService.Items {
			if _, found := (*kindMap)[constants.KindService][item.Name]; found {
				aggr.AddServices(item.DeepCopy())
			} else {
				danglingList = append(danglingList, item.DeepCopy())
			}
		}
	}

	if ingressListCh != nil {
		currentIngress := <-ingressListCh
		for _, item := range currentIngress.Items {
			if _, found := (*kindMap)[constants.KindIngress][item.Name]; found {
				aggr.AddIngresses(item.DeepCopy())
			} else {
				danglingList = append(danglingList, item.DeepCopy())
			}
		}
	}

	if routeListCh != nil {
		currentRoute := <-routeListCh
		for _, item := range currentRoute.Items {
			if _, found := (*kindMap)[constants.KindRoute][item.Name]; found {
				aggr.AddRoutes(item.DeepCopy())
			} else {
				danglingList = append(danglingList, item.DeepCopy())
			}
		}
	}

	if httpRouteListCh != nil {
		currentHTTPRoute := <-httpRouteListCh
		for _, item := range currentHTTPRoute.Items {
			if _, found := (*kindMap)[constants.KindHTTPRoute][item.Name]; found {
				aggr.AddHTTPRoutes(item.DeepCopy())
			} else {
				danglingList = append(danglingList, item.DeepCopy())
			}
		}
	}

	if pvcListCh != nil {
		currentPvc := <-pvcListCh
		for _, item := range currentPvc.Items {
			aggr.AddPvcs(item.DeepCopy())
		}
	}

	if computeDomainCh != nil {
		computeDomain := <-computeDomainCh
		aggr.SetComputeDomain(computeDomain)
	}

	if secretListCh != nil {
		currentSecret := <-secretListCh
		for _, item := range currentSecret.Items {
			aggr.AddSecrets(item.DeepCopy())
		}
	}

	for err := range errChannel {
		if !apierrors.IsNotFound(err) {
			errList = append(errList, err)
		}
	}
	return danglingList, errors.NewAggregate(errList)
}

func (krm *K8sResourceManager) groupResourceByType(ctx context.Context, objects []common.Resource) map[string]map[string]common.Resource {
	log := logf.FromContext(ctx)
	var (
		kindMap      string
		ok           bool
		getResources = make(map[string]map[string]common.Resource)
	)

	for _, resObj := range objects {
		kindMap = resObj.Kind.Kind
		if _, ok = getResources[kindMap]; !ok {
			switch resObj.Kind.Kind {
			case constants.KindService, constants.KindIngress, constants.KindRoute, constants.KindHTTPRoute, constants.KindPvc, constants.KindComputeDomain, constants.KindSecret:
				getResources[kindMap] = make(map[string]common.Resource)
			case constants.KindRunaiJob, constants.KindPyTorchJob, constants.KindTFJob, constants.KindXGBoostJob, constants.KindMpiJob, constants.KindJAXJob:
				kindMap = constants.JobType
				getResources[kindMap] = make(map[string]common.Resource)
			default:
				log.V(2).Info("resource type was not recognized by resource manager.", "kindType", kindMap)
				continue
			}
		}
		getResources[kindMap][resObj.Name] = resObj
	}
	return getResources
}

func (krm *K8sResourceManager) configureK8sClientOptions(options ResourceOptions) (client.CreateOptions, client.UpdateOptions) {
	createOptions := client.CreateOptions{}
	updateOptions := client.UpdateOptions{}

	if options.DryRun {
		createOptions.DryRun = []string{"All"}
		updateOptions.DryRun = []string{"All"}
	}

	return createOptions, updateOptions
}
func (krm *K8sResourceManager) UpsertResources(ctx context.Context, workload *job.JobWorkload, workloadSubResources []client.Object, options ResourceOptions) []*issue.Issue {
	logger := logf.FromContext(ctx)
	logger.V(4).Info("Updating workload sub-resources")
	createOptions, updateOptions := krm.configureK8sClientOptions(options)
	var wg sync.WaitGroup
	upgradeResponseCh := make(chan string, len(workloadSubResources))
	issuesChan := make(chan *issue.Issue, len(workloadSubResources))

	for _, subResource := range workloadSubResources {
		// this newResource resolve mutual exclusion issue of pointer
		newResource := subResource
		wg.Add(1)
		go ObjectUpsert(ctx, krm.client, krm.apiReadClient, newResource, workload, createOptions, updateOptions, upgradeResponseCh, issuesChan, &wg, krm.rg)
	}
	// update ready condition if some resources were upgraded
	// if no resources were upgraded, the ready condition will set to False
	// with info about the resources that are not uptodate and need to be upgraded
	wg.Wait()
	close(issuesChan)

	var issues []*issue.Issue
	// Collect and report errors
	for iss := range issuesChan {
		issues = append(issues, iss)
	}

	close(upgradeResponseCh)
	if len(issues) > 0 {
		// one of the goroutines failed
		logger.V(2).Info("upsert issue found for workload", "name", workload.Name)
		return issues
	}
	krm.checkUpgradedResource(ctx, upgradeResponseCh, workload)
	logger.V(4).Info("Completed updating  workload sub-resources", "workload", workload.Name)
	return nil
}

func ObjectUpsert(ctx context.Context, client client.Client, apiReadClient client.Reader, newResource client.Object, workload *job.JobWorkload,
	createOptions client.CreateOptions, updateOptions client.UpdateOptions,
	upgradeResponseCh chan<- string, issueChan chan<- *issue.Issue, wg *sync.WaitGroup, rg issue.ResourceGrouper) {
	defer wg.Done()
	logger := logf.FromContext(ctx)
	updateResource := true
	resourceType := resources.GetResourceType(newResource)
	if resourceType == constants.KindPvc || resourceType == constants.KindComputeDomain || resourceType == constants.KindSecret {
		updateResource = false
	}
	upgraded, iss := UpsertIfNeededCheckOwnerUID(ctx, client, apiReadClient, newResource, workload.UID, workload.GetChildName(), createOptions, updateOptions, updateResource, rg)
	if iss != nil {
		logger.Info("Failed to upsert resource", "resource", newResource, "resourceType", resourceType, "error", iss)
		issueChan <- iss
		return
	}
	if upgraded {
		upgradeResponseCh <- fmt.Sprintf("%s/%s", resourceType, newResource.GetName())
	}
	err := status.PatchOwnedResource(newResource, &workload.Status.OwnedResources)
	if err != nil {
		issueChan <- issue.New(workload, err, false, "", rg)
	}
}

func (krm *K8sResourceManager) checkUpgradedResource(ctx context.Context, workloadUpgradedResourcesChannel chan string, workload *job.JobWorkload) {
	var message bytes.Buffer
	log := logf.FromContext(ctx)
	log.V(4).Info("Checking if some resources were upgraded")
	for upgradedResource := range workloadUpgradedResourcesChannel {
		message.WriteString(fmt.Sprintf("%s ", upgradedResource))
	}
	if message.Len() > 0 {
		log.V(4).Info("Some resources were upgraded", "resources", message.String())
		krm.setReadyConditionFalse(workload, message.String())
		return
	}
	log.V(4).Info("No resources were upgraded")
}

func (krm *K8sResourceManager) setReadyConditionFalse(workload *job.JobWorkload, upgradedResources string) {

	readyCondition := status.GetWorkloadCondition(&workload.Status.Conditions, common.Ready)
	var message bytes.Buffer
	if readyCondition != nil && readyCondition.Message != "" {
		message.WriteString(fmt.Sprintf("%s \n", readyCondition.Message))
	}
	message.WriteString("the following workload's resource are stalled and need to be upgraded: ")
	message.WriteString(upgradedResources)
	status.UpsertCondition(&workload.Status.Conditions, common.Ready, v1.ConditionFalse, "", message.String())
}

func (krm *K8sResourceManager) getPvc(ctx context.Context, labelMatch *map[string]string, workloadNS string, waitGroup *sync.WaitGroup, pvcListCh chan v1.PersistentVolumeClaimList, errChannel chan error) {
	defer waitGroup.Done()
	pvcList := &v1.PersistentVolumeClaimList{}
	err := krm.client.List(ctx, pvcList,
		client.InNamespace(workloadNS),
		client.MatchingLabels(*labelMatch))
	if err != nil {
		errChannel <- err
		return
	}
	pvcListCh <- *pvcList
}

func (krm *K8sResourceManager) getSecret(ctx context.Context, labelMatch *map[string]string, workloadNS string, waitGroup *sync.WaitGroup, secretListCh chan v1.SecretList, errChannel chan error) {
	defer waitGroup.Done()
	secretList := &v1.SecretList{}
	err := krm.client.List(ctx, secretList,
		client.InNamespace(workloadNS),
		client.MatchingLabels(*labelMatch))
	if err != nil {
		errChannel <- err
		return
	}
	secretListCh <- *secretList
}

func (krm *K8sResourceManager) getComputeDomain(ctx context.Context, labelMatch *map[string]string, workloadName, workloadNS string, waitGroup *sync.WaitGroup, computeDomainCh chan *nvapi.ComputeDomain, errChannel chan error) {
	defer waitGroup.Done()

	computeDomainList := &nvapi.ComputeDomainList{}
	err := krm.client.List(ctx, computeDomainList,
		client.InNamespace(workloadNS),
		client.MatchingLabels(*labelMatch))
	if err != nil {
		errChannel <- err
		return
	}

	if len(computeDomainList.Items) > 1 {
		errChannel <- fmt.Errorf("found multiple %s for workload %s", constants.KindComputeDomain, workloadName)
		return
	}

	if len(computeDomainList.Items) == 1 {
		computeDomainCh <- &computeDomainList.Items[0]
	} else {
		computeDomainCh <- nil
	}
}
