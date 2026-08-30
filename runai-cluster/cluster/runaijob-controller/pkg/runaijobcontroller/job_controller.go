/*
Copyright 2015 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package runaijobcontroller

import (
	"context"
	"fmt"
	"math"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/run-ai/runai/runai-cluster/cluster/runaijob-controller/pkg/common"
	"github.com/run-ai/runai/runai-cluster/cluster/runaijob-controller/pkg/common/runaiJobUtils"
	"k8s.io/client-go/rest"
	"k8s.io/utils/pointer"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	core1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	kubeinformers "k8s.io/client-go/informers"
	clientset "k8s.io/client-go/kubernetes"
	corev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	corelisters "k8s.io/client-go/listers/core/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/util/workqueue"
	"k8s.io/klog/v2"
	"k8s.io/kubernetes/pkg/controller"
	"k8s.io/utils/integer"

	runaijob "github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1"
	clientsetrunai "github.com/run-ai/runai/runai-cluster/cluster/sdk/client/clientset/versioned"
	"github.com/run-ai/runai/runai-cluster/cluster/sdk/client/clientset/versioned/scheme"
	runaijobinformers "github.com/run-ai/runai/runai-cluster/cluster/sdk/client/informers/externalversions"
	runaijoblisters "github.com/run-ai/runai/runai-cluster/cluster/sdk/client/listers/run/v1"
)

const (
	labelOfPodIndex         = "runai/pod-index"
	kubeClientUpdateRetries = 3

	GPUResourceName         = "nvidia.com/gpu"
	GPUFractionResourceName = "gpu-fraction"

	failureMessageSwapPreempted = "run.ai/preempted"

	defaultOfTerminateAfterPreemption = false

	DefaultCompletions  = 1
	DefaultParallelism  = 1
	DefaultBackoffLimit = 0
)

type podStatus string

const (
	phaseDeleted podStatus = "Deleted"
	// nolint
	phasePending podStatus = "Pending"
	// nolint
	phaseFailed podStatus = "Failed"
	// nolint
	phaseSucceeded podStatus = "Succeeded"
	// nolint
	phaseSuspended podStatus = "Suspended"

	phaseCreating podStatus = "Creating"
	phaseDeleting podStatus = "Deleting"

	jobSuspendedMessage = "Job suspended"
	jobResumedMessage   = "Job resumed"
)

// resourcesTypes maps the user readable resource name to it ResourceName type value
var resourcesTypes = map[string]v1.ResourceName{
	"GPU": GPUResourceName,
	"CPU": v1.ResourceCPU,
	"MEM": v1.ResourceMemory,
}

type podLastIndexAndStatus struct {
	lastIndexAttempt int64
	status           podStatus
}

// controllerKind contains the schema.GroupVersionKind for this controller type.
var controllerKind = runaijob.SchemeGroupVersion.WithKind("RunaiJob")

var (
	// DefaultJobBackOff is the default backoff period, exported for the e2e test
	DefaultJobBackOff = 10 * time.Second
	// MaxJobBackOff is the max backoff period, exported for the e2e test
	MaxJobBackOff = 360 * time.Second
)

// Controller ensures that all RunaiJob objects have corresponding pods to
// run their configured workload.
type Controller struct {
	config          *rest.Config
	kubeRunaiClient clientsetrunai.Interface
	kubeClient      clientset.Interface
	podControl      controller.PodControlInterface
	control         PodControlInterface

	// To allow injection of updateJobStatus for testing.
	updateHandler func(job *runaijob.RunaiJob) error
	syncHandler   func(jobKey ctrl.Request) (bool, error)
	// podStoreSynced returns true if the pod store has been synced at least once.
	// Added as a member to the struct to allow injection for testing.
	podStoreSynced cache.InformerSynced
	// jobStoreSynced returns true if the job store has been synced at least once.
	// Added as a member to the struct to allow injection for testing.
	jobStoreSynced  cache.InformerSynced
	configMapSynced cache.InformerSynced

	// A TTLCache of pod creates/deletes each rc expects to see
	expectations controller.ControllerExpectationsInterface

	// A store of jobs
	// jobLister runaijoblisters.RunaiJobLister
	jobLister runaijoblisters.RunaiJobLister

	// A store of pods, populated by the podController
	podStore       corelisters.PodLister
	configMapStore corelisters.ConfigMapLister

	// Jobs that need to be updated
	queue workqueue.TypedRateLimitingInterface[ctrl.Request]

	recorder record.EventRecorder
}

// NewController creates a new RunaiJob controller that keeps the relevant pods
// in sync with their corresponding RunaiJob objects.
func NewController(config *rest.Config, podInformer kubeinformers.SharedInformerFactory, jobInformer runaijobinformers.SharedInformerFactory,
	kubeClient clientset.Interface, kubeRunaiClient clientsetrunai.Interface) *Controller {

	broadcaster := record.NewBroadcaster()
	broadcaster.StartRecordingToSink(&corev1.EventSinkImpl{Interface: kubeClient.CoreV1().Events("")})
	jm := &Controller{
		config:          config,
		kubeRunaiClient: kubeRunaiClient,
		kubeClient:      kubeClient,
		podControl: controller.RealPodControl{
			KubeClient: kubeClient,
			Recorder:   broadcaster.NewRecorder(scheme.Scheme, v1.EventSource{Component: "runaijobcontroller"}),
		},
		control:      NewPodControl(kubeClient, broadcaster.NewRecorder(scheme.Scheme, v1.EventSource{Component: "runaijobcontroller"})),
		expectations: controller.NewControllerExpectations(),
		queue:        workqueue.NewTypedRateLimitingQueue(workqueue.NewTypedItemExponentialFailureRateLimiter[ctrl.Request](DefaultJobBackOff, MaxJobBackOff)),
		recorder:     broadcaster.NewRecorder(scheme.Scheme, v1.EventSource{Component: "runaijobcontroller"}),
	}

	jobInformer.Run().V1().RunaiJobs().Informer().AddEventHandler(
		cache.ResourceEventHandlerFuncs{
			AddFunc:    jm.addJob,
			UpdateFunc: jm.updateJob,
			DeleteFunc: func(obj interface{}) {
				job := obj.(*runaijob.RunaiJob)
				jm.enqueueController(job, true)
			},
		})

	jm.jobLister = jobInformer.Run().V1().RunaiJobs().Lister()
	jm.jobStoreSynced = jobInformer.Run().V1().RunaiJobs().Informer().HasSynced

	podInformer.Core().V1().Pods().Informer().AddEventHandler(
		cache.ResourceEventHandlerFuncs{
			AddFunc:    jm.addPod,
			UpdateFunc: jm.updatePod,
			DeleteFunc: jm.deletePod,
		})

	jm.podStore = podInformer.Core().V1().Pods().Lister()
	jm.podStoreSynced = podInformer.Core().V1().Pods().Informer().HasSynced

	jm.configMapStore = podInformer.Core().V1().ConfigMaps().Lister()
	jm.configMapSynced = podInformer.Core().V1().ConfigMaps().Informer().HasSynced

	jm.updateHandler = jm.updateJobStatus
	jm.syncHandler = jm.syncJob

	return jm
}

// Run the main goroutine responsible for watching and syncing jobs.
func (jm *Controller) Run(workers int, stopCh <-chan struct{}) {
	defer utilruntime.HandleCrash()
	defer jm.queue.ShutDown()

	klog.Infof("Starting runaijob controller")
	defer klog.Infof("Shutting down runaijob controller")

	if !cache.WaitForNamedCacheSync("runaijob", stopCh, jm.podStoreSynced, jm.jobStoreSynced, jm.configMapSynced) {
		return
	}

	for i := 0; i < workers; i++ {
		go wait.Until(jm.worker, time.Second, stopCh)
	}

	<-stopCh
}

func listPodJobs(jobLister runaijoblisters.RunaiJobLister, pod *v1.Pod) (jobs []runaijob.RunaiJob, err error) {
	if len(pod.Labels) == 0 {
		err = fmt.Errorf("no jobs found for pod %v because it has no labels", pod.Name)
		return
	}

	var list []*runaijob.RunaiJob
	list, err = jobLister.RunaiJobs(pod.Namespace).List(labels.Everything())
	if err != nil {
		return
	}
	for _, job := range list {
		selector, _ := metav1.LabelSelectorAsSelector(job.Spec.Selector)
		if !selector.Matches(labels.Set(pod.Labels)) {
			continue
		}
		jobs = append(jobs, *job)
	}
	if len(jobs) == 0 {
		err = fmt.Errorf("could not find jobs for pod %s in namespace %s with labels: %v", pod.Name, pod.Namespace, pod.Labels)
	}
	return
}

// getPodJobs returns a list of Jobs that potentially match a Pod.
func (jm *Controller) getPodJobs(pod *v1.Pod) []*runaijob.RunaiJob {
	jobs, err := listPodJobs(jm.jobLister, pod)
	if err != nil {
		return nil
	}
	if len(jobs) > 1 {
		// ControllerRef will ensure we don't process multiple owners, but more than one
		// item in this list nevertheless constitutes user error.
		utilruntime.HandleError(fmt.Errorf("user error! more than one job is selecting pods with labels: %+v", pod.Labels))
	}
	ret := make([]*runaijob.RunaiJob, 0, len(jobs))
	for i := range jobs {
		ret = append(ret, &jobs[i])
	}
	return ret
}

// resolveControllerRef returns the controller referenced by a ControllerRef,
// or nil if the ControllerRef could not be resolved to a matching controller
// of the correct Kind.
func (jm *Controller) resolveControllerRef(namespace string, controllerRef *metav1.OwnerReference) *runaijob.RunaiJob {
	// We can't look up by UID, so look up by Name and then verify UID.
	// Don't even try to look up by Name if it's the wrong Kind.
	if controllerRef.Kind != controllerKind.Kind {
		return nil
	}
	job, err := jm.jobLister.RunaiJobs(namespace).Get(controllerRef.Name)
	if err != nil {
		return nil
	}
	if job.UID != controllerRef.UID {
		// The controller we found with this Name is not the same one that the
		// ControllerRef points to.
		return nil
	}
	return job
}

// When a pod is created, enqueue the controller that manages it and update it's expectations.
func (jm *Controller) addPod(obj interface{}) {
	pod := obj.(*v1.Pod)
	if pod.DeletionTimestamp != nil {
		// on a restart of the controller, it's possible a new pod shows up in a state that
		// is already pending deletion. Prevent the pod from being a creation observation.
		jm.deletePod(pod)
		return
	}

	// If it has a ControllerRef, that's all that matters.
	if controllerRef := metav1.GetControllerOf(pod); controllerRef != nil {
		job := jm.resolveControllerRef(pod.Namespace, controllerRef)
		if job == nil {
			return
		}
		jobKey, err := controller.KeyFunc(job)
		if err != nil {
			return
		}
		jm.expectations.CreationObserved(klog.FromContext(context.Background()), jobKey)
		jm.enqueueController(job, true)
		return
	}

	// Otherwise, it's an orphan. Get a list of all matching controllers and sync
	// them to see if anyone wants to adopt it.
	// DO NOT observe creation because no controller should be waiting for an
	// orphan.
	for _, job := range jm.getPodJobs(pod) {
		jm.enqueueController(job, true)
	}
}

// When a pod is updated, figure out what job/s manage it and wake them up.
// If the labels of the pod have changed we need to awaken both the old
// and new job. old and cur must be *v1.Pod types.
func (jm *Controller) updatePod(old, cur interface{}) {
	curPod := cur.(*v1.Pod)
	oldPod := old.(*v1.Pod)
	if curPod.ResourceVersion == oldPod.ResourceVersion {
		// Periodic resync will send update events for all known pods.
		// Two different versions of the same pod will always have different RVs.
		return
	}
	if curPod.DeletionTimestamp != nil {
		// when a pod is deleted gracefully it's deletion timestamp is first modified to reflect a grace period,
		// and after such time has passed, the kubelet actually deletes it from the store. We receive an update
		// for modification of the deletion timestamp and expect an job to create more pods asap, not wait
		// until the kubelet actually deletes the pod.
		jm.deletePod(curPod)
		return
	}

	// the only time we want the backoff to kick-in, is when the pod failed
	immediate := curPod.Status.Phase != v1.PodFailed

	curControllerRef := metav1.GetControllerOf(curPod)
	oldControllerRef := metav1.GetControllerOf(oldPod)
	controllerRefChanged := !reflect.DeepEqual(curControllerRef, oldControllerRef)
	if controllerRefChanged && oldControllerRef != nil {
		// The ControllerRef was changed. Sync the old controller, if any.
		if job := jm.resolveControllerRef(oldPod.Namespace, oldControllerRef); job != nil {
			jm.enqueueController(job, immediate)
		}
	}

	// If it has a ControllerRef, that's all that matters.
	if curControllerRef != nil {
		job := jm.resolveControllerRef(curPod.Namespace, curControllerRef)
		if job == nil {
			return
		}
		jm.enqueueController(job, immediate)
		return
	}

	// Otherwise, it's an orphan. If anything changed, sync matching controllers
	// to see if anyone wants to adopt it now.
	labelChanged := !reflect.DeepEqual(curPod.Labels, oldPod.Labels)
	if labelChanged || controllerRefChanged {
		for _, job := range jm.getPodJobs(curPod) {
			jm.enqueueController(job, immediate)
		}
	}
}

// When a pod is deleted, enqueue the job that manages the pod and update its expectations.
// obj could be an *v1.Pod, or a DeletionFinalStateUnknown marker item.
func (jm *Controller) deletePod(obj interface{}) {
	pod, ok := obj.(*v1.Pod)

	// When a delete is dropped, the relist will notice a pod in the store not
	// in the list, leading to the insertion of a tombstone object which contains
	// the deleted key/value. Note that this value might be stale. If the pod
	// changed labels the new job will not be woken up till the periodic resync.
	if !ok {
		tombstone, ok := obj.(cache.DeletedFinalStateUnknown)
		if !ok {
			utilruntime.HandleError(fmt.Errorf("couldn't get object from tombstone %+v", obj))
			return
		}
		pod, ok = tombstone.Obj.(*v1.Pod)
		if !ok {
			utilruntime.HandleError(fmt.Errorf("tombstone contained object that is not a pod %+v", obj))
			return
		}
	}

	controllerRef := metav1.GetControllerOf(pod)
	if controllerRef == nil {
		// No controller should care about orphans being deleted.
		return
	}
	job := jm.resolveControllerRef(pod.Namespace, controllerRef)
	if job == nil {
		return
	}
	jobKey, err := controller.KeyFunc(job)
	if err != nil {
		return
	}
	jm.expectations.DeletionObserved(klog.FromContext(context.Background()), jobKey)
	jm.enqueueController(job, true)
}

func (jm *Controller) addJob(obj interface{}) {
	job := obj.(*runaijob.RunaiJob)
	jm.enqueueController(job, true)
}

func (jm *Controller) updateJob(old, cur interface{}) {
	curJob := cur.(*runaijob.RunaiJob)

	jm.enqueueController(curJob, true)
}

// obj could be an *batch.RunaiJob, or a DeletionFinalStateUnknown marker item,
// immediate tells the controller to update the status right away, and should
// happen ONLY when there was a successful pod run.
func (jm *Controller) enqueueController(obj *runaijob.RunaiJob, immediate bool) {
	key := client.ObjectKeyFromObject(obj)
	req := ctrl.Request{NamespacedName: key}

	backoff := time.Duration(0)
	if !immediate {
		backoff = getBackoff(jm.queue, req)
	}

	// TODO: Handle overlapping controllers better. Either disallow them at admission time or
	// deterministically avoid syncing controllers that fight over pods. Currently, we only
	// ensure that the same controller is synced for a given pod. When we periodically relist
	// all controllers there will still be some replica instability. One way to handle this is
	// by querying the store for all controllers that this rc overlaps, as well as all
	// controllers that overlap this rc, and sorting them.
	jm.queue.AddAfter(req, backoff)
}

// worker runs a worker thread that just dequeues items, processes them, and marks them done.
// It enforces that the syncHandler is never invoked concurrently with the same key.
func (jm *Controller) worker() {
	for jm.processNextWorkItem() {
	}
}

func (jm *Controller) processNextWorkItem() bool {
	key, quit := jm.queue.Get()
	if quit {
		return false
	}
	defer jm.queue.Done(key)

	forget, err := jm.syncHandler(key)
	if err == nil {
		if forget {
			jm.queue.Forget(key)
		}
		return true
	}

	utilruntime.HandleError(fmt.Errorf("Error syncing job: %v", err))
	jm.queue.AddRateLimited(key)

	return true
}

// getPodsForJob returns the set of pods that this RunaiJob should manage.
// It also reconciles ControllerRef by adopting/orphaning.
// Note that the returned Pods are pointers into the cache.
func (jm *Controller) getPodsForJob(j *runaijob.RunaiJob) ([]*v1.Pod, error) {
	selector, err := metav1.LabelSelectorAsSelector(&metav1.LabelSelector{
		MatchLabels: map[string]string{
			"controller-uid": string(j.UID),
		},
	})
	if err != nil {
		return nil, fmt.Errorf("couldn't convert RunaiJob selector: %v", err)
	}
	// List all pods to include those that don't match the selector anymore
	// but have a ControllerRef pointing to this controller.
	pods, err := jm.podStore.Pods(j.Namespace).List(labels.Everything())
	if err != nil {
		return nil, err
	}
	// If any adoptions are attempted, we should first recheck for deletion
	// with an uncached quorum read sometime after listing Pods (see #42639).
	canAdoptFunc := controller.RecheckDeletionTimestamp(func(c context.Context) (metav1.Object, error) {
		fresh, err := jm.kubeRunaiClient.RunV1().RunaiJobs(j.Namespace).Get(c, j.Name, metav1.GetOptions{})
		if err != nil {
			return nil, err
		}
		if fresh.UID != j.UID {
			return nil, fmt.Errorf("original RunaiJob %v/%v is gone: got uid %v, wanted %v", j.Namespace, j.Name, fresh.UID, j.UID)
		}
		return fresh, nil
	})
	cm := controller.NewPodControllerRefManager(jm.podControl, j, selector, controllerKind, canAdoptFunc)
	return cm.ClaimPods(context.TODO(), pods)
}

func isContainerConfigMisMatching(c *v1.Container, job *runaijob.RunaiJob, i int, podName string) bool {
	for resourceName, resourceType := range resourcesTypes {
		podResourceVal, podResourceFound := c.Resources.Limits[resourceType]
		jobResourceVal, jobResourceFound := job.Spec.Template.Spec.Containers[i].Resources.Limits[resourceType]
		if jobResourceFound && (podResourceFound != jobResourceFound || podResourceVal.Value() != jobResourceVal.Value()) {
			klog.V(4).Infof("Old pod: %s of job: <%v/%v>. Different %s values. podResourceFound: %t, "+
				"jobResourceFound: %t, podResourceVal: %v, jobResourceVal: %v", podName, job.Namespace, job.Name, resourceName, podResourceFound,
				jobResourceFound, podResourceVal, jobResourceVal)
			return true
		}
	}
	return false
}

func getPodsNotMatchingSpec(job *runaijob.RunaiJob, activePods []*v1.Pod) []*v1.Pod {
	var ret []*v1.Pod
	for _, pod := range activePods {
		// currently gpu-fraction resource is defined at the pod-level. In the future it might be defined at the
		//container level and then we should change the following logic to take the gpu fraction at the container-level
		podGpuFracVal, podGpuFracFound := strconv.ParseFloat(pod.Annotations[GPUFractionResourceName], 64)
		jobGpuFracVal, jobGpuFracFound := strconv.ParseFloat(job.Spec.Template.ObjectMeta.Annotations[GPUFractionResourceName], 64)
		if (podGpuFracFound == nil) != (jobGpuFracFound == nil) || podGpuFracVal != jobGpuFracVal {
			klog.V(4).Infof("Old pod: %s of job: <%v/%v>. Different GPU fraction values. "+
				"podGpuFracFound: %v, jobGpuFracFound: %v, podGpuFracVal: %v, jobGpuFracVal: %v", pod.Name,
				job.Namespace, job.Name, podGpuFracFound, jobGpuFracFound, podGpuFracVal, jobGpuFracVal)
			ret = append(ret, pod)
		} else {
			jobTemplateSpecContainerNames := make(map[string]bool)
			for _, container := range job.Spec.Template.Spec.Containers {
				jobTemplateSpecContainerNames[container.Name] = true
			}

			var podContainers []v1.Container
			for _, container := range pod.Spec.Containers {
				if jobTemplateSpecContainerNames[container.Name] {
					podContainers = append(podContainers, container)
				}
			}

			for i, c := range podContainers {
				if isContainerConfigMisMatching(&c, job, i, pod.Name) {
					ret = append(ret, pod)
				}
			}
		}
	}
	return ret
}

func wasPodSwapPreempted(pod *v1.Pod) bool {
	for _, container := range pod.Status.ContainerStatuses {
		if container.State.Terminated == nil {
			continue
		}

		if strings.Contains(container.State.Terminated.Message, failureMessageSwapPreempted) {
			return true
		}
	}

	return false
}

func getNumOfPodsToNotCountInBackoffLimit(pods []*v1.Pod) int32 {
	var result int32 = 0
	for _, pod := range pods {
		if wasPodSwapPreempted(pod) {
			result += 1
		}
	}

	return result
}

func getSwapPreemptedPods(job *runaijob.RunaiJob, pods []*v1.Pod) (swapPreemptedPods []*v1.Pod, otherPods []*v1.Pod) {
	for _, pod := range pods {
		if wasPodSwapPreempted(pod) {
			klog.V(4).Infof("Found a pod <%v/%v> in job <%v/%v> that was swap-preempted - deleting it",
				pod.Name, pod.UID, job.Namespace, job.Name)
			swapPreemptedPods = append(swapPreemptedPods, pod)
		} else {
			otherPods = append(otherPods, pod)
		}
	}

	return swapPreemptedPods, otherPods
}

func didExceedBackoffLimit(jobHaveNewFailure bool, active, failed int32, job *runaijob.RunaiJob) bool {
	// new failures happen when status does not reflect the failures and active
	// is different than parallelism, otherwise the previous controller loop
	// failed updating status so even if we pick up failure it is not a new one
	if !jobHaveNewFailure || active == pointer.Int32Deref(job.Spec.Parallelism, DefaultParallelism) {
		return false
	}

	return failed > pointer.Int32Deref(job.Spec.BackoffLimit, DefaultBackoffLimit)
}

func (jm *Controller) deleteSwapPreemptedPods(job *runaijob.RunaiJob, jobKey string, pods []*v1.Pod) (otherPods []*v1.Pod) {
	swapPreemptedPods, otherPods := getSwapPreemptedPods(job, pods)

	deletedPods := make(map[*v1.Pod]bool)
	err := jm.deletePods(job, jobKey, swapPreemptedPods, deletedPods)
	if err != nil {
		klog.Errorf("Error during attempt to delete swap-preempted pods, job: %v, err: %v", job.Name, err.Error())
		return pods
	}

	return otherPods
}

// syncJob will sync the job with the given key if it has had its expectations fulfilled, meaning
// it did not expect to see any more of its pods created or deleted. This function is not meant to be invoked
// concurrently with the same key.
func (jm *Controller) syncJob(key ctrl.Request) (bool, error) {
	startTime := time.Now()
	defer func() {
		klog.V(4).Infof("Finished syncing job %q (%v)", key, time.Since(startTime))
	}()

	ns := key.Namespace
	name := key.Name
	if len(ns) == 0 || len(name) == 0 {
		return false, fmt.Errorf("invalid job key %q: either namespace or name is missing", key)
	}

	sharedJob, err := jm.jobLister.RunaiJobs(ns).Get(name)
	if err != nil {
		if errors.IsNotFound(err) {
			klog.V(4).Infof("RunaiJob has been deleted: %v", key)
			jm.expectations.DeleteExpectations(klog.FromContext(context.Background()), key.String())
			return true, nil
		}
		return false, err
	}
	scheme.Scheme.Default(sharedJob)
	job := *sharedJob

	klog.Infof("### Sync -> %s ###", job.Name)

	// if job was finished previously, we don't want to redo the termination
	if IsJobFinished(&job) {

		klog.V(4).Infof("Job: <%v/%v> has finished", job.Namespace, job.Name)
		return true, nil
	}

	// Check the expectations of the job before counting active pods, otherwise a new pod can sneak in
	// and update the expectations after we've retrieved active pods from the store. If a new pod enters
	// the store after we've checked the expectation, the job sync is just deferred till the next relist.
	jobNeedsSync := jm.expectations.SatisfiedExpectations(klog.FromContext(context.Background()), key.String())

	pods, activePods, jobConfigMap, err := jm.getJobPodsAndUpdateConfigMap(job)
	if err != nil {
		klog.Errorf("Error during attempt to get pods, job: %v, number of pods: %v, err: %v", job.Name, len(pods), err)
		return false, err
	}

	for _, pod := range activePods {
		handleS3NotReadyState(pod, jm)
	}

	pods = jm.deleteSwapPreemptedPods(&job, key.String(), pods)

	invalidPods := getPodsNotMatchingSpec(&job, activePods)
	active := int32(len(activePods))
	succeeded, failed := getStatus(&job, pods)

	klog.Infof("active %d succeeded %d failed %d", active, succeeded, failed)

	if job.Status.StartTime == nil && !jobSuspended(&job) {
		now := metav1.Now()
		job.Status.StartTime = &now
	}

	var manageJobErr error
	jobFailed := false
	terminatedAfterPreemption := false
	var failureReason string
	var failureMessage string

	jobHaveNewFailure := failed > job.Status.Failed
	exceedsBackoffLimit := didExceedBackoffLimit(jobHaveNewFailure, active, failed, &job)

	if exceedsBackoffLimit || pastBackoffLimitOnFailure(&job, pods) {
		// check if the number of pod restart exceeds backoff (for restart OnFailure only)
		// OR if the number of failed jobs increased since the last syncJob
		jobFailed = true
		failureReason = "BackoffLimitExceeded"
		failureMessage = "RunaiJob has reached the specified backoff limit"
		klog.Infof("backoff exceeded, job marked as failed")
	} else if jm.isJobAfterPreemption(&job, pods, active, succeeded, failed) {
		jobFailed = true
		failureReason = "TerminatedAfterPreemption"
		failureMessage = "Due to project policy, this workload will not be rescheduled"
		terminatedAfterPreemption = true
		klog.Infof("job terminated after preemption")
	}

	manageJobCalled := false
	jobConditionsChanged := false

	if jobFailed || terminatedAfterPreemption {

		klog.Infof("job is marked as failed, delete remaining pods")
		manageJobErr = jm.deletePods(&job, key.String(), activePods, nil)
		if manageJobErr == nil {
			common.ClearPodsStatus(&job)
		}

		// update status values accordingly
		failed += active
		active = 0

		setConditionFalseWhenFiniteState(&job)
		job.Status.Conditions = append(job.Status.Conditions, runaiJobUtils.NewCondition(runaijob.JobFailed, v1.ConditionTrue, failureReason, failureMessage))
		jobConditionsChanged = true
		jm.recorder.Event(&job, v1.EventTypeWarning, failureReason, failureMessage)
	} else {
		if jobNeedsSync && job.DeletionTimestamp == nil {
			runaiPods := common.GenerateExtendedPods(pods)
			jobConditionsChanged = common.ShouldUpdateRunaiJobStatus(&job, runaiPods)
			active, manageJobErr = jm.manageJob(activePods, invalidPods, succeeded, &job, jobConfigMap)
			klog.Infof("invoked manageJob to re-sync the job, now active pods = %d", active)
			manageJobCalled = true
		}
		completions := succeeded
		complete := false
		if job.Spec.Completions == nil {
			// This type of job is complete when any pod exits with success.
			// Each pod is capable of
			// determining whether or not the entire RunaiJob is done.  Subsequent pods are
			// not expected to fail, but if they do, the failure is ignored.  Once any
			// pod succeeds, the controller waits for remaining pods to finish, and
			// then the job is complete.
			if succeeded > 0 && active == 0 {
				complete = true
			}
		} else {
			// RunaiJob specifies a number of completions.  This type of job signals
			// success by having that number of successes.  Since we do not
			// start more pods than there are remaining completions, there should
			// not be any remaining active pods once this count is reached.
			if completions >= pointer.Int32Deref(job.Spec.Completions, DefaultCompletions) {
				complete = true
				if active > 0 {
					jm.recorder.Event(&job, v1.EventTypeWarning, "TooManyActivePods", "Too many active pods running after completion count reached")
				}
				if completions > pointer.Int32Deref(job.Spec.Completions, DefaultCompletions) {
					jm.recorder.Event(&job, v1.EventTypeWarning, "TooManySucceededPods", "Too many succeeded pods running after completion count reached")
				}
			}
		}
		if complete {
			setConditionFalseWhenFiniteState(&job)
			job.Status.Conditions = append(job.Status.Conditions, runaiJobUtils.NewCondition(runaijob.JobComplete, v1.ConditionTrue, "", ""))
			jobConditionsChanged = true
			now := metav1.Now()
			job.Status.CompletionTime = &now
			jm.recorder.Event(&job, v1.EventTypeNormal, "Completed", "RunaiJob completed")
		} else if manageJobCalled {
			// Update the conditions / emit events only if manageJob was called in
			// this syncJob. Otherwise wait for the right syncJob call to make
			// updates.
			if jobSuspended(&job) {
				var isUpdated bool
				if job.Status.Conditions, isUpdated = runaiJobUtils.EnsureJobConditionStatus(job.Status.Conditions, runaijob.JobSuspended, v1.ConditionTrue, "JobSuspended", jobSuspendedMessage); isUpdated {
					jobConditionsChanged = true
					jm.recorder.Event(&job, v1.EventTypeNormal, "Suspended", jobSuspendedMessage)
					job.Status.StartTime = nil
				}
			} else {
				// Job not suspended
				var isUpdated bool
				if job.Status.Conditions, isUpdated = runaiJobUtils.EnsureJobConditionStatus(job.Status.Conditions, runaijob.JobSuspended, v1.ConditionFalse, "JobResumed", jobResumedMessage); isUpdated {
					jobConditionsChanged = true
					jm.recorder.Event(&job, v1.EventTypeNormal, "Resumed", jobResumedMessage)
					// Resumed jobs will always reset StartTime to current time.
					now := metav1.Now()
					job.Status.StartTime = &now
				}
			}
		}
	}

	forget := false
	// Check if the number of jobs succeeded increased since the last check. If yes "forget" should be true
	// This logic is linked to the issue: https://github.com/kubernetes/kubernetes/issues/56853 that aims to
	// improve the RunaiJob backoff policy when parallelism > 1 and few Jobs failed but others succeed.
	// In this case, we should clear the backoff delay.
	if job.Status.Succeeded < succeeded {
		forget = true
	}

	// no need to update the job if the status hasn't changed since last time
	if job.Status.Active != active || job.Status.Succeeded != succeeded || job.Status.Failed != failed || jobConditionsChanged {
		job.Status.Active = active
		job.Status.Succeeded = succeeded
		job.Status.Failed = failed

		if err := jm.updateHandler(&job); err != nil {
			klog.Errorf("Failed to set status of job: <%v/%v>, err: <%v>", job.Namespace, job.Name, err)
			return forget, err
		}

		if jobHaveNewFailure && !IsJobFinished(&job) {
			// returning an error will re-enqueue RunaiJob after the backoff period
			return forget, fmt.Errorf("failed pod(s) detected for job key %q", key)
		}

		forget = true
	}
	return forget, manageJobErr
}

func setConditionFalseWhenFiniteState(job *runaijob.RunaiJob) {
	if job == nil {
		return
	}
	job.Status.Conditions, _ = runaiJobUtils.EnsureJobConditionStatus(job.Status.Conditions, runaijob.JobSuspended, v1.ConditionFalse, "", "")
	job.Status.Conditions, _ = runaiJobUtils.EnsureJobConditionStatus(job.Status.Conditions, runaijob.JobPending, v1.ConditionFalse, "", "")
	job.Status.Conditions, _ = runaiJobUtils.EnsureJobConditionStatus(job.Status.Conditions, runaijob.JobRunning, v1.ConditionFalse, "", "")
}

func (jm *Controller) isJobAfterPreemption(job *runaijob.RunaiJob, pods []*v1.Pod, active int32, succeeded int32, failed int32) bool {
	if job == nil {
		return false
	}
	return pointer.BoolDeref(job.Spec.TerminateAfterPreemption, defaultOfTerminateAfterPreemption) &&
		pods != nil && active < int32(len(pods)) && succeeded == 0 && failed == 0 && !jobSuspended(job)
}

func (jm *Controller) getJobPodsAndUpdateConfigMap(job runaijob.RunaiJob) ([]*v1.Pod, []*v1.Pod, *v1.ConfigMap, error) {
	pods, err := jm.getPodsForJob(&job)
	if err != nil {
		klog.Errorf("Failed to get pods for job: <%v/%v>", job.Namespace, job.Name)
		return pods, nil, nil, fmt.Errorf("failed to get pods for job")
	}

	activePods := controller.FilterActivePods(klog.FromContext(context.Background()), pods)

	jobConfigMap, err := jm.handleJobConfigMap(pods, &job)
	if err != nil {
		klog.Errorf("Failed to handle configmap for job: <%v/%v>", job.Namespace, job.Name)
		return pods, activePods, jobConfigMap, fmt.Errorf("failed to get or create configmap")
	}
	return pods, activePods, jobConfigMap, nil
}

func (jm *Controller) getPodNameToDelete(job *runaijob.RunaiJob, mapOfLastIndexPods *map[int64]podLastIndexAndStatus) (string, error) {
	// look for the pod with the highest index
	biggestActiveIndex := int64(-1)
	for podIndex, value := range *mapOfLastIndexPods {
		if value.status == podStatus(v1.PodRunning) || value.status == podStatus(v1.PodPending) {
			if podIndex > biggestActiveIndex {
				biggestActiveIndex = podIndex
			}
		}
	}
	if biggestActiveIndex == -1 {
		return "", fmt.Errorf("failed to find an existing pod to delete")
	}

	(*mapOfLastIndexPods)[biggestActiveIndex] = podLastIndexAndStatus{
		lastIndexAttempt: (*mapOfLastIndexPods)[biggestActiveIndex].lastIndexAttempt,
		status:           phaseDeleting,
	}
	return fmt.Sprintf("%s-%d-%d", job.Name, biggestActiveIndex, (*mapOfLastIndexPods)[biggestActiveIndex].lastIndexAttempt), nil
}

func (jm *Controller) getPodIndexAndAttemptToCreate(job *runaijob.RunaiJob, mapOfLastIndexPods *map[int64]podLastIndexAndStatus) (int64, int64, error) {
	// look for deletion or failure:
	biggestIndex := int64(-1)
	for podIndex, value := range *mapOfLastIndexPods {
		if value.status == podStatus(v1.PodFailed) || value.status == phaseDeleted {
			(*mapOfLastIndexPods)[podIndex] = podLastIndexAndStatus{
				lastIndexAttempt: value.lastIndexAttempt + 1,
				status:           phaseCreating,
			}
			klog.V(4).Infof("Job: <%v/%v> had pod with index of: <%v>, attempt: <%v> status: <%v>, creating pod with attempt: <%v>", job.Namespace, job.Name, podIndex, value.lastIndexAttempt, value.status, value.lastIndexAttempt+1)
			return podIndex, value.lastIndexAttempt + 1, nil
		}
		if biggestIndex < podIndex {
			biggestIndex = podIndex
		}
	}

	// create a new index
	if job.Spec.Completions == nil || *job.Spec.Completions > int32(biggestIndex)+1 {
		(*mapOfLastIndexPods)[biggestIndex+1] = podLastIndexAndStatus{
			lastIndexAttempt: 0,
			status:           phaseCreating,
		}
		klog.V(4).Infof("Job: <%v/%v> biggest index was: <%v>, creating pod with index of: <%v>", job.Namespace, job.Name, biggestIndex, biggestIndex+1)
		return biggestIndex + 1, 0, nil
	}

	return 0, 0, fmt.Errorf("all pods of job already existing, failed to get name for pod")
}

func (jm *Controller) updateJobConfigMap(job *runaijob.RunaiJob, activePods []*v1.Pod, cm *v1.ConfigMap) error {
	shouldUpdateConfigMap := false
	// fill configmap with latest pod statuses
	for _, pod := range activePods {
		if pod.Labels == nil {
			continue
		}
		podSuffix, found := pod.Labels[labelOfPodIndex]
		if !found {
			klog.V(4).Infof("Job: <%v/%v> had no label of pod index, probably an old job", job.Namespace, job.Name)
			continue
		}
		status := podStatus(pod.Status.Phase)
		if pod.DeletionTimestamp != nil {
			status = phaseDeleted
		}
		if cmStatus, found := cm.Data[podSuffix]; !found || cmStatus != string(status) {
			if len(cm.Data) == 0 {
				cm.Data = map[string]string{}
			}
			if cm.Data[podSuffix] != string(status) {
				cm.Data[podSuffix] = string(status)
				shouldUpdateConfigMap = true
			}
		}
	}

	// verify we dont have leftovers in active state
	for podSuffix, podStatus := range cm.Data {
		if podStatus != string(v1.PodPending) && podStatus != string(v1.PodRunning) {
			continue // pod is finished
		}
		wasPodFoundAsActive := false
		for _, currentPod := range activePods {
			if currentPod.Labels == nil {
				continue
			}
			currentPodSuffix, found := currentPod.Labels[labelOfPodIndex]
			if !found {
				klog.V(4).Infof("Job: <%v/%v> had no label of pod index, probably an old job", job.Namespace, job.Name)
				continue
			}
			if currentPodSuffix == podSuffix {
				wasPodFoundAsActive = true
				break
			}
		}
		if !wasPodFoundAsActive {
			klog.V(4).Infof("job: <%v/%v>, suffix: <%v>, is not longer active", job.Namespace, job.Name, podSuffix)
			cm.Data[podSuffix] = string(phaseDeleted)
			shouldUpdateConfigMap = true
		}
	}

	if shouldUpdateConfigMap {
		_, err := jm.kubeClient.CoreV1().ConfigMaps(job.Namespace).Update(context.TODO(), cm, metav1.UpdateOptions{})
		if err != nil {
			klog.Warningf("Failed to update configmap: <%v/%v>", cm.Namespace, cm.Name)
			return err
		} else {
			klog.V(4).Infof("Updated configmap: <%v/%v>", cm.Namespace, cm.Name)
		}
	} else {
		klog.V(4).Infof("No need to updated configmap: <%v/%v>", cm.Namespace, cm.Name)
	}
	return nil
}

func (jm *Controller) getOrCreateJobConfigmap(job *runaijob.RunaiJob) (*v1.ConfigMap, error) {
	configmapName := fmt.Sprintf("%s-pod-attempts", job.Name)
	cm, err := jm.configMapStore.ConfigMaps(job.Namespace).Get(configmapName)

	if err == nil && len(cm.OwnerReferences) > 0 && job.UID == cm.OwnerReferences[0].UID {
		return cm, nil
	}

	cmRequiresUpdate := cm != nil && (len(cm.OwnerReferences) == 0 || job.UID != cm.OwnerReferences[0].UID)

	cm = &core1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name: configmapName,
		},
		Data: make(map[string]string),
	}
	cm.OwnerReferences = append(cm.OwnerReferences, *metav1.NewControllerRef(job, controllerKind))

	if cmRequiresUpdate {
		for i := 0; i < kubeClientUpdateRetries; i++ {
			cm, err = jm.kubeClient.CoreV1().ConfigMaps(job.Namespace).Update(context.TODO(), cm, metav1.UpdateOptions{})
			if err == nil {
				return cm, err
			}
		}

	}
	cm, err = jm.kubeClient.CoreV1().ConfigMaps(job.Namespace).Create(context.TODO(), cm, metav1.CreateOptions{})
	return cm, err
}

// pastBackoffLimitOnFailure checks if container restartCounts sum exceeds BackoffLimit
// this method applies only to pods with restartPolicy == OnFailure
func pastBackoffLimitOnFailure(job *runaijob.RunaiJob, pods []*v1.Pod) bool {
	if job.Spec.Template.Spec.RestartPolicy != v1.RestartPolicyOnFailure {
		return false
	}
	result := int32(0)
	for i := range pods {
		po := pods[i]
		if po.Status.Phase == v1.PodRunning || po.Status.Phase == v1.PodPending {
			for j := range po.Status.InitContainerStatuses {
				stat := po.Status.InitContainerStatuses[j]
				result += stat.RestartCount
			}
			for j := range po.Status.ContainerStatuses {
				stat := po.Status.ContainerStatuses[j]
				result += stat.RestartCount
			}
		}
	}
	if pointer.Int32Deref(job.Spec.BackoffLimit, DefaultBackoffLimit) == 0 {
		return result > 0
	}
	return result >= pointer.Int32Deref(job.Spec.BackoffLimit, DefaultBackoffLimit)
}

// jobSuspended returns whether a Job is suspended.
func jobSuspended(job *runaijob.RunaiJob) bool {
	return job.Spec.Suspend != nil && *job.Spec.Suspend
}

// getStatus returns no of succeeded and failed pods running a job
func getStatus(job *runaijob.RunaiJob, pods []*v1.Pod) (succeeded, failed int32) {
	ignoreDeletedPods := true
	succeeded = int32(filterPods(pods, v1.PodSucceeded, ignoreDeletedPods))
	failed = int32(filterPods(pods, v1.PodFailed, ignoreDeletedPods))

	numOfPodsToNotCountAsFailed := getNumOfPodsToNotCountInBackoffLimit(pods)
	if numOfPodsToNotCountAsFailed > 0 {
		klog.V(4).Infof("Found %v pods in job <%v/%v> that were terminated by runai - not counting them as failed",
			numOfPodsToNotCountAsFailed, job.Namespace, job.Name)
		failed -= numOfPodsToNotCountAsFailed
	}

	return
}

func (jm *Controller) deletePods(job *runaijob.RunaiJob, jobKey string, podsToDelete []*v1.Pod,
	deletedPods map[*v1.Pod]bool) error {
	wait := sync.WaitGroup{}
	diff := len(podsToDelete)
	errCh := make(chan error, diff)
	err := jm.expectations.ExpectDeletions(klog.FromContext(context.Background()), jobKey, diff)
	if err != nil {
		klog.Errorf("Error from ExpectDeletions, error: <%v>", err)
	}
	sort.Sort(controller.ActivePods(podsToDelete))
	deletedPodsLock := sync.Mutex{}
	wait.Add(diff)
	for _, p := range podsToDelete {
		go func(pod *v1.Pod) {
			defer wait.Done()
			if err := jm.control.DeletePod(job.Namespace, pod.Name, job); err != nil {
				jm.expectations.DeletionObserved(klog.FromContext(context.Background()), jobKey)
				if !apierrors.IsNotFound(err) {
					klog.V(2).Infof("Failed to delete %v, decremented expectations for job %q/%q", pod.Name, job.Namespace, job.Name)
					errCh <- err
					utilruntime.HandleError(err)
				}
			} else {
				deletedPodsLock.Lock()
				if deletedPods != nil {
					deletedPods[pod] = true
				}
				deletedPodsLock.Unlock()
			}
		}(p)
	}
	wait.Wait()
	return errorFromChannel(errCh)
}

func (jm *Controller) locallySetPodStatus(mapOfLastIndexPods *map[int64]podLastIndexAndStatus, pod *v1.Pod,
	newStatus podStatus) error {
	podNameParts := strings.Split(pod.Name, "-")
	if len(podNameParts) != 3 {
		return fmt.Errorf("failed to parse pod name: %s", pod.Name)
	}
	podIndex, err := strconv.ParseInt(podNameParts[1], 10, 64)
	if err != nil {
		return fmt.Errorf("failed to parse pod index. Pod name: %s", pod.Name)
	}
	podAttempt, err := strconv.ParseInt(podNameParts[2], 10, 64)
	if err != nil {
		return fmt.Errorf("failed to parse pod attempt. Pod name: %s", pod.Name)
	}
	(*mapOfLastIndexPods)[podIndex] = podLastIndexAndStatus{
		lastIndexAttempt: podAttempt,
		status:           newStatus,
	}
	return nil
}

// manageJob is the core method responsible for managing the number of running
// pods according to what is specified in the job.Spec.
// Does NOT modify <activePods>.
func (jm *Controller) manageJob(activePods []*v1.Pod, invalidPods []*v1.Pod, succeeded int32, job *runaijob.RunaiJob, jobConfigMap *v1.ConfigMap) (int32, error) {
	var activeLock sync.Mutex
	var podNameLock sync.Mutex
	active := int32(len(activePods))
	var parallelism int32 = 1
	if job.Spec.Parallelism != nil {
		parallelism = pointer.Int32Deref(job.Spec.Parallelism, DefaultParallelism)
	}
	jobKey, err := controller.KeyFunc(job)
	if err != nil {
		utilruntime.HandleError(fmt.Errorf("couldn't get key for job %#v: %v", job, err))
		return 0, nil
	}

	if jobSuspended(job) {
		klog.V(4).InfoS("Deleting all active pods in suspended job", "job", klog.KObj(job), "active", active)
		if expectedDeletionsErr := jm.expectations.ExpectDeletions(klog.FromContext(context.Background()),
			jobKey, len(activePods)); expectedDeletionsErr != nil {
			return active, expectedDeletionsErr
		}
		deletedPods := make(map[*v1.Pod]bool)
		err := jm.deletePods(job, jobKey, activePods, deletedPods)
		active -= int32(len(deletedPods))
		return active, err
	}

	mapOfLastIndexPods := jm.getMapOfLastIndexPodAndStatus(jobConfigMap)
	var invalidPodsErrCh, errCh chan error
	if len(invalidPods) > 0 {
		deletedPods := make(map[*v1.Pod]bool)
		invalidPodsErrCh = make(chan error, len(invalidPods))
		invalidPodsErrCh <- jm.deletePods(job, jobKey, invalidPods, deletedPods)
		var newActive []*v1.Pod
		for _, pod := range activePods {
			if deletedPods[pod] {
				if err := jm.locallySetPodStatus(&mapOfLastIndexPods, pod, phaseDeleting); err != nil {
					return int32(len(newActive)), fmt.Errorf("error in updating pod status, error: <%v>, pod: %s", err, pod.Name)
				}
			} else {
				newActive = append(newActive, pod)
			}
		}
		activePods = newActive
		active = int32(len(activePods))
	}
	if active > parallelism {
		diff := active - parallelism
		errCh = make(chan error, diff)
		err := jm.expectations.ExpectDeletions(klog.FromContext(context.Background()), jobKey, int(diff))
		if err != nil {
			klog.Errorf("Error from ExpectDeletions, error: <%v>", err)
		}
		klog.V(4).Infof("Too many pods running job %q, need %d, deleting %d", jobKey, parallelism, diff)
		// Sort the pods in the order such that not-ready < ready, unscheduled
		// < scheduled, and pending < running. This ensures that we delete pods
		// in the earlier stages whenever possible.
		sort.Sort(controller.ActivePods(activePods))

		active -= diff
		wait := sync.WaitGroup{}
		wait.Add(int(diff))
		for i := int32(0); i < diff; i++ {
			go func() {
				defer wait.Done()
				podNameLock.Lock()
				podName, err := jm.getPodNameToDelete(job, &mapOfLastIndexPods)
				podNameLock.Unlock()
				if err != nil {
					// Decrement the expected number of deletes because the informer won't observe this deletion
					jm.expectations.DeletionObserved(klog.FromContext(context.Background()), jobKey)
					klog.V(2).Infof("Failed to delete %v, decremented expectations for job %q/%q", podName, job.Namespace, job.Name)
					activeLock.Lock()
					active++
					activeLock.Unlock()
					errCh <- err
					utilruntime.HandleError(err)
					return
				}

				if err := jm.control.DeletePod(job.Namespace, podName, job); err != nil {
					// Decrement the expected number of deletes because the informer won't observe this deletion
					jm.expectations.DeletionObserved(klog.FromContext(context.Background()), jobKey)
					if !apierrors.IsNotFound(err) {
						klog.V(2).Infof("Failed to delete %v, decremented expectations for job %q/%q", podName, job.Namespace, job.Name)
						activeLock.Lock()
						active++
						activeLock.Unlock()
						errCh <- err
						utilruntime.HandleError(err)
					}
				}
			}()
		}
		wait.Wait()

	} else if active < parallelism {
		wantActive := int32(0)
		if job.Spec.Completions == nil {
			// RunaiJob does not specify a number of completions.  Therefore, number active
			// should be equal to parallelism, unless the job has seen at least
			// once success, in which leave whatever is running, running.
			if succeeded > 0 {
				wantActive = active
			} else {
				wantActive = parallelism
			}
		} else {
			// RunaiJob specifies a specific number of completions.  Therefore, number
			// active should not ever exceed number of remaining completions.
			wantActive = *job.Spec.Completions - succeeded
			if wantActive > parallelism {
				wantActive = parallelism
			}
		}
		diff := wantActive - active
		if diff < 0 {
			utilruntime.HandleError(fmt.Errorf("More active than wanted: job %q, want %d, have %d", jobKey, wantActive, active))
			diff = 0
		}
		if diff == 0 {
			return active, nil
		}
		err := jm.expectations.ExpectCreations(klog.FromContext(context.Background()), jobKey, int(diff))
		if err != nil {
			klog.Errorf("Error from ExpectDeletions, error: <%v>", err)
		}
		errCh = make(chan error, diff)
		klog.V(4).Infof("Too few pods running job %q, need %d, creating %d", jobKey, wantActive, diff)

		active += diff
		wait := sync.WaitGroup{}

		// Batch the pod creates. Batch sizes start at SlowStartInitialBatchSize
		// and double with each successful iteration in a kind of "slow start".
		// This handles attempts to start large numbers of pods that would
		// likely all fail with the same error. For example a project with a
		// low quota that attempts to create a large number of pods will be
		// prevented from spamming the API service with the pod create requests
		// after one of its pods fails.  Conveniently, this also prevents the
		// event spam that those failures would generate.
		for batchSize := int32(integer.IntMin(int(diff), controller.SlowStartInitialBatchSize)); diff > 0; batchSize = integer.Int32Min(2*batchSize, diff) {
			errorCount := len(errCh)
			wait.Add(int(batchSize))
			if job.Spec.Template.Labels == nil {
				job.Spec.Template.Labels = map[string]string{}
			}
			job.Spec.Template.Labels["controller-uid"] = string(job.UID)
			for i := int32(0); i < batchSize; i++ {
				go func() {
					defer wait.Done()
					podNameLock.Lock()
					podIndex, podAttempt, err := jm.getPodIndexAndAttemptToCreate(job, &mapOfLastIndexPods)
					podName := fmt.Sprintf("%s-%d-%d", job.Name, podIndex, podAttempt)
					podNameLock.Unlock()
					if err != nil {
						defer utilruntime.HandleError(err)
						jm.handleErrorDuringPodCreation(job, jobKey, &activeLock, &active, &errCh, err)
						return
					}
					err = jm.control.CreatePod(job, podName, podIndex, podAttempt)
					if err != nil {
						if errors.HasStatusCause(err, v1.NamespaceTerminatingCause) {
							// If the namespace is being torn down, we can safely ignore
							// this error since all subsequent creations will fail.
							return
						}
					}
					if err != nil {
						defer utilruntime.HandleError(err)
						jm.handleErrorDuringPodCreation(job, jobKey, &activeLock, &active, &errCh, err)
					}
				}()
			}
			wait.Wait()
			// any skipped pods that we never attempted to start shouldn't be expected.
			skippedPods := diff - batchSize
			if errorCount < len(errCh) && skippedPods > 0 {
				klog.V(2).Infof("Slow-start failure. Skipping creation of %d pods, decrementing expectations for job %q/%q", skippedPods, job.Namespace, job.Name)
				active -= skippedPods
				for i := int32(0); i < skippedPods; i++ {
					// Decrement the expected number of creates because the informer won't observe this pod
					jm.expectations.CreationObserved(klog.FromContext(context.Background()), jobKey)
				}
				// The skipped pods will be retried later. The next controller resync will
				// retry the slow start process.
				break
			}
			diff -= batchSize
		}
	}

	select {
	// all errors have been reported before, we only need to inform the controller that there was an error and it should re-try this job once more next time.
	case err := <-invalidPodsErrCh:
		if err != nil {
			return active, err
		}
	case err := <-errCh:
		if err != nil {
			return active, err
		}
	default:
	}

	return active, nil
}

func (jm *Controller) handleErrorDuringPodCreation(job *runaijob.RunaiJob, jobKey string, activeLock *sync.Mutex, active *int32, errCh *chan error, err error) {
	// Decrement the expected number of creates because the informer won't observe this pod
	klog.V(2).Infof("Failed creation, decrementing expectations for job %q/%q", job.Namespace, job.Name)
	jm.expectations.CreationObserved(klog.FromContext(context.Background()), jobKey)
	activeLock.Lock()
	*active--
	activeLock.Unlock()
	*errCh <- err
}

func (jm *Controller) getMapOfLastIndexPodAndStatus(jobConfigMap *v1.ConfigMap) map[int64]podLastIndexAndStatus {
	mapOfLastIndexPods := map[int64]podLastIndexAndStatus{}
	for podSuffix, status := range jobConfigMap.Data {
		podSuffixArray := strings.Split(podSuffix, "-")
		if len(podSuffix) < 2 {
			klog.Errorf("Failed to split pod suffix: <%v>", podSuffix)
			continue
		}
		podIndex, err := strconv.ParseInt(podSuffixArray[0], 10, 64)
		if err != nil {
			klog.Errorf("Failed to parse pod index: <%v>", podSuffix)
			continue
		}
		podIndexAttempt, err := strconv.ParseInt(podSuffixArray[1], 10, 64)
		if err != nil {
			klog.Errorf("Failed to parse pod index attempt: <%v>", podSuffix)
			continue
		}
		_, found := mapOfLastIndexPods[podIndex]
		if !found || mapOfLastIndexPods[podIndex].lastIndexAttempt < podIndexAttempt {
			mapOfLastIndexPods[podIndex] = podLastIndexAndStatus{
				lastIndexAttempt: podIndexAttempt,
				status:           podStatus(status),
			}
		}
	}
	return mapOfLastIndexPods
}

func (jm *Controller) handleJobConfigMap(activePods []*v1.Pod, job *runaijob.RunaiJob) (*v1.ConfigMap, error) {
	cm, err := jm.getOrCreateJobConfigmap(job)
	if err != nil {
		klog.Errorf("Failed to get or create configmap for job: <%v/%v>, error: <%v>", job.Namespace, job.Name, err)
		return nil, fmt.Errorf("failed to create or get configmap for job")
	}
	err = jm.updateJobConfigMap(job, activePods, cm)
	return cm, err
}

func (jm *Controller) updateJobStatus(job *runaijob.RunaiJob) error {
	jobClient := jm.kubeRunaiClient.RunV1().RunaiJobs(job.Namespace)
	var err error
	for i := 0; i <= kubeClientUpdateRetries; i = i + 1 {
		var newJob *runaijob.RunaiJob
		newJob, err = jobClient.Get(context.TODO(), job.Name, metav1.GetOptions{})
		if err != nil {
			break
		}
		newJob.Status = job.Status
		if _, err = jobClient.UpdateStatus(context.TODO(), newJob, metav1.UpdateOptions{}); err == nil {
			break
		}
		klog.Errorf("failed to update status of job: %v/%v, attempt: %v, err: %v", job.Namespace, job.Name, i, err)
	}

	return err
}

func getBackoff(queue workqueue.TypedRateLimitingInterface[ctrl.Request], key ctrl.Request) time.Duration {
	exp := queue.NumRequeues(key)

	if exp <= 0 {
		return time.Duration(0)
	}

	// The backoff is capped such that 'calculated' value never overflows.
	backoff := float64(DefaultJobBackOff.Nanoseconds()) * math.Pow(2, float64(exp-1))
	if backoff > math.MaxInt64 {
		return MaxJobBackOff
	}

	calculated := time.Duration(backoff)
	if calculated > MaxJobBackOff {
		return MaxJobBackOff
	}
	return calculated
}

// filterPods returns pods based on their phase.
func filterPods(pods []*v1.Pod, phase v1.PodPhase, ignoreDeletedPods bool) int {
	result := 0
	for i := range pods {
		if ignoreDeletedPods && pods[i].DeletionTimestamp != nil {
			// This part was added due to a bug in k8s where the job might be set to completed/failed when the pod is during terminating state, this was fixed in: https://github.com/kubernetes/kubernetes/pull/88440 but was not merged to all releases yet.
			continue
		}

		if phase == pods[i].Status.Phase {
			result++
		}
	}
	return result
}

func errorFromChannel(errCh <-chan error) error {
	select {
	case err := <-errCh:
		return err
	default:
	}
	return nil
}
