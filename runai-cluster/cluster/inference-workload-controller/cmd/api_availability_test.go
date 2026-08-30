package main

import (
	"errors"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	dinfw "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/distributedinferenceworkload/v1alpha1"
	infp "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/inferencepolicy/v2alpha1"
	infw "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/inferenceworkload/v2alpha1"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

var _ = Describe("resolveAPIAvailability", func() {
	inferenceWorkload := infw.GroupVersion.WithKind("InferenceWorkload")
	inferencePolicy := infp.GroupVersion.WithKind("InferencePolicy")
	distributedInferenceWorkload := dinfw.GroupVersion.WithKind("DistributedInferenceWorkload")
	allGVKs := []schema.GroupVersionKind{inferenceWorkload, inferencePolicy, distributedInferenceWorkload}
	discoveryErr := errors.New("discovery unavailable")

	DescribeTable("resolves exact API availability",
		func(mappingErrors map[schema.GroupVersionKind]error, want apiAvailability, wantErr error, wantMapped []schema.GroupVersionKind) {
			resolver := &apiAvailabilityTestResolver{errors: mappingErrors}

			got, err := resolveAPIAvailability(resolver)

			if wantErr == nil {
				Expect(err).NotTo(HaveOccurred())
				Expect(got).To(Equal(want))
			} else {
				Expect(err).To(MatchError(wantErr))
			}
			Expect(resolver.mappedGVKs).To(Equal(wantMapped))
		},
		Entry("reports every API as available when exact mappings exist",
			nil,
			apiAvailability{inferenceWorkload: true, inferencePolicy: true, distributedInferenceWorkload: true},
			nil,
			allGVKs,
		),
		Entry("tracks a partial legacy API set independently",
			map[schema.GroupVersionKind]error{
				inferencePolicy:              noMatch(inferencePolicy),
				distributedInferenceWorkload: noMatch(distributedInferenceWorkload),
			},
			apiAvailability{inferenceWorkload: true},
			nil,
			allGVKs,
		),
		Entry("reports every API as unavailable when all CRDs are absent",
			map[schema.GroupVersionKind]error{
				inferenceWorkload:            noMatch(inferenceWorkload),
				inferencePolicy:              noMatch(inferencePolicy),
				distributedInferenceWorkload: noMatch(distributedInferenceWorkload),
			},
			apiAvailability{},
			nil,
			allGVKs,
		),
		Entry("returns non NoMatch discovery errors",
			map[schema.GroupVersionKind]error{
				inferencePolicy: discoveryErr,
			},
			apiAvailability{},
			discoveryErr,
			allGVKs,
		),
	)
})

func noMatch(gvk schema.GroupVersionKind) error {
	return &meta.NoKindMatchError{GroupKind: gvk.GroupKind(), SearchedVersions: []string{gvk.Version}}
}

type apiAvailabilityTestResolver struct {
	errors     map[schema.GroupVersionKind]error
	mappedGVKs []schema.GroupVersionKind
}

func (r *apiAvailabilityTestResolver) RESTMapping(groupKind schema.GroupKind, versions ...string) (*meta.RESTMapping, error) {
	gvk := schema.GroupVersionKind{Group: groupKind.Group, Kind: groupKind.Kind}
	if len(versions) > 0 {
		gvk.Version = versions[0]
	}
	r.mappedGVKs = append(r.mappedGVKs, gvk)
	return &meta.RESTMapping{}, r.errors[gvk]
}
