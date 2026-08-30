package cachedclient

import (
	"context"

	nrtv1alpha2 "github.com/k8stopologyawareschedwg/noderesourcetopology-api/pkg/apis/topology/v1alpha2"
	kaiv1alpha1 "github.com/kai-scheduler/KAI-scheduler/pkg/apis/kai/v1alpha1"
	runaikaiv1alpha1 "github.com/kai-scheduler/kai-resource-management-api/kai/v1alpha1"

	"go.uber.org/zap"

	kaiv2alpha2 "github.com/kai-scheduler/KAI-scheduler/pkg/apis/scheduling/v2alpha2"
	kartav1alpha1 "github.com/run-ai/karta/pkg/api/runai/v1alpha1"
	"github.com/run-ai/runai/runai-cluster/cluster/cluster-sync/pkg/config"
	"github.com/run-ai/runai/runai-cluster/cluster/cluster-sync/pkg/constants"
	runaiv1 "github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1"
	"github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1alpha1"
	dinfwv1 "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/distributedinferenceworkload/v1alpha1"
	dipv2 "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/distributedpolicy/v2alpha1"
	diwv2 "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/distributedworkload/v2alpha1"
	infpv2 "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/inferencepolicy/v2alpha1"
	infwv2 "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/inferenceworkload/v2alpha1"
	ipv2 "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/interactivepolicy/v2alpha1"
	iwv2 "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/interactiveworkload/v2alpha1"
	tpv2 "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/trainingpolicy/v2alpha1"
	twv2 "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/trainingworkload/v2alpha1"
	commonconsts "github.com/run-ai/runai/runai-cluster/common/constants"
	supportedworkloads "github.com/run-ai/runai/runai-cluster/common/supported-workloads"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	schedulingv1 "k8s.io/api/scheduling/v1"
	storagev1 "k8s.io/api/storage/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	apiruntime "k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/rest"
	clientgocache "k8s.io/client-go/tools/cache"
	knv1 "knative.dev/serving/pkg/apis/serving/v1"
	"sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type CachedClient interface {
	// Client - Get and List would be cached, other operations would be done on the writer client
	client.Client

	Start(ctx context.Context, stopCh <-chan struct{})
	Shutdown()

	GetInformer(objectType SupportedType) CachedInformer

	NewRateLimitedController(resourceName RateLimiterName, handler Handler) *RateLimitedController
}

type CachedInformer interface {
	AddEventHandler(handler clientgocache.ResourceEventHandler) (clientgocache.ResourceEventHandlerRegistration, error)
	HasSynced() bool
}

type ClusterSyncCachedClient struct {
	client.Client
	cachedClient           cache.Cache
	scheme                 *apiruntime.Scheme
	objectInformers        map[SupportedType]CachedInformer
	supportedGroupKinds    map[schema.GroupKind]bool
	rateLimitedControllers map[string]*RateLimitedController
}

type SupportedType schema.GroupVersionKind

var (
	NodeType = SupportedType{
		Group:   corev1.SchemeGroupVersion.Group,
		Version: corev1.SchemeGroupVersion.Version,
		Kind:    "Node",
	}
	NodeResourceTopologyType = SupportedType{
		Group:   nrtv1alpha2.SchemeGroupVersion.Group,
		Version: nrtv1alpha2.SchemeGroupVersion.Version,
		Kind:    "NodeResourceTopology",
	}
	NodePoolType = SupportedType{
		Group:   runaikaiv1alpha1.SchemeGroupVersion.Group,
		Version: runaikaiv1alpha1.SchemeGroupVersion.Version,
		Kind:    "NodePool",
	}
	DatavolumeType = SupportedType{
		Group:   v1alpha1.SchemeGroupVersion.Group,
		Version: v1alpha1.SchemeGroupVersion.Version,
		Kind:    "DataVolume",
	}
	AccessRuleType = SupportedType{
		Group:   v1alpha1.SchemeGroupVersion.Group,
		Version: v1alpha1.SchemeGroupVersion.Version,
		Kind:    "AccessRule",
	}
	ProjectType = SupportedType{
		Group:   runaikaiv1alpha1.GroupVersion.Group,
		Version: runaikaiv1alpha1.GroupVersion.Version,
		Kind:    "Project",
	}
	DepartmentType = SupportedType{
		Group:   runaikaiv1alpha1.GroupVersion.Group,
		Version: runaikaiv1alpha1.GroupVersion.Version,
		Kind:    "Department",
	}
	PodType = SupportedType{
		Group:   corev1.SchemeGroupVersion.Group,
		Version: corev1.SchemeGroupVersion.Version,
		Kind:    "Pod",
	}
	PodGroupType = SupportedType{
		Group:   kaiv2alpha2.SchemeGroupVersion.Group,
		Version: kaiv2alpha2.SchemeGroupVersion.Version,
		Kind:    "PodGroup",
	}
	DeploymentType = SupportedType{
		Group:   appsv1.SchemeGroupVersion.Group,
		Version: appsv1.SchemeGroupVersion.Version,
		Kind:    "Deployment",
	}
	ServiceType = SupportedType{
		Group:   knv1.SchemeGroupVersion.Group,
		Version: knv1.SchemeGroupVersion.Version,
		Kind:    "Service",
	}
	PersistentVolumeClaimType = SupportedType{
		Group:   corev1.SchemeGroupVersion.Group,
		Version: corev1.SchemeGroupVersion.Version,
		Kind:    "PersistentVolumeClaim",
	}
	SecretType = SupportedType{
		Group:   corev1.SchemeGroupVersion.Group,
		Version: corev1.SchemeGroupVersion.Version,
		Kind:    "Secret",
	}
	DistributedWorkloadType = SupportedType{
		Group:   diwv2.GroupVersion.Group,
		Version: diwv2.GroupVersion.Version,
		Kind:    "DistributedWorkload",
	}
	TrainingWorkloadType = SupportedType{
		Group:   twv2.GroupVersion.Group,
		Version: twv2.GroupVersion.Version,
		Kind:    "TrainingWorkload",
	}
	InferenceWorkloadType = SupportedType{
		Group:   infwv2.GroupVersion.Group,
		Version: infwv2.GroupVersion.Version,
		Kind:    "InferenceWorkload",
	}
	DistributedInferenceWorkloadType = SupportedType{
		Group:   dinfwv1.GroupVersion.Group,
		Version: dinfwv1.GroupVersion.Version,
		Kind:    "DistributedInferenceWorkload",
	}
	RevisionType = SupportedType{
		Group:   knv1.SchemeGroupVersion.Group,
		Version: knv1.SchemeGroupVersion.Version,
		Kind:    "Revision",
	}
	InteractiveWorkloadType = SupportedType{
		Group:   iwv2.GroupVersion.Group,
		Version: iwv2.GroupVersion.Version,
		Kind:    "InteractiveWorkload",
	}
	StorageClassType = SupportedType{
		Group:   storagev1.SchemeGroupVersion.Group,
		Version: storagev1.SchemeGroupVersion.Version,
		Kind:    "StorageClass",
	}
	ExternalWorkloadType = SupportedType{
		Group:   v1alpha1.SchemeGroupVersion.Group,
		Version: v1alpha1.SchemeGroupVersion.Version,
		Kind:    "ExternalWorkload",
	}
	InteractivePolicyType = SupportedType{
		Group:   ipv2.GroupVersion.Group,
		Version: ipv2.GroupVersion.Version,
		Kind:    "InteractivePolicy",
	}
	TrainingPolicyType = SupportedType{
		Group:   tpv2.GroupVersion.Group,
		Version: tpv2.GroupVersion.Version,
		Kind:    "TrainingPolicy",
	}
	InferencePolicyType = SupportedType{
		Group:   infpv2.GroupVersion.Group,
		Version: infpv2.GroupVersion.Version,
		Kind:    "InferencePolicy",
	}
	DistributedPolicyType = SupportedType{
		Group:   dipv2.GroupVersion.Group,
		Version: dipv2.GroupVersion.Version,
		Kind:    "DistributedPolicy",
	}
	ConfigMapType = SupportedType{
		Group:   corev1.SchemeGroupVersion.Group,
		Version: corev1.SchemeGroupVersion.Version,
		Kind:    "ConfigMap",
	}
	PriorityClassType = SupportedType{
		Group:   schedulingv1.SchemeGroupVersion.Group,
		Version: schedulingv1.SchemeGroupVersion.Version,
		Kind:    "PriorityClass",
	}
	TopologyType = SupportedType{
		Group:   kaiv1alpha1.SchemeGroupVersion.Group,
		Version: kaiv1alpha1.SchemeGroupVersion.Version,
		Kind:    "Topology",
	}
	ApplicationType = SupportedType{
		Group:   v1alpha1.SchemeGroupVersion.Group,
		Version: v1alpha1.SchemeGroupVersion.Version,
		Kind:    "Application",
	}
	KartaType = SupportedType{
		Group:   kartav1alpha1.SchemeGroupVersion.Group,
		Version: kartav1alpha1.SchemeGroupVersion.Version,
		Kind:    "Karta",
	}
	WorkloadRunnerType = SupportedType{
		Group:   v1alpha1.SchemeGroupVersion.Group,
		Version: v1alpha1.SchemeGroupVersion.Version,
		Kind:    "WorkloadRunner",
	}
	AllSupportedTypes = []SupportedType{
		NodeType,
		NodeResourceTopologyType,
		NodePoolType,
		DatavolumeType,
		AccessRuleType,
		ProjectType,
		DepartmentType,
		PodType,
		PodGroupType,
		DeploymentType,
		ServiceType,
		PersistentVolumeClaimType,
		SecretType,
		DistributedWorkloadType,
		TrainingWorkloadType,
		InferenceWorkloadType,
		DistributedInferenceWorkloadType,
		RevisionType,
		InteractiveWorkloadType,
		StorageClassType,
		ExternalWorkloadType,
		InteractivePolicyType,
		TrainingPolicyType,
		InferencePolicyType,
		DistributedPolicyType,
		ConfigMapType,
		PriorityClassType,
		TopologyType,
		ApplicationType,
		KartaType,
		WorkloadRunnerType,
	}
)

func (t SupportedType) String() string {
	gvk := schema.GroupVersionKind(t)
	return gvk.String()
}

func NewCachedClient(ctx context.Context, writerClient client.Client,
	kubeConfig *rest.Config, scheme *apiruntime.Scheme) CachedClient {
	cachedClient := createCache(kubeConfig, scheme)

	clusterSyncCachedClient := &ClusterSyncCachedClient{
		Client:                 writerClient,
		cachedClient:           cachedClient,
		scheme:                 scheme,
		objectInformers:        map[SupportedType]CachedInformer{},
		supportedGroupKinds:    getInitialSupportedGroupKinds(),
		rateLimitedControllers: map[string]*RateLimitedController{},
	}

	clusterSyncCachedClient.setInformers(ctx, supportedTypes())
	return clusterSyncCachedClient
}

func supportedTypes() []SupportedType {
	return append([]SupportedType{}, AllSupportedTypes...)
}

func (c *ClusterSyncCachedClient) setInformers(ctx context.Context, supportedTypes []SupportedType) {
	for _, gvk := range supportedTypes {
		object := createObject(gvk, c.scheme)
		objectInformer, err := c.cachedClient.GetInformer(ctx, object)
		if err != nil {
			if shouldSkipInformerError(gvk, err) {
				zap.S().Warnf("Optional type <%s> is not installed; its informer will not be started: %s",
					gvk.String(), err.Error())
				continue
			}

			zap.S().Fatalf("Failed to create informer for type <%s>, error: %s",
				gvk.String(), err.Error())
		}

		c.objectInformers[gvk] = objectInformer
		c.supportedGroupKinds[schema.GroupVersionKind(gvk).GroupKind()] = true
	}
}

func shouldSkipInformerError(gvk SupportedType, err error) bool {
	return isOptionalInformer(gvk) && meta.IsNoMatchError(err)
}

func isOptionalInformer(gvk SupportedType) bool {
	return gvk == ServiceType ||
		gvk == RevisionType ||
		gvk == NodeResourceTopologyType ||
		isLegacyWorkloadType(gvk) ||
		isLegacyPolicyType(gvk)
}

func isLegacyWorkloadType(gvk SupportedType) bool {
	return gvk == TrainingWorkloadType ||
		gvk == InteractiveWorkloadType ||
		gvk == DistributedWorkloadType ||
		gvk == InferenceWorkloadType ||
		gvk == DistributedInferenceWorkloadType
}

func isLegacyPolicyType(gvk SupportedType) bool {
	return gvk == TrainingPolicyType ||
		gvk == InteractivePolicyType ||
		gvk == DistributedPolicyType ||
		gvk == InferencePolicyType
}

func getInitialSupportedGroupKinds() map[schema.GroupKind]bool {
	allTypesThatNeedSupport := supportedworkloads.GetAllTypesThatNeedSupport()

	// kinds that are not part of "GetAllTypesThatNeedSupport", but we do have permissions and caches for them.
	supportedGroupKinds := map[schema.GroupKind]bool{
		{Group: kaiv2alpha2.GroupVersion.Group, Kind: "PodGroup"}: true,
	}

	runaiJobGroupKind := schema.GroupKind{Group: runaiv1.GroupVersion.Group, Kind: "RunaiJob"}
	for gvk := range allTypesThatNeedSupport {
		groupKind := schema.GroupKind{Group: gvk.Group, Kind: gvk.Kind}
		if groupKind != runaiJobGroupKind {
			supportedGroupKinds[groupKind] = true
		}
	}

	return supportedGroupKinds
}

func createObject(gvk SupportedType, scheme *apiruntime.Scheme) client.Object {
	runtimeObject, err := scheme.New(schema.GroupVersionKind(gvk))
	if err != nil {
		zap.S().Fatalf("Failed creating new runtime object from gvk <%s>, error: %s",
			gvk.String(), err.Error())
		return nil
	}

	object, ok := runtimeObject.(client.Object)
	if !ok {
		zap.S().Fatalf("Failed to parse runtimeObject of kind <%s> to client.Object", gvk.String())
		return nil
	}
	return object
}

func createCache(kubeConfig *rest.Config, scheme *apiruntime.Scheme) cache.Cache {
	cacheOptions := cache.Options{
		Scheme: scheme,
	}

	cachedClient, err := cache.New(kubeConfig, cacheOptions)
	if err != nil {
		zap.S().Fatalf("Failed to create cached client, error: %s", err.Error())
	}

	addIndexers(cachedClient)

	return cachedClient
}

// Get is a cached operation by default.
// There's a bug in k8s cached client, where "Get" for objects we don't have the
// RBAC for are left waiting forever. So this is a wrapper to avoid those issues.
// https://github.com/kubernetes-sigs/controller-runtime/issues/550
func (c *ClusterSyncCachedClient) Get(ctx context.Context, key client.ObjectKey, obj client.Object,
	opts ...client.GetOption) error {
	gvk := obj.GetObjectKind().GroupVersionKind()
	if key.Namespace == "" {
		return c.cachedClient.Get(ctx, key, obj, opts...)
	}

	groupKind := gvk.GroupKind()

	if c.supportedGroupKinds[groupKind] || groupKind.Kind == "" {
		// when we are getting an object explicitly, its gvk is not set,
		// and it's fine, we should have permissions for it.
		return c.cachedClient.Get(ctx, key, obj, opts...)
	}

	// if it's an unrecognized type - used Uncached client to Get it to avoid the k8s bug
	return c.Client.Get(ctx, key, obj, opts...)
}

// List is a cached operation
func (c *ClusterSyncCachedClient) List(ctx context.Context, list client.ObjectList,
	opts ...client.ListOption) error {
	if workloadType, found := legacyWorkloadTypeForList(list); found && c.GetInformer(workloadType) == nil {
		clearLegacyWorkloadList(list)
		return nil
	}
	return c.cachedClient.List(ctx, list, opts...)
}

func legacyWorkloadTypeForList(list client.ObjectList) (SupportedType, bool) {
	switch list.(type) {
	case *twv2.TrainingWorkloadList:
		return TrainingWorkloadType, true
	case *iwv2.InteractiveWorkloadList:
		return InteractiveWorkloadType, true
	case *diwv2.DistributedWorkloadList:
		return DistributedWorkloadType, true
	case *infwv2.InferenceWorkloadList:
		return InferenceWorkloadType, true
	case *dinfwv1.DistributedInferenceWorkloadList:
		return DistributedInferenceWorkloadType, true
	default:
		return SupportedType{}, false
	}
}

func clearLegacyWorkloadList(list client.ObjectList) {
	switch typed := list.(type) {
	case *twv2.TrainingWorkloadList:
		typed.Items = nil
	case *iwv2.InteractiveWorkloadList:
		typed.Items = nil
	case *diwv2.DistributedWorkloadList:
		typed.Items = nil
	case *infwv2.InferenceWorkloadList:
		typed.Items = nil
	case *dinfwv1.DistributedInferenceWorkloadList:
		typed.Items = nil
	}
}

func (c *ClusterSyncCachedClient) Start(ctx context.Context, stopCh <-chan struct{}) {
	go func() {
		if err := c.cachedClient.Start(ctx); err != nil {
			zap.S().Fatalf("Failed to start cache client, error: %s", err.Error())
		}
	}()

	if !c.cachedClient.WaitForCacheSync(ctx) {
		zap.S().Fatal("failed waiting for cache to sync")
	}

	for gvk, objectInformer := range c.objectInformers {
		if objectInformer != nil {
			if !clientgocache.WaitForNamedCacheSync(gvk.String(), stopCh, objectInformer.HasSynced) {
				zap.S().Fatalf("failed waiting for %s cache to sync", gvk.String())
			}
		}
	}

	c.startRateLimitedControllers()
}

func (c *ClusterSyncCachedClient) GetInformer(objectType SupportedType) CachedInformer {
	objectInformer, found := c.objectInformers[objectType]
	if found && objectInformer != nil {
		return objectInformer
	}
	return nil
}

// startRateLimitedControllers - should be called after the cache is synced
// and the informers are running
func (c *ClusterSyncCachedClient) startRateLimitedControllers() {
	for _, rateLimitedController := range c.rateLimitedControllers {
		rateLimitedController.Run()
	}
}

func (c *ClusterSyncCachedClient) NewRateLimitedController(resourceName RateLimiterName, handler Handler) *RateLimitedController {
	rateLimitedController := NewRateLimitedController(resourceName, handler)
	c.rateLimitedControllers[string(resourceName)] = rateLimitedController
	return rateLimitedController
}

func (c *ClusterSyncCachedClient) Shutdown() {
	for _, rateLimitedController := range c.rateLimitedControllers {
		rateLimitedController.ShutDown()
	}
}

func addIndexers(cachedClient cache.Cache) {
	addPodIndexers(cachedClient)
	addRunaiWorkloadIndexers(cachedClient)
}

func addPodIndexers(cachedClient cache.Cache) {
	// TODO: remove when podgroups deprecated
	err := cachedClient.IndexField(context.Background(), &corev1.Pod{},
		constants.PodByPodGroupIndexerName, PodByPodGroupIndexer)
	if err != nil {
		zap.S().Fatalf("Failed indexing pod field: %s, err: %s",
			constants.PodByPodGroupIndexerName, err.Error())
	}

	if config.Config().EnableNodeNvidiaComponentsReadiness {
		err = cachedClient.IndexField(context.Background(), &corev1.Pod{},
			constants.NonRunaiPodByNodeNameIndexer, NonRunaiPodByNodeNameIndexer)
		if err != nil {
			zap.S().Fatalf("Failed indexing pod field: %s, err: %s",
				constants.NonRunaiPodByNodeNameIndexer, err.Error())
		}
	}

	err = cachedClient.IndexField(context.Background(), &corev1.Pod{},
		constants.PodBySchedulerNameIndexer, PodBySchedulerNameIndexer)
	if err != nil {
		zap.S().Fatalf("Failed indexing pod field: %s, err: %s",
			constants.PodBySchedulerNameIndexer, err.Error())
	}
}

func PodByPodGroupIndexer(obj client.Object) []string {
	pod, ok := obj.(*corev1.Pod)
	if !ok {
		zap.S().Errorf("Failed converting object to pod in PodByPodGroupIndexer: %v", obj)
		return []string{""}
	}

	podGroupName, found := pod.Annotations[commonconsts.PodGroupAnnotationForPod]
	if !found {
		return []string{""}
	}
	return []string{podGroupName}
}

func NonRunaiPodByNodeNameIndexer(obj client.Object) []string {
	pod, ok := obj.(*corev1.Pod)
	if !ok {
		zap.S().Errorf("Failed converting object to pod in NonRunaiPodByNodeNameIndexer: %v", obj)
		return []string{""}
	}

	if pod.Spec.SchedulerName != "default-scheduler" ||
		pod.Status.Phase != corev1.PodRunning ||
		pod.Spec.NodeName == "" {
		return []string{""}
	}

	return []string{pod.Spec.NodeName}
}

func PodBySchedulerNameIndexer(obj client.Object) []string {
	pod, ok := obj.(*corev1.Pod)
	if !ok {
		zap.S().Errorf("Failed converting object to pod in PodBySchedulerNameIndexer: %v", obj)
		return []string{""}
	}
	return []string{pod.Spec.SchedulerName}
}

func addRunaiWorkloadIndexers(cachedClient cache.Cache) {
	workloadTypesToCache := []client.Object{
		&twv2.TrainingWorkload{},
		&iwv2.InteractiveWorkload{},
		&infwv2.InferenceWorkload{},
		&diwv2.DistributedWorkload{},
		&dinfwv1.DistributedInferenceWorkload{},
	}

	for _, workloadTypeToCache := range workloadTypesToCache {
		if _, err := cachedClient.GetInformer(context.Background(), workloadTypeToCache); err != nil {
			zap.S().Warnf("Optional legacy workload type <%T> is not installed; its index will not be registered: %s",
				workloadTypeToCache, err.Error())
			continue
		}
		err := cachedClient.IndexField(context.Background(), workloadTypeToCache,
			constants.RunaiWorkloadByUsageIndexer, RunaiWorkloadByUsageIndexer)
		if err != nil {
			zap.S().Fatalf("Failed indexing workload field: %s, err: %s",
				constants.RunaiWorkloadByUsageIndexer, err.Error())
		}
	}
}

func RunaiWorkloadByUsageIndexer(obj client.Object) []string {
	switch workload := obj.(type) {
	case *twv2.TrainingWorkload:
		return []string{workload.Spec.Usage}
	case *iwv2.InteractiveWorkload:
		return []string{workload.Spec.Usage}
	case *infwv2.InferenceWorkload:
		return []string{workload.Spec.Usage}
	case *diwv2.DistributedWorkload:
		return []string{workload.Spec.Usage}
	case *dinfwv1.DistributedInferenceWorkload:
		return []string{workload.Spec.Usage}
	}
	return []string{""}
}
