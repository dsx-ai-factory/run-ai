package deletion_test

import (
	"context"
	"errors"

	kaiv1alpha1 "github.com/kai-scheduler/kai-resource-management-api/kai/v1alpha1"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/run-ai/runai/runai-cluster/cluster/project-controller/pkg/handlers/deletion"
	. "github.com/run-ai/runai/runai-cluster/cluster/project-controller/pkg/test"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/yaml"
)

var _ = Describe("BlockerGroupsFromConfigMapData", func() {
	It("returns no groups for empty data", func() {
		groups, err := BlockerGroupsFromConfigMapData(nil)
		Expect(err).NotTo(HaveOccurred())
		Expect(groups).To(BeEmpty())
	})

	It("groups blockers sharing a displayName into one group", func() {
		// A flat list where two Blockers share the "Workloads" displayName and one is "Secrets".
		// They must collapse into two groups: Workloads (2 queries) and Secrets (1 query).
		// Group order is not significant, so look groups up by displayName.
		blockers := []Blocker{
			{DisplayName: "Workloads", Group: "run.ai", Version: "v2alpha1", Kind: "TrainingWorkload"},
			{DisplayName: "Workloads", Group: "run.ai", Version: "v2alpha1", Kind: "InferenceWorkload"},
			{DisplayName: "Secrets", Group: "", Version: "v1", Kind: "Secret", LabelSelector: &metav1.LabelSelector{
				MatchExpressions: []metav1.LabelSelectorRequirement{
					{Key: "run.ai/department", Operator: metav1.LabelSelectorOpDoesNotExist},
					{Key: "run.ai/resource", Operator: metav1.LabelSelectorOpIn, Values: []string{"password"}},
				},
			}},
		}

		encoded, err := yaml.Marshal(blockers)
		Expect(err).NotTo(HaveOccurred())
		data := map[string]string{"blockers.yaml": string(encoded)}

		groups, err := BlockerGroupsFromConfigMapData(data)
		Expect(err).NotTo(HaveOccurred())
		Expect(groups).To(HaveLen(2))

		byName := map[string]BlockerGroup{}
		for _, g := range groups {
			byName[g.DisplayName] = g
		}

		Expect(byName).To(HaveKey("Workloads"))
		Expect(byName["Workloads"].ConditionType()).To(Equal("WorkloadsReady"))
		Expect(byName["Workloads"].Reason()).To(Equal("WorkloadsDeletionHandlerFailed"))
		Expect(byName["Workloads"].Blockers).To(HaveLen(2))

		Expect(byName).To(HaveKey("Secrets"))
		Expect(byName["Secrets"].Blockers).To(HaveLen(1))
	})

	It("rejects a blocker with no displayName", func() {
		data := map[string]string{"blockers.yaml": "- version: v1\n  kind: Secret\n"}
		_, err := BlockerGroupsFromConfigMapData(data)
		Expect(err).To(HaveOccurred())
	})

	It("rejects a blocker with no kind", func() {
		data := map[string]string{"blockers.yaml": "- displayName: Secrets\n  version: v1\n"}
		_, err := BlockerGroupsFromConfigMapData(data)
		Expect(err).To(HaveOccurred())
	})
})

var _ = Describe("ConfigurableBlocker", func() {
	var project kaiv1alpha1.Project

	BeforeEach(func() {
		project = *TestProject.DeepCopy()
	})

	It("does not block a fresh installation when a configured legacy API has no matching kind", func() {
		gvk := schema.GroupVersionKind{Group: "run.ai", Version: "v2alpha1", Kind: "TrainingWorkloadList"}
		cachedClient := fake.NewClientBuilder().WithScheme(scheme).Build()
		Expect(cachedClient.Create(context.Background(), TestNamespace.DeepCopy())).To(Succeed())
		k8sClient := listErrorClient{
			Client: cachedClient,
			errors: map[schema.GroupVersionKind]error{
				gvk: &meta.NoKindMatchError{GroupKind: gvk.GroupKind(), SearchedVersions: []string{gvk.Version}},
			},
		}
		handler := NewConfigurableBlocker(k8sClient, BlockerGroup{
			DisplayName: "Workloads",
			Blockers:    []Blocker{{Group: "run.ai", Version: "v2alpha1", Kind: "TrainingWorkload"}},
		})

		conditions, err := handler.OnDelete(&project)

		Expect(err).NotTo(HaveOccurred())
		Expect(conditions).To(HaveLen(1))
		Expect(conditions[0].Status).To(Equal(corev1.ConditionTrue))
	})

	It("does not block when the configured resource API is not found", func() {
		gvk := schema.GroupVersionKind{Group: "run.ai", Version: "v2alpha1", Kind: "TrainingWorkloadList"}
		cachedClient := fake.NewClientBuilder().WithScheme(scheme).Build()
		Expect(cachedClient.Create(context.Background(), TestNamespace.DeepCopy())).To(Succeed())
		k8sClient := listErrorClient{
			Client: cachedClient,
			errors: map[schema.GroupVersionKind]error{
				gvk: apierrors.NewNotFound(schema.GroupResource{Group: gvk.Group, Resource: "trainingworkloads"}, ""),
			},
		}
		handler := NewConfigurableBlocker(k8sClient, BlockerGroup{
			DisplayName: "Workloads",
			Blockers:    []Blocker{{Group: "run.ai", Version: "v2alpha1", Kind: "TrainingWorkload"}},
		})

		conditions, err := handler.OnDelete(&project)

		Expect(err).NotTo(HaveOccurred())
		Expect(conditions).To(HaveLen(1))
		Expect(conditions[0].Status).To(Equal(corev1.ConditionTrue))
	})

	It("returns a forbidden error from the configured resource API", func() {
		gvk := schema.GroupVersionKind{Group: "run.ai", Version: "v2alpha1", Kind: "TrainingWorkloadList"}
		forbiddenErr := apierrors.NewForbidden(schema.GroupResource{Group: gvk.Group, Resource: "trainingworkloads"}, "", errors.New("denied"))
		cachedClient := fake.NewClientBuilder().WithScheme(scheme).Build()
		Expect(cachedClient.Create(context.Background(), TestNamespace.DeepCopy())).To(Succeed())
		k8sClient := listErrorClient{
			Client: cachedClient,
			errors: map[schema.GroupVersionKind]error{
				gvk: forbiddenErr,
			},
		}
		handler := NewConfigurableBlocker(k8sClient, BlockerGroup{
			DisplayName: "Workloads",
			Blockers:    []Blocker{{Group: "run.ai", Version: "v2alpha1", Kind: "TrainingWorkload"}},
		})

		conditions, err := handler.OnDelete(&project)

		Expect(err).To(MatchError(forbiddenErr))
		Expect(conditions).To(HaveLen(1))
		Expect(conditions[0].Status).To(Equal(corev1.ConditionFalse))
	})

	It("returns an unexpected error from the configured resource API", func() {
		gvk := schema.GroupVersionKind{Group: "run.ai", Version: "v2alpha1", Kind: "TrainingWorkloadList"}
		unexpectedErr := errors.New("reader failed")
		cachedClient := fake.NewClientBuilder().WithScheme(scheme).Build()
		Expect(cachedClient.Create(context.Background(), TestNamespace.DeepCopy())).To(Succeed())
		k8sClient := listErrorClient{
			Client: cachedClient,
			errors: map[schema.GroupVersionKind]error{
				gvk: unexpectedErr,
			},
		}
		handler := NewConfigurableBlocker(k8sClient, BlockerGroup{
			DisplayName: "Workloads",
			Blockers:    []Blocker{{Group: "run.ai", Version: "v2alpha1", Kind: "TrainingWorkload"}},
		})

		conditions, err := handler.OnDelete(&project)

		Expect(err).To(MatchError(unexpectedErr))
		Expect(conditions).To(HaveLen(1))
		Expect(conditions[0].Status).To(Equal(corev1.ConditionFalse))
	})
})
