/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package converter_test

import (
	"context"
	"errors"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	kaiv1alpha1 "github.com/kai-scheduler/kai-resource-management-api/kai/v1alpha1"
	"github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1alpha1"
	"github.com/run-ai/runai/runai-cluster/common/constants"
	"github.com/run-ai/runai/runai-cluster/common/node-pool-utils/converter"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	fakeclient "sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/client/interceptor"
)

const (
	labelKeyNodePool = "runai/node-pool"
	kindPod          = "Pod"
)

// testIdentifiers carries the runai-flavored vocabulary used by the test
// suite. The converter package itself no longer hardcodes these values;
// each binary passes its own identifiers (here we use runai's so existing
// behavioral assertions stay valid).
var testIdentifiers = converter.NodePoolIdentifiers{
	NodePoolAssignmentLabelKey: labelKeyNodePool,
	DefaultNodepoolName:        v1alpha1.DefaultNodePoolName,
	UnexistingNodepoolSentinel: v1alpha1.UnexistingRunaiNodePool,
	AnnotationNodepoolsKey:     constants.AnnotationNodepools,
}

var _ = Describe("GetRequestedNodePools", func() {
	var (
		ctx        context.Context
		objectMeta metav1.ObjectMeta
		sources    converter.NodePoolsSources
		project    *kaiv1alpha1.Project
	)

	BeforeEach(func() {
		ctx = context.TODO()
		objectMeta = metav1.ObjectMeta{
			Namespace:   "ns",
			Name:        "obj",
			Annotations: map[string]string{},
			Labels:      map[string]string{},
		}
		project = &kaiv1alpha1.Project{Spec: kaiv1alpha1.ProjectSpec{DefaultNodePools: nil}}
		sources = converter.NodePoolsSources{
			PodAnnotations: nil,
			AffinitySource: converter.AffinitySource{},
			Project:        nil,
		}
	})

	Describe("Pod Annotations", func() {
		It("returns node pools from annotation if present and non-empty", func() {
			annotations := map[string]string{constants.AnnotationNodepools: "np1 np2"}
			sources.PodAnnotations = annotations
			result, err := converter.GetRequestedNodePools(ctx, objectMeta, kindPod, testIdentifiers, sources)
			Expect(err).ToNot(HaveOccurred())
			Expect(result).To(Equal([]string{"np1", "np2"}))
		})
		It("continues if annotation is present but empty", func() {
			objectMeta.Annotations = map[string]string{constants.AnnotationNodepools: ""}
			sources.PodAnnotations = objectMeta.Annotations
			result, err := converter.GetRequestedNodePools(ctx, objectMeta, kindPod, testIdentifiers, sources)
			Expect(err).ToNot(HaveOccurred())
			Expect(result).To(Equal([]string{v1alpha1.DefaultNodePoolName}))
		})
		It("continues if annotation is missing", func() {
			sources.PodAnnotations = map[string]string{}
			result, err := converter.GetRequestedNodePools(ctx, objectMeta, kindPod, testIdentifiers, sources)
			Expect(err).ToNot(HaveOccurred())
			Expect(result).To(Equal([]string{v1alpha1.DefaultNodePoolName}))
		})
	})

	Describe("Node Affinity", func() {
		const (
			existingNodePoolName  = "npA"
			existingNodePoolKey   = "key"
			existingNodePoolValue = "value"
		)
		var (
			affinity     *corev1.Affinity
			readerClient client.Reader
		)
		scheme := runtime.NewScheme()
		Expect(v1alpha1.AddToScheme(scheme)).To(Succeed())
		Expect(kaiv1alpha1.AddToScheme(scheme)).To(Succeed())

		BeforeEach(func() {
			nodepoolA := &v1alpha1.NodePool{
				ObjectMeta: metav1.ObjectMeta{
					Name: existingNodePoolName,
				},
				Spec: v1alpha1.NodePoolSpec{
					LabelKey:   existingNodePoolKey,
					LabelValue: existingNodePoolValue,
				},
			}
			// kai.resources twin of the fixture, so these specs resolve the same node pool
			// under both values of the compile-time UseKaiNodePools flag.
			// TODO: drop the legacy nodepoolA once clusterconstants.UseKaiNodePools=true
			kaiNodepoolA := &kaiv1alpha1.NodePool{
				ObjectMeta: metav1.ObjectMeta{
					Name: existingNodePoolName,
				},
				Spec: kaiv1alpha1.NodePoolSpec{
					LabelKey:   existingNodePoolKey,
					LabelValue: existingNodePoolValue,
				},
			}
			fakeClientBuilder := fakeclient.NewClientBuilder().WithScheme(scheme).WithObjects(nodepoolA, kaiNodepoolA)

			fakeListFn := func(ctx context.Context, client client.WithWatch, list client.ObjectList, opts ...client.ListOption) error {
				if ctx.Value("fail") == true {
					return errors.New("i was told to fail")
				}

				return client.List(ctx, list, opts...)
			}

			fakeClientBuilder.WithInterceptorFuncs(interceptor.Funcs{
				List: fakeListFn,
			})

			readerClient = fakeClientBuilder.Build()

			affinity = &corev1.Affinity{
				NodeAffinity: &corev1.NodeAffinity{
					RequiredDuringSchedulingIgnoredDuringExecution: &corev1.NodeSelector{},
				},
			}
			sources.AffinitySource = converter.AffinitySource{
				Affinity:     affinity,
				ReaderClient: readerClient,
			}
		})

		It("returns node pools from affinity if conversion succeeds and non-empty", func() {
			// term that matches the existing nodepool
			affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms = []corev1.NodeSelectorTerm{
				{
					MatchExpressions: []corev1.NodeSelectorRequirement{
						{
							Key:      existingNodePoolKey,
							Operator: corev1.NodeSelectorOpIn,
							Values:   []string{existingNodePoolValue},
						},
					},
				},
			}

			sources.AffinitySource.Affinity = affinity
			result, err := converter.GetRequestedNodePools(ctx, objectMeta, kindPod, testIdentifiers, sources)
			Expect(err).ToNot(HaveOccurred())
			Expect(result).To(Equal([]string{existingNodePoolName}))
		})
		It("continues if affinity conversion returns empty list", func() {
			// term that matches non-existing nodepool
			affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms = []corev1.NodeSelectorTerm{
				{
					MatchExpressions: []corev1.NodeSelectorRequirement{
						{
							Key:      "blabla",
							Operator: corev1.NodeSelectorOpIn,
							Values:   []string{"blabla2"},
						},
					},
				},
			}

			result, err := converter.GetRequestedNodePools(ctx, objectMeta, kindPod, testIdentifiers, sources)
			Expect(err).ToNot(HaveOccurred())
			Expect(result).To(Equal([]string{v1alpha1.DefaultNodePoolName}))
		})

		Context("affinity conversion errors", func() {
			BeforeEach(func() {
				// term that matches the existing nodepool
				affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms = []corev1.NodeSelectorTerm{
					{
						MatchExpressions: []corev1.NodeSelectorRequirement{
							{
								Key:      existingNodePoolKey,
								Operator: corev1.NodeSelectorOpIn,
								Values:   []string{existingNodePoolValue},
							},
						},
					},
				}

				// make the mock client return error
				ctx = context.WithValue(ctx, "fail", true)
			})
			It("continues if IgnoreConversionErrors is true", func() {
				sources.AffinitySource.IgnoreConversionErrors = true

				result, err := converter.GetRequestedNodePools(ctx, objectMeta, kindPod, testIdentifiers, sources)
				Expect(err).ToNot(HaveOccurred())
				Expect(result).To(Equal([]string{v1alpha1.DefaultNodePoolName}))
			})
			It("returns error if IgnoreConversionErrors is false", func() {
				result, err := converter.GetRequestedNodePools(ctx, objectMeta, kindPod, testIdentifiers, sources)
				Expect(err).To(HaveOccurred())
				Expect(errors.As(err, &converter.NodeAffinityConversionError{})).To(BeTrue())
				Expect(err.Error()).To(Equal("i was told to fail"))

				Expect(result).To(Equal([]string{}))
			})
		})
		It("continues if affinity is not set", func() {
			sources.AffinitySource.Affinity = nil
			result, err := converter.GetRequestedNodePools(ctx, objectMeta, kindPod, testIdentifiers, sources)
			Expect(err).ToNot(HaveOccurred())
			Expect(result).To(Equal([]string{v1alpha1.DefaultNodePoolName}))
		})
		It("continues if affinity is empty", func() {
			result, err := converter.GetRequestedNodePools(ctx, objectMeta, kindPod, testIdentifiers, sources)
			Expect(err).ToNot(HaveOccurred())
			Expect(result).To(Equal([]string{v1alpha1.DefaultNodePoolName}))
		})
	})

	Describe("Object Labels", func() {
		It("returns node pool from label if present and valid", func() {
			objectMeta.Labels = map[string]string{labelKeyNodePool: "npLabel"}
			result, err := converter.GetRequestedNodePools(ctx, objectMeta, kindPod, testIdentifiers, sources)
			Expect(err).ToNot(HaveOccurred())
			Expect(result).To(Equal([]string{"npLabel"}))
		})
		It("continues if label is present but UnexistingRunaiNodePool", func() {
			objectMeta.Labels = map[string]string{labelKeyNodePool: v1alpha1.UnexistingRunaiNodePool}
			result, err := converter.GetRequestedNodePools(ctx, objectMeta, kindPod, testIdentifiers, sources)
			Expect(err).ToNot(HaveOccurred())
			Expect(result).To(Equal([]string{v1alpha1.DefaultNodePoolName}))
		})
		It("continues if label is missing", func() {
			objectMeta.Labels = map[string]string{}
			result, err := converter.GetRequestedNodePools(ctx, objectMeta, kindPod, testIdentifiers, sources)
			Expect(err).ToNot(HaveOccurred())
			Expect(result).To(Equal([]string{v1alpha1.DefaultNodePoolName}))
		})
	})

	Describe("Project Default Node Pools", func() {
		It("returns project default node pools if present and non-empty", func() {
			project.Spec.DefaultNodePools = []string{"npProj1", "npProj2"}
			sources.Project = project
			result, err := converter.GetRequestedNodePools(ctx, objectMeta, kindPod, testIdentifiers, sources)
			Expect(err).ToNot(HaveOccurred())
			Expect(result).To(Equal([]string{"npProj1", "npProj2"}))
		})
		It("continues if project default node pools is nil or empty", func() {
			project.Spec.DefaultNodePools = nil
			sources.Project = project
			result, err := converter.GetRequestedNodePools(ctx, objectMeta, kindPod, testIdentifiers, sources)
			Expect(err).ToNot(HaveOccurred())
			Expect(result).To(Equal([]string{v1alpha1.DefaultNodePoolName}))
		})
		It("continues if project is not set", func() {
			sources.Project = nil
			result, err := converter.GetRequestedNodePools(ctx, objectMeta, kindPod, testIdentifiers, sources)
			Expect(err).ToNot(HaveOccurred())
			Expect(result).To(Equal([]string{v1alpha1.DefaultNodePoolName}))
		})
	})

	Describe("Fallback to Default Node Pool", func() {
		It("returns default node pool if all other sources fail", func() {
			result, err := converter.GetRequestedNodePools(ctx, objectMeta, kindPod, testIdentifiers, sources)
			Expect(err).ToNot(HaveOccurred())
			Expect(result).To(Equal([]string{v1alpha1.DefaultNodePoolName}))
		})
	})

	Describe("Edge Cases", func() {
		It("uses only the highest-priority source if multiple are set", func() {
			objectMeta.Annotations = map[string]string{constants.AnnotationNodepools: "np1 np2"}
			sources.PodAnnotations = objectMeta.Annotations
			objectMeta.Labels = map[string]string{labelKeyNodePool: "npLabel"}
			project.Spec.DefaultNodePools = []string{"npProj1"}
			sources.Project = project
			result, err := converter.GetRequestedNodePools(ctx, objectMeta, kindPod, testIdentifiers, sources)
			Expect(err).ToNot(HaveOccurred())
			Expect(result).To(Equal([]string{"np1", "np2"}))
		})
		It("handles annotations/labels with extra spaces", func() {
			objectMeta.Annotations = map[string]string{constants.AnnotationNodepools: " np1   np2  "}
			sources.PodAnnotations = objectMeta.Annotations
			result, err := converter.GetRequestedNodePools(ctx, objectMeta, kindPod, testIdentifiers, sources)
			Expect(err).ToNot(HaveOccurred())
			Expect(result).To(ContainElements("np1", "np2"))
		})
		It("handles nil maps for annotations/labels without panic", func() {
			objectMeta.Annotations = nil
			objectMeta.Labels = nil
			sources.PodAnnotations = nil
			result, err := converter.GetRequestedNodePools(ctx, objectMeta, kindPod, testIdentifiers, sources)
			Expect(err).ToNot(HaveOccurred())
			Expect(result).To(Equal([]string{v1alpha1.DefaultNodePoolName}))
		})
		It("returns default node pool if all sources are missing", func() {
			objectMeta.Annotations = nil
			objectMeta.Labels = nil
			sources.PodAnnotations = nil
			sources.Project = nil
			sources.AffinitySource = converter.AffinitySource{}
			result, err := converter.GetRequestedNodePools(ctx, objectMeta, kindPod, testIdentifiers, sources)
			Expect(err).ToNot(HaveOccurred())
			Expect(result).To(Equal([]string{v1alpha1.DefaultNodePoolName}))
		})
	})
})

// Covers the NodePoolsSources constructors: NewNodePoolsSources (legacy run.ai Project) and its
// kai.resources twin NewNodePoolSources (kai Project). Both wire the affinity/reader/annotations/
// project into the struct and map the AffinityConversionErrorHandling flag to IgnoreConversionErrors.
var _ = Describe("NewNodePoolsSources", func() {
	var (
		affinity     *corev1.Affinity
		readerClient client.Reader
		podAnnots    map[string]string
	)

	BeforeEach(func() {
		affinity = &corev1.Affinity{NodeAffinity: &corev1.NodeAffinity{}}
		readerClient = fakeclient.NewClientBuilder().Build()
		podAnnots = map[string]string{constants.AnnotationNodepools: "np1"}
	})

	Describe("NewNodePoolsSources", func() {
		It("populates every field and maps IgnoreAffinityConversionErrors to true", func() {
			project := &kaiv1alpha1.Project{Spec: kaiv1alpha1.ProjectSpec{DefaultNodePools: []string{"npA"}}}
			s := converter.NewNodePoolsSources(affinity, readerClient,
				converter.IgnoreAffinityConversionErrors, podAnnots, project)

			Expect(s.PodAnnotations).To(Equal(podAnnots))
			Expect(s.AffinitySource.Affinity).To(Equal(affinity))
			Expect(s.AffinitySource.ReaderClient).To(Equal(readerClient))
			Expect(s.AffinitySource.IgnoreConversionErrors).To(BeTrue())
			Expect(s.AffinitySource.HasValue()).To(BeTrue())
			Expect(s.Project).To(Equal(project))
		})

		It("maps ReturnAffinityConversionErrors to false and allows a nil project", func() {
			s := converter.NewNodePoolsSources(nil, nil,
				converter.ReturnAffinityConversionErrors, nil, nil)

			Expect(s.AffinitySource.IgnoreConversionErrors).To(BeFalse())
			Expect(s.AffinitySource.HasValue()).To(BeFalse())
			Expect(s.Project).To(BeNil())
		})

		It("produces sources that drive GetRequestedNodePools' project source", func() {
			project := &kaiv1alpha1.Project{Spec: kaiv1alpha1.ProjectSpec{DefaultNodePools: []string{"npA", "npB"}}}
			// no annotations/affinity/labels -> falls through to the project default node pools
			s := converter.NewNodePoolsSources(nil, nil,
				converter.IgnoreAffinityConversionErrors, nil, project)

			result, err := converter.GetRequestedNodePools(context.TODO(), metav1.ObjectMeta{}, kindPod, testIdentifiers, s)
			Expect(err).ToNot(HaveOccurred())
			Expect(result).To(Equal([]string{"npA", "npB"}))
		})
	})

})
