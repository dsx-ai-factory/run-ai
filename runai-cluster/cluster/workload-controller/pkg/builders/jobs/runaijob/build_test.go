package runaijob

import (
	"context"
	"reflect"

	flds "github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/workload_fields/field_types"
	nodepoolbuilders "github.com/run-ai/runai/runai-cluster/cluster/workload-common/builders/nodepool"
	commonconstants "github.com/run-ai/runai/runai-cluster/common/constants"

	kaiorg "github.com/kai-scheduler/kai-resource-management-api/kai/v1alpha1"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	clusterconstants "github.com/run-ai/runai/runai-cluster/cluster/common/constants"
	rjv1 "github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1"
	"github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1alpha1"
	runaiv2 "github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v2"
	workloadproject "github.com/run-ai/runai/runai-cluster/cluster/workload-common/project"
	"github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/jobs"
	"github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/builders/common"
	workloadCommon "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/common/constants"
	fakeClient "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/utils/testing_utils/client"
	"github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/utils/testing_utils/register"
	workload2 "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/utils/testing_utils/resource/workload"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	testProject = "test-project"
	name        = "test-job"
)

var (
	testNamespace = "runai-" + testProject
	nodePoolName  = "nodepoola"
	dataVolumeId  = "5e0445ea-6d67-4641-a791-b4a372ef722c"
)

var _ = Describe("Builder", func() {
	var (
		builder         *Builder
		workload        *jobs.JobWorkload
		exitingPodsSpec map[workloadCommon.ReplicaType]*corev1.PodSpec
		k8sClient       client.Client
		ctx             context.Context
	)
	BeforeEach(func() {
		scheme := runtime.NewScheme()
		register.InitSchema(scheme)
		ctx = workloadproject.SetProjectInContext(context.Background(), testProject)
		project := createProject(testProject)
		kaiProject := createKaiProject(testProject)
		nodePool := createNodePool(nodePoolName)

		k8sClient = fakeClient.InitFakeClientWithObjectNoStatus(scheme, project, kaiProject, nodePool, createKaiNodePool(nodePool.Name))
		builder = &Builder{
			k8sClient: k8sClient,
		}
		workload = workload2.NewJobWorkloadBuilder().WithInteractive().
			WithNamespace(testNamespace).WithNodePools(nodePoolName).WithImage("bla").WithName(name).Build()
		exitingPodsSpec = make(map[workloadCommon.ReplicaType]*corev1.PodSpec)
		exitingPodsSpec[workloadCommon.Worker] = &corev1.PodSpec{
			Affinity: &corev1.Affinity{
				NodeAffinity: &corev1.NodeAffinity{
					RequiredDuringSchedulingIgnoredDuringExecution: &corev1.NodeSelector{
						NodeSelectorTerms: []corev1.NodeSelectorTerm{
							{
								MatchExpressions: []corev1.NodeSelectorRequirement{
									{
										Key:      "other-key",
										Operator: corev1.NodeSelectorOpIn,
										Values:   []string{"test-value"},
									},
								},
							},
						},
					},
				},
			},
		}
	})

	It("builds RunaiJob spec successfully without existing node affinity", func() {
		computeResources, err := builder.Build(ctx, workload, &common.ComputeBuilderOptions{})
		Expect(err).To(BeNil())
		expectedAffinity := &corev1.Affinity{}
		runaiJobObj, ok := computeResources.Job.(*rjv1.RunaiJob)
		Expect(ok).To(BeTrue())

		err = nodepoolbuilders.CreateNodeAffinityForNodePools(ctx, k8sClient, workload.Spec.NodePools.Deref(), workload.Spec.NodePool.Deref(), &expectedAffinity)
		Expect(err).To(BeNil())
		Expect(runaiJobObj.Spec.Template.Spec.Affinity.NodeAffinity).To(Equal(expectedAffinity.NodeAffinity))

	})
	It("builds RunaiJob spec successfully with existing node affinity", func() {
		computeResources, err := builder.Build(ctx, workload, &common.ComputeBuilderOptions{ExitingPodsSpec: exitingPodsSpec})
		Expect(err).To(BeNil())
		affinity := &corev1.Affinity{}
		runaiJobObj, ok := computeResources.Job.(*rjv1.RunaiJob)
		Expect(ok).To(BeTrue())

		err = nodepoolbuilders.CreateNodeAffinityForNodePools(ctx, k8sClient, workload.Spec.NodePools.Deref(), workload.Spec.NodePool.Deref(), &affinity)
		Expect(err).To(BeNil())
		// don't use nodepool node affinity
		Expect(runaiJobObj.Spec.Template.Spec.Affinity.NodeAffinity).ToNot(Equal(affinity.NodeAffinity))
		// use existing node affinity
		Expect(reflect.DeepEqual(runaiJobObj.Spec.Template.Spec.Affinity.NodeAffinity,
			exitingPodsSpec[workloadCommon.Worker].Affinity.NodeAffinity)).To(BeTrue())
	})
	Context("job has old image", func() {
		const oldImage = "old-image"
		BeforeEach(func() {
			exitingPodsSpec = make(map[workloadCommon.ReplicaType]*corev1.PodSpec)
			exitingPodsSpec[workloadCommon.Worker] = &corev1.PodSpec{
				Containers: []corev1.Container{
					{
						Name:  "worker",
						Image: oldImage,
					},
				},
			}
		})
		It("builds RunaiJob spec successfully", func() {
			computeResources, err := builder.Build(ctx, workload, &common.ComputeBuilderOptions{ExitingPodsSpec: exitingPodsSpec})
			Expect(err).To(BeNil())
			runaiJobObj, ok := computeResources.Job.(*rjv1.RunaiJob)
			Expect(ok).To(BeTrue())
			Expect(runaiJobObj.Spec.Template.Spec.Containers[0].Image).To(Equal(oldImage))
		})
	})

	Context("Nodepools annotation", func() {
		Context("When has no nodepools", func() {
			It("when has no nodepools", func() {
				workload.Spec.NodePools = nil

				computeResources, err := builder.Build(ctx, workload, &common.ComputeBuilderOptions{ExitingPodsSpec: exitingPodsSpec})
				Expect(err).To(BeNil())
				runaiJobObj, ok := computeResources.Job.(*rjv1.RunaiJob)
				Expect(ok).To(BeTrue())

				_, exist := runaiJobObj.Spec.Template.Annotations[commonconstants.AnnotationNodepools]
				Expect(exist).To(BeFalse())
			})
		})

		Context("When has nodepools", func() {
			It("when has Nodepools", func() {
				computeResources, err := builder.Build(ctx, workload, &common.ComputeBuilderOptions{ExitingPodsSpec: exitingPodsSpec})
				Expect(err).To(BeNil())
				runaiJobObj, ok := computeResources.Job.(*rjv1.RunaiJob)
				Expect(ok).To(BeTrue())

				value, exist := runaiJobObj.Spec.Template.Annotations[commonconstants.AnnotationNodepools]
				Expect(exist).To(BeTrue())
				Expect(value).To(Equal(nodePoolName))
			})

			It("when has Nodepool", func() {
				workload.Spec.NodePools = nil
				workload.Spec.NodePool = flds.NewOptionalStringField(nodePoolName, "")

				computeResources, err := builder.Build(ctx, workload, &common.ComputeBuilderOptions{ExitingPodsSpec: exitingPodsSpec})
				Expect(err).To(BeNil())
				runaiJobObj, ok := computeResources.Job.(*rjv1.RunaiJob)
				Expect(ok).To(BeTrue())

				value, exist := runaiJobObj.Spec.Template.Annotations[commonconstants.AnnotationNodepools]
				Expect(exist).To(BeTrue())
				Expect(value).To(Equal(nodePoolName))
			})

			It("when has Nodepools and Nodepool - uses Nodepools", func() {
				workload.Spec.NodePool = flds.NewOptionalStringField("blabla", "")

				computeResources, err := builder.Build(ctx, workload, &common.ComputeBuilderOptions{ExitingPodsSpec: exitingPodsSpec})
				Expect(err).To(BeNil())
				runaiJobObj, ok := computeResources.Job.(*rjv1.RunaiJob)
				Expect(ok).To(BeTrue())

				value, exist := runaiJobObj.Spec.Template.Annotations[commonconstants.AnnotationNodepools]
				Expect(exist).To(BeTrue())
				Expect(value).To(Equal(nodePoolName))
			})
		})
	})
})

func createProject(name string) *runaiv2.Project {
	fakeProject := runaiv2.Project{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
		Spec: runaiv2.ProjectSpec{
			InteractiveJobTimeLimitSecs:                  120,
			InteractiveJobMaxIdleDurationSecs:            1200,
			InteractivePreemptibleJobMaxIdleDurationSecs: 1800,
			TrainingJobMaxIdleDurationSecs:               2400,
			TrainingJobTimeLimitSecs:                     120,
		},
	}
	return &fakeProject
}

// createKaiProject builds the kai.resources twin of createProject, carrying the same
// limits as the run.ai/*-seconds annotations the builder reads.
func createKaiProject(name string) *kaiorg.Project {
	return &kaiorg.Project{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
			Annotations: map[string]string{
				clusterconstants.InteractiveJobTimeLimitSecondsAnnotation:                  "120",
				clusterconstants.InteractiveJobMaxIdleDurationSecondsAnnotation:            "1200",
				clusterconstants.InteractivePreemptibleJobMaxIdleDurationSecondsAnnotation: "1800",
				clusterconstants.TrainingJobMaxIdleDurationSecondsAnnotation:               "2400",
				clusterconstants.TrainingJobTimeLimitSecondsAnnotation:                     "120",
			},
		},
	}
}

func createNodePool(name string) *v1alpha1.NodePool {
	fakeNodePool := v1alpha1.NodePool{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
		Spec: v1alpha1.NodePoolSpec{
			LabelKey:   "test-key",
			LabelValue: "test-value",
		},
	}
	return &fakeNodePool
}

// createKaiNodePool is the kai.resources twin of createNodePool. Seed both so these
// specs resolve the same node pool under either value of the compile-time
// UseKaiNodePools flag.
// TODO: delete once constants.UseKaiNodePools=true and createNodePool goes
func createKaiNodePool(name string) *kaiorg.NodePool {
	return &kaiorg.NodePool{
		ObjectMeta: metav1.ObjectMeta{Name: name},
		Spec: kaiorg.NodePoolSpec{
			LabelKey:   "test-key",
			LabelValue: "test-value",
		},
	}
}
