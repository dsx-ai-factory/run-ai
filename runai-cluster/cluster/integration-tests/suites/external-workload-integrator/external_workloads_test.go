package external_workload_integrator

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
	"k8s.io/apimachinery/pkg/types"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	kubeflowv2beta1 "github.com/kubeflow/mpi-operator/pkg/apis/kubeflow/v2beta1"
	kubeflowv1 "github.com/kubeflow/training-operator/pkg/apis/kubeflow.org/v1"
	rayv1alpha1 "github.com/ray-project/kuberay/ray-operator/apis/ray/v1alpha1"
	workflow "github.com/run-ai/runai/runai-cluster/cluster/external-workload-integrator/pkg/thirdpartyapi/argo/v1alpha1"
	nvcfv1 "github.com/run-ai/runai/runai-cluster/cluster/external-workload-integrator/pkg/thirdpartyapi/nvcf/v1"
	"github.com/run-ai/runai/runai-cluster/cluster/integration-tests/common"
	"github.com/run-ai/runai/runai-cluster/cluster/integration-tests/config"
	"github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var _ = Describe("External Workloads Test", Ordered, func() {
	var (
		testConfig *config.TestConfig
		kubeClient client.Client
		ctx        context.Context

		testCreatedObjects = map[string]client.Object{}
		imageName          = "ubuntu"
	)

	BeforeAll(func() {
		testConfig = config.GetTestConfig(Default)
		kubeClient = testConfig.KubeClient
		ctx = testConfig.Ctx
	})

	AfterAll(func() {
		testConfig.Cleanup()
	})

	AfterEach(func() {
		for _, testCreatedObject := range testCreatedObjects {
			err := kubeClient.Delete(ctx, testCreatedObject)
			if err != nil && !errors.IsNotFound(err) {
				Expect(err).ToNot(HaveOccurred(), "Failed to delete object %s/%s/%s: %v",
					testCreatedObject.GetObjectKind().GroupVersionKind().String(), testCreatedObject.GetNamespace(), testCreatedObject.GetName(), err)
			}
		}
		testCreatedObjects = map[string]client.Object{}
	})

	Context("Test external-workload crd created for supported integrations", func() {
		It("Argocd workflow", func() {
			testName := "Argocd workflow"
			log.Printf("Starting test: %s", testName)

			argoWorkflow := &workflow.Workflow{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "argo-workflow",
					Namespace: testConfig.TestNsName,
				},
				Spec: workflow.WorkflowSpec{
					Entrypoint: "main",
				},
				Status: workflow.WorkflowStatus{
					Phase: workflow.WorkflowPending,
				},
			}
			err := kubeClient.Create(ctx, argoWorkflow)
			Expect(err).ToNot(HaveOccurred(), "failed to create argo workflow")
			testCreatedObjects["argoworkflow1"] = argoWorkflow

			ewArgo := waitForExternalWorkloadToExist(kubeClient, argoWorkflow, testName, testCreatedObjects)

			expectedEwArgo := createExpectedEw(
				argoWorkflow.ObjectMeta, "Workflow", testConfig.TestProject, testConfig.TestDepartment, v1alpha1.RequestedPods{}, "",
			)
			validateEwSpec(ewArgo, expectedEwArgo, testName)
		})

		It("Kubeflow xgboost job", func() {
			testName := "Kubeflow xgboost job"
			log.Printf("Starting test: %s", testName)

			masterReplicas := int32(1)
			workerReplicas := int32(2)

			xgboostJob := &kubeflowv1.XGBoostJob{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "xgboost-job",
					Namespace: testConfig.TestNsName,
				},
				Spec: kubeflowv1.XGBoostJobSpec{
					XGBReplicaSpecs: map[kubeflowv1.ReplicaType]*kubeflowv1.ReplicaSpec{
						kubeflowv1.XGBoostJobReplicaTypeMaster: {
							Replicas: &masterReplicas,
							Template: corev1.PodTemplateSpec{
								Spec: common.CreatePodSpec("xgboost", "", "", "", "", imageName),
							},
						},
						kubeflowv1.XGBoostJobReplicaTypeWorker: {
							Replicas: &workerReplicas,
							Template: corev1.PodTemplateSpec{
								Spec: common.CreatePodSpec("xgboost", "", "", "", "", imageName),
							},
						},
					},
				},
			}

			err := kubeClient.Create(ctx, xgboostJob)
			Expect(err).ToNot(HaveOccurred(), "failed to create xgboost job")
			testCreatedObjects["xgboostjob1"] = xgboostJob

			ewXgboost := waitForExternalWorkloadToExist(kubeClient, xgboostJob, testName, testCreatedObjects)

			expectedEwXgboost := createExpectedEw(
				xgboostJob.ObjectMeta, "XGBoostJob", testConfig.TestProject, testConfig.TestDepartment,
				v1alpha1.RequestedPods{Replicas: ptr.To(masterReplicas + workerReplicas)}, imageName,
			)
			validateEwSpec(ewXgboost, expectedEwXgboost, testName)
		})

		It("Kubeflow TF Job", func() {
			testName := "Kubeflow TF job"
			log.Printf("Starting test: %s", testName)

			workerReplicas := int32(2)

			tfJob := &kubeflowv1.TFJob{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "tf-job",
					Namespace: testConfig.TestNsName,
				},
				Spec: kubeflowv1.TFJobSpec{
					TFReplicaSpecs: map[kubeflowv1.ReplicaType]*kubeflowv1.ReplicaSpec{
						"Worker": {
							Replicas:      &workerReplicas,
							RestartPolicy: kubeflowv1.RestartPolicy(kubeflowv1.TFJobDefaultRestartPolicy),
							Template: corev1.PodTemplateSpec{
								Spec: common.CreatePodSpec("tf", "", "", "", "", imageName),
							},
						},
					},
				},
			}

			err := kubeClient.Create(ctx, tfJob)
			Expect(err).ToNot(HaveOccurred(), "failed to create tf job")
			testCreatedObjects["tfjob1"] = tfJob

			ewTf := waitForExternalWorkloadToExist(kubeClient, tfJob, testName, testCreatedObjects)

			expectedEwTf := createExpectedEw(
				tfJob.ObjectMeta, "TFJob", testConfig.TestProject, testConfig.TestDepartment,
				v1alpha1.RequestedPods{Replicas: ptr.To(workerReplicas)}, imageName,
			)
			validateEwSpec(ewTf, expectedEwTf, testName)
		})

		It("Kubeflow PyTorch Job", func() {
			testName := "Kubeflow PyTorch job"
			log.Printf("Starting test: %s", testName)

			masterReplicas := int32(1)
			workerReplicas := int32(2)

			pytorchJob := &kubeflowv1.PyTorchJob{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "pytorch-job",
					Namespace: testConfig.TestNsName,
				},
				Spec: kubeflowv1.PyTorchJobSpec{
					PyTorchReplicaSpecs: map[kubeflowv1.ReplicaType]*kubeflowv1.ReplicaSpec{
						"Master": {
							Replicas:      &masterReplicas,
							RestartPolicy: kubeflowv1.RestartPolicy(kubeflowv1.TFJobDefaultRestartPolicy),
							Template: corev1.PodTemplateSpec{
								Spec: common.CreatePodSpec("pytorch", "", "", "", "", imageName),
							},
						},
						"Worker": {
							Replicas:      &workerReplicas,
							RestartPolicy: kubeflowv1.RestartPolicy(kubeflowv1.TFJobDefaultRestartPolicy),
							Template: corev1.PodTemplateSpec{
								Spec: common.CreatePodSpec("pytorch", "", "", "", "", imageName),
							},
						},
					},
				},
			}

			err := kubeClient.Create(ctx, pytorchJob)
			Expect(err).ToNot(HaveOccurred(), "failed to create pytorch job")
			testCreatedObjects["pytorchjob1"] = pytorchJob

			ewPytorch := waitForExternalWorkloadToExist(kubeClient, pytorchJob, testName, testCreatedObjects)

			expectedEwPytorch := createExpectedEw(
				pytorchJob.ObjectMeta, "PyTorchJob", testConfig.TestProject, testConfig.TestDepartment,
				v1alpha1.RequestedPods{Replicas: ptr.To(masterReplicas + workerReplicas)}, imageName,
			)
			validateEwSpec(ewPytorch, expectedEwPytorch, testName)
		})

		It("Kubeflow MPI Job", func() {
			testName := "Kubeflow MPI job"
			log.Printf("Starting test: %s", testName)

			masterReplicas := int32(1)
			workerReplicas := int32(2)
			cleanPodPolicyRunning := kubeflowv2beta1.CleanPodPolicyRunning

			mpiJob := &kubeflowv2beta1.MPIJob{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "mpi-job",
					Namespace: testConfig.TestNsName,
				},
				Spec: kubeflowv2beta1.MPIJobSpec{
					RunPolicy: kubeflowv2beta1.RunPolicy{CleanPodPolicy: &cleanPodPolicyRunning},
					MPIReplicaSpecs: map[kubeflowv2beta1.MPIReplicaType]*kubeflowv2beta1.ReplicaSpec{
						kubeflowv2beta1.MPIReplicaTypeLauncher: {
							Replicas: &masterReplicas,
							Template: corev1.PodTemplateSpec{
								Spec: common.CreatePodSpec("mpi", "", "", "", "", imageName),
							},
						},
						kubeflowv2beta1.MPIReplicaTypeWorker: {
							Replicas: &workerReplicas,
							Template: corev1.PodTemplateSpec{
								Spec: common.CreatePodSpec("mpi", "", "", "", "", imageName),
							},
						},
					},
				},
			}

			err := kubeClient.Create(ctx, mpiJob)
			Expect(err).ToNot(HaveOccurred(), "failed to create mpi job")
			testCreatedObjects["mpijob1"] = mpiJob

			ewMpi := waitForExternalWorkloadToExist(kubeClient, mpiJob, testName, testCreatedObjects)

			expectedEwMpi := createExpectedEw(
				mpiJob.ObjectMeta, "MPIJob", testConfig.TestProject, testConfig.TestDepartment,
				v1alpha1.RequestedPods{Replicas: ptr.To(masterReplicas + workerReplicas)}, imageName,
			)
			validateEwSpec(ewMpi, expectedEwMpi, testName)
		})

		It("Ray Job", func() {
			testName := "Ray job"
			log.Printf("Starting test: %s", testName)

			workerReplicas := int32(1)
			maxReplicas := int32(3)
			minReplicas := int32(1)

			rayJob := &rayv1alpha1.RayJob{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "ray-job",
					Namespace: testConfig.TestNsName,
				},
				Spec: rayv1alpha1.RayJobSpec{
					Entrypoint:               "main",
					ShutdownAfterJobFinishes: true,
					RayClusterSpec: &rayv1alpha1.RayClusterSpec{
						RayVersion: "2.3.0",
						HeadGroupSpec: rayv1alpha1.HeadGroupSpec{
							RayStartParams: map[string]string{
								"password": "YourPassword",
							},
							Template: corev1.PodTemplateSpec{
								Spec: common.CreatePodSpec("ray", "", "", "", "", imageName),
							},
						},
						WorkerGroupSpecs: []rayv1alpha1.WorkerGroupSpec{
							{
								Replicas:    &workerReplicas,
								MaxReplicas: &maxReplicas,
								MinReplicas: &minReplicas,
								RayStartParams: map[string]string{
									"password": "YourPassword",
								},
								Template: corev1.PodTemplateSpec{
									Spec: common.CreatePodSpec("ray", "", "", "", "", imageName),
								},
							},
						},
					},
				},
			}

			err := kubeClient.Create(ctx, rayJob)
			Expect(err).ToNot(HaveOccurred(), "failed to create ray job")
			testCreatedObjects["rayjob1"] = rayJob

			ewRay := waitForExternalWorkloadToExist(kubeClient, rayJob, testName, testCreatedObjects)

			expectedEwRay := createExpectedEw(
				rayJob.ObjectMeta, "RayJob", testConfig.TestProject, testConfig.TestDepartment,
				v1alpha1.RequestedPods{Replicas: ptr.To(workerReplicas + 1), MinReplicas: ptr.To(minReplicas), MaxReplicas: ptr.To(maxReplicas)}, "ubuntu",
			)
			validateEwSpec(ewRay, expectedEwRay, testName)
		})
		It("NVCF Function", func() {
			testName := "NVCF Function"
			log.Printf("Starting test: %s", testName)

			// create a spotRequest instance
			uid := uuid.NewString()
			spotRequest := buildNVCF("nvcf-func", testConfig.TestNsName, uid)
			err := kubeClient.Create(ctx, spotRequest)
			Expect(err).ToNot(HaveOccurred(), "failed to create spotRequest")
			testCreatedObjects["spotRequest1"] = spotRequest

			nvcfPod := buildNVCFPod(spotRequest, imageName)
			err = kubeClient.Create(ctx, nvcfPod)
			Expect(err).ToNot(HaveOccurred(), "failed to create nvcf pod")
			testCreatedObjects["nvcfPod1"] = nvcfPod

			// Expect ew to be created
			// change the spotRequest uid, name in local run only, since NVCF ew is calculated differently
			clonedSpotRequest := spotRequest.DeepCopy()
			clonedSpotRequest.UID = types.UID(uid)
			clonedSpotRequest.Name = fmt.Sprintf("fv-%s", uid)
			ewSpotRequest := waitForExternalWorkloadToExist(kubeClient, clonedSpotRequest, testName, testCreatedObjects)

			// Validate ew
			expectedEwSpotRequest := createExpectedEw(
				clonedSpotRequest.ObjectMeta, "NVCF Function", testConfig.TestProject, testConfig.TestDepartment,
				v1alpha1.RequestedPods{Replicas: ptr.To(int32(1))}, imageName,
			)
			validateEwSpec(ewSpotRequest, expectedEwSpotRequest, testName)

			err = kubeClient.Delete(ctx, nvcfPod)
			Expect(err).ToNot(HaveOccurred(), "failed to delete pod")

			err = kubeClient.Delete(ctx, spotRequest)
			Expect(err).ToNot(HaveOccurred(), "failed to delete spotRequest")

			waitForExternalWorkloadToBeDeleted(kubeClient, clonedSpotRequest, testName)
		})
		It("NVCF Function - multiple spot requests", func() {
			testName := "NVCF Function - multiple spot requests"
			log.Printf("Starting test: %s", testName)

			functionUid := uuid.NewString()
			spotRequests := []*nvcfv1.SPOTRequest{}
			pods := []*corev1.Pod{}

			for i := 0; i < 3; i++ {
				// create a spotRequest instance
				spotRequest := buildNVCF(fmt.Sprintf("nvcf-func-%d", i), testConfig.TestNsName, functionUid)
				err := kubeClient.Create(ctx, spotRequest)
				Expect(err).ToNot(HaveOccurred(), "failed to create spotRequest")
				testCreatedObjects[fmt.Sprintf("spot-request-%d", i)] = spotRequest

				nvcfPod := buildNVCFPod(spotRequest, imageName)
				err = kubeClient.Create(ctx, nvcfPod)
				Expect(err).ToNot(HaveOccurred(), "failed to create nvcf pod")
				testCreatedObjects[fmt.Sprintf("nvcf-pod-%d", i)] = nvcfPod

				spotRequests = append(spotRequests, spotRequest)
				pods = append(pods, nvcfPod)
			}

			// Expect ew to be created
			// change the spotRequest uid, name in local run only, since NVCF ew is calculated differently
			clonedSpotRequest := spotRequests[0].DeepCopy()
			clonedSpotRequest.UID = types.UID(functionUid)
			clonedSpotRequest.Name = fmt.Sprintf("fv-%s", functionUid)

			// Validate ew
			expectedEwSpotRequest := createExpectedEw(
				clonedSpotRequest.ObjectMeta, "NVCF Function", testConfig.TestProject, testConfig.TestDepartment,
				v1alpha1.RequestedPods{Replicas: ptr.To(int32(3))}, imageName,
			)
			time.Sleep(3 * time.Second)
			ewSpotRequest := waitForExternalWorkloadToExist(kubeClient, clonedSpotRequest, testName, testCreatedObjects)
			validateEwSpec(ewSpotRequest, expectedEwSpotRequest, testName)

			for _, pod := range pods {
				err := kubeClient.Delete(ctx, pod)
				Expect(err).ToNot(HaveOccurred(), "failed to delete pod")
			}

			for _, spotRequest := range spotRequests {
				err := kubeClient.Delete(ctx, spotRequest)
				Expect(err).ToNot(HaveOccurred(), "failed to delete spotRequest")
			}

			waitForExternalWorkloadToBeDeleted(kubeClient, clonedSpotRequest, testName)
		})
	})

	Context("Ray CRDs + EW integration: Priority Class, Resource Reflection, Ingress, and Status", func() {
		It("should validate Ray Cluster EW integration (priority class, resources, ingress, status)", func() {
			testName := "Ray cluster ew integration"
			log.Printf("Starting test: %s", testName)
			priorityClassName := "build"
			rayCluster := BuildRayCluster("raycluster-standard", testConfig.TestNsName, testConfig.TestProject.Name, 2, 2, 2, "rayproject/ray:2.4.0", priorityClassName)

			err := kubeClient.Create(ctx, rayCluster)
			Expect(err).ToNot(HaveOccurred(), "failed to create ray cluster")
			testCreatedObjects["raycluster-validation"] = rayCluster

			rayClusterEw := waitForExternalWorkloadToExist(kubeClient, rayCluster, testName, testCreatedObjects)

			expectedSpec := ExpectedExternalWorkloadSpec{
				Replicas:          ptr.To(int32(3)), // 2 workers + 1 head
				MinReplicas:       ptr.To(int32(2)),
				MaxReplicas:       ptr.To(int32(2)),
				PriorityClassName: &priorityClassName,
			}
			ValidateEWIntegration(rayClusterEw, expectedSpec)
		})

		It("should validate Ray Job EW integration (priority class, resources, ingress, status)", func() {
			testName := "Ray job ew integration"
			log.Printf("Starting test: %s", testName)
			priorityClassName := "train"
			rayJob := BuildRayJob("rayjob-sample", testConfig.TestNsName, testConfig.TestProject.Name, 2, 2, 2, "rayproject/ray:2.4.0", priorityClassName)

			err := kubeClient.Create(ctx, rayJob)
			Expect(err).ToNot(HaveOccurred(), "failed to create ray job")
			testCreatedObjects["rayjob-validation"] = rayJob

			rayJobEw := waitForExternalWorkloadToExist(kubeClient, rayJob, testName, testCreatedObjects)

			expectedSpec := ExpectedExternalWorkloadSpec{
				Replicas:          ptr.To(int32(3)), // 2 workers + 1 head
				MinReplicas:       ptr.To(int32(2)),
				MaxReplicas:       ptr.To(int32(2)),
				PriorityClassName: &priorityClassName,
			}
			ValidateEWIntegration(rayJobEw, expectedSpec)
		})

		It("should validate Ray Serve EW integration (priority class, resources, ingress, status)", func() {
			testName := "Ray serve ew integration"
			log.Printf("Starting test: %s", testName)
			priorityClassName := "inference"
			rayService := BuildRayService("rayservice-sample", testConfig.TestNsName, testConfig.TestProject.Name, 1, 1, 5, "rayproject/ray:2.6.3", priorityClassName)

			err := kubeClient.Create(ctx, rayService)
			Expect(err).ToNot(HaveOccurred(), "failed to create ray service")
			testCreatedObjects["rayservice-validation"] = rayService

			rayServiceEw := waitForExternalWorkloadToExist(kubeClient, rayService, testName, testCreatedObjects)

			expectedSpec := ExpectedExternalWorkloadSpec{
				Replicas:          ptr.To(int32(2)), // 1 worker + 1 head
				MinReplicas:       ptr.To(int32(1)),
				MaxReplicas:       ptr.To(int32(5)),
				PriorityClassName: &priorityClassName,
			}
			ValidateEWIntegration(rayServiceEw, expectedSpec)
		})
	})

	Context("DRA (Dynamic Resource Allocation) GPU resource resolution", func() {
		BeforeAll(func() {
			if !isDRAInstalled(ctx, kubeClient) {
				Skip("DRA is not installed on the cluster (gpu.nvidia.com DeviceClass not found)")
			}
		})

		It("should resolve GPU resources from DRA ResourceClaimTemplate for a Deployment", func() {
			testName := "Deployment with DRA ResourceClaimTemplate"
			log.Printf("Starting test: %s", testName)

			draGpuCount := int64(4)
			replicas := int32(2)

			// Create ResourceClaimTemplate with 4 GPUs
			claimTemplate := BuildResourceClaimTemplate("dra-gpu-template", testConfig.TestNsName, draGpuCount)
			createOrSkipIfUnsupported(ctx, kubeClient, claimTemplate, "v1beta2 ResourceClaimTemplate")
			testCreatedObjects["dra-gpu-template"] = claimTemplate

			// Create Deployment referencing the ResourceClaimTemplate
			deployment := BuildDeploymentWithDRATemplate("depl-dra-template", testConfig.TestNsName, replicas, imageName, claimTemplate.Name)
			err := kubeClient.Create(ctx, deployment)
			Expect(err).ToNot(HaveOccurred(), "failed to create deployment with DRA template")
			testCreatedObjects["depl-dra-template"] = deployment

			ewDeployment := waitForExternalWorkloadToExist(kubeClient, deployment, testName, testCreatedObjects)

			// Expected: 2 replicas * 4 GPUs = 8 GPUs total
			expectedGpuQty := resource.NewQuantity(int64(replicas)*draGpuCount, resource.DecimalSI)
			expectedSpec := ExpectedExternalWorkloadSpec{
				Replicas: ptr.To(replicas),
				RequestedResources: &corev1.ResourceRequirements{
					Requests: corev1.ResourceList{
						"nvidia.com/gpu": *expectedGpuQty,
					},
					Limits: corev1.ResourceList{
						"nvidia.com/gpu": *expectedGpuQty,
					},
				},
			}
			ValidateEWIntegration(ewDeployment, expectedSpec)
		})

		It("should resolve GPU resources from DRA ResourceClaim for a Deployment", func() {
			testName := "Deployment with DRA ResourceClaim"
			log.Printf("Starting test: %s", testName)

			draGpuCount := int64(2)
			replicas := int32(3)

			// Create ResourceClaim with 2 GPUs
			claim := BuildResourceClaim("dra-gpu-claim", testConfig.TestNsName, draGpuCount)
			createOrSkipIfUnsupported(ctx, kubeClient, claim, "v1beta2 ResourceClaim")
			testCreatedObjects["dra-gpu-claim"] = claim

			// Create Deployment referencing the ResourceClaim
			deployment := BuildDeploymentWithDRAClaim("depl-dra-claim", testConfig.TestNsName, replicas, imageName, claim.Name)
			err := kubeClient.Create(ctx, deployment)
			Expect(err).ToNot(HaveOccurred(), "failed to create deployment with DRA claim")
			testCreatedObjects["depl-dra-claim"] = deployment

			ewDeployment := waitForExternalWorkloadToExist(kubeClient, deployment, testName, testCreatedObjects)

			// Expected: 3 replicas * 2 GPUs = 6 GPUs total
			expectedGpuQty := resource.NewQuantity(int64(replicas)*draGpuCount, resource.DecimalSI)
			expectedSpec := ExpectedExternalWorkloadSpec{
				Replicas: ptr.To(replicas),
				RequestedResources: &corev1.ResourceRequirements{
					Requests: corev1.ResourceList{
						"nvidia.com/gpu": *expectedGpuQty,
					},
					Limits: corev1.ResourceList{
						"nvidia.com/gpu": *expectedGpuQty,
					},
				},
			}
			ValidateEWIntegration(ewDeployment, expectedSpec)
		})

		It("should resolve GPU resources from DRA v1 ResourceClaimTemplate for a Deployment", func() {
			testName := "Deployment with DRA v1 ResourceClaimTemplate"
			log.Printf("Starting test: %s", testName)

			draGpuCount := int64(3)
			replicas := int32(2)

			// Create v1 ResourceClaimTemplate with 3 GPUs
			claimTemplate := BuildResourceClaimTemplateV1("dra-gpu-template-v1", testConfig.TestNsName, draGpuCount)
			createOrSkipIfUnsupported(ctx, kubeClient, claimTemplate, "v1 ResourceClaimTemplate")
			testCreatedObjects["dra-gpu-template-v1"] = claimTemplate

			// Create Deployment referencing the ResourceClaimTemplate
			deployment := BuildDeploymentWithDRATemplate("depl-dra-template-v1", testConfig.TestNsName, replicas, imageName, claimTemplate.Name)
			err := kubeClient.Create(ctx, deployment)
			Expect(err).ToNot(HaveOccurred(), "failed to create deployment with DRA v1 template")
			testCreatedObjects["depl-dra-template-v1"] = deployment

			ewDeployment := waitForExternalWorkloadToExist(kubeClient, deployment, testName, testCreatedObjects)

			// Expected: 2 replicas * 3 GPUs = 6 GPUs total
			expectedGpuQty := resource.NewQuantity(int64(replicas)*draGpuCount, resource.DecimalSI)
			expectedSpec := ExpectedExternalWorkloadSpec{
				Replicas: ptr.To(replicas),
				RequestedResources: &corev1.ResourceRequirements{
					Requests: corev1.ResourceList{
						"nvidia.com/gpu": *expectedGpuQty,
					},
					Limits: corev1.ResourceList{
						"nvidia.com/gpu": *expectedGpuQty,
					},
				},
			}
			ValidateEWIntegration(ewDeployment, expectedSpec)
		})

		It("should resolve GPU resources from DRA v1 ResourceClaim for a Deployment", func() {
			testName := "Deployment with DRA v1 ResourceClaim"
			log.Printf("Starting test: %s", testName)

			draGpuCount := int64(5)
			replicas := int32(2)

			// Create v1 ResourceClaim with 5 GPUs
			claim := BuildResourceClaimV1("dra-gpu-claim-v1", testConfig.TestNsName, draGpuCount)
			createOrSkipIfUnsupported(ctx, kubeClient, claim, "v1 ResourceClaim")
			testCreatedObjects["dra-gpu-claim-v1"] = claim

			// Create Deployment referencing the ResourceClaim
			deployment := BuildDeploymentWithDRAClaim("depl-dra-claim-v1", testConfig.TestNsName, replicas, imageName, claim.Name)
			err := kubeClient.Create(ctx, deployment)
			Expect(err).ToNot(HaveOccurred(), "failed to create deployment with DRA v1 claim")
			testCreatedObjects["depl-dra-claim-v1"] = deployment

			ewDeployment := waitForExternalWorkloadToExist(kubeClient, deployment, testName, testCreatedObjects)

			// Expected: 2 replicas * 5 GPUs = 10 GPUs total
			expectedGpuQty := resource.NewQuantity(int64(replicas)*draGpuCount, resource.DecimalSI)
			expectedSpec := ExpectedExternalWorkloadSpec{
				Replicas: ptr.To(replicas),
				RequestedResources: &corev1.ResourceRequirements{
					Requests: corev1.ResourceList{
						"nvidia.com/gpu": *expectedGpuQty,
					},
					Limits: corev1.ResourceList{
						"nvidia.com/gpu": *expectedGpuQty,
					},
				},
			}
			ValidateEWIntegration(ewDeployment, expectedSpec)
		})
	})
})
