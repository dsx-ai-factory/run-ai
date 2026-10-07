/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package tests

import (
	"context"
	"testing"
	"time"

	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/constants"
	workloadTypeConfig "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-common/workload-type-config"
	kaiv2 "github.com/kai-scheduler/KAI-scheduler/pkg/apis/scheduling/v2"
	kaiv2alpha2 "github.com/kai-scheduler/KAI-scheduler/pkg/apis/scheduling/v2alpha2"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/json"

	dinfw "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-controller/pkg/api/distributedinferenceworkload/v1alpha1"
	infwv2 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-controller/pkg/api/inferenceworkload/v2alpha1"

	scheme_init "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/scheme-init"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1alpha1"
	diwv2 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-controller/pkg/api/distributedworkload/v2alpha1"
	iwv2a1 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-controller/pkg/api/interactiveworkload/v2alpha1"
	twv2a1 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-controller/pkg/api/trainingworkload/v2alpha1"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-exporter/pkg/controller/node"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-exporter/pkg/controller/pod"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-exporter/pkg/controller/podGroup"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-exporter/pkg/controller/workload"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-exporter/pkg/metrics"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-exporter/pkg/metrics/data"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-exporter/pkg/metrics/node_mig_exporter"
	kaiorg "github.com/kai-scheduler/kai-resource-management-api/kai/v1alpha1"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/client/interceptor"
	"sigs.k8s.io/controller-runtime/pkg/event"
)

// backfillGVK restores the GroupVersionKind (TypeMeta) on an object fetched from
// the fake client, which strips it. Production code dispatches on the object's
// GVK, so this keeps test behaviour aligned with a real client.
func backfillGVK(obj client.Object) {
	if !obj.GetObjectKind().GroupVersionKind().Empty() {
		return
	}
	if gvks, _, err := scheme.ObjectKinds(obj); err == nil && len(gvks) > 0 {
		obj.GetObjectKind().SetGroupVersionKind(gvks[0])
	}
}

var scheme *runtime.Scheme

var _ = BeforeSuite(func() {
	scheme = scheme_init.Scheme()
	Expect(corev1.AddToScheme(scheme)).Should(Succeed())
	Expect(v1alpha1.AddToScheme(scheme)).Should(Succeed())
	Expect(diwv2.AddToScheme(scheme)).Should(Succeed())
	Expect(iwv2a1.AddToScheme(scheme)).Should(Succeed())
	Expect(twv2a1.AddToScheme(scheme)).Should(Succeed())
	Expect(infwv2.AddToScheme(scheme)).Should(Succeed())
	Expect(dinfw.AddToScheme(scheme)).Should(Succeed())
	Expect(kaiv2.AddToScheme(scheme)).Should(Succeed())
	Expect(kaiv2alpha2.AddToScheme(scheme)).Should(Succeed())
	Expect(kaiorg.AddToScheme(scheme)).Should(Succeed())
})

func TestWorkloadExporter(t *testing.T) {
	zerolog.SetGlobalLevel(zerolog.DebugLevel)
	log.Info().Msgf("Started service tests")
	RegisterFailHandler(Fail)
	RunSpecs(t, "workload-exporter Test Suite")
}

type TestPodData struct {
	testName                string
	pod                     *corev1.Pod
	shouldCreatePod         bool
	expectedPodMetricLabels *ExpectedPodMetricLabels
	expectedPodMetricValues *ExpectedPodMetricValues

	expectedError bool
}

type ExpectedPodMetricLabels struct {
	name           string
	uuid           string
	namespace      string
	queue          string
	nodepool       string
	project        string
	department     string
	workloadId     string
	node           string
	phase          string
	detailedStatus string
}

type ExpectedPodMetricValues struct {
	requestedGpus      float64
	requestedGPUMemory float64
	limitGpus          float64
	limitGPUMemory     float64
	allocatedGpus      float64
	allocatedGPUMemory float64
	allocatedCpus      float64
	allocatedCpuMemory float64
}

type TestWorkloadData struct {
	testName                     string
	workload                     client.Object
	otherObjectsToCreate         []client.Object
	expectedWorkloadMetricLabels *ExpectedWorkloadMetricLabels
}

type ExpectedWorkloadMetricLabels struct {
	name             string
	uuid             string
	workloadType     data.MetricsWorkloadType
	workloadCategory string
	queue            string
	nodepool         string
	project          string
	department       string
	phase            string
	detailedStatus   string
	status           string
	user             string
	objectName       string
	namespace        string
	source           string
}

type TestNodeData struct {
	testName            string
	node                *corev1.Node
	expectedMigGpuCount float64
	shouldHaveMigMetric bool
}

type TestPodGroupData struct {
	testName                     string
	podGroup                     *kaiv2alpha2.PodGroup
	expectedPodGroupMetricLabels *ExpectedPodGroupMetricLabels
	expectedMigDevices           map[string]int64
	shouldHaveMetric             bool
}

type ExpectedPodGroupMetricLabels struct {
	queueName    string
	podGroupName string
	podGroupUUID string
	workloadName string
	workloadType string
	project      string
}

var expectedPodMetricValueNames = []string{"requestedGpus", "requestedGPUMemory", "limitGpus", "limitGPUMemory", "allocatedGpus", "allocatedGPUMemory", "allocatedCpus", "allocatedCpuMemory"}

func getWorkloadsConfigMapEntity() *corev1.ConfigMap {
	m := []workloadTypeConfig.WorkloadTypeConfig{
		{
			TypeName:     "JAXJob",
			Group:        "kubeflow.org",
			CategoryName: "Train",
			PriorityName: "",
		},
		{
			TypeName:     "JAXJob",
			Group:        "some-other-group.kubeflow.org",
			CategoryName: "Train2",
			PriorityName: "",
		},
		{
			TypeName:     "Notebook",
			Group:        "kubeflow.org",
			CategoryName: "Build",
			PriorityName: "",
		},
		{
			TypeName:     "Deployment",
			Group:        "apps",
			CategoryName: "Deploy",
			PriorityName: "",
		},
		{
			TypeName:     "Deployment",
			Group:        "apps.some-other-group",
			CategoryName: "Deploy222",
			PriorityName: "",
		},
		{
			TypeName:     "TrainingWorkload",
			Group:        "run.ai",
			CategoryName: "Train",
			PriorityName: "",
		},
		{
			TypeName:     "TrainingWorkload",
			Group:        "run.ai.some.other.group",
			CategoryName: "Train2222",
			PriorityName: "",
		},
		{
			TypeName:     "DistributedWorkload",
			Group:        "run.ai",
			CategoryName: "Train",
			PriorityName: "",
		},
		{
			TypeName:     "InteractiveWorkload",
			Group:        "run.ai",
			CategoryName: "Build",
			PriorityName: "",
		},
		{
			TypeName:     "InferenceWorkload",
			Group:        "run.ai",
			CategoryName: "Deploy",
			PriorityName: "",
		},
		{
			TypeName:     "DistributedInferenceWorkload",
			Group:        "run.ai",
			CategoryName: "Deploy",
			PriorityName: "",
		},
	}
	jsonM, _ := json.Marshal(m)
	cm := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      constants.WorkloadsConfigMapName,
			Namespace: constants.WorkloadsConfigMapNamespace,
		},
	}
	cm.Data = map[string]string{constants.WorkloadsConfigMapTypesKey: string(jsonM)}
	list := []string{"Train", "Deploy", "Inference", "Build"}
	categoriesJson, _ := json.Marshal(list)
	cm.Data[constants.WorkloadsConfigMapCategoriesKey] = string(categoriesJson)
	return cm
}

var _ = Describe("TestWorkloadExporter", Ordered, func() {
	var (
		k8sClient      client.Client
		podController  *pod.PodController
		nodeController *node.NodeController
		pgController   *podgroup.PodGroupController

		externalWorkloadController             *workload.ExternalWorkloadController
		trainingWorkloadController             *workload.TrainingWorkloadController
		interactiveWorkloadController          *workload.InteractiveWorkloadController
		distributedWorkloadController          *workload.DistributedWorkloadController
		inferenceWorkloadController            *workload.InferenceWorkloadController
		distributedInferenceWorkloadController *workload.DistributedInferenceWorkloadController

		ctx     context.Context
		cancel  context.CancelFunc
		stopper chan struct{}

		testCreatedObjects    = []client.Object{}
		podIdentifiersToReset = map[string]string{}
		workloadIdsToReset    = [][]string{}
		nodeNamesToReset      = []string{}
		podGroupNamesToReset  = map[string]string{}
	)

	BeforeAll(func() {
		k8sClient = fake.NewClientBuilder().WithScheme(scheme).WithObjects(
			TestProject, TestKaiProject, TestNamespace, TestNamespace2,
			TestNodePool, TestNodePool2, TestKaiNodePool, TestKaiNodePool2, TestQueue, TestQueue2,
			PriorityClassTrain, PriorityClassInteractivePreemtible,
			PriorityClassBuild, PriorityClassInference,
		).WithInterceptorFuncs(interceptor.Funcs{
			// The newer controller-runtime fake client strips TypeMeta (apiVersion/kind)
			// from objects returned by Get/List. Production code dispatches on the
			// workload's GroupVersionKind (e.g. to resolve the workload category), so
			// backfill the GVK from the scheme after the call.
			Get: func(ctx context.Context, c client.WithWatch, key client.ObjectKey, obj client.Object, opts ...client.GetOption) error {
				if err := c.Get(ctx, key, obj, opts...); err != nil {
					return err
				}
				backfillGVK(obj)
				return nil
			},
		}).Build()
		err := k8sClient.Create(ctx, getWorkloadsConfigMapEntity())
		Expect(err).To(BeNil())
		podController = pod.NewPodController(k8sClient)
		pgController = podgroup.NewPodGroupController(k8sClient)
		nodeController = node.NewNodeController(k8sClient)

		externalWorkloadController = workload.NewExternalWorkloadController(k8sClient)
		trainingWorkloadController = workload.NewTrainingWorkloadController(k8sClient)
		interactiveWorkloadController = workload.NewInteractiveWorkloadController(k8sClient)
		distributedWorkloadController = workload.NewDistributedWorkloadController(k8sClient)
		inferenceWorkloadController = workload.NewInferenceWorkloadController(k8sClient)
		distributedInferenceWorkloadController = workload.NewDistributedInferenceWorkloadController(k8sClient)
		stopper = make(chan struct{}, 1)
		ctx, cancel = context.WithCancel(context.Background())
		DeferCleanup(func() {
			close(stopper)
			cancel()
		})
	})

	AfterEach(func() {
		for _, testCreatedObject := range testCreatedObjects {
			err := k8sClient.Delete(ctx, testCreatedObject)
			if err != nil {
				Expect(errors.IsNotFound(err)).To(BeTrue())
			}
		}
		testCreatedObjects = []client.Object{}

		for name, namespace := range podIdentifiersToReset {
			metrics.ResetPodMetrics(name, namespace)
		}
		for _, identifier := range workloadIdsToReset {
			metrics.ResetWorkloadMetrics(identifier[0], identifier[1], metrics.WorkloadSource(identifier[2]))
		}
		for name, namespace := range podGroupNamesToReset {
			metrics.ResetPodGroupMetrics(name, namespace)
		}
		podGroupNamesToReset = map[string]string{}
		for _, nodeName := range nodeNamesToReset {
			metrics.ResetNodeMetrics(nodeName)
		}
		nodeNamesToReset = []string{}
	})

	DescribeTable("Pod metrics tests",
		func(testData *TestPodData) {
			err := k8sClient.Create(ctx, testData.pod)
			Expect(err).To(BeNil())
			testCreatedObjects = append(testCreatedObjects, testData.pod)
			podIdentifiersToReset[testData.pod.Name] = testData.pod.Namespace

			_, err = podController.Reconcile(ctx, getReconcileRequest(testData.pod.Name, testData.pod.Namespace))
			Expect(err).To(BeNil())

			podIdentifier := prometheus.Labels{
				metrics.PodNameLabel:    testData.pod.Name,
				metrics.PodUUIDLabel:    string(testData.pod.UID),
				metrics.NamespaceLabel:  testData.pod.Namespace,
				metrics.WorkloadIdLabel: testData.expectedPodMetricLabels.workloadId,
			}

			podMetrics := metrics.GetPodResourcesMetrics()
			exp := testData.expectedPodMetricLabels
			Eventually(func(g Gomega) {
				podInfoGauge, err := metrics.GetPodInfoMetric().GetMetricWithLabelValues(exp.name, exp.uuid, exp.namespace, exp.queue, exp.nodepool, exp.project, exp.department, exp.workloadId, exp.node, exp.phase, exp.detailedStatus)

				g.Expect(err).To(BeNil(), "Test: %s, Metric podInfo not error", testData.testName)
				g.Expect(podInfoGauge).ToNot(BeNil(), "Test: %s, Metric podInfo not nil", testData.testName)
				g.Expect(testutil.ToFloat64(podInfoGauge)).To(BeEquivalentTo(1), "Test: %s, Metric value podInfo", testData.testName)

				for i := range podMetrics {
					expectedMetricValue := getExpectedMetricValue(testData.expectedPodMetricValues, i)
					podMetric := podMetrics[i]
					podMetricGauge, err := podMetric.GetMetricWith(podIdentifier)
					g.Expect(err).To(BeNil(), "Test: %s, Metric not error: %s", testData.testName, expectedPodMetricValueNames[i])
					g.Expect(podMetricGauge).ToNot(BeNil(), "Test: %s, Metric not nil: %s", testData.testName, expectedPodMetricValueNames[i])
					g.Expect(testutil.ToFloat64(podMetricGauge)).To(Equal(expectedMetricValue), "Test: %s, Metric value: %s", testData.testName, expectedPodMetricValueNames[i])
				}
			}, 3*time.Second, 200*time.Millisecond).Should(Succeed())

			// delete the pod and Reconcile again
			err = k8sClient.Delete(ctx, testData.pod)
			Expect(err).To(BeNil())
			_, err = podController.Reconcile(ctx, getReconcileRequest(testData.pod.Name, testData.pod.Namespace))
			Expect(err).To(BeNil())

			podInfoGauge, _ := metrics.GetPodInfoMetric().GetMetricWithLabelValues(exp.name, exp.uuid, exp.namespace, exp.queue, exp.nodepool, exp.project, exp.department, exp.workloadId, exp.node, exp.phase, exp.detailedStatus)
			Expect(testutil.ToFloat64(podInfoGauge)).To(BeEquivalentTo(0), "Test: %s, Metric value podInfo should be 0 after deletion", testData.testName)

			for i := range podMetrics {
				podMetric := podMetrics[i]
				podMetricGauge, _ := podMetric.GetMetricWith(podIdentifier)
				Expect(testutil.ToFloat64(podMetricGauge)).To(BeEquivalentTo(0), "Test: %s, Metric value: %s, should be 0 after deletion", testData.testName, expectedPodMetricValueNames[i])
			}
		},
		Entry("Sanity - basic pod", &TestPodData{
			testName: "Sanity - basic pod",
			pod:      buildPod("test-pod", TestNamespaceName, "test-workload-id", TestNodeName, string(corev1.PodRunning)),
			expectedPodMetricLabels: &ExpectedPodMetricLabels{
				name:           "test-pod",
				uuid:           "test-pod",
				namespace:      TestNamespaceName,
				queue:          TestQueueName,
				nodepool:       TestNodePoolName,
				project:        TestProjectName,
				department:     TestDepartmentName,
				workloadId:     "test-workload-id",
				node:           TestNodeName,
				phase:          "Running",
				detailedStatus: string(corev1.PodRunning),
			},
			expectedPodMetricValues: TestPodExpectedPodMetricValues,
		}),
		Entry("Sanity - pending pod", &TestPodData{
			testName: "Sanity - pending pod",
			pod:      buildPod("pending-pod", TestNamespaceName, "pending-workload-id", "", string(corev1.PodPending)),
			expectedPodMetricLabels: &ExpectedPodMetricLabels{
				name:           "pending-pod",
				uuid:           "pending-pod",
				namespace:      TestNamespaceName,
				queue:          TestQueueName,
				nodepool:       TestNodePoolName,
				project:        TestProjectName,
				department:     TestDepartmentName,
				workloadId:     "pending-workload-id",
				node:           "",
				phase:          "Pending",
				detailedStatus: string(corev1.PodPending),
			},
			expectedPodMetricValues: &ExpectedPodMetricValues{
				requestedGpus:      1,
				requestedGPUMemory: 0,
				limitGpus:          1,
				limitGPUMemory:     0,
				allocatedGpus:      0,
				allocatedGPUMemory: 0,
				allocatedCpus:      0,
				allocatedCpuMemory: 0,
			},
		}),
		Entry("Sanity - suspended pod", &TestPodData{
			testName: "Sanity - suspended pod",
			pod:      buildPod("pending-pod", TestNamespaceName, "pending-workload-id", "", "Suspended"),
			expectedPodMetricLabels: &ExpectedPodMetricLabels{
				name:           "pending-pod",
				uuid:           "pending-pod",
				namespace:      TestNamespaceName,
				queue:          TestQueueName,
				nodepool:       TestNodePoolName,
				project:        TestProjectName,
				department:     TestDepartmentName,
				workloadId:     "pending-workload-id",
				node:           "",
				phase:          "Suspended",
				detailedStatus: "Suspended",
			},
			expectedPodMetricValues: &ExpectedPodMetricValues{
				requestedGpus:      1,
				requestedGPUMemory: 0,
				limitGpus:          1,
				limitGPUMemory:     0,
				allocatedGpus:      0,
				allocatedGPUMemory: 0,
				allocatedCpus:      0,
				allocatedCpuMemory: 0,
			},
		}),
		Entry("Sanity - running pod in nodepool", &TestPodData{
			testName: "Sanity - running pod in nodepool",
			pod:      buildPodWithNodePool("test-pod-np", TestNamespaceName, "test-workload-id-np", TestNodeName, string(corev1.PodRunning)),
			expectedPodMetricLabels: &ExpectedPodMetricLabels{
				name:           "test-pod-np",
				uuid:           "test-pod-np",
				namespace:      TestNamespaceName,
				queue:          TestQueueName2,
				nodepool:       TestNodePoolName2,
				project:        TestProjectName,
				department:     TestDepartmentName,
				workloadId:     "test-workload-id-np",
				node:           TestNodeName,
				phase:          "Running",
				detailedStatus: string(corev1.PodRunning),
			},
			expectedPodMetricValues: TestPodExpectedPodMetricValues,
		}),
		Entry("Sanity - pending pod in nodepool", &TestPodData{
			testName: "Sanity - pending pod in nodepool",
			pod:      buildPodWithNodePool("pending-pod2", TestNamespaceName, "pending-workload-id2", "", string(corev1.PodPending)),
			expectedPodMetricLabels: &ExpectedPodMetricLabels{
				name:           "pending-pod2",
				uuid:           "pending-pod2",
				namespace:      TestNamespaceName,
				queue:          TestQueueName2,
				nodepool:       TestNodePoolName2,
				project:        TestProjectName,
				department:     TestDepartmentName,
				workloadId:     "pending-workload-id2",
				node:           "",
				phase:          "Pending",
				detailedStatus: string(corev1.PodPending),
			},
			expectedPodMetricValues: &ExpectedPodMetricValues{
				requestedGpus:      1,
				requestedGPUMemory: 0,
				limitGpus:          1,
				limitGPUMemory:     0,
				allocatedGpus:      0,
				allocatedGPUMemory: 0,
				allocatedCpus:      0,
				allocatedCpuMemory: 0,
			},
		}),
	)

	DescribeTable("Pod metrics tests - with error",
		func(testData *TestPodData) {
			if testData.shouldCreatePod {
				err := k8sClient.Create(ctx, testData.pod)
				Expect(err).To(BeNil())
				testCreatedObjects = append(testCreatedObjects, testData.pod)
				podIdentifiersToReset[testData.pod.Name] = testData.pod.Namespace
			}

			result, err := podController.Reconcile(ctx, getReconcileRequest(testData.pod.Name, testData.pod.Namespace))
			if testData.expectedError {
				Expect(err).To(HaveOccurred(), "Test: %s, expected an error", testData.testName)
			} else {
				Expect(err).To(BeNil(), "Test: %s, unexpected error", testData.testName)
			}
			Expect(result.RequeueAfter).To(BeZero(),
				"Test: %s, the rate limiter must decide the retry delay", testData.testName)

			podIdentifier := prometheus.Labels{
				metrics.PodNameLabel:    testData.pod.Name,
				metrics.PodUUIDLabel:    string(testData.pod.UID),
				metrics.NamespaceLabel:  testData.pod.Namespace,
				metrics.WorkloadIdLabel: testData.expectedPodMetricLabels.workloadId,
			}

			Eventually(func(g Gomega) {
				podMetrics := metrics.GetPodResourcesMetrics()
				for i := range podMetrics {
					podMetric := podMetrics[i]
					podMetricGauge, _ := podMetric.GetMetricWith(podIdentifier)

					g.Expect(testutil.ToFloat64(podMetricGauge)).To(BeEquivalentTo(0), "Test: %s, Metric shouldn't exist (validating value is 0): %s", testData.testName, expectedPodMetricValueNames[i])
				}
			}, 3*time.Second, 200*time.Millisecond).Should(Succeed())
		},
		Entry("Pod in unexisting project - metrics shouldn't exist", &TestPodData{
			testName: "Pod in unexisting project - metrics shouldn't exist",
			pod:      buildPod("test-pod2", TestNamespaceName2, "test-workload-id2", TestNodeName, string(corev1.PodRunning)),
			expectedPodMetricLabels: &ExpectedPodMetricLabels{
				workloadId: "test-workload-id2",
			},
			shouldCreatePod: true,
			expectedError:   true,
		}),
		Entry("Unexisting pod - metrics shouldn't exist", &TestPodData{
			testName: "Unexisting pod - metrics shouldn't exist",
			pod:      buildPod("test-pod3", TestNamespaceName, "test-workload-id3", TestNodeName, string(corev1.PodRunning)),
			expectedPodMetricLabels: &ExpectedPodMetricLabels{
				workloadId: "test-workload-id3",
			},
			shouldCreatePod: false,
			expectedError:   false,
		}),
	)

	It("Pod elements metric - pod with element IDs", func() {
		pod := buildPodWithElementIDs("elem-pod", TestNamespaceName, "elem-wl-id", TestNodeName, string(corev1.PodRunning), "e1-uuid", "e2-uuid", "")
		err := k8sClient.Create(ctx, pod)
		Expect(err).To(BeNil())
		testCreatedObjects = append(testCreatedObjects, pod)
		podIdentifiersToReset[pod.Name] = pod.Namespace

		_, err = podController.Reconcile(ctx, getReconcileRequest(pod.Name, pod.Namespace))
		Expect(err).To(BeNil())

		Eventually(func(g Gomega) {
			gauge1, err := metrics.GetPodElementsMetric().GetMetricWithLabelValues(pod.Name, pod.Namespace, string(pod.UID), "elem-wl-id", "e1-uuid", "1")
			g.Expect(err).To(BeNil())
			g.Expect(testutil.ToFloat64(gauge1)).To(BeEquivalentTo(1))

			gauge2, err := metrics.GetPodElementsMetric().GetMetricWithLabelValues(pod.Name, pod.Namespace, string(pod.UID), "elem-wl-id", "e2-uuid", "2")
			g.Expect(err).To(BeNil())
			g.Expect(testutil.ToFloat64(gauge2)).To(BeEquivalentTo(1))
		}, 3*time.Second, 200*time.Millisecond).Should(Succeed())

		// delete pod and verify cleanup
		err = k8sClient.Delete(ctx, pod)
		Expect(err).To(BeNil())
		_, err = podController.Reconcile(ctx, getReconcileRequest(pod.Name, pod.Namespace))
		Expect(err).To(BeNil())

		gauge1, _ := metrics.GetPodElementsMetric().GetMetricWithLabelValues(pod.Name, pod.Namespace, string(pod.UID), "elem-wl-id", "e1-uuid", "1")
		Expect(testutil.ToFloat64(gauge1)).To(BeEquivalentTo(0))
	})

	It("Pod elements metric - pod without element IDs", func() {
		pod := buildPod("no-elem-pod", TestNamespaceName, "no-elem-wl-id", TestNodeName, string(corev1.PodRunning))
		err := k8sClient.Create(ctx, pod)
		Expect(err).To(BeNil())
		testCreatedObjects = append(testCreatedObjects, pod)
		podIdentifiersToReset[pod.Name] = pod.Namespace

		_, err = podController.Reconcile(ctx, getReconcileRequest(pod.Name, pod.Namespace))
		Expect(err).To(BeNil())

		// pod_elements should have no series for this pod
		count := testutil.CollectAndCount(metrics.GetPodElementsMetric())
		Expect(count).To(Equal(0), "Pod without element labels should not have pod_elements series")
	})

	DescribeTable("Workload metrics tests",
		func(testData *TestWorkloadData) {
			err := k8sClient.Create(ctx, testData.workload)
			Expect(err).To(BeNil())
			testCreatedObjects = append(testCreatedObjects, testData.workload)
			workloadIdsToReset = append(workloadIdsToReset, []string{testData.workload.GetName(), testData.workload.GetNamespace(), testData.expectedWorkloadMetricLabels.source})

			for _, otherObject := range testData.otherObjectsToCreate {
				err := k8sClient.Create(ctx, otherObject)
				Expect(err).To(BeNil())
				testCreatedObjects = append(testCreatedObjects, otherObject)
			}

			err = ReconcileTestWorkload(ctx, testData.workload, externalWorkloadController, trainingWorkloadController, interactiveWorkloadController, distributedWorkloadController, inferenceWorkloadController, distributedInferenceWorkloadController)
			Expect(err).To(BeNil())

			exp := testData.expectedWorkloadMetricLabels
			Eventually(func(g Gomega) {
				workloadInfoGauge, err := metrics.GetWorkloadInfoMetric().GetMetricWithLabelValues(exp.name, exp.uuid, string(exp.workloadType), exp.workloadCategory,
					exp.queue, exp.nodepool, exp.project, exp.department, exp.phase,
					exp.detailedStatus, exp.status, exp.user, exp.objectName, exp.namespace, exp.source)

				g.Expect(err).To(BeNil(), "Test: %s, Metric workloadInfo not error", testData.testName)
				g.Expect(workloadInfoGauge).ToNot(BeNil(), "Test: %s, Metric workloadInfo not nil", testData.testName)
				g.Expect(testutil.ToFloat64(workloadInfoGauge)).To(BeEquivalentTo(1), "Test: %s, Metric value workloadInfo", testData.testName)
			}, 3*time.Second, 200*time.Millisecond).Should(Succeed())

			// delete the workload and Reconcile again
			err = k8sClient.Delete(ctx, testData.workload)
			Expect(err).To(BeNil())
			err = ReconcileTestWorkload(ctx, testData.workload, externalWorkloadController, trainingWorkloadController, interactiveWorkloadController, distributedWorkloadController, inferenceWorkloadController, distributedInferenceWorkloadController)
			Expect(err).To(BeNil())

			workloadInfoGauge, _ := metrics.GetWorkloadInfoMetric().GetMetricWithLabelValues(exp.name, exp.uuid, string(exp.workloadType), exp.workloadCategory,
				exp.queue, exp.nodepool, exp.project, exp.department, exp.phase,
				exp.detailedStatus, exp.status, exp.user, exp.objectName, exp.namespace, exp.source)
			Expect(testutil.ToFloat64(workloadInfoGauge)).To(BeEquivalentTo(0), "Test: %s, Metric value workloadInfo should be 0 after deletion", testData.testName)
		},
		Entry("Sanity - external workload", &TestWorkloadData{
			testName: "Sanity - external workload",
			workload: buildExternalWorkload("test-ew", TestNamespaceName, "test-workload-id-ew", "id1", "", "JAXJob", "kubeflow.org", v1alpha1.ExWRunningPhase, []string{}),
			otherObjectsToCreate: []client.Object{
				buildPodGroup("test-pg-ew", TestNamespaceName, "train", "Running"),
				buildPodOfWorkload("test-pod-ew", TestNamespaceName, "id1", TestNodeName, string(corev1.PodRunning), "test-pg-ew"),
			},
			expectedWorkloadMetricLabels: &ExpectedWorkloadMetricLabels{
				name:             "test-workload-id-ew",
				uuid:             "id1",
				workloadType:     data.JobTypeTrain,
				workloadCategory: "Train",
				queue:            TestQueueName,
				nodepool:         TestNodePoolName,
				project:          TestProjectName,
				department:       TestDepartmentName,
				phase:            "Running",
				detailedStatus:   "Running",
				status:           string(v1alpha1.ExWRunningPhase),
				user:             "",
				objectName:       "test-ew",
				namespace:        TestNamespaceName,
				source:           string(metrics.ExternalSource),
			},
		}),
		Entry("Sanity - training workload", &TestWorkloadData{
			testName: "Sanity - training workload",
			workload: buildTrainingWorkload("test-train", TestNamespaceName, "test-workload-id-train", "", "Running", "user1"),
			otherObjectsToCreate: []client.Object{
				buildPodGroup("test-pg-train", TestNamespaceName, "train", "Running"),
				buildPodOfWorkload("test-pod-train", TestNamespaceName, "test-workload-id-train", TestNodeName, string(corev1.PodRunning), "test-pg-train"),
			},
			expectedWorkloadMetricLabels: &ExpectedWorkloadMetricLabels{
				name:             "test-train",
				uuid:             "test-workload-id-train",
				workloadType:     data.JobTypeTrain,
				workloadCategory: "Train",
				queue:            TestQueueName,
				nodepool:         TestNodePoolName,
				project:          TestProjectName,
				department:       TestDepartmentName,
				phase:            "Running",
				detailedStatus:   "Running",
				status:           "Running",
				user:             "user1",
				objectName:       "test-train",
				namespace:        TestNamespaceName,
				source:           string(metrics.TrainingSource),
			},
		}),
		Entry("Sanity - distributed workload", &TestWorkloadData{
			testName: "Sanity - distributed workload",
			workload: buildDistributedWorkload("test-dist", TestNamespaceName, "test-workload-id-dist", "", "Running", "user3"),
			otherObjectsToCreate: []client.Object{
				buildPodGroup("test-pg-dist", TestNamespaceName, "train", "Running"),
				buildPodOfWorkload("test-pod-dist", TestNamespaceName, "test-workload-id-dist", TestNodeName, string(corev1.PodRunning), "test-pg-dist"),
			},
			expectedWorkloadMetricLabels: &ExpectedWorkloadMetricLabels{
				name:             "test-dist",
				uuid:             "test-workload-id-dist",
				workloadType:     data.JobTypeTrain,
				workloadCategory: "Train",
				queue:            TestQueueName,
				nodepool:         TestNodePoolName,
				project:          TestProjectName,
				department:       TestDepartmentName,
				phase:            "Running",
				detailedStatus:   "Running",
				status:           "Running",
				user:             "user3",
				objectName:       "test-dist",
				namespace:        TestNamespaceName,
				source:           string(metrics.DistributedSource),
			},
		}),
		Entry("Sanity - interactive workload", &TestWorkloadData{
			testName: "Sanity - interactive workload",
			workload: buildInteractiveWorkload("test-int", TestNamespaceName, "test-workload-id-int", "", "Running", "user4", false),
			otherObjectsToCreate: []client.Object{
				buildPodGroup("test-pg-int", TestNamespaceName, "build", "Running"),
				buildPodOfWorkload("test-pod-int", TestNamespaceName, "test-workload-id-int", TestNodeName, string(corev1.PodRunning), "test-pg-int"),
			},
			expectedWorkloadMetricLabels: &ExpectedWorkloadMetricLabels{
				name:             "test-int",
				uuid:             "test-workload-id-int",
				workloadType:     data.JobTypeInteractive,
				workloadCategory: "Build",
				queue:            TestQueueName,
				nodepool:         TestNodePoolName,
				project:          TestProjectName,
				department:       TestDepartmentName,
				phase:            "Running",
				detailedStatus:   "Running",
				status:           "Running",
				user:             "user4",
				objectName:       "test-int",
				namespace:        TestNamespaceName,
				source:           string(metrics.InteractiveSource),
			},
		}),
		Entry("Sanity - inference workload", &TestWorkloadData{
			testName: "Sanity - inference workload",
			workload: buildInferenceWorkload("test-inf", TestNamespaceName, "test-workload-id-inf", "", "Running", "user1"),
			otherObjectsToCreate: []client.Object{
				buildPodGroup("test-pg-inf", TestNamespaceName, "inference", "Running"),
				buildPodOfWorkload("test-pod-inf", TestNamespaceName, "test-workload-id-inf", TestNodeName, string(corev1.PodRunning), "test-pg-inf"),
			},
			expectedWorkloadMetricLabels: &ExpectedWorkloadMetricLabels{
				name:             "test-inf",
				uuid:             "test-workload-id-inf",
				workloadType:     data.JobTypeInference,
				workloadCategory: "Deploy",
				queue:            TestQueueName,
				nodepool:         TestNodePoolName,
				project:          TestProjectName,
				department:       TestDepartmentName,
				phase:            "Running",
				detailedStatus:   "Running",
				status:           "Running",
				user:             "user1",
				objectName:       "test-inf",
				namespace:        TestNamespaceName,
				source:           string(metrics.InferenceSource),
			},
		}),
		Entry("Sanity - distributed inference workload", &TestWorkloadData{
			testName: "Sanity - distributed inference workload",
			workload: buildDistributedInferenceWorkload("test-dist-inf", TestNamespaceName, "test-workload-id-dist-inf", "", "Running", "user1"),
			otherObjectsToCreate: []client.Object{
				buildPodGroup("test-pg-dist-inf", TestNamespaceName, "inference", "Running"),
				buildPodOfWorkload("test-pod-dist-inf", TestNamespaceName, "test-workload-id-dist-inf", TestNodeName, string(corev1.PodRunning), "test-pg-dist-inf"),
			},
			expectedWorkloadMetricLabels: &ExpectedWorkloadMetricLabels{
				name:             "test-dist-inf",
				uuid:             "test-workload-id-dist-inf",
				workloadType:     data.JobTypeDistributedInference,
				workloadCategory: "Deploy",
				queue:            TestQueueName,
				nodepool:         TestNodePoolName,
				project:          TestProjectName,
				department:       TestDepartmentName,
				phase:            "Running",
				detailedStatus:   "Running",
				status:           "Running",
				user:             "user1",
				objectName:       "test-dist-inf",
				namespace:        TestNamespaceName,
				source:           string(metrics.DistributedInferenceSource),
			},
		}),
		Entry("External workload with nodepool", &TestWorkloadData{
			testName: "External workload with nodepool",
			workload: buildExternalWorkload("test-ew", TestNamespaceName, "test-workload-ew-name", "id1", "id2", "Deployment", "apps", v1alpha1.ExWRunningPhase, []string{"unexisting-np1", "unexisting-np2"}),
			otherObjectsToCreate: []client.Object{
				buildPodGroup("test-pg-ew", TestNamespaceName, "train", "Running"),
				buildPodOfWorkloadInNodePool("test-pod-ew", TestNamespaceName, "id2", TestNodeName, string(corev1.PodRunning), "test-pg-ew"),
			},
			expectedWorkloadMetricLabels: &ExpectedWorkloadMetricLabels{
				name:             "test-workload-ew-name",
				uuid:             "id2",
				workloadType:     data.JobTypeTrain,
				workloadCategory: "Deploy",
				queue:            TestQueueName2,
				nodepool:         TestNodePoolName2,
				project:          TestProjectName,
				department:       TestDepartmentName,
				phase:            "Running",
				detailedStatus:   "Running",
				status:           string(v1alpha1.ExWRunningPhase),
				user:             "",
				objectName:       "test-ew",
				namespace:        TestNamespaceName,
				source:           string(metrics.ExternalSource),
			},
		}),
		Entry("External workload with requested nodepool", &TestWorkloadData{
			testName: "External workload with requested nodepool",
			workload: buildExternalWorkload("test-ew", TestNamespaceName, "test-workload-ew-name", "id1", "id2", "Notebook", "kubeflow.org", v1alpha1.ExWRunningPhase, []string{TestNodePoolName2}),
			otherObjectsToCreate: []client.Object{
				buildPodGroup("test-pg-ew", TestNamespaceName, "train", "Running"),
				buildPodOfWorkload("test-pod-ew", TestNamespaceName, "id2", TestNodeName, string(corev1.PodRunning), "test-pg-ew"),
			},
			expectedWorkloadMetricLabels: &ExpectedWorkloadMetricLabels{
				name:             "test-workload-ew-name",
				uuid:             "id2",
				workloadType:     data.JobTypeTrain,
				workloadCategory: "Build",
				queue:            TestQueueName2,
				nodepool:         TestNodePoolName2,
				project:          TestProjectName,
				department:       TestDepartmentName,
				phase:            "Running",
				detailedStatus:   "Running",
				status:           string(v1alpha1.ExWRunningPhase),
				user:             "",
				objectName:       "test-ew",
				namespace:        TestNamespaceName,
				source:           string(metrics.ExternalSource),
			},
		}),
		Entry("Initializing training workload", &TestWorkloadData{
			testName: "Initializing training workload",
			workload: buildTrainingWorkload("test-train", TestNamespaceName, "test-workload-id-train", "test-workload-id-train2", "Initializing", "user2"),
			otherObjectsToCreate: []client.Object{
				buildPodGroup("test-pg-train", TestNamespaceName, "train", "Pending"),
				buildPodOfWorkload("test-pod-train", TestNamespaceName, "test-workload-id-train2", "", "", "test-pg-train"),
			},
			expectedWorkloadMetricLabels: &ExpectedWorkloadMetricLabels{
				name:             "test-train",
				uuid:             "test-workload-id-train2",
				workloadType:     data.JobTypeTrain,
				workloadCategory: "Train",
				queue:            TestQueueName,
				nodepool:         TestNodePoolName,
				project:          TestProjectName,
				department:       TestDepartmentName,
				phase:            "Running",
				detailedStatus:   "Initializing",
				status:           "Initializing",
				user:             "user2",
				objectName:       "test-train",
				namespace:        TestNamespaceName,
				source:           string(metrics.TrainingSource),
			},
		}),
		Entry("Pending training workload", &TestWorkloadData{
			testName: "Pending training workload",
			workload: buildTrainingWorkload("test-train", TestNamespaceName, "test-workload-id-train", "test-workload-id-train2", "Pending", "user2"),
			otherObjectsToCreate: []client.Object{
				buildPodGroup("test-pg-train", TestNamespaceName, "train", "Pending"),
				buildPodOfWorkload("test-pod-train", TestNamespaceName, "test-workload-id-train2", "", "", "test-pg-train"),
			},
			expectedWorkloadMetricLabels: &ExpectedWorkloadMetricLabels{
				name:             "test-train",
				uuid:             "test-workload-id-train2",
				workloadType:     data.JobTypeTrain,
				workloadCategory: "Train",
				queue:            TestQueueName,
				nodepool:         TestNodePoolName,
				project:          TestProjectName,
				department:       TestDepartmentName,
				phase:            "Pending",
				detailedStatus:   "Pending",
				status:           "Pending",
				user:             "user2",
				objectName:       "test-train",
				namespace:        TestNamespaceName,
				source:           string(metrics.TrainingSource),
			},
		}),
		Entry("Distributed workload without podgroups", &TestWorkloadData{
			testName: "Distributed workload without podgroups",
			workload: buildDistributedWorkload("test-dist", TestNamespaceName, "test-workload-id-dist", "test-workload-id-dist2", "Degraded", "user5"),
			otherObjectsToCreate: []client.Object{
				buildPodOfWorkload("test-pod-dist", TestNamespaceName, "test-workload-id-dist2", TestNodeName, string(corev1.PodRunning), "test-pg-dist"),
			},
			expectedWorkloadMetricLabels: &ExpectedWorkloadMetricLabels{
				name:             "test-dist",
				uuid:             "test-workload-id-dist2",
				workloadType:     data.JobTypeTrain,
				workloadCategory: "Train",
				queue:            TestQueueName,
				nodepool:         TestNodePoolName,
				project:          TestProjectName,
				department:       TestDepartmentName,
				phase:            "Running",
				detailedStatus:   "Degraded",
				status:           "Degraded",
				user:             "user5",
				objectName:       "test-dist",
				namespace:        TestNamespaceName,
				source:           string(metrics.DistributedSource),
			},
		}),
		Entry("Interactive workload without pods", &TestWorkloadData{
			testName:             "Interactive workload without pods",
			workload:             buildInteractiveWorkload("test-int", TestNamespaceName, "test-workload-id-int", "test-workload-id-int2", "Failed", "user8", false),
			otherObjectsToCreate: []client.Object{},
			expectedWorkloadMetricLabels: &ExpectedWorkloadMetricLabels{
				name:             "test-int",
				uuid:             "test-workload-id-int2",
				workloadType:     data.JobTypeInteractive,
				workloadCategory: "Build",
				queue:            "",
				nodepool:         "",
				project:          TestProjectName,
				department:       TestDepartmentName,
				phase:            "",
				detailedStatus:   "Failed",
				status:           "Failed",
				user:             "user8",
				objectName:       "test-int",
				namespace:        TestNamespaceName,
				source:           string(metrics.InteractiveSource),
			},
		}),
		Entry("Interactive workload without podgroups", &TestWorkloadData{
			testName: "Interactive workload without podgroups",
			workload: buildInteractiveWorkload("test-int", TestNamespaceName, "test-workload-id-int", "", "Failed", "", false),
			otherObjectsToCreate: []client.Object{
				buildPodOfWorkload("test-pod-int", TestNamespaceName, "test-workload-id-int", TestNodeName, string(corev1.PodRunning), "test-pg-int"),
			},
			expectedWorkloadMetricLabels: &ExpectedWorkloadMetricLabels{
				name:             "test-int",
				uuid:             "test-workload-id-int",
				workloadType:     data.JobTypeInteractive,
				workloadCategory: "Build",
				queue:            TestQueueName,
				nodepool:         TestNodePoolName,
				project:          TestProjectName,
				department:       TestDepartmentName,
				phase:            "",
				detailedStatus:   "Failed",
				status:           "Failed",
				user:             "",
				objectName:       "test-int",
				namespace:        TestNamespaceName,
				source:           string(metrics.InteractiveSource),
			},
		}),
		Entry("Interactive workload interactive-preemtible (without pods)", &TestWorkloadData{
			testName:             "Interactive workload interactive-preemtible (without pods)",
			workload:             buildInteractiveWorkload("test-int", TestNamespaceName, "test-workload-id-int", "test-workload-id-int2", "Failed", "user8", true),
			otherObjectsToCreate: []client.Object{},
			expectedWorkloadMetricLabels: &ExpectedWorkloadMetricLabels{
				name:             "test-int",
				uuid:             "test-workload-id-int2",
				workloadType:     data.JobTypeInteractivePreemptible,
				workloadCategory: "Build",
				queue:            "",
				nodepool:         "",
				project:          TestProjectName,
				department:       TestDepartmentName,
				phase:            "",
				detailedStatus:   "Failed",
				status:           "Failed",
				user:             "user8",
				objectName:       "test-int",
				namespace:        TestNamespaceName,
				source:           string(metrics.InteractiveSource),
			},
		}),
	)

	DescribeTable("Node metrics tests - MIG mode GPU count",
		func(testData *TestNodeData) {
			err := k8sClient.Create(ctx, testData.node)
			Expect(err).To(BeNil())
			testCreatedObjects = append(testCreatedObjects, testData.node)
			nodeNamesToReset = append(nodeNamesToReset, testData.node.Name)

			// Use event-based triggering - only reconcile if the event filter allows it
			eventFilter := node.MigNodesEventFilter()
			createEvent := event.CreateEvent{Object: testData.node}
			if eventFilter.Create(createEvent) {
				_, err = nodeController.Reconcile(ctx, getReconcileRequest(testData.node.Name, ""))
				Expect(err).To(BeNil())
			}

			migMetric := metrics.GetMigModeGPUsMetric()
			Eventually(func(g Gomega) {
				migGpuGauge, err := migMetric.GetMetricWithLabelValues(testData.node.Name, testData.node.Name)
				g.Expect(err).To(BeNil(), "Test: %s, Metric migModeGPUs not error", testData.testName)

				if testData.shouldHaveMigMetric {
					g.Expect(testutil.ToFloat64(migGpuGauge)).To(Equal(testData.expectedMigGpuCount),
						"Test: %s, Metric value migModeGPUs", testData.testName)
				} else {
					g.Expect(testutil.ToFloat64(migGpuGauge)).To(BeEquivalentTo(0),
						"Test: %s, Metric migModeGPUs should be 0 for non-MIG node", testData.testName)
				}
			}, 3*time.Second, 200*time.Millisecond).Should(Succeed())

			// delete the node and Reconcile again (only if filter allows)
			err = k8sClient.Delete(ctx, testData.node)
			Expect(err).To(BeNil())
			deleteEvent := event.DeleteEvent{Object: testData.node}
			if eventFilter.Delete(deleteEvent) {
				_, err = nodeController.Reconcile(ctx, getReconcileRequest(testData.node.Name, ""))
				Expect(err).To(BeNil())
			}

			migGpuGauge, _ := migMetric.GetMetricWithLabelValues(testData.node.Name, testData.node.Name)
			Expect(testutil.ToFloat64(migGpuGauge)).To(BeEquivalentTo(0),
				"Test: %s, Metric value migModeGPUs should be 0 after deletion", testData.testName)
		},
		Entry("MIG node with mixed strategy - should report GPU count", &TestNodeData{
			testName:            "MIG node with mixed strategy",
			node:                buildMigNode("mig-node-mixed", 4, node_mig_exporter.MigStrategyMixed),
			expectedMigGpuCount: 4,
			shouldHaveMigMetric: true,
		}),
		Entry("MIG node with single strategy - should report GPU count", &TestNodeData{
			testName:            "MIG node with single strategy",
			node:                buildMigNode("mig-node-single", 8, node_mig_exporter.MigStrategySingle),
			expectedMigGpuCount: 8,
			shouldHaveMigMetric: true,
		}),
		Entry("Non-MIG node - should not report MIG GPU count", &TestNodeData{
			testName:            "Non-MIG node",
			node:                buildNonMigNode("non-mig-node", 2),
			expectedMigGpuCount: 0,
			shouldHaveMigMetric: false,
		}),
		Entry("MIG node with zero GPUs - should report zero", &TestNodeData{
			testName:            "MIG node with zero GPUs",
			node:                buildMigNode("mig-node-zero-gpu", 0, node_mig_exporter.MigStrategyMixed),
			expectedMigGpuCount: 0,
			shouldHaveMigMetric: true,
		}),
	)

	DescribeTable("Node metrics tests - unready nodes",
		func(testData *TestNodeData) {
			err := k8sClient.Create(ctx, testData.node)
			Expect(err).To(BeNil())
			testCreatedObjects = append(testCreatedObjects, testData.node)
			nodeNamesToReset = append(nodeNamesToReset, testData.node.Name)

			// Use event-based triggering - only reconcile if the event filter allows it
			eventFilter := node.MigNodesEventFilter()
			createEvent := event.CreateEvent{Object: testData.node}
			if eventFilter.Create(createEvent) {
				_, err = nodeController.Reconcile(ctx, getReconcileRequest(testData.node.Name, ""))
				Expect(err).To(BeNil())
			}

			migMetric := metrics.GetMigModeGPUsMetric()
			Eventually(func(g Gomega) {
				migGpuGauge, err := migMetric.GetMetricWithLabelValues(testData.node.Name, testData.node.Name)
				g.Expect(err).To(BeNil(), "Test: %s, Metric migModeGPUs not error", testData.testName)
				g.Expect(testutil.ToFloat64(migGpuGauge)).To(BeEquivalentTo(0),
					"Test: %s, Metric migModeGPUs should be 0 for unready node", testData.testName)
			}, 3*time.Second, 200*time.Millisecond).Should(Succeed())
		},
		Entry("Unready MIG node - should not report GPU count", &TestNodeData{
			testName:            "Unready MIG node",
			node:                buildUnreadyNode("unready-mig-node", 4, true, node_mig_exporter.MigStrategyMixed),
			expectedMigGpuCount: 0,
			shouldHaveMigMetric: false,
		}),
	)

	DescribeTable("PodGroup MIG metrics tests",
		func(testData *TestPodGroupData) {
			err := k8sClient.Create(ctx, testData.podGroup)
			Expect(err).To(BeNil())
			testCreatedObjects = append(testCreatedObjects, testData.podGroup)
			podGroupNamesToReset[testData.podGroup.Name] = testData.podGroup.Namespace

			_, err = pgController.Reconcile(ctx, getReconcileRequest(testData.podGroup.Name, testData.podGroup.Namespace))
			Expect(err).To(BeNil())

			migMetric := metrics.GetPodGroupRequestedMigDevicesMetric()

			if testData.shouldHaveMetric {
				exp := testData.expectedPodGroupMetricLabels
				Eventually(func(g Gomega) {
					for deviceType, expectedCount := range testData.expectedMigDevices {
						gauge, err := migMetric.GetMetricWithLabelValues(
							exp.queueName, exp.podGroupName, exp.podGroupUUID,
							exp.workloadName, exp.workloadType, deviceType, exp.project)
						g.Expect(err).To(BeNil(), "Test: %s, Metric error for device %s", testData.testName, deviceType)
						g.Expect(testutil.ToFloat64(gauge)).To(BeEquivalentTo(expectedCount),
							"Test: %s, Metric value for device %s", testData.testName, deviceType)
					}
				}, 3*time.Second, 200*time.Millisecond).Should(Succeed())
			}

			// delete the podGroup and Reconcile again
			err = k8sClient.Delete(ctx, testData.podGroup)
			Expect(err).To(BeNil())
			_, err = pgController.Reconcile(ctx, getReconcileRequest(testData.podGroup.Name, testData.podGroup.Namespace))
			Expect(err).To(BeNil())

			// Verify metrics are reset after deletion
			if testData.shouldHaveMetric {
				exp := testData.expectedPodGroupMetricLabels
				for deviceType := range testData.expectedMigDevices {
					gauge, _ := migMetric.GetMetricWithLabelValues(
						exp.queueName, exp.podGroupName, exp.podGroupUUID,
						exp.workloadName, exp.workloadType, deviceType, exp.project)
					Expect(testutil.ToFloat64(gauge)).To(BeEquivalentTo(0),
						"Test: %s, Metric value for device %s should be 0 after deletion", testData.testName, deviceType)
				}
			}
		},
		Entry("PodGroup with MIG devices - should report requested devices", &TestPodGroupData{
			testName: "PodGroup with MIG devices",
			podGroup: buildPodGroupWithMigDevices("test-pg-mig", TestNamespaceName, TestQueueName,
				"test-workload", "test-workload-uid", map[string]int64{"1g.5gb": 2, "2g.10gb": 1}),
			expectedPodGroupMetricLabels: &ExpectedPodGroupMetricLabels{
				queueName:    TestQueueName,
				podGroupName: "test-pg-mig",
				podGroupUUID: "test-workload-uid",
				workloadName: "test-workload",
				workloadType: string(data.JobTypeTrain),
				project:      TestProjectName,
			},
			expectedMigDevices: map[string]int64{"1g.5gb": 2, "2g.10gb": 1},
			shouldHaveMetric:   true,
		}),
		Entry("PodGroup without MIG annotation - should not report metric", &TestPodGroupData{
			testName:         "PodGroup without MIG annotation",
			podGroup:         buildPodGroupWithoutMigDevices("test-pg-no-mig", TestNamespaceName, TestQueueName, "test-workload-2", "test-workload-uid-2"),
			shouldHaveMetric: false,
		}),
		Entry("PodGroup with empty MIG devices - should not report metric", &TestPodGroupData{
			testName: "PodGroup with empty MIG devices",
			podGroup: buildPodGroupWithMigDevices("test-pg-empty-mig", TestNamespaceName, TestQueueName,
				"test-workload-3", "test-workload-uid-3", map[string]int64{}),
			expectedPodGroupMetricLabels: &ExpectedPodGroupMetricLabels{
				queueName:    TestQueueName,
				podGroupName: "test-pg-empty-mig",
				podGroupUUID: "test-workload-uid-3",
				workloadName: "test-workload-3",
				workloadType: string(data.JobTypeTrain),
				project:      TestProjectName,
			},
			expectedMigDevices: map[string]int64{},
			shouldHaveMetric:   false,
		}),
	)
})

func ReconcileTestWorkload(ctx context.Context, workload client.Object,
	externalController *workload.ExternalWorkloadController,
	trainingController *workload.TrainingWorkloadController,
	interactiveController *workload.InteractiveWorkloadController,
	distributedController *workload.DistributedWorkloadController,
	inferenceController *workload.InferenceWorkloadController,
	distributedInferenceController *workload.DistributedInferenceWorkloadController) error {

	switch workload.(type) {
	case *v1alpha1.ExternalWorkload:
		_, err := externalController.Reconcile(ctx, getReconcileRequest(workload.GetName(), workload.GetNamespace()))
		return err
	case *twv2a1.TrainingWorkload:
		_, err := trainingController.Reconcile(ctx, getReconcileRequest(workload.GetName(), workload.GetNamespace()))
		return err
	case *iwv2a1.InteractiveWorkload:
		_, err := interactiveController.Reconcile(ctx, getReconcileRequest(workload.GetName(), workload.GetNamespace()))
		return err
	case *diwv2.DistributedWorkload:
		_, err := distributedController.Reconcile(ctx, getReconcileRequest(workload.GetName(), workload.GetNamespace()))
		return err
	case *infwv2.InferenceWorkload:
		_, err := inferenceController.Reconcile(ctx, getReconcileRequest(workload.GetName(), workload.GetNamespace()))
		return err
	case *dinfw.DistributedInferenceWorkload:
		_, err := distributedInferenceController.Reconcile(ctx, getReconcileRequest(workload.GetName(), workload.GetNamespace()))
		return err
	default:
		return nil
	}
}

func getExpectedMetricValue(expectedPodMetricValues *ExpectedPodMetricValues, index int) float64 {
	switch index {
	case 0:
		return expectedPodMetricValues.requestedGpus
	case 1:
		return expectedPodMetricValues.requestedGPUMemory
	case 2:
		return expectedPodMetricValues.limitGpus
	case 3:
		return expectedPodMetricValues.limitGPUMemory
	case 4:
		return expectedPodMetricValues.allocatedGpus
	case 5:
		return expectedPodMetricValues.allocatedGPUMemory
	case 6:
		return expectedPodMetricValues.allocatedCpus
	case 7:
		return expectedPodMetricValues.allocatedCpuMemory
	default:
		return 0
	}
}
