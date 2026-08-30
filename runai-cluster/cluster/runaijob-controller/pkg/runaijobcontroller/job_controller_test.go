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
	"strconv"
	"testing"
	"time"

	"github.com/run-ai/runai/runai-cluster/cluster/runaijob-controller/pkg/common/runaiJobUtils"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/google/go-cmp/cmp"
	v1 "k8s.io/api/core/v1"
	apiequality "k8s.io/apimachinery/pkg/api/equality"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/uuid"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/informers"
	kubeinformers "k8s.io/client-go/informers"
	clientset "k8s.io/client-go/kubernetes"
	clientfeatures "k8s.io/client-go/features"
	clientfeaturestesting "k8s.io/client-go/features/testing"
	"k8s.io/client-go/kubernetes/fake"
	restclient "k8s.io/client-go/rest"
	core "k8s.io/client-go/testing"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/util/workqueue"
	"k8s.io/klog/v2"
	_ "k8s.io/kubernetes/pkg/apis/core/install"
	"k8s.io/kubernetes/pkg/controller"
	"k8s.io/utils/pointer"

	runaijobv1 "github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1"
	clientsetrunai "github.com/run-ai/runai/runai-cluster/cluster/sdk/client/clientset/versioned"
	fakeclientset "github.com/run-ai/runai/runai-cluster/cluster/sdk/client/clientset/versioned/fake"
	runaijobinformers "github.com/run-ai/runai/runai-cluster/cluster/sdk/client/informers/externalversions"
)

var alwaysReady = func() bool { return true }

const (
	JobSuspendedReason = "JobSuspended"
)

func getCondition(job *runaijobv1.RunaiJob, condition runaijobv1.JobConditionType, status v1.ConditionStatus, reason string) bool {
	for _, v := range job.Status.Conditions {
		if v.Type == condition && v.Status == status && v.Reason == reason {
			return true
		}
	}
	return false
}
func newJob(parallelism, completions, backoffLimit int32) *runaijobv1.RunaiJob {
	jobUid := uuid.NewUUID()
	j := &runaijobv1.RunaiJob{
		TypeMeta: metav1.TypeMeta{Kind: "RunaiJob"},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "foobar",
			UID:       jobUid,
			Namespace: metav1.NamespaceDefault,
		},
		Spec: runaijobv1.JobSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{"controller-uid": string(jobUid)},
			},
			Template: v1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"controller-uid": string(jobUid),
					},
				},
				Spec: v1.PodSpec{
					Containers: []v1.Container{
						{Image: "foo/bar", Name: "foo"},
					},
				},
			},
		},
	}
	// Special case: -1 for either completions or parallelism means leave nil (negative is not allowed
	// in practice by validation.
	if completions >= 0 {
		j.Spec.Completions = &completions
	} else {
		j.Spec.Completions = nil
	}
	if parallelism >= 0 {
		j.Spec.Parallelism = &parallelism
	} else {
		j.Spec.Parallelism = nil
	}
	j.Spec.BackoffLimit = &backoffLimit

	return j
}

func newControllerFromClient(kubeClient clientset.Interface, clientsetrunai clientsetrunai.Interface,
	resyncPeriod controller.ResyncPeriodFunc) (*Controller, informers.SharedInformerFactory, runaijobinformers.SharedInformerFactory) {
	runaiSharedInformers := runaijobinformers.NewSharedInformerFactory(clientsetrunai, resyncPeriod())
	kubeInformerFactory := kubeinformers.NewSharedInformerFactory(kubeClient, resyncPeriod())

	klog.V(3).Infof("Initializing Runai Job Controller")
	jm := NewController(
		nil,
		kubeInformerFactory,
		runaiSharedInformers,
		kubeClient,
		clientsetrunai,
	)
	jm.podControl = &controller.FakePodControl{}
	jm.control = &FakePodControl{}

	return jm, kubeInformerFactory, runaiSharedInformers
}

func newController() (*Controller, informers.SharedInformerFactory, runaijobinformers.SharedInformerFactory) {
	clientset := clientset.NewForConfigOrDie(&restclient.Config{Host: "", ContentConfig: restclient.ContentConfig{GroupVersion: &schema.GroupVersion{Group: "", Version: "v1"}}})
	clientsetrunai := clientsetrunai.NewForConfigOrDie(&restclient.Config{Host: "", ContentConfig: restclient.ContentConfig{GroupVersion: &schema.GroupVersion{Group: "", Version: "v1"}}})
	jm, sharedInformerFactory, runaiSharedInformersFactory := newControllerFromClient(clientset, clientsetrunai, controller.NoResyncPeriodFunc)
	jm.podStoreSynced = alwaysReady
	jm.jobStoreSynced = alwaysReady

	return jm, sharedInformerFactory, runaiSharedInformersFactory
}

func newControllerWithRunaiClientset(job *runaijobv1.RunaiJob) (*Controller, informers.SharedInformerFactory, runaijobinformers.SharedInformerFactory) {
	clientsetrunai := fakeclientset.NewSimpleClientset(job)
	clientset := clientset.NewForConfigOrDie(&restclient.Config{Host: "", ContentConfig: restclient.ContentConfig{GroupVersion: &schema.GroupVersion{Group: "", Version: "v1"}}})
	jm, sharedInformerFactory, runaiSharedInformersFactory := newControllerFromClient(clientset, clientsetrunai, controller.NoResyncPeriodFunc)
	jm.podStoreSynced = alwaysReady
	jm.jobStoreSynced = alwaysReady

	return jm, sharedInformerFactory, runaiSharedInformersFactory
}

func newPod(podSuffix string, job *runaijobv1.RunaiJob, status v1.PodPhase, conditions *[]v1.PodCondition) *v1.Pod {
	labels := map[string]string{}
	for key, value := range job.Spec.Selector.MatchLabels {
		labels[key] = value
	}
	labels[labelOfPodIndex] = podSuffix
	pod := &v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:            fmt.Sprintf("%s-%s", job.Name, podSuffix),
			Labels:          labels,
			Namespace:       job.Namespace,
			OwnerReferences: []metav1.OwnerReference{*metav1.NewControllerRef(job, controllerKind)},
		},
		Status: v1.PodStatus{
			Phase: status,
		},
		Spec: v1.PodSpec{
			Containers: []v1.Container{
				{Image: "foo/bar", Name: "foo"},
			},
		},
	}

	if conditions != nil {
		pod.Status.Conditions = *conditions
	}

	return pod
}

// create count pods with the given phase for the given job
func newPodList(podIndexer cache.Indexer, count int32, status v1.PodPhase, conditions *[]v1.PodCondition, job *runaijobv1.RunaiJob, index *int, configMap *v1.ConfigMap, numberOfPodRestarts []int32) []v1.Pod {
	var pods []v1.Pod
	for i := int32(0); i < count; i++ {
		podSuffix := fmt.Sprintf("%d-0", *index)
		configMap.Data[podSuffix] = string(status)
		newPod := newPod(podSuffix, job, status, conditions)
		err := podIndexer.Add(newPod)
		if err != nil {
			klog.Errorf("failed to add pod")
		}
		if numberOfPodRestarts != nil {
			newPod.Status.ContainerStatuses = []v1.ContainerStatus{{RestartCount: numberOfPodRestarts[i]}}

		}
		*index++
		pods = append(pods, *newPod)
	}
	return pods
}

func setPodsStatusesAndGetConfigMap(podIndexer cache.Indexer, job *runaijobv1.RunaiJob, pendingPods, activePods, succeededPods, failedPods int32) *v1.ConfigMap {
	configmap := v1.ConfigMap{
		TypeMeta: metav1.TypeMeta{Kind: "ConfigMap", APIVersion: "v1"},
		ObjectMeta: metav1.ObjectMeta{
			Name:            fmt.Sprintf("%s-pod-attempts", job.Name),
			Namespace:       job.Namespace,
			UID:             uuid.NewUUID(),
			ResourceVersion: "1",
			OwnerReferences: []metav1.OwnerReference{*metav1.NewControllerRef(job, controllerKind)},
		},
		Data: make(map[string]string),
	}
	index := 0
	newPodList(podIndexer, pendingPods, v1.PodPending, nil, job, &index, &configmap, nil)
	newPodList(podIndexer, activePods, v1.PodRunning, createCondition(CreateConditionRunningTrue), job, &index, &configmap, nil)
	newPodList(podIndexer, succeededPods, v1.PodSucceeded, nil, job, &index, &configmap, nil)
	newPodList(podIndexer, failedPods, v1.PodFailed, nil, job, &index, &configmap, nil)
	return &configmap
}

func TestControllerSyncJob(t *testing.T) {
	jobConditionComplete := runaijobv1.JobComplete
	jobConditionFailed := runaijobv1.JobFailed
	jobConditionSuspended := runaijobv1.JobSuspended

	testCases := map[string]struct {
		// job setup
		parallelism  int32
		completions  int32
		backoffLimit int32
		deleting     bool
		podLimit     int
		wasSuspended bool
		suspend      bool

		// pod setup
		podControllerError        error
		jobKeyForget              bool
		pendingPods               int32
		activePods                int32
		succeededPods             int32
		failedPods                int32
		fakeExpectationAtCreation int32 // negative: ExpectDeletions, positive: ExpectCreations

		// expectations
		expectedCreations       int32
		expectedDeletions       int32
		expectedActive          int32
		expectedSucceeded       int32
		expectedFailed          int32
		expectedCondition       *runaijobv1.JobConditionType
		expectedConditionStatus v1.ConditionStatus
		expectedConditionReason string
	}{
		// "job start": {
		// 	2, 5, 6, false, 0,
		// 	nil, true, 0, 0, 0, 0,
		// 	2, 0, 2, 0, 0, nil, "",
		// },
		// "WQ job start": {
		// 	2, -1, 6, false, 0,
		// 	nil, true, 0, 0, 0, 0,
		// 	2, 0, 2, 0, 0, nil, "",
		// },
		// "pending pods": {
		// 	2, 5, 6, false, 0,
		// 	nil, true, 2, 0, 0, 0,
		// 	0, 0, 2, 0, 0, nil, "",
		// },
		// "correct # of pods": {
		// 	2, 5, 6, false, 0,
		// 	nil, true, 0, 2, 0, 0,
		// 	0, 0, 2, 0, 0, nil, "",
		// },
		// "WQ job: correct # of pods": {
		// 	2, -1, 6, false, 0,
		// 	nil, true, 0, 2, 0, 0,
		// 	0, 0, 2, 0, 0, nil, "",
		// },
		// "too few active pods": {
		// 	2, 5, 6, false, 0,
		// 	nil, true, 0, 1, 1, 0,
		// 	1, 0, 2, 1, 0, nil, "",
		// },
		// "too few active pods with a dynamic job": {
		// 	2, -1, 6, false, 0,
		// 	nil, true, 0, 1, 0, 0,
		// 	1, 0, 2, 0, 0, nil, "",
		// },
		// "too few active pods, with controller error": {
		// 	2, 5, 6, false, 0,
		// 	fmt.Errorf("fake error"), true, 0, 1, 1, 0,
		// 	1, 0, 1, 1, 0, nil, "",
		// },
		// "too many active pods": {
		// 	2, 5, 6, false, 0,
		// 	nil, true, 0, 3, 0, 0,
		// 	0, 1, 2, 0, 0, nil, "",
		// },
		// "too many active pods, with controller error": {
		// 	2, 5, 6, false, 0,
		// 	fmt.Errorf("fake error"), true, 0, 3, 0, 0,
		// 	0, 1, 3, 0, 0, nil, "",
		// },
		// "failed + succeed pods: reset backoff delay": {
		// 	2, 5, 6, false, 0,
		// 	fmt.Errorf("fake error"), true, 0, 1, 1, 1,
		// 	1, 0, 1, 1, 1, nil, "",
		// },
		// "only new failed pod": {
		// 	2, 5, 6, false, 0,
		// 	fmt.Errorf("fake error"), false, 0, 1, 0, 1,
		// 	1, 0, 1, 0, 1, nil, "",
		// },
		// "job finish": {
		// 	2, 5, 6, false, 0,
		// 	nil, true, 0, 0, 5, 0,
		// 	0, 0, 0, 5, 0, nil, "",
		// },
		// "WQ job finishing": {
		// 	2, -1, 6, false, 0,
		// 	nil, true, 0, 1, 1, 0,
		// 	0, 0, 1, 1, 0, nil, "",
		// },
		"WQ job all finished": {
			parallelism: 2, completions: -1, backoffLimit: 6, deleting: false, podLimit: 0,
			podControllerError: nil, jobKeyForget: true, pendingPods: 0, activePods: 0, succeededPods: 2, failedPods: 0,
			expectedCreations: 0, expectedDeletions: 0, expectedActive: 0, expectedSucceeded: 2, expectedFailed: 0, expectedCondition: &jobConditionComplete, expectedConditionStatus: v1.ConditionTrue, expectedConditionReason: "",
		},
		"WQ job all finished despite one failure": {
			parallelism: 2, completions: -1, backoffLimit: 6, deleting: false, podLimit: 0,
			podControllerError: nil, jobKeyForget: true, pendingPods: 0, activePods: 0, succeededPods: 1, failedPods: 1,
			expectedCreations: 0, expectedDeletions: 0, expectedActive: 0, expectedSucceeded: 1, expectedFailed: 1, expectedCondition: &jobConditionComplete, expectedConditionStatus: v1.ConditionTrue, expectedConditionReason: "",
		},
		"more active pods than completions": {
			parallelism: 2, completions: 5, backoffLimit: 6, deleting: false, podLimit: 0,
			podControllerError: nil, jobKeyForget: true, pendingPods: 0, activePods: 10, succeededPods: 0, failedPods: 0,
			expectedCreations: 0, expectedDeletions: 8, expectedActive: 2, expectedSucceeded: 0, expectedFailed: 0, expectedCondition: nil, expectedConditionReason: "",
		},
		"status change": {
			parallelism: 2, completions: 5, backoffLimit: 6, deleting: false, podLimit: 0,
			podControllerError: nil, jobKeyForget: true, pendingPods: 0, activePods: 2, succeededPods: 2, failedPods: 0,
			expectedCreations: 0, expectedDeletions: 0, expectedActive: 2, expectedSucceeded: 2, expectedFailed: 0, expectedCondition: nil, expectedConditionReason: "",
		},
		"deleting job": {
			parallelism: 2, completions: 5, backoffLimit: 6, deleting: true, podLimit: 0,
			podControllerError: nil, jobKeyForget: true, pendingPods: 1, activePods: 1, succeededPods: 1, failedPods: 0,
			expectedCreations: 0, expectedDeletions: 0, expectedActive: 2, expectedSucceeded: 1, expectedFailed: 0, expectedCondition: nil, expectedConditionReason: "",
		},
		"limited pods": {
			parallelism: 100, completions: 200, backoffLimit: 6, deleting: false, podLimit: 10,
			podControllerError: nil, jobKeyForget: true, pendingPods: 0, activePods: 0, succeededPods: 0, failedPods: 0,
			expectedCreations: 10, expectedDeletions: 0, expectedActive: 10, expectedSucceeded: 0, expectedFailed: 0, expectedCondition: nil, expectedConditionReason: "",
		},
		"too many job failures": {
			parallelism: 2, completions: 5, backoffLimit: 0, deleting: true, podLimit: 0,
			podControllerError: nil, jobKeyForget: true, pendingPods: 0, activePods: 0, succeededPods: 0, failedPods: 1,
			expectedCreations: 0, expectedDeletions: 0, expectedActive: 0, expectedSucceeded: 0, expectedFailed: 1, expectedCondition: &jobConditionFailed, expectedConditionStatus: v1.ConditionTrue, expectedConditionReason: "BackoffLimitExceeded",
		},
		"suspending a job with satisfied expectations": {
			suspend: true, parallelism: 2, activePods: 2, completions: 4, backoffLimit: 6, jobKeyForget: true,
			expectedCreations: 0, expectedActive: 0, expectedDeletions: 2, expectedCondition: &jobConditionSuspended, expectedConditionStatus: v1.ConditionTrue, expectedConditionReason: JobSuspendedReason,
		},
		"suspending a job with unsatisfied expectations": {
			// Unlike the previous test, we expect the controller to NOT suspend the
			// Job in the syncJob call because the controller will wait for
			// expectations to be satisfied first. The next syncJob call (not tested
			// here) will be the same as the previous test.
			suspend: true, parallelism: 2, activePods: 3, fakeExpectationAtCreation: -1, completions: 4, backoffLimit: 6, jobKeyForget: true,
			expectedCreations: 0, expectedActive: 3,
		},
		"resuming a suspended job": {
			wasSuspended: true, suspend: false, parallelism: 2, activePods: 0, completions: 4, backoffLimit: 6, jobKeyForget: true,
			expectedCreations: 2, expectedActive: 2, expectedCondition: &jobConditionSuspended, expectedConditionStatus: v1.ConditionFalse, expectedConditionReason: "JobResumed",
		},
		"suspending a deleted job": {
			// The job is already deleting so nothing will be done
			deleting: true, suspend: true, parallelism: 2, activePods: 2, completions: 4, backoffLimit: 6, jobKeyForget: true,
			expectedCreations: 0, expectedActive: 2, expectedDeletions: 0,
		},
	}

	for name, tc := range testCases {
		// job manager setup
		manager, sharedInformerFactory, runaiSharedInformersFactory := newController()
		manager.podControl = &controller.FakePodControl{Err: tc.podControllerError, CreateLimit: tc.podLimit}
		fakePodControl := FakePodControl{Err: tc.podControllerError, CreateLimit: tc.podLimit}
		manager.control = &fakePodControl
		manager.configMapSynced = alwaysReady
		manager.recorder = record.NewFakeRecorder(1024)
		var actual *runaijobv1.RunaiJob

		manager.updateHandler = func(job *runaijobv1.RunaiJob) error {
			actual = job
			return nil
		}

		// job & pods setup
		job := newJob(tc.parallelism, tc.completions, tc.backoffLimit)
		job.Spec.Suspend = pointer.BoolPtr(tc.suspend)
		key := client.ObjectKeyFromObject(job)

		if tc.fakeExpectationAtCreation < 0 {
			if expectationsErr := manager.expectations.ExpectDeletions(klog.FromContext(context.Background()),
				key.String(), int(-tc.fakeExpectationAtCreation)); expectationsErr != nil {
				t.Errorf("Failed to set expectedDeletions: %v", expectationsErr)
			}
		} else if tc.fakeExpectationAtCreation > 0 {
			if expectationsErr := manager.expectations.ExpectCreations(klog.FromContext(context.Background()),
				key.String(), int(tc.fakeExpectationAtCreation)); expectationsErr != nil {
				t.Errorf("Failed to set expectedDeletions: %v", expectationsErr)
			}
		}
		if tc.wasSuspended {
			job.Status.Conditions = append(job.Status.Conditions, runaiJobUtils.NewCondition(runaijobv1.JobSuspended, v1.ConditionTrue, JobSuspendedReason, jobSuspendedMessage))
		}

		if tc.deleting {
			now := metav1.Now()
			job.DeletionTimestamp = &now
		}

		klog.V(3).Infof("Initializing Runai Job Controller")
		err := runaiSharedInformersFactory.Run().V1().RunaiJobs().Informer().GetIndexer().Add(job)
		if err != nil {
			klog.Errorf("Failed to add runaiJob, error: <%v>", err)
		}

		podIndexer := sharedInformerFactory.Core().V1().Pods().Informer().GetIndexer()

		configMap := setPodsStatusesAndGetConfigMap(podIndexer, job, tc.pendingPods, tc.activePods, tc.succeededPods, tc.failedPods)
		err = sharedInformerFactory.Core().V1().ConfigMaps().Informer().GetIndexer().Add(configMap)
		if err != nil {
			klog.Errorf("Failed to add configmap, error: <%v>", err)
		}

		// key := GetKey(job)

		// run
		forget, err := manager.syncJob(ctrl.Request{NamespacedName: key})

		// We need requeue syncJob task if podController error
		if tc.podControllerError != nil {
			if err == nil {
				t.Errorf("%s: Syncing jobs would return error when podController exception", name)
			}
		} else {
			if err != nil && (tc.podLimit == 0 || fakePodControl.CreateCallCount < tc.podLimit) {
				t.Errorf("%s: unexpected error when syncing jobs %v", name, err)
			}
		}
		if forget != tc.jobKeyForget {
			t.Errorf("%s: unexpected forget value. Expected %v, saw %v\n", name, tc.jobKeyForget, forget)
		}
		// validate created/deleted pods
		if int32(len(fakePodControl.Templates)) != tc.expectedCreations {
			t.Errorf("%s: unexpected number of creates.  Expected %d, saw %d\n", name, tc.expectedCreations, len(fakePodControl.Templates))
		}
		if int32(len(fakePodControl.DeletePodName)) != tc.expectedDeletions {
			t.Errorf("%s: unexpected number of deletes.  Expected %d, saw %d\n", name, tc.expectedDeletions, len(fakePodControl.DeletePodName))
		}
		// Each create should have an accompanying ControllerRef.
		if len(fakePodControl.ControllerRefs) != int(tc.expectedCreations) {
			t.Errorf("%s: unexpected number of ControllerRefs.  Expected %d, saw %d\n", name, tc.expectedCreations, len(fakePodControl.ControllerRefs))
		}
		// Make sure the ControllerRefs are correct.
		for _, controllerRef := range fakePodControl.ControllerRefs {
			if got, want := controllerRef.APIVersion, "run.ai/v1"; got != want {
				t.Errorf("controllerRef.APIVersion = %q, want %q", got, want)
			}
			if got, want := controllerRef.Kind, "RunaiJob"; got != want {
				t.Errorf("controllerRef.Kind = %q, want %q", got, want)
			}
			if got, want := controllerRef.Name, job.Name; got != want {
				t.Errorf("controllerRef.Name = %q, want %q", got, want)
			}
			if got, want := controllerRef.UID, job.UID; got != want {
				t.Errorf("controllerRef.UID = %q, want %q", got, want)
			}
			if controllerRef.Controller == nil || *controllerRef.Controller != true {
				t.Errorf("controllerRef.Controller is not set to true")
			}
		}
		// validate status
		if actual.Status.Active != tc.expectedActive {
			t.Errorf("%s: unexpected number of active pods.  Expected %d, saw %d\n", name, tc.expectedActive, actual.Status.Active)
		}
		if actual.Status.Succeeded != tc.expectedSucceeded {
			t.Errorf("%s: unexpected number of succeeded pods.  Expected %d, saw %d\n", name, tc.expectedSucceeded, actual.Status.Succeeded)
		}
		if actual.Status.Failed != tc.expectedFailed {
			t.Errorf("%s: unexpected number of failed pods.  Expected %d, saw %d\n", name, tc.expectedFailed, actual.Status.Failed)
		}
		if actual.Status.StartTime != nil && tc.suspend {
			t.Errorf("%s: Unexpected .status.startTime not nil when suspend is true", name)
		}
		if actual.Status.StartTime == nil && !tc.suspend {
			t.Errorf("%s: .status.startTime was not set", name)
		}
		// validate conditions
		if tc.expectedCondition != nil && !getCondition(actual, *tc.expectedCondition, tc.expectedConditionStatus, tc.expectedConditionReason) {
			t.Errorf("%s: expected completion condition.  Got %#v", name, actual.Status.Conditions)
		}
		if tc.expectedCondition == nil && tc.suspend && len(actual.Status.Conditions) != 0 {
			t.Errorf("%s: Unexpected conditions: %v", name, actual.Status.Conditions)
		}
		// validate slow start
		expectedLimit := 0
		for pass := uint8(0); expectedLimit <= tc.podLimit; pass++ {
			expectedLimit += controller.SlowStartInitialBatchSize << pass
		}
		if tc.podLimit > 0 && fakePodControl.CreateCallCount > expectedLimit {
			t.Errorf("%s: Unexpected number of create calls.  Expected <= %d, saw %d\n", name, fakePodControl.CreateLimit*2, fakePodControl.CreateCallCount)
		}
	}
}

func TestSyncPastDeadlineJobFinished(t *testing.T) {
	manager, _, runaiSharedInformersFactory := newController()
	fakePodControl := controller.FakePodControl{}
	manager.podControl = &fakePodControl
	var actual *runaijobv1.RunaiJob
	manager.updateHandler = func(job *runaijobv1.RunaiJob) error {
		actual = job
		return nil
	}

	job := newJob(1, 1, 6)
	start := metav1.Unix(metav1.Now().Time.Unix()-15, 0)
	job.Status.StartTime = &start
	job.Status.Conditions = append(job.Status.Conditions, runaiJobUtils.NewCondition(runaijobv1.JobFailed, v1.ConditionTrue, "DeadlineExceeded", "RunaiJob was active longer than specified deadline"))
	err := runaiSharedInformersFactory.Run().V1().RunaiJobs().Informer().GetIndexer().Add(job)
	if err != nil {
		klog.Errorf("Failed to add runaiJob, error: <%v>", err)
	}
	forget, err := manager.syncJob(GetKey(job))
	if err != nil {
		t.Errorf("Unexpected error when syncing jobs %v", err)
	}
	if !forget {
		t.Errorf("Unexpected forget value. Expected %v, saw %v\n", true, forget)
	}
	if len(fakePodControl.Templates) != 0 {
		t.Errorf("Unexpected number of creates.  Expected %d, saw %d\n", 0, len(fakePodControl.Templates))
	}
	if len(fakePodControl.DeletePodName) != 0 {
		t.Errorf("Unexpected number of deletes.  Expected %d, saw %d\n", 0, len(fakePodControl.DeletePodName))
	}
	if actual != nil {
		t.Error("Unexpected job modification")
	}
}

func TestSyncJobComplete(t *testing.T) {
	manager, _, runaiSharedInformersFactory := newController()
	fakePodControl := controller.FakePodControl{}
	manager.podControl = &fakePodControl

	job := newJob(1, 1, 6)
	job.Status.Conditions = append(job.Status.Conditions, runaiJobUtils.NewCondition(runaijobv1.JobComplete, v1.ConditionTrue, "", ""))
	err := runaiSharedInformersFactory.Run().V1().RunaiJobs().Informer().GetIndexer().Add(job)
	if err != nil {
		klog.Errorf("Failed to add runaiJob, error: <%v>", err)
	}
	forget, err := manager.syncJob(GetKey(job))
	if err != nil {
		t.Fatalf("Unexpected error when syncing jobs %v", err)
	}
	if !forget {
		t.Errorf("Unexpected forget value. Expected %v, saw %v\n", true, forget)
	}
	actual, err := manager.jobLister.RunaiJobs(job.Namespace).Get(job.Name)
	if err != nil {
		t.Fatalf("Unexpected error when trying to get job from the store: %v", err)
	}
	// Verify that after syncing a complete job, the conditions are the same.
	if got, expected := len(actual.Status.Conditions), 1; got != expected {
		t.Fatalf("Unexpected job status conditions amount; expected %d, got %d", expected, got)
	}
}

func TestSyncJobDeleted(t *testing.T) {
	manager, _, _ := newController()
	fakePodControl := controller.FakePodControl{}
	manager.podControl = &fakePodControl
	manager.updateHandler = func(job *runaijobv1.RunaiJob) error { return nil }
	job := newJob(2, 2, 6)
	forget, err := manager.syncJob(GetKey(job))
	if err != nil {
		t.Errorf("Unexpected error when syncing jobs %v", err)
	}
	if !forget {
		t.Errorf("Unexpected forget value. Expected %v, saw %v\n", true, forget)
	}
	if len(fakePodControl.Templates) != 0 {
		t.Errorf("Unexpected number of creates.  Expected %d, saw %d\n", 0, len(fakePodControl.Templates))
	}
	if len(fakePodControl.DeletePodName) != 0 {
		t.Errorf("Unexpected number of deletes.  Expected %d, saw %d\n", 0, len(fakePodControl.DeletePodName))
	}
}

func TestSyncJobUpdateRequeue(t *testing.T) {
	DefaultJobBackOff = time.Duration(0) // overwrite the default value for testing
	manager, sharedInformersFactory, runaiSharedInformersFactory := newController()
	fakePodControl := controller.FakePodControl{}
	manager.podControl = &fakePodControl
	manager.recorder = record.NewFakeRecorder(1024)
	updateError := fmt.Errorf("update error")
	manager.updateHandler = func(job *runaijobv1.RunaiJob) error {
		manager.queue.AddRateLimited(GetKey(job))
		return updateError
	}
	job := newJob(2, 2, 6)
	err := runaiSharedInformersFactory.Run().V1().RunaiJobs().Informer().GetIndexer().Add(job)
	if err != nil {
		klog.Errorf("Failed to add runaiJob, error: <%v>", err)
	}
	configmap := v1.ConfigMap{
		TypeMeta: metav1.TypeMeta{Kind: "ConfigMap", APIVersion: "v1"},
		ObjectMeta: metav1.ObjectMeta{
			Name:            fmt.Sprintf("%s-pod-attempts", job.Name),
			Namespace:       job.Namespace,
			UID:             uuid.NewUUID(),
			ResourceVersion: "1",
			OwnerReferences: []metav1.OwnerReference{*metav1.NewControllerRef(job, controllerKind)},
		},
		Data: make(map[string]string),
	}
	err = sharedInformersFactory.Core().V1().ConfigMaps().Informer().GetIndexer().Add(&configmap)
	if err != nil {
		klog.Errorf("Failed to add configmap, error: <%v>", err)
	}

	forget, err := manager.syncJob(GetKey(job))
	if err == nil || err != updateError {
		t.Errorf("Expected error %v when syncing jobs, got %v", updateError, err)
	}
	if forget != false {
		t.Errorf("Unexpected forget value. Expected %v, saw %v\n", false, forget)
	}
	t.Log("Waiting for a job in the queue")
	actualKey, _ := manager.queue.Get()
	expectedKey := GetKey(job)
	if actualKey != expectedKey {
		t.Errorf("Expected requeue of job with key %s got %s", expectedKey, actualKey)
	}
}

func TestSyncJobUpdateJobResources(t *testing.T) {
	testCases := map[string]struct {
		jobGpuFraction       float64
		podGpuFraction       float64
		jobRes               []v1.ResourceRequirements
		podRes               []v1.ResourceRequirements
		additionalActivePods int // additional active pods (with job-matching spec) besides the pod with the unmatching spec
		expectedDeletions    int32
		parallelism          int32
	}{
		"identical resources for job and pod": {
			jobRes: []v1.ResourceRequirements{{
				Limits: v1.ResourceList{
					GPUResourceName: *resource.NewQuantity(1, resource.DecimalSI),
				},
				Requests: v1.ResourceList{
					v1.ResourceCPU:    *resource.NewQuantity(1, resource.DecimalSI),
					v1.ResourceMemory: *resource.NewQuantity(1e9, resource.DecimalSI),
				}},
			},
			podRes: []v1.ResourceRequirements{{
				Limits: v1.ResourceList{
					GPUResourceName: *resource.NewQuantity(1, resource.DecimalSI),
				},
				Requests: v1.ResourceList{
					v1.ResourceCPU:    *resource.NewQuantity(1, resource.DecimalSI),
					v1.ResourceMemory: *resource.NewQuantity(1e9, resource.DecimalSI),
				}},
			},
			expectedDeletions: 0,
		},
		"equal resources for job and pod but different cpu scale": {
			jobRes: []v1.ResourceRequirements{{
				Limits: v1.ResourceList{
					GPUResourceName: *resource.NewQuantity(1, resource.DecimalSI),
				},
				Requests: v1.ResourceList{
					v1.ResourceCPU:    *resource.NewMilliQuantity(1e3, resource.DecimalSI),
					v1.ResourceMemory: *resource.NewQuantity(1e9, resource.DecimalSI),
				}},
			},
			podRes: []v1.ResourceRequirements{{
				Limits: v1.ResourceList{
					GPUResourceName: *resource.NewQuantity(1, resource.DecimalSI),
				},
				Requests: v1.ResourceList{
					v1.ResourceCPU:    *resource.NewQuantity(1, resource.DecimalSI),
					v1.ResourceMemory: *resource.NewQuantity(1e9, resource.DecimalSI),
				},
			}},
			expectedDeletions: 0,
		},
		"different resources for job and pod": {
			jobRes: []v1.ResourceRequirements{{
				Limits: v1.ResourceList{
					GPUResourceName: *resource.NewQuantity(1, resource.DecimalSI),
				},
				Requests: v1.ResourceList{
					v1.ResourceCPU:    *resource.NewMilliQuantity(1e3, resource.DecimalSI),
					v1.ResourceMemory: *resource.NewQuantity(1e9, resource.DecimalSI),
				},
			}},
			podRes: []v1.ResourceRequirements{{
				Limits: v1.ResourceList{
					GPUResourceName: *resource.NewQuantity(2, resource.DecimalSI),
				},
				Requests: v1.ResourceList{
					v1.ResourceCPU:    *resource.NewMilliQuantity(2e3, resource.DecimalSI),
					v1.ResourceMemory: *resource.NewQuantity(2e9, resource.DecimalSI),
				}},
			},
			expectedDeletions: 1,
		},
		"missing pod limits": {
			jobRes: []v1.ResourceRequirements{{
				Limits: v1.ResourceList{
					GPUResourceName:   *resource.NewQuantity(2, resource.DecimalSI),
					v1.ResourceCPU:    *resource.NewMilliQuantity(2e3, resource.DecimalSI),
					v1.ResourceMemory: *resource.NewQuantity(2e9, resource.DecimalSI),
				},
			}},
			podRes: []v1.ResourceRequirements{{
				Limits: v1.ResourceList{
					GPUResourceName: *resource.NewQuantity(2, resource.DecimalSI),
					v1.ResourceCPU:  *resource.NewMilliQuantity(2e3, resource.DecimalSI),
				},
			}},
			expectedDeletions: 1,
		},
		"missing job limits": {
			jobRes: []v1.ResourceRequirements{{
				Limits: v1.ResourceList{
					GPUResourceName: *resource.NewQuantity(2, resource.DecimalSI),
				},
			}},
			podRes: []v1.ResourceRequirements{{
				Limits: v1.ResourceList{
					GPUResourceName: *resource.NewQuantity(2, resource.DecimalSI),
					v1.ResourceCPU:  *resource.NewMilliQuantity(2e3, resource.DecimalSI),
				},
			}},
			expectedDeletions: 0,
		},
		"different gpu fraction for job and pod": {
			jobGpuFraction: 0.5,
			podRes: []v1.ResourceRequirements{{
				Limits: v1.ResourceList{
					GPUResourceName: *resource.NewQuantity(1, resource.DecimalSI),
				}},
			},
			expectedDeletions: 1,
		},
		"Multiple containers with identical res": {
			jobGpuFraction: 0.5,
			podGpuFraction: 0.5,
			jobRes: []v1.ResourceRequirements{{
				Requests: v1.ResourceList{
					v1.ResourceCPU:    *resource.NewMilliQuantity(1e3, resource.DecimalSI),
					v1.ResourceMemory: *resource.NewQuantity(1e9, resource.DecimalSI),
				}}, {
				Requests: v1.ResourceList{
					v1.ResourceCPU:    *resource.NewMilliQuantity(1e3, resource.DecimalSI),
					v1.ResourceMemory: *resource.NewQuantity(1e9, resource.DecimalSI),
				}},
			},
			podRes: []v1.ResourceRequirements{{
				Requests: v1.ResourceList{
					v1.ResourceCPU:    *resource.NewMilliQuantity(1e3, resource.DecimalSI),
					v1.ResourceMemory: *resource.NewQuantity(1e9, resource.DecimalSI),
				}}, {
				Requests: v1.ResourceList{
					v1.ResourceCPU:    *resource.NewMilliQuantity(1e3, resource.DecimalSI),
					v1.ResourceMemory: *resource.NewQuantity(1e9, resource.DecimalSI),
				}},
			},
			expectedDeletions: 0,
		},
		"Multiple containers with different res": {
			jobRes: []v1.ResourceRequirements{{
				Limits: v1.ResourceList{
					GPUResourceName: *resource.NewQuantity(1, resource.DecimalSI),
				},
				Requests: v1.ResourceList{
					v1.ResourceCPU:    *resource.NewMilliQuantity(1e3, resource.DecimalSI),
					v1.ResourceMemory: *resource.NewQuantity(1e9, resource.DecimalSI),
				}}, {
				Limits: v1.ResourceList{
					GPUResourceName: *resource.NewQuantity(2, resource.DecimalSI),
				},
				Requests: v1.ResourceList{
					v1.ResourceCPU:    *resource.NewMilliQuantity(1e3, resource.DecimalSI),
					v1.ResourceMemory: *resource.NewQuantity(1e9, resource.DecimalSI),
				}},
			},
			podRes: []v1.ResourceRequirements{{
				Limits: v1.ResourceList{
					GPUResourceName: *resource.NewQuantity(1, resource.DecimalSI),
				},
				Requests: v1.ResourceList{
					v1.ResourceCPU:    *resource.NewMilliQuantity(1e3, resource.DecimalSI),
					v1.ResourceMemory: *resource.NewQuantity(1e9, resource.DecimalSI),
				}}, {
				Limits: v1.ResourceList{
					GPUResourceName: *resource.NewQuantity(1, resource.DecimalSI),
				},
				Requests: v1.ResourceList{
					v1.ResourceCPU:    *resource.NewMilliQuantity(1e3, resource.DecimalSI),
					v1.ResourceMemory: *resource.NewQuantity(1e9, resource.DecimalSI),
				}},
			},
			expectedDeletions: 1,
		},
		"Different number of containers": {
			jobRes: []v1.ResourceRequirements{{
				Limits: v1.ResourceList{
					GPUResourceName: *resource.NewQuantity(2, resource.DecimalSI),
				},
				Requests: v1.ResourceList{
					v1.ResourceCPU:    *resource.NewMilliQuantity(1e3, resource.DecimalSI),
					v1.ResourceMemory: *resource.NewQuantity(1e9, resource.DecimalSI),
				}},
			},
			podRes: []v1.ResourceRequirements{{
				Limits: v1.ResourceList{
					GPUResourceName: *resource.NewQuantity(1, resource.DecimalSI),
				},
				Requests: v1.ResourceList{
					v1.ResourceCPU:    *resource.NewMilliQuantity(1e3, resource.DecimalSI),
					v1.ResourceMemory: *resource.NewQuantity(1e9, resource.DecimalSI),
				}}, {
				Limits: v1.ResourceList{
					GPUResourceName: *resource.NewQuantity(1, resource.DecimalSI),
				},
				Requests: v1.ResourceList{
					v1.ResourceCPU:    *resource.NewMilliQuantity(1e3, resource.DecimalSI),
					v1.ResourceMemory: *resource.NewQuantity(1e9, resource.DecimalSI),
				}},
			},
			expectedDeletions: 1,
		},
		"Multiple pods": {
			jobRes: []v1.ResourceRequirements{{
				Limits: v1.ResourceList{
					GPUResourceName: *resource.NewQuantity(1, resource.DecimalSI),
				},
				Requests: v1.ResourceList{
					v1.ResourceCPU:    *resource.NewQuantity(1, resource.DecimalSI),
					v1.ResourceMemory: *resource.NewQuantity(1e9, resource.DecimalSI),
				}},
			},
			podRes: []v1.ResourceRequirements{{
				Limits: v1.ResourceList{
					GPUResourceName: *resource.NewQuantity(1, resource.DecimalSI),
				},
				Requests: v1.ResourceList{
					v1.ResourceCPU:    *resource.NewQuantity(1, resource.DecimalSI),
					v1.ResourceMemory: *resource.NewQuantity(1e9, resource.DecimalSI),
				}},
			},
			additionalActivePods: 4,
			expectedDeletions:    0,
			parallelism:          5,
		},
		"1 pods with unmatch spec and 1 pod above parallelism": {
			jobRes: []v1.ResourceRequirements{{
				Limits: v1.ResourceList{
					GPUResourceName: *resource.NewQuantity(1, resource.DecimalSI),
				},
				Requests: v1.ResourceList{
					v1.ResourceCPU:    *resource.NewQuantity(1, resource.DecimalSI),
					v1.ResourceMemory: *resource.NewQuantity(1e9, resource.DecimalSI),
				}},
			},
			podRes: []v1.ResourceRequirements{{
				Limits: v1.ResourceList{
					GPUResourceName: *resource.NewQuantity(2, resource.DecimalSI),
				},
				Requests: v1.ResourceList{
					v1.ResourceCPU:    *resource.NewQuantity(2, resource.DecimalSI),
					v1.ResourceMemory: *resource.NewQuantity(2e9, resource.DecimalSI),
				}},
			},
			additionalActivePods: 2,
			expectedDeletions:    2,
		},
	}
	for name, tc := range testCases {
		// job controller setup
		manager, sharedInformerFactory, runaiSharedInformersFactory := newController()
		fakePodControl := controller.FakePodControl{}
		manager.podControl = &fakePodControl
		fakeControl := &FakePodControl{}
		manager.control = fakeControl
		manager.recorder = record.NewFakeRecorder(1024)
		manager.updateHandler = func(job *runaijobv1.RunaiJob) error { return nil }

		// creating the job
		parallelism := tc.parallelism
		if parallelism == 0 {
			parallelism = 1
		}
		job := newJob(parallelism, 5, 6)
		for i, v := range tc.jobRes {
			if len(job.Spec.Template.Spec.Containers) > i {
				job.Spec.Template.Spec.Containers[i].Resources = *v.DeepCopy()
			} else {
				job.Spec.Template.Spec.Containers = append(job.Spec.Template.Spec.Containers,
					v1.Container{Image: "foo/bar", Resources: *v.DeepCopy()})
			}
		}
		if tc.jobGpuFraction > 0 {
			if job.Spec.Template.ObjectMeta.Annotations == nil {
				job.Spec.Template.ObjectMeta.Annotations = make(map[string]string)
			}
			job.Spec.Template.ObjectMeta.Annotations[GPUFractionResourceName] = strconv.FormatFloat(tc.jobGpuFraction, 'f', -1, 64)
		}
		err := runaiSharedInformersFactory.Run().V1().RunaiJobs().Informer().GetIndexer().Add(job)
		if err != nil {
			klog.Errorf("Failed to add new job, error: <%v>", err)
		}

		// creating pods with the pod resource (that might be different than the job's resources and then it should be
		// deleted)
		podSuffix := "0-0"
		pod := newPod(podSuffix, job, "Running", createCondition(CreateConditionRunningTrue))
		for i, v := range tc.podRes {
			if len(pod.Spec.Containers) > i {
				pod.Spec.Containers[i].Resources = *v.DeepCopy()
			} else {
				pod.Spec.Containers = append(pod.Spec.Containers,
					v1.Container{Resources: *v.DeepCopy()})
			}
		}
		if tc.podGpuFraction > 0 {
			if pod.Annotations == nil {
				pod.Annotations = make(map[string]string)
			}
			pod.Annotations[GPUFractionResourceName] = strconv.FormatFloat(tc.podGpuFraction, 'f', -1, 64)
		}
		podIndexer := sharedInformerFactory.Core().V1().Pods().Informer().GetIndexer()
		err = podIndexer.Add(pod)
		if err != nil {
			klog.Errorf("failed to add pod to new job")
		}
		configMap := setPodsStatusesAndGetConfigMap(podIndexer, job, 0, 0, 0, 0)
		configMap.Data[podSuffix] = "Running"
		for i := 1; i <= tc.additionalActivePods; i++ {
			podSuffix := fmt.Sprintf("%d-0", i)
			pod := newPod(podSuffix, job, "Running", createCondition(CreateConditionRunningTrue))
			for i, v := range tc.jobRes {
				if len(pod.Spec.Containers) > i {
					pod.Spec.Containers[i].Resources = *v.DeepCopy()
				} else {
					pod.Spec.Containers = append(pod.Spec.Containers,
						v1.Container{Resources: *v.DeepCopy()})
				}
			}
			if tc.jobGpuFraction > 0 {
				if pod.Annotations == nil {
					pod.Annotations = make(map[string]string)
				}
				pod.Annotations[GPUFractionResourceName] = strconv.FormatFloat(tc.podGpuFraction, 'f', -1, 64)
			}
			podIndexer := sharedInformerFactory.Core().V1().Pods().Informer().GetIndexer()
			err = podIndexer.Add(pod)
			if err != nil {
				klog.Errorf("failed to add pod to new job")
			}
			configMap.Data[podSuffix] = "Running"
		}
		err = sharedInformerFactory.Core().V1().ConfigMaps().Informer().GetIndexer().Add(configMap)
		if err != nil {
			klog.Errorf("Failed to add configmap, error: <%v>", err)
		}

		//run
		_, err = manager.syncJob(GetKey(job))
		if err != nil {
			t.Errorf("Error syncing job")
		}

		if int32(len(fakeControl.DeletePodName)) != tc.expectedDeletions {
			t.Errorf("%s: unexpected number of deletes.  Expected %d, saw %d\n",
				name, tc.expectedDeletions, len(fakeControl.DeletePodName))
		}
	}
}

func TestJobPodLookup(t *testing.T) {
	manager, _, runaiSharedInformersFactory := newController()
	testCases := []struct {
		job *runaijobv1.RunaiJob
		pod *v1.Pod

		expectedName string
	}{
		// pods without labels don't match any job
		{
			job: &runaijobv1.RunaiJob{
				ObjectMeta: metav1.ObjectMeta{Name: "basic"},
			},
			pod: &v1.Pod{
				ObjectMeta: metav1.ObjectMeta{Name: "foo1", Namespace: metav1.NamespaceAll},
			},
			expectedName: "",
		},
		// matching labels, different namespace
		{
			job: &runaijobv1.RunaiJob{
				ObjectMeta: metav1.ObjectMeta{Name: "foo"},
				Spec: runaijobv1.JobSpec{
					Selector: &metav1.LabelSelector{
						MatchLabels: map[string]string{"foo": "bar"},
					},
				},
			},
			pod: &v1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "foo2",
					Namespace: "ns",
					Labels:    map[string]string{"foo": "bar"},
				},
			},
			expectedName: "",
		},
		// matching ns and labels returns
		{
			job: &runaijobv1.RunaiJob{
				ObjectMeta: metav1.ObjectMeta{Name: "bar", Namespace: "ns"},
				Spec: runaijobv1.JobSpec{
					Selector: &metav1.LabelSelector{
						MatchExpressions: []metav1.LabelSelectorRequirement{
							{
								Key:      "foo",
								Operator: metav1.LabelSelectorOpIn,
								Values:   []string{"bar"},
							},
						},
					},
				},
			},
			pod: &v1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "foo3",
					Namespace: "ns",
					Labels:    map[string]string{"foo": "bar"},
				},
			},
			expectedName: "bar",
		},
	}
	for _, tc := range testCases {
		err := runaiSharedInformersFactory.Run().V1().RunaiJobs().Informer().GetIndexer().Add(tc.job)
		if err != nil {
			klog.Errorf("Failed to add runaiJob, error: <%v>", err)
		}
		if jobs := manager.getPodJobs(tc.pod); len(jobs) > 0 {
			if got, want := len(jobs), 1; got != want {
				t.Errorf("len(jobs) = %v, want %v", got, want)
			}
			job := jobs[0]
			if tc.expectedName != job.Name {
				t.Errorf("Got job %+v expected %+v", job.Name, tc.expectedName)
			}
		} else if tc.expectedName != "" {
			t.Errorf("Expected a job %v pod %v, found none", tc.expectedName, tc.pod.Name)
		}
	}
}

func TestGetPodsForJob(t *testing.T) {
	manager, sharedInformerFactory, runaiSharedInformersFactory := newController()

	job1 := newJob(1, 1, 6)
	job1.Name = "job1"
	job2 := newJob(1, 1, 6)
	job2.Name = "job2"
	err := runaiSharedInformersFactory.Run().V1().RunaiJobs().Informer().GetIndexer().Add(job1)
	if err != nil {
		if err != nil {
			klog.Errorf("Failed to add runaiJob, error: <%v>", err)
		}
	}

	err = runaiSharedInformersFactory.Run().V1().RunaiJobs().Informer().GetIndexer().Add(job2)
	if err != nil {
		klog.Errorf("failed to add job, error: <%v>", err)
	}

	podIndexer := sharedInformerFactory.Core().V1().Pods().Informer().GetIndexer()
	configMap := setPodsStatusesAndGetConfigMap(podIndexer, job1, 0, 1, 0, 0)

	// Make pod3 an orphan that doesn't match. It should be ignored.
	podSuffix := "1-0"
	configMap.Data[podSuffix] = "Running"
	pod3 := newPod(podSuffix, job1, "Running", createCondition(CreateConditionRunningTrue))
	pod3.OwnerReferences = nil
	pod3.Labels = nil
	err = podIndexer.Add(pod3)
	if err != nil {
		klog.Errorf("failed to add pod")
	}
	err = sharedInformerFactory.Core().V1().ConfigMaps().Informer().GetIndexer().Add(configMap)
	if err != nil {
		klog.Errorf("Failed to add configmap, error: <%v>", err)
	}

	configMap2 := setPodsStatusesAndGetConfigMap(podIndexer, job2, 1, 0, 0, 0)
	err = sharedInformerFactory.Core().V1().ConfigMaps().Informer().GetIndexer().Add(configMap2)
	if err != nil {
		klog.Errorf("Failed to add configmap, error: <%v>", err)
	}

	pods, err := manager.getPodsForJob(job1)
	if err != nil {
		t.Fatalf("getPodsForJob() error: %v", err)
	}
	if got, want := len(pods), 1; got != want {
		t.Errorf("len(pods) = %v, want %v", got, want)
	}
	if got, want := pods[0].Name, "job1-0-0"; got != want {
		t.Errorf("pod.Name = %v, want %v", got, want)
	}

	pods, err = manager.getPodsForJob(job2)
	if err != nil {
		t.Fatalf("getPodsForJob() error: %v", err)
	}
	if got, want := len(pods), 1; got != want {
		t.Errorf("len(pods) = %v, want %v", got, want)
	}
	if got, want := pods[0].Name, "job2-0-0"; got != want {
		t.Errorf("pod.Name = %v, want %v", got, want)
	}
}

func TestGetPodsForJobAdopt(t *testing.T) {
	job1 := newJob(1, 1, 6)
	job1.Name = "job1"
	jm, sharedInformerFactory, runaiSharedInformersFactory := newControllerWithRunaiClientset(job1)

	err := runaiSharedInformersFactory.Run().V1().RunaiJobs().Informer().GetIndexer().Add(job1)
	if err != nil {
		klog.Errorf("failed to add job, error: <%v>", err)
	}
	podIndexer := sharedInformerFactory.Core().V1().Pods().Informer().GetIndexer()
	configMap := setPodsStatusesAndGetConfigMap(podIndexer, job1, 0, 1, 0, 0)

	// Make pod3 an orphan that doesn't match. It should be ignored.
	podSuffix := "1-0"
	configMap.Data[podSuffix] = "Running"
	pod2 := newPod(podSuffix, job1, "Running", createCondition(CreateConditionRunningTrue))
	// Make this pod an orphan. It should still be returned because it's adopted.
	pod2.OwnerReferences = nil
	err = podIndexer.Add(pod2)
	if err != nil {
		klog.Errorf("failed to add pod")
	}
	err = sharedInformerFactory.Core().V1().ConfigMaps().Informer().GetIndexer().Add(configMap)
	if err != nil {
		klog.Errorf("failed to add configmap")
	}

	pods, err := jm.getPodsForJob(job1)
	if err != nil {
		t.Fatalf("getPodsForJob() error: %v", err)
	}
	if got, want := len(pods), 2; got != want {
		t.Errorf("len(pods) = %v, want %v", got, want)
	}
}

func TestGetPodsForJobNoAdoptIfBeingDeleted(t *testing.T) {
	job1 := newJob(1, 1, 6)
	job1.Name = "job1"
	job1.DeletionTimestamp = &metav1.Time{}
	jm, sharedInformerFactory, runaiSharedInformersFactory := newControllerWithRunaiClientset(job1)

	err := runaiSharedInformersFactory.Run().V1().RunaiJobs().Informer().GetIndexer().Add(job1)
	if err != nil {
		klog.Errorf("failed to add job, error: <%v>", err)
	}

	podIndexer := sharedInformerFactory.Core().V1().Pods().Informer().GetIndexer()
	configMap := setPodsStatusesAndGetConfigMap(podIndexer, job1, 0, 1, 0, 0)

	// Make pod3 an orphan that doesn't match. It should be ignored.
	podSuffix := "1-0"
	configMap.Data[podSuffix] = "Running"
	pod2 := newPod(podSuffix, job1, "Running", createCondition(CreateConditionRunningTrue))
	// Make this pod an orphan. It should not be adopted because the RunaiJob is being deleted.
	pod2.OwnerReferences = nil
	err = podIndexer.Add(pod2)
	if err != nil {
		klog.Errorf("failed to add pod")
	}
	err = sharedInformerFactory.Core().V1().ConfigMaps().Informer().GetIndexer().Add(configMap)
	if err != nil {
		klog.Errorf("failed to add configmap")
	}

	pods, err := jm.getPodsForJob(job1)
	if err != nil {
		t.Fatalf("getPodsForJob() error: %v", err)
	}
	if got, want := len(pods), 1; got != want {
		t.Errorf("len(pods) = %v, want %v", got, want)
	}
	if got, want := pods[0].Name, fmt.Sprintf("%s-0-0", job1.Name); got != want {
		t.Errorf("pod.Name = %q, want %q", got, want)
	}
}

func TestGetPodsForJobNoAdoptIfBeingDeletedRace(t *testing.T) {
	job1 := newJob(1, 1, 6)
	job1.Name = "job1"
	// The up-to-date object says it's being deleted.
	job1.DeletionTimestamp = &metav1.Time{}
	jm, sharedInformerFactory, runaiSharedInformersFactory := newControllerWithRunaiClientset(job1)

	// The cache says it's NOT being deleted.
	cachedJob := *job1
	cachedJob.DeletionTimestamp = nil
	err := runaiSharedInformersFactory.Run().V1().RunaiJobs().Informer().GetIndexer().Add(job1)
	if err != nil {
		klog.Errorf("failed to add job, error: <%v>", err)
	}

	podIndexer := sharedInformerFactory.Core().V1().Pods().Informer().GetIndexer()
	configMap := setPodsStatusesAndGetConfigMap(podIndexer, job1, 0, 1, 0, 0)

	// Make pod3 an orphan that doesn't match. It should be ignored.
	podSuffix := "1-0"
	configMap.Data[podSuffix] = "Running"
	pod2 := newPod(podSuffix, job1, "Running", createCondition(CreateConditionRunningTrue))
	// Make this pod an orphan. It should not be adopted because the RunaiJob is being deleted.
	pod2.OwnerReferences = nil
	err = podIndexer.Add(pod2)
	if err != nil {
		klog.Errorf("failed to add pod")
	}
	err = sharedInformerFactory.Core().V1().ConfigMaps().Informer().GetIndexer().Add(configMap)
	if err != nil {
		klog.Errorf("failed to add configmap")
	}

	pods, err := jm.getPodsForJob(job1)
	if err != nil {
		t.Fatalf("getPodsForJob() error: %v", err)
	}
	if got, want := len(pods), 1; got != want {
		t.Errorf("len(pods) = %v, want %v", got, want)
	}
	if got, want := pods[0].Name, fmt.Sprintf("%s-0-0", job1.Name); got != want {
		t.Errorf("pod.Name = %q, want %q", got, want)
	}
}

func TestGetPodsForJobRelease(t *testing.T) {
	manager, sharedInformerFactory, runaiSharedInformersFactory := newController()

	job1 := newJob(1, 1, 6)
	job1.Name = "job1"
	err := runaiSharedInformersFactory.Run().V1().RunaiJobs().Informer().GetIndexer().Add(job1)
	if err != nil {
		klog.Errorf("failed to add job, error: <%v>", err)
	}

	podIndexer := sharedInformerFactory.Core().V1().Pods().Informer().GetIndexer()
	configMap := setPodsStatusesAndGetConfigMap(podIndexer, job1, 0, 1, 0, 0)

	// Make pod3 an orphan that doesn't match. It should be ignored.
	podSuffix := "1-0"
	configMap.Data[podSuffix] = "Running"
	pod2 := newPod(podSuffix, job1, "Running", createCondition(CreateConditionRunningTrue))
	// Make this pod not match, even though it's owned. It should be released.
	pod2.Labels = nil
	err = podIndexer.Add(pod2)
	if err != nil {
		klog.Errorf("failed to add pod")
	}
	err = sharedInformerFactory.Core().V1().ConfigMaps().Informer().GetIndexer().Add(configMap)
	if err != nil {
		klog.Errorf("failed to add configmap")
	}

	pods, err := manager.getPodsForJob(job1)
	if err != nil {
		t.Fatalf("getPodsForJob() error: %v", err)
	}
	if got, want := len(pods), 1; got != want {
		t.Errorf("len(pods) = %v, want %v", got, want)
	}
	if got, want := pods[0].Name, fmt.Sprintf("%s-0-0", job1.Name); got != want {
		t.Errorf("pod.Name = %v, want %v", got, want)
	}
}

func TestAddPod(t *testing.T) {
	jm, sharedInformerFactory, runaiSharedInformersFactory := newController()

	job1 := newJob(1, 1, 6)
	job1.Name = "job1"
	job2 := newJob(1, 1, 6)
	job2.Name = "job2"
	err := runaiSharedInformersFactory.Run().V1().RunaiJobs().Informer().GetIndexer().Add(job1)
	if err != nil {
		klog.Errorf("failed to add job, error: <%v>", err)
	}
	err = runaiSharedInformersFactory.Run().V1().RunaiJobs().Informer().GetIndexer().Add(job2)
	if err != nil {
		klog.Errorf("failed to add job, error: <%v>", err)
	}

	podIndexer := sharedInformerFactory.Core().V1().Pods().Informer().GetIndexer()
	configMap := setPodsStatusesAndGetConfigMap(podIndexer, job1, 0, 0, 0, 0)

	// Make pod3 an orphan that doesn't match. It should be ignored.
	podSuffix := "0-0"
	configMap.Data[podSuffix] = "Running"
	pod1 := newPod(podSuffix, job1, "Running", createCondition(CreateConditionRunningTrue))
	err = podIndexer.Add(pod1)
	if err != nil {
		klog.Errorf("failed to add pod")
	}
	err = sharedInformerFactory.Core().V1().ConfigMaps().Informer().GetIndexer().Add(configMap)
	if err != nil {
		klog.Errorf("failed to add configmap")
	}

	configMap2 := setPodsStatusesAndGetConfigMap(podIndexer, job1, 0, 0, 0, 0)

	// Make pod3 an orphan that doesn't match. It should be ignored.
	podSuffix2 := "0-0"
	configMap.Data[podSuffix2] = "Running"
	pod2 := newPod(podSuffix2, job2, "Running", createCondition(CreateConditionRunningTrue))
	err = podIndexer.Add(pod1)
	if err != nil {
		klog.Errorf("failed to add pod")
	}
	err = sharedInformerFactory.Core().V1().ConfigMaps().Informer().GetIndexer().Add(configMap2)
	if err != nil {
		klog.Errorf("failed to add configmap")
	}

	jm.addPod(pod1)
	if got, want := jm.queue.Len(), 1; got != want {
		t.Fatalf("queue.Len() = %v, want %v", got, want)
	}
	key, done := jm.queue.Get()
	if done {
		t.Fatalf("failed to enqueue controller for pod %v", pod1.Name)
	}
	expectedKey := GetKey(job1)
	if key.String() != expectedKey.String() {
		t.Errorf("queue.Get() = %s, want %s", key, expectedKey)
	}

	jm.addPod(pod2)
	if got, want := jm.queue.Len(), 1; got != want {
		t.Fatalf("queue.Len() = %v, want %v", got, want)
	}
	key, done = jm.queue.Get()
	if done {
		t.Fatalf("failed to enqueue controller for pod %v", pod2.Name)
	}
	expectedKey = GetKey(job2)
	if key.String() != expectedKey.String() {
		t.Errorf("queue.Get() = %s, want %s", key, expectedKey)
	}
}

func TestAddPodOrphan(t *testing.T) {
	jm, sharedInformerFactory, runaiSharedInformersFactory := newController()

	job1 := newJob(1, 1, 6)
	job1.Name = "job1"
	job2 := newJob(1, 1, 6)
	job2.Name = "job2"
	job3 := newJob(1, 1, 6)
	job3.Name = "job3"
	job3.Spec.Selector.MatchLabels = map[string]string{"other": "labels"}
	err := runaiSharedInformersFactory.Run().V1().RunaiJobs().Informer().GetIndexer().Add(job1)
	if err != nil {
		klog.Errorf("failed to add job, error: <%v>", err)
	}
	err = runaiSharedInformersFactory.Run().V1().RunaiJobs().Informer().GetIndexer().Add(job2)
	if err != nil {
		klog.Errorf("failed to add job, error: <%v>", err)
	}
	err = runaiSharedInformersFactory.Run().V1().RunaiJobs().Informer().GetIndexer().Add(job3)
	if err != nil {
		klog.Errorf("failed to add job, error: <%v>", err)
	}

	podIndexer := sharedInformerFactory.Core().V1().Pods().Informer().GetIndexer()
	configMap := setPodsStatusesAndGetConfigMap(podIndexer, job1, 0, 0, 0, 0)

	podSuffix := "0-0"
	configMap.Data[podSuffix] = "Running"
	pod1 := newPod(podSuffix, job1, "Running", createCondition(CreateConditionRunningTrue))
	// Make pod an orphan. Expect all matching controllers to be queued.
	pod1.OwnerReferences = nil
	err = podIndexer.Add(pod1)
	if err != nil {
		klog.Errorf("failed to add pod")
	}
	err = sharedInformerFactory.Core().V1().ConfigMaps().Informer().GetIndexer().Add(configMap)
	if err != nil {
		klog.Errorf("failed to add configmap")
	}

	jm.addPod(pod1)
	if got, want := jm.queue.Len(), 1; got != want {
		t.Fatalf("queue.Len() = %v, want %v", got, want)
	}
}

func TestUpdatePod(t *testing.T) {
	jm, sharedInformerFactory, runaiSharedInformersFactory := newController()

	job1 := newJob(1, 1, 6)
	job1.Name = "job1"
	job2 := newJob(1, 1, 6)
	job2.Name = "job2"
	err := runaiSharedInformersFactory.Run().V1().RunaiJobs().Informer().GetIndexer().Add(job1)
	if err != nil {
		klog.Errorf("failed to add job, error: <%v>", err)
	}
	err = runaiSharedInformersFactory.Run().V1().RunaiJobs().Informer().GetIndexer().Add(job2)
	if err != nil {
		klog.Errorf("failed to add job, error: <%v>", err)
	}

	podIndexer := sharedInformerFactory.Core().V1().Pods().Informer().GetIndexer()
	configMap := setPodsStatusesAndGetConfigMap(podIndexer, job1, 0, 0, 0, 0)

	podSuffix := "0-0"
	configMap.Data[podSuffix] = "Running"
	pod1 := newPod(podSuffix, job1, "Running", createCondition(CreateConditionRunningTrue))
	err = podIndexer.Add(pod1)
	if err != nil {
		klog.Errorf("failed to add pod")
	}
	err = sharedInformerFactory.Core().V1().ConfigMaps().Informer().GetIndexer().Add(configMap)
	if err != nil {
		klog.Errorf("failed to add configmap")
	}
	configMap2 := setPodsStatusesAndGetConfigMap(podIndexer, job2, 0, 0, 0, 0)

	podSuffix2 := "0-0"
	configMap2.Data[podSuffix2] = "Running"
	pod2 := newPod(podSuffix2, job2, "Running", createCondition(CreateConditionRunningTrue))
	err = podIndexer.Add(pod2)
	if err != nil {
		klog.Errorf("failed to add pod")
	}
	err = sharedInformerFactory.Core().V1().ConfigMaps().Informer().GetIndexer().Add(configMap2)
	if err != nil {
		klog.Errorf("failed to add configmap")
	}

	prev := *pod1
	bumpResourceVersion(pod1)
	jm.updatePod(&prev, pod1)
	if got, want := jm.queue.Len(), 1; got != want {
		t.Fatalf("queue.Len() = %v, want %v", got, want)
	}
	key, done := jm.queue.Get()
	if done {
		t.Fatalf("failed to enqueue controller for pod %v", pod1.Name)
	}
	expectedKey := GetKey(job1)
	if key.String() != expectedKey.String() {
		t.Errorf("queue.Get() = %s, want %s", key, expectedKey)
	}

	prev = *pod2
	bumpResourceVersion(pod2)
	jm.updatePod(&prev, pod2)
	if got, want := jm.queue.Len(), 1; got != want {
		t.Fatalf("queue.Len() = %v, want %v", got, want)
	}
	key, done = jm.queue.Get()
	if done {
		t.Fatalf("failed to enqueue controller for pod %v", pod2.Name)
	}
	expectedKey = GetKey(job2)
	if key.String() != expectedKey.String() {
		t.Errorf("queue.Get() = %s, want %s", key, expectedKey)
	}
}

func TestUpdatePodOrphanWithNewLabels(t *testing.T) {
	jm, sharedInformerFactory, runaiSharedInformersFactory := newController()

	job1 := newJob(1, 1, 6)
	job1.Name = "job1"
	job2 := newJob(1, 1, 6)
	job2.Name = "job2"
	err := runaiSharedInformersFactory.Run().V1().RunaiJobs().Informer().GetIndexer().Add(job1)
	if err != nil {
		klog.Errorf("failed to add job, error: <%v>", err)
	}
	err = runaiSharedInformersFactory.Run().V1().RunaiJobs().Informer().GetIndexer().Add(job2)
	if err != nil {
		klog.Errorf("failed to add job, error: <%v>", err)
	}

	podIndexer := sharedInformerFactory.Core().V1().Pods().Informer().GetIndexer()
	configMap := setPodsStatusesAndGetConfigMap(podIndexer, job1, 0, 0, 0, 0)

	podSuffix := "0-0"
	configMap.Data[podSuffix] = "Running"
	pod1 := newPod(podSuffix, job1, "Running", createCondition(CreateConditionRunningTrue))
	pod1.OwnerReferences = nil
	err = podIndexer.Add(pod1)
	if err != nil {
		klog.Errorf("failed to add pod")
	}
	err = sharedInformerFactory.Core().V1().ConfigMaps().Informer().GetIndexer().Add(configMap)
	if err != nil {
		klog.Errorf("failed to add configmap")
	}

	// Labels changed on orphan. Expect newly matching controllers to queue.
	prev := *pod1
	prev.Labels = map[string]string{"foo2": "bar2"}
	bumpResourceVersion(pod1)
	jm.updatePod(&prev, pod1)
	if got, want := jm.queue.Len(), 1; got != want {
		t.Fatalf("queue.Len() = %v, want %v", got, want)
	}
}

func TestUpdatePodChangeControllerRef(t *testing.T) {
	jm, sharedInformerFactory, runaiSharedInformersFactory := newController()

	job1 := newJob(1, 1, 6)
	job1.Name = "job1"
	job2 := newJob(1, 1, 6)
	job2.Name = "job2"
	err := runaiSharedInformersFactory.Run().V1().RunaiJobs().Informer().GetIndexer().Add(job1)
	if err != nil {
		klog.Errorf("failed to add job, error: <%v>", err)
	}
	err = runaiSharedInformersFactory.Run().V1().RunaiJobs().Informer().GetIndexer().Add(job2)
	if err != nil {
		klog.Errorf("failed to add job, error: <%v>", err)
	}

	podIndexer := sharedInformerFactory.Core().V1().Pods().Informer().GetIndexer()
	configMap := setPodsStatusesAndGetConfigMap(podIndexer, job1, 0, 0, 0, 0)

	podSuffix := "0-0"
	configMap.Data[podSuffix] = "Running"
	pod1 := newPod(podSuffix, job1, "Running", createCondition(CreateConditionRunningTrue))
	err = podIndexer.Add(pod1)
	if err != nil {
		klog.Errorf("failed to add pod")
	}
	err = sharedInformerFactory.Core().V1().ConfigMaps().Informer().GetIndexer().Add(configMap)
	if err != nil {
		klog.Errorf("failed to add configmap")
	}

	// Changed ControllerRef. Expect both old and new to queue.
	prev := *pod1
	prev.OwnerReferences = []metav1.OwnerReference{*metav1.NewControllerRef(job2, controllerKind)}
	bumpResourceVersion(pod1)
	jm.updatePod(&prev, pod1)
	if got, want := jm.queue.Len(), 2; got != want {
		t.Fatalf("queue.Len() = %v, want %v", got, want)
	}
}

func TestUpdatePodRelease(t *testing.T) {
	jm, sharedInformerFactory, runaiSharedInformersFactory := newController()

	job1 := newJob(1, 1, 6)
	job1.Name = "job1"
	job2 := newJob(1, 1, 6)
	job2.Name = "job2"
	err := runaiSharedInformersFactory.Run().V1().RunaiJobs().Informer().GetIndexer().Add(job1)
	if err != nil {
		klog.Errorf("failed to add job, error: <%v>", err)
	}
	err = runaiSharedInformersFactory.Run().V1().RunaiJobs().Informer().GetIndexer().Add(job2)
	if err != nil {
		klog.Errorf("failed to add job, error: <%v>", err)
	}

	podIndexer := sharedInformerFactory.Core().V1().Pods().Informer().GetIndexer()
	configMap := setPodsStatusesAndGetConfigMap(podIndexer, job1, 0, 0, 0, 0)

	podSuffix := "0-0"
	configMap.Data[podSuffix] = "Running"
	pod1 := newPod(podSuffix, job1, "Running", createCondition(CreateConditionRunningTrue))
	err = podIndexer.Add(pod1)
	if err != nil {
		klog.Errorf("failed to add pod")
	}
	err = sharedInformerFactory.Core().V1().ConfigMaps().Informer().GetIndexer().Add(configMap)
	if err != nil {
		klog.Errorf("failed to add configmap")
	}

	// Remove ControllerRef. Expect all matching to queue for adoption.
	prev := *pod1
	pod1.OwnerReferences = nil
	bumpResourceVersion(pod1)
	jm.updatePod(&prev, pod1)
	if got, want := jm.queue.Len(), 1; got != want {
		t.Fatalf("queue.Len() = %v, want %v", got, want)
	}
}

func TestDeletePod(t *testing.T) {
	jm, sharedInformerFactory, runaiSharedInformersFactory := newController()

	job1 := newJob(1, 1, 6)
	job1.Name = "job1"
	job2 := newJob(1, 1, 6)
	job2.Name = "job2"
	err := runaiSharedInformersFactory.Run().V1().RunaiJobs().Informer().GetIndexer().Add(job1)
	if err != nil {
		klog.Errorf("failed to add job, error: <%v>", err)
	}
	err = runaiSharedInformersFactory.Run().V1().RunaiJobs().Informer().GetIndexer().Add(job2)
	if err != nil {
		klog.Errorf("failed to add job, error: <%v>", err)
	}

	podIndexer := sharedInformerFactory.Core().V1().Pods().Informer().GetIndexer()
	configMap := setPodsStatusesAndGetConfigMap(podIndexer, job1, 0, 0, 0, 0)

	podSuffix := "0-0"
	configMap.Data[podSuffix] = "Running"
	pod1 := newPod(podSuffix, job1, "Running", createCondition(CreateConditionRunningTrue))
	err = podIndexer.Add(pod1)
	if err != nil {
		klog.Errorf("failed to add pod")
	}
	err = sharedInformerFactory.Core().V1().ConfigMaps().Informer().GetIndexer().Add(configMap)
	if err != nil {
		klog.Errorf("failed to add configmap")
	}

	configMap2 := setPodsStatusesAndGetConfigMap(podIndexer, job2, 0, 0, 0, 0)
	configMap2.Data[podSuffix] = "Running"
	pod2 := newPod(podSuffix, job2, "Running", createCondition(CreateConditionRunningTrue))
	err = podIndexer.Add(pod2)
	if err != nil {
		klog.Errorf("failed to add pod")
	}
	err = sharedInformerFactory.Core().V1().ConfigMaps().Informer().GetIndexer().Add(configMap2)
	if err != nil {
		klog.Errorf("failed to add configmap")
	}
	err = sharedInformerFactory.Core().V1().Pods().Informer().GetIndexer().Add(pod2)
	if err != nil {
		klog.Errorf("Failed to add pod, error: <%v>", err)
	}

	jm.deletePod(pod1)
	if got, want := jm.queue.Len(), 1; got != want {
		t.Fatalf("queue.Len() = %v, want %v", got, want)
	}
	key, done := jm.queue.Get()
	if done {
		t.Fatalf("failed to enqueue controller for pod %v", pod1.Name)
	}
	expectedKey := GetKey(job1)
	if key.String() != expectedKey.String() {
		t.Errorf("queue.Get() = %s, want %s", key, expectedKey)
	}

	jm.deletePod(pod2)
	if got, want := jm.queue.Len(), 1; got != want {
		t.Fatalf("queue.Len() = %v, want %v", got, want)
	}
	key, done = jm.queue.Get()
	if done {
		t.Fatalf("failed to enqueue controller for pod %v", pod2.Name)
	}
	expectedKey = GetKey(job2)
	if key.String() != expectedKey.String() {
		t.Errorf("queue.Get() = %s, want %s", key, expectedKey)
	}
}

func TestDeletePodOrphan(t *testing.T) {
	jm, sharedInformerFactory, runaiSharedInformersFactory := newController()

	job1 := newJob(1, 1, 6)
	job1.Name = "job1"
	job2 := newJob(1, 1, 6)
	job2.Name = "job2"
	job3 := newJob(1, 1, 6)
	job3.Name = "job3"
	job3.Spec.Selector.MatchLabels = map[string]string{"other": "labels"}
	err := runaiSharedInformersFactory.Run().V1().RunaiJobs().Informer().GetIndexer().Add(job1)
	if err != nil {
		klog.Errorf("failed to add job, error: <%v>", err)
	}
	err = runaiSharedInformersFactory.Run().V1().RunaiJobs().Informer().GetIndexer().Add(job2)
	if err != nil {
		klog.Errorf("failed to add job, error: <%v>", err)
	}
	err = runaiSharedInformersFactory.Run().V1().RunaiJobs().Informer().GetIndexer().Add(job3)
	if err != nil {
		klog.Errorf("failed to add job, error: <%v>", err)
	}
	podIndexer := sharedInformerFactory.Core().V1().Pods().Informer().GetIndexer()
	configMap := setPodsStatusesAndGetConfigMap(podIndexer, job1, 0, 0, 0, 0)

	podSuffix := "0-0"
	configMap.Data[podSuffix] = "Running"
	pod1 := newPod(podSuffix, job1, "Running", createCondition(CreateConditionRunningTrue))
	pod1.OwnerReferences = nil
	err = podIndexer.Add(pod1)
	if err != nil {
		klog.Errorf("failed to add pod")
	}
	err = sharedInformerFactory.Core().V1().ConfigMaps().Informer().GetIndexer().Add(configMap)
	if err != nil {
		klog.Errorf("failed to add configmap")
	}

	jm.deletePod(pod1)
	if got, want := jm.queue.Len(), 0; got != want {
		t.Fatalf("queue.Len() = %v, want %v", got, want)
	}
}

type FakeJobExpectations struct {
	*controller.ControllerExpectations
	satisfied    bool
	expSatisfied func()
}

func (fe FakeJobExpectations) SatisfiedExpectations(logger klog.Logger, controllerKey string) bool {
	fe.expSatisfied()
	return fe.satisfied
}

// TestSyncJobExpectations tests that a pod cannot sneak in between counting active pods
// and checking expectations.
func TestSyncJobExpectations(t *testing.T) {
	manager, sharedInformerFactory, runaiSharedInformersFactory := newController()
	fakePodControl := controller.FakePodControl{}
	manager.podControl = &fakePodControl
	manager.control = &FakePodControl{}
	manager.recorder = record.NewFakeRecorder(1024)

	manager.updateHandler = func(job *runaijobv1.RunaiJob) error { return nil }

	job := newJob(2, 2, 6)
	err := runaiSharedInformersFactory.Run().V1().RunaiJobs().Informer().GetIndexer().Add(job)
	if err != nil {
		klog.Errorf("Failed to add runaiJob, error: <%v>", err)
	}
	configmap := v1.ConfigMap{
		TypeMeta: metav1.TypeMeta{Kind: "ConfigMap", APIVersion: "v1"},
		ObjectMeta: metav1.ObjectMeta{
			Name:            fmt.Sprintf("%s-pod-attempts", job.Name),
			Namespace:       job.Namespace,
			UID:             uuid.NewUUID(),
			ResourceVersion: "1",
			OwnerReferences: []metav1.OwnerReference{*metav1.NewControllerRef(job, controllerKind)},
		},
		Data: make(map[string]string),
	}
	index := 0
	podIndexer := sharedInformerFactory.Core().V1().Pods().Informer().GetIndexer()

	pods := newPodList(podIndexer, 2, v1.PodPending, nil, job, &index, &configmap, nil)
	err = sharedInformerFactory.Core().V1().ConfigMaps().Informer().GetIndexer().Add(&configmap)
	if err != nil {
		klog.Errorf("failed to add configmap")
	}

	err = podIndexer.Add(pods[0])
	if err != nil {
		klog.Errorf("Failed to add pod, error: <%v>", err)
	}

	manager.expectations = FakeJobExpectations{
		controller.NewControllerExpectations(), true, func() {
			// If we check active pods before checking expectations, the job
			// will create a new replica because it doesn't see this pod, but
			// has fulfilled its expectations.
			err := podIndexer.Add(&pods[1])
			if err != nil {
				klog.Errorf("Failed to add pod, error: <%v>", err)
			}
		},
	}

	_, err = manager.syncJob(GetKey(job))
	if err != nil {
		t.Errorf("Unexpected error was returned, error: <%v>", err)
	}
	if len(fakePodControl.Templates) != 0 {
		t.Errorf("Unexpected number of creates.  Expected %d, saw %d\n", 0, len(fakePodControl.Templates))
	}
	if len(fakePodControl.DeletePodName) != 0 {
		t.Errorf("Unexpected number of deletes.  Expected %d, saw %d\n", 0, len(fakePodControl.DeletePodName))
	}
}

func TestWatchJobs(t *testing.T) {
	// client-go now enables the WatchListClient feature gate by default, which makes the
	// reflector use the streaming watch-list protocol. The fake clientset's watch reactor
	// does not implement that protocol, so the reflector blocks forever. Disable it for the test.
	clientfeaturestesting.SetFeatureDuringTest(t, clientfeatures.WatchListClient, false)
	clientset := clientset.NewForConfigOrDie(&restclient.Config{Host: "", ContentConfig: restclient.ContentConfig{GroupVersion: &schema.GroupVersion{Group: "", Version: "v1"}}})
	clientsetrunai := fakeclientset.NewSimpleClientset()
	fakeWatch := watch.NewFake()
	clientsetrunai.PrependWatchReactor("runaijobs", core.DefaultWatchReactor(fakeWatch, nil))
	manager, sharedInformerFactory, runaiSharedInformersFactory := newControllerFromClient(clientset, clientsetrunai, controller.NoResyncPeriodFunc)
	manager.podStoreSynced = alwaysReady
	manager.jobStoreSynced = alwaysReady
	manager.configMapSynced = alwaysReady

	var testJob runaijobv1.RunaiJob
	received := make(chan struct{})

	// The update sent through the fakeWatcher should make its way into the workqueue,
	// and eventually into the syncHandler.
	manager.syncHandler = func(key ctrl.Request) (bool, error) {
		defer close(received)
		ns, name := key.Namespace, key.Name

		job, err := manager.jobLister.RunaiJobs(ns).Get(name)
		if err != nil || job == nil {
			t.Errorf("Expected to find job under key %v: %v", key, err)
			return true, nil
		}
		if !apiequality.Semantic.DeepDerivative(*job, testJob) {
			t.Errorf("Expected %#v, but got %#v", testJob, *job)
		}
		return true, nil
	}
	// Start only the job watcher and the workqueue, send a watch event,
	// and make sure it hits the sync method.
	stopCh := make(chan struct{})
	defer close(stopCh)
	sharedInformerFactory.Start(stopCh)
	runaiSharedInformersFactory.Start(stopCh)
	go manager.Run(1, stopCh)

	// We're sending new job to see if it reaches syncHandler.
	testJob.Namespace = "bar"
	testJob.Name = "foo"
	fakeWatch.Add(&testJob)
	t.Log("Waiting for job to reach syncHandler")
	<-received
}

func TestWatchPods(t *testing.T) {
	// See TestWatchJobs: disable the streaming watch-list protocol so the fake watch reactor works.
	clientfeaturestesting.SetFeatureDuringTest(t, clientfeatures.WatchListClient, false)
	testJob := newJob(2, 2, 6)
	clientsetrunai := fakeclientset.NewSimpleClientset(testJob)
	fakeWatch := watch.NewFake()
	clientset := fake.NewSimpleClientset()
	clientset.PrependWatchReactor("pods", core.DefaultWatchReactor(fakeWatch, nil))
	manager, sharedInformerFactory, runaiSharedInformersFactory := newControllerFromClient(clientset, clientsetrunai, controller.NoResyncPeriodFunc)
	manager.podStoreSynced = alwaysReady
	manager.jobStoreSynced = alwaysReady

	// Put one job and one pod into the store
	err := runaiSharedInformersFactory.Run().V1().RunaiJobs().Informer().GetIndexer().Add(testJob)
	if err != nil {
		klog.Errorf("Failed to add runaiJob, error: <%v>", err)
	}
	// The pod update sent through the fakeWatcher should figure out the managing job and
	// send it into the syncHandler.
	manager.syncHandler = func(key ctrl.Request) (bool, error) {
		ns, name := key.Namespace, key.Name
		job, err := manager.jobLister.RunaiJobs(ns).Get(name)
		if err != nil {
			t.Errorf("Expected to find job under key %v: %v", key, err)
		}
		if !apiequality.Semantic.DeepDerivative(job, testJob) {
			t.Errorf("\nExpected %#v,\nbut got %#v", testJob, job)
			return true, nil
		}
		return true, nil
	}
	// Start only the pod watcher and the workqueue, send a watch event,
	// and make sure it hits the sync method for the right job.
	stopCh := make(chan struct{})
	defer close(stopCh)
	go sharedInformerFactory.Core().V1().Pods().Informer().Run(stopCh)
	go runaiSharedInformersFactory.Run().V1().RunaiJobs().Informer().Run(stopCh)

	go wait.Until(manager.worker, 100*time.Millisecond, stopCh)

	index := 0
	podIndexer := sharedInformerFactory.Core().V1().Pods().Informer().GetIndexer()
	configmap := v1.ConfigMap{
		TypeMeta: metav1.TypeMeta{Kind: "ConfigMap", APIVersion: "v1"},
		ObjectMeta: metav1.ObjectMeta{
			Name:            fmt.Sprintf("%s-pod-attempts", testJob.Name),
			Namespace:       testJob.Namespace,
			UID:             uuid.NewUUID(),
			ResourceVersion: "1",
			OwnerReferences: []metav1.OwnerReference{*metav1.NewControllerRef(testJob, controllerKind)},
		},
		Data: make(map[string]string),
	}

	pods := newPodList(podIndexer, 1, v1.PodRunning, createCondition(CreateConditionRunningTrue), testJob, &index, &configmap, nil)
	testPod := pods[0]
	testPod.Status.Phase = v1.PodFailed
	err = sharedInformerFactory.Core().V1().ConfigMaps().Informer().GetIndexer().Add(&configmap)
	if err != nil {
		klog.Errorf("failed to add configmap")
	}

	err = podIndexer.Add(pods[0])
	if err != nil {
		klog.Errorf("Failed to add pod, error: <%v>", err)
	}

	fakeWatch.Add(&testPod)

	t.Log("Waiting for pod to reach syncHandler")
}

func bumpResourceVersion(obj metav1.Object) {
	ver, _ := strconv.ParseInt(obj.GetResourceVersion(), 10, 32)
	obj.SetResourceVersion(strconv.FormatInt(ver+1, 10))
}

type pods struct {
	pending int32
	active  int32
	succeed int32
	failed  int32
}

func TestJobBackoffReset(t *testing.T) {
	testCases := map[string]struct {
		// job setup
		parallelism  int32
		completions  int32
		backoffLimit int32

		// pod setup - each row is additive!
		pods []pods
	}{
		"parallelism=1": {
			1, 2, 1,
			[]pods{
				{0, 1, 0, 1},
				{0, 0, 1, 0},
			},
		},
		"parallelism=2 (just failure)": {
			2, 2, 1,
			[]pods{
				{0, 2, 0, 1},
				{0, 0, 1, 0},
			},
		},
	}

	for name, tc := range testCases {
		DefaultJobBackOff = time.Duration(0) // overwrite the default value for testing
		manager, sharedInformerFactory, runaiSharedInformersFactory := newController()
		fakePodControl := controller.FakePodControl{}
		manager.podControl = &fakePodControl
		var actual *runaijobv1.RunaiJob
		manager.updateHandler = func(job *runaijobv1.RunaiJob) error {
			actual = job
			return nil
		}

		// job & pods setup
		job := newJob(tc.parallelism, tc.completions, tc.backoffLimit)
		err := runaiSharedInformersFactory.Run().V1().RunaiJobs().Informer().GetIndexer().Add(job)
		if err != nil {
			klog.Errorf("Failed to add runaiJob, error: <%v>", err)
		}
		podIndexer := sharedInformerFactory.Core().V1().Pods().Informer().GetIndexer()

		configmap := setPodsStatusesAndGetConfigMap(podIndexer, job, tc.pods[0].pending, tc.pods[0].active, tc.pods[0].succeed, tc.pods[0].failed)
		err = sharedInformerFactory.Core().V1().ConfigMaps().Informer().GetIndexer().Add(configmap)
		if err != nil {
			klog.Errorf("failed to add configmap")
		}
		key := GetKey(job)
		manager.queue.Add(key)
		manager.processNextWorkItem()
		retries := manager.queue.NumRequeues(key)
		if retries != 1 {
			t.Errorf("%s: expected exactly 1 retry, got %d", name, retries)
		}

		podList := podIndexer.List()
		for _, podObj := range podList {
			pod, ok := podObj.(*v1.Pod)
			if !ok {
				klog.Errorf("Failed to delete pods")
			}
			err = podIndexer.Delete(pod)
			if err != nil {
				klog.Errorf("Failed to delete pods")
			}
		}
		job = actual
		err = runaiSharedInformersFactory.Run().V1().RunaiJobs().Informer().GetIndexer().Replace([]interface{}{actual}, actual.ResourceVersion)
		if err != nil {
			klog.Errorf("Failed to add runaiJob, error: <%v>", err)
		}
		configmap2 := setPodsStatusesAndGetConfigMap(podIndexer, job, tc.pods[1].pending, tc.pods[1].active, tc.pods[1].succeed, tc.pods[1].failed)
		err = sharedInformerFactory.Core().V1().ConfigMaps().Informer().GetIndexer().Delete(configmap)
		if err != nil {
			klog.Errorf("failed to delete configmap")
		}
		err = sharedInformerFactory.Core().V1().ConfigMaps().Informer().GetIndexer().Add(configmap2)
		if err != nil {
			klog.Errorf("failed to add configmap")
		}
		manager.processNextWorkItem()
		retries = manager.queue.NumRequeues(key)
		if retries != 0 {
			t.Errorf("%s: expected exactly 0 retries, got %d", name, retries)
		}
		if getCondition(actual, runaijobv1.JobFailed, v1.ConditionTrue, "BackoffLimitExceeded") {
			t.Errorf("%s: unexpected job failure", name)
		}
	}
}

var _ workqueue.TypedRateLimitingInterface[ctrl.Request] = &fakeRateLimitingQueue{}

type fakeRateLimitingQueue struct {
	workqueue.TypedInterface[ctrl.Request]
	requeues int
	item     interface{}
	duration time.Duration
}

func (f *fakeRateLimitingQueue) AddRateLimited(item ctrl.Request) {}
func (f *fakeRateLimitingQueue) Forget(item ctrl.Request) {
	f.requeues = 0
}
func (f *fakeRateLimitingQueue) NumRequeues(item ctrl.Request) int {
	return f.requeues
}
func (f *fakeRateLimitingQueue) AddAfter(item ctrl.Request, duration time.Duration) {
	f.item = item
	f.duration = duration
}

func TestJobBackoff(t *testing.T) {
	job := newJob(1, 1, 1)
	oldPod := newPod("0-0", job, "Running", createCondition(CreateConditionRunningTrue))
	newPod := oldPod.DeepCopy()
	newPod.ResourceVersion = "2"

	testCases := map[string]struct {
		// inputs
		requeues int
		phase    v1.PodPhase

		// expectation
		backoff int
	}{
		"1st failure": {0, v1.PodFailed, 0},
		"2nd failure": {1, v1.PodFailed, 1},
		"3rd failure": {2, v1.PodFailed, 2},
		"1st success": {0, v1.PodSucceeded, 0},
		"2nd success": {1, v1.PodSucceeded, 0},
		"1st running": {0, v1.PodSucceeded, 0},
		"2nd running": {1, v1.PodSucceeded, 0},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			manager, _, runaiSharedInformersFactory := newController()
			manager.podControl = &controller.FakePodControl{}
			fakeControl := FakePodControl{}
			manager.control = &fakeControl
			queue := &fakeRateLimitingQueue{}
			manager.queue = queue
			err := runaiSharedInformersFactory.Run().V1().RunaiJobs().Informer().GetIndexer().Add(job)
			if err != nil {
				klog.Errorf("Failed to add runaiJob, error: <%v>", err)
			}

			queue.requeues = tc.requeues
			newPod.Status.Phase = tc.phase
			manager.updatePod(oldPod, newPod)

			if queue.duration.Nanoseconds() != int64(tc.backoff)*DefaultJobBackOff.Nanoseconds() {
				t.Errorf("unexpected backoff %v", queue.duration)
			}
		})
	}
}

func TestJobBackoffForOnFailure(t *testing.T) {
	jobConditionFailed := runaijobv1.JobFailed

	testCases := map[string]struct {
		// job setup
		parallelism  int32
		completions  int32
		backoffLimit int32
		pending      int32

		// pod setup
		jobKeyForget  bool
		restartCounts []int32
		podPhase      v1.PodPhase

		// expectations
		expectedActive          int32
		expectedSucceeded       int32
		expectedFailed          int32
		expectedPending         int32
		expectedCondition       *runaijobv1.JobConditionType
		expectedConditionReason string
	}{
		"backoffLimit 0 should have 1 pod active": {
			1, 1, 0, 1,
			true, []int32{0}, v1.PodRunning,
			1, 0, 0, 0,
			nil, "",
		},
		"backoffLimit 1 with restartCount 0 should have 1 pod active": {
			1, 1, 1, 1,
			true, []int32{0}, v1.PodRunning,
			1, 0, 0, 0,
			nil, "",
		},
		"backoffLimit 1 with restartCount 1 and podRunning should have 0 pod active": {
			1, 1, 1, 1,
			true, []int32{1}, v1.PodRunning,
			0, 0, 1, 0,
			&jobConditionFailed, "BackoffLimitExceeded",
		},
		"backoffLimit 1 with restartCount 0 and podPending should have 0 pod active, 1 pending": {
			1, 1, 1, 1,
			true, []int32{0}, v1.PodPending,
			1, 0, 0, 1,
			nil, "BackoffLimitExceeded",
		},
		"backoffLimit 1 with restartCount 1 and podPending should have 0 pod active": {
			1, 1, 1, 1,
			true, []int32{1}, v1.PodPending,
			0, 0, 1, 0,
			&jobConditionFailed, "BackoffLimitExceeded",
		},
		"too many job failures with podRunning - single pod": {
			1, 5, 2, 1,
			true, []int32{2}, v1.PodRunning,
			0, 0, 1, 0,
			&jobConditionFailed, "BackoffLimitExceeded",
		},
		"too many job failures with podPending - single pod": {
			1, 5, 2, 1,
			true, []int32{2}, v1.PodPending,
			0, 0, 1, 0,
			&jobConditionFailed, "BackoffLimitExceeded",
		},
		"too many job failures with podRunning - multiple pods": {
			2, 5, 2, 1,
			true, []int32{1, 1}, v1.PodRunning,
			0, 0, 2, 0,
			&jobConditionFailed, "BackoffLimitExceeded",
		},
		"too many job failures with podPending - multiple pods": {
			2, 5, 2, 1,
			true, []int32{1, 1}, v1.PodPending,
			0, 0, 3, 0,
			&jobConditionFailed, "BackoffLimitExceeded",
		},
		"not enough failures": {
			2, 5, 3, 1,
			true, []int32{1, 1}, v1.PodRunning,
			2, 0, 0, 0,
			nil, "",
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			// job manager setup
			manager, sharedInformerFactory, runaiSharedInformersFactory := newController()
			fakePodControl := controller.FakePodControl{}
			manager.podControl = &fakePodControl
			manager.control = &FakePodControl{}
			manager.configMapSynced = alwaysReady
			var actual *runaijobv1.RunaiJob
			manager.updateHandler = func(job *runaijobv1.RunaiJob) error {
				actual = job
				return nil
			}

			// job & pods setup
			job := newJob(tc.parallelism, tc.completions, tc.backoffLimit)
			job.Spec.Template.Spec.RestartPolicy = v1.RestartPolicyOnFailure
			job.Status.Pending = tc.pending
			err := runaiSharedInformersFactory.Run().V1().RunaiJobs().Informer().GetIndexer().Add(job)
			if err != nil {
				klog.Errorf("Failed to add runaiJob, error: <%v>", err)
			}
			index := 0
			podIndexer := sharedInformerFactory.Core().V1().Pods().Informer().GetIndexer()
			configmap := v1.ConfigMap{
				TypeMeta: metav1.TypeMeta{Kind: "ConfigMap", APIVersion: "v1"},
				ObjectMeta: metav1.ObjectMeta{
					Name:            fmt.Sprintf("%s-pod-attempts", job.Name),
					Namespace:       job.Namespace,
					UID:             uuid.NewUUID(),
					ResourceVersion: "1",
					OwnerReferences: []metav1.OwnerReference{*metav1.NewControllerRef(job, controllerKind)},
				},
				Data: make(map[string]string),
			}

			newPodList(podIndexer, int32(len(tc.restartCounts)), tc.podPhase, createCondition(CreateConditionRunningTrue), job, &index, &configmap, tc.restartCounts)
			if name == "too many job failures with podPending - multiple pods" {
				newPodList(podIndexer, 1, v1.PodPending, createCondition(CreateConditionScheduledFalse), job, &index, &configmap, []int32{0})
			}

			err = sharedInformerFactory.Core().V1().ConfigMaps().Informer().GetIndexer().Add(&configmap)
			if err != nil {
				klog.Errorf("failed to add configmap")
			}

			// run
			forget, err := manager.syncJob(GetKey(job))

			if err != nil {
				t.Errorf("unexpected error syncing job.  Got %#v", err)
			}
			if forget != tc.jobKeyForget {
				t.Errorf("unexpected forget value. Expected %v, saw %v\n", tc.jobKeyForget, forget)
			}
			// validate status
			if actual.Status.Active != tc.expectedActive {
				t.Errorf("unexpected number of active pods.  Expected %d, saw %d\n", tc.expectedActive, actual.Status.Active)
			}
			if actual.Status.Succeeded != tc.expectedSucceeded {
				t.Errorf("unexpected number of succeeded pods.  Expected %d, saw %d\n", tc.expectedSucceeded, actual.Status.Succeeded)
			}
			if actual.Status.Failed != tc.expectedFailed {
				t.Errorf("unexpected number of failed pods.  Expected %d, saw %d\n", tc.expectedFailed, actual.Status.Failed)
			}
			if actual.Status.Pending != tc.expectedPending {
				t.Errorf("unexpected number of pending pods.  Expected %d, saw %d\n", tc.expectedFailed, actual.Status.Failed)
			}
			// validate conditions
			if tc.expectedCondition != nil && !getCondition(actual, *tc.expectedCondition, v1.ConditionTrue, tc.expectedConditionReason) {
				t.Errorf("expected completion condition.  Got %#v", actual.Status.Conditions)
			}
		})
	}
}

func TestJobBackoffOnRestartPolicyNever(t *testing.T) {
	jobConditionFailed := runaijobv1.JobFailed

	testCases := map[string]struct {
		// job setup
		parallelism  int32
		completions  int32
		backoffLimit int32

		// pod setup
		activePodsPhase v1.PodPhase
		activePods      int32
		failedPods      int32

		// expectations
		isExpectingAnError      bool
		jobKeyForget            bool
		expectedActive          int32
		expectedSucceeded       int32
		expectedFailed          int32
		expectedCondition       *runaijobv1.JobConditionType
		expectedConditionReason string
	}{
		"not enough failures with backoffLimit 0 - single pod": {
			1, 1, 0,
			v1.PodRunning, 1, 0,
			false, true, 1, 0, 0, nil, "",
		},
		"not enough failures with backoffLimit 1 - single pod": {
			1, 1, 1,
			"", 0, 1,
			true, false, 1, 0, 1, nil, "",
		},
		"too many failures with backoffLimit 1 - single pod": {
			1, 1, 1,
			"", 0, 2,
			false, true, 0, 0, 2, &jobConditionFailed, "BackoffLimitExceeded",
		},
		"not enough failures with backoffLimit 6 - multiple pods": {
			2, 2, 6,
			v1.PodRunning, 1, 6,
			true, false, 2, 0, 6, nil, "",
		},
		"too many failures with backoffLimit 6 - multiple pods": {
			2, 2, 6,
			"", 0, 7,
			false, true, 0, 0, 7, &jobConditionFailed, "BackoffLimitExceeded",
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			// job manager setup
			manager, sharedInformerFactory, runaiSharedInformersFactory := newController()
			fakePodControl := controller.FakePodControl{}
			manager.podControl = &fakePodControl
			manager.control = &FakePodControl{}
			var actual *runaijobv1.RunaiJob
			manager.updateHandler = func(job *runaijobv1.RunaiJob) error {
				actual = job
				return nil
			}

			// job & pods setup
			job := newJob(tc.parallelism, tc.completions, tc.backoffLimit)
			job.Spec.Template.Spec.RestartPolicy = v1.RestartPolicyNever
			err := runaiSharedInformersFactory.Run().V1().RunaiJobs().Informer().GetIndexer().Add(job)
			if err != nil {
				klog.Errorf("Failed to add runaiJob, error: <%v>", err)
			}
			index := 0
			podIndexer := sharedInformerFactory.Core().V1().Pods().Informer().GetIndexer()
			configmap := v1.ConfigMap{
				TypeMeta: metav1.TypeMeta{Kind: "ConfigMap", APIVersion: "v1"},
				ObjectMeta: metav1.ObjectMeta{
					Name:            fmt.Sprintf("%s-pod-attempts", job.Name),
					Namespace:       job.Namespace,
					UID:             uuid.NewUUID(),
					ResourceVersion: "1",
					OwnerReferences: []metav1.OwnerReference{*metav1.NewControllerRef(job, controllerKind)},
				},
				Data: make(map[string]string),
			}
			newPodList(podIndexer, tc.failedPods, v1.PodFailed, nil, job, &index, &configmap, nil)
			newPodList(podIndexer, tc.activePods, v1.PodRunning, createCondition(CreateConditionRunningTrue), job, &index, &configmap, nil)

			err = sharedInformerFactory.Core().V1().ConfigMaps().Informer().GetIndexer().Add(&configmap)
			if err != nil {
				klog.Errorf("failed to add configmap")
			}

			// run
			forget, err := manager.syncJob(GetKey(job))

			if (err != nil) != tc.isExpectingAnError {
				t.Errorf("unexpected error syncing job. Got %#v, isExpectingAnError: %v\n", err, tc.isExpectingAnError)
			}
			if forget != tc.jobKeyForget {
				t.Errorf("unexpected forget value. Expected %v, saw %v\n", tc.jobKeyForget, forget)
			}
			// validate status
			if actual.Status.Active != tc.expectedActive {
				t.Errorf("unexpected number of active pods. Expected %d, saw %d\n", tc.expectedActive, actual.Status.Active)
			}
			if actual.Status.Succeeded != tc.expectedSucceeded {
				t.Errorf("unexpected number of succeeded pods. Expected %d, saw %d\n", tc.expectedSucceeded, actual.Status.Succeeded)
			}
			if actual.Status.Failed != tc.expectedFailed {
				t.Errorf("unexpected number of failed pods. Expected %d, saw %d\n", tc.expectedFailed, actual.Status.Failed)
			}
			// validate conditions
			if tc.expectedCondition != nil && !getCondition(actual, *tc.expectedCondition, v1.ConditionTrue, tc.expectedConditionReason) {
				t.Errorf("expected completion condition. Got %#v", actual.Status.Conditions)
			}
		})
	}
}

func TestBackoffLimitFailedPodsNotCounted(t *testing.T) {
	testCases := map[string]struct {
		// job setup
		jobHaveNewFailure bool
		parallelism       int32
		backoffLimit      int32

		// pods setup
		activePods             int32
		failedPods             int32
		podsTerminatedMessages []string

		// expectations
		expectedDidExceedBackoffLimit bool
		expectedPodsAfterDeletion     int
	}{
		"test 1 - job doesnt have new failure": {
			jobHaveNewFailure:             false,
			parallelism:                   2,
			backoffLimit:                  6,
			activePods:                    1,
			failedPods:                    7,
			expectedDidExceedBackoffLimit: false,
			expectedPodsAfterDeletion:     7,
		},
		"test 2 - job active pods matches parallelism": {
			jobHaveNewFailure:             true,
			parallelism:                   2,
			activePods:                    2,
			backoffLimit:                  6,
			failedPods:                    7,
			expectedDidExceedBackoffLimit: false,
			expectedPodsAfterDeletion:     7,
		},
		"test 3 - job has not exceeded backoff limit": {
			jobHaveNewFailure:             true,
			parallelism:                   2,
			backoffLimit:                  6,
			activePods:                    1,
			failedPods:                    6,
			expectedDidExceedBackoffLimit: false,
			expectedPodsAfterDeletion:     6,
		},
		"test 4 - job has exceeded backoff limit": {
			jobHaveNewFailure:             true,
			parallelism:                   2,
			backoffLimit:                  6,
			activePods:                    1,
			failedPods:                    7,
			expectedDidExceedBackoffLimit: true,
			expectedPodsAfterDeletion:     7,
		},
		"test 5 - job has not exceeded backoff limit because of pods preempted by us": {
			jobHaveNewFailure:             true,
			parallelism:                   2,
			backoffLimit:                  6,
			activePods:                    1,
			failedPods:                    7,
			podsTerminatedMessages:        []string{failureMessageSwapPreempted},
			expectedDidExceedBackoffLimit: false,
			expectedPodsAfterDeletion:     6,
		},
		"test 6 - job has exceeded backoff limit because of pods with other termination message": {
			jobHaveNewFailure:             true,
			parallelism:                   2,
			backoffLimit:                  6,
			activePods:                    1,
			failedPods:                    7,
			podsTerminatedMessages:        []string{"some-termination-message"},
			expectedDidExceedBackoffLimit: true,
			expectedPodsAfterDeletion:     7,
		},
		"test 7 - job has not exceeded backoff limit; some pods were preempted by us and some not": {
			jobHaveNewFailure:             true,
			parallelism:                   2,
			backoffLimit:                  6,
			activePods:                    1,
			failedPods:                    7,
			podsTerminatedMessages:        []string{failureMessageSwapPreempted + "some-suffix", "some-termination-message"},
			expectedDidExceedBackoffLimit: false,
			expectedPodsAfterDeletion:     6,
		},
		"test 8 - job has exceeded backoff limit even tough some pods were preempted by us": {
			jobHaveNewFailure:             true,
			parallelism:                   2,
			backoffLimit:                  6,
			activePods:                    1,
			failedPods:                    8,
			podsTerminatedMessages:        []string{failureMessageSwapPreempted + "some-suffix", "some-termination-message"},
			expectedDidExceedBackoffLimit: true,
			expectedPodsAfterDeletion:     7,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			// job setup
			job := newJob(tc.parallelism, -1, tc.backoffLimit)

			pods := getPodsForTestBackoffLimitFailedPodsNotCounted(tc.failedPods, tc.podsTerminatedMessages, job)

			_, failed := getStatus(job, pods)
			result := didExceedBackoffLimit(tc.jobHaveNewFailure, tc.activePods, failed, job)

			// validate result
			if tc.expectedDidExceedBackoffLimit != result {
				t.Errorf("unexpected result of didExceedBackoffLimit. Got %v, expected %v\n", result, tc.expectedDidExceedBackoffLimit)
			}

			// test deletion of pods
			jm, _, _ := newControllerWithRunaiClientset(job)
			key := GetKey(job)
			updatedPods := jm.deleteSwapPreemptedPods(job, key.String(), pods)

			if tc.expectedPodsAfterDeletion != len(updatedPods) {
				t.Errorf("unexpected result of pods after deletion. Got %v, expected %v\n", len(updatedPods), tc.expectedPodsAfterDeletion)
			}
		})
	}
}

func TestEnsureJobConditions(t *testing.T) {
	testCases := []struct {
		name         string
		haveList     []runaijobv1.JobCondition
		wantType     runaijobv1.JobConditionType
		wantStatus   v1.ConditionStatus
		wantReason   string
		expectList   []runaijobv1.JobCondition
		expectUpdate bool
	}{
		{
			name:         "append true condition",
			haveList:     []runaijobv1.JobCondition{},
			wantType:     runaijobv1.JobSuspended,
			wantStatus:   v1.ConditionTrue,
			wantReason:   "foo",
			expectList:   []runaijobv1.JobCondition{runaiJobUtils.NewCondition(runaijobv1.JobSuspended, v1.ConditionTrue, "foo", "")},
			expectUpdate: true,
		},
		{
			name:         "append false condition",
			haveList:     []runaijobv1.JobCondition{},
			wantType:     runaijobv1.JobSuspended,
			wantStatus:   v1.ConditionFalse,
			wantReason:   "foo",
			expectList:   []runaijobv1.JobCondition{},
			expectUpdate: false,
		},
		{
			name:         "update true condition reason",
			haveList:     []runaijobv1.JobCondition{runaiJobUtils.NewCondition(runaijobv1.JobSuspended, v1.ConditionTrue, "foo", "")},
			wantType:     runaijobv1.JobSuspended,
			wantStatus:   v1.ConditionTrue,
			wantReason:   "bar",
			expectList:   []runaijobv1.JobCondition{runaiJobUtils.NewCondition(runaijobv1.JobSuspended, v1.ConditionTrue, "bar", "")},
			expectUpdate: true,
		},
		{
			name:         "update true condition status",
			haveList:     []runaijobv1.JobCondition{runaiJobUtils.NewCondition(runaijobv1.JobSuspended, v1.ConditionTrue, "foo", "")},
			wantType:     runaijobv1.JobSuspended,
			wantStatus:   v1.ConditionFalse,
			wantReason:   "foo",
			expectList:   []runaijobv1.JobCondition{runaiJobUtils.NewCondition(runaijobv1.JobSuspended, v1.ConditionFalse, "foo", "")},
			expectUpdate: true,
		},
		{
			name:         "update false condition status",
			haveList:     []runaijobv1.JobCondition{runaiJobUtils.NewCondition(runaijobv1.JobSuspended, v1.ConditionFalse, "foo", "")},
			wantType:     runaijobv1.JobSuspended,
			wantStatus:   v1.ConditionTrue,
			wantReason:   "foo",
			expectList:   []runaijobv1.JobCondition{runaiJobUtils.NewCondition(runaijobv1.JobSuspended, v1.ConditionTrue, "foo", "")},
			expectUpdate: true,
		},
		{
			name:         "condition already exists",
			haveList:     []runaijobv1.JobCondition{runaiJobUtils.NewCondition(runaijobv1.JobSuspended, v1.ConditionTrue, "foo", "")},
			wantType:     runaijobv1.JobSuspended,
			wantStatus:   v1.ConditionTrue,
			wantReason:   "foo",
			expectList:   []runaijobv1.JobCondition{runaiJobUtils.NewCondition(runaijobv1.JobSuspended, v1.ConditionTrue, "foo", "")},
			expectUpdate: false,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			gotList, isUpdated := runaiJobUtils.EnsureJobConditionStatus(tc.haveList, tc.wantType, tc.wantStatus, tc.wantReason, "")
			if isUpdated != tc.expectUpdate {
				t.Errorf("Got isUpdated=%v, want %v", isUpdated, tc.expectUpdate)
			}
			if len(gotList) != len(tc.expectList) {
				t.Errorf("got a list of length %d, want %d", len(gotList), len(tc.expectList))
			}
			for i := range gotList {
				// Make timestamps the same before comparing the two lists.
				gotList[i].LastProbeTime = tc.expectList[i].LastProbeTime
				gotList[i].LastTransitionTime = tc.expectList[i].LastTransitionTime
			}
			if diff := cmp.Diff(tc.expectList, gotList); diff != "" {
				t.Errorf("Unexpected JobCondition list: (-want,+got):\n%s", diff)
			}
		})
	}
}

func getPodsForTestBackoffLimitFailedPodsNotCounted(failedPods int32, podsTerminatedMessages []string, job *runaijobv1.RunaiJob) (pods []*v1.Pod) {
	for i := 0; i < int(failedPods); i++ {
		podSuffix := fmt.Sprintf("0-%d", i)
		pod := newPod(podSuffix, job, "Failed", nil)

		if i < len(podsTerminatedMessages) {
			pod.Status.ContainerStatuses = []v1.ContainerStatus{
				{
					State: v1.ContainerState{
						Terminated: &v1.ContainerStateTerminated{
							Message: podsTerminatedMessages[i],
						},
					},
				}}
		}

		pods = append(pods, pod)
	}

	return pods
}

type CreateCondition int

const (
	CreateConditionInitializingTrue CreateCondition = iota
	CreateConditionInitializingFalse
	CreateConditionScheduledTrue
	CreateConditionScheduledFalse
	CreateConditionRunningFalse
	CreateConditionRunningTrue
)

func createCondition(createCondition CreateCondition) *[]v1.PodCondition {
	var conditions []v1.PodCondition
	switch createCondition {
	case CreateConditionRunningTrue:
		conditions = []v1.PodCondition{
			{
				Type:   v1.PodReady,
				Status: v1.ConditionTrue,
			},
		}
	case CreateConditionRunningFalse:
		conditions = []v1.PodCondition{
			{
				Type:   v1.PodReady,
				Status: v1.ConditionFalse,
				Reason: "RunFalse",
			},
		}
	case CreateConditionInitializingTrue:
		conditions = []v1.PodCondition{
			{
				Type:   v1.PodScheduled,
				Status: v1.ConditionTrue,
			},
		}
	case CreateConditionInitializingFalse:
		conditions = []v1.PodCondition{
			{
				Type:   v1.PodScheduled,
				Status: v1.ConditionFalse,
				Reason: "InitFalse",
			},
		}
	case CreateConditionScheduledTrue:
		conditions = []v1.PodCondition{
			{
				Type:   v1.ContainersReady,
				Status: v1.ConditionTrue,
			},
		}
	case CreateConditionScheduledFalse:
		conditions = []v1.PodCondition{
			{
				Type:   v1.ContainersReady,
				Status: v1.ConditionFalse,
				Reason: "SchedFalse",
			},
		}
	default:
		conditions = nil
	}
	return &conditions
}
