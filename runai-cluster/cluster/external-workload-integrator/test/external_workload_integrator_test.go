package test

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	pluginsConstants "github.com/kai-scheduler/KAI-scheduler/pkg/podgrouper/podgrouper/plugins/constants"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	kaiv2alpha2 "github.com/kai-scheduler/KAI-scheduler/pkg/apis/scheduling/v2alpha2"
	rayv1alpha1 "github.com/ray-project/kuberay/ray-operator/apis/ray/v1alpha1"
	"github.com/run-ai/runai/runai-cluster/cluster/common/dra"
	"github.com/run-ai/runai/runai-cluster/cluster/external-workload-integrator/pkg/cachedclient"
	"github.com/run-ai/runai/runai-cluster/cluster/external-workload-integrator/pkg/config"
	"github.com/run-ai/runai/runai-cluster/cluster/external-workload-integrator/pkg/constants"
	"github.com/run-ai/runai/runai-cluster/cluster/external-workload-integrator/pkg/externalworkloads"
	kservev1beta1 "github.com/run-ai/runai/runai-cluster/cluster/external-workload-integrator/pkg/thirdpartyapi/kserve/v1beta1"
	kubeflowv1beta1 "github.com/run-ai/runai/runai-cluster/cluster/external-workload-integrator/pkg/thirdpartyapi/kubeflownotebook/v1beta1"
	nim "github.com/run-ai/runai/runai-cluster/cluster/external-workload-integrator/pkg/thirdpartyapi/nim/v1alpha1"
	"github.com/run-ai/runai/runai-cluster/cluster/external-workload-integrator/pkg/utils"
	"github.com/run-ai/runai/runai-cluster/cluster/external-workload-integrator/test/fakecachedclient"
	"github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1alpha1"
	wlconfig "github.com/run-ai/runai/runai-cluster/cluster/workload-common/config"
	commonconsts "github.com/run-ai/runai/runai-cluster/common/constants"

	servingv1 "knative.dev/serving/pkg/apis/serving/v1"

	appsv1 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/rand"
	kubefake "k8s.io/client-go/kubernetes/fake"
	clienttesting "k8s.io/client-go/testing"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	infv1 "sigs.k8s.io/gateway-api-inference-extension/api/v1"
	gwv1 "sigs.k8s.io/gateway-api/apis/v1"
	gwv1alpha2 "sigs.k8s.io/gateway-api/apis/v1alpha2"
)

const (
	timeout             = time.Second * 5
	interval            = time.Millisecond * 500
	consistentlyTimeout = time.Millisecond * 2500
	testStatusMessage   = "some-mesagee"
)

func TestExternalWorkloadIntegrator(t *testing.T) {
	zerolog.SetGlobalLevel(zerolog.DebugLevel)
	log.Info().Msgf("Started service tests")
	RegisterFailHandler(Fail)
	MockEnv(t)
	InitVars()
	RunSpecs(t, "external-workload-integrator Test Suite")
}

type TestData struct {
	testName                 string
	ownerObjectToCreate      client.Object
	childObjectsToCreate     []client.Object
	expectedExternalWorkload *v1alpha1.ExternalWorkload
}

var _ = Describe("TestExternalWorkloadIntegrator", Ordered, func() {
	var (
		watchClient             client.WithWatch
		fakeCachedClient        cachedclient.CachedClient
		externalWorkloadHandler *externalworkloads.ExternalWorkloadHandler
		podGroup                *kaiv2alpha2.PodGroup

		ctx     context.Context
		cancel  context.CancelFunc
		stopper chan struct{}

		testCreatedObjects = []client.Object{}

		kartasForTest         = []client.Object{KartaForKserveInferenceService, KartaForDeployment}
		additionalGvksForTest = []schema.GroupVersionKind{kservev1beta1.SchemeGroupVersion.WithKind("InferenceService")}
	)

	BeforeAll(func() {
		stopper = make(chan struct{}, 1)
		ctx, cancel = context.WithCancel(context.Background())
		DeferCleanup(func() {
			close(stopper)
			cancel()
		})

		ns := BuildNamespace(TestNamespace, map[string]string{constants.RunaiQueueLabel: testProjectName})
		ns2 := BuildNamespace(TestNonProjectNamespace, map[string]string{})
		project := buildProject(testProjectName, testDepartmentName, testProjectId, testDepartmentId)
		kaiProject := buildKaiProject(testProjectName, testDepartmentName, testProjectId, testDepartmentId)
		podGroup = TestPodGroup.DeepCopy()

		scheme := getTestScheme()
		workloadConfigMap := CreateWorkloadsConfigMapObject()
		// Use a plain object tracker (no field management). The default fake-client
		// tracker computes managed fields via structured-merge-diff, which panics with
		// "unsupported type: uint64" on third-party types such as kubevirt's
		// VirtualMachineInstance. A plain tracker restores the pre-upgrade behavior.
		watchClient = fake.NewClientBuilder().WithScheme(scheme).WithStatusSubresource(&v1alpha1.ExternalWorkload{}).
			WithObjectTracker(clienttesting.NewObjectTracker(scheme, serializer.NewCodecFactory(scheme).UniversalDecoder())).
			WithObjects(ns, ns2, project, kaiProject, TestNodePool, TestKaiNodePool, podGroup, workloadConfigMap).WithObjects(kartasForTest...).Build()

		draParser := dra.NewDRAParser(
			fake.NewClientBuilder().WithScheme(getTestScheme()).WithObjects(TestDRAGpuClaimTemplate.DeepCopy(), TestDRAGpuClaim.DeepCopy()).Build(),
			kubefake.NewClientset(TestDRAGpuClaimTemplate.DeepCopy(), TestDRAGpuClaim.DeepCopy()),
		)

		fakeCachedClient = fakecachedclient.NewFakeCachedClient(watchClient, stopper, false)

		externalWorkloadHandler = externalworkloads.StartExternalWorkloadIntegrator(ctx, stopper,
			fakeCachedClient, draParser, config.Config().AllowNonSupportedTypes, additionalGvksForTest)
		externalWorkloadHandler.StructureUpdater().SetOwnedObjectEnqueueDelay(0)
	})

	AfterEach(func() {
		for i := range testCreatedObjects {
			testCreatedObject := testCreatedObjects[i]
			err := watchClient.Delete(ctx, testCreatedObject)
			if err != nil {
				Expect(errors.IsNotFound(err)).To(BeTrue())
			}
		}
		testCreatedObjects = []client.Object{}
	})

	AfterAll(func() {
		externalWorkloadHandler = nil
	})

	DescribeTable("Sanity - externalworkload exists with correct info for all supported types",
		Ordered,
		func(testData *TestData) {
			// create the top owned object
			testData.ownerObjectToCreate = createObject(ctx, testData.ownerObjectToCreate,
				watchClient, &testCreatedObjects)

			// some types (e.g. nvcf) have a different UID for the external workload
			expectedTopOwnerWorkloadId := testData.ownerObjectToCreate.GetUID()
			expectedClusterTopOwnerId := expectedTopOwnerWorkloadId
			if testData.expectedExternalWorkload.Spec.OriginalWorkloadUID != "" {
				expectedTopOwnerWorkloadId = testData.expectedExternalWorkload.Spec.OriginalWorkloadUID
			}

			// create its children - each object is the next's owner
			expectedChildIds := []v1alpha1.ChildId{}
			prevOwner := testData.ownerObjectToCreate
			for i := range testData.childObjectsToCreate {
				if len(testData.childObjectsToCreate[i].GetOwnerReferences()) == 0 {
					testData.childObjectsToCreate[i].SetOwnerReferences([]metav1.OwnerReference{{
						UID:  prevOwner.GetUID(),
						Name: prevOwner.GetName(),
					}})
				}
				// sleep - so the creation of the objects doesn't happen at the exact same time.
				// the tests infra doesn't handle that well;
				// in a real cluster scenario the objects are never created at the exact same time.
				time.Sleep(30 * time.Millisecond)
				testData.childObjectsToCreate[i] = createObject(ctx, testData.childObjectsToCreate[i],
					watchClient, &testCreatedObjects)
				prevOwner = testData.childObjectsToCreate[i]

				if externalWorkloadHandler.EventHandlerFilter(ctx, testData.childObjectsToCreate[i],
					testData.childObjectsToCreate[i].GetObjectKind().GroupVersionKind()) &&
					externalworkloads.IsValidTypeForChildId(testData.childObjectsToCreate[i]) {
					// only add to expected child ids if object passed the filter
					expectedChildIds = append(expectedChildIds, v1alpha1.ChildId{
						UID:  string(testData.childObjectsToCreate[i].GetUID()),
						Type: testData.childObjectsToCreate[i].GetObjectKind().GroupVersionKind().Kind,
					})
				}
			}

			// update the original object to cause an update event and hence another reconciliation
			updateObjectWithRandomLabel(ctx, testData.ownerObjectToCreate, watchClient)

			// validate external workload object for the top owned object with correct child ids
			externalWorkloadOfObject := eventuallyFindAndValidateExternalWorkloadWithChildIds(fakeCachedClient,
				testData.expectedExternalWorkload, testData.ownerObjectToCreate, expectedChildIds,
				testData.testName, &testCreatedObjects)

			Expect(len(externalWorkloadOfObject.Status.ObservedStatus.ChildIds)).To(
				Equal(len(expectedChildIds)),
				fmt.Sprintf("Test: <%s>, ChildIds-len", testData.testName))

			for i := range testData.childObjectsToCreate {
				externalWorkloadOfChildObject := findExternalWorkloadOfObject(fakeCachedClient,
					testData.childObjectsToCreate[i].GetUID(), testData.childObjectsToCreate[i].GetNamespace())
				Expect(externalWorkloadOfChildObject).To(BeNil(),
					"Test: <%s>, external workload for child object with UID <%s> should not exist",
					testData.testName, testData.childObjectsToCreate[i].GetUID())

				pod, ok := testData.childObjectsToCreate[i].(*corev1.Pod)
				if ok && pod != nil && pod.Spec.SchedulerName == constants.RunaiSchedulerName {
					eventuallyExpectPodLabeledWithTopOwnerWorkloadId(
						pod,
						string(expectedTopOwnerWorkloadId), string(expectedClusterTopOwnerId),
						fakeCachedClient)
				}
			}
		},
		Entry("Sanity - create deployment", &TestData{
			testName:            "Sanity - create deployment",
			ownerObjectToCreate: TestDeployment.DeepCopy(),
			childObjectsToCreate: []client.Object{
				testDeploymentChild1RS.DeepCopy(),
				testDeploymentChild2Pod.DeepCopy(),
			},
			expectedExternalWorkload: testDeploymentExpectedEw,
		}),
		Entry("Create deployment - pod without runai scheduler - not in child ids", &TestData{
			testName:            "Create deployment - pod without runai scheduler - not in child ids",
			ownerObjectToCreate: TestDeployment.DeepCopy(),
			childObjectsToCreate: []client.Object{
				testDeploymentChild1RS.DeepCopy(),
				testDeploymentChildPodBadSchedulerName.DeepCopy(),
			},
			expectedExternalWorkload: testDeploymentExpectedEw,
		}),
		Entry("Create deployment - additional child id of PodGroup", &TestData{
			testName:            "Create deployment - additional child id of PodGroup",
			ownerObjectToCreate: TestDeployment.DeepCopy(),
			childObjectsToCreate: []client.Object{
				testDeploymentChild1RS.DeepCopy(),
				testDeploymentChildPodWithPodGroup.DeepCopy(),
			},
			expectedExternalWorkload: testDeploymentExpectedEw,
		}),
		Entry("Sanity - create pod", &TestData{
			testName:                 "Sanity - create pod",
			ownerObjectToCreate:      testPod2.DeepCopy(),
			childObjectsToCreate:     []client.Object{},
			expectedExternalWorkload: testPodExpectedEw2,
		}),
		Entry("Create pod with additional PodGroupId as childId", &TestData{
			testName:                 "Create pod with additional PodGroupId as childId",
			ownerObjectToCreate:      testPodWithPodGroupAnnotation.DeepCopy(),
			childObjectsToCreate:     []client.Object{},
			expectedExternalWorkload: testPodExpectedEw2,
		}),
		Entry("Sanity - create replicaset", &TestData{
			testName:            "Sanity - create replicaset",
			ownerObjectToCreate: TestReplicaSet.DeepCopy(),
			childObjectsToCreate: []client.Object{
				testReplicaSetChild1Pod.DeepCopy(),
			},
			expectedExternalWorkload: testReplicaSetExpectedEw,
		}),
		Entry("Sanity - create statefulset", &TestData{
			testName:            "Sanity - create statefulset",
			ownerObjectToCreate: TestStatefulSet.DeepCopy(),
			childObjectsToCreate: []client.Object{
				testStatefulSetChild1Pod.DeepCopy(),
			},
			expectedExternalWorkload: testStatefulSetExpectedEw,
		}),
		Entry("Sanity - create NIM from statefulset", &TestData{
			testName:            "Sanity - create NIM from statefulset",
			ownerObjectToCreate: testNimStatefulSet.DeepCopy(),
			childObjectsToCreate: []client.Object{
				testNimStatefulSetChild1Pod.DeepCopy(),
			},
			expectedExternalWorkload: testNimStatefulSetExpectedEw,
		}),
		Entry("Sanity - create v1beta1 DynamoGraphDeployment", &TestData{
			testName:                 "Sanity - create v1beta1 DynamoGraphDeployment",
			ownerObjectToCreate:      DynamoGraphDeploymentV1beta1Basic.DeepCopy(),
			childObjectsToCreate:     []client.Object{},
			expectedExternalWorkload: DynamoGraphDeploymentV1beta1BasicExpectedEw,
		}),
		Entry("Sanity - create batch job", &TestData{
			testName:            "Sanity - create batch job",
			ownerObjectToCreate: TestBatchJob.DeepCopy(),
			childObjectsToCreate: []client.Object{
				testBatchJobChild1Pod.DeepCopy(),
				testBatchJobChild2Pod.DeepCopy(),
			},
			expectedExternalWorkload: testBatchJobExpectedEw,
		}),
		Entry("Sanity - create cronjob", &TestData{
			testName:            "Sanity - create cronjob",
			ownerObjectToCreate: TestCronJob.DeepCopy(),
			childObjectsToCreate: []client.Object{
				testCronJobChild1Job.DeepCopy(),
				testCronJobChild2Pod.DeepCopy(),
			},
			expectedExternalWorkload: testCronJobExpectedEw,
		}),
		Entry("Sanity - create rayjob", &TestData{
			testName:            "Sanity - create rayjob",
			ownerObjectToCreate: TestRayJob.DeepCopy(),
			childObjectsToCreate: []client.Object{
				testRayJobChildRayCluster.DeepCopy(),
				testRayJobChild1Pod.DeepCopy(),
				testRayJobChild2Pod.DeepCopy(),
			},
			expectedExternalWorkload: TestRayJobExpectedEw,
		}),
		Entry("Sanity - create raycluster", &TestData{
			testName:            "Sanity - create raycluster",
			ownerObjectToCreate: TestRayCluster.DeepCopy(),
			childObjectsToCreate: []client.Object{
				testRayClusterChildPod.DeepCopy(),
			},
			expectedExternalWorkload: TestRayClusterExpectedEw,
		}),
		Entry("Sanity - create rayservice", &TestData{
			testName:                 "Sanity - create rayservice",
			ownerObjectToCreate:      TestRayService.DeepCopy(),
			expectedExternalWorkload: TestRayServiceExpectedEw,
		}),
		Entry("Sanity - create rayjob v1", &TestData{
			testName:                 "Sanity - create rayjob v1",
			ownerObjectToCreate:      TestRayJobV1.DeepCopy(),
			expectedExternalWorkload: TestRayJobExpectedEw,
		}),
		Entry("Sanity - create raycluster v1", &TestData{
			testName:                 "Sanity - create raycluster v1",
			ownerObjectToCreate:      TestRayClusterV1.DeepCopy(),
			expectedExternalWorkload: TestRayClusterExpectedEw,
		}),
		Entry("Sanity - create rayservice v1", &TestData{
			testName:                 "Sanity - create rayservice v1",
			ownerObjectToCreate:      TestRayServiceV1.DeepCopy(),
			expectedExternalWorkload: TestRayServiceExpectedEw,
		}),
		Entry("Sanity - create knative service", &TestData{
			testName:            "Sanity - create knative service",
			ownerObjectToCreate: TestKnativeService.DeepCopy(),
			childObjectsToCreate: []client.Object{
				testKnativeChildConfiguration.DeepCopy(),
				testKnativeChildRevision.DeepCopy(),
			},
			expectedExternalWorkload: TestKnativeServiceExpectedEw,
		}),
		Entry("Sanity - create knative service inference", &TestData{
			testName:            "Sanity - create knative service inference",
			ownerObjectToCreate: TestKnativeServiceInference.DeepCopy(),
			childObjectsToCreate: []client.Object{
				testKnativeChildConfiguration.DeepCopy(),
				testKnativeChildRevision.DeepCopy(),
			},
			expectedExternalWorkload: TestKnativeServiceInferenceExpectedEw,
		}),
		Entry("Sanity - create seldon deployment v1", &TestData{
			testName:                 "Sanity - create seldon deployment v1",
			ownerObjectToCreate:      TestSeldonDeployment1.DeepCopy(),
			expectedExternalWorkload: TestSeldonDeploymentExpectedEw,
		}),
		Entry("Sanity - create seldon deployment v1alpha2", &TestData{
			testName:                 "Sanity - create seldon deployment v1alpha2",
			ownerObjectToCreate:      TestSeldonDeployment2.DeepCopy(),
			expectedExternalWorkload: TestSeldonDeploymentExpectedEw,
		}),
		Entry("Sanity - create TFJob", &TestData{
			testName:                 "Sanity - create TFJob",
			ownerObjectToCreate:      TestTFJob1.DeepCopy(),
			expectedExternalWorkload: TestTFJobExpectedEw,
		}),
		Entry("Sanity - create PytorchJob", &TestData{
			testName:            "Sanity - create PytorchJob",
			ownerObjectToCreate: TestPytorchJob1.DeepCopy(),
			childObjectsToCreate: []client.Object{
				testPytorchJobChild1Pod.DeepCopy(),
			},
			expectedExternalWorkload: TestPytorchJobExpectedEw,
		}),
		Entry("Sanity - create XGBoostJob", &TestData{
			testName:                 "Sanity - create XGBoostJob",
			ownerObjectToCreate:      TestXGBoostJob1.DeepCopy(),
			expectedExternalWorkload: TestXGBoostJobExpectedEw,
		}),
		Entry("Sanity - create JAXJob", &TestData{
			testName:                 "Sanity - create JAXJob",
			ownerObjectToCreate:      TestJAXJob1.DeepCopy(),
			expectedExternalWorkload: TestJAXJobExpectedEw,
		}),
		Entry("Sanity - create MPIJob v1", &TestData{
			testName:                 "Sanity - create MPIJob v1",
			ownerObjectToCreate:      TestMPIJobV1.DeepCopy(),
			expectedExternalWorkload: TestMPIJobV1ExpectedEw,
		}),
		Entry("Sanity - create MPIJob v2beta1", &TestData{
			testName:                 "Sanity - create MPIJob v2beta1",
			ownerObjectToCreate:      TestMPIJobV2beta1.DeepCopy(),
			expectedExternalWorkload: TestMPIJobV2beta1ExpectedEw,
		}),
		Entry("Sanity - create Notebook", &TestData{
			testName:                 "Sanity - create Notebook",
			ownerObjectToCreate:      TestNotebook1.DeepCopy(),
			expectedExternalWorkload: TestNotebookExpectedEw,
		}),
		Entry("Sanity - create kubevirt VirtualMachineInstance", &TestData{
			testName:                 "Sanity - kubevirt VirtualMachineInstance",
			ownerObjectToCreate:      TestKubevirtVMI.DeepCopy(),
			expectedExternalWorkload: TestKubevirtVMIExpectedEw,
		}),
		Entry("Sanity - create DevWorkspace", &TestData{
			testName:            "Sanity - DevWorkspace",
			ownerObjectToCreate: TestDevWorkspace.DeepCopy(),
			childObjectsToCreate: []client.Object{
				testTestDevWorkspaceChild1Pod.DeepCopy(),
			},
			expectedExternalWorkload: TestDevWorkspaceExpectedEw,
		}),
		Entry("Sanity - create nimcache", &TestData{
			testName:                 "Sanity - create nimcache",
			ownerObjectToCreate:      testNimCache.DeepCopy(),
			expectedExternalWorkload: TestNimCacheExpectedBasicEw,
		}),
		Entry("Sanity - create nimservice", &TestData{
			testName:                 "Sanity - create nimservice",
			ownerObjectToCreate:      testNimService.DeepCopy(),
			expectedExternalWorkload: TestNimServiceExpectedBasicEw,
		}),
		Entry("Sanity - nimservice with update in progress", &TestData{
			testName:                 "Sanity - create nimservice with update in progress",
			ownerObjectToCreate:      testNimServiceUpdating.DeepCopy(),
			expectedExternalWorkload: TestNimServiceUpdatingExpectedEw,
		}),
		Entry("Sanity - nimservice during creation", &TestData{
			testName:                 "Sanity - create nimservice during creation",
			ownerObjectToCreate:      testNimServiceCreating.DeepCopy(),
			expectedExternalWorkload: TestNimServiceCreatingExpectedEw,
		}),
		Entry("Sanity - nimservice failed with no message", &TestData{
			testName:                 "Sanity - nimservice failed with no message",
			ownerObjectToCreate:      testNimServiceFailed.DeepCopy(),
			expectedExternalWorkload: TestNimServiceFailedExpectedEw,
		}),
		Entry("Sanity - nimservice failed with error message", &TestData{
			testName:                 "Sanity - nimservice failed with error message",
			ownerObjectToCreate:      testNimServiceFailedWithMessage.DeepCopy(),
			expectedExternalWorkload: TestNimServiceFailedWithMessageExpectedEw,
		}),
		// MultiNode NIMService tests - validate RequestedPods calculation
		Entry("Sanity - nimservice multiNode with multiple replicas (2 replicas * 4 nodes = 8 pods)", &TestData{
			testName:                 "Sanity - nimservice multiNode with multiple replicas",
			ownerObjectToCreate:      testNimServiceMultiNode.DeepCopy(),
			expectedExternalWorkload: TestNimServiceMultiNodeExpectedEw,
		}),
		Entry("Sanity - nimservice multiNode with single replica (1 replica * 2 nodes = 2 pods)", &TestData{
			testName:                 "Sanity - nimservice multiNode with single replica",
			ownerObjectToCreate:      testNimServiceMultiNodeSingleReplica.DeepCopy(),
			expectedExternalWorkload: TestNimServiceMultiNodeSingleReplicaExpectedEw,
		}),
		Entry("Sanity - nimservice without multiNode (3 replicas * 1 = 3 pods)", &TestData{
			testName:                 "Sanity - nimservice without multiNode",
			ownerObjectToCreate:      testNimServiceNoMultiNode.DeepCopy(),
			expectedExternalWorkload: TestNimServiceNoMultiNodeExpectedEw,
		}),
		// DRA GPU tests
		Entry("Sanity - deployment with DRA GPU claims (2 replicas * 4 GPUs = 8 GPUs)", &TestData{
			testName:                 "Sanity - deployment with DRA GPU claims",
			ownerObjectToCreate:      TestDeploymentWithDRA.DeepCopy(),
			expectedExternalWorkload: TestDeploymentWithDRAExpectedEw,
		}),
		Entry("Sanity - nimservice with DRA ResourceClaim (3 replicas * 2 GPUs = 6 GPUs)", &TestData{
			testName:                 "Sanity - nimservice with DRA ResourceClaim",
			ownerObjectToCreate:      TestNimServiceWithDRAClaim.DeepCopy(),
			expectedExternalWorkload: TestNimServiceWithDRAClaimExpectedEw,
		}),
		Entry("Sanity - nimservice with DRA ResourceClaimTemplate (2 replicas * 4 GPUs = 8 GPUs)", &TestData{
			testName:                 "Sanity - nimservice with DRA ResourceClaimTemplate",
			ownerObjectToCreate:      TestNimServiceWithDRATemplate.DeepCopy(),
			expectedExternalWorkload: TestNimServiceWithDRATemplateExpectedEw,
		}),
		Entry("Sanity - create nvcf function", &TestData{
			testName:            "Sanity - create nvcf function",
			ownerObjectToCreate: TestNvcfFunction.DeepCopy(),
			childObjectsToCreate: []client.Object{
				TestNvcfFunctionChildPod.DeepCopy(),
			},
			expectedExternalWorkload: TestNvcfFunctionExpectedBasicEw,
		}),
		Entry("Sanity - using KartaD - create kserve InferenceService", &TestData{
			testName:                 "Sanity - using KartaD - create kserve InferenceService",
			ownerObjectToCreate:      TestKserveInferenceService.DeepCopy(),
			expectedExternalWorkload: TestKserveExpectedEw,
		}),
	)

	Describe("Test workload objects that no external workload should be created for them", Ordered, func() {
		It("Test all objects - objects that don't pass FilterFunc + onlyOwnedSupportedGVKs objects", func() {
			testPodWithoutSchedulerName := buildPodInner("some-other-nameeee",
				TestNamespace, "some-other-uidddddd", "")
			testPodInDifferentNamespace := buildPodInner("some-other-nameeee2",
				TestNonProjectNamespace, "some-other-uidddddd2", constants.RunaiSchedulerName)
			testDeploymentWithoutSchedulerName := buildDeployment("some-other-depl-name",
				TestNamespace, "some-other-depl-uid", "")
			testDeploymentInDifferentNamespace := buildDeployment("some-other-depl-name2",
				TestNonProjectNamespace, "some-other-depl-uid2", constants.RunaiSchedulerName)
			testReplicaSetWithoutSchedulerName := buildReplicaSet("some-other-replicaset-name",
				TestNamespace, "some-other-replicaset-uid", "")
			testReplicaSetInDifferentNamespace := buildReplicaSet("some-other-replicaset-name2",
				TestNonProjectNamespace, "some-other-replicaset-uid2", constants.RunaiSchedulerName)
			testStatefulSetWithoutSchedulerName := buildStatefulSet("some-other-statefulset-name",
				TestNamespace, "some-other-statefulset-uid", "")
			testStatefulSetInDifferentNamespace := buildStatefulSet("some-other-statefulset-name2",
				TestNonProjectNamespace, "some-other-statefulset-uid2", constants.RunaiSchedulerName)
			testBatchJobWithoutSchedulerName := buildBatchJob("some-other-batchjob-name",
				TestNamespace, "some-other-batchjob-uid", "")
			testBatchJobInDifferentNamespace := buildBatchJob("some-other-batchjob-name2",
				TestNonProjectNamespace, "some-other-batchjob-uid2", constants.RunaiSchedulerName)
			testCronJobWithoutSchedulerName := buildCronJob("some-other-cronjob-name",
				TestNamespace, "some-other-cronjob-uid", "")
			testCronJobInDifferentNamespace := buildCronJob("some-other-cronjob-name2",
				TestNonProjectNamespace, "some-other-cronjob-uid2", constants.RunaiSchedulerName)
			testRayJobWithoutSchedulerName := buildRayJob("some-other-rayjob-name",
				TestNamespace, "some-other-rayjob-uid", "")
			testRayJobInDifferentNamespace := buildRayJob("some-other-rayjob-name2",
				TestNonProjectNamespace, "some-other-rayjob-uid2", constants.RunaiSchedulerName)
			testRayClusterWithoutSchedulerName := buildRayCluster("some-other-raycluster-name",
				TestNamespace, "some-other-raycluster-uid", "")
			testRayClusterInDifferentNamespace := buildRayCluster("some-other-raycluster-name2",
				TestNonProjectNamespace, "some-other-raycluster-uid2", constants.RunaiSchedulerName)
			testRayServiceWithoutSchedulerName := buildRayService("some-other-rayservice-name",
				TestNamespace, "some-other-rayservice-uid", "")
			testRayServiceInDifferentNamespace := buildRayService("some-other-rayservice-name2",
				TestNonProjectNamespace, "some-other-rayservice-uid2", constants.RunaiSchedulerName)
			testRayJobV1WithoutSchedulerName := buildRayJobV1(testRayJobWithoutSchedulerName)
			testRayJobV1InDifferentNamespace := buildRayJobV1(testRayJobInDifferentNamespace)
			testRayClusterV1WithoutSchedulerName := buildRayClusterV1(testRayClusterWithoutSchedulerName)
			testRayClusterV1InDifferentNamespace := buildRayClusterV1(testRayClusterInDifferentNamespace)
			testRayServiceV1WithoutSchedulerName := buildRayServiceV1(testRayServiceWithoutSchedulerName)
			testRayServiceV1InDifferentNamespace := buildRayServiceV1(testRayServiceInDifferentNamespace)
			testKNativeServiceWithoutSchedulerName := buildKNativeService("some-other-knative-name",
				TestNamespace, "some-other-knative-uid", "")
			testKNativeServiceInDifferentNamespace := buildKNativeService("some-other-knative-name2",
				TestNonProjectNamespace, "some-other-knative-uid2", constants.RunaiSchedulerName)
			testKNativeRevision := buildKNativeRevision("some-other-knative-revision-name",
				TestNamespace, "some-other-knative-revision-uid")
			testKNativeConfiguration := buildKNativeConfiguration("some-other-knative-configuration-name",
				TestNamespace, "some-other-knative-configuration-uid")
			testTFJobWithoutSchedulerName := buildTFJob("some-other-tfjob-name",
				TestNamespace, "some-other-tfjob-uid", constants.RunaiSchedulerName, "")
			testTFJobInDifferentNamespace := buildTFJob("some-other-tfjob-name2",
				TestNonProjectNamespace, "some-other-tfjob-uid2", constants.RunaiSchedulerName, constants.RunaiSchedulerName)
			testPytorchJobWithoutSchedulerName := buildPytorchJob("some-other-pytorchjob-name",
				TestNamespace, "some-other-pytorchjob-uid", constants.RunaiSchedulerName, "")
			testPytorchJobInDifferentNamespace := buildPytorchJob("some-other-pytorchjob-name2",
				TestNonProjectNamespace, "some-other-pytorchjob-uid2", constants.RunaiSchedulerName, constants.RunaiSchedulerName)
			testXGBoostJobWithoutSchedulerName := buildXGBoostJob("some-other-xgboostjob-name",
				TestNamespace, "some-other-xgboostjob-uid", "", constants.RunaiSchedulerName)
			testXGBoostJobInDifferentNamespace := buildXGBoostJob("some-other-xgboostjob-name2",
				TestNonProjectNamespace, "some-other-xgboostjob-uid2", constants.RunaiSchedulerName, constants.RunaiSchedulerName)
			testJAXJobWithoutSchedulerName := buildJAXJob("some-other-jaxjob-name",
				TestNamespace, "some-other-jaxjob-uid", "", constants.RunaiSchedulerName)
			testJAXJobInDifferentNamespace := buildJAXJob("some-other-jaxjob-name2",
				TestNonProjectNamespace, "some-other-jaxjob-uid2", constants.RunaiSchedulerName, constants.RunaiSchedulerName)
			testMPIJobV1WithoutSchedulerName := buildMPIJobV1("some-other-mpijob-name",
				TestNamespace, "some-other-mpijob-uid", constants.RunaiSchedulerName, "")
			testMPIJobV1InDifferentNamespace := buildMPIJobV1("some-other-mpijob-name2",
				TestNonProjectNamespace, "some-other-mpijob-uid2", constants.RunaiSchedulerName, constants.RunaiSchedulerName)
			testMPIJobV2beta1WithoutSchedulerName := buildMPIJobV2beta1("some-other-mpijob2-name",
				TestNamespace, "some-other-mpijob-uid", "", constants.RunaiSchedulerName)
			testMPIJobV2beta1InDifferentNamespace := buildMPIJobV2beta1("some-other-mpijob2-name2",
				TestNonProjectNamespace, "some-other-mpijob-uid2", constants.RunaiSchedulerName, constants.RunaiSchedulerName)
			testNotebookWithoutSchedulerName := buildNotebook("some-other-notebook-name",
				TestNamespace, "some-other-notebook-uid", "")
			testNotebookInDifferentNamespace := buildNotebook("some-other-notebook-name2",
				TestNonProjectNamespace, "some-other-notebook-uid2", constants.RunaiSchedulerName)
			testKubevirtVMIWithoutSchedulerName := buildKubevirtVMI("some-other-kubevirt-vmi-name",
				TestNamespace, "some-other-kubevirt-vmi-uid", "")
			testKubevirtVMIInDifferentNamespace := buildKubevirtVMI("some-other-kubevirt-vmi-name2",
				TestNonProjectNamespace, "some-other-kubevirt-vmi-uid2", constants.RunaiSchedulerName)
			testNvcfWithoutPods := buildNvcf("some-other-nvcf-name",
				TestNamespace, "some-other-nvcf-uid", constants.RunaiSchedulerName)

			// add objects here as you add more integrations with EventHandlerFilter
			objectsNotPassingFilter := []client.Object{
				testDeploymentWithoutSchedulerName, testDeploymentInDifferentNamespace,
				testPodWithoutSchedulerName, testPodInDifferentNamespace,
				testReplicaSetWithoutSchedulerName, testReplicaSetInDifferentNamespace,
				testStatefulSetWithoutSchedulerName, testStatefulSetInDifferentNamespace,
				testBatchJobWithoutSchedulerName, testBatchJobInDifferentNamespace,
				testCronJobWithoutSchedulerName, testCronJobInDifferentNamespace,
				testRayJobWithoutSchedulerName, testRayJobInDifferentNamespace,
				testRayClusterWithoutSchedulerName, testRayClusterInDifferentNamespace,
				testRayServiceWithoutSchedulerName, testRayServiceInDifferentNamespace,
				testRayJobV1WithoutSchedulerName, testRayJobV1InDifferentNamespace,
				testRayClusterV1WithoutSchedulerName, testRayClusterV1InDifferentNamespace,
				testRayServiceV1WithoutSchedulerName, testRayServiceV1InDifferentNamespace,
				testKNativeServiceWithoutSchedulerName, testKNativeServiceInDifferentNamespace,
				testTFJobWithoutSchedulerName, testTFJobInDifferentNamespace,
				testPytorchJobWithoutSchedulerName, testPytorchJobInDifferentNamespace,
				testXGBoostJobWithoutSchedulerName, testXGBoostJobInDifferentNamespace,
				testJAXJobWithoutSchedulerName, testJAXJobInDifferentNamespace,
				testMPIJobV1WithoutSchedulerName, testMPIJobV1InDifferentNamespace,
				testMPIJobV2beta1WithoutSchedulerName, testMPIJobV2beta1InDifferentNamespace,
				testNotebookWithoutSchedulerName, testNotebookInDifferentNamespace,
				testKubevirtVMIWithoutSchedulerName, testKubevirtVMIInDifferentNamespace,
				testNvcfWithoutPods,
			}
			onlyOwnedSupportedGVKsObjects := []client.Object{
				testKNativeRevision, testKNativeConfiguration,
			}
			allObjects := append(objectsNotPassingFilter, onlyOwnedSupportedGVKsObjects...)

			for _, obj := range allObjects {
				// create the object
				_ = createObject(ctx, obj, watchClient, &testCreatedObjects)
			}

			// create a sanity object
			sanityObjectThatExternalWorkloadShouldBeCreatedFor := buildPodInner("some-other-nameeee3",
				TestNamespace, "some-other-uidddddd3", constants.RunaiSchedulerName)
			createObject(ctx, sanityObjectThatExternalWorkloadShouldBeCreatedFor, watchClient, &testCreatedObjects)

			// validate en external workload was created for the sanity object
			eventuallyFindAndValidateExternalWorkload(fakeCachedClient,
				TestPodExpectedEw, sanityObjectThatExternalWorkloadShouldBeCreatedFor,
				"Test workload objects that don't pass FilterFunc", &testCreatedObjects)

			for _, objectNotPassingFilter := range objectsNotPassingFilter {
				Expect(externalWorkloadHandler.EventHandlerFilter(ctx, objectNotPassingFilter,
					objectNotPassingFilter.GetObjectKind().GroupVersionKind())).To(BeFalse())
			}

			for _, obj := range allObjects {
				externalWorkloadOfUnrecognizedWorkloadObject := findExternalWorkloadOfObject(fakeCachedClient,
					obj.GetUID(), obj.GetNamespace())
				Expect(externalWorkloadOfUnrecognizedWorkloadObject).To(BeNil(),
					"Test: <Test workload objects that don't pass FilterFunc>, "+
						"external workload for object with UID <%s> should not exist", obj.GetUID())
			}
		})

		It("Test objects in namespace with enforce runai scheduler", func() {
			// add the annotation to the namespace
			ns := &corev1.Namespace{}
			err := watchClient.Get(ctx, types.NamespacedName{Name: TestNamespace}, ns)
			Expect(err).ToNot(HaveOccurred())
			ns.Annotations = map[string]string{commonconsts.EnforceSchedulerAnnotationName: "true"}
			err = watchClient.Update(ctx, ns)
			Expect(err).ToNot(HaveOccurred())

			testDeploymentWithoutSchedulerName := buildDeployment("some-other-depl-name",
				TestNamespace, "some-other-depl-uid", "")
			testNotebookWithoutSchedulerName := buildNotebook("some-other-notebook-name",
				TestNamespace, "some-other-notebook-uid", "")
			testPodWithoutSchedulerName := buildPodInner("some-other-nameeee",
				TestNamespace, "some-other-uidddddd", "")

			createObject(ctx, testDeploymentWithoutSchedulerName, watchClient, &testCreatedObjects)
			createObject(ctx, testNotebookWithoutSchedulerName, watchClient, &testCreatedObjects)
			createObject(ctx, testPodWithoutSchedulerName, watchClient, &testCreatedObjects)

			// validate en external workload was created for the notebook and deployment
			validateExternalWorkloadExists(fakeCachedClient, testDeploymentWithoutSchedulerName.GetUID(),
				testDeploymentWithoutSchedulerName.GetNamespace(), "Deployment")
			validateExternalWorkloadExists(fakeCachedClient, testNotebookWithoutSchedulerName.GetUID(),
				testNotebookWithoutSchedulerName.GetNamespace(), "Notebook")
			// pod object still shouldn't have external-workload - it doesn't pass the filter func
			validateExternalWorkloadDoesNotExist(fakeCachedClient, testPodWithoutSchedulerName.GetUID(),
				testPodWithoutSchedulerName.GetNamespace(), "Pod")
		})
	})

	Describe("Test very long workload object name", func() {
		It("Pods with very long name", func() {
			// generate a max-length name for the pod
			podName0 := fmt.Sprintf("p-%s", rand.String(61))               // length 63
			podName1 := fmt.Sprintf("%s%s", podName0[:59], rand.String(4)) // length 63
			podName2 := podName0[:50]                                      // length 50
			testPodWithLongName0 := buildPod(podName0, TestNamespace, "uid1")
			testPodWithLongName1 := buildPod(podName1, TestNamespace, "uid2")
			testPodWithLongName2 := buildPod(podName2, TestNamespace, "uid3")
			testPods := []*corev1.Pod{testPodWithLongName0, testPodWithLongName1, testPodWithLongName2}

			for _, objectNotPassingFilter := range testPods {
				// create the object
				createObject(ctx, objectNotPassingFilter, watchClient, &testCreatedObjects)
			}

			ewOfPod0 := eventuallyFindAndValidateExternalWorkload(fakeCachedClient, nil, testPods[0],
				"Pods with very long name", &testCreatedObjects)
			ewOfPod1 := eventuallyFindAndValidateExternalWorkload(fakeCachedClient, nil, testPods[1],
				"Pods with very long name", &testCreatedObjects)
			ewOfPod2 := eventuallyFindAndValidateExternalWorkload(fakeCachedClient, nil, testPods[2],
				"Pods with very long name", &testCreatedObjects)

			Expect(ewOfPod0.GetName()[:58]).To(Equal(ewOfPod1.GetName()[:58]))
			Expect(ewOfPod2.GetName()[:50]).To(Equal(ewOfPod1.GetName()[:50]))
			Expect(ewOfPod0.GetName()).ToNot(Equal(ewOfPod1.GetName()))
			Expect(ewOfPod2.GetName()).ToNot(Equal(ewOfPod1.GetName()))
			Expect(ewOfPod0.GetName()).ToNot(Equal(ewOfPod2.GetName()))
		})
	})

	Describe("Test original workload update - external workload updated as well", func() {
		It("Deployment update", func() {
			deployment1 := TestDeployment.DeepCopy()
			expectedEW := testDeploymentExpectedEw.DeepCopy()

			obj := createObject(ctx, deployment1, watchClient, &testCreatedObjects)
			deployment1 = obj.(*appsv1.Deployment)
			eventuallyFindAndValidateExternalWorkload(fakeCachedClient,
				expectedEW, deployment1, "Deployment update - 0", &testCreatedObjects)

			// update some fields
			newNumOfReplicas := int64(testDeploymentNumberOfReplicas + 1)
			deployment1.Spec.Replicas = ptr.To(int32(newNumOfReplicas))
			deployment1.Spec.Template.Spec.PriorityClassName = testDeploymentPriority + "2"
			deployment1.Spec.Template.Spec.Containers[0].Image = "new-image"
			deployment1.Spec.Template.Annotations = map[string]string{}
			deployment1.Spec.Template.Spec.Containers[0].Resources = corev1.ResourceRequirements{
				Limits: corev1.ResourceList{
					corev1.ResourceCPU:        *resource.NewQuantity(testDeploymentCpuLimit+1, resource.DecimalSI),
					corev1.ResourceMemory:     *resource.NewQuantity(testDeploymentMemoryLimit+1, resource.DecimalSI),
					constants.GpuResourceName: *resource.NewQuantity(testDeploymentGpuLimit+1, resource.DecimalSI),
				},
				Requests: corev1.ResourceList{
					corev1.ResourceCPU:        *resource.NewQuantity(testDeploymentCpuRequest+1, resource.DecimalSI),
					corev1.ResourceMemory:     *resource.NewQuantity(testDeploymentMemoryRequest+1, resource.DecimalSI),
					constants.GpuResourceName: *resource.NewQuantity(testDeploymentGpuRequest+1, resource.DecimalSI),
				},
			}
			deployment1.Status.Conditions = []appsv1.DeploymentCondition{
				{
					Type:    appsv1.DeploymentProgressing,
					Status:  corev1.ConditionFalse,
					Reason:  "ProgressDeadlineExceeded",
					Message: "Deployment failed",
				},
				{
					Type:    appsv1.DeploymentAvailable,
					Status:  corev1.ConditionFalse,
					Reason:  "RandomReason",
					Message: "Random message",
				},
			}

			// update the expected external workload object
			expectedEW.Spec.RequestedPods.Replicas = ptr.To(int32(newNumOfReplicas))
			expectedEW.Spec.PriorityClassName = testDeploymentPriority + "2"
			expectedEW.Spec.ImageName = "new-image"
			expectedEW.Spec.RequestedResources = corev1.ResourceRequirements{
				Limits: corev1.ResourceList{
					corev1.ResourceCPU:        *resource.NewQuantity(newNumOfReplicas*(testDeploymentCpuLimit+1), resource.DecimalSI),
					corev1.ResourceMemory:     *resource.NewQuantity(newNumOfReplicas*(testDeploymentMemoryLimit+1), resource.DecimalSI),
					constants.GpuResourceName: *resource.NewMilliQuantity(1000*newNumOfReplicas*(testDeploymentGpuLimit+1), resource.DecimalSI),
				},
				Requests: corev1.ResourceList{
					corev1.ResourceCPU:        *resource.NewQuantity(newNumOfReplicas*(testDeploymentCpuRequest+1), resource.DecimalSI),
					corev1.ResourceMemory:     *resource.NewQuantity(newNumOfReplicas*(testDeploymentMemoryRequest+1), resource.DecimalSI),
					constants.GpuResourceName: *resource.NewQuantity(newNumOfReplicas*(testDeploymentGpuRequest+1), resource.DecimalSI),
				},
			}
			expectedEW.Status.ObservedStatus = v1alpha1.ExternalWorkloadObservedStatus{
				Phase:     v1alpha1.ExWDegradedPhase,
				K8sStatus: "Degraded",
				Message:   "Deployment failed",
				Conditions: []v1alpha1.ObservedCondition{
					{
						Type:    string(appsv1.DeploymentProgressing),
						Status:  corev1.ConditionFalse,
						Reason:  "ProgressDeadlineExceeded",
						Message: "Deployment failed",
					},
					{
						Type:    string(appsv1.DeploymentAvailable),
						Status:  corev1.ConditionFalse,
						Reason:  "RandomReason",
						Message: "Random message",
					},
				},
			}

			// update the deployment objects
			deplStatusCopy := deployment1.Status.DeepCopy()
			err := watchClient.Update(ctx, deployment1)
			Expect(err).To(BeNil())
			deployment1.Status = *deplStatusCopy
			err = watchClient.Status().Update(ctx, deployment1)
			Expect(err).To(BeNil())

			// validate the external workload was updated
			eventuallyFindAndValidateExternalWorkload(fakeCachedClient,
				expectedEW, deployment1, "Deployment update - 1", &testCreatedObjects)

			// now only update a condition that doesn't affect the calculated phase
			deployment1.Status.Conditions[1] = appsv1.DeploymentCondition{
				Type:    appsv1.DeploymentAvailable,
				Status:  corev1.ConditionTrue,
				Reason:  "RandomReason2",
				Message: "Random message 2",
			}
			expectedEW.Status.ObservedStatus.Conditions[1] = v1alpha1.ObservedCondition{
				Type:    string(appsv1.DeploymentAvailable),
				Status:  corev1.ConditionTrue,
				Reason:  "RandomReason2",
				Message: "Random message 2",
			}
			err = watchClient.Status().Update(ctx, deployment1)
			Expect(err).To(BeNil())

			// validate the external workload was updated
			eventuallyFindAndValidateExternalWorkload(fakeCachedClient,
				expectedEW, deployment1, "Deployment update - 2", &testCreatedObjects)

			// update only resources
			deployment1.Spec.Template.Spec.Containers[0].Resources = corev1.ResourceRequirements{
				Limits: corev1.ResourceList{
					corev1.ResourceCPU:        *resource.NewQuantity(testDeploymentCpuLimit+5, resource.DecimalSI),
					corev1.ResourceMemory:     *resource.NewQuantity(testDeploymentMemoryLimit+5, resource.DecimalSI),
					constants.GpuResourceName: *resource.NewQuantity(testDeploymentGpuLimit+5, resource.DecimalSI),
				},
				Requests: corev1.ResourceList{
					corev1.ResourceCPU:        *resource.NewQuantity(testDeploymentCpuRequest+10, resource.DecimalSI),
					corev1.ResourceMemory:     *resource.NewQuantity(testDeploymentMemoryRequest+10, resource.DecimalSI),
					constants.GpuResourceName: *resource.NewQuantity(testDeploymentGpuRequest+10, resource.DecimalSI),
				},
			}
			expectedEW.Spec.RequestedResources = corev1.ResourceRequirements{
				Limits: corev1.ResourceList{
					corev1.ResourceCPU:        *resource.NewQuantity(newNumOfReplicas*(testDeploymentCpuLimit+5), resource.DecimalSI),
					corev1.ResourceMemory:     *resource.NewQuantity(newNumOfReplicas*(testDeploymentMemoryLimit+5), resource.DecimalSI),
					constants.GpuResourceName: *resource.NewMilliQuantity(1000*newNumOfReplicas*(testDeploymentGpuLimit+5), resource.DecimalSI),
				},
				Requests: corev1.ResourceList{
					corev1.ResourceCPU:        *resource.NewQuantity(newNumOfReplicas*(testDeploymentCpuRequest+10), resource.DecimalSI),
					corev1.ResourceMemory:     *resource.NewQuantity(newNumOfReplicas*(testDeploymentMemoryRequest+10), resource.DecimalSI),
					constants.GpuResourceName: *resource.NewQuantity(newNumOfReplicas*(testDeploymentGpuRequest+10), resource.DecimalSI),
				},
			}
			err = watchClient.Update(ctx, deployment1)
			Expect(err).To(BeNil())

			// validate the external workload was updated
			eventuallyFindAndValidateExternalWorkload(fakeCachedClient,
				expectedEW, deployment1, "Deployment update - 3", &testCreatedObjects)
		})

		It("Knative status update", func() {
			knative1 := TestKnativeService.DeepCopy()
			expectedEW := TestKnativeServiceExpectedEw.DeepCopy()
			knative1.Status.Status.Conditions[0].Status = corev1.ConditionFalse
			knative1.Status.Status.Conditions[0].Message = testStatusMessage
			expectedEW.Status.ObservedStatus.Conditions[0].Status = corev1.ConditionFalse
			expectedEW.Status.ObservedStatus.Conditions[0].Message = testStatusMessage
			expectedEW.Status.ObservedStatus.K8sStatus = "NotReady"
			expectedEW.Status.ObservedStatus.Phase = v1alpha1.ExWPendingPhase
			expectedEW.Status.ObservedStatus.Message = testStatusMessage

			obj := createObject(ctx, knative1, watchClient, &testCreatedObjects)
			knative1 = obj.(*servingv1.Service)
			eventuallyFindAndValidateExternalWorkload(fakeCachedClient,
				expectedEW, knative1, "Knative update - 0", &testCreatedObjects)

			// update status to make it ready - validate message is empty now
			knative1.Status.Status.Conditions[0].Status = corev1.ConditionTrue
			knative1.Status.Status.Conditions[0].Message = ""
			expectedEW.Status.ObservedStatus.Conditions[0].Status = corev1.ConditionTrue
			expectedEW.Status.ObservedStatus.Conditions[0].Message = ""
			expectedEW.Status.ObservedStatus.K8sStatus = "Ready"
			expectedEW.Status.ObservedStatus.Phase = v1alpha1.ExWRunningPhase
			expectedEW.Status.ObservedStatus.Message = ""

			// update the knative object
			err := watchClient.Update(ctx, knative1)
			Expect(err).To(BeNil())

			// validate the external workload was updated
			eventuallyFindAndValidateExternalWorkload(fakeCachedClient,
				expectedEW, knative1, "Knative update - 1", &testCreatedObjects)
		})

		It("NIMService category and priority label update", func() {
			// Create NIM service with initial category label
			nimService := testNimService.DeepCopy()
			nimService.ObjectMeta.Labels = map[string]string{
				commonconsts.WorkloadCategoryLabel: "Deploy",
				pluginsConstants.PriorityLabelKey:  "medium",
			}
			expectedEW := TestNimServiceExpectedBasicEw.DeepCopy()
			expectedEW.ObjectMeta.Labels = map[string]string{
				commonconsts.WorkloadCategoryLabel: "Deploy",
				pluginsConstants.PriorityLabelKey:  "medium",
			}

			obj := createObject(ctx, nimService, watchClient, &testCreatedObjects)
			nimService = obj.(*nim.NIMService)
			eventuallyFindAndValidateExternalWorkload(fakeCachedClient,
				expectedEW, nimService, "NIMService category update - 0", &testCreatedObjects)

			// Update category and priority labels (simulating API update)
			nimService.ObjectMeta.Labels[commonconsts.WorkloadCategoryLabel] = "Train"
			nimService.ObjectMeta.Labels[pluginsConstants.PriorityLabelKey] = "high"
			expectedEW.ObjectMeta.Labels[commonconsts.WorkloadCategoryLabel] = "Train"
			expectedEW.ObjectMeta.Labels[pluginsConstants.PriorityLabelKey] = "high"

			// Update the NIM service object
			err := watchClient.Update(ctx, nimService)
			Expect(err).To(BeNil())

			// Validate the external workload was updated with new category and priority
			eventuallyFindAndValidateExternalWorkload(fakeCachedClient,
				expectedEW, nimService, "NIMService category update - 1", &testCreatedObjects)
		})

		It("Deployment category label update", func() {
			// Create deployment with initial category label
			deployment := TestDeployment.DeepCopy()
			deployment.ObjectMeta.Labels[commonconsts.WorkloadCategoryLabel] = "Deploy"
			deployment.ObjectMeta.Labels[pluginsConstants.PriorityLabelKey] = "low"
			expectedEW := testDeploymentExpectedEw.DeepCopy()
			expectedEW.ObjectMeta.Labels = map[string]string{
				commonconsts.WorkloadCategoryLabel: "Deploy",
				pluginsConstants.PriorityLabelKey:  "low",
			}
			// The handler uses the priorityClassName label to set Spec.PriorityClassName
			expectedEW.Spec.PriorityClassName = "low"

			obj := createObject(ctx, deployment, watchClient, &testCreatedObjects)
			deployment = obj.(*appsv1.Deployment)
			eventuallyFindAndValidateExternalWorkload(fakeCachedClient,
				expectedEW, deployment, "Deployment category update - 0", &testCreatedObjects)

			// Update category and priority labels
			deployment.ObjectMeta.Labels[commonconsts.WorkloadCategoryLabel] = "Train"
			deployment.ObjectMeta.Labels[pluginsConstants.PriorityLabelKey] = "high"
			expectedEW.ObjectMeta.Labels[commonconsts.WorkloadCategoryLabel] = "Train"
			expectedEW.ObjectMeta.Labels[pluginsConstants.PriorityLabelKey] = "high"
			expectedEW.Spec.PriorityClassName = "high"

			// Update the deployment object
			err := watchClient.Update(ctx, deployment)
			Expect(err).To(BeNil())

			// Validate the external workload was updated with new category and priority
			eventuallyFindAndValidateExternalWorkload(fakeCachedClient,
				expectedEW, deployment, "Deployment category update - 1", &testCreatedObjects)
		})
	})

	Describe("Test runai-workload owner: no externalworkload created + pod labeled with top owner workload id", func() {
		It("TrainingWorkload top owner", func() {
			// create the top owner runai training workload
			runaiWorkloadOwner := testRunaiWorkload.DeepCopy()
			createObject(ctx, runaiWorkloadOwner, watchClient, &testCreatedObjects)

			// create a pod with the runai workload as owner
			time.Sleep(30 * time.Millisecond)
			childPodWithRunaiWorkloadOwner := testChildPodOfRunaiWorkload.DeepCopy()
			createObject(ctx, childPodWithRunaiWorkloadOwner, watchClient, &testCreatedObjects)

			// validate en external workload was NOT created for the workload or the pod
			consistentlyValidateExternalWorkloadDoesNotExist(fakeCachedClient, runaiWorkloadOwner.GetUID(),
				runaiWorkloadOwner.GetNamespace(), "TrainingWorkload")
			consistentlyValidateExternalWorkloadDoesNotExist(fakeCachedClient, childPodWithRunaiWorkloadOwner.GetUID(),
				childPodWithRunaiWorkloadOwner.GetNamespace(), "Pod")

			eventuallyExpectPodLabeledWithTopOwnerWorkloadId(childPodWithRunaiWorkloadOwner,
				string(runaiWorkloadOwner.GetUID()), string(runaiWorkloadOwner.GetUID()), fakeCachedClient)
		})

		It("InteractiveWorkload with backend workloadId annotation top owner", func() {
			// create the top owner runai interactive workload
			runaiWorkloadOwner := testRunaiWorkload.DeepCopy()
			runaiWorkloadOwner.SetGroupVersionKind(schema.GroupVersionKind{
				Group:   "run.ai",
				Version: "v2alpha1",
				Kind:    "InteractiveWorkload",
			})
			backendWorkspaceId := "backend-workspace-id"
			runaiWorkloadOwner.SetAnnotations(map[string]string{
				commonconsts.ControlPlaneWorkloadIdAnnotationForWorkload: backendWorkspaceId,
			})
			createObject(ctx, runaiWorkloadOwner, watchClient, &testCreatedObjects)

			// create a pod with the runai workload as owner
			time.Sleep(30 * time.Millisecond)
			childPodWithRunaiWorkloadOwner := testChildPodOfRunaiWorkload.DeepCopy()
			childPodWithRunaiWorkloadOwner.OwnerReferences = []metav1.OwnerReference{
				{
					APIVersion:         runaiWorkloadOwner.GetAPIVersion(),
					BlockOwnerDeletion: ptr.To(true),
					Controller:         ptr.To(true),
					Kind:               runaiWorkloadOwner.GetKind(),
					Name:               runaiWorkloadOwner.GetName(),
					UID:                runaiWorkloadOwner.GetUID(),
				},
			}
			createObject(ctx, childPodWithRunaiWorkloadOwner, watchClient, &testCreatedObjects)

			// validate en external workload was NOT created for the workload or the pod
			consistentlyValidateExternalWorkloadDoesNotExist(fakeCachedClient, runaiWorkloadOwner.GetUID(),
				runaiWorkloadOwner.GetNamespace(), "InteractiveWorkload")
			consistentlyValidateExternalWorkloadDoesNotExist(fakeCachedClient, childPodWithRunaiWorkloadOwner.GetUID(),
				childPodWithRunaiWorkloadOwner.GetNamespace(), "Pod")

			eventuallyExpectPodLabeledWithTopOwnerWorkloadId(childPodWithRunaiWorkloadOwner,
				backendWorkspaceId, string(runaiWorkloadOwner.GetUID()), fakeCachedClient)
		})
	})

	Describe("Test unrecognized owner with AllowNonSupportedTypes=true", func() {
		It("Test unrecognized owner - use self as top owned object", func() {
			// create the top owner object with unrecognized kind
			unrecognizedOwnerWorkload := testUnrecognizedOwnerWorkload.DeepCopy()
			createObject(ctx, unrecognizedOwnerWorkload, watchClient, &testCreatedObjects)

			// create a pod with the unrecognized object as owner
			time.Sleep(30 * time.Millisecond)
			childPodWithUnrecognizedOwner := testChildPodWithUnrecognizedOwner.DeepCopy()
			createObject(ctx, childPodWithUnrecognizedOwner, watchClient, &testCreatedObjects)

			// validate en external workload was created for the pod
			eventuallyFindAndValidateExternalWorkload(fakeCachedClient,
				TestPodExpectedEw, childPodWithUnrecognizedOwner,
				"Test unrecognized owner - use self as top owned object", &testCreatedObjects)

			externalWorkloadOfUnrecognizedWorkloadObject := findExternalWorkloadOfObject(fakeCachedClient,
				unrecognizedOwnerWorkload.GetUID(), unrecognizedOwnerWorkload.GetNamespace())
			Expect(externalWorkloadOfUnrecognizedWorkloadObject).To(BeNil(),
				"Test: <Test unrecognized owner - use self as top owned object>, "+
					"external workload for unrecognized owner object with UID <%s> should not exist",
				unrecognizedOwnerWorkload.GetUID())
		})

		It("Test unrecognized owner - use last known top owner object", func() {
			deploymentWithUnrecognizedOwner := TestDeployment.DeepCopy()
			deploymentWithUnrecognizedOwner.SetOwnerReferences(unrecognizedWorkloadOwnerReferences)
			testChildPodWithUnrecognizedTopOwner := testChildPodWithUnrecognizedOwner.DeepCopy()
			testChildPodWithUnrecognizedTopOwner.SetOwnerReferences([]metav1.OwnerReference{
				{
					APIVersion:         deploymentWithUnrecognizedOwner.APIVersion,
					BlockOwnerDeletion: ptr.To(true),
					Controller:         ptr.To(true),
					Kind:               deploymentWithUnrecognizedOwner.Kind,
					Name:               deploymentWithUnrecognizedOwner.GetName(),
					UID:                deploymentWithUnrecognizedOwner.GetUID(),
				},
			})
			unrecognizedOwnerWorkload := testUnrecognizedOwnerWorkload.DeepCopy()

			// create the top owner object with unrecognized kind
			createObject(ctx, unrecognizedOwnerWorkload, watchClient, &testCreatedObjects)

			// create a deployment with the unrecognized object as owner + pod with the deployment as owner
			time.Sleep(30 * time.Millisecond)
			createObject(ctx, deploymentWithUnrecognizedOwner, watchClient, &testCreatedObjects)
			time.Sleep(30 * time.Millisecond)
			createObject(ctx, testChildPodWithUnrecognizedTopOwner, watchClient, &testCreatedObjects)

			// validate en external workload was created for the deployment
			eventuallyFindAndValidateExternalWorkloadWithChildIds(fakeCachedClient,
				testDeploymentExpectedEw, deploymentWithUnrecognizedOwner, []v1alpha1.ChildId{},
				"Test unrecognized owner - use last known top owner object", &testCreatedObjects)

			externalWorkloadOfUnrecognizedWorkloadObject := findExternalWorkloadOfObject(fakeCachedClient,
				unrecognizedOwnerWorkload.GetUID(), unrecognizedOwnerWorkload.GetNamespace())
			Expect(externalWorkloadOfUnrecognizedWorkloadObject).To(BeNil(),
				"Test: <Test unrecognized owner - use last known top owner object>, "+
					"external workload for unrecognized owner object with UID <%s> should not exist",
				unrecognizedOwnerWorkload.GetUID())
		})
	})

	Describe("Test unrecognized owner with AllowNonSupportedTypes=false", func() {
		It("Test unrecognized owner - error", func() {
			externalWorkloadHandler.SetAllowNonSupportedTypes(false)
			unrecognizedOwnerWorkload := testUnrecognizedOwnerWorkload.DeepCopy()
			createObject(ctx, unrecognizedOwnerWorkload, watchClient, &testCreatedObjects)

			// create a pod with the unrecognized object as owner
			time.Sleep(30 * time.Millisecond)
			childPodWithUnrecognizedOwner := testChildPodWithUnrecognizedOwner.DeepCopy()
			createObject(ctx, childPodWithUnrecognizedOwner, watchClient, &testCreatedObjects)

			// create a regular pod without owners just for sanity
			time.Sleep(30 * time.Millisecond)
			pod := TestPod.DeepCopy()
			pod.UID = "different-uid"
			pod.Name = "different-name"
			createObject(ctx, pod, watchClient, &testCreatedObjects)

			// validate en external workload was created for that pod without owners
			eventuallyFindAndValidateExternalWorkload(fakeCachedClient,
				TestPodExpectedEw, pod,
				"Test unrecognized owner - error", &testCreatedObjects)

			// validate en external workload was not created for the pod with unrecognized owner
			externalWorkloadOfPodWithUnrecognizedOwner := findExternalWorkloadOfObject(fakeCachedClient,
				childPodWithUnrecognizedOwner.GetUID(), childPodWithUnrecognizedOwner.GetNamespace())
			Expect(externalWorkloadOfPodWithUnrecognizedOwner).To(BeNil(),
				"Test: <Test unrecognized owner - error>, "+
					"external workload for pod with unrecognized owner with UID <%s> should not exist",
				unrecognizedOwnerWorkload.GetUID())

			// validate en external workload was not created for the unrecognized object
			externalWorkloadOfUnrecognizedWorkloadObject := findExternalWorkloadOfObject(fakeCachedClient,
				unrecognizedOwnerWorkload.GetUID(), unrecognizedOwnerWorkload.GetNamespace())
			Expect(externalWorkloadOfUnrecognizedWorkloadObject).To(BeNil(),
				"Test: <Test unrecognized owner - error>, "+
					"external workload for unrecognized owner object with UID <%s> should not exist",
				unrecognizedOwnerWorkload.GetUID())
		})
	})

	// Test: When top owner has no handler but an intermediate owner does, use the highest owner with a handler.
	// Scenario: Pod -> Deployment (has handler) -> DaemonSet (no handler)
	// Expected: ExternalWorkload is created for the Deployment (highest owner with handler), not for Pod or DaemonSet.
	Describe("Test readable owner without handler - use highest owner with handler", func() {
		It("Pod -> Deployment -> DaemonSet: EW should be created for Deployment (highest with handler)", func() {
			// DaemonSet is in the scheme (apps/v1) but does NOT have a handler in supportedGVKsHandlers
			// (and it shouldn't have a handler - it is used here only for tests purposes).
			// The EW should be created for the Deployment, not the Pod or DaemonSet.

			// 1. Create DaemonSet (top owner, no handler)
			daemonSet := buildDaemonSet("test-daemonset", TestNamespace, "daemonset-uid-123", constants.RunaiSchedulerName)
			createObject(ctx, daemonSet, watchClient, &testCreatedObjects)

			// 2. Create Deployment owned by DaemonSet (has handler)
			time.Sleep(30 * time.Millisecond)
			deployment := buildDeploymentWithOwner("deployment-under-daemonset", TestNamespace, "deployment-uid-456",
				constants.RunaiSchedulerName, daemonSet)
			createObject(ctx, deployment, watchClient, &testCreatedObjects)

			// 3. Create Pod owned by Deployment
			time.Sleep(30 * time.Millisecond)
			podUnderDeployment := buildPod("pod-under-deployment", TestNamespace, "pod-uid-789")
			podUnderDeployment.OwnerReferences = []metav1.OwnerReference{
				{
					APIVersion:         deployment.APIVersion,
					BlockOwnerDeletion: ptr.To(true),
					Controller:         ptr.To(true),
					Kind:               deployment.Kind,
					Name:               deployment.GetName(),
					UID:                deployment.GetUID(),
				},
			}
			createObject(ctx, podUnderDeployment, watchClient, &testCreatedObjects)

			// Validate: external workload IS created for Deployment (highest owner with handler)
			eventuallyFindAndValidateExternalWorkload(fakeCachedClient,
				TestDeploymentExpectedEw, deployment,
				"Deployment (highest owner with handler) should get the EW", &testCreatedObjects)

			// Validate: no external workload for DaemonSet (no handler)
			externalWorkloadOfDaemonSet := findExternalWorkloadOfObject(fakeCachedClient,
				daemonSet.GetUID(), daemonSet.GetNamespace())
			Expect(externalWorkloadOfDaemonSet).To(BeNil(),
				"external workload for DaemonSet should not exist (no handler)")

			// Validate: no external workload for Pod (it has an owner with handler)
			externalWorkloadOfPod := findExternalWorkloadOfObject(fakeCachedClient,
				podUnderDeployment.GetUID(), podUnderDeployment.GetNamespace())
			Expect(externalWorkloadOfPod).To(BeNil(),
				"external workload for Pod should not exist (Deployment should be used)")
		})
	})

	Describe("Test multiple workload objects that a single external workload should be created for them", func() {
		var createdNvcfObjects []client.Object
		var createdPodObjects []client.Object

		It("Creates a single external workload with two replicas", func() {
			// create two (nvcf, pod) tuples and validate that:
			// 1. only a single external workload is created for both of them
			// 2. the external workload has both nvcf instances as owners
			var testCreatedObjects []client.Object
			for i := 0; i < 2; i++ {
				testNvcf := TestNvcfFunction.DeepCopy()
				testNvcf.Name = fmt.Sprintf("%s-%d", testNvcf.Name, i)
				testNvcf.UID = types.UID(fmt.Sprintf("%s-%d", testNvcf.UID, i))
				createdNvcfObjects = append(createdNvcfObjects, createObject(ctx, testNvcf, watchClient, &testCreatedObjects))

				testPod := TestNvcfFunctionChildPod.DeepCopy()
				testPod.Name = fmt.Sprintf("%s-%d", testPod.Name, i)
				testPod.UID = types.UID(fmt.Sprintf("%s-%d", testPod.UID, i))
				testPod.SetOwnerReferences([]metav1.OwnerReference{
					{
						APIVersion: testNvcf.APIVersion,
						Kind:       testNvcf.Kind,
						Name:       createdNvcfObjects[i].GetName(),
						UID:        createdNvcfObjects[i].GetUID(),
					},
				})

				createdPodObjects = append(createdPodObjects, createObject(ctx, testPod, watchClient, &testCreatedObjects))
				time.Sleep(30 * time.Millisecond)
			}

			// update the original object to cause an update event and hence another reconciliation
			updateObjectWithRandomLabel(ctx, createdNvcfObjects[0], watchClient)

			// validate external workload
			Eventually(func(g Gomega) {
				externalWorkloads := findAllExternalWorkloadOfObject(watchClient, testNvcfFunctionExpectedEw.Spec.OriginalWorkloadUID, createdNvcfObjects[0].GetNamespace())
				g.Expect(externalWorkloads).ToNot(BeNil())
				g.Expect(externalWorkloads).To(HaveLen(1))
				externalWorkloadOfObject := &externalWorkloads[0]
				g.Expect(externalWorkloadOfObject.OwnerReferences).To(HaveLen(2))
				g.Expect(externalWorkloadOfObject.OwnerReferences[0].UID).To(Equal(types.UID(fmt.Sprintf("%s-0", TestNvcfFunction.UID))))
				g.Expect(externalWorkloadOfObject.OwnerReferences[1].UID).To(Equal(types.UID(fmt.Sprintf("%s-1", TestNvcfFunction.UID))))
				g.Expect(externalWorkloadOfObject.Spec.OriginalWorkloadUID).To(Equal(testNvcfFunctionExpectedEw.Spec.OriginalWorkloadUID))
				g.Expect(externalWorkloadOfObject.Spec.RequestedPods.Replicas).To(Equal(ptr.To(int32(2))))
			}, timeout, interval).Should(Succeed())
		})

		It("Updates the owner references of the external workload when deleting one of the instances", func() {
			err := watchClient.Delete(ctx, createdPodObjects[1])
			Expect(err).To(BeNil())

			Eventually(func(g Gomega) {
				externalWorkloads := findAllExternalWorkloadOfObject(watchClient, testNvcfFunctionExpectedEw.Spec.OriginalWorkloadUID, createdNvcfObjects[0].GetNamespace())
				g.Expect(externalWorkloads).ToNot(BeNil())
				g.Expect(externalWorkloads).To(HaveLen(1))
				externalWorkloadOfObject := &externalWorkloads[0]
				g.Expect(externalWorkloadOfObject.OwnerReferences).To(HaveLen(1))
			}, timeout, interval).Should(Succeed())
		})
	})

	Describe("Test StorageRefs rebuild on reconcile", func() {
		It("rebuilds StorageRefs from the current pod template and drops refs removed from the source", func() {
			deployment := buildDeployment("depl-storage-refs", TestNamespace, "depl-storage-refs-uid",
				constants.RunaiSchedulerName)
			deployment.Spec.Template.Spec.Volumes = []corev1.Volume{
				{Name: "data", VolumeSource: corev1.VolumeSource{
					PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{ClaimName: "pvc-data"}}},
				{Name: "creds", VolumeSource: corev1.VolumeSource{
					Secret: &corev1.SecretVolumeSource{SecretName: "secret-creds"}}},
			}

			// The EWI resolves each ref to its UID, so the referenced resources must exist.
			createObject(ctx, &corev1.PersistentVolumeClaim{ObjectMeta: metav1.ObjectMeta{
				Name: "pvc-data", Namespace: TestNamespace, UID: "pvc-data-uid"}}, watchClient, &testCreatedObjects)
			createObject(ctx, &corev1.Secret{ObjectMeta: metav1.ObjectMeta{
				Name: "secret-creds", Namespace: TestNamespace, UID: "secret-creds-uid"}}, watchClient, &testCreatedObjects)

			obj := createObject(ctx, deployment, watchClient, &testCreatedObjects)
			deployment = obj.(*appsv1.Deployment)

			// Both refs are present after the initial reconcile.
			Eventually(func(g Gomega) {
				ew := findExternalWorkloadOfObject(fakeCachedClient, deployment.GetUID(), deployment.GetNamespace())
				g.Expect(ew).ToNot(BeNil())
				g.Expect(ew.Spec.StorageRefs).To(ConsistOf(
					v1alpha1.StorageRef{Kind: v1alpha1.StorageRefKindPVC, Name: "pvc-data", UID: "pvc-data-uid"},
					v1alpha1.StorageRef{Kind: v1alpha1.StorageRefKindSecret, Name: "secret-creds", UID: "secret-creds-uid"},
				))
			}, timeout, interval).Should(Succeed())

			// Remove the secret volume from the source and re-reconcile.
			err := watchClient.Get(ctx, types.NamespacedName{Name: deployment.Name, Namespace: deployment.Namespace}, deployment)
			Expect(err).ToNot(HaveOccurred())
			deployment.GetObjectKind().SetGroupVersionKind(appsv1.SchemeGroupVersion.WithKind("Deployment"))
			deployment.Spec.Template.Spec.Volumes = []corev1.Volume{
				{Name: "data", VolumeSource: corev1.VolumeSource{
					PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{ClaimName: "pvc-data"}}},
			}
			err = watchClient.Update(ctx, deployment)
			Expect(err).ToNot(HaveOccurred())

			// The removed secret ref drops out; only the PVC remains.
			Eventually(func(g Gomega) {
				ew := findExternalWorkloadOfObject(fakeCachedClient, deployment.GetUID(), deployment.GetNamespace())
				g.Expect(ew).ToNot(BeNil())
				g.Expect(ew.Spec.StorageRefs).To(ConsistOf(
					v1alpha1.StorageRef{Kind: v1alpha1.StorageRefKindPVC, Name: "pvc-data", UID: "pvc-data-uid"},
				))
			}, timeout, interval).Should(Succeed())
		})
	})

	Describe("Test network_status end-to-end", func() {
		It("network_status updated on external-workload on pod add, not updated on pod update", func() {
			// Create a Node so the NetworkStatusUpdater can resolve NodePort endpoint addresses
			node := &corev1.Node{
				ObjectMeta: metav1.ObjectMeta{Name: "test-node-net-status"},
				Status: corev1.NodeStatus{
					Addresses: []corev1.NodeAddress{
						{Type: corev1.NodeInternalIP, Address: "10.0.0.99"},
					},
				},
			}
			createObject(ctx, node, watchClient, &testCreatedObjects)

			// Create a NodePort Service that selects pods with label app=net-status-test
			svc := &corev1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "svc-net-status",
					Namespace: TestNamespace,
				},
				Spec: corev1.ServiceSpec{
					Type:     corev1.ServiceTypeNodePort,
					Selector: map[string]string{"app": "net-status-test"},
					Ports: []corev1.ServicePort{
						{Port: 80, NodePort: 30080, Protocol: corev1.ProtocolTCP},
					},
				},
			}
			createObject(ctx, svc, watchClient, &testCreatedObjects)

			// Create a Deployment as the top owner - its EW will receive the network_status
			deployment := buildDeployment("depl-net-status", TestNamespace, "depl-net-status-uid",
				constants.RunaiSchedulerName)
			obj := createObject(ctx, deployment, watchClient, &testCreatedObjects)
			deployment = obj.(*appsv1.Deployment)

			// Wait for the deployment's EW to be created
			eventuallyFindAndValidateExternalWorkload(fakeCachedClient,
				nil, deployment, "network_status e2e - deployment EW created", &testCreatedObjects)

			// Create a child Pod (POD ADD) with labels matching the service and NodeName for IP resolution.
			// Network status should be updated on the deployment's EW when this pod is added.
			time.Sleep(30 * time.Millisecond)
			pod := buildPod("pod-net-status", TestNamespace, "pod-net-status-uid")
			pod.Labels = map[string]string{"app": "net-status-test"}
			pod.Spec.NodeName = "test-node-net-status"
			pod.OwnerReferences = []metav1.OwnerReference{
				{
					APIVersion: deployment.APIVersion,
					Kind:       deployment.Kind,
					Name:       deployment.GetName(),
					UID:        deployment.GetUID(),
				},
			}
			createObject(ctx, pod, watchClient, &testCreatedObjects)

			// Verify the deployment's EW has network_status updated with the NodePort endpoint
			Eventually(func(g Gomega) {
				ew := findExternalWorkloadOfObject(fakeCachedClient,
					deployment.GetUID(), deployment.GetNamespace())
				g.Expect(ew).ToNot(BeNil(), "should find external workload for deployment")
				g.Expect(ew.Status.NetworkStatus.ExternalEndpoints).To(HaveLen(1),
					"network_status should have 1 endpoint after pod add")
				ep := ew.Status.NetworkStatus.ExternalEndpoints[0]
				g.Expect(ep.SourceType).To(Equal(v1alpha1.ExternalEndpointNodePortSource))
				g.Expect(ep.Host).To(Equal("10.0.0.99"))
				g.Expect(ep.Port).To(Equal("30080"))
				g.Expect(ep.Protocol).To(Equal("TCP"))
				g.Expect(ep.Url).To(Equal("10.0.0.99:30080"))
				g.Expect(ep.SourceRef).To(Equal(v1alpha1.ExternalEndpointSourceRef{
					Kind: "Service", Name: "svc-net-status",
				}))
			}, timeout, interval).Should(Succeed())

			// Re-fetch the pod (handler may have patched it with labels, changing ResourceVersion)
			err := watchClient.Get(ctx, types.NamespacedName{Name: pod.Name, Namespace: pod.Namespace}, pod)
			Expect(err).ToNot(HaveOccurred())
			pod.GetObjectKind().SetGroupVersionKind(corev1.SchemeGroupVersion.WithKind("Pod"))

			// Update the pod (POD UPDATE event, not ADD).
			// Network status should NOT be re-evaluated on pod update.
			updateObjectWithRandomLabel(ctx, pod, watchClient)

			// Verify network_status was NOT updated on pod update - still 1 endpoint.
			Consistently(func(g Gomega) {
				ew := findExternalWorkloadOfObject(fakeCachedClient,
					deployment.GetUID(), deployment.GetNamespace())
				g.Expect(ew).ToNot(BeNil(), "should find external workload for deployment")
				g.Expect(ew.Status.NetworkStatus.ExternalEndpoints).To(HaveLen(1),
					"network_status should NOT be updated on pod update - still 1 endpoint")
			}, consistentlyTimeout, interval).Should(Succeed())
		})
	})

	Describe("Test network resource watcher e2e", func() {
		// These tests verify that when network resources (Service, Ingress, Routes) are
		// created AFTER a pod already exists, the ExternalWorkload endpoints are updated
		// via the network resource watchers. Also tests deletion removes endpoints.

		// Helper: create a Deployment + Pod that already exist, wait for the EW to be created.
		// Returns the deployment (with UID set by fake client) and the pod.
		setupDeploymentAndPod := func(suffix string) (*appsv1.Deployment, *corev1.Pod) {
			node := &corev1.Node{
				ObjectMeta: metav1.ObjectMeta{Name: "node-nrw-" + suffix},
				Status: corev1.NodeStatus{
					Addresses: []corev1.NodeAddress{
						{Type: corev1.NodeInternalIP, Address: "10.0.1.1"},
					},
				},
			}
			createObject(ctx, node, watchClient, &testCreatedObjects)

			deployment := buildDeployment("depl-nrw-"+suffix, TestNamespace, "depl-nrw-"+suffix+"-uid",
				constants.RunaiSchedulerName)
			obj := createObject(ctx, deployment, watchClient, &testCreatedObjects)
			deployment = obj.(*appsv1.Deployment)

			eventuallyFindAndValidateExternalWorkload(fakeCachedClient,
				nil, deployment, "nrw-"+suffix+" - EW created", &testCreatedObjects)

			time.Sleep(30 * time.Millisecond)
			pod := buildPod("pod-nrw-"+suffix, TestNamespace, "pod-nrw-"+suffix+"-uid")
			pod.Labels = map[string]string{"app": "nrw-" + suffix}
			pod.Spec.NodeName = "node-nrw-" + suffix
			pod.OwnerReferences = []metav1.OwnerReference{
				{
					APIVersion: deployment.APIVersion,
					Kind:       deployment.Kind,
					Name:       deployment.GetName(),
					UID:        deployment.GetUID(),
				},
			}
			createObject(ctx, pod, watchClient, &testCreatedObjects)

			// Wait for pod to be processed (pod-ADD sets labels etc.)
			time.Sleep(100 * time.Millisecond)

			return deployment, pod
		}

		It("NodePort service add updates endpoints, delete removes them", func() {
			deployment, _ := setupDeploymentAndPod("np")

			// Create a NodePort service AFTER the pod exists
			svc := &corev1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "svc-nrw-np",
					Namespace: TestNamespace,
				},
				Spec: corev1.ServiceSpec{
					Type:     corev1.ServiceTypeNodePort,
					Selector: map[string]string{"app": "nrw-np"},
					Ports: []corev1.ServicePort{
						{Port: 80, NodePort: 30090, Protocol: corev1.ProtocolTCP},
					},
				},
			}
			createObject(ctx, svc, watchClient, &testCreatedObjects)

			// Verify the EW gets the NodePort endpoint via the network resource watcher
			Eventually(func(g Gomega) {
				ew := findExternalWorkloadOfObject(fakeCachedClient,
					deployment.GetUID(), deployment.GetNamespace())
				g.Expect(ew).ToNot(BeNil())
				g.Expect(ew.Status.NetworkStatus.ExternalEndpoints).To(HaveLen(1))
				ep := ew.Status.NetworkStatus.ExternalEndpoints[0]
				g.Expect(ep.SourceType).To(Equal(v1alpha1.ExternalEndpointNodePortSource))
				g.Expect(ep.Host).To(Equal("10.0.1.1"))
				g.Expect(ep.Port).To(Equal("30090"))
				g.Expect(ep.SourceRef.Name).To(Equal("svc-nrw-np"))
			}, timeout, interval).Should(Succeed())

			// Delete the service
			err := watchClient.Delete(ctx, svc)
			Expect(err).ToNot(HaveOccurred())

			// Verify the endpoints are removed
			Eventually(func(g Gomega) {
				ew := findExternalWorkloadOfObject(fakeCachedClient,
					deployment.GetUID(), deployment.GetNamespace())
				g.Expect(ew).ToNot(BeNil())
				g.Expect(ew.Status.NetworkStatus.ExternalEndpoints).To(BeEmpty(),
					"endpoints should be removed after service deletion")
			}, timeout, interval).Should(Succeed())
		})

		It("LoadBalancer service add updates endpoints, delete removes them", func() {
			deployment, _ := setupDeploymentAndPod("lb")

			svc := &corev1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "svc-nrw-lb",
					Namespace: TestNamespace,
				},
				Spec: corev1.ServiceSpec{
					Type:     corev1.ServiceTypeLoadBalancer,
					Selector: map[string]string{"app": "nrw-lb"},
					Ports: []corev1.ServicePort{
						{Port: 443, Protocol: corev1.ProtocolTCP},
					},
				},
				Status: corev1.ServiceStatus{
					LoadBalancer: corev1.LoadBalancerStatus{
						Ingress: []corev1.LoadBalancerIngress{
							{IP: "203.0.113.10"},
						},
					},
				},
			}
			createObject(ctx, svc, watchClient, &testCreatedObjects)

			Eventually(func(g Gomega) {
				ew := findExternalWorkloadOfObject(fakeCachedClient,
					deployment.GetUID(), deployment.GetNamespace())
				g.Expect(ew).ToNot(BeNil())
				g.Expect(ew.Status.NetworkStatus.ExternalEndpoints).To(HaveLen(1))
				ep := ew.Status.NetworkStatus.ExternalEndpoints[0]
				g.Expect(ep.SourceType).To(Equal(v1alpha1.ExternalEndpointLoadBalanceSource))
				g.Expect(ep.Host).To(Equal("203.0.113.10"))
				g.Expect(ep.Port).To(Equal("443"))
				g.Expect(ep.SourceRef.Name).To(Equal("svc-nrw-lb"))
			}, timeout, interval).Should(Succeed())

			err := watchClient.Delete(ctx, svc)
			Expect(err).ToNot(HaveOccurred())

			Eventually(func(g Gomega) {
				ew := findExternalWorkloadOfObject(fakeCachedClient,
					deployment.GetUID(), deployment.GetNamespace())
				g.Expect(ew).ToNot(BeNil())
				g.Expect(ew.Status.NetworkStatus.ExternalEndpoints).To(BeEmpty(),
					"endpoints should be removed after LB service deletion")
			}, timeout, interval).Should(Succeed())
		})

		It("Ingress add updates endpoints, delete removes them", func() {
			deployment, _ := setupDeploymentAndPod("ing")

			os.Setenv(wlconfig.EnvVarClusterDomain, "https://app.example.com:443")
			defer os.Unsetenv(wlconfig.EnvVarClusterDomain)

			// Create a ClusterIP service (ingress routes to it)
			svc := &corev1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "svc-nrw-ing",
					Namespace: TestNamespace,
				},
				Spec: corev1.ServiceSpec{
					Type:     corev1.ServiceTypeClusterIP,
					Selector: map[string]string{"app": "nrw-ing"},
					Ports: []corev1.ServicePort{
						{Port: 80, Protocol: corev1.ProtocolTCP},
					},
				},
			}
			createObject(ctx, svc, watchClient, &testCreatedObjects)

			// Create an Ingress pointing to the service
			pathType := networkingv1.PathTypePrefix
			ing := &networkingv1.Ingress{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "ing-nrw",
					Namespace: TestNamespace,
				},
				Spec: networkingv1.IngressSpec{
					Rules: []networkingv1.IngressRule{
						{
							Host: "myapp.example.com",
							IngressRuleValue: networkingv1.IngressRuleValue{
								HTTP: &networkingv1.HTTPIngressRuleValue{
									Paths: []networkingv1.HTTPIngressPath{
										{
											Path:     "/api",
											PathType: &pathType,
											Backend: networkingv1.IngressBackend{
												Service: &networkingv1.IngressServiceBackend{
													Name: "svc-nrw-ing",
													Port: networkingv1.ServiceBackendPort{Number: 80},
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
			createObject(ctx, ing, watchClient, &testCreatedObjects)

			Eventually(func(g Gomega) {
				ew := findExternalWorkloadOfObject(fakeCachedClient,
					deployment.GetUID(), deployment.GetNamespace())
				g.Expect(ew).ToNot(BeNil())
				g.Expect(ew.Status.NetworkStatus.ExternalEndpoints).To(HaveLen(1))
				ep := ew.Status.NetworkStatus.ExternalEndpoints[0]
				g.Expect(ep.SourceType).To(Equal(v1alpha1.ExternalEndpointIngressSource))
				g.Expect(ep.Host).To(Equal("myapp.example.com"))
				g.Expect(ep.Path).To(Equal("/api"))
				g.Expect(ep.SourceRef.Kind).To(Equal("Ingress"))
				g.Expect(ep.SourceRef.Name).To(Equal("ing-nrw"))
			}, timeout, interval).Should(Succeed())

			err := watchClient.Delete(ctx, ing)
			Expect(err).ToNot(HaveOccurred())

			Eventually(func(g Gomega) {
				ew := findExternalWorkloadOfObject(fakeCachedClient,
					deployment.GetUID(), deployment.GetNamespace())
				g.Expect(ew).ToNot(BeNil())
				g.Expect(ew.Status.NetworkStatus.ExternalEndpoints).To(BeEmpty(),
					"endpoints should be removed after ingress deletion")
			}, timeout, interval).Should(Succeed())
		})

		It("HTTPRoute add updates endpoints, delete removes them", func() {
			deployment, _ := setupDeploymentAndPod("http")

			os.Setenv(wlconfig.EnvVarClusterDomain, "https://app.example.com:443")
			defer os.Unsetenv(wlconfig.EnvVarClusterDomain)

			svc := &corev1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "svc-nrw-http",
					Namespace: TestNamespace,
				},
				Spec: corev1.ServiceSpec{
					Type:     corev1.ServiceTypeClusterIP,
					Selector: map[string]string{"app": "nrw-http"},
					Ports: []corev1.ServicePort{
						{Port: 80, Protocol: corev1.ProtocolTCP},
					},
				},
			}
			createObject(ctx, svc, watchClient, &testCreatedObjects)

			// Create Gateway that the HTTPRoute references
			gw := &gwv1.Gateway{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "gw-nrw-http",
					Namespace: TestNamespace,
				},
				Spec: gwv1.GatewaySpec{
					GatewayClassName: "test-class",
					Listeners: []gwv1.Listener{
						{
							Name:     "http",
							Port:     8080,
							Protocol: gwv1.HTTPProtocolType,
						},
					},
				},
			}
			createObject(ctx, gw, watchClient, &testCreatedObjects)

			gwNamespace := gwv1.Namespace(TestNamespace)
			route := &gwv1.HTTPRoute{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "httproute-nrw",
					Namespace: TestNamespace,
				},
				Spec: gwv1.HTTPRouteSpec{
					Hostnames: []gwv1.Hostname{"http.example.com"},
					CommonRouteSpec: gwv1.CommonRouteSpec{
						ParentRefs: []gwv1.ParentReference{
							{Name: "gw-nrw-http", Namespace: &gwNamespace},
						},
					},
					Rules: []gwv1.HTTPRouteRule{
						{
							BackendRefs: []gwv1.HTTPBackendRef{
								{BackendRef: gwv1.BackendRef{BackendObjectReference: gwv1.BackendObjectReference{
									Name: "svc-nrw-http",
								}}},
							},
						},
					},
				},
				Status: gwv1.HTTPRouteStatus{
					RouteStatus: gwv1.RouteStatus{
						Parents: []gwv1.RouteParentStatus{
							{
								ParentRef: gwv1.ParentReference{Name: "gw-nrw-http", Namespace: &gwNamespace},
								Conditions: []metav1.Condition{
									{Type: string(gwv1.RouteConditionAccepted), Status: metav1.ConditionTrue},
								},
							},
						},
					},
				},
			}
			createObject(ctx, route, watchClient, &testCreatedObjects)

			Eventually(func(g Gomega) {
				ew := findExternalWorkloadOfObject(fakeCachedClient,
					deployment.GetUID(), deployment.GetNamespace())
				g.Expect(ew).ToNot(BeNil())
				g.Expect(ew.Status.NetworkStatus.ExternalEndpoints).To(HaveLen(1))
				ep := ew.Status.NetworkStatus.ExternalEndpoints[0]
				g.Expect(ep.SourceType).To(Equal(v1alpha1.ExternalEndpointGatewayAPISource))
				g.Expect(ep.Host).To(Equal("http.example.com"))
				g.Expect(ep.SourceRef.Kind).To(Equal("HTTPRoute"))
				g.Expect(ep.SourceRef.Name).To(Equal("httproute-nrw"))
			}, timeout, interval).Should(Succeed())

			err := watchClient.Delete(ctx, route)
			Expect(err).ToNot(HaveOccurred())

			Eventually(func(g Gomega) {
				ew := findExternalWorkloadOfObject(fakeCachedClient,
					deployment.GetUID(), deployment.GetNamespace())
				g.Expect(ew).ToNot(BeNil())
				g.Expect(ew.Status.NetworkStatus.ExternalEndpoints).To(BeEmpty(),
					"endpoints should be removed after HTTPRoute deletion")
			}, timeout, interval).Should(Succeed())
		})

		It("GRPCRoute add updates endpoints, delete removes them", func() {
			deployment, _ := setupDeploymentAndPod("grpc")

			os.Setenv(wlconfig.EnvVarClusterDomain, "https://app.example.com:443")
			defer os.Unsetenv(wlconfig.EnvVarClusterDomain)

			svc := &corev1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "svc-nrw-grpc",
					Namespace: TestNamespace,
				},
				Spec: corev1.ServiceSpec{
					Type:     corev1.ServiceTypeClusterIP,
					Selector: map[string]string{"app": "nrw-grpc"},
					Ports: []corev1.ServicePort{
						{Port: 50051, Protocol: corev1.ProtocolTCP},
					},
				},
			}
			createObject(ctx, svc, watchClient, &testCreatedObjects)

			gw := &gwv1.Gateway{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "gw-nrw-grpc",
					Namespace: TestNamespace,
				},
				Spec: gwv1.GatewaySpec{
					GatewayClassName: "test-class",
					Listeners: []gwv1.Listener{
						{
							Name:     "http",
							Port:     8080,
							Protocol: gwv1.HTTPProtocolType,
						},
					},
				},
			}
			createObject(ctx, gw, watchClient, &testCreatedObjects)

			gwNamespace := gwv1.Namespace(TestNamespace)
			route := &gwv1.GRPCRoute{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "grpcroute-nrw",
					Namespace: TestNamespace,
				},
				Spec: gwv1.GRPCRouteSpec{
					Hostnames: []gwv1.Hostname{"grpc.example.com"},
					CommonRouteSpec: gwv1.CommonRouteSpec{
						ParentRefs: []gwv1.ParentReference{
							{Name: "gw-nrw-grpc", Namespace: &gwNamespace},
						},
					},
					Rules: []gwv1.GRPCRouteRule{
						{
							BackendRefs: []gwv1.GRPCBackendRef{
								{BackendRef: gwv1.BackendRef{BackendObjectReference: gwv1.BackendObjectReference{
									Name: "svc-nrw-grpc",
								}}},
							},
						},
					},
				},
				Status: gwv1.GRPCRouteStatus{
					RouteStatus: gwv1.RouteStatus{
						Parents: []gwv1.RouteParentStatus{
							{
								ParentRef: gwv1.ParentReference{Name: "gw-nrw-grpc", Namespace: &gwNamespace},
								Conditions: []metav1.Condition{
									{Type: string(gwv1.RouteConditionAccepted), Status: metav1.ConditionTrue},
								},
							},
						},
					},
				},
			}
			createObject(ctx, route, watchClient, &testCreatedObjects)

			Eventually(func(g Gomega) {
				ew := findExternalWorkloadOfObject(fakeCachedClient,
					deployment.GetUID(), deployment.GetNamespace())
				g.Expect(ew).ToNot(BeNil())
				g.Expect(ew.Status.NetworkStatus.ExternalEndpoints).To(HaveLen(1))
				ep := ew.Status.NetworkStatus.ExternalEndpoints[0]
				g.Expect(ep.SourceType).To(Equal(v1alpha1.ExternalEndpointGatewayAPISource))
				g.Expect(ep.Host).To(Equal("grpc.example.com"))
				g.Expect(ep.SourceRef.Kind).To(Equal("GRPCRoute"))
				g.Expect(ep.SourceRef.Name).To(Equal("grpcroute-nrw"))
			}, timeout, interval).Should(Succeed())

			err := watchClient.Delete(ctx, route)
			Expect(err).ToNot(HaveOccurred())

			Eventually(func(g Gomega) {
				ew := findExternalWorkloadOfObject(fakeCachedClient,
					deployment.GetUID(), deployment.GetNamespace())
				g.Expect(ew).ToNot(BeNil())
				g.Expect(ew.Status.NetworkStatus.ExternalEndpoints).To(BeEmpty(),
					"endpoints should be removed after GRPCRoute deletion")
			}, timeout, interval).Should(Succeed())
		})

		It("TLSRoute add updates endpoints, delete removes them", func() {
			deployment, _ := setupDeploymentAndPod("tls")

			os.Setenv(wlconfig.EnvVarClusterDomain, "https://app.example.com:443")
			defer os.Unsetenv(wlconfig.EnvVarClusterDomain)

			svc := &corev1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "svc-nrw-tls",
					Namespace: TestNamespace,
				},
				Spec: corev1.ServiceSpec{
					Type:     corev1.ServiceTypeClusterIP,
					Selector: map[string]string{"app": "nrw-tls"},
					Ports: []corev1.ServicePort{
						{Port: 443, Protocol: corev1.ProtocolTCP},
					},
				},
			}
			createObject(ctx, svc, watchClient, &testCreatedObjects)

			gw := &gwv1.Gateway{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "gw-nrw-tls",
					Namespace: TestNamespace,
				},
				Spec: gwv1.GatewaySpec{
					GatewayClassName: "test-class",
					Listeners: []gwv1.Listener{
						{
							Name:     "tls",
							Port:     8443,
							Protocol: gwv1.TLSProtocolType,
							TLS:      &gwv1.ListenerTLSConfig{},
						},
					},
				},
			}
			createObject(ctx, gw, watchClient, &testCreatedObjects)

			gwNamespace := gwv1.Namespace(TestNamespace)
			route := &gwv1alpha2.TLSRoute{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "tlsroute-nrw",
					Namespace: TestNamespace,
				},
				Spec: gwv1alpha2.TLSRouteSpec{
					Hostnames: []gwv1.Hostname{"tls.example.com"},
					CommonRouteSpec: gwv1.CommonRouteSpec{
						ParentRefs: []gwv1.ParentReference{
							{Name: "gw-nrw-tls", Namespace: &gwNamespace},
						},
					},
					Rules: []gwv1alpha2.TLSRouteRule{
						{
							BackendRefs: []gwv1.BackendRef{
								{BackendObjectReference: gwv1.BackendObjectReference{
									Name: "svc-nrw-tls",
								}},
							},
						},
					},
				},
				Status: gwv1alpha2.TLSRouteStatus{
					RouteStatus: gwv1.RouteStatus{
						Parents: []gwv1.RouteParentStatus{
							{
								ParentRef: gwv1.ParentReference{Name: "gw-nrw-tls", Namespace: &gwNamespace},
								Conditions: []metav1.Condition{
									{Type: string(gwv1.RouteConditionAccepted), Status: metav1.ConditionTrue},
								},
							},
						},
					},
				},
			}
			createObject(ctx, route, watchClient, &testCreatedObjects)

			Eventually(func(g Gomega) {
				ew := findExternalWorkloadOfObject(fakeCachedClient,
					deployment.GetUID(), deployment.GetNamespace())
				g.Expect(ew).ToNot(BeNil())
				g.Expect(ew.Status.NetworkStatus.ExternalEndpoints).To(HaveLen(1))
				ep := ew.Status.NetworkStatus.ExternalEndpoints[0]
				g.Expect(ep.SourceType).To(Equal(v1alpha1.ExternalEndpointGatewayAPISource))
				g.Expect(ep.Host).To(Equal("tls.example.com"))
				g.Expect(ep.TlsEnabled).To(HaveValue(BeTrue()))
				g.Expect(ep.SourceRef.Kind).To(Equal("TLSRoute"))
				g.Expect(ep.SourceRef.Name).To(Equal("tlsroute-nrw"))
			}, timeout, interval).Should(Succeed())

			err := watchClient.Delete(ctx, route)
			Expect(err).ToNot(HaveOccurred())

			Eventually(func(g Gomega) {
				ew := findExternalWorkloadOfObject(fakeCachedClient,
					deployment.GetUID(), deployment.GetNamespace())
				g.Expect(ew).ToNot(BeNil())
				g.Expect(ew.Status.NetworkStatus.ExternalEndpoints).To(BeEmpty(),
					"endpoints should be removed after TLSRoute deletion")
			}, timeout, interval).Should(Succeed())
		})

		It("InferencePool add updates endpoints, delete removes them", func() {
			deployment, _ := setupDeploymentAndPod("pool")

			_ = os.Setenv(wlconfig.EnvVarClusterDomain, "https://app.example.com:443")
			defer func() {
				_ = os.Unsetenv(wlconfig.EnvVarClusterDomain)
			}()
			// Create Gateway that the HTTPRoute references
			gw := &gwv1.Gateway{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "gw-nrw-pool",
					Namespace: TestNamespace,
				},
				Spec: gwv1.GatewaySpec{
					GatewayClassName: "test-class",
					Listeners: []gwv1.Listener{
						{
							Name:     "http",
							Port:     8080,
							Protocol: gwv1.HTTPProtocolType,
						},
					},
				},
			}
			createObject(ctx, gw, watchClient, &testCreatedObjects)

			// Create an HTTPRoute that references an InferencePool as backend
			gwNamespace := gwv1.Namespace(TestNamespace)
			route := &gwv1.HTTPRoute{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "httproute-nrw-pool",
					Namespace: TestNamespace,
				},
				Spec: gwv1.HTTPRouteSpec{
					Hostnames: []gwv1.Hostname{"inference.example.com"},
					CommonRouteSpec: gwv1.CommonRouteSpec{
						ParentRefs: []gwv1.ParentReference{
							{Name: "gw-nrw-pool", Namespace: &gwNamespace},
						},
					},
					Rules: []gwv1.HTTPRouteRule{
						{
							BackendRefs: []gwv1.HTTPBackendRef{
								{BackendRef: gwv1.BackendRef{BackendObjectReference: gwv1.BackendObjectReference{
									Kind:  ptr.To(gwv1.Kind("InferencePool")),
									Group: ptr.To(gwv1.Group(infv1.GroupName)),
									Name:  "pool-nrw",
								}}},
							},
						},
					},
				},
				Status: gwv1.HTTPRouteStatus{
					RouteStatus: gwv1.RouteStatus{
						Parents: []gwv1.RouteParentStatus{
							{
								ParentRef: gwv1.ParentReference{Name: "gw-nrw-pool", Namespace: &gwNamespace},
								Conditions: []metav1.Condition{
									{Type: string(gwv1.RouteConditionAccepted), Status: metav1.ConditionTrue},
								},
							},
						},
					},
				},
			}
			createObject(ctx, route, watchClient, &testCreatedObjects)

			// Create the InferencePool — this triggers the network resource watcher
			pool := &infv1.InferencePool{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "pool-nrw",
					Namespace: TestNamespace,
				},
				Spec: infv1.InferencePoolSpec{
					Selector: infv1.LabelSelector{
						MatchLabels: map[infv1.LabelKey]infv1.LabelValue{
							"app": "nrw-pool",
						},
					},
				},
			}
			createObject(ctx, pool, watchClient, &testCreatedObjects)

			// Verify the EW gets the GatewayAPI endpoint via InferencePool + HTTPRoute
			Eventually(func(g Gomega) {
				ew := findExternalWorkloadOfObject(fakeCachedClient,
					deployment.GetUID(), deployment.GetNamespace())
				g.Expect(ew).ToNot(BeNil())
				g.Expect(ew.Status.NetworkStatus.ExternalEndpoints).To(HaveLen(1))
				ep := ew.Status.NetworkStatus.ExternalEndpoints[0]
				g.Expect(ep.SourceType).To(Equal(v1alpha1.ExternalEndpointGatewayAPISource))
				g.Expect(ep.Host).To(Equal("inference.example.com"))
				g.Expect(ep.SourceRef.Kind).To(Equal("HTTPRoute"))
				g.Expect(ep.SourceRef.Name).To(Equal("httproute-nrw-pool"))
			}, timeout, interval).Should(Succeed())

			// Delete the InferencePool
			err := watchClient.Delete(ctx, pool)
			Expect(err).ToNot(HaveOccurred())

			// Verify the endpoints are removed
			Eventually(func(g Gomega) {
				ew := findExternalWorkloadOfObject(fakeCachedClient,
					deployment.GetUID(), deployment.GetNamespace())
				g.Expect(ew).ToNot(BeNil())
				g.Expect(ew.Status.NetworkStatus.ExternalEndpoints).To(BeEmpty(),
					"endpoints should be removed after InferencePool deletion")
			}, timeout, interval).Should(Succeed())
		})
	})

	It("Deployment with topology annotations", func() {
		testName := "Deployment with topology"
		log.Printf("Starting test: %s", testName)

		replicas := int32(1)
		deployment := &appsv1.Deployment{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "deployment-topology",
				Namespace: TestNamespace,
				Annotations: map[string]string{
					"kai.scheduler/topology":                     "net-topo-test",
					"kai.scheduler/topology-preferred-placement": "pref-json",
					"kai.scheduler/topology-required-placement":  "req-json",
				},
			},
			Spec: appsv1.DeploymentSpec{
				Replicas: &replicas,
				Selector: &metav1.LabelSelector{MatchLabels: map[string]string{"app": "x"}},
				Template: corev1.PodTemplateSpec{
					ObjectMeta: metav1.ObjectMeta{Labels: map[string]string{"app": "x"}},
					Spec: corev1.PodSpec{
						Containers: []corev1.Container{{
							Name:  "main",
							Image: "ubuntu",
						}},
					},
				},
			},
		}

		err := watchClient.Create(ctx, deployment)
		Expect(err).ToNot(HaveOccurred(), "failed to create deployment")
		testCreatedObjects = append(testCreatedObjects, deployment)

		ew := eventuallyFindAndValidateExternalWorkloadWithChildIds(fakeCachedClient,
			nil, deployment, []v1alpha1.ChildId{}, testName, &testCreatedObjects)

		// Topology assertions (end-to-end from top owner annotations)
		Expect(ew.Spec.Topology.Name).To(Equal("net-topo-test"))
		Expect(ew.Spec.Topology.PreferredPlacement).To(Equal("pref-json"))
		Expect(ew.Spec.Topology.RequiredPlacement).To(Equal("req-json"))
	})

	Describe("Test structure e2e", func() {
		// Helper: create Deployment, wait for EW, then create RS + Pod under it.
		// Returns all three with UIDs set by fake client.
		setupDeploymentRSAndPod := func(suffix string) (*appsv1.Deployment, *appsv1.ReplicaSet, *corev1.Pod) {
			deployment := buildDeployment("depl-struct-"+suffix, TestNamespace, "depl-struct-"+suffix+"-uid",
				constants.RunaiSchedulerName)
			obj := createObject(ctx, deployment, watchClient, &testCreatedObjects)
			deployment = obj.(*appsv1.Deployment)

			eventuallyFindAndValidateExternalWorkload(fakeCachedClient,
				nil, deployment, "structure-"+suffix+" - EW created", &testCreatedObjects)

			time.Sleep(30 * time.Millisecond)
			rs := buildReplicaSet("rs-struct-"+suffix, TestNamespace, "rs-struct-"+suffix+"-uid",
				constants.RunaiSchedulerName)
			rs.OwnerReferences = []metav1.OwnerReference{
				{APIVersion: "apps/v1", Kind: "Deployment", Name: deployment.Name, UID: deployment.UID},
			}
			rsObj := createObject(ctx, rs, watchClient, &testCreatedObjects)
			rs = rsObj.(*appsv1.ReplicaSet)

			time.Sleep(30 * time.Millisecond)
			pod := buildPod("pod-struct-"+suffix, TestNamespace, "pod-struct-"+suffix+"-uid")
			pod.OwnerReferences = []metav1.OwnerReference{
				{APIVersion: "apps/v1", Kind: "ReplicaSet", Name: rs.Name, UID: rs.UID},
			}
			createObject(ctx, pod, watchClient, &testCreatedObjects)

			return deployment, rs, pod
		}

		// Helper: add a second RS + Pod under an existing Deployment.
		addSecondRSAndPod := func(suffix string, deployment *appsv1.Deployment) (*appsv1.ReplicaSet, *corev1.Pod) {
			time.Sleep(30 * time.Millisecond)
			rs2 := buildReplicaSet("rs-struct-"+suffix+"-2", TestNamespace, "rs-struct-"+suffix+"-2-uid",
				constants.RunaiSchedulerName)
			rs2.OwnerReferences = []metav1.OwnerReference{
				{APIVersion: "apps/v1", Kind: "Deployment", Name: deployment.Name, UID: deployment.UID},
			}
			rsObj := createObject(ctx, rs2, watchClient, &testCreatedObjects)
			rs2 = rsObj.(*appsv1.ReplicaSet)

			time.Sleep(30 * time.Millisecond)
			pod2 := buildPod("pod-struct-"+suffix+"-2", TestNamespace, "pod-struct-"+suffix+"-2-uid")
			pod2.OwnerReferences = []metav1.OwnerReference{
				{APIVersion: "apps/v1", Kind: "ReplicaSet", Name: rs2.Name, UID: rs2.UID},
			}
			createObject(ctx, pod2, watchClient, &testCreatedObjects)

			return rs2, pod2
		}

		It("structure elements populated on pod add for workload with Karta", func() {
			deployment, rs, pod := setupDeploymentRSAndPod("add")

			// Verify EW gets StructureElements with the RS element
			var elementID string
			Eventually(func(g Gomega) {
				ew := findExternalWorkloadOfObject(fakeCachedClient,
					deployment.GetUID(), deployment.GetNamespace())
				g.Expect(ew).ToNot(BeNil(), "should find EW for deployment")
				g.Expect(ew.Status.StructureElements).ToNot(BeNil(),
					"StructureElements should be populated after pod add")
				g.Expect(ew.Status.StructureElements.Elements).To(HaveLen(1),
					"should have 1 element (replicaset)")

				elem := ew.Status.StructureElements.Elements[0]
				g.Expect(elem.Name).To(Equal("replicaset"))
				g.Expect(elem.Level).To(Equal(1))
				g.Expect(elem.ParentID).To(BeNil(), "level 1 element has no parent")
				g.Expect(elem.K8sRef).ToNot(BeNil(), "replicaset element should be object-backed")
				g.Expect(elem.K8sRef.UID).To(Equal(rs.UID))
				g.Expect(elem.K8sRef.Name).To(Equal(rs.Name))
				elementID = elem.ID
			}, timeout, interval).Should(Succeed())

			// Verify the pod label VALUE equals the element ID
			Eventually(func(g Gomega) {
				freshPod := &corev1.Pod{}
				err := watchClient.Get(ctx, types.NamespacedName{Name: pod.Name, Namespace: pod.Namespace}, freshPod)
				g.Expect(err).ToNot(HaveOccurred())
				g.Expect(freshPod.Labels["run.ai/element-1-id"]).To(Equal(elementID),
					"pod element-1-id label value should equal the element ID")
			}, timeout, interval).Should(Succeed())
		})

		It("second pod shares same structure element", func() {
			deployment, rs, pod1 := setupDeploymentRSAndPod("shared")

			// Wait for structure to be populated from first pod
			var firstElementID string
			Eventually(func(g Gomega) {
				ew := findExternalWorkloadOfObject(fakeCachedClient,
					deployment.GetUID(), deployment.GetNamespace())
				g.Expect(ew).ToNot(BeNil())
				g.Expect(ew.Status.StructureElements).ToNot(BeNil())
				g.Expect(ew.Status.StructureElements.Elements).To(HaveLen(1))
				firstElementID = ew.Status.StructureElements.Elements[0].ID
			}, timeout, interval).Should(Succeed())

			// Verify pod1 label value matches the element ID
			Eventually(func(g Gomega) {
				freshPod1 := &corev1.Pod{}
				err := watchClient.Get(ctx, types.NamespacedName{Name: pod1.Name, Namespace: pod1.Namespace}, freshPod1)
				g.Expect(err).ToNot(HaveOccurred())
				g.Expect(freshPod1.Labels["run.ai/element-1-id"]).To(Equal(firstElementID),
					"pod1 element-1-id label value should equal the element ID")
			}, timeout, interval).Should(Succeed())

			// Add a second pod owned by the same RS
			time.Sleep(30 * time.Millisecond)
			pod2 := buildPod("pod-struct-shared-2", TestNamespace, "pod-struct-shared-2-uid")
			pod2.OwnerReferences = []metav1.OwnerReference{
				{APIVersion: "apps/v1", Kind: "ReplicaSet", Name: rs.Name, UID: rs.UID},
			}
			createObject(ctx, pod2, watchClient, &testCreatedObjects)

			// Verify structure still has 1 element (same RS), and pod2 label value matches
			Eventually(func(g Gomega) {
				ew := findExternalWorkloadOfObject(fakeCachedClient,
					deployment.GetUID(), deployment.GetNamespace())
				g.Expect(ew).ToNot(BeNil())
				g.Expect(ew.Status.StructureElements).ToNot(BeNil())
				g.Expect(ew.Status.StructureElements.Elements).To(HaveLen(1),
					"still 1 element - same RS for both pods")
				g.Expect(ew.Status.StructureElements.Elements[0].K8sRef.UID).To(Equal(rs.UID))

				// Pod2 label value should equal the same element ID
				freshPod2 := &corev1.Pod{}
				err := watchClient.Get(ctx, types.NamespacedName{Name: pod2.Name, Namespace: pod2.Namespace}, freshPod2)
				g.Expect(err).ToNot(HaveOccurred())
				g.Expect(freshPod2.Labels["run.ai/element-1-id"]).To(Equal(firstElementID),
					"pod2 element-1-id label value should equal the same element ID as pod1")
			}, timeout, interval).Should(Succeed())
		})

		It("structure grows when pod from different RS is added, and shrinks when deleted", func() {
			// Phase 1: Create Deployment → RS1 → Pod1, wait for 1 element
			deployment, rs1, pod1 := setupDeploymentRSAndPod("2rs")

			var rs1ElementID string
			Eventually(func(g Gomega) {
				ew := findExternalWorkloadOfObject(fakeCachedClient,
					deployment.GetUID(), deployment.GetNamespace())
				g.Expect(ew).ToNot(BeNil())
				g.Expect(ew.Status.StructureElements).ToNot(BeNil())
				g.Expect(ew.Status.StructureElements.Elements).To(HaveLen(1))
				g.Expect(ew.Status.StructureElements.Elements[0].K8sRef.UID).To(Equal(rs1.UID))
				rs1ElementID = ew.Status.StructureElements.Elements[0].ID
			}, timeout, interval).Should(Succeed())

			// Verify pod1 label value equals the RS1 element ID
			Eventually(func(g Gomega) {
				freshPod1 := &corev1.Pod{}
				err := watchClient.Get(ctx, types.NamespacedName{Name: pod1.Name, Namespace: pod1.Namespace}, freshPod1)
				g.Expect(err).ToNot(HaveOccurred())
				g.Expect(freshPod1.Labels["run.ai/element-1-id"]).To(Equal(rs1ElementID),
					"pod1 label value should equal the RS1 element ID")
			}, timeout, interval).Should(Succeed())

			// Phase 2: Add RS2 → Pod2 (different RS), structure should grow to 2 elements
			rs2, pod2 := addSecondRSAndPod("2rs", deployment)

			var rs2ElementID string
			Eventually(func(g Gomega) {
				ew := findExternalWorkloadOfObject(fakeCachedClient,
					deployment.GetUID(), deployment.GetNamespace())
				g.Expect(ew).ToNot(BeNil())
				g.Expect(ew.Status.StructureElements).ToNot(BeNil())
				g.Expect(ew.Status.StructureElements.Elements).To(HaveLen(2),
					"should have 2 elements (RS1 + RS2)")

				for _, elem := range ew.Status.StructureElements.Elements {
					g.Expect(elem.Name).To(Equal("replicaset"))
					g.Expect(elem.Level).To(Equal(1))
					g.Expect(elem.K8sRef).ToNot(BeNil())
					if elem.K8sRef.UID == rs1.UID {
						g.Expect(elem.ID).To(Equal(rs1ElementID), "RS1 element ID should be stable")
					} else {
						g.Expect(elem.K8sRef.UID).To(Equal(rs2.UID))
						rs2ElementID = elem.ID
					}
				}
				g.Expect(rs2ElementID).ToNot(BeEmpty(), "should have found the RS2 element")
			}, timeout, interval).Should(Succeed())

			// Verify pod2 label value equals the RS2 element ID
			Eventually(func(g Gomega) {
				freshPod2 := &corev1.Pod{}
				err := watchClient.Get(ctx, types.NamespacedName{Name: pod2.Name, Namespace: pod2.Namespace}, freshPod2)
				g.Expect(err).ToNot(HaveOccurred())
				g.Expect(freshPod2.Labels["run.ai/element-1-id"]).To(Equal(rs2ElementID),
					"pod2 label value should equal the RS2 element ID")
			}, timeout, interval).Should(Succeed())

			// Phase 3: Delete Pod2 → structure should shrink back to 1 element (RS1 only)
			err := watchClient.Get(ctx, types.NamespacedName{Name: pod2.Name, Namespace: pod2.Namespace}, pod2)
			Expect(err).ToNot(HaveOccurred())
			pod2.GetObjectKind().SetGroupVersionKind(corev1.SchemeGroupVersion.WithKind("Pod"))
			err = watchClient.Delete(ctx, pod2)
			Expect(err).ToNot(HaveOccurred())

			Eventually(func(g Gomega) {
				ew := findExternalWorkloadOfObject(fakeCachedClient,
					deployment.GetUID(), deployment.GetNamespace())
				g.Expect(ew).ToNot(BeNil())
				g.Expect(ew.Status.StructureElements).ToNot(BeNil())
				g.Expect(ew.Status.StructureElements.Elements).To(HaveLen(1),
					"should be back to 1 element after deleting pod2 (RS2 element removed)")
				g.Expect(ew.Status.StructureElements.Elements[0].K8sRef.UID).To(Equal(rs1.UID),
					"remaining element should be RS1")
				g.Expect(ew.Status.StructureElements.Elements[0].ID).To(Equal(rs1ElementID),
					"RS1 element ID should be stable across recalculations")
			}, timeout, interval).Should(Succeed())

			// Delete Pod1 (last pod) → structure should become empty
			err = watchClient.Get(ctx, types.NamespacedName{Name: pod1.Name, Namespace: pod1.Namespace}, pod1)
			Expect(err).ToNot(HaveOccurred())
			pod1.GetObjectKind().SetGroupVersionKind(corev1.SchemeGroupVersion.WithKind("Pod"))
			err = watchClient.Delete(ctx, pod1)
			Expect(err).ToNot(HaveOccurred())

			Eventually(func(g Gomega) {
				ew := findExternalWorkloadOfObject(fakeCachedClient,
					deployment.GetUID(), deployment.GetNamespace())
				g.Expect(ew).ToNot(BeNil())
				g.Expect(ew.Status.StructureElements).ToNot(BeNil(),
					"StructureElements should still exist (was calculated)")
				g.Expect(ew.Status.StructureElements.Elements).To(BeEmpty(),
					"Elements should be empty after all pods are deleted")
			}, timeout, interval).Should(Succeed())
		})

		It("structure not changed on pod update", func() {
			deployment, _, pod := setupDeploymentRSAndPod("upd")

			// Wait for structure to be populated and pod labeled
			var elementID string
			Eventually(func(g Gomega) {
				ew := findExternalWorkloadOfObject(fakeCachedClient,
					deployment.GetUID(), deployment.GetNamespace())
				g.Expect(ew).ToNot(BeNil())
				g.Expect(ew.Status.StructureElements).ToNot(BeNil())
				g.Expect(ew.Status.StructureElements.Elements).To(HaveLen(1))
				elementID = ew.Status.StructureElements.Elements[0].ID

				freshPod := &corev1.Pod{}
				err := watchClient.Get(ctx, types.NamespacedName{Name: pod.Name, Namespace: pod.Namespace}, freshPod)
				g.Expect(err).ToNot(HaveOccurred())
				g.Expect(freshPod.Labels["run.ai/element-1-id"]).To(Equal(elementID),
					"pod label value should equal the element ID")
			}, timeout, interval).Should(Succeed())

			// Capture LastCalculated
			ew := findExternalWorkloadOfObject(fakeCachedClient,
				deployment.GetUID(), deployment.GetNamespace())
			lastCalcBefore := ew.Status.StructureElements.LastCalculated

			// Re-fetch pod to get current ResourceVersion
			err := watchClient.Get(ctx, types.NamespacedName{Name: pod.Name, Namespace: pod.Namespace}, pod)
			Expect(err).ToNot(HaveOccurred())
			pod.GetObjectKind().SetGroupVersionKind(corev1.SchemeGroupVersion.WithKind("Pod"))

			// Update pod (triggers UPDATE event, not ADD)
			updateObjectWithRandomLabel(ctx, pod, watchClient)

			// Structure should NOT be recalculated — LastCalculated stays the same
			Consistently(func(g Gomega) {
				ew := findExternalWorkloadOfObject(fakeCachedClient,
					deployment.GetUID(), deployment.GetNamespace())
				g.Expect(ew).ToNot(BeNil())
				g.Expect(ew.Status.StructureElements).ToNot(BeNil())
				g.Expect(ew.Status.StructureElements.LastCalculated.Time).To(Equal(lastCalcBefore.Time),
					"LastCalculated should NOT change on pod update")
			}, consistentlyTimeout, interval).Should(Succeed())
		})

		It("no structure elements when no Karta exists for workload type", func() {
			// StatefulSet has no Karta registered in the test setup
			ss := buildStatefulSet("ss-struct-nori", TestNamespace, "ss-struct-nori-uid",
				constants.RunaiSchedulerName)
			obj := createObject(ctx, ss, watchClient, &testCreatedObjects)
			ss = obj.(*appsv1.StatefulSet)

			eventuallyFindAndValidateExternalWorkload(fakeCachedClient,
				nil, ss, "structure-nori - EW created", &testCreatedObjects)

			// Create a pod owned by the StatefulSet
			time.Sleep(30 * time.Millisecond)
			pod := buildPod("pod-struct-nori", TestNamespace, "pod-struct-nori-uid")
			pod.OwnerReferences = []metav1.OwnerReference{
				{APIVersion: "apps/v1", Kind: "StatefulSet", Name: ss.Name, UID: ss.UID},
			}
			createObject(ctx, pod, watchClient, &testCreatedObjects)

			// Wait for pod labeling (ensures the pod add event was processed)
			Eventually(func(g Gomega) {
				freshPod := &corev1.Pod{}
				err := watchClient.Get(ctx, types.NamespacedName{Name: pod.Name, Namespace: pod.Namespace}, freshPod)
				g.Expect(err).ToNot(HaveOccurred())
				g.Expect(freshPod.Labels).To(HaveKey(commonconsts.ClusterTopOwnerUidLabel),
					"pod should be labeled with top-owner UID")
			}, timeout, interval).Should(Succeed())

			// StructureElements should be with timestamp and no elements — no Karta for StatefulSet
			Consistently(func(g Gomega) {
				ew := findExternalWorkloadOfObject(fakeCachedClient,
					ss.GetUID(), ss.GetNamespace())
				g.Expect(ew).ToNot(BeNil())
				g.Expect(ew.Status.StructureElements).ToNot(BeNil(),
					"StructureElements should be nil when no Karta exists")
				g.Expect(ew.Status.StructureElements.Elements).To(BeEmpty(), "should have no elements when no Karta exists")
				g.Expect(ew.Status.StructureElements.LastCalculated).ToNot(BeZero(),
					"LastCalculated should be set even when no Karta exists")
			}, consistentlyTimeout, interval).Should(Succeed())
		})
	})

	Describe("NodeAffinity propagation per handler type", Ordered, func() {
		// Generic test affinity that's distinct from the ones in shared fixtures.
		buildTestAffinity := func() *corev1.NodeAffinity {
			return &corev1.NodeAffinity{
				RequiredDuringSchedulingIgnoredDuringExecution: &corev1.NodeSelector{
					NodeSelectorTerms: []corev1.NodeSelectorTerm{{
						MatchExpressions: []corev1.NodeSelectorRequirement{{
							Key:      "node-affinity-test/key",
							Operator: corev1.NodeSelectorOpIn,
							Values:   []string{"a", "b"},
						}},
					}},
				},
			}
		}
		expectedSingle := func(specName string, na *corev1.NodeAffinity) []v1alpha1.NodeAffinityEntry {
			return []v1alpha1.NodeAffinityEntry{{SpecName: specName, NodeAffinity: na}}
		}

		// Helpers — clone an existing fixture, replace its podTemplate NodeAffinity with the test one,
		// give it a unique UID/Name so it doesn't collide with other test runs.
		mkDeployment := func(uid string, na *corev1.NodeAffinity) *appsv1.Deployment {
			d := TestDeployment.DeepCopy()
			d.Name = "na-depl-" + uid
			d.UID = types.UID("na-depl-uid-" + uid)
			d.Spec.Template.Spec.Affinity = &corev1.Affinity{NodeAffinity: na}
			return d
		}
		mkStatefulSet := func(uid string, na *corev1.NodeAffinity) *appsv1.StatefulSet {
			s := TestStatefulSet.DeepCopy()
			s.Name = "na-ss-" + uid
			s.UID = types.UID("na-ss-uid-" + uid)
			s.Spec.Template.Spec.Affinity = &corev1.Affinity{NodeAffinity: na}
			return s
		}
		mkReplicaSet := func(uid string, na *corev1.NodeAffinity) *appsv1.ReplicaSet {
			r := TestReplicaSet.DeepCopy()
			r.Name = "na-rs-" + uid
			r.UID = types.UID("na-rs-uid-" + uid)
			r.OwnerReferences = nil // make it top-owned
			r.Spec.Template.Spec.Affinity = &corev1.Affinity{NodeAffinity: na}
			return r
		}
		mkJob := func(uid string, na *corev1.NodeAffinity) *batchv1.Job {
			j := TestBatchJob.DeepCopy()
			j.Name = "na-job-" + uid
			j.UID = types.UID("na-job-uid-" + uid)
			j.Spec.Template.Spec.Affinity = &corev1.Affinity{NodeAffinity: na}
			return j
		}
		mkCronJob := func(uid string, na *corev1.NodeAffinity) *batchv1.CronJob {
			c := TestCronJob.DeepCopy()
			c.Name = "na-cj-" + uid
			c.UID = types.UID("na-cj-uid-" + uid)
			c.Spec.JobTemplate.Spec.Template.Spec.Affinity = &corev1.Affinity{NodeAffinity: na}
			return c
		}
		mkKnativeService := func(uid string, na *corev1.NodeAffinity) *servingv1.Service {
			k := TestKnativeService.DeepCopy()
			k.Name = "na-ks-" + uid
			k.UID = types.UID("na-ks-uid-" + uid)
			k.Spec.Template.Spec.Affinity = &corev1.Affinity{NodeAffinity: na}
			return k
		}
		mkNotebook := func(uid string, na *corev1.NodeAffinity) *kubeflowv1beta1.Notebook {
			n := TestNotebook1.DeepCopy()
			n.Name = "na-nb-" + uid
			n.UID = types.UID("na-nb-uid-" + uid)
			n.Spec.Template.Spec.Affinity = &corev1.Affinity{NodeAffinity: na}
			return n
		}
		mkPod := func(uid string, na *corev1.NodeAffinity) *corev1.Pod {
			p := TestPod.DeepCopy()
			p.Name = "na-pod-" + uid
			p.UID = types.UID("na-pod-uid-" + uid)
			p.Spec.Affinity = &corev1.Affinity{NodeAffinity: na}
			return p
		}

		assertEntries := func(owner client.Object, expected []v1alpha1.NodeAffinityEntry) {
			Eventually(func(g Gomega) {
				ew := findExternalWorkloadOfObject(fakeCachedClient, owner.GetUID(), owner.GetNamespace())
				g.Expect(ew).ToNot(BeNil(), "external workload not found for owner %s/%s",
					owner.GetNamespace(), owner.GetName())
				g.Expect(ew.Spec.NodeAffinities).To(ConsistOf(expected),
					"NodeAffinities for %s/%s", owner.GetNamespace(), owner.GetName())
			}, timeout, interval).Should(Succeed())
		}

		It("Deployment → single entry under DefaultNodeAffinitySpecName", func() {
			na := buildTestAffinity()
			owner := createObject(ctx, mkDeployment("depl", na), watchClient, &testCreatedObjects)
			assertEntries(owner, expectedSingle(v1alpha1.DefaultNodeAffinitySpecName, na))
		})
		It("StatefulSet → single entry under DefaultNodeAffinitySpecName", func() {
			na := buildTestAffinity()
			owner := createObject(ctx, mkStatefulSet("ss", na), watchClient, &testCreatedObjects)
			assertEntries(owner, expectedSingle(v1alpha1.DefaultNodeAffinitySpecName, na))
		})
		It("ReplicaSet → single entry under DefaultNodeAffinitySpecName", func() {
			na := buildTestAffinity()
			owner := createObject(ctx, mkReplicaSet("rs", na), watchClient, &testCreatedObjects)
			assertEntries(owner, expectedSingle(v1alpha1.DefaultNodeAffinitySpecName, na))
		})
		It("Job → single entry under DefaultNodeAffinitySpecName", func() {
			na := buildTestAffinity()
			owner := createObject(ctx, mkJob("job", na), watchClient, &testCreatedObjects)
			assertEntries(owner, expectedSingle(v1alpha1.DefaultNodeAffinitySpecName, na))
		})
		It("CronJob → single entry under DefaultNodeAffinitySpecName", func() {
			na := buildTestAffinity()
			owner := createObject(ctx, mkCronJob("cj", na), watchClient, &testCreatedObjects)
			assertEntries(owner, expectedSingle(v1alpha1.DefaultNodeAffinitySpecName, na))
		})
		It("Knative Service → single entry under DefaultNodeAffinitySpecName", func() {
			na := buildTestAffinity()
			owner := createObject(ctx, mkKnativeService("ks", na), watchClient, &testCreatedObjects)
			assertEntries(owner, expectedSingle(v1alpha1.DefaultNodeAffinitySpecName, na))
		})
		It("Kubeflow Notebook → single entry under DefaultNodeAffinitySpecName", func() {
			na := buildTestAffinity()
			owner := createObject(ctx, mkNotebook("nb", na), watchClient, &testCreatedObjects)
			assertEntries(owner, expectedSingle(v1alpha1.DefaultNodeAffinitySpecName, na))
		})
		It("Bare Pod → single entry under DefaultNodeAffinitySpecName", func() {
			na := buildTestAffinity()
			owner := createObject(ctx, mkPod("pod", na), watchClient, &testCreatedObjects)
			assertEntries(owner, expectedSingle(v1alpha1.DefaultNodeAffinitySpecName, na))
		})

		It("Karta (kserve InferenceService) → distinct entries per component", func() {
			// Drives the Karta path (FillExternalWorkloadDataFromPodTemplateSpecIfEmpty with
			// specName = riComponent.Name()). The KartaForKserveInferenceService fixture
			// declares a root "inferenceservice" component and a child "predictor" component.
			// Injecting affinity into the Predictor's PodSpec should produce a "predictor" entry.
			predictorNa := &corev1.NodeAffinity{
				RequiredDuringSchedulingIgnoredDuringExecution: &corev1.NodeSelector{
					NodeSelectorTerms: []corev1.NodeSelectorTerm{{MatchExpressions: []corev1.NodeSelectorRequirement{
						{Key: "node-affinity-test/role", Operator: corev1.NodeSelectorOpIn, Values: []string{"predictor"}},
					}}},
				},
			}

			isvc := TestKserveInferenceService.DeepCopy()
			isvc.Name = "na-isvc"
			isvc.UID = "na-isvc-uid"
			isvc.Spec.Predictor.Affinity = &corev1.Affinity{NodeAffinity: predictorNa}

			owner := createObject(ctx, isvc, watchClient, &testCreatedObjects)
			Eventually(func(g Gomega) {
				ew := findExternalWorkloadOfObject(fakeCachedClient, owner.GetUID(), owner.GetNamespace())
				g.Expect(ew).ToNot(BeNil())
				g.Expect(ew.Spec.NodeAffinities).To(ContainElement(
					v1alpha1.NodeAffinityEntry{SpecName: "predictor", NodeAffinity: predictorNa}))
			}, timeout, interval).Should(Succeed())
		})

		It("RayCluster → distinct entries for head and each worker group", func() {
			rc := TestRayCluster.DeepCopy()
			rc.Name = "na-ray"
			rc.UID = "na-ray-uid"
			headNa := &corev1.NodeAffinity{
				RequiredDuringSchedulingIgnoredDuringExecution: &corev1.NodeSelector{
					NodeSelectorTerms: []corev1.NodeSelectorTerm{{MatchExpressions: []corev1.NodeSelectorRequirement{
						{Key: "node-affinity-test/role", Operator: corev1.NodeSelectorOpIn, Values: []string{"head"}},
					}}},
				},
			}
			workerNa := &corev1.NodeAffinity{
				RequiredDuringSchedulingIgnoredDuringExecution: &corev1.NodeSelector{
					NodeSelectorTerms: []corev1.NodeSelectorTerm{{MatchExpressions: []corev1.NodeSelectorRequirement{
						{Key: "node-affinity-test/role", Operator: corev1.NodeSelectorOpIn, Values: []string{"worker"}},
					}}},
				},
			}
			rc.Spec.HeadGroupSpec.Template.Spec.Affinity = &corev1.Affinity{NodeAffinity: headNa}

			workerGroupName := "wg-test"
			rc.Spec.WorkerGroupSpecs = []rayv1alpha1.WorkerGroupSpec{{
				GroupName: workerGroupName,
				Replicas:  ptr.To(int32(1)),
				Template: corev1.PodTemplateSpec{
					Spec: corev1.PodSpec{
						SchedulerName: constants.RunaiSchedulerName,
						Affinity:      &corev1.Affinity{NodeAffinity: workerNa},
						Containers:    []corev1.Container{{Image: testImage}},
					},
				},
			}}

			owner := createObject(ctx, rc, watchClient, &testCreatedObjects)
			assertEntries(owner, []v1alpha1.NodeAffinityEntry{
				{SpecName: "head", NodeAffinity: headNa},
				{SpecName: "worker-" + workerGroupName, NodeAffinity: workerNa},
			})
		})
	})
})

func createObject(ctx context.Context, object client.Object, watchClient client.WithWatch, testCreatedObjects *[]client.Object) client.Object {
	// Preserve the GVK before creation as the fake client may clear it
	gvk := object.GetObjectKind().GroupVersionKind()

	err := watchClient.Create(ctx, object)
	ExpectWithOffset(1, err).ToNot(HaveOccurred())

	// Restore the GVK after creation
	object.GetObjectKind().SetGroupVersionKind(gvk)

	*testCreatedObjects = append(*testCreatedObjects, object)
	return object
}

func updateObjectWithRandomLabel(ctx context.Context, object client.Object, watchClient client.WithWatch) client.Object {
	// Preserve the GVK before update as the fake client may clear it
	gvk := object.GetObjectKind().GroupVersionKind()

	labels := object.GetLabels()
	if labels == nil {
		labels = map[string]string{}
	}
	labels["random-label"] = rand.String(5)
	object.SetLabels(labels)
	err := watchClient.Update(ctx, object)
	ExpectWithOffset(1, err).ToNot(HaveOccurred())

	// Restore the GVK after update
	object.GetObjectKind().SetGroupVersionKind(gvk)

	return object
}

func eventuallyFindAndValidateExternalWorkload(k8sClient client.Reader, expected *v1alpha1.ExternalWorkload,
	originalObject client.Object, testName string,
	testCreatedObjects *[]client.Object,
) *v1alpha1.ExternalWorkload {
	return eventuallyFindAndValidateExternalWorkloadWithChildIds(k8sClient, expected, originalObject,
		[]v1alpha1.ChildId{}, testName, testCreatedObjects)
}

func eventuallyFindAndValidateExternalWorkloadWithChildIds(k8sClient client.Reader, expected *v1alpha1.ExternalWorkload,
	originalObject client.Object, expectedChildIds []v1alpha1.ChildId, testName string,
	testCreatedObjects *[]client.Object,
) *v1alpha1.ExternalWorkload {
	var externalWorkloadOfObject *v1alpha1.ExternalWorkload
	Eventually(func(g Gomega) {
		externalWorkloadOfObject = findExternalWorkloadOfObject(k8sClient,
			getExpectedExternalWorkloadUid(expected, originalObject), originalObject.GetNamespace())
		g.Expect(externalWorkloadOfObject).ToNot(BeNil(),
			"Test: <%s>, should find external workload for object <%s> with UID <%s>",
			testName, originalObject.GetObjectKind().GroupVersionKind().String(), getExpectedExternalWorkloadUid(expected, originalObject))
		*testCreatedObjects = append(*testCreatedObjects, externalWorkloadOfObject)

		if expected == nil {
			return
		}

		expected.Status.ObservedStatus.ChildIds = expectedChildIds
		validateExternalWorkload(g, 2, externalWorkloadOfObject,
			expected, originalObject)
	}, timeout, interval).Should(Succeed(), fmt.Sprintf("Test <%s>", testName))

	return externalWorkloadOfObject
}

func getExpectedExternalWorkloadUid(expected *v1alpha1.ExternalWorkload, originalObject client.Object) types.UID {
	// some handlers (e.g. nvcf) set a non-default UID for the external workload
	if expected != nil && expected.Spec.OriginalWorkloadUID != "" {
		return expected.Spec.OriginalWorkloadUID
	}

	return originalObject.GetUID()
}

func getExpectedExternalWorkloadName(expected *v1alpha1.ExternalWorkload, originalObject client.Object) string {
	// some handlers (e.g. nvcf) set a non-default name for the external workload
	if expected != nil && expected.Spec.OriginalWorkloadName != "" {
		return expected.Spec.OriginalWorkloadName
	}

	return originalObject.GetName()
}

func findExternalWorkloadOfObject(k8sClient client.Reader,
	objectUID types.UID, namespace string,
) *v1alpha1.ExternalWorkload {
	externalWorkloads := findAllExternalWorkloadOfObject(k8sClient, objectUID, namespace)
	if len(externalWorkloads) > 0 {
		return &externalWorkloads[0]
	}

	return nil
}

func findAllExternalWorkloadOfObject(k8sClient client.Reader,
	objectUID types.UID, namespace string,
) []v1alpha1.ExternalWorkload {
	var matchingExternalWorkloads []v1alpha1.ExternalWorkload
	externalWorkloads := listExternalWorkloadsFromClient(k8sClient)
	for i, externalWorkload := range externalWorkloads {
		if externalWorkload.Spec.OriginalWorkloadUID == objectUID && externalWorkload.Namespace == namespace {
			matchingExternalWorkloads = append(matchingExternalWorkloads, externalWorkloads[i])
		}
	}

	return matchingExternalWorkloads
}

func validateExternalWorkload(g Gomega, offset int, actual, expected *v1alpha1.ExternalWorkload, originalObject client.Object) {
	if expected == nil {
		return
	}

	// The fake client strips TypeMeta on Create/Update, so the in-memory original
	// object may have an empty GVK; recover it from the scheme. In production the
	// object comes from a dynamic (unstructured) informer that retains apiVersion/kind.
	gvk := originalObject.GetObjectKind().GroupVersionKind()
	if gvk.Empty() {
		if gvks, _, err := getTestScheme().ObjectKinds(originalObject); err == nil && len(gvks) > 0 {
			gvk = gvks[0]
		}
	}

	expectedName := utils.GenerateExternalWorkloadName(gvk, getExpectedExternalWorkloadName(expected, originalObject))
	if len(expectedName) > 58 {
		g.ExpectWithOffset(offset, actual.Name[:len(actual.Name)-4]).To(Equal(expectedName[:len(expectedName)-4]),
			"ExternWorkload.Name")
	} else {
		g.ExpectWithOffset(offset, actual.Name).To(Equal(expectedName), "ExternWorkload.Name")
	}
	g.ExpectWithOffset(offset, actual.Namespace).To(Equal(originalObject.GetNamespace()),
		"ExternWorkload.Namespace")
	g.ExpectWithOffset(offset, len(actual.OwnerReferences)).To(Equal(1),
		"ExternWorkload.OwnerReferences")
	g.ExpectWithOffset(offset, actual.OwnerReferences[0]).To(Equal(metav1.OwnerReference{
		APIVersion: gvk.GroupVersion().String(),
		Kind:       gvk.Kind,
		Name:       originalObject.GetName(),
		UID:        originalObject.GetUID(),
	}), "ExternWorkload.OwnerReferences[0]")
	g.ExpectWithOffset(offset, actual.Spec.OriginalWorkloadUID).To(Equal(getExpectedExternalWorkloadUid(expected, originalObject)),
		"OriginalWorkloadUID")
	g.ExpectWithOffset(offset, actual.Spec.OriginalWorkloadName).To(Equal(getExpectedExternalWorkloadName(expected, originalObject)),
		"OriginalWorkloadName")
	creationTimestamp := originalObject.GetCreationTimestamp()
	g.ExpectWithOffset(offset, creationTimestamp.IsZero()).To(BeFalse(),
		"CreationTimestampExists")
	g.ExpectWithOffset(offset, actual.Spec.CreationTimestamp.Unix()).To(Equal(creationTimestamp.Unix()),
		"CreationTimestamp")

	if expected.Annotations != nil {
		for key, value := range expected.Annotations {
			g.ExpectWithOffset(offset, actual.Annotations[key]).To(Equal(value),
				"Annotations[%s]", key)
		}
	}

	if expected.Labels != nil {
		for key, value := range expected.Labels {
			g.ExpectWithOffset(offset, actual.Labels[key]).To(Equal(value),
				"Labels[%s]", key)
		}
	}

	if expected.Spec.Type != "" {
		g.ExpectWithOffset(offset, actual.Spec.Type).To(Equal(expected.Spec.Type),
			"Type")
	} else {
		g.ExpectWithOffset(offset, actual.Spec.Type).To(Equal(gvk.Kind),
			"Type")
	}
	if expected.Spec.Group != "" {
		g.ExpectWithOffset(offset, actual.Spec.Group).To(Equal(expected.Spec.Group),
			"Group")
	} else {
		g.ExpectWithOffset(offset, actual.Spec.Group).To(Equal(gvk.Group),
			"Group")
	}
	if expected.Spec.Version != "" {
		g.ExpectWithOffset(offset, actual.Spec.Version).To(Equal(expected.Spec.Version),
			"Version")
	} else {
		g.ExpectWithOffset(offset, actual.Spec.Version).To(Equal(gvk.Version),
			"Version")
	}
	if expected.Spec.ProjectName != "" {
		g.ExpectWithOffset(offset, actual.Spec.ProjectName).To(Equal(expected.Spec.ProjectName),
			"ProjectName")
	}
	if expected.Spec.ProjectId != "" {
		g.ExpectWithOffset(offset, actual.Spec.ProjectId).To(Equal(expected.Spec.ProjectId),
			"ProjectId")
	}
	if expected.Spec.DepartmentName != "" {
		g.ExpectWithOffset(offset, actual.Spec.DepartmentName).To(Equal(expected.Spec.DepartmentName),
			"DepartmentName")
	}
	if expected.Spec.DepartmentId != "" {
		g.ExpectWithOffset(offset, actual.Spec.DepartmentId).To(Equal(expected.Spec.DepartmentId),
			"DepartmentId")
	}

	ValidateExternalWorkloadInnerSpec(g, offset+1, actual, expected)

	actualObservedChildIds := actual.Status.ObservedStatus.ChildIds
	if actualObservedChildIds == nil {
		actualObservedChildIds = []v1alpha1.ChildId{}
	}
	expectedObservedChildIds := expected.Status.ObservedStatus.ChildIds
	if len(expectedObservedChildIds) > 0 {
		g.ExpectWithOffset(offset, len(actualObservedChildIds)).To(Equal(len(expectedObservedChildIds)),
			"observedChildIds-len, expected: <%+v>, actual: <%+v>",
			expectedObservedChildIds, actualObservedChildIds)
		g.ExpectWithOffset(offset, actualObservedChildIds).To(ContainElements(expectedObservedChildIds),
			"observedChildIds")
	} else {
		g.ExpectWithOffset(offset, len(actualObservedChildIds)).To(Equal(0),
			"observedChildIds-len")
	}
}

func consistentlyValidateExternalWorkloadDoesNotExist(k8sClient client.Reader, objectUID types.UID,
	namespace, kind string,
) {
	Consistently(func(g Gomega) {
		validateExternalWorkloadExistence(g, false, k8sClient, objectUID, namespace, kind)
	}, consistentlyTimeout, interval).Should(Succeed())
}

func validateExternalWorkloadDoesNotExist(k8sClient client.Reader, objectUID types.UID,
	namespace, kind string,
) {
	Eventually(func(g Gomega) {
		validateExternalWorkloadExistence(g, false, k8sClient, objectUID, namespace, kind)
	}).Should(Succeed())
}

func validateExternalWorkloadExists(k8sClient client.Reader, objectUID types.UID,
	namespace, kind string,
) {
	Eventually(func(g Gomega) {
		validateExternalWorkloadExistence(g, true, k8sClient, objectUID, namespace, kind)
	}).Should(Succeed())
}

func validateExternalWorkloadExistence(g Gomega, shouldExist bool, k8sClient client.Reader, objectUID types.UID,
	namespace, kind string,
) {
	externalWorkloadOfObject := findExternalWorkloadOfObject(k8sClient,
		objectUID, namespace)
	if shouldExist {
		g.Expect(externalWorkloadOfObject).ToNot(BeNil(),
			"Should find external workload for object <%s/%s> of kind <%s>",
			objectUID, namespace, kind)
	} else {
		g.Expect(externalWorkloadOfObject).To(BeNil(),
			"Should not find external workload for object <%s/%s> of kind <%s>",
			objectUID, namespace, kind)
	}
}

func eventuallyExpectPodLabeledWithTopOwnerWorkloadId(pod *corev1.Pod, expectedTopOwnerWorkloadId, expectedClusterTopOwnerId string,
	k8sClient client.Reader,
) {
	Eventually(func(g Gomega) {
		apiPod := &corev1.Pod{}
		err := k8sClient.Get(context.Background(), types.NamespacedName{Name: pod.Name, Namespace: pod.Namespace}, apiPod)
		g.Expect(err).To(BeNil())

		topOwnerWorkloadId, found := apiPod.Labels[commonconsts.TopOwnerWorkloadIdLabel]
		g.Expect(found).To(BeTrue())
		g.Expect(topOwnerWorkloadId).To(Equal(expectedTopOwnerWorkloadId))

		topOwnerWorkloadId, found = apiPod.Labels[commonconsts.TopOwnerUidLabel]
		g.Expect(found).To(BeTrue())
		g.Expect(topOwnerWorkloadId).To(Equal(expectedTopOwnerWorkloadId))

		clusterTopOwnerId, clusterTopOwnerIdFound := apiPod.Labels[commonconsts.ClusterTopOwnerUidLabel]
		g.Expect(clusterTopOwnerIdFound).To(BeTrue())
		g.Expect(clusterTopOwnerId).To(Equal(expectedClusterTopOwnerId))
	}, timeout, interval).Should(Succeed())
}
