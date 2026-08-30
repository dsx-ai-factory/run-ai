package main

import (
	"errors"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	dp "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/distributedpolicy/v2alpha1"
	dw "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/distributedworkload/v2alpha1"
	ip "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/interactivepolicy/v2alpha1"
	iw "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/interactiveworkload/v2alpha1"
	tp "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/trainingpolicy/v2alpha1"
	tw "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/trainingworkload/v2alpha1"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

var _ = Describe("resolveAPIAvailability", func() {
	gvks := []schema.GroupVersionKind{
		iw.GroupVersion.WithKind("InteractiveWorkload"), ip.GroupVersion.WithKind("InteractivePolicy"),
		tw.GroupVersion.WithKind("TrainingWorkload"), tp.GroupVersion.WithKind("TrainingPolicy"),
		dw.GroupVersion.WithKind("DistributedWorkload"), dp.GroupVersion.WithKind("DistributedPolicy"),
	}
	discoveryErr := errors.New("discovery unavailable")

	DescribeTable("resolves exact API availability",
		func(mappingErrors map[schema.GroupVersionKind]error, want apiAvailability, wantErr error, wantMapped []schema.GroupVersionKind) {
			resolver := &availabilityResolver{errors: mappingErrors}

			got, err := resolveAPIAvailability(resolver)

			if wantErr == nil {
				Expect(err).NotTo(HaveOccurred())
				Expect(got).To(Equal(want))
			} else {
				Expect(err).To(MatchError(wantErr))
			}
			Expect(resolver.mapped).To(Equal(wantMapped))
		},
		Entry("all present",
			nil,
			apiAvailability{true, true, true, true, true, true},
			nil,
			gvks,
		),
		Entry("mixed workloads and policies",
			map[schema.GroupVersionKind]error{gvks[1]: noMatch(gvks[1]), gvks[4]: noMatch(gvks[4])},
			apiAvailability{interactiveWorkload: true, trainingWorkload: true, trainingPolicy: true, distributedPolicy: true},
			nil,
			gvks,
		),
		Entry("all missing", allNoMatch(gvks), apiAvailability{}, nil, gvks),
		Entry("discovery error",
			map[schema.GroupVersionKind]error{gvks[2]: discoveryErr},
			apiAvailability{},
			discoveryErr,
			gvks,
		),
	)
})

func noMatch(gvk schema.GroupVersionKind) error {
	return &meta.NoKindMatchError{GroupKind: gvk.GroupKind(), SearchedVersions: []string{gvk.Version}}
}

func allNoMatch(gvks []schema.GroupVersionKind) map[schema.GroupVersionKind]error {
	errs := make(map[schema.GroupVersionKind]error, len(gvks))
	for _, gvk := range gvks {
		errs[gvk] = noMatch(gvk)
	}
	return errs
}

type availabilityResolver struct {
	errors map[schema.GroupVersionKind]error
	mapped []schema.GroupVersionKind
}

func (r *availabilityResolver) RESTMapping(groupKind schema.GroupKind, versions ...string) (*meta.RESTMapping, error) {
	gvk := schema.GroupVersionKind{Group: groupKind.Group, Kind: groupKind.Kind}
	if len(versions) > 0 {
		gvk.Version = versions[0]
	}
	r.mapped = append(r.mapped, gvk)
	return &meta.RESTMapping{}, r.errors[gvk]
}
