/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package managers

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/constants"
	runaiJobV1 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-common/config"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-common/issue"
	workloadproject "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-common/project"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-controller/pkg/api/common"
	iw "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-controller/pkg/api/interactiveworkload/v2alpha1"
	job "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-controller/pkg/api/jobs"
	mpi "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-controller/pkg/api/mpijob/v2beta1"
	issue2 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-controller/pkg/builders/common/issue"
	resourceaggr "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-controller/pkg/builders/common/resource"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-controller/pkg/utils/events"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-controller/pkg/utils/testing_utils/client"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-controller/pkg/utils/testing_utils/register"
	resourceutils "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-controller/pkg/utils/testing_utils/resource"
	workloadutils "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-controller/pkg/utils/testing_utils/resource/workload"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-controller/pkg/utils/workloads/convertors"
	kube "github.com/kubeflow/training-operator/pkg/apis/kubeflow.org/v1"
	nvapi "github.com/NVIDIA/k8s-dra-driver-gpu/api/nvidia.com/resource/v1beta1"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	routev1 "github.com/openshift/api/route/v1"
	v1 "k8s.io/api/core/v1"
	network "k8s.io/api/networking/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	k8sclient "sigs.k8s.io/controller-runtime/pkg/client"
	gatewayv1 "sigs.k8s.io/gateway-api/apis/v1"
)

// Additional constants that should be defined
const (
	TestRunaiJobName      = "test-runai-job-1"
	TestPytorchJobName    = "test-pytorch-job-1"
	TestJaxJobName        = "test-jax-job-1"
	TestTFJobName         = "test-tf-job-1"
	TestXGBJobName        = "test-xgb-job-1"
	TestMPIJobName        = "test-mpi-job-1"
	TestServiceName1      = "test-service-1"
	TestServiceName2      = "test-service-2"
	TestIngressName1      = "test-ingress-1"
	TestIngressName2      = "test-ingress-2"
	TestIngressName3      = "test-ingress-3"
	TestPvc1              = "test-pvc-1"
	testRoute2            = "test-route-2"
	testRoute3            = "test-route-3"
	TestRoute1            = "test-route-1"
	workloadName          = "workload-name"
	projectNamespace      = "runai-test-project"
	projectName           = "test-project"
	pvcSize               = "10Gi"
	TestComputeDomainName = "test-compute-domain"
	TestJaxWorkloadName   = "test-jax-workload"
	TestSecretName        = "test-secret-1"
)

var (
	runaiJobResource = common.Resource{
		Kind: metav1.GroupVersionKind{
			Kind: constants.KindRunaiJob,
		},
		Name:      TestRunaiJobName,
		Namespace: projectNamespace,
	}
	pytorchJobResource = common.Resource{
		Kind: metav1.GroupVersionKind{
			Kind: constants.KindPyTorchJob,
		},
		Name:      TestPytorchJobName,
		Namespace: projectNamespace,
	}
	jaxJobResource = common.Resource{
		Kind: metav1.GroupVersionKind{
			Kind: constants.KindJAXJob,
		},
		Name:      TestJaxJobName,
		Namespace: projectNamespace,
	}
	tfJobResource = common.Resource{
		Kind: metav1.GroupVersionKind{
			Kind: constants.KindTFJob,
		},
		Name:      TestTFJobName,
		Namespace: projectNamespace,
	}
	mpiJobResource = common.Resource{
		Kind: metav1.GroupVersionKind{
			Kind: constants.KindMpiJob,
		},
		Name:      TestMPIJobName,
		Namespace: projectNamespace,
	}
)

var _ = Describe("Resource manager", func() {
	var (
		k8sClient           k8sclient.Client
		scheme              *runtime.Scheme
		K8sResourcesManager *K8sResourceManager
		ctx                 context.Context
		cancel              context.CancelFunc
		recorder            *record.FakeRecorder
	)

	BeforeEach(func() {
		scheme = runtime.NewScheme()
		register.InitSchema(scheme)
		ctx = workloadproject.SetProjectInContext(context.Background(), projectNamespace)
		ctx, cancel = context.WithCancel(ctx)

		recorder = record.NewFakeRecorder(200)
	})
	AfterEach(func() {
		defer cancel()
	})

	Context("Get All Resources ", func() {
		var expectedResources []common.Resource

		BeforeEach(func() {
			expectedResources = []common.Resource{
				{
					Kind: metav1.GroupVersionKind{
						Group:   "",
						Version: "",
						Kind:    constants.KindService,
					},
					Name:      TestServiceName1,
					Namespace: projectNamespace,
				},
				{
					Kind: metav1.GroupVersionKind{
						Group:   "",
						Version: "",
						Kind:    constants.KindService,
					},
					Name:      TestServiceName2,
					Namespace: projectNamespace,
				},
				{Kind: metav1.GroupVersionKind{
					Kind: constants.KindIngress,
				},
					Name:      TestIngressName1,
					Namespace: projectNamespace,
				},
				{Kind: metav1.GroupVersionKind{
					Kind: constants.KindIngress,
				},
					Name:      TestIngressName2,
					Namespace: projectNamespace,
				},
				{Kind: metav1.GroupVersionKind{
					Kind: constants.KindRoute,
				},
					Name:      TestRoute1,
					Namespace: projectNamespace,
				},
				{Kind: metav1.GroupVersionKind{
					Kind: constants.KindRoute,
				},
					Name:      testRoute2,
					Namespace: projectNamespace,
				},
				{
					Kind: metav1.GroupVersionKind{
						Kind: constants.KindPvc,
					},
					Name:      TestPvc1,
					Namespace: projectNamespace,
				},
				{
					Kind: metav1.GroupVersionKind{
						Kind: constants.KindComputeDomain,
					},
					Name:      TestComputeDomainName,
					Namespace: projectNamespace,
				},
				{
					Kind: metav1.GroupVersionKind{
						Kind: constants.KindSecret,
					},
					Name:      TestSecretName,
					Namespace: projectNamespace,
				},
			}
		})
		Context("No resources", func() {
			BeforeEach(func() {
				k8sClient = client.InitFakeClientWithObjectNoStatus(scheme)
				K8sResourcesManager = NewResourceManager(k8sClient, k8sClient, &issue2.ComputeResourceGrouper{})
			})
			It("Should return empty list", func() {
				aggr := resourceaggr.NewAggregator(nil)
				ra, err := K8sResourcesManager.GetAllResources(ctx, aggr, expectedResources)
				Expect(err).To(HaveOccurred())
				Expect(ra).To(BeNil())
			})
		})
		Context("resources exists", func() {
			var (
				workload     *job.JobWorkload
				ownerUid     types.UID
				kind         string
				serviceOne   *v1.Service
				serviceTwo   *v1.Service
				ingressOne   *network.Ingress
				ingressTwo   *network.Ingress
				ingressThree *network.Ingress
				routeOne     *routev1.Route
				routeTwo     *routev1.Route
				routeThree   *routev1.Route
				runaiJob     *runaiJobV1.RunaiJob
				pytorchJob   *kube.PyTorchJob
				jaxJob       *kube.JAXJob
				mpiJob       *mpi.MPIJob
				xgbJob       *kube.XGBoostJob
				tfJob        *kube.TFJob
				pvc          *v1.PersistentVolumeClaim
				cd           *nvapi.ComputeDomain
				secret       *v1.Secret
			)

			BeforeEach(func() {
				ownerUid = "test-uid"
				kind = constants.KindInteractiveWorkload
				workload = &job.JobWorkload{}
				workload.SetName(workloadName)
				workload.SetNamespace(projectNamespace)
				labels := map[string]string{constants.LabelRelease: workload.Name}
				serviceOne = resourceutils.CreateDumbService(TestServiceName1, projectNamespace, ownerUid, kind, labels)
				serviceTwo = resourceutils.CreateDumbService(TestServiceName2, projectNamespace, ownerUid, kind, labels)
				ingressOne = resourceutils.CreateDumbIngress(TestIngressName1, projectNamespace, ownerUid, kind, labels)
				ingressTwo = resourceutils.CreateDumbIngress(TestIngressName2, projectNamespace, ownerUid, kind, labels)
				ingressThree = resourceutils.CreateDumbIngress(TestIngressName3, projectNamespace, ownerUid, kind, labels)
				routeOne = resourceutils.CreateDumbRoute(TestRoute1, projectNamespace, ownerUid, kind, labels)
				routeTwo = resourceutils.CreateDumbRoute(testRoute2, projectNamespace, ownerUid, kind, labels)
				routeThree = resourceutils.CreateDumbRoute(testRoute3, projectNamespace, ownerUid, kind, labels)
				runaiJob = resourceutils.CreateDumbRunaiJob(TestRunaiJobName, projectNamespace, ownerUid, kind, runaiJobV1.JobSpec{})
				pytorchJob = resourceutils.CreateDumbPytorch(TestPytorchJobName, projectNamespace, ownerUid, kind, labels)
				jaxJob = resourceutils.CreateDumbJAXJob(TestJaxJobName, projectNamespace, ownerUid, kind, labels)
				tfJob = resourceutils.CreateDumbTensorflow(TestTFJobName, projectNamespace, ownerUid, kind, labels)
				xgbJob = resourceutils.CreateDumbXGBoost(TestXGBJobName, projectNamespace, ownerUid, kind, labels)
				mpiJob = resourceutils.CreateDumbMPI(TestMPIJobName, projectNamespace, ownerUid, kind, labels)
				pvc = resourceutils.CreatePvc(TestPvc1, projectNamespace, resource.MustParse(pvcSize),
					ownerUid, kind, "", nil, nil, labels)
				cd = resourceutils.CreateDumbComputeDomain(TestComputeDomainName, projectNamespace, ownerUid, kind, labels)
				secret = resourceutils.CreateSecret(TestSecretName, projectNamespace, ownerUid, kind, labels)
				k8sClient = client.InitFakeClientWithObjectNoStatus(scheme, serviceOne, serviceTwo,
					ingressOne, ingressTwo, routeOne, routeTwo, runaiJob, pytorchJob,
					ingressThree, routeThree, pvc, mpiJob, xgbJob, tfJob, jaxJob, cd, secret)

				K8sResourcesManager = NewResourceManager(k8sClient, k8sClient, &issue2.ComputeResourceGrouper{})
			})
			DescribeTable("Getting resources for different job types",
				func(jobResource common.Resource) {
					resources := append(expectedResources, jobResource)

					workload.SetUID(ownerUid)
					workload.SetNamespace(projectNamespace)
					workload.SetName(workloadName)
					aggr := resourceaggr.NewAggregator(workload) // Initialize with appropriate workload

					_, err := K8sResourcesManager.GetAllResources(ctx, aggr, resources)
					Expect(err).ToNot(HaveOccurred())
					Expect(aggr.GetJob()).ToNot(BeNil())
					Expect(aggr.GetJob().GetName()).To(Equal(jobResource.Name))

					// Common expectations for all job types
					Expect(aggr.GetServices()).To(HaveLen(2))
					Expect(aggr.GetExternalNetwork()).To(HaveLen(4))
					Expect(aggr.GetPvcs()).To(HaveLen(1))
					Expect(aggr.GetComputeDomain()).NotTo(BeNil())
					Expect(aggr.GetSecrets()).To(HaveLen(1))

					// Validate external network resources
					lookupMap := map[string]bool{
						TestIngressName1: false,
						TestIngressName2: false,
						TestRoute1:       false,
						testRoute2:       false,
					}

					for _, ingress := range aggr.GetExternalNetwork() {
						_, found := lookupMap[ingress.GetName()]
						if found {
							lookupMap[ingress.GetName()] = true
						}
						Expect(ingress.GetName()).ToNot(Equal(TestIngressName3))
						Expect(ingress.GetName()).ToNot(Equal(testRoute3))
					}

					// Verify all expected network resources were found
					for _, value := range lookupMap {
						Expect(value).To(Equal(true))
					}
				},
				Entry("RunAI Job",
					runaiJobResource,
				),
				Entry("PyTorch Job",
					pytorchJobResource,
				),
				Entry("JAX Job",
					jaxJobResource,
				),
				Entry("TensorFlow Job",
					tfJobResource,
				),
				Entry("MPI Job",
					mpiJobResource,
				),
			)

		})
		Context("mode transition cleanup - orphaned networking resources detected as dangling", func() {
			var (
				workload     *job.JobWorkload
				ownerUid     types.UID
				kind         string
				serviceOne   *v1.Service
				ingressOne   *network.Ingress
				httpRouteOne *gatewayv1.HTTPRoute
			)

			const (
				testHTTPRouteName = "test-httproute-1"
			)

			BeforeEach(func() {
				ownerUid = "test-uid"
				kind = constants.KindInteractiveWorkload
				workload = &job.JobWorkload{}
				workload.SetName(workloadName)
				workload.SetNamespace(projectNamespace)
				labels := map[string]string{constants.LabelRelease: workload.Name}
				serviceOne = resourceutils.CreateDumbService(TestServiceName1, projectNamespace, ownerUid, kind, labels)
				ingressOne = resourceutils.CreateDumbIngress(TestIngressName1, projectNamespace, ownerUid, kind, labels)
				httpRouteOne = resourceutils.CreateDumbHTTPRoute(testHTTPRouteName, projectNamespace, ownerUid, kind, labels)
			})

			AfterEach(func() {
				_ = os.Unsetenv(config.EnvVarIngressEnabled)
				_ = os.Unsetenv(config.EnvVarGatewayAPIEnabled)
			})

			It("should detect orphaned Ingresses when switching to Gateway-only mode", func() {
				k8sClient = client.InitFakeClientWithObjectNoStatus(scheme, serviceOne, ingressOne)
				K8sResourcesManager = NewResourceManager(k8sClient, k8sClient, &issue2.ComputeResourceGrouper{})

				_ = os.Setenv(config.EnvVarIngressEnabled, "false")
				_ = os.Setenv(config.EnvVarGatewayAPIEnabled, "true")

				kindList := []common.Resource{
					{Kind: metav1.GroupVersionKind{Kind: constants.KindService}, Name: TestServiceName1, Namespace: projectNamespace},
					{Kind: metav1.GroupVersionKind{Kind: constants.KindHTTPRoute}, Name: testHTTPRouteName, Namespace: projectNamespace},
				}

				workload.SetUID(ownerUid)
				aggr := resourceaggr.NewAggregator(workload)
				dangling, err := K8sResourcesManager.GetAllResources(ctx, aggr, kindList)
				Expect(err).ToNot(HaveOccurred())
				Expect(dangling).To(HaveLen(1))
				Expect(dangling[0].GetName()).To(Equal(TestIngressName1))
			})

			It("should detect orphaned HTTPRoutes when switching to Ingress-only mode", func() {
				k8sClient = client.InitFakeClientWithObjectNoStatus(scheme, serviceOne, httpRouteOne)
				K8sResourcesManager = NewResourceManager(k8sClient, k8sClient, &issue2.ComputeResourceGrouper{})

				_ = os.Setenv(config.EnvVarIngressEnabled, "true")
				_ = os.Setenv(config.EnvVarGatewayAPIEnabled, "false")

				kindList := []common.Resource{
					{Kind: metav1.GroupVersionKind{Kind: constants.KindService}, Name: TestServiceName1, Namespace: projectNamespace},
					{Kind: metav1.GroupVersionKind{Kind: constants.KindIngress}, Name: TestIngressName1, Namespace: projectNamespace},
				}

				workload.SetUID(ownerUid)
				aggr := resourceaggr.NewAggregator(workload)
				dangling, err := K8sResourcesManager.GetAllResources(ctx, aggr, kindList)
				Expect(err).ToNot(HaveOccurred())
				Expect(dangling).To(HaveLen(1))
				Expect(dangling[0].GetName()).To(Equal(testHTTPRouteName))
			})

			It("should not detect dangling resources in dual mode", func() {
				k8sClient = client.InitFakeClientWithObjectNoStatus(scheme, serviceOne, ingressOne, httpRouteOne)
				K8sResourcesManager = NewResourceManager(k8sClient, k8sClient, &issue2.ComputeResourceGrouper{})

				_ = os.Setenv(config.EnvVarIngressEnabled, "true")
				_ = os.Setenv(config.EnvVarGatewayAPIEnabled, "true")

				kindList := []common.Resource{
					{Kind: metav1.GroupVersionKind{Kind: constants.KindService}, Name: TestServiceName1, Namespace: projectNamespace},
					{Kind: metav1.GroupVersionKind{Kind: constants.KindIngress}, Name: TestIngressName1, Namespace: projectNamespace},
					{Kind: metav1.GroupVersionKind{Kind: constants.KindHTTPRoute}, Name: testHTTPRouteName, Namespace: projectNamespace},
				}

				workload.SetUID(ownerUid)
				aggr := resourceaggr.NewAggregator(workload)
				dangling, err := K8sResourcesManager.GetAllResources(ctx, aggr, kindList)
				Expect(err).ToNot(HaveOccurred())
				Expect(dangling).To(BeEmpty())
			})
		})
	})
	Context("UpsertResources", func() {
		var (
			workload *job.JobWorkload
		)
		ownerUid := types.UID("test-uid")
		kind := constants.KindInteractiveWorkload
		Context("No resources create it", func() {

			var (
				objects    []k8sclient.Object
				serviceOne *v1.Service
				ingressOne *network.Ingress
				runaiJob   *runaiJobV1.RunaiJob
				pvc        *v1.PersistentVolumeClaim
				secret     *v1.Secret
				cd         *nvapi.ComputeDomain
			)
			BeforeEach(func() {
				k8sClient = client.InitFakeClientWithObjectNoStatus(scheme)
				K8sResourcesManager = NewResourceManager(k8sClient, k8sClient, &issue2.ComputeResourceGrouper{})
				workload = &job.JobWorkload{}
				workload.Kind = kind
				workload.SetName(workloadName)
				workload.SetUID(ownerUid)
				labels := map[string]string{constants.LabelRelease: workload.Name}
				objects = []k8sclient.Object{}
				serviceOne = resourceutils.CreateDumbService(TestServiceName1, projectNamespace, ownerUid, kind, labels)
				objects = append(objects, serviceOne)
				ingressOne = resourceutils.CreateDumbIngress(TestIngressName1, projectNamespace, ownerUid, kind, labels)
				objects = append(objects, ingressOne)
				runaiJob = resourceutils.CreateDumbRunaiJob(TestRunaiJobName, projectNamespace, ownerUid, kind, runaiJobV1.JobSpec{})
				objects = append(objects, runaiJob)
				pvc = resourceutils.CreatePvc(TestPvc1, projectNamespace, resource.MustParse(pvcSize), ownerUid, kind, "", nil, nil, labels)
				objects = append(objects, pvc)
				cd = resourceutils.CreateDumbComputeDomain(TestComputeDomainName, projectNamespace, ownerUid, kind, labels)
				objects = append(objects, cd)
				secret = resourceutils.CreateSecret(TestSecretName, projectNamespace, ownerUid, kind, labels)
				objects = append(objects, secret)

				ctx = injectWorkloadEventRecorder(ctx, workload, recorder)
			})
			It("Should create all resources", func() {
				issues := K8sResourcesManager.UpsertResources(ctx, workload, objects, ResourceOptions{})
				Expect(issues).To(BeNil())
				newRunaiJob := &runaiJobV1.RunaiJob{}
				err := k8sClient.Get(ctx, k8sclient.ObjectKeyFromObject(runaiJob), newRunaiJob)
				Expect(err).ToNot(HaveOccurred())
				Expect(newRunaiJob.GetName()).To(Equal(runaiJob.GetName()))
				service := &v1.Service{}
				err = k8sClient.Get(ctx, k8sclient.ObjectKeyFromObject(serviceOne), service)
				Expect(err).ToNot(HaveOccurred())
				Expect(service.GetName()).To(Equal(serviceOne.GetName()))
				ingress := &network.Ingress{}
				err = k8sClient.Get(ctx, k8sclient.ObjectKeyFromObject(ingressOne), ingress)
				Expect(err).ToNot(HaveOccurred())
				Expect(ingress.GetName()).To(Equal(ingressOne.GetName()))

				getPvc := &v1.PersistentVolumeClaim{}
				err = k8sClient.Get(ctx, k8sclient.ObjectKeyFromObject(pvc), getPvc)
				Expect(err).ToNot(HaveOccurred())
				Expect(getPvc.GetName()).To(Equal(pvc.GetName()))

				getSecret := &v1.Secret{}
				err = k8sClient.Get(ctx, k8sclient.ObjectKeyFromObject(secret), getSecret)
				Expect(err).ToNot(HaveOccurred())
				Expect(getSecret.GetName()).To(Equal(secret.GetName()))

				expectedCd := &nvapi.ComputeDomain{}
				err = k8sClient.Get(ctx, k8sclient.ObjectKeyFromObject(cd), expectedCd)
				Expect(err).ToNot(HaveOccurred())
				Expect(expectedCd.GetName()).To(Equal(cd.GetName()))
				Eventually(recorder.Events).WithTimeout(3 * time.Second).
					Should(Receive(ContainSubstring(fmt.Sprintf("Successfully created %s %s", constants.KindComputeDomain, cd.Name))))
			})
			It("Ready condition should be false", func() {
				issues := K8sResourcesManager.UpsertResources(ctx, workload, objects, ResourceOptions{})
				Expect(issues).To(BeNil())
				Expect(workload.Status.Conditions).To(HaveLen(1))
				Expect(string(workload.Status.Conditions[0].Type)).To(Equal(common.Ready))
				Expect(workload.Status.Conditions[0].Status).To(Equal(v1.ConditionFalse))
			})
		})
		Context("Resources exists update it", func() {

			var (
				objects    []k8sclient.Object
				serviceOne *v1.Service
				ingressOne *network.Ingress
				runaiJob   *runaiJobV1.RunaiJob
				cd         *nvapi.ComputeDomain
			)
			BeforeEach(func() {
				workload = &job.JobWorkload{}
				workload.Kind = kind
				workload.SetName(workloadName)
				workload.SetUID(ownerUid)
				objects = []k8sclient.Object{}
				labels := map[string]string{constants.LabelRelease: workload.Name}
				serviceOne = resourceutils.CreateDumbService(TestServiceName1, projectNamespace, ownerUid, kind, labels)
				objects = append(objects, serviceOne)
				ingressOne = resourceutils.CreateDumbIngress(TestIngressName1, projectNamespace, ownerUid, kind, labels)
				objects = append(objects, ingressOne)
				runaiJob = resourceutils.CreateDumbRunaiJob(TestRunaiJobName, projectNamespace, ownerUid, kind, runaiJobV1.JobSpec{})
				objects = append(objects, runaiJob)
				cd = resourceutils.CreateDumbComputeDomain(TestComputeDomainName, projectNamespace, ownerUid, kind, labels)
				objects = append(objects, cd)

				k8sClient = client.InitFakeClientWithObjectNoStatus(scheme, runaiJob, serviceOne, ingressOne, cd)
				K8sResourcesManager = NewResourceManager(k8sClient, k8sClient, &issue2.ComputeResourceGrouper{})

				ctx = injectWorkloadEventRecorder(ctx, workload, recorder)
			})
			It("Should copy finalizer from the existing resource before upsert", func() {
				newRunaiJob := &runaiJobV1.RunaiJob{}
				err := k8sClient.Get(ctx, k8sclient.ObjectKeyFromObject(runaiJob), newRunaiJob)
				Expect(err).ToNot(HaveOccurred())
				newRunaiJob.SetFinalizers([]string{"test-finalizer"})
				err = k8sClient.Update(ctx, newRunaiJob)
				Expect(err).ToNot(HaveOccurred())
				errs := K8sResourcesManager.UpsertResources(ctx, workload, objects, ResourceOptions{})
				Expect(errs).To(BeNil())
				err = k8sClient.Get(ctx, k8sclient.ObjectKeyFromObject(runaiJob), newRunaiJob)
				Expect(err).ToNot(HaveOccurred())
				Expect(newRunaiJob.GetFinalizers()).To(ContainElement("test-finalizer"))
			})
			It("Should copy runtime status labels before upsert", func() {
				newRunaiJob := &runaiJobV1.RunaiJob{}
				err := k8sClient.Get(ctx, k8sclient.ObjectKeyFromObject(runaiJob), newRunaiJob)
				Expect(err).ToNot(HaveOccurred())
				newRunaiJob.Labels = map[string]string{
					"test": "test",
				}
				err = k8sClient.Update(ctx, newRunaiJob)
				Expect(err).ToNot(HaveOccurred())
				errs := K8sResourcesManager.UpsertResources(ctx, workload, objects, ResourceOptions{})
				Expect(errs).To(BeNil())
				err = k8sClient.Get(ctx, k8sclient.ObjectKeyFromObject(runaiJob), newRunaiJob)
				Expect(err).ToNot(HaveOccurred())
				Expect(newRunaiJob.GetLabels()).To(HaveKeyWithValue("test", "test"))
			})
			It("Should copy runtime status annotation before upsert", func() {
				newRunaiJob := &runaiJobV1.RunaiJob{}
				err := k8sClient.Get(ctx, k8sclient.ObjectKeyFromObject(runaiJob), newRunaiJob)
				Expect(err).ToNot(HaveOccurred())
				newRunaiJob.Annotations = map[string]string{
					common.RunaiCurrentAllocatedGpusAnnotation:      "2",
					common.RunaiCurrentAllocatedGpuMemoryAnnotation: "2400",
					"test": "test",
				}
				err = k8sClient.Update(ctx, newRunaiJob)
				Expect(err).ToNot(HaveOccurred())
				errs := K8sResourcesManager.UpsertResources(ctx, workload, objects, ResourceOptions{})
				Expect(errs).To(BeNil())
				err = k8sClient.Get(ctx, k8sclient.ObjectKeyFromObject(runaiJob), newRunaiJob)
				Expect(err).ToNot(HaveOccurred())
				Expect(newRunaiJob.GetAnnotations()).To(HaveLen(3))
				Expect(newRunaiJob.GetAnnotations()).To(HaveKeyWithValue(common.RunaiCurrentAllocatedGpuMemoryAnnotation, "2400"))
				Expect(newRunaiJob.GetAnnotations()).To(HaveKeyWithValue(common.RunaiCurrentAllocatedGpusAnnotation, "2"))
				Expect(newRunaiJob.GetAnnotations()).To(HaveKeyWithValue("test", "test"))

			})
			It("Should upsert all resources", func() {
				errs := K8sResourcesManager.UpsertResources(ctx, workload, objects, ResourceOptions{})
				Expect(errs).To(BeNil())
				newRunaiJob := &runaiJobV1.RunaiJob{}
				err := k8sClient.Get(ctx, k8sclient.ObjectKeyFromObject(runaiJob), newRunaiJob)
				Expect(err).ToNot(HaveOccurred())
				Expect(newRunaiJob.GetName()).To(Equal(runaiJob.GetName()))
				service := &v1.Service{}
				err = k8sClient.Get(ctx, k8sclient.ObjectKeyFromObject(serviceOne), service)
				Expect(err).ToNot(HaveOccurred())
				Expect(service.GetName()).To(Equal(serviceOne.GetName()))
				ingress := &network.Ingress{}
				err = k8sClient.Get(ctx, k8sclient.ObjectKeyFromObject(ingressOne), ingress)
				Expect(err).ToNot(HaveOccurred())
				Expect(ingress.GetName()).To(Equal(ingressOne.GetName()))

				expectedCd := &nvapi.ComputeDomain{}
				err = k8sClient.Get(ctx, k8sclient.ObjectKeyFromObject(cd), expectedCd)
				Expect(err).ToNot(HaveOccurred())
				Expect(expectedCd.GetName()).To(Equal(cd.GetName()))
				// making sure we don't send event for update to not spam
				Consistently(recorder.Events).ShouldNot(Receive())
			})
			It("Ready condition should not exists, no changes to resources", func() {
				errs := K8sResourcesManager.UpsertResources(ctx, workload, objects, ResourceOptions{})
				Expect(errs).To(BeNil())
				Expect(workload.Status.Conditions).To(HaveLen(0))
			})
			It("should failed to update owner reference is not equal", func() {
				workload.SetUID("other-uid")
				issues := K8sResourcesManager.UpsertResources(ctx, workload, objects, ResourceOptions{})
				var msgList []string
				Expect(issues).ToNot(BeNil())
				Expect(len(issues)).To(Equal(len(objects)))

				for _, iss := range issues {
					msgList = append(msgList, iss.Message())
				}

				Expect(msgList).To(ContainElement("[Service:test-service-1]  \"test-service-1\" already exists"))
				Expect(msgList).To(ContainElement("[RunaiJob:test-runai-job-1]  \"test-runai-job-1\" already exists"))
				Expect(msgList).To(ContainElement("[Ingress:test-ingress-1]  \"test-ingress-1\" already exists"))
			})
		})
		Context("Backward compatibility, resources exits no owner ref", func() {
			var (
				objects    []k8sclient.Object
				serviceOne *v1.Service
				ingressOne *network.Ingress
				runaiJob   *runaiJobV1.RunaiJob
			)

			BeforeEach(func() {
				workload = workloadutils.NewJobWorkloadBuilder().WithUid(ownerUid).WithName("test-job").WithNamespace("test-namespace").Build()
				labels := map[string]string{constants.LabelRelease: workload.Name}
				objects = []k8sclient.Object{}
				serviceOne = resourceutils.CreateDumbService(TestServiceName1, projectNamespace, "", "", labels)
				serviceOne.SetLabels(map[string]string{
					constants.LabelRelease: workload.Name,
				})
				objects = append(objects, serviceOne)
				ingressOne = resourceutils.CreateDumbIngress(TestIngressName1, projectNamespace, "", "", labels)
				ingressOne.SetLabels(map[string]string{
					constants.LabelRelease: workload.Name,
				})
				objects = append(objects, ingressOne)
				runaiJob = resourceutils.CreateDumbRunaiJob(TestRunaiJobName, projectNamespace, "", "", runaiJobV1.JobSpec{})
				runaiJob.SetLabels(map[string]string{
					constants.LabelRelease: workload.Name,
				})
				objects = append(objects, runaiJob)
				k8sClient = client.InitFakeClientWithObjectNoStatus(scheme, runaiJob, serviceOne, ingressOne)
				K8sResourcesManager = NewResourceManager(k8sClient, k8sClient, &issue2.ComputeResourceGrouper{})
			})

			It("Should copy runtime status annotation before upsert", func() {
				newRunaiJob := &runaiJobV1.RunaiJob{}
				err := k8sClient.Get(ctx, k8sclient.ObjectKeyFromObject(runaiJob), newRunaiJob)
				Expect(err).ToNot(HaveOccurred())
				newRunaiJob.Annotations = map[string]string{
					common.RunaiCurrentAllocatedGpusAnnotation:      "2",
					common.RunaiCurrentAllocatedGpuMemoryAnnotation: "2400",
				}
				err = k8sClient.Update(ctx, newRunaiJob)
				Expect(err).ToNot(HaveOccurred())
				errs := K8sResourcesManager.UpsertResources(ctx, workload, objects, ResourceOptions{})
				Expect(errs).To(BeNil())
				err = k8sClient.Get(ctx, k8sclient.ObjectKeyFromObject(runaiJob), newRunaiJob)
				Expect(err).ToNot(HaveOccurred())
				Expect(newRunaiJob.GetAnnotations()).To(HaveKeyWithValue(common.RunaiCurrentAllocatedGpusAnnotation, "2"))

			})
			It("Should upsert all resources", func() {
				errs := K8sResourcesManager.UpsertResources(ctx, workload, objects, ResourceOptions{})
				Expect(errs).To(BeNil())
				newRunaiJob := &runaiJobV1.RunaiJob{}
				err := k8sClient.Get(ctx, k8sclient.ObjectKeyFromObject(runaiJob), newRunaiJob)
				Expect(err).ToNot(HaveOccurred())
				Expect(newRunaiJob.GetName()).To(Equal(runaiJob.GetName()))
				service := &v1.Service{}
				err = k8sClient.Get(ctx, k8sclient.ObjectKeyFromObject(serviceOne), service)
				Expect(err).ToNot(HaveOccurred())
				Expect(service.GetName()).To(Equal(serviceOne.GetName()))
				ingress := &network.Ingress{}
				err = k8sClient.Get(ctx, k8sclient.ObjectKeyFromObject(ingressOne), ingress)
				Expect(err).ToNot(HaveOccurred())
				Expect(ingress.GetName()).To(Equal(ingressOne.GetName()))
			})
			It("Ready condition should not exists, no changes to resources", func() {
				errs := K8sResourcesManager.UpsertResources(ctx, workload, objects, ResourceOptions{})
				Expect(errs).To(BeNil())
				Expect(workload.Status.Conditions).To(HaveLen(0))
			})
			It("should failed to update release label is not equal", func() {
				otherWorkload := workloadutils.NewJobWorkloadBuilder().WithUid(ownerUid).WithName("other-test-job").WithNamespace("test-namespace").Build()
				errs := K8sResourcesManager.UpsertResources(ctx, otherWorkload, objects, ResourceOptions{})
				Expect(errs).ToNot(BeNil())
				Expect(errs[0].Error()).To(ContainSubstring("already exists"))
			})
		})
	})
	Context("Upsert Issues", func() {
		var (
			fakeClientPtr *client.FakeClientWithErrorRules

			workload       *job.JobWorkload
			allTestObjects []k8sclient.Object
			service        *v1.Service
			ingress        *network.Ingress
			runaiJob       *runaiJobV1.RunaiJob
			pvc            *v1.PersistentVolumeClaim
			cd             *nvapi.ComputeDomain
			jax            *kube.JAXJob

			labels   map[string]string
			ownerUid = types.UID("test-uid")
			kind     = constants.KindInteractiveWorkload
		)

		BeforeEach(func() {
			fakeClientPtr = client.NewFakeClientWithErrorRules(scheme, []k8sclient.Object{&runaiJobV1.RunaiJob{}, &iw.InteractiveWorkload{}}, nil)
			k8sClient = fakeClientPtr
			K8sResourcesManager = NewResourceManager(k8sClient, k8sClient, &issue2.ComputeResourceGrouper{})

			workload = &job.JobWorkload{}
			workload.Kind = kind
			workload.SetName(workloadName)
			workload.SetUID(ownerUid)
			labels = map[string]string{constants.LabelRelease: workload.Name}

			service = resourceutils.CreateDumbService(TestServiceName1, projectNamespace, ownerUid, kind, labels)
			ingress = resourceutils.CreateDumbIngress(TestIngressName1, projectNamespace, ownerUid, kind, labels)
			runaiJob = resourceutils.CreateDumbRunaiJob(TestRunaiJobName, projectNamespace, ownerUid, kind, runaiJobV1.JobSpec{})
			pvc = resourceutils.CreatePvc(TestPvc1, projectNamespace, resource.MustParse(pvcSize), ownerUid, kind, "", nil, nil, labels)
			cd = resourceutils.CreateDumbComputeDomain(TestComputeDomainName, projectNamespace, ownerUid, kind, labels)
			jax = resourceutils.CreateDumbJAXJob(TestJaxWorkloadName, projectNamespace, ownerUid, kind, labels)

			allTestObjects = []k8sclient.Object{service, ingress, runaiJob, pvc, cd, jax}

			ctx = injectWorkloadEventRecorder(ctx, workload, recorder)
		})

		It("Create issues, compute resource not found - critical", func() {
			rules := []client.ResourceFailingRule{
				{
					ClientAction: client.Create,
					ResourceKind: constants.KindIngress,
					ResourceName: ingress.GetName(),
					StatusError:  &k8serrors.StatusError{ErrStatus: metav1.Status{Message: client.UrlAlreadyExistsMsg, Code: http.StatusBadRequest}},
				},
				{
					ClientAction: client.Create,
					ResourceKind: constants.KindPvc,
					ResourceName: pvc.GetName(),
				},
				{
					ClientAction: client.Create,
					ResourceKind: constants.KindComputeDomain,
					ResourceName: cd.GetName(),
				},
				{
					ClientAction: client.Create,
					ResourceKind: constants.KindJAXJob,
					ResourceName: jax.GetName(),
					StatusError: &k8serrors.StatusError{ErrStatus: metav1.Status{
						Code:    http.StatusInternalServerError,
						Message: "Internal error occurred: failed calling webhook \"bla bla bla\"",
						Details: &metav1.StatusDetails{Causes: []metav1.StatusCause{{
							Message: "failed calling webhook \"validator.jaxjob.training-operator.kubeflow.org\": failed to call webhook: the server could not find the requested resource",
						}}}},
					},
				},
			}
			fakeClientPtr.UpdateRules(rules)

			issues := K8sResourcesManager.UpsertResources(ctx, workload, allTestObjects, ResourceOptions{})
			Expect(issues).NotTo(BeNil())
			Expect(issues).To(HaveLen(len(rules)))

			var ingressIssue, pvcIssue, cdIssue, jaxIssue *issue.Issue
			for _, iss := range issues {
				if iss.Kind() == constants.KindPvc {
					pvcIssue = iss
				}
				if iss.Kind() == constants.KindIngress {
					ingressIssue = iss
				}
				if iss.Kind() == constants.KindComputeDomain {
					cdIssue = iss
				}
				if iss.Kind() == constants.KindJAXJob {
					jaxIssue = iss
				}
			}

			Expect(ingressIssue).NotTo(BeNil())
			Expect(ingressIssue.IsCritical()).To(BeTrue())
			Expect(ingressIssue.Message()).To(ContainSubstring(client.UrlAlreadyExistsMsg))
			Expect(ingressIssue.GetGroup()).To(Equal(issue.Network))

			Expect(pvcIssue).NotTo(BeNil())
			Expect(pvcIssue.IsCritical()).To(BeTrue())
			Expect(pvcIssue.Message()).To(ContainSubstring(fmt.Sprintf(client.GeneralErrorMsg, pvc.GetName())))
			Expect(pvcIssue.GetGroup()).To(Equal(issue.Storage))

			Expect(cdIssue).NotTo(BeNil())
			Expect(cdIssue.IsCritical()).To(BeFalse()) //Compute upsert issues are not critical
			Expect(cdIssue.Message()).To(ContainSubstring(fmt.Sprintf(client.GeneralErrorMsg, cd.GetName())))
			Expect(cdIssue.GetGroup()).To(Equal(issue.Compute))

			Expect(jaxIssue).NotTo(BeNil())
			Expect(jaxIssue.IsCritical()).To(BeTrue()) //Compute upsert issues with webhook error are critical
			Expect(jaxIssue.Message()).To(Equal("[JAXJob:test-jax-workload] Internal error occurred: failed calling webhook \"bla bla bla\""))
			Expect(jaxIssue.GetGroup()).To(Equal(issue.Compute))

			Eventually(recorder.Events).WithTimeout(3 * time.Second).
				Should(Receive(ContainSubstring(fmt.Sprintf("Failed to create %s %s", constants.KindComputeDomain, cd.Name))))
		})

		It("Create issues, compute resource certificate request error - not critical", func() {
			rules := []client.ResourceFailingRule{
				{
					ClientAction: client.Create,
					ResourceKind: constants.KindJAXJob,
					ResourceName: jax.GetName(),
					StatusError: &k8serrors.StatusError{ErrStatus: metav1.Status{
						Code:    http.StatusInternalServerError,
						Message: "Internal error occurred: failed calling webhook \"bla bla bla2\"",
						Details: &metav1.StatusDetails{Causes: []metav1.StatusCause{{
							Message: "Internal error occurred: failed calling webhook \"runai-admission-controller.runai.svc\": failed to call webhook: Post \"https://runai-admission-controller.runai.svc:443/mutate-pod-controller?timeout=10s\": tls: failed to verify certificate: x509: certificate signed by unknown authority (possibly because of \"x509: invalid signature: parent certificate cannot sign this kind of certificate\" while trying to verify candidate authority certificate \"runai-admission-controller.runai.svc\")",
						}}}},
					},
				},
			}
			fakeClientPtr.UpdateRules(rules)

			issues := K8sResourcesManager.UpsertResources(ctx, workload, allTestObjects, ResourceOptions{})
			Expect(issues).NotTo(BeNil())
			Expect(issues).To(HaveLen(len(rules)))

			var jaxIssue *issue.Issue
			for _, iss := range issues {
				if iss.Kind() == constants.KindJAXJob {
					jaxIssue = iss
				}
			}

			Expect(jaxIssue).NotTo(BeNil())
			Expect(jaxIssue.IsCritical()).To(BeFalse()) //Compute upsert issues with webhook error are critical
			Expect(jaxIssue.Message()).To(Equal("[JAXJob:test-jax-workload] Internal error occurred: failed calling webhook \"bla bla bla2\""))
			Expect(jaxIssue.GetGroup()).To(Equal(issue.Compute))

			Eventually(recorder.Events).WithTimeout(3 * time.Second).
				Should(Receive(Equal("Normal SuccessfulCreateResource Successfully created ComputeDomain test-compute-domain")))
		})

		It("Update issues", func() {
			for _, obj := range allTestObjects {
				Expect(k8sClient.Create(ctx, obj)).To(Succeed())
			}

			rules := []client.ResourceFailingRule{
				{
					ClientAction: client.Update,
					ResourceKind: constants.KindService,
					ResourceName: service.GetName(),
				},
				{
					ClientAction: client.Update,
					ResourceKind: constants.KindComputeDomain,
					ResourceName: cd.GetName(),
				},
			}
			fakeClientPtr.UpdateRules(rules)

			issues := K8sResourcesManager.UpsertResources(ctx, workload, allTestObjects, ResourceOptions{})
			Expect(issues).NotTo(BeNil())
			// we currently don't update compute domain so it won't produce an issue
			Expect(issues).To(HaveLen(1))

			var serviceIssue, cdIssue *issue.Issue
			for _, iss := range issues {
				if iss.Kind() == constants.KindService {
					serviceIssue = iss
				}
				if iss.Kind() == constants.KindComputeDomain {
					cdIssue = iss
				}
			}

			Expect(serviceIssue).NotTo(BeNil())
			Expect(serviceIssue.IsCritical()).To(BeTrue())
			Expect(serviceIssue.Message()).To(ContainSubstring(fmt.Sprintf(client.GeneralErrorMsg, service.GetName())))
			Expect(serviceIssue.GetGroup()).To(Equal(issue.Network))

			// we currently don't update compute domain
			Expect(cdIssue).To(BeNil())

			// making sure we don't send event for update to not spam
			Consistently(recorder.Events).ShouldNot(Receive())
		})

		It("Compute upsert issues are not critical", func() {
			pyJob := resourceutils.CreateDumbPytorch(TestPytorchJobName, projectNamespace, ownerUid, kind, labels)
			mpiJob := resourceutils.CreateDumbMPI(TestMPIJobName, projectNamespace, ownerUid, kind, labels)
			jaxJob := resourceutils.CreateDumbJAXJob(TestJaxJobName, projectNamespace, ownerUid, kind, labels)
			tfjob := resourceutils.CreateDumbTensorflow(TestTFJobName, projectNamespace, ownerUid, kind, labels)
			xgbJob := resourceutils.CreateDumbXGBoost(TestXGBJobName, projectNamespace, ownerUid, kind, labels)

			allTestObjects = append(allTestObjects, pyJob)
			allTestObjects = append(allTestObjects, mpiJob)
			allTestObjects = append(allTestObjects, jaxJob)
			allTestObjects = append(allTestObjects, tfjob)
			allTestObjects = append(allTestObjects, xgbJob)

			for _, obj := range allTestObjects {
				Expect(k8sClient.Create(ctx, obj)).To(Succeed())
			}

			rules := []client.ResourceFailingRule{
				{
					ClientAction: client.Update,
					ResourceKind: constants.KindRunaiJob,
					ResourceName: runaiJob.GetName(),
					StatusError:  &k8serrors.StatusError{ErrStatus: metav1.Status{Message: "Unauthorized", Code: http.StatusUnauthorized}},
				},
				{
					ClientAction: client.Update,
					ResourceKind: constants.KindPyTorchJob,
					ResourceName: pyJob.GetName(),
					StatusError:  &k8serrors.StatusError{ErrStatus: metav1.Status{Message: "Unauthorized", Code: http.StatusUnauthorized}},
				},
				{
					ClientAction: client.Update,
					ResourceKind: constants.KindMpiJob,
					ResourceName: mpiJob.GetName(),
					StatusError:  &k8serrors.StatusError{ErrStatus: metav1.Status{Message: "Unauthorized", Code: http.StatusUnauthorized}},
				},
				{
					ClientAction: client.Update,
					ResourceKind: constants.KindJAXJob,
					ResourceName: jaxJob.GetName(),
					StatusError:  &k8serrors.StatusError{ErrStatus: metav1.Status{Message: "Unauthorized", Code: http.StatusUnauthorized}},
				},
				{
					ClientAction: client.Update,
					ResourceKind: constants.KindTFJob,
					ResourceName: tfjob.GetName(),
					StatusError:  &k8serrors.StatusError{ErrStatus: metav1.Status{Message: "Unauthorized", Code: http.StatusUnauthorized}},
				},
				{
					ClientAction: client.Update,
					ResourceKind: constants.KindXGBoostJob,
					ResourceName: xgbJob.GetName(),
					StatusError:  &k8serrors.StatusError{ErrStatus: metav1.Status{Message: "Unauthorized", Code: http.StatusUnauthorized}},
				},
			}
			fakeClientPtr.UpdateRules(rules)

			issues := K8sResourcesManager.UpsertResources(ctx, workload, allTestObjects, ResourceOptions{})
			Expect(issues).NotTo(BeNil())
			Expect(issues).To(HaveLen(len(rules)))

			issueMap := make(map[string]*issue.Issue)
			for _, iss := range issues {
				issueMap[iss.Kind()] = iss
			}

			expectedKinds := []string{
				constants.KindRunaiJob,
				constants.KindPyTorchJob,
				constants.KindMpiJob,
				constants.KindJAXJob,
				constants.KindTFJob,
				constants.KindXGBoostJob,
			}
			for _, kind := range expectedKinds {
				Expect(issueMap[kind]).NotTo(BeNil())
				Expect(issueMap[kind].IsCritical()).To(BeFalse())
				Expect(issueMap[kind].Message()).To(ContainSubstring("Unauthorized"))
				Expect(issueMap[kind].GetGroup()).To(Equal(issue.Compute))
			}
		})

		DescribeTable("non-critical errors",
			func(errorCode int, errorMsg string, errorReason metav1.StatusReason) {
				rules := []client.ResourceFailingRule{
					{
						ClientAction: client.Create,
						ResourceKind: constants.KindIngress,
						ResourceName: ingress.GetName(),
						StatusError:  &k8serrors.StatusError{ErrStatus: metav1.Status{Message: errorMsg, Code: int32(errorCode), Reason: errorReason}},
					},
				}
				fakeClientPtr.UpdateRules(rules)

				issues := K8sResourcesManager.UpsertResources(ctx, workload, allTestObjects, ResourceOptions{})
				Expect(issues).NotTo(BeNil())
				Expect(issues).To(HaveLen(len(rules)))

				var ingressIssue *issue.Issue
				for _, iss := range issues {
					if iss.Kind() == constants.KindIngress {
						ingressIssue = iss
					}
				}

				Expect(ingressIssue).NotTo(BeNil())
				Expect(ingressIssue.IsCritical()).To(BeFalse())
				Expect(ingressIssue.Message()).To(ContainSubstring(errorMsg))
				Expect(ingressIssue.GetGroup()).To(Equal(issue.Network))
			},
			Entry("429 - too many requests", http.StatusTooManyRequests, "too many, try again later", metav1.StatusReasonTooManyRequests),
			Entry("409 - conflict - the object has been modified", http.StatusConflict, "the object has been modified; please apply your changes to the latest version and try again", metav1.StatusReasonConflict),
			Entry("500 - internal server error", http.StatusInternalServerError, "ERROR", metav1.StatusReasonInternalError),
			Entry("303", http.StatusSeeOther, "ERROR", metav1.StatusReasonUnknown),
		)
	})
})

func injectWorkloadEventRecorder(ctx context.Context, workload *job.JobWorkload, recorder record.EventRecorder) context.Context {
	converter := convertors.NewGeneralWorkloadConverter()
	workloadObj := converter.Convert(workload)
	Expect(workloadObj).NotTo(BeNil())

	workloadRecorder := events.NewWorkloadRecorder(workloadObj, recorder)
	return events.ToContext(ctx, workloadRecorder)
}
