package inference_workload_controller

import (
	"context"
	"fmt"
	"time"

	commonconstants "github.com/run-ai/runai/runai-cluster/cluster/common/constants"
	"github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/workload_fields/field_types"
	"github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/inferenceworkload/v2alpha1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/utils/ptr"
	v1 "knative.dev/serving/pkg/apis/serving/v1"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/run-ai/runai/runai-cluster/cluster/integration-tests/common"
	"github.com/run-ai/runai/runai-cluster/cluster/integration-tests/config"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const image = "gcr.io/run-ai-lab/knative-helloworld-go"

var _ = Describe("Inference Workload Controller Test", Ordered, func() {
	var (
		testConfig         *config.TestConfig
		kubeClient         client.Client
		ctx                context.Context
		testCreatedObjects = map[string]client.Object{}
	)

	BeforeAll(func() {
		Skip("")
		testConfig = config.GetTestConfig(Default)
		kubeClient = testConfig.KubeClient
		ctx = testConfig.Ctx
	})

	AfterAll(func() {
		for _, obj := range testCreatedObjects {
			if err := kubeClient.Delete(ctx, obj); err != nil {
				fmt.Printf("Failed to delete resource %v: %v\n", obj, err)
			}
		}
		testConfig.Cleanup()
	})
	It("Submit inference workload with single pod and revision running phase", func() {
		podNamePrefix := common.GetRandomName("infw")
		inf, err := common.LoadWorkload("../../resources/inference.yaml", common.GetRandomName("inference-wl"), testConfig.TestNamespace.Name, podNamePrefix, &v2alpha1.InferenceWorkload{}, nil)
		Expect(err).ToNot(HaveOccurred(), "Failed to load InferenceWorkload")

		testCreatedObjects[podNamePrefix] = inf
		err = kubeClient.Create(ctx, inf)
		Expect(err).ToNot(HaveOccurred(), "Failed to create inference workload")

		var createdPod *corev1.Pod
		err = common.WaitForCondition(ctx, 10*time.Second, 180*time.Second, func() bool {
			createdPod, _ = common.GetPodByPrefix(ctx, kubeClient, podNamePrefix, testConfig.TestNamespace.Name)
			return createdPod != nil && createdPod.Status.Phase == corev1.PodRunning
		})
		Expect(err).ToNot(HaveOccurred(),
			fmt.Sprintf("Expected one pod with a name starting with '%s' in namespace '%s'.",
				podNamePrefix, testConfig.TestNamespace.Name))
		Expect(createdPod.Spec.Containers[0].Image).To(ContainSubstring(image), fmt.Sprintf("Expected the pod's image to be %s", image))

		revisions := &v1.RevisionList{}
		err = kubeClient.List(context.Background(), revisions, &client.ListOptions{Namespace: testConfig.TestNamespace.Name})
		Expect(err).To(BeNil())
		Expect(len(revisions.Items)).To(Equal(1))

		revision := revisions.Items[0]
		phase, ok := revision.Status.Annotations[commonconstants.RunaiPhaseAnnotationKey]
		Expect(ok).To(BeTrue())
		Expect(phase).To(Equal("Running"))

		phase, ok = revision.Status.Annotations[commonconstants.RunaiPhaseMessageAnnotationKey]
		Expect(ok).To(BeTrue())
		Expect(phase).To(Equal(""))

		err = kubeClient.Delete(ctx, inf)
		Expect(err).To(BeNil())
	})
	It("Submit inference workload with an error", func() {
		podNamePrefix := common.GetRandomName("infw")
		inf, err := common.LoadWorkload("../../resources/inference.yaml", common.GetRandomName("inference-wl"), testConfig.TestNamespace.Name, podNamePrefix, &v2alpha1.InferenceWorkload{}, nil)
		Expect(err).ToNot(HaveOccurred(), "Failed to load InferenceWorkload")
		inf.Spec.Image = &field_types.StringField{Value: ptr.To("wrong-image")}

		testCreatedObjects[podNamePrefix] = inf
		err = kubeClient.Create(ctx, inf)
		Expect(err).ToNot(HaveOccurred(), "Failed to create inference workload")

		revisions := &v1.RevisionList{}
		err = common.WaitForCondition(ctx, 10*time.Second, 180*time.Second, func() bool {
			err = kubeClient.List(context.Background(), revisions, &client.ListOptions{Namespace: testConfig.TestNamespace.Name})
			return err == nil && len(revisions.Items) == 1 && revisions.Items[0].Status.Annotations[commonconstants.RunaiPhaseAnnotationKey] == "Failed"
		})
		Expect(err).ToNot(HaveOccurred(),
			fmt.Sprintf("Expected one failed revision in namespace '%s'.",
				testConfig.TestNamespace.Name))

		phase, ok := revisions.Items[0].Status.Annotations[commonconstants.RunaiPhaseMessageAnnotationKey]
		Expect(ok).To(BeTrue())
		Expect(phase).To(ContainSubstring("Revision failed: Unable to fetch image"))

		err = kubeClient.Delete(ctx, inf)
		Expect(err).To(BeNil())
	})
	It("Submit over quota inference workload with single pod and revision pending phase, update and wait for running", func() {
		podNamePrefix := common.GetRandomName("infw")
		inf, err := common.LoadWorkload("../../resources/inference.yaml", common.GetRandomName("inference-wl"), testConfig.TestNamespace.Name, podNamePrefix, &v2alpha1.InferenceWorkload{}, nil)
		Expect(err).ToNot(HaveOccurred(), "Failed to load InferenceWorkload")
		inf.Spec.Gpu = &field_types.NumberField{
			Value: ptr.To("100"),
		}

		testCreatedObjects[podNamePrefix] = inf
		err = kubeClient.Create(ctx, inf)
		Expect(err).ToNot(HaveOccurred(), "Failed to create inference workload")

		var createdPod *corev1.Pod
		err = common.WaitForCondition(ctx, 10*time.Second, 180*time.Second, func() bool {
			createdPod, _ = common.GetPodByPrefix(ctx, kubeClient, podNamePrefix, testConfig.TestNamespace.Name)
			return createdPod != nil && createdPod.Status.Phase == corev1.PodPending
		})
		Expect(err).ToNot(HaveOccurred(),
			fmt.Sprintf("Expected one pod with a name starting with '%s' in namespace '%s'.",
				podNamePrefix, testConfig.TestNamespace.Name))
		Expect(createdPod.Spec.Containers[0].Image).To(ContainSubstring(image), fmt.Sprintf("Expected the pod's image to be %s", image))

		revisions := &v1.RevisionList{}
		err = kubeClient.List(context.Background(), revisions, &client.ListOptions{Namespace: testConfig.TestNamespace.Name})
		Expect(err).To(BeNil())
		Expect(len(revisions.Items)).To(Equal(1))

		revision := revisions.Items[0]
		phase, ok := revision.Status.Annotations[commonconstants.RunaiPhaseAnnotationKey]
		Expect(ok).To(BeTrue())
		Expect(phase).To(Equal("Pending"))

		phase, ok = revision.Status.Annotations[commonconstants.RunaiPhaseMessageAnnotationKey]
		Expect(ok).To(BeTrue())
		Expect(phase).To(Equal("Revision is waiting to be scheduled"))

		err = kubeClient.Get(ctx, types.NamespacedName{Namespace: inf.Namespace, Name: inf.Name}, inf)
		Expect(err).ToNot(HaveOccurred(), "Failed to get inference workload")
		inf.Spec.Gpu = nil

		err = kubeClient.Update(ctx, inf)
		Expect(err).ToNot(HaveOccurred(), "Failed to updating inference workload")

		err = common.WaitForCondition(ctx, 10*time.Second, 180*time.Second, func() bool {
			err = kubeClient.List(context.Background(), revisions, &client.ListOptions{Namespace: testConfig.TestNamespace.Name})
			return len(revisions.Items) == 2 && revisions.Items[1].Status.Annotations[commonconstants.RunaiPhaseAnnotationKey] == "Running"
		})

		Expect(err).ToNot(HaveOccurred(),
			fmt.Sprintf("Expected new revision to become running."))

		err = kubeClient.Delete(ctx, inf)
		Expect(err).To(BeNil())
	})
})
