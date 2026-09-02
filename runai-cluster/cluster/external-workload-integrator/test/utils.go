/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package test

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"testing"

	dynamov1alpha1 "github.com/ai-dynamo/dynamo/deploy/operator/api/v1alpha1"
	dynamov1beta1 "github.com/ai-dynamo/dynamo/deploy/operator/api/v1beta1"
	kartav1alpha1 "github.com/run-ai/karta/pkg/api/runai/v1alpha1"
	constants2 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/constants"
	kservev1beta1 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/external-workload-integrator/pkg/thirdpartyapi/kserve/v1beta1"
	workloadTypeConfig "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-common/workload-type-config"

	nvcfv1 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/external-workload-integrator/pkg/thirdpartyapi/nvcf/v1"

	. "github.com/onsi/gomega" //nolint:staticcheck // dot import is acceptable for test assertions

	devv1alpha2 "github.com/devfile/api/v2/pkg/apis/workspaces/v1alpha2"
	kaiorg "github.com/kai-scheduler/kai-resource-management-api/kai/v1alpha1"
	kubeflowv2beta1 "github.com/kubeflow/mpi-operator/pkg/apis/kubeflow/v2beta1"
	kubeflowv1 "github.com/kubeflow/training-operator/pkg/apis/kubeflow.org/v1"
	rayv1alpha1 "github.com/ray-project/kuberay/ray-operator/apis/ray/v1alpha1"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/external-workload-integrator/pkg/constants"
	kubeflowv1beta1 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/external-workload-integrator/pkg/thirdpartyapi/kubeflownotebook/v1beta1"
	nim "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/external-workload-integrator/pkg/thirdpartyapi/nim/v1alpha1"
	rayv1 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/external-workload-integrator/pkg/thirdpartyapi/ray/v1"
	seldonv1 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/external-workload-integrator/pkg/thirdpartyapi/seldon/v1"
	seldonv1alpha2 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/external-workload-integrator/pkg/thirdpartyapi/seldon/v1alpha2"
	scheme_init "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/go-operator/pkg/scheme-init"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1alpha1"
	v2 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v2"
	schedulingv2alpha2 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/scheduling/v2alpha2"
	appsv1 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	resourcev1 "k8s.io/api/resource/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/utils/ptr"
	servingv1 "knative.dev/serving/pkg/apis/serving/v1"
	kubevirtv1 "kubevirt.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	infv1 "sigs.k8s.io/gateway-api-inference-extension/api/v1"
	gwv1 "sigs.k8s.io/gateway-api/apis/v1"
	gwv1alpha2 "sigs.k8s.io/gateway-api/apis/v1alpha2"
)

// MockEnv Notice that after test is done env variables are cleaned up
func MockEnv(t *testing.T) {
	t.Setenv("USE_DEBUG_LOG_LEVEL", "true")
	t.Setenv("ALLOW_NON_SUPPORTED_TYPES", "true")
}

var (
	testScheme     *runtime.Scheme
	testSchemeOnce sync.Once
)

// getTestScheme returns the shared test scheme. scheme_init.Scheme() is a global
// singleton, so the AddToScheme calls below mutate shared state. They are run
// exactly once (sync.Once): re-running them on every BeforeEach raced the scheme's
// internal maps against the controller goroutines that read it during List,
// causing "concurrent map read and map write" crashes on the parallel CI runner.
func getTestScheme() *runtime.Scheme {
	testSchemeOnce.Do(buildTestScheme)
	return testScheme
}

func buildTestScheme() {
	scheme := scheme_init.Scheme()
	utilruntime.Must(rayv1alpha1.AddToScheme(scheme))
	utilruntime.Must(rayv1.AddToScheme(scheme))
	utilruntime.Must(seldonv1.AddToScheme(scheme))
	utilruntime.Must(seldonv1alpha2.AddToScheme(scheme))
	utilruntime.Must(servingv1.AddToScheme(scheme))
	utilruntime.Must(kubeflowv1.AddToScheme(scheme))
	utilruntime.Must(kubeflowv2beta1.AddToScheme(scheme))
	utilruntime.Must(kubeflowv1beta1.AddToScheme(scheme))
	utilruntime.Must(kubevirtv1.AddToScheme(scheme))
	utilruntime.Must(devv1alpha2.AddToScheme(scheme))
	utilruntime.Must(nim.AddToScheme(scheme))
	utilruntime.Must(nvcfv1.AddToScheme(scheme))
	utilruntime.Must(kservev1beta1.AddToScheme(scheme))
	utilruntime.Must(gwv1.Install(scheme))
	utilruntime.Must(gwv1alpha2.Install(scheme))
	utilruntime.Must(infv1.Install(scheme))
	utilruntime.Must(kartav1alpha1.AddToScheme(scheme))
	utilruntime.Must(resourcev1.AddToScheme(scheme))
	utilruntime.Must(dynamov1alpha1.AddToScheme(scheme))
	utilruntime.Must(dynamov1beta1.AddToScheme(scheme))
	// kai.resources so the handlers can store/read kai Projects.
	utilruntime.Must(kaiorg.AddToScheme(scheme))
	testScheme = scheme
}

func ValidateExternalWorkloadInnerSpec(g Gomega, offset int, actual, expected *v1alpha1.ExternalWorkload) {
	if expected == nil {
		return
	}

	if len(expected.Spec.RequestedNodePools) > 0 {
		g.ExpectWithOffset(offset, actual.Spec.RequestedNodePools).To(HaveExactElements(expected.Spec.RequestedNodePools),
			"RequestedNodePools")
	}
	validateRequestedPods(g, offset+1, actual.Spec.RequestedPods, expected.Spec.RequestedPods)
	if expected.Spec.PriorityClassName != "" {
		g.ExpectWithOffset(offset, actual.Spec.PriorityClassName).To(Equal(expected.Spec.PriorityClassName),
			"PriorityClassName")
	}
	if expected.Spec.Preemptibility != "" {
		g.ExpectWithOffset(offset, actual.Spec.Preemptibility).To(Equal(expected.Spec.Preemptibility),
			"Preemptibility")
	}
	if expected.Spec.ImageName != "" {
		g.ExpectWithOffset(offset, actual.Spec.ImageName).To(Equal(expected.Spec.ImageName),
			"ImageName")
	}
	ValidateRequestedResources(g, offset+1, actual.Spec.RequestedResources, expected.Spec.RequestedResources)
	g.ExpectWithOffset(offset, actual.Spec.AdditionalFields).To(Equal(expected.Spec.AdditionalFields),
		"AdditionalFields")
	g.ExpectWithOffset(offset, actual.Spec.Urls).To(Equal(expected.Spec.Urls),
		"Urls")
	if len(expected.Spec.NodeAffinities) > 0 {
		g.ExpectWithOffset(offset, actual.Spec.NodeAffinities).To(ConsistOf(expected.Spec.NodeAffinities),
			"NodeAffinities")
	}
	if expected.Spec.Topology.Name != "" || expected.Spec.Topology.PreferredPlacement != "" || expected.Spec.Topology.RequiredPlacement != "" {
		g.ExpectWithOffset(offset, actual.Spec.Topology).To(Equal(expected.Spec.Topology), "Topology")
	}

	g.ExpectWithOffset(offset, actual.Status.ObservedStatus.Phase).To(Equal(expected.Status.ObservedStatus.Phase),
		"ObservedStatus.Phase")
	g.ExpectWithOffset(offset, actual.Status.ObservedStatus.Message).To(
		Equal(expected.Status.ObservedStatus.Message), "ObservedStatus.Message")
	g.ExpectWithOffset(offset, actual.Status.ObservedStatus.K8sStatus).To(
		Equal(expected.Status.ObservedStatus.K8sStatus), "ObservedStatus.K8sStatus")
	g.ExpectWithOffset(offset, actual.Status.ObservedStatus.Conditions).To(
		ConsistOf(expected.Status.ObservedStatus.Conditions), "ObservedStatus.Conditions")
}

func ValidateRequestedResources(g Gomega, offset int, actual, expected corev1.ResourceRequirements) {
	validateResourceList(g, offset+1, actual.Limits, expected.Limits, "Limits")
	validateResourceList(g, offset+1, actual.Requests, expected.Requests, "Requests")
}

func validateResourceList(g Gomega, offset int, actual, expected corev1.ResourceList, listType string) {
	g.ExpectWithOffset(offset, len(actual)).To(Equal(len(expected)), fmt.Sprintf("ResourceList.%s-len", listType))

	for resourceName, actualResources := range actual {
		expectedResources, found := expected[resourceName]
		g.ExpectWithOffset(offset, found).To(BeTrue(),
			fmt.Sprintf("expected ResourceRequirements.%s of <%s> to be found", listType, resourceName))
		g.ExpectWithOffset(offset, actualResources.Equal(expectedResources)).To(BeTrue(),
			fmt.Sprintf("ResourceRequirements.%s of <%s>, actual: %s, expected: %s",
				listType, resourceName, actualResources.String(), expectedResources.String()))
	}
}

func validateRequestedPods(g Gomega, offset int, actual, expected v1alpha1.RequestedPods) {
	if expected.Replicas != nil {
		g.ExpectWithOffset(offset, actual.Replicas).ToNot(BeNil(), "RequestedPods.Replicas")
		g.ExpectWithOffset(offset, *actual.Replicas).To(Equal(*expected.Replicas), "RequestedPods.Replicas")
	} else {
		g.ExpectWithOffset(offset, actual.Replicas).To(BeNil(), "RequestedPods.Replicas")
	}

	if expected.MinReplicas != nil {
		g.ExpectWithOffset(offset, actual.MinReplicas).ToNot(BeNil(), "RequestedPods.MinReplicas")
		g.ExpectWithOffset(offset, *actual.MinReplicas).To(Equal(*expected.MinReplicas), "RequestedPods.MinReplicas")
	} else {
		g.ExpectWithOffset(offset, actual.MinReplicas).To(BeNil(), "RequestedPods.MinReplicas")
	}

	if expected.MaxReplicas != nil {
		g.ExpectWithOffset(offset, actual.MaxReplicas).ToNot(BeNil(), "RequestedPods.MaxReplicas")
		g.ExpectWithOffset(offset, *actual.MaxReplicas).To(Equal(*expected.MaxReplicas), "RequestedPods.MaxReplicas")
	} else {
		g.ExpectWithOffset(offset, actual.MaxReplicas).To(BeNil(), "RequestedPods.MaxReplicas")
	}

	if expected.Parallelism != nil {
		g.ExpectWithOffset(offset, actual.Parallelism).ToNot(BeNil(), "RequestedPods.Parallelism")
		g.ExpectWithOffset(offset, *actual.Parallelism).To(Equal(*expected.Parallelism), "RequestedPods.Parallelism")
	} else {
		g.ExpectWithOffset(offset, actual.Parallelism).To(BeNil(), "RequestedPods.Parallelism")
	}

	if expected.Completions != nil {
		g.ExpectWithOffset(offset, actual.Completions).ToNot(BeNil(), "RequestedPods.Completions")
		g.ExpectWithOffset(offset, *actual.Completions).To(Equal(*expected.Completions), "RequestedPods.Completions")
	} else {
		g.ExpectWithOffset(offset, actual.Completions).To(BeNil(), "RequestedPods.Completions")
	}
}

func listExternalWorkloadsFromClient(k8sClient client.Reader) []v1alpha1.ExternalWorkload {
	externalWorkloads := &v1alpha1.ExternalWorkloadList{}
	err := k8sClient.List(context.Background(), externalWorkloads)
	if err != nil {
		return nil
	}
	return externalWorkloads.Items
}

func BuildNamespace(name string, labels map[string]string) *corev1.Namespace {
	return &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name:   name,
			Labels: labels,
		},
	}
}

func buildProject(projectName, departmentName string, projectId, departmentId string) *v2.Project {
	return &v2.Project{
		ObjectMeta: metav1.ObjectMeta{
			Name: projectName,
			Labels: map[string]string{
				v2.ProjectIdLabel:                    projectId,
				schedulingv2alpha2.DepartmentIdLabel: departmentId,
			},
		},
		Spec: v2.ProjectSpec{
			Department: departmentName,
		},
	}
}

// buildKaiProject is the kai.resources twin of buildProject, used so tests pass under both values of
// the kai.resources CRD. Same id labels; Spec.Parent replaces Spec.Department.
func buildKaiProject(projectName, departmentName string, projectId, departmentId string) *kaiorg.Project {
	return &kaiorg.Project{
		ObjectMeta: metav1.ObjectMeta{
			Name: projectName,
			Labels: map[string]string{
				v2.ProjectIdLabel:                    projectId,
				schedulingv2alpha2.DepartmentIdLabel: departmentId,
			},
		},
		Spec: kaiorg.ProjectSpec{
			Parent: departmentName,
		},
	}
}

func buildPod(name, namespace, uid string) *corev1.Pod {
	return buildPodInner(name, namespace, uid, constants.RunaiSchedulerName)
}

func buildPodInner(name, namespace, uid, schedulerName string) *corev1.Pod {
	return &corev1.Pod{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Pod",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			UID:               types.UID(uid),
			Name:              name,
			Namespace:         namespace,
			Labels:            map[string]string{},
			CreationTimestamp: metav1.Now(),
		},
		Spec: corev1.PodSpec{
			SchedulerName: schedulerName,
		},
	}
}

func buildDeployment(name, namespace, uid, schedulerName string) *appsv1.Deployment {
	return &appsv1.Deployment{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Deployment",
			APIVersion: "apps/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			UID:               types.UID(uid),
			Name:              name,
			Namespace:         namespace,
			Labels:            map[string]string{},
			CreationTimestamp: metav1.Now(),
		},
		Spec: appsv1.DeploymentSpec{
			Template: corev1.PodTemplateSpec{
				Spec: corev1.PodSpec{
					SchedulerName: schedulerName,
				},
			},
		},
	}
}

func buildDeploymentWithOwner(name, namespace, uid, schedulerName string, owner client.Object) *appsv1.Deployment {
	deployment := buildDeployment(name, namespace, uid, schedulerName)
	deployment.OwnerReferences = []metav1.OwnerReference{
		{
			APIVersion:         owner.GetObjectKind().GroupVersionKind().GroupVersion().String(),
			BlockOwnerDeletion: ptr.To(true),
			Controller:         ptr.To(true),
			Kind:               owner.GetObjectKind().GroupVersionKind().Kind,
			Name:               owner.GetName(),
			UID:                owner.GetUID(),
		},
	}
	return deployment
}

func buildReplicaSet(name, namespace, uid, schedulerName string) *appsv1.ReplicaSet {
	return &appsv1.ReplicaSet{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ReplicaSet",
			APIVersion: "apps/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			UID:               types.UID(uid),
			Name:              name,
			Namespace:         namespace,
			Labels:            map[string]string{},
			CreationTimestamp: metav1.Now(),
		},
		Spec: appsv1.ReplicaSetSpec{
			Template: corev1.PodTemplateSpec{
				Spec: corev1.PodSpec{
					SchedulerName: schedulerName,
				},
			},
		},
	}
}

func buildStatefulSet(name, namespace, uid, schedulerName string) *appsv1.StatefulSet {
	return &appsv1.StatefulSet{
		TypeMeta: metav1.TypeMeta{
			Kind:       "StatefulSet",
			APIVersion: "apps/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			UID:               types.UID(uid),
			Name:              name,
			Namespace:         namespace,
			Labels:            map[string]string{},
			CreationTimestamp: metav1.Now(),
		},
		Spec: appsv1.StatefulSetSpec{
			Template: corev1.PodTemplateSpec{
				Spec: corev1.PodSpec{
					SchedulerName: schedulerName,
				},
			},
		},
	}
}

func buildBatchJob(name, namespace, uid, schedulerName string) *batchv1.Job {
	return &batchv1.Job{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Job",
			APIVersion: "batch/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			UID:               types.UID(uid),
			Name:              name,
			Namespace:         namespace,
			Labels:            map[string]string{},
			CreationTimestamp: metav1.Now(),
		},
		Spec: batchv1.JobSpec{
			Template: corev1.PodTemplateSpec{
				Spec: corev1.PodSpec{
					SchedulerName: schedulerName,
				},
			},
		},
	}
}

func buildCronJob(name, namespace, uid, schedulerName string) *batchv1.CronJob {
	return &batchv1.CronJob{
		TypeMeta: metav1.TypeMeta{
			Kind:       "CronJob",
			APIVersion: "batch/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			UID:               types.UID(uid),
			Name:              name,
			Namespace:         namespace,
			Labels:            map[string]string{},
			CreationTimestamp: metav1.Now(),
		},
		Spec: batchv1.CronJobSpec{
			JobTemplate: batchv1.JobTemplateSpec{
				Spec: batchv1.JobSpec{
					Template: corev1.PodTemplateSpec{
						Spec: corev1.PodSpec{
							SchedulerName: schedulerName,
						},
					},
				},
			},
		},
	}
}

func buildRayJob(name, namespace, uid, schedulerName string) *rayv1alpha1.RayJob {
	return &rayv1alpha1.RayJob{
		TypeMeta: metav1.TypeMeta{
			Kind:       "RayJob",
			APIVersion: "ray.io/v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			UID:               types.UID(uid),
			Name:              name,
			Namespace:         namespace,
			Labels:            map[string]string{},
			CreationTimestamp: metav1.Now(),
		},
		Spec: rayv1alpha1.RayJobSpec{
			RayClusterSpec: &rayv1alpha1.RayClusterSpec{
				HeadGroupSpec: rayv1alpha1.HeadGroupSpec{
					Template: corev1.PodTemplateSpec{
						Spec: corev1.PodSpec{
							SchedulerName: schedulerName,
						},
					},
				},
			},
		},
	}
}

func buildRayJobV1(rayJob *rayv1alpha1.RayJob) *rayv1.RayJob {
	return &rayv1.RayJob{
		TypeMeta: metav1.TypeMeta{
			Kind:       "RayJob",
			APIVersion: "ray.io/v1",
		},
		ObjectMeta: rayJob.ObjectMeta,
		Spec:       rayJob.Spec,
		Status:     rayJob.Status,
	}
}

func buildRayCluster(name, namespace, uid, schedulerName string) *rayv1alpha1.RayCluster {
	return &rayv1alpha1.RayCluster{
		TypeMeta: metav1.TypeMeta{
			Kind:       "RayCluster",
			APIVersion: "ray.io/v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			UID:               types.UID(uid),
			Name:              name,
			Namespace:         namespace,
			Labels:            map[string]string{},
			CreationTimestamp: metav1.Now(),
		},
		Spec: rayv1alpha1.RayClusterSpec{
			HeadGroupSpec: rayv1alpha1.HeadGroupSpec{
				Template: corev1.PodTemplateSpec{
					Spec: corev1.PodSpec{
						SchedulerName: schedulerName,
					},
				},
			},
		},
	}
}

func buildRayClusterV1(rayCluster *rayv1alpha1.RayCluster) *rayv1.RayCluster {
	return &rayv1.RayCluster{
		TypeMeta: metav1.TypeMeta{
			Kind:       "RayCluster",
			APIVersion: "ray.io/v1",
		},
		ObjectMeta: rayCluster.ObjectMeta,
		Spec:       rayCluster.Spec,
		Status:     rayCluster.Status,
	}
}

func buildRayService(name, namespace, uid, schedulerName string) *rayv1alpha1.RayService {
	return &rayv1alpha1.RayService{
		TypeMeta: metav1.TypeMeta{
			Kind:       "RayService",
			APIVersion: "ray.io/v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			UID:               types.UID(uid),
			Name:              name,
			Namespace:         namespace,
			Labels:            map[string]string{},
			CreationTimestamp: metav1.Now(),
		},
		Spec: rayv1alpha1.RayServiceSpec{
			RayClusterSpec: rayv1alpha1.RayClusterSpec{
				HeadGroupSpec: rayv1alpha1.HeadGroupSpec{
					Template: corev1.PodTemplateSpec{
						Spec: corev1.PodSpec{
							SchedulerName: schedulerName,
						},
					},
				},
			},
		},
	}
}

func buildRayServiceV1(rayService *rayv1alpha1.RayService) *rayv1.RayService {
	return &rayv1.RayService{
		TypeMeta: metav1.TypeMeta{
			Kind:       "RayService",
			APIVersion: "ray.io/v1",
		},
		ObjectMeta: rayService.ObjectMeta,
		Spec:       rayService.Spec,
		Status:     rayService.Status,
	}
}

func buildKNativeService(name, namespace, uid, schedulerName string) *servingv1.Service {
	return &servingv1.Service{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Service",
			APIVersion: "serving.knative.dev/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			UID:               types.UID(uid),
			Name:              name,
			Namespace:         namespace,
			Labels:            map[string]string{},
			CreationTimestamp: metav1.Now(),
		},
		Spec: servingv1.ServiceSpec{
			ConfigurationSpec: servingv1.ConfigurationSpec{
				Template: servingv1.RevisionTemplateSpec{
					Spec: servingv1.RevisionSpec{
						PodSpec: corev1.PodSpec{
							SchedulerName: schedulerName,
						},
					},
				},
			},
		},
	}
}

func buildKNativeRevision(name, namespace, uid string) *servingv1.Revision {
	return &servingv1.Revision{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Revision",
			APIVersion: "serving.knative.dev/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			UID:               types.UID(uid),
			Name:              name,
			Namespace:         namespace,
			CreationTimestamp: metav1.Now(),
		},
	}
}

func buildKNativeConfiguration(name, namespace, uid string) *servingv1.Configuration {
	return &servingv1.Configuration{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Configuration",
			APIVersion: "serving.knative.dev/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			UID:               types.UID(uid),
			Name:              name,
			Namespace:         namespace,
			CreationTimestamp: metav1.Now(),
		},
	}
}

func buildTFJob(name, namespace, uid, schedulerName1, schedulerName2 string) *kubeflowv1.TFJob {
	return &kubeflowv1.TFJob{
		TypeMeta: metav1.TypeMeta{
			Kind:       "TFJob",
			APIVersion: "kubeflow.org/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			UID:               types.UID(uid),
			Name:              name,
			Namespace:         namespace,
			CreationTimestamp: metav1.Now(),
		},
		Spec: kubeflowv1.TFJobSpec{
			TFReplicaSpecs: map[kubeflowv1.ReplicaType]*kubeflowv1.ReplicaSpec{
				"Worker": {
					Template: corev1.PodTemplateSpec{
						Spec: corev1.PodSpec{
							SchedulerName: schedulerName1,
						},
					},
				},
				"PS": {
					Template: corev1.PodTemplateSpec{
						Spec: corev1.PodSpec{
							SchedulerName: schedulerName2,
						},
					},
				},
			},
		},
	}
}

func buildPytorchJob(name, namespace, uid, schedulerName1, schedulerName2 string) *kubeflowv1.PyTorchJob {
	return &kubeflowv1.PyTorchJob{
		TypeMeta: metav1.TypeMeta{
			Kind:       "PyTorchJob",
			APIVersion: "kubeflow.org/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			UID:               types.UID(uid),
			Name:              name,
			Namespace:         namespace,
			CreationTimestamp: metav1.Now(),
		},
		Spec: kubeflowv1.PyTorchJobSpec{
			PyTorchReplicaSpecs: map[kubeflowv1.ReplicaType]*kubeflowv1.ReplicaSpec{
				"Worker": {
					Template: corev1.PodTemplateSpec{
						Spec: corev1.PodSpec{
							SchedulerName: schedulerName1,
						},
					},
				},
				"Master": {
					Template: corev1.PodTemplateSpec{
						Spec: corev1.PodSpec{
							SchedulerName: schedulerName2,
						},
					},
				},
			},
		},
	}
}

func buildXGBoostJob(name, namespace, uid, schedulerName1, schedulerName2 string) *kubeflowv1.XGBoostJob {
	return &kubeflowv1.XGBoostJob{
		TypeMeta: metav1.TypeMeta{
			Kind:       "XGBoostJob",
			APIVersion: "kubeflow.org/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			UID:               types.UID(uid),
			Name:              name,
			Namespace:         namespace,
			CreationTimestamp: metav1.Now(),
		},
		Spec: kubeflowv1.XGBoostJobSpec{
			XGBReplicaSpecs: map[kubeflowv1.ReplicaType]*kubeflowv1.ReplicaSpec{
				"Worker": {
					Template: corev1.PodTemplateSpec{
						Spec: corev1.PodSpec{
							SchedulerName: schedulerName1,
						},
					},
				},
				"Master": {
					Template: corev1.PodTemplateSpec{
						Spec: corev1.PodSpec{
							SchedulerName: schedulerName2,
						},
					},
				},
			},
		},
	}
}

func buildJAXJob(name, namespace, uid, schedulerName1, schedulerName2 string) *kubeflowv1.JAXJob {
	return &kubeflowv1.JAXJob{
		TypeMeta: metav1.TypeMeta{
			Kind:       "JAXJob",
			APIVersion: "kubeflow.org/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			UID:               types.UID(uid),
			Name:              name,
			Namespace:         namespace,
			CreationTimestamp: metav1.Now(),
		},
		Spec: kubeflowv1.JAXJobSpec{
			JAXReplicaSpecs: map[kubeflowv1.ReplicaType]*kubeflowv1.ReplicaSpec{
				"Worker": {
					Template: corev1.PodTemplateSpec{
						Spec: corev1.PodSpec{
							SchedulerName: schedulerName1,
						},
					},
				},
				"Master": {
					Template: corev1.PodTemplateSpec{
						Spec: corev1.PodSpec{
							SchedulerName: schedulerName2,
						},
					},
				},
			},
		},
	}
}

func buildMPIJobV1(name, namespace, uid, schedulerName1, schedulerName2 string) *kubeflowv1.MPIJob {
	return &kubeflowv1.MPIJob{
		TypeMeta: metav1.TypeMeta{
			Kind:       "MPIJob",
			APIVersion: "kubeflow.org/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			UID:               types.UID(uid),
			Name:              name,
			Namespace:         namespace,
			CreationTimestamp: metav1.Now(),
		},
		Spec: kubeflowv1.MPIJobSpec{
			MPIReplicaSpecs: map[kubeflowv1.ReplicaType]*kubeflowv1.ReplicaSpec{
				"Worker": {
					Template: corev1.PodTemplateSpec{
						Spec: corev1.PodSpec{
							SchedulerName: schedulerName1,
						},
					},
				},
				"Launcher": {
					Template: corev1.PodTemplateSpec{
						Spec: corev1.PodSpec{
							SchedulerName: schedulerName2,
						},
					},
				},
			},
		},
	}
}

func buildMPIJobV2beta1(name, namespace, uid, schedulerName1, schedulerName2 string) *kubeflowv2beta1.MPIJob {
	return &kubeflowv2beta1.MPIJob{
		TypeMeta: metav1.TypeMeta{
			Kind:       "MPIJob",
			APIVersion: "kubeflow.org/v2beta1",
		},
		ObjectMeta: metav1.ObjectMeta{
			UID:               types.UID(uid),
			Name:              name,
			Namespace:         namespace,
			CreationTimestamp: metav1.Now(),
		},
		Spec: kubeflowv2beta1.MPIJobSpec{
			MPIReplicaSpecs: map[kubeflowv2beta1.MPIReplicaType]*kubeflowv2beta1.ReplicaSpec{
				"Worker": {
					Template: corev1.PodTemplateSpec{
						Spec: corev1.PodSpec{
							SchedulerName: schedulerName1,
						},
					},
				},
				"Launcher": {
					Template: corev1.PodTemplateSpec{
						Spec: corev1.PodSpec{
							SchedulerName: schedulerName2,
						},
					},
				},
			},
		},
	}
}

func buildNotebook(name, namespace, uid, schedulerName string) *kubeflowv1beta1.Notebook {
	return &kubeflowv1beta1.Notebook{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Notebook",
			APIVersion: "kubeflow.org/v1beta1",
		},
		ObjectMeta: metav1.ObjectMeta{
			UID:               types.UID(uid),
			Name:              name,
			Namespace:         namespace,
			Generation:        1,
			CreationTimestamp: metav1.Now(),
		},
		Spec: kubeflowv1beta1.NotebookSpec{
			Template: kubeflowv1beta1.NotebookTemplateSpec{
				Spec: corev1.PodSpec{
					SchedulerName: schedulerName,
				},
			},
		},
	}
}

func buildKubevirtVMI(name, namespace, uid, schedulerName string) *kubevirtv1.VirtualMachineInstance {
	return &kubevirtv1.VirtualMachineInstance{
		TypeMeta: metav1.TypeMeta{
			Kind:       "VirtualMachineInstance",
			APIVersion: "kubevirt.io/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			UID:               types.UID(uid),
			Name:              name,
			Namespace:         namespace,
			Generation:        1,
			CreationTimestamp: metav1.Now(),
		},
		Spec: kubevirtv1.VirtualMachineInstanceSpec{
			SchedulerName: schedulerName,
		},
	}
}

func buildNvcf(name, namespace, uid, schedulerName string) *nvcfv1.SPOTRequest {
	return &nvcfv1.SPOTRequest{
		TypeMeta: metav1.TypeMeta{
			Kind:       "SPOTRequest",
			APIVersion: "nvcf.nvidia.com/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			UID:               types.UID(uid),
			Name:              name,
			Namespace:         namespace,
			Generation:        1,
			CreationTimestamp: metav1.Now(),
		},
		Spec: nvcfv1.SPOTRequestSpec{
			FunctionVersionID: name,
		},
	}
}

func buildDaemonSet(name, namespace, uid, schedulerName string) *appsv1.DaemonSet {
	return &appsv1.DaemonSet{
		TypeMeta: metav1.TypeMeta{
			Kind:       "DaemonSet",
			APIVersion: "apps/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			UID:               types.UID(uid),
			Name:              name,
			Namespace:         namespace,
			Generation:        1,
			CreationTimestamp: metav1.Now(),
		},
		Spec: appsv1.DaemonSetSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{"app": name},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{"app": name},
				},
				Spec: corev1.PodSpec{
					SchedulerName: schedulerName,
					Containers: []corev1.Container{
						{
							Name:  "main",
							Image: "ubuntu",
						},
					},
				},
			},
		},
	}
}

func CreateWorkloadsConfigMapObject() *corev1.ConfigMap {
	m := []workloadTypeConfig.WorkloadTypeConfig{
		{
			TypeName:       "JAXJob",
			Group:          "kubeflow.org",
			CategoryName:   "Train",
			PriorityName:   "",
			Preemptibility: "non-preemptible",
		},
		{
			TypeName:       "Notebook",
			Group:          "kubeflow.org",
			CategoryName:   "Build",
			PriorityName:   "",
			Preemptibility: "non-preemptible",
		},
		{
			TypeName:       "Deployment",
			Group:          "apps",
			CategoryName:   "Deploy",
			PriorityName:   "",
			Preemptibility: "preemptible",
		},
		{
			TypeName:       "Deployment",
			Group:          "some-other-groupppp",
			CategoryName:   "Deploy222",
			PriorityName:   "high2222",
			Preemptibility: "preemptible",
		},
		{
			TypeName:       "PyTorchJob",
			Group:          "kubeflow.org",
			CategoryName:   "Train",
			PriorityName:   "low",
			Preemptibility: "non-preemptible",
		},
		{
			TypeName:       "XGBoostJob",
			Group:          "kubeflow.org",
			CategoryName:   "Train",
			PriorityName:   "",
			Preemptibility: "non-preemptible",
		},
		{
			TypeName:       "MPIJob",
			Group:          "kubeflow.org",
			CategoryName:   "Train",
			PriorityName:   "",
			Preemptibility: "non-preemptible",
		},
		{
			TypeName:       "StatefulSet",
			Group:          "apps",
			CategoryName:   "Deploy",
			PriorityName:   "very-high",
			Preemptibility: "preemptible",
		},
		{
			TypeName:       "RayCluster",
			Group:          "ray.io",
			CategoryName:   "Train",
			PriorityName:   "low",
			Preemptibility: "preemptible",
		},
		{
			TypeName:       "RayCluster",
			Group:          "some-other-groupppp",
			CategoryName:   "Train33",
			PriorityName:   "low33",
			Preemptibility: "preemptible",
		},
		{
			TypeName:       "RayJob",
			Group:          "ray.io",
			CategoryName:   "Train",
			PriorityName:   "low",
			Preemptibility: "preemptible",
		},
		{
			TypeName:       "DynamoGraphDeployment",
			Group:          "nvidia.com",
			CategoryName:   "Deploy",
			PriorityName:   "very-high",
			Preemptibility: "preemptible",
		},
	}
	jsonM, _ := json.Marshal(m)
	cm := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      constants2.WorkloadsConfigMapName,
			Namespace: constants2.WorkloadsConfigMapNamespace,
		},
	}
	cm.Data = map[string]string{constants2.WorkloadsConfigMapTypesKey: string(jsonM)}
	list := []string{"Train", "Deploy", "Inference", "Build"}
	categoriesJson, _ := json.Marshal(list)
	cm.Data[constants2.WorkloadsConfigMapCategoriesKey] = string(categoriesJson)
	return cm
}
