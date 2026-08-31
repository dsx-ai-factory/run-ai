/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package external_workload_integrator

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"k8s.io/utils/ptr"

	kaiv1alpha1 "github.com/kai-scheduler/kai-resource-management-api/kai/v1alpha1"
	rayv1alpha1 "github.com/ray-project/kuberay/ray-operator/apis/ray/v1alpha1"
	nvcfv1 "github.com/run-ai/runai/runai-cluster/cluster/external-workload-integrator/pkg/thirdpartyapi/nvcf/v1"
	rayv1 "github.com/run-ai/runai/runai-cluster/cluster/external-workload-integrator/pkg/thirdpartyapi/ray/v1"
	"github.com/run-ai/runai/runai-cluster/cluster/integration-tests/common"
	"github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1alpha1"
	v2 "github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v2"
	schedulingv2alpha2 "github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/scheduling/v2alpha2"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	resourcev1 "k8s.io/api/resource/v1"
	resourcev1beta1 "k8s.io/api/resource/v1beta1"
	resourcev1beta2 "k8s.io/api/resource/v1beta2"
	apimeta "k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func validateEwStatus(ew *v1alpha1.ExternalWorkload, expectedObservedStatus v1alpha1.ExternalWorkloadObservedStatus, testName string) {
	Expect(ew.Status.ObservedStatus.Phase).To(Equal(expectedObservedStatus.Phase),
		"Test: <%s>, ObservedStatus.Phase", testName)
	Expect(ew.Status.ObservedStatus.Message).To(Equal(expectedObservedStatus.Message),
		"Test: <%s>, ObservedStatus.Message", testName)
	Expect(ew.Status.ObservedStatus.K8sStatus).To(Equal(expectedObservedStatus.K8sStatus),
		"Test: <%s>, ObservedStatus.K8sStatus", testName)
	Expect(ew.Status.ObservedStatus.Conditions).To(ConsistOf(expectedObservedStatus.Conditions),
		"Test: <%s>, ObservedStatus.Conditions", testName)
}

func validateEwSpec(ew *v1alpha1.ExternalWorkload, expectedSpec v1alpha1.ExternalWorkload, testName string) {
	Expect(ew.Spec.OriginalWorkloadUID).To(Equal(expectedSpec.Spec.OriginalWorkloadUID),
		"Test: <%s>, Spec.OriginalWorkloadUID", testName)
	Expect(ew.Spec.OriginalWorkloadName).To(Equal(expectedSpec.Spec.OriginalWorkloadName),
		"Test: <%s>, Spec.OriginalWorkloadName", testName)
	Expect(ew.Spec.Type).To(Equal(expectedSpec.Spec.Type), "Test: <%s>, Spec.Type", testName)
	Expect(ew.Spec.ProjectName).To(Equal(expectedSpec.Spec.ProjectName), "Test: <%s>, Spec.ProjectName", testName)
	Expect(ew.Spec.ProjectId).To(Equal(expectedSpec.Spec.ProjectId), "Test: <%s>, Spec.ProjectId", testName)
	Expect(ew.Spec.DepartmentName).To(Equal(expectedSpec.Spec.DepartmentName), "Test: <%s>, Spec.DepartmentName", testName)
	Expect(ew.Spec.DepartmentId).To(Equal(expectedSpec.Spec.DepartmentId), "Test: <%s>, Spec.DepartmentId", testName)
	Expect(ew.Spec.RequestedPods).To(Equal(expectedSpec.Spec.RequestedPods), "Test: <%s>, Spec.RequestedPods", testName)
	Expect(ew.Spec.ImageName).To(Equal(expectedSpec.Spec.ImageName), "Test: <%s>, Spec.ImageName", testName)
	// Expect(ew.Spec.PriorityClassName).To(Equal(expectedSpec.Spec.PriorityClassName), "Test: <%s>, Spec.PriorityClassName", testName)
	// Expect(ew.Spec.RequestedNodePools).To(ConsistOf(expectedSpec.Spec.RequestedNodePools), "Test: <%s>, Spec.RequestedNodePools", testName)
}

func waitForExternalWorkloadToExist(k8sClient client.Reader, originalObject client.Object,
	testName string, testCreatedObjects map[string]client.Object,
) *v1alpha1.ExternalWorkload {
	var externalWorkloadOfObject *v1alpha1.ExternalWorkload = nil
	Eventually(func(g Gomega) {
		externalWorkloadOfObject = findExternalWorkloadOfObject(k8sClient,
			originalObject.GetUID(), originalObject.GetNamespace())
		g.Expect(externalWorkloadOfObject).ToNot(BeNil(),
			"Test: <%s>, should find external workload for object <%s> with UID <%s>",
			testName, originalObject.GetObjectKind().GroupVersionKind().String(), originalObject.GetUID())
		testCreatedObjects[externalWorkloadOfObject.Name] = externalWorkloadOfObject
	}, timeout, interval).Should(Succeed(), fmt.Sprintf("Test <%s>", testName))
	return externalWorkloadOfObject
}

func waitForExternalWorkloadToBeDeleted(k8sClient client.Reader, originalObject client.Object,
	testName string,
) {
	var externalWorkloadOfObject *v1alpha1.ExternalWorkload = nil
	Eventually(func(g Gomega) {
		externalWorkloadOfObject = findExternalWorkloadOfObject(k8sClient,
			originalObject.GetUID(), originalObject.GetNamespace())
		g.Expect(externalWorkloadOfObject).To(BeNil(),
			"Test: <%s>, should not find external workload for object <%s> with UID <%s>",
			testName, originalObject.GetObjectKind().GroupVersionKind().String(), originalObject.GetUID())
	}, timeout, interval).Should(Succeed(), fmt.Sprintf("Test <%s>", testName))
	return
}

func findExternalWorkloadOfObject(k8sClient client.Reader,
	objectUID types.UID, namespace string,
) *v1alpha1.ExternalWorkload {
	externalWorkloads := listExternalWorkloadsFromClient(k8sClient)
	for i, externalWorkload := range externalWorkloads {
		if externalWorkload.Spec.OriginalWorkloadUID == objectUID && externalWorkload.Namespace == namespace {
			return &externalWorkloads[i]
		}
	}
	return nil
}

func listExternalWorkloadsFromClient(k8sClient client.Reader) []v1alpha1.ExternalWorkload {
	externalWorkloads := &v1alpha1.ExternalWorkloadList{}
	err := k8sClient.List(context.Background(), externalWorkloads)
	if err != nil {
		return nil
	}
	return externalWorkloads.Items
}

func createExpectedEw(ewObjectMeta metav1.ObjectMeta, ewType string, project *kaiv1alpha1.Project, department *kaiv1alpha1.Department, requestedPods v1alpha1.RequestedPods, imageName string) v1alpha1.ExternalWorkload {
	return v1alpha1.ExternalWorkload{
		ObjectMeta: metav1.ObjectMeta{
			Name:      ewObjectMeta.Name,
			Namespace: ewObjectMeta.Namespace,
		},
		Spec: v1alpha1.ExternalWorkloadSpec{
			OriginalWorkloadUID:  ewObjectMeta.UID,
			OriginalWorkloadName: ewObjectMeta.Name,
			Type:                 ewType,
			ProjectName:          project.Name,
			ProjectId:            project.Labels[v2.ProjectIdLabel],
			DepartmentName:       department.Name,
			DepartmentId:         project.Labels[schedulingv2alpha2.DepartmentIdLabel],
			RequestedPods:        requestedPods,
			ImageName:            imageName,
			// RequestedNodePools:   []string{"default"},
			// PriorityClassName:    priorityClassName,
		},
	}
}

func buildNVCF(name, namespace, uid string) *nvcfv1.SPOTRequest {
	return &nvcfv1.SPOTRequest{
		TypeMeta: metav1.TypeMeta{
			Kind:       "SPOTRequest",
			APIVersion: "egx.nvidia.io/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			UID:               types.UID(uuid.NewString()),
			Name:              name,
			Namespace:         namespace,
			CreationTimestamp: metav1.Now(),
		},
		Spec: nvcfv1.SPOTRequestSpec{
			FunctionID:        "some-function-id",
			FunctionVersionID: uid,
		},
		Status: nvcfv1.SPOTRequestStatus{},
	}
}

func buildNVCFPod(spot *nvcfv1.SPOTRequest, imageName string) *corev1.Pod {
	podSpec := common.CreatePodSpec("sr", "", "", "", "", imageName)

	podSpec.InitContainers = []corev1.Container{
		{
			Name: "init-sr",
			Env: []corev1.EnvVar{
				{
					Name:  "INFERENCE_CONTAINER",
					Value: imageName,
				},
			},
			Image:   imageName,
			Command: []string{"bash", "-c", "sleep infinity"},
			Resources: corev1.ResourceRequirements{
				Requests: corev1.ResourceList{},
			},
		},
	}
	return &corev1.Pod{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Pod",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			OwnerReferences: []metav1.OwnerReference{{
				APIVersion:         "egx.nvidia.io/v1",
				BlockOwnerDeletion: ptr.To(true),
				Controller:         ptr.To(true),
				Kind:               "SPOTRequest",
				Name:               spot.GetName(),
				UID:                spot.GetUID(),
			}},
			UID:       types.UID(uuid.NewString()),
			Name:      spot.Name,
			Namespace: spot.Namespace,
			Labels: map[string]string{
				"function-version-id": spot.Spec.FunctionVersionID,
			},
			CreationTimestamp: metav1.Now(),
		},
		Spec: podSpec,
	}
}

// ExpectedExternalWorkloadSpec defines the expected specification for an external workload
type ExpectedExternalWorkloadSpec struct {
	PriorityClassName  *string
	Replicas           *int32
	MinReplicas        *int32
	MaxReplicas        *int32
	ImageName          *string
	ExpectedStatuses   []string
	RequestedResources *corev1.ResourceRequirements
}

// ValidateEWIntegration validates the provided EW against the expected values.
func ValidateEWIntegration(
	ew *v1alpha1.ExternalWorkload,
	expectedSpec ExpectedExternalWorkloadSpec,
) {
	Expect(ew).ToNot(BeNil(), "EW should not be nil")

	// Validate PriorityClassName if provided
	if expectedSpec.PriorityClassName != nil {
		Expect(ew.Spec.PriorityClassName).To(Equal(*expectedSpec.PriorityClassName), "EW PriorityClassName should match expected")
	}

	// Validate replicas if provided
	if expectedSpec.Replicas != nil {
		Expect(ew.Spec.RequestedPods.Replicas).ToNot(BeNil())
		Expect(*ew.Spec.RequestedPods.Replicas).To(Equal(*expectedSpec.Replicas), "EW Replicas should match expected")
	}

	// Validate minReplicas if provided
	if expectedSpec.MinReplicas != nil {
		Expect(ew.Spec.RequestedPods.MinReplicas).ToNot(BeNil())
		Expect(*ew.Spec.RequestedPods.MinReplicas).To(Equal(*expectedSpec.MinReplicas), "EW MinReplicas should match expected")
	}

	// Validate maxReplicas if provided
	if expectedSpec.MaxReplicas != nil {
		Expect(ew.Spec.RequestedPods.MaxReplicas).ToNot(BeNil())
		Expect(*ew.Spec.RequestedPods.MaxReplicas).To(Equal(*expectedSpec.MaxReplicas), "EW MaxReplicas should match expected")
	}

	// Validate imageName if provided
	if expectedSpec.ImageName != nil {
		Expect(ew.Spec.ImageName).To(Equal(*expectedSpec.ImageName), "EW ImageName should match expected")
	}

	// Validate requested resources if provided
	if expectedSpec.RequestedResources != nil {
		if expectedSpec.RequestedResources.Requests != nil {
			for resourceName, expectedQty := range expectedSpec.RequestedResources.Requests {
				actualQty, ok := ew.Spec.RequestedResources.Requests[resourceName]
				Expect(ok).To(BeTrue(), "EW RequestedResources.Requests should contain %s", resourceName)
				Expect(actualQty.Equal(expectedQty)).To(BeTrue(),
					"EW RequestedResources.Requests[%s] should be %s, got %s", resourceName, expectedQty.String(), actualQty.String())
			}
		}
		if expectedSpec.RequestedResources.Limits != nil {
			for resourceName, expectedQty := range expectedSpec.RequestedResources.Limits {
				actualQty, ok := ew.Spec.RequestedResources.Limits[resourceName]
				Expect(ok).To(BeTrue(), "EW RequestedResources.Limits should contain %s", resourceName)
				Expect(actualQty.Equal(expectedQty)).To(BeTrue(),
					"EW RequestedResources.Limits[%s] should be %s, got %s", resourceName, expectedQty.String(), actualQty.String())
			}
		}
	}

	// Validate status/phase if provided
	if len(expectedSpec.ExpectedStatuses) > 0 {
		phase := string(ew.Status.ObservedStatus.Phase)
		found := false
		for _, s := range expectedSpec.ExpectedStatuses {
			if phase == s {
				found = true
				break
			}
		}
		Expect(found).To(BeTrue(), "EW status phase %s should be in expected statuses", phase)
	}
}

// BuildRayCluster creates a RayCluster object with the specified parameters
func BuildRayCluster(name, namespace string, projectName string, workerReplicas, minReplicas, maxReplicas int32, imageName string, priorityClassName string) *rayv1alpha1.RayCluster {
	return &rayv1alpha1.RayCluster{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			Labels: map[string]string{
				"controller-tools.k8s.io": "1.0",
				"project":                 projectName,
				"runai/queue":             projectName,
			},
		},
		Spec: rayv1alpha1.RayClusterSpec{
			RayVersion: "2.4.0",
			HeadGroupSpec: rayv1alpha1.HeadGroupSpec{
				RayStartParams: map[string]string{
					"dashboard-host": "0.0.0.0",
				},
				Template: corev1.PodTemplateSpec{
					Spec: corev1.PodSpec{
						Containers: []corev1.Container{
							{
								Name:  "ray-head",
								Image: imageName,
								Ports: []corev1.ContainerPort{
									{ContainerPort: 6379, Name: "gcs"},
									{ContainerPort: 8265, Name: "dashboard"},
									{ContainerPort: 10001, Name: "client"},
								},
								Lifecycle: &corev1.Lifecycle{
									PreStop: &corev1.LifecycleHandler{
										Exec: &corev1.ExecAction{
											Command: []string{"/bin/sh", "-c", "ray stop"},
										},
									},
								},
								Resources: corev1.ResourceRequirements{
									Limits: corev1.ResourceList{
										corev1.ResourceCPU:    resource.MustParse("500m"),
										corev1.ResourceMemory: resource.MustParse("2G"),
									},
									Requests: corev1.ResourceList{
										corev1.ResourceCPU:    resource.MustParse("500m"),
										corev1.ResourceMemory: resource.MustParse("2G"),
									},
								},
							},
						},
						SchedulerName:     "runai-scheduler",
						PriorityClassName: priorityClassName,
					},
				},
			},
			WorkerGroupSpecs: []rayv1alpha1.WorkerGroupSpec{
				{
					Replicas:    &workerReplicas,
					MinReplicas: &minReplicas,
					MaxReplicas: &maxReplicas,
					GroupName:   "small-group",
					RayStartParams: map[string]string{
						"num-cpus": "1",
					},
					Template: corev1.PodTemplateSpec{
						Spec: corev1.PodSpec{
							Containers: []corev1.Container{
								{
									Name:  "ray-worker",
									Image: imageName,
									Lifecycle: &corev1.Lifecycle{
										PreStop: &corev1.LifecycleHandler{
											Exec: &corev1.ExecAction{
												Command: []string{"/bin/sh", "-c", "ray stop"},
											},
										},
									},
									Resources: corev1.ResourceRequirements{
										Limits: corev1.ResourceList{
											corev1.ResourceCPU:    resource.MustParse("500m"),
											corev1.ResourceMemory: resource.MustParse("2G"),
										},
										Requests: corev1.ResourceList{
											corev1.ResourceCPU:    resource.MustParse("500m"),
											corev1.ResourceMemory: resource.MustParse("2G"),
										},
									},
								},
							},
							SchedulerName: "runai-scheduler",
						},
					},
				},
			},
		},
	}
}

// BuildRayJob creates a RayJob object with the specified parameters
func BuildRayJob(name, namespace string, projectName string, workerReplicas, minReplicas, maxReplicas int32, imageName string, priorityClassName string) *rayv1.RayJob {
	return &rayv1.RayJob{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			Labels: map[string]string{
				"project":     projectName,
				"runai/queue": projectName,
			},
		},
		Spec: rayv1alpha1.RayJobSpec{
			ShutdownAfterJobFinishes: true,
			TTLSecondsAfterFinished:  int32(60),
			Entrypoint:               "python /home/ray/samples/sample_code.py",
			RuntimeEnvYAML: `pip:
  - requests==2.26.0
  - pendulum==2.1.2
env_vars:
  counter_name: "test_counter"`,
			RayClusterSpec: &rayv1alpha1.RayClusterSpec{
				RayVersion: "2.4.0",
				HeadGroupSpec: rayv1alpha1.HeadGroupSpec{
					RayStartParams: map[string]string{
						"dashboard-host": "0.0.0.0",
						"num-cpus":       "1",
					},
					Template: corev1.PodTemplateSpec{
						Spec: corev1.PodSpec{
							Containers: []corev1.Container{
								{
									Name:  "ray-head",
									Image: imageName,
									Ports: []corev1.ContainerPort{
										{ContainerPort: 6379, Name: "gcs-server"},
										{ContainerPort: 8265, Name: "dashboard"},
										{ContainerPort: 10001, Name: "client"},
										{ContainerPort: 8000, Name: "serve"},
									},
									Resources: corev1.ResourceRequirements{
										Limits: corev1.ResourceList{
											corev1.ResourceCPU: resource.MustParse("1"),
										},
										Requests: corev1.ResourceList{
											corev1.ResourceCPU: resource.MustParse("200m"),
										},
									},
									VolumeMounts: []corev1.VolumeMount{
										{
											Name:      "code-sample",
											MountPath: "/home/ray/samples",
										},
									},
								},
							},
							Volumes: []corev1.Volume{
								{
									Name: "code-sample",
									VolumeSource: corev1.VolumeSource{
										ConfigMap: &corev1.ConfigMapVolumeSource{
											LocalObjectReference: corev1.LocalObjectReference{
												Name: "ray-job-code-sample",
											},
											Items: []corev1.KeyToPath{
												{
													Key:  "sample_code.py",
													Path: "sample_code.py",
												},
											},
										},
									},
								},
							},
							SchedulerName:     "runai-scheduler",
							PriorityClassName: priorityClassName,
						},
					},
				},
				WorkerGroupSpecs: []rayv1alpha1.WorkerGroupSpec{
					{
						Replicas:       &workerReplicas,
						MinReplicas:    &minReplicas,
						MaxReplicas:    &maxReplicas,
						GroupName:      "small-group",
						RayStartParams: map[string]string{},
						Template: corev1.PodTemplateSpec{
							Spec: corev1.PodSpec{
								Containers: []corev1.Container{
									{
										Name:  "ray-worker",
										Image: imageName,
										Lifecycle: &corev1.Lifecycle{
											PreStop: &corev1.LifecycleHandler{
												Exec: &corev1.ExecAction{
													Command: []string{"/bin/sh", "-c", "ray stop"},
												},
											},
										},
										Resources: corev1.ResourceRequirements{
											Limits: corev1.ResourceList{
												corev1.ResourceCPU: resource.MustParse("1"),
											},
											Requests: corev1.ResourceList{
												corev1.ResourceCPU: resource.MustParse("200m"),
											},
										},
									},
								},
								SchedulerName: "runai-scheduler",
							},
						},
					},
				},
			},
		},
	}
}

// BuildRayService creates a RayService object with the specified parameters
func BuildRayService(name, namespace string, projectName string, workerReplicas, minReplicas, maxReplicas int32, imageName string, priorityClassName string) *rayv1alpha1.RayService {
	return &rayv1alpha1.RayService{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			Labels: map[string]string{
				"runai/queue": projectName,
			},
			Annotations: map[string]string{
				"user": "test",
			},
		},
		Spec: rayv1alpha1.RayServiceSpec{
			ServiceUnhealthySecondThreshold:    ptr.To(int32(900)),
			DeploymentUnhealthySecondThreshold: ptr.To(int32(300)),
			ServeConfigV2: `applications:
  - name: text_ml_app
    import_path: text_ml.app
    route_prefix: /summarize_translate
    runtime_env:
      working_dir: "https://github.com/ray-project/serve_config_examples/archive/36862c251615e258a58285934c7c41cffd1ee3b7.zip"
      pip:
        - torch
        - transformers
    deployments:
      - name: Translator
        num_replicas: 1
        ray_actor_options:
          num_cpus: 0.2
        user_config:
          language: french
      - name: Summarizer
        num_replicas: 1
        ray_actor_options:
          num_cpus: 0.2
`,
			RayClusterSpec: rayv1alpha1.RayClusterSpec{
				RayVersion: "2.6.3",
				HeadGroupSpec: rayv1alpha1.HeadGroupSpec{
					RayStartParams: map[string]string{
						"dashboard-host": "0.0.0.0",
					},
					Template: corev1.PodTemplateSpec{
						ObjectMeta: metav1.ObjectMeta{
							Labels: map[string]string{
								"runai/queue": projectName,
							},
							Annotations: map[string]string{
								"user": "test",
							},
						},
						Spec: corev1.PodSpec{
							SchedulerName:     "runai-scheduler",
							PriorityClassName: priorityClassName,
							Containers: []corev1.Container{
								{
									Name:  "ray-head",
									Image: imageName,
									Resources: corev1.ResourceRequirements{
										Limits: corev1.ResourceList{
											corev1.ResourceCPU:    resource.MustParse("2"),
											corev1.ResourceMemory: resource.MustParse("2Gi"),
										},
										Requests: corev1.ResourceList{
											corev1.ResourceCPU:    resource.MustParse("2"),
											corev1.ResourceMemory: resource.MustParse("2Gi"),
										},
									},
									Ports: []corev1.ContainerPort{
										{ContainerPort: 6379, Name: "gcs-server"},
										{ContainerPort: 8265, Name: "dashboard"},
										{ContainerPort: 10001, Name: "client"},
										{ContainerPort: 8000, Name: "serve"},
									},
								},
							},
						},
					},
				},
				WorkerGroupSpecs: []rayv1alpha1.WorkerGroupSpec{
					{
						Replicas:       &workerReplicas,
						MinReplicas:    &minReplicas,
						MaxReplicas:    &maxReplicas,
						GroupName:      "small-group",
						RayStartParams: map[string]string{},
						Template: corev1.PodTemplateSpec{
							ObjectMeta: metav1.ObjectMeta{
								Labels: map[string]string{
									"runai/queue": projectName,
								},
								Annotations: map[string]string{
									"user": "test",
								},
							},
							Spec: corev1.PodSpec{
								SchedulerName: "runai-scheduler",
								Containers: []corev1.Container{
									{
										Name:  "ray-worker",
										Image: imageName,
										Lifecycle: &corev1.Lifecycle{
											PreStop: &corev1.LifecycleHandler{
												Exec: &corev1.ExecAction{
													Command: []string{"/bin/sh", "-c", "ray stop"},
												},
											},
										},
										Resources: corev1.ResourceRequirements{
											Limits: corev1.ResourceList{
												corev1.ResourceCPU:    resource.MustParse("1"),
												corev1.ResourceMemory: resource.MustParse("2Gi"),
											},
											Requests: corev1.ResourceList{
												corev1.ResourceCPU:    resource.MustParse("500m"),
												corev1.ResourceMemory: resource.MustParse("2Gi"),
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}
}

// isDRAInstalled checks whether DRA is installed on the cluster by verifying
// that the gpu.nvidia.com DeviceClass exists in any supported API version.
func isDRAInstalled(ctx context.Context, c client.Client) bool {
	probes := []client.Object{
		&resourcev1.DeviceClass{},
		&resourcev1beta2.DeviceClass{},
		&resourcev1beta1.DeviceClass{},
	}
	for _, probe := range probes {
		if err := c.Get(ctx, types.NamespacedName{Name: "gpu.nvidia.com"}, probe); err == nil {
			return true
		}
	}
	return false
}

// createOrSkipIfUnsupported creates the given object or skips the test
// if the API version is not served by the cluster.
func createOrSkipIfUnsupported(ctx context.Context, c client.Client, obj client.Object, description string) {
	err := c.Create(ctx, obj)
	if apimeta.IsNoMatchError(err) {
		ginkgo.Skip(fmt.Sprintf("API not available on cluster for %s: %v", description, err))
	}
	Expect(err).ToNot(HaveOccurred(), "failed to create %s", description)
}

// BuildResourceClaimTemplate creates a DRA ResourceClaimTemplate with the specified GPU count.
func BuildResourceClaimTemplate(name, namespace string, gpuCount int64) *resourcev1beta2.ResourceClaimTemplate {
	return &resourcev1beta2.ResourceClaimTemplate{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: resourcev1beta2.ResourceClaimTemplateSpec{
			Spec: resourcev1beta2.ResourceClaimSpec{
				Devices: resourcev1beta2.DeviceClaim{
					Requests: []resourcev1beta2.DeviceRequest{
						{
							Name: "gpu",
							Exactly: &resourcev1beta2.ExactDeviceRequest{
								DeviceClassName: "gpu.nvidia.com",
								AllocationMode:  resourcev1beta2.DeviceAllocationModeExactCount,
								Count:           gpuCount,
							},
						},
					},
				},
			},
		},
	}
}

// BuildResourceClaim creates a DRA ResourceClaim with the specified GPU count.
func BuildResourceClaim(name, namespace string, gpuCount int64) *resourcev1beta2.ResourceClaim {
	return &resourcev1beta2.ResourceClaim{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: resourcev1beta2.ResourceClaimSpec{
			Devices: resourcev1beta2.DeviceClaim{
				Requests: []resourcev1beta2.DeviceRequest{
					{
						Name: "gpu",
						Exactly: &resourcev1beta2.ExactDeviceRequest{
							DeviceClassName: "gpu.nvidia.com",
							AllocationMode:  resourcev1beta2.DeviceAllocationModeExactCount,
							Count:           gpuCount,
						},
					},
				},
			},
		},
	}
}

// BuildDeploymentWithDRATemplate creates a Deployment that references a DRA ResourceClaimTemplate.
func BuildDeploymentWithDRATemplate(name, namespace string, replicas int32, imageName, claimTemplateName string) *appsv1.Deployment {
	return &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: ptr.To(replicas),
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{"app": name},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{"app": name},
				},
				Spec: corev1.PodSpec{
					SchedulerName: "runai-scheduler",
					Containers: []corev1.Container{
						{
							Name:    "main",
							Image:   imageName,
							Command: []string{"bash", "-c", "sleep infinity"},
						},
					},
					ResourceClaims: []corev1.PodResourceClaim{
						{
							Name:                      "gpu-claim",
							ResourceClaimTemplateName: ptr.To(claimTemplateName),
						},
					},
				},
			},
		},
	}
}

// BuildDeploymentWithDRAClaim creates a Deployment that references a pre-existing DRA ResourceClaim.
func BuildDeploymentWithDRAClaim(name, namespace string, replicas int32, imageName, claimName string) *appsv1.Deployment {
	return &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: ptr.To(replicas),
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{"app": name},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{"app": name},
				},
				Spec: corev1.PodSpec{
					SchedulerName: "runai-scheduler",
					Containers: []corev1.Container{
						{
							Name:    "main",
							Image:   imageName,
							Command: []string{"bash", "-c", "sleep infinity"},
						},
					},
					ResourceClaims: []corev1.PodResourceClaim{
						{
							Name:              "gpu-claim",
							ResourceClaimName: ptr.To(claimName),
						},
					},
				},
			},
		},
	}
}

// DRA v1 builders

// BuildResourceClaimTemplateV1 creates a DRA v1 ResourceClaimTemplate with the specified GPU count.
func BuildResourceClaimTemplateV1(name, namespace string, gpuCount int64) *resourcev1.ResourceClaimTemplate {
	return &resourcev1.ResourceClaimTemplate{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: resourcev1.ResourceClaimTemplateSpec{
			Spec: resourcev1.ResourceClaimSpec{
				Devices: resourcev1.DeviceClaim{
					Requests: []resourcev1.DeviceRequest{
						{
							Name: "gpu",
							Exactly: &resourcev1.ExactDeviceRequest{
								DeviceClassName: "gpu.nvidia.com",
								AllocationMode:  resourcev1.DeviceAllocationModeExactCount,
								Count:           gpuCount,
							},
						},
					},
				},
			},
		},
	}
}

// BuildResourceClaimV1 creates a DRA v1 ResourceClaim with the specified GPU count.
func BuildResourceClaimV1(name, namespace string, gpuCount int64) *resourcev1.ResourceClaim {
	return &resourcev1.ResourceClaim{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: resourcev1.ResourceClaimSpec{
			Devices: resourcev1.DeviceClaim{
				Requests: []resourcev1.DeviceRequest{
					{
						Name: "gpu",
						Exactly: &resourcev1.ExactDeviceRequest{
							DeviceClassName: "gpu.nvidia.com",
							AllocationMode:  resourcev1.DeviceAllocationModeExactCount,
							Count:           gpuCount,
						},
					},
				},
			},
		},
	}
}

// DRA v1beta1 builders

// BuildResourceClaimTemplateV1Beta1 creates a DRA v1beta1 ResourceClaimTemplate with the specified GPU count.
func BuildResourceClaimTemplateV1Beta1(name, namespace string, gpuCount int64) *resourcev1beta1.ResourceClaimTemplate {
	return &resourcev1beta1.ResourceClaimTemplate{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: resourcev1beta1.ResourceClaimTemplateSpec{
			Spec: resourcev1beta1.ResourceClaimSpec{
				Devices: resourcev1beta1.DeviceClaim{
					Requests: []resourcev1beta1.DeviceRequest{
						{
							Name:            "gpu",
							DeviceClassName: "gpu.nvidia.com",
							AllocationMode:  resourcev1beta1.DeviceAllocationModeExactCount,
							Count:           gpuCount,
						},
					},
				},
			},
		},
	}
}

// BuildResourceClaimV1Beta1 creates a DRA v1beta1 ResourceClaim with the specified GPU count.
func BuildResourceClaimV1Beta1(name, namespace string, gpuCount int64) *resourcev1beta1.ResourceClaim {
	return &resourcev1beta1.ResourceClaim{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: resourcev1beta1.ResourceClaimSpec{
			Devices: resourcev1beta1.DeviceClaim{
				Requests: []resourcev1beta1.DeviceRequest{
					{
						Name:            "gpu",
						DeviceClassName: "gpu.nvidia.com",
						AllocationMode:  resourcev1beta1.DeviceAllocationModeExactCount,
						Count:           gpuCount,
					},
				},
			},
		},
	}
}
